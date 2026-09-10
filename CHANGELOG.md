# Changelog

Все значимые изменения проекта Zorion.
Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/),
проект следует [Semantic Versioning](https://semver.org/lang/ru/).

---

## [Unreleased]

### Добавлено

**Описания планет (фазы 1+2)**
- **Фаза 1 (скелет):** 8 типов × (30 openings + 30 closings) = 480 текстов:
  `volcanic`, `desert`, `earthlike`, `radioactive`, `organic`, `glass`, `metal`, `rocky`.
- **Ранее:** `gas_giant`, `icy`, `oceanic` — по 150+150 (уже в репозитории).
- **Итого:** 11 типов, ~705 зачинов, ~690 концовок.
- **Фаза 2 (Go-код):** новые файлы в `internal/generator/planet/`:
  - `descriptions_types.go` — структуры (`Opening`, `Closing`, `DescriptionContext`, `descriptionsManager`).
  - `descriptions_tags.go` — вычисление тегов из контекста планеты.
  - `descriptions_load.go` — автопоиск JSON-файлов при старте, загрузка в память.
  - `descriptions_pick.go` — фильтрация по тегам и выбор по хэшу (FNV-1a).
  - `descriptions_manager.go` — точка входа `GenerateDescription`, склейка, fallback.
  - `descriptions_mapping.go` — маппинг «геймдизайнерский тип → папка» (`TypeIce` → `icy`).
- Автопоиск: сервер сам сканирует `config/descriptions/<type>/`, подхватывает
  все `openings_*.json` и `closings_*.json`. Новые файлы — через `git push`, без правок кода.
- Детерминированный выбор по хэшу от `planet.id`. Fallback — 7 нейтральных вариантов.
- Интеграция: `planet_data_generate.go`, `planet_data_gas.go`, `main.go`.
- **Детали:** `docs/DESCRIPTIONS_WORK.md`.

**Серверная кластеризация карты**
- `filter_worlds_handler.go` переписан под SQL `GROUP BY` по ячейкам сетки.
- Клиент присылает `x_min/x_max/y_min/y_max/cell`, сервер отдаёт 100–5000 кластеров
  вместо 100 000 миров. Gzip-сжатие ответа.
- Фронт: `map_render.js` рисует кластеры (кружки с числами), `data.js` грузит
  их с debounce 180 мс, `events.js` — pan/zoom/click, `main.js`, `animation.js`, `navigation.js`.
- `minZoom: 0.02 → 0.001` — можно отдалить до всей галактики.
- **Результат:** карта на 100k миров отвечает за < 100 мс.

**Миграции БД**
- `000010_assignments_type_reward.up.sql` / `.down.sql` — добавить колонки
  `type` и `reward` в `assignments` (используются кодом, но отсутствовали в схеме).

**Безопасность**
- JWT-секрет читается из env `JWT_SECRET` через `auth.InitJWTSecret` в `main.go`.
  Валидация длины ≥ 32 байта. Без секрета сервер не стартует.

### Изменено

**Безопасность**
- `internal/auth/jwt.go` — секрет больше не хардкод, живёт под `RWMutex`.
- `internal/auth/jwt.go` — в `VerifyToken` добавлена проверка алгоритма
  (защита от подмены `alg: none`).
- `internal/config/config.go` — `JWT_SECRET` обязателен, `log.Fatal` при пустом.
- `cmd/server/main.go` — вызов `auth.InitJWTSecret` в начале `main()`.

**Классификатор планет (`classify.go`)**
- **Ледяная:** `T < 150` теперь требует `ShareOf(Glaciers) >= 30`. Раньше любая
  холодная планета становилась ледяной (перекос 28%).
- **Органик:** порог по сумме биосферных форм `>= 25%` вместо требования доминанты
  (тип был почти недостижим: 0,1%).
- **Стекло/металл:** пороги `15 → 10` (эти формы в архетипах редко доходят до 15%).

**Карта миров**
- `filter_worlds_handler.go` — облегчённые поля (убраны `created_at`, `updated_at`),
  gzip, таймеры в лог, `QueryContext`, фикс `superfluous WriteHeader`.
- `GetWorld` (`world_handlers.go`) — устойчивость к падениям locations/assignments:
  если вспомогательный запрос падает, ответ всё равно уходит. `sql.ErrNoRows` → 404, не 500.

**Описания планет**
- `planet_data_generate.go` — UUID генерится раньше (нужен для выбора описания),
  все три генератора заполняют `DescriptionContext`.
- `planet_data_gas.go` — газовый гигант тоже получает описание из библиотеки.
- Фикс бага с ключом `surface:` в `generateRadioactivePlanet` — заменено
  на явную переменную `dominantSurface`.

### Исправлено

- **JWT-секрет — хардкод `your-secret-key`** → env. Критично, было в публичном репозитории.
- **`GetWorld` 500 при несуществующих locations/assignments** → устойчивость.
- **`column "type" does not exist`** в `assignmentRepo.GetByWorld` → миграция `000010`.
- **`concurrent map writes`** — не в этой сессии, но фикс в `planet_image.go`
  (`sync.Mutex` на `cache`/`cacheOrder`/`rand`).
- **`superfluous response.WriteHeader call`** в `filter_worlds_handler.go` —
  убран `http.Error` после начала записи ответа.

### Удалено

- `internal/generator/planet/planet_data_description.go` — старая `generateDescription`
  больше не вызывается. **Восстановлен** после ошибочного удаления: файл содержит
  функцию `clamp`, нужную `physics.go`. `generateDescription` в нём — мёртвый код,
  пусть лежит.

### В планах

**Аномалии как контент** (следующий крупный блок)
- Библиотека готова (22 JSON-файла в `config/anomalies/`).
- Осталось: Go-код чтения при старте, API, детекция на фронте, показ в карточке планеты.

**LLM-генерация описаний**
- Groq / YandexGPT при создании планеты, сохранение в БД.
- Fallback — библиотека `config/descriptions/` + библиотека аномалий.

**Тонкая настройка генерации**
- Все хардкод-константы в конфиг, редактируемый через админку.

**Расширение описаний**
- Добить до 150+150 популярные типы: `desert`, `volcanic`, `earthlike`.
- Ревизия `gas_giant`/`icy` (600 текстов) под актуальные теги.

**Безопасность и инфраструктура**
- `ADMIN_PASSWORD` — сейчас дефолт `admin123`, нужен обязательный из env.
- Обработка `unique_violation` (23505) в `Register` — 409 вместо 500.
- Вкладка «Совместимость» в админке (frontend для готового backend).
- `compatibility_matrix` — миграция `009` не применена, таблицы нет.
- Модель энергии: планетарный рынок, микроконтракты.
- Миграция старых категорий ресурсов (`energy` → `fuel`).
- Рефакторинг остальных генераторов имён на `LocalizedName`.
- CI с `go vet`, `go test`, `go vet -race`.

---

## [0.4.0] — 2026-09-10

Обновление: **аудит планет**, **карточка планеты в UI**,
**фикс гонок конкурентности**, **балансировка биосферы в холоде**.

### Добавлено

**Аудит планет**
- Пакет `internal/audit` с дженерик-движком `Run[T]` — расширяемо на другие сущности.
- Подпакет `internal/audit/planet` — 43 правила проверки планет:
  - физика (T, M/R/ρ, суммы композиций, отрицательные проценты);
  - композиция vs T (джунгли/леса/луга/болота/рифы/лёд/лава/океаны);
  - композиция vs вода (океаны, биосфера);
  - атмосфера vs T (парник, кислород, метан, водород);
  - ядро (диапазоны, возраст, флаг `is_metallic`);
  - спутники (температура, масса);
  - жизнь и обитаемость;
  - `type` ↔ `surface_dominant`;
  - мусор в данных (NaN, Inf, нули, пустое имя).
- HTTP-хендлер `GET /admin/audit`.
- Английские коды проблем (`jungles_in_cold`, `lava_in_cold`, ...).
- Защита от паник в правилах (`safeCheck`).

**Вкладка «🔍 Аудит» в админке**
- Сводка (4 плитки): всего планет, с проблемами, всего проблем, время.
- Таблица кодов проблем (код, количество, severity).
- Таблица примеров (планета, код, описание).
- Кнопка «показать все» (50 → 200).

**Карточка планеты (UI)**
- Модалка разбита: `tabs.js` + `panel.js` + `index.js`.
- Новые поля: масса, размер, плотность, температура, ядро (тип, доля, активность, радиоактивность, возраст), композиция поверхности, композиция недр, спутники газовых гигантов.
- Цветная полоска + иконки для композиции поверхности и недр.
- Обрезка топ-7 форм с кнопкой «показать все».
- Список спутников у газовых гигантов.

**Редирект на логин**
- Автоматический редирект на `/login-page` при 401/403 (в `main.js`).
- Удаление токена из `localStorage` при 401.

**LLM-генерация (обсуждение в GDD)**
- Решено: генерировать описания планет и аномалий **при создании** планеты и сохранять в БД.
- Планируемые API: Groq (free tier) или YandexGPT.

### Изменено

- **`PlanetGenerator`** в `planet_image.go` защищён `sync.Mutex` (гонка `cache`/`cacheOrder`/`rand`).
- **`WebSocketHub`** переведён на `wsClient` с per-connection мьютексом (gorilla/websocket не потокобезопасна для записи).
- **`TravelManager`** — старый полёт не удаляет новый (проверка `current == flight`).
- **`StatusManager`** — новый метод `TryStart` (атомарная проверка + запуск).
- **`admin_universe.go`** — `TryStart` вместо `if + Start`, безопасный `recoverErr`, `ClearUniverse` блокируется при активной генерации.
- **`composition_modifiers.go`** — биосферные формы при `T < 250` умножаются на ×0.05 (было ×0.2–0.3). Абсолютные запреты: лава при `T < 500`, лёд при `T > 320`.
- **`computeSatelliteTemp`** — приливный нагрев снижен с 400 до 100, добавлен потолок `giantTemp + 50`.
- **`checks_physics.go`** — severity биосферных форм в холоде `high` → `low`; газовые гиганты исключены из `hydrogen_in_heat`.
- **`models.Planet`** расширена: `Density`, `Core`, `SurfaceComposition`, `SubterrainComposition`, `Satellites`, `Climate`, `SystemAge`, `Hydrosphere`, `Biosphere`, `Radioactive`.
- **`planet_repo.go`** — парсинг ~20 полей (было 9).

### Исправлено

- **`concurrent map writes`** в `PlanetGenerator` (паника всего сервера при двух одновременных запросах картинок).
- **Concurrent write** в `WebSocketHub` (запись в одно соединение из двух горутин).
- **Race** в `TravelManager` (старый полёт удалял новый).
- **Race** в `StatusManager` (проверка + старт были неатомарны).
- **Паника от `r.(string)`** в `recover` (падало на `runtime.Error`).
- **Джунгли/леса/луга/болота** при `T 220–250 K` — теперь ×0.05, редко (2% планет).
- **`satellite_hotter_than_giant`** — 584 случая → ~0.
- **`biosphere_without_water`** — 1089 случаев → ~0.

### Удалено

- Дублирующие файлы `internal/audit/parser.go`, `auditor.go`, `checks_physics.go`, `checks_consistency.go` (переехали в `internal/audit/planet/`).

---

## [0.3.0] — 2026-09-10

Обновление: **ядро планеты**, **плотность и размер через массу**,
**правильная формула температуры**, **реалистичные спектры**.

### Добавлено

- **Ядро планеты** (`core.go`): тип, доля массы, активность, радиоактивность, возраст.
- **Физика температуры** (`physics.go`): `T_eq = 278.7 × L^0.25 / sqrt(r)`, альбедо, парниковый эффект, вклад ядра.
- **Плотность и размер**: `R = (M/ρ)^(1/3)`.
- **Спектральные классы**: взвешенное распределение (O 0.5% → M 32%).
- **Классификация планет** по композиции (`classify.go`).
- **Статистика** с распределениями по формам, недрам, ядрам.

### Изменено

- Архетипы: `size_min/max` → `mass_min/max`.
- JSON планеты: добавлены `density`, `core`, `system_age`.
- Атмосферы расширены до 14 типов.

---

## [0.2.0] — 2026-09-10

Обновление: **композиция поверхности и недр**, **6 категорий ресурсов**,
**спутники газовых гигантов**, **билингвальные имена**.

### Добавлено

- 16 форм поверхности, 17 типов недр.
- Матрица совместимости (backend + кеш).
- Спутники газовых гигантов (3–10 штук).
- 6 категорий ресурсов, маппинг «форма → категория».
- Тип `LocalizedName` (кириллица + латиница).

---

## [0.1.0] — 2026-09-09

Первый MVP-релиз: генерация вселенной, планеты, модальное окно, текстуры, админка.

---

[Unreleased]: https://github.com/lersss/zorion/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/lersss/zorion/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/lersss/zorion/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/lersss/zorion/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/lersss/zorion/releases/tag/v0.1.0