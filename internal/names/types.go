// internal/names/types.go
package names

// LocalizedName — имя сущности в двух вариантах:
//   - Cyr — кириллица (для UI, логов, отображения игроку);
//   - Lat — латиница (для URL, экспорта, технических идентификаторов).
//
// Оба варианта генерируются из одной пары слогов, поэтому всегда
// соответствуют друг другу (Атрокс ↔ Atrox).
type LocalizedName struct {
	Cyr string `json:"name"`     // кириллица
	Lat string `json:"name_lat"` // латиница
}

// String — возвращает кириллический вариант как основной.
// Удобно для логов, ошибок, отладки.
func (n LocalizedName) String() string {
	return n.Cyr
}

// IsEmpty — есть ли вообще имя
func (n LocalizedName) IsEmpty() bool {
	return n.Cyr == "" && n.Lat == ""
}