// internal/audit/types.go
package audit

// ==================== УРОВНИ КРИТИЧНОСТИ ====================

const (
	SeverityLow    = "low"    // мелкие несоответствия, не влияют на геймплей
	SeverityMedium = "medium" // заметные противоречия, странно выглядят
	SeverityHigh   = "high"   // физически невозможно или сломает геймплей
)

// ==================== ПРОБЛЕМА ====================

// Issue — одна найденная проблема в планете.
type Issue struct {
	PlanetID    string                 `json:"planet_id"`
	PlanetName  string                 `json:"planet_name"`
	WorldID     string                 `json:"world_id"`
	Code        string                 `json:"code"`        // "jungles_in_cold"
	Severity    string                 `json:"severity"`    // "high" / "medium" / "low"
	Description string                 `json:"description"` // человекочитаемое
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ==================== РЕЗУЛЬТАТ ====================

// AuditResult — результат полного аудита галактики.
type AuditResult struct {
	// Итоги
	TotalPlanets      int   `json:"total_planets"`
	PlanetsWithIssues int   `json:"planets_with_issues"`
	TotalIssues       int   `json:"total_issues"`
	DurationMs        int64 `json:"duration_ms"`

	// Агрегаты по коду проблемы (для быстрого обзора)
	IssuesByCode map[string]int `json:"issues_by_code"`

	// Агрегаты по уровню критичности
	IssuesBySeverity map[string]int `json:"issues_by_severity"`

	// Примеры проблем (до SampleLimit)
	SampleIssues []Issue `json:"sample_issues"`

	// Был ли результат усечён (если проблем > SampleLimit)
	Truncated bool `json:"truncated"`
}

// SampleLimit — максимальное количество примеров в ответе.
const SampleLimit = 200

// ==================== ПРАВИЛА ====================

// Rule — одно правило проверки.
//
// Принимает "сырую" планету (map из JSON) и её ID/имя/world_id.
// Возвращает список найденных проблем (может быть пустым).
type Rule struct {
	Code     string
	Severity string
	Check    func(p *PlanetView) []Issue
}

// PlanetView — представление планеты для проверок.
// Заполняется один раз в начале аудита, чтобы не парсить JSON в каждой проверке.
type PlanetView struct {
	ID         string
	Name       string
	WorldID    string

	// Физика
	Size         float64
	Mass         float64
	Density      float64
	Temperature  float64
	WaterPercent float64

	// Типы и флаги
	Type            string
	SurfaceDominant string
	Climate         string
	Atmosphere      string
	Hydrosphere     string
	Biosphere       string

	IsGasGiant  bool
	IsRadioactive bool
	Habitable   bool
	Life        bool
	Population  int64

	// Композиции
	Surface    map[string]float64
	Subterrain map[string]float64

	// Ядро (nil если нет)
	Core *CoreView

	// Спутники
	Satellites []SatelliteView

	// Сырой JSON (на случай, если правило хочет свои поля)
	Raw map[string]interface{}
}

// CoreView — ядро планеты для проверок.
type CoreView struct {
	Type          string
	MassPercent   float64
	Activity      float64
	Radioactivity float64
	Age           float64
	IsActive      bool
	IsMetallic    bool
}

// SatelliteView — спутник газового гиганта для проверок.
type SatelliteView struct {
	ID           string
	Name         string
	Mass         float64
	Size         float64
	Temperature  float64
	WaterPercent float64
	Atmosphere   string
	Habitable    bool
	Life         bool
}