// internal/generator/planet/descriptions_load.go
package planet

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ==================== ЗАГРУЗКА ====================
//
// Менеджер сканирует базовую папку (обычно config/descriptions/),
// обходит все подпапки (по одной на геймдизайнерский тип),
// читает openings_*.json и closings_*.json, складывает записи
// в плоские срезы. Хардкод-списка файлов нет — новые файлы
// подхватываются автоматически при следующем старте сервера.

// LoadDescriptions — читает все описания из базовой папки.
// Возвращает готовый менеджер или ошибку (тогда сервер не стартует).
func LoadDescriptions(baseDir string) (*descriptionsManager, error) {
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("descriptions: папка %q недоступна: %w", baseDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("descriptions: %q — не папка", baseDir)
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("descriptions: чтение %q: %w", baseDir, err)
	}

	m := &descriptionsManager{
		openings: make(map[string][]Opening),
		closings: make(map[string][]Closing),
		stats:    make(map[string]descriptionsStats),
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		typeName := entry.Name()
		folder := filepath.Join(baseDir, typeName)

		openings, openingFiles, err := loadOpenings(folder)
		if err != nil {
			return nil, fmt.Errorf("descriptions/%s: %w", typeName, err)
		}
		closings, closingFiles, err := loadClosings(folder)
		if err != nil {
			return nil, fmt.Errorf("descriptions/%s: %w", typeName, err)
		}

		if len(openings) == 0 && len(closings) == 0 {
			// Пустая папка — не ошибка, но сообщаем в лог.
			log.Printf("descriptions: %s: папка пуста, пропускаю", typeName)
			continue
		}

		m.openings[typeName] = openings
		m.closings[typeName] = closings
		m.stats[typeName] = descriptionsStats{
			OpeningsCount: len(openings),
			ClosingsCount: len(closings),
			OpeningsFiles: openingFiles,
			ClosingsFiles: closingFiles,
		}
	}

	logSummary(m)
	return m, nil
}

// ==================== ЗАЧИНЫ ====================

// loadOpenings — читает все openings_*.json из папки типа.
// Возвращает записи и число прочитанных файлов.
func loadOpenings(folder string) ([]Opening, int, error) {
	files, err := findFiles(folder, openingsFilePrefix)
	if err != nil {
		return nil, 0, err
	}

	var result []Opening
	for _, name := range files {
		path := filepath.Join(folder, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, 0, fmt.Errorf("чтение %s: %w", name, err)
		}

		var parsed openingFile
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return nil, 0, fmt.Errorf("парсинг %s: %w", name, err)
		}

		for i, op := range parsed.Openings {
			if strings.TrimSpace(op.Text) == "" {
				return nil, 0, fmt.Errorf("%s: запись #%d пустая", name, i+1)
			}
			result = append(result, op)
		}
	}

	return result, len(files), nil
}

// ==================== КОНЦОВКИ ====================

// loadClosings — читает все closings_*.json из папки типа.
func loadClosings(folder string) ([]Closing, int, error) {
	files, err := findFiles(folder, closingsFilePrefix)
	if err != nil {
		return nil, 0, err
	}

	var result []Closing
	for _, name := range files {
		path := filepath.Join(folder, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, 0, fmt.Errorf("чтение %s: %w", name, err)
		}

		var parsed closingFile
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return nil, 0, fmt.Errorf("парсинг %s: %w", name, err)
		}

		for i, cl := range parsed.Closings {
			if strings.TrimSpace(cl.Text) == "" {
				return nil, 0, fmt.Errorf("%s: запись #%d пустая", name, i+1)
			}
			result = append(result, cl)
		}
	}

	return result, len(files), nil
}

// ==================== ПОИСК ФАЙЛОВ ====================

// findFiles — возвращает имена файлов в папке, начинающиеся с prefix
// и заканчивающиеся на .json. Отсортированы лексикографически.
func findFiles(folder, prefix string) ([]string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("чтение папки: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if !strings.HasSuffix(name, jsonSuffix) {
			continue
		}
		names = append(names, name)
	}

	sort.Strings(names)
	return names, nil
}

// ==================== ЛОГ ====================

// logSummary — пишет в лог сводку по всем загруженным типам.
func logSummary(m *descriptionsManager) {
	totalOpenings, totalClosings := 0, 0
	for typeName, s := range m.stats {
		log.Printf(
			"descriptions: %-12s openings=%3d (%d файлов), closings=%3d (%d файлов)",
			typeName, s.OpeningsCount, s.OpeningsFiles,
			s.ClosingsCount, s.ClosingsFiles,
		)
		totalOpenings += s.OpeningsCount
		totalClosings += s.ClosingsCount
	}
	log.Printf(
		"descriptions: итого типов=%d, зачинов=%d, концовок=%d",
		len(m.stats), totalOpenings, totalClosings,
	)
}