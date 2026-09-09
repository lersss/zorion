package resource

import (
	"math/rand"

	"github.com/google/uuid"
	"zorion/internal/models"
	"zorion/internal/names"
)

// GenerateResources создаёт ресурсы для планеты
func GenerateResources(planetID, planetType, spectralClass string, rng *rand.Rand) []*models.PlanetResource {
	totalResources := determineTotal(planetType)
	if totalResources == 0 {
		return nil
	}

	knownCount := totalResources / 2
	if knownCount < 1 {
		knownCount = 1
	}
	unknownCount := totalResources - knownCount

	categories := weightedCategories(planetType, spectralClass)

	existingNames := make(map[string]bool)
	var resources []*models.PlanetResource

	for i := 0; i < knownCount; i++ {
		cat := pickCategory(categories, rng)
		res := generateResource(planetID, cat, true, rng, existingNames)
		if res != nil {
			resources = append(resources, res)
		}
	}

	for i := 0; i < unknownCount; i++ {
		cat := pickCategory(categories, rng)
		res := generateResource(planetID, cat, false, rng, existingNames)
		if res != nil {
			resources = append(resources, res)
		}
	}

	return resources
}

func determineTotal(planetType string) int {
	switch planetType {
	case "землеподобная", "терраформированная":
		return 10 + rand.Intn(5)
	case "океаническая":
		return 8 + rand.Intn(5)
	case "пустынная", "ледяная", "вулканическая":
		return 6 + rand.Intn(5)
	case "газовый_гигант":
		return rand.Intn(2)
	default:
		return 6 + rand.Intn(5)
	}
}

func weightedCategories(planetType, spectralClass string) map[string]int {
	base := map[string]int{
		"mineral": 25,
		"organic": 25,
		"energy":  25,
		"rare":    25,
	}
	switch planetType {
	case "землеподобная":
		base["organic"] += 10
		base["mineral"] += 5
		base["rare"] -= 5
	case "океаническая":
		base["organic"] += 15
		base["mineral"] -= 5
		base["energy"] += 5
		base["rare"] -= 5
	case "пустынная":
		base["mineral"] += 10
		base["organic"] -= 10
		base["energy"] += 5
		base["rare"] -= 5
	case "ледяная":
		base["organic"] += 5
		base["mineral"] -= 5
		base["energy"] += 5
		base["rare"] += 5
	case "вулканическая":
		base["mineral"] += 15
		base["organic"] -= 10
		base["energy"] += 10
		base["rare"] += 5
	case "газовый_гигант":
		base["energy"] += 20
		base["mineral"] -= 10
		base["organic"] -= 10
		base["rare"] += 5
	}
	switch spectralClass {
	case "O", "B", "A":
		base["mineral"] += 10
		base["rare"] += 10
		base["organic"] -= 10
		base["energy"] -= 5
	case "K", "M":
		base["organic"] += 10
		base["energy"] += 10
		base["mineral"] -= 10
		base["rare"] -= 5
	}
	for k := range base {
		if base[k] < 0 {
			base[k] = 0
		}
	}
	return base
}

func pickCategory(categories map[string]int, rng *rand.Rand) string {
	total := 0
	for _, v := range categories {
		total += v
	}
	if total == 0 {
		return "mineral"
	}
	r := rng.Intn(total)
	for cat, weight := range categories {
		r -= weight
		if r < 0 {
			return cat
		}
	}
	return "mineral"
}

func generateResource(planetID, category string, known bool, rng *rand.Rand, existingNames map[string]bool) *models.PlanetResource {
	name := names.GenerateResourceName(category, rng, existingNames)
	baseProps := baseProperties(category)
	return &models.PlanetResource{
		ID:                uuid.New().String(),
		PlanetID:          planetID,
		Name:              name,
		Category:          category,
		Hardness:          clamp(baseProps.Hardness+(rng.Float64()-0.5)*20, 0, 100),
		Elasticity:        clamp(baseProps.Elasticity+(rng.Float64()-0.5)*20, 0, 100),
		Conductivity:      clamp(baseProps.Conductivity+(rng.Float64()-0.5)*20, 0, 100),
		HeatResistance:    clamp(baseProps.HeatResistance+(rng.Float64()-0.5)*20, 0, 100),
		ChemicalActivity:  clamp(baseProps.ChemicalActivity+(rng.Float64()-0.5)*20, 0, 100),
		Density:           clamp(baseProps.Density+(rng.Float64()-0.5)*20, 0, 100),
		Biocompatibility:  clamp(baseProps.Biocompatibility+(rng.Float64()-0.5)*20, 0, 100),
		EnergyDensity:     clamp(baseProps.EnergyDensity+(rng.Float64()-0.5)*20, 0, 100),
		Volatility:        clamp(baseProps.Volatility+(rng.Float64()-0.5)*20, 0, 100),
		Quantity:          100 + rng.Intn(401),
		IsKnown:           known,
	}
}

type prop struct {
	Hardness          float64
	Elasticity        float64
	Conductivity      float64
	HeatResistance    float64
	ChemicalActivity  float64
	Density           float64
	Biocompatibility  float64
	EnergyDensity     float64
	Volatility        float64
}

func baseProperties(category string) prop {
	switch category {
	case "mineral":
		return prop{70, 20, 50, 65, 30, 65, 10, 25, 10}
	case "organic":
		return prop{20, 70, 10, 25, 15, 20, 80, 35, 25}
	case "energy":
		return prop{10, 20, 5, 40, 20, 15, 5, 80, 60}
	case "rare":
		return prop{60, 25, 60, 55, 25, 50, 15, 35, 15}
	default:
		return prop{50, 50, 50, 50, 50, 50, 50, 50, 50}
	}
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}