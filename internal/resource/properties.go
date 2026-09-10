// internal/resource/properties.go
package resource

// BaseProps — 9 числовых свойств ресурса (0–100).
// Используется как база для каждой категории. При генерации конкретного
// ресурса значения случайно сдвигаются на ±propJitter.
type BaseProps struct {
	Hardness         float64
	Elasticity       float64
	Conductivity     float64
	HeatResistance   float64
	ChemicalActivity float64
	Density          float64
	Biocompatibility float64
	EnergyDensity    float64
	Volatility       float64
}

// propJitter — максимальный сдвиг при генерации конкретного ресурса.
const propJitter = 10.0

// baseProperties — базовый набор свойств для каждой из 6 категорий.
//
// Значения подобраны так, чтобы:
//   - разные категории давали разные профили;
//   - идеальные профили товаров (см. GDD, раздел 11.2) могли быть
//     достигнуты смесью ресурсов из разных категорий;
//   - Топливо, Вода, Газы имели ярко выраженные «крайние» значения
//     (энергоёмкость, биосовместимость, летучесть).
var baseProperties = map[string]BaseProps{
	// 🪨 Минералы: твёрдые, плотные, средняя проводимость
	CategoryMineral: {
		Hardness:         70,
		Elasticity:       20,
		Conductivity:     50,
		HeatResistance:   65,
		ChemicalActivity: 30,
		Density:          65,
		Biocompatibility: 10,
		EnergyDensity:    25,
		Volatility:       10,
	},

	// 🌿 Органика: мягкие, эластичные, биосовместимые
	CategoryOrganic: {
		Hardness:         20,
		Elasticity:       70,
		Conductivity:     10,
		HeatResistance:   25,
		ChemicalActivity: 15,
		Density:          20,
		Biocompatibility: 80,
		EnergyDensity:    35,
		Volatility:       25,
	},

	// ⭐ Редкие: средние по всем свойствам, чуть выше проводимость
	CategoryRare: {
		Hardness:         60,
		Elasticity:       25,
		Conductivity:     60,
		HeatResistance:   55,
		ChemicalActivity: 25,
		Density:          50,
		Biocompatibility: 15,
		EnergyDensity:    35,
		Volatility:       15,
	},

	// 🔥 Топливо: летучее, энергоёмкое, хим. активное
	CategoryFuel: {
		Hardness:         10,
		Elasticity:       10,
		Conductivity:     10,
		HeatResistance:   30,
		ChemicalActivity: 60,
		Density:          50,
		Biocompatibility: 10,
		EnergyDensity:    95,
		Volatility:       80,
	},

	// 💧 Вода: жидкая, биосовместимая, хим. активная
	CategoryWater: {
		Hardness:         5,
		Elasticity:       10,
		Conductivity:     40,
		HeatResistance:   20,
		ChemicalActivity: 50,
		Density:          30,
		Biocompatibility: 95,
		EnergyDensity:    10,
		Volatility:       60,
	},

	// 💨 Газы: лёгкие, крайне летучие, энергоёмкие
	CategoryGas: {
		Hardness:         2,
		Elasticity:       5,
		Conductivity:     20,
		HeatResistance:   30,
		ChemicalActivity: 40,
		Density:          5,
		Biocompatibility: 20,
		EnergyDensity:    70,
		Volatility:       95,
	},
}

// GetBaseProps — базовые свойства для категории.
// Если категория неизвестна — возвращает средний профиль (все 50).
func GetBaseProps(category string) BaseProps {
	if p, ok := baseProperties[category]; ok {
		return p
	}
	return BaseProps{
		Hardness:         50,
		Elasticity:       50,
		Conductivity:     50,
		HeatResistance:   50,
		ChemicalActivity: 50,
		Density:          50,
		Biocompatibility: 50,
		EnergyDensity:    50,
		Volatility:       50,
	}
}