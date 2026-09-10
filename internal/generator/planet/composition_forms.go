// internal/generator/planet/composition_forms.go
package planet

// ==================== ФОРМЫ ПОВЕРХНОСТИ (16) ====================

const (
	SurfaceRocks          = "скалы"
	SurfaceSands          = "пески_пустыни"
	SurfaceCraters        = "кратеры"
	SurfaceGlassFields    = "стеклянные_поля"
	SurfaceMetalFields    = "металлические_поля"
	SurfaceLavaFields     = "лавовые_поля"
	SurfaceVolcanicFields = "вулканические_поля"
	SurfaceGlaciers       = "ледники"
	SurfaceFrozenGases    = "мёрзлые_газы"
	SurfaceOceans         = "океаны"
	SurfaceLakes          = "озёра_реки"
	SurfaceMeadows        = "луга_степи"
	SurfaceForests        = "леса"
	SurfaceJungles        = "джунгли"
	SurfaceSwamps         = "болота"
	SurfaceCoralReefs     = "коралловые_рифы"
)

// AllSurfaceForms — все формы поверхности (для валидации, перебора, UI)
var AllSurfaceForms = []string{
	SurfaceRocks,
	SurfaceSands,
	SurfaceCraters,
	SurfaceGlassFields,
	SurfaceMetalFields,
	SurfaceLavaFields,
	SurfaceVolcanicFields,
	SurfaceGlaciers,
	SurfaceFrozenGases,
	SurfaceOceans,
	SurfaceLakes,
	SurfaceMeadows,
	SurfaceForests,
	SurfaceJungles,
	SurfaceSwamps,
	SurfaceCoralReefs,
}

// ==================== ТИПЫ НЕДР (17) ====================

const (
	SubterrainEmptyRock        = "пустая_порода"
	SubterrainMagmaticRocks    = "магматические_породы"
	SubterrainMetamorphicRocks = "метаморфические_породы"
	SubterrainSedimentaryRocks = "осадочные_породы"
	SubterrainOreVeins         = "рудные_жилы"
	SubterrainRareEarthVeins   = "редкоземельные_жилы"
	SubterrainRadioactiveZones = "радиоактивные_зоны"
	SubterrainCoalSeams        = "угольные_пласты"
	SubterrainOilPockets       = "нефтяные_карманы"
	SubterrainGasPockets       = "газовые_карманы"
	SubterrainGroundwater      = "подземные_воды"
	SubterrainGroundIce        = "подземные_льды"
	SubterrainMagmaChambers    = "магматические_камеры"
	SubterrainCrystalVeins     = "кристаллические_жилы"
	SubterrainSaltDomes        = "соляные_купола"
	SubterrainCaveSystems      = "пещерные_системы"
	SubterrainMetalCores       = "металлические_ядра"
)

// AllSubterrainTypes — все типы недр
var AllSubterrainTypes = []string{
	SubterrainEmptyRock,
	SubterrainMagmaticRocks,
	SubterrainMetamorphicRocks,
	SubterrainSedimentaryRocks,
	SubterrainOreVeins,
	SubterrainRareEarthVeins,
	SubterrainRadioactiveZones,
	SubterrainCoalSeams,
	SubterrainOilPockets,
	SubterrainGasPockets,
	SubterrainGroundwater,
	SubterrainGroundIce,
	SubterrainMagmaChambers,
	SubterrainCrystalVeins,
	SubterrainSaltDomes,
	SubterrainCaveSystems,
	SubterrainMetalCores,
}