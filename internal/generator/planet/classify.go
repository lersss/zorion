// internal/generator/planet/classify.go
package planet

// ==================== ГЕЙМДИЗАЙНЕРСКИЕ ТИПЫ ====================

const (
	TypeGasGiant    = "газовый гигант"
	TypeRadioactive = "радиоактивная"
	TypeEarthlike   = "землеподобная"
	TypeOceanic     = "океаническая"
	TypeIce         = "ледяная"
	TypeVolcanic    = "вулканическая"
	TypeDesert      = "пустынная"
	TypeGlass       = "стеклянная"
	TypeMetal       = "металлическая"
	TypeOrganic     = "органик"
	TypeRocky       = "скалистая (базовая)"
)

// AllGameDesignTypes — все возможные типы (для UI, фильтров, статистики).
var AllGameDesignTypes = []string{
	TypeEarthlike,
	TypeOceanic,
	TypeIce,
	TypeVolcanic,
	TypeDesert,
	TypeGlass,
	TypeMetal,
	TypeOrganic,
	TypeRocky,
	TypeRadioactive,
	TypeGasGiant,
}

// classificationThreshold — минимальная доля формы, при которой она
// считается «присутствующей» для классификации.
const classificationThreshold = 15.0

// PlanetClassificationInput — входные данные для классификации.
type PlanetClassificationInput struct {
	IsGasGiant    bool
	IsRadioactive bool
	Surface       Composition
	Temperature   float64
	WaterPercent  float64
	Habitable     bool
	Life          bool
}

// ClassifyGameDesignType — определяет геймдизайнерский тип планеты
// на основе её свойств и композиции поверхности.
//
// Порядок проверок важен:
//  1. Сначала — специфичные флаги (газовый гигант, радиоактивная).
//  2. Затем — землеподобная (самый узкий набор условий).
//  3. Затем — остальные типы по убыванию специфичности.
//  4. Fallback — «скалистая (базовая)».
func ClassifyGameDesignType(in PlanetClassificationInput) string {
	// 1. Газовый гигант — самый специфичный случай
	if in.IsGasGiant {
		return TypeGasGiant
	}

	// 2. Радиоактивная
	if in.IsRadioactive {
		return TypeRadioactive
	}

	// 3. Землеподобная: обитаема + есть жизнь + и скалы, и вода
	if in.Habitable && in.Life &&
		in.Surface.Has(SurfaceRocks) &&
		(in.Surface.Has(SurfaceOceans) || in.Surface.Has(SurfaceLakes)) {
		return TypeEarthlike
	}

	// 4. Океаническая: доминируют океаны, много воды
	if in.Surface.DominantForm() == SurfaceOceans && in.WaterPercent > 60 {
		return TypeOceanic
	}

	// 5. Ледяная: доминируют ледники, или очень холодно
	if in.Surface.DominantForm() == SurfaceGlaciers || in.Temperature < 200 {
		return TypeIce
	}

	// 6. Вулканическая: заметная доля лавы или вулканических полей
	if in.Surface.ShareOf(SurfaceLavaFields)+in.Surface.ShareOf(SurfaceVolcanicFields) >= classificationThreshold {
		return TypeVolcanic
	}

	// 7. Пустынная: доминируют пески, мало воды
	if in.Surface.DominantForm() == SurfaceSands && in.WaterPercent < 20 {
		return TypeDesert
	}

	// 8. Стеклянная
	if in.Surface.ShareOf(SurfaceGlassFields) >= classificationThreshold {
		return TypeGlass
	}

	// 9. Металлическая
	if in.Surface.ShareOf(SurfaceMetalFields) >= classificationThreshold {
		return TypeMetal
	}

	// 10. Органик — доминирует биосферная форма
	if isBiosphereForm(in.Surface.DominantForm()) {
		return TypeOrganic
	}

	// 11. Fallback — «скалистая (базовая)»
	return TypeRocky
}

// ==================== ХЕЛПЕРЫ ====================

// isBiosphereForm — относится ли форма к биосферным (жизнь).
func isBiosphereForm(form string) bool {
	switch form {
	case SurfaceMeadows,
		SurfaceForests,
		SurfaceJungles,
		SurfaceSwamps,
		SurfaceCoralReefs:
		return true
	}
	return false
}

// GameDesignTypeName — читаемое название типа (для UI).
// Сейчас совпадает с кодом, но оставляем на случай переименования.
func GameDesignTypeName(code string) string {
	return code
}