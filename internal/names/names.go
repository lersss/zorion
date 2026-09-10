package names

import (
	"math/rand"
)

// GenerateStarName генерирует название для звезды
func GenerateStarName(rng *rand.Rand, usedNames map[string]bool) string {
	prefixes := []string{
		"Al", "Ar", "Bel", "Ben", "Cal", "Cel", "Dal", "Den", "El", "En",
		"Fal", "Fin", "Gal", "Gor", "Hal", "Hel", "In", "Is", "Jal", "Jen",
		"Kal", "Kel", "Lan", "Lor", "Mal", "Mar", "Nan", "Nel", "Ol", "Or",
		"Pal", "Pel", "Qu", "Quel", "Ral", "Ren", "Sal", "Sel", "Tal", "Ten",
		"Ul", "Ur", "Val", "Vel", "Wal", "Wel", "Xal", "Xen", "Yal", "Yen",
		"Zal", "Zen", "Ael", "Aer", "Bael", "Bri", "Cer", "Cor", "Dae", "Dor",
		"Astr", "Bore", "Cael", "Drac", "Elys", "Fen", "Gla", "Heli", "Ion", "Jov",
		"Kry", "Lyr", "Mira", "Neb", "Orion", "Peg", "Phae", "Qui", "Reg", "Sir",
		"Tau", "Urs", "Vega", "Xan", "Yor", "Zeph",
		"Aquil", "Ara", "Aur", "Boot", "Caelum", "Camel", "Canc", "Capr", "Carr", "Cassi",
		"Cen", "Ceph", "Cetus", "Col", "Com", "Corv", "Crat", "Cru", "Cyg", "Del",
		"Dor", "Dra", "Equ", "Eri", "For", "Gem", "Grus", "Her", "Hor", "Hyd",
		"Ind", "Lac", "Leo", "Lep", "Lib", "Lup", "Lyn", "Lyr", "Mic", "Mon",
		"Mus", "Nor", "Oct", "Oph", "Ori", "Pav", "Phe", "Pic", "Pis", "Pup",
		"Pyx", "Ret", "Sag", "Sco", "Scul", "Ser", "Sex", "Sge", "Sgr", "Tau",
		"Tel", "Tri", "Tuc", "UMa", "UMi", "Vel", "Vir", "Vol", "Vul", "Xer",
	}
	roots := []string{
		"ar", "en", "in", "on", "or", "um", "us", "is", "os", "al",
		"an", "ir", "ur", "yn", "el", "am", "ed", "id", "ul", "em",
		"ab", "ac", "ad", "ag", "ak", "ap", "as", "at", "ax", "az",
		"eb", "ec", "ef", "eg", "ek", "ep", "es", "et", "ex", "ez",
		"ib", "ic", "if", "ig", "ik", "ip", "is", "it", "ix", "iz",
		"ob", "oc", "of", "og", "ok", "op", "os", "ot", "ox", "oz",
		"ub", "uc", "uf", "ug", "uk", "up", "us", "ut", "ux", "uz",
	}
	suffixes := []string{
		"ia", "ar", "on", "is", "us", "os", "um", "a", "e", "i", "o", "y",
		"en", "or", "an", "ir", "yn", "el", "am", "ed", "id", "ul", "em",
		"ae", "ai", "ao", "au", "ei", "eu", "ie", "io", "iu", "oe",
		"oi", "ou", "ua", "ue", "ui", "uo", "ya", "ye", "yi", "yo",
	}

	for attempt := 0; attempt < 50; attempt++ {
		prefix := prefixes[rng.Intn(len(prefixes))]
		root := roots[rng.Intn(len(roots))]
		suf := suffixes[rng.Intn(len(suffixes))]
		name := prefix + root + suf
		if rng.Float64() < 0.3 {
			name = prefix + suf
		}
		if !usedNames[name] {
			usedNames[name] = true
			return name
		}
	}
	return "Star-" + randomSuffix(rng)
}

// GeneratePlanetName генерирует название для планеты
func GeneratePlanetName(rng *rand.Rand, usedNames map[string]bool) string {
	prefixes := []string{
		"Al", "Ar", "Bel", "Ben", "Cal", "Cel", "Dal", "Den", "El", "En",
		"Fal", "Fin", "Gal", "Gor", "Hal", "Hel", "In", "Is", "Jal", "Jen",
		"Kal", "Kel", "Lan", "Lor", "Mal", "Mar", "Nan", "Nel", "Ol", "Or",
		"Pal", "Pel", "Ral", "Ren", "Sal", "Sel", "Tal", "Ten", "Ul", "Ur",
		"Val", "Vel", "Wal", "Wel", "Xal", "Xen", "Yal", "Yen", "Zal", "Zen",
		"Aer", "Astr", "Bore", "Cael", "Drac", "Elys", "Fen", "Gla", "Heli", "Ion",
		"Jov", "Kry", "Lyr", "Mira", "Neb", "Orion", "Peg", "Phae", "Qui", "Reg",
		"Sir", "Tau", "Urs", "Vega", "Xan", "Yor", "Zeph",
		"Aquil", "Ara", "Aur", "Boot", "Caelum", "Camel", "Canc", "Capr", "Carr", "Cassi",
		"Cen", "Ceph", "Cetus", "Col", "Com", "Corv", "Crat", "Cru", "Cyg", "Del",
		"Dor", "Dra", "Equ", "Eri", "For", "Gem", "Grus", "Her", "Hor", "Hyd",
		"Ind", "Lac", "Leo", "Lep", "Lib", "Lup", "Lyn", "Lyr", "Mic", "Mon",
		"Mus", "Nor", "Oct", "Oph", "Ori", "Pav", "Phe", "Pic", "Pis", "Pup",
		"Pyx", "Ret", "Sag", "Sco", "Scul", "Ser", "Sex", "Sge", "Sgr", "Tau",
		"Tel", "Tri", "Tuc", "UMa", "UMi", "Vel", "Vir", "Vol", "Vul", "Xer",
	}
	suffixes := []string{
		"ia", "ar", "on", "is", "us", "os", "um", "or", "an", "il",
		"en", "ir", "yn", "el", "am", "ed", "id", "ul", "em", "a", "e", "i", "o", "y",
		"ae", "ai", "ao", "au", "ei", "eu", "ie", "io", "iu", "oe",
		"oi", "ou", "ua", "ue", "ui", "uo", "ya", "ye", "yi", "yo",
		"as", "es", "is", "os", "us", "ys", "um", "ym", "on", "yn",
	}

	for attempt := 0; attempt < 50; attempt++ {
		prefix := prefixes[rng.Intn(len(prefixes))]
		suffix := suffixes[rng.Intn(len(suffixes))]
		name := prefix + suffix
		if !usedNames[name] {
			usedNames[name] = true
			return name
		}
	}
	return "Planet-" + randomSuffix(rng)
}

// GenerateFactionName генерирует название для фракции
func GenerateFactionName(rng *rand.Rand, usedNames map[string]bool) string {
	prefixes := []string{
		"Al", "Ar", "Bel", "Ben", "Cal", "Cel", "Dal", "Den", "El", "En",
		"Fal", "Fin", "Gal", "Gor", "Hal", "Hel", "In", "Is", "Jal", "Jen",
		"Kal", "Kel", "Lan", "Lor", "Mal", "Mar", "Nan", "Nel", "Ol", "Or",
		"Pal", "Pel", "Qu", "Quel", "Ral", "Ren", "Sal", "Sel", "Tal", "Ten",
		"Ul", "Ur", "Val", "Vel", "Wal", "Wel", "Xal", "Xen", "Yal", "Yen",
		"Zal", "Zen", "Ael", "Aer", "Bael", "Bri", "Cer", "Cor", "Dae", "Dor",
		"Astr", "Bore", "Cael", "Drac", "Elys", "Fen", "Gla", "Heli", "Ion", "Jov",
		"Kry", "Lyr", "Mira", "Neb", "Orion", "Peg", "Phae", "Qui", "Reg", "Sir",
		"Tau", "Urs", "Vega", "Xan", "Yor", "Zeph",
		"Aquil", "Ara", "Aur", "Boot", "Caelum", "Camel", "Canc", "Capr", "Carr", "Cassi",
		"Cen", "Ceph", "Cetus", "Col", "Com", "Corv", "Crat", "Cru", "Cyg", "Del",
		"Dor", "Dra", "Equ", "Eri", "For", "Gem", "Grus", "Her", "Hor", "Hyd",
		"Ind", "Lac", "Leo", "Lep", "Lib", "Lup", "Lyn", "Lyr", "Mic", "Mon",
		"Mus", "Nor", "Oct", "Oph", "Ori", "Pav", "Phe", "Pic", "Pis", "Pup",
		"Pyx", "Ret", "Sag", "Sco", "Scul", "Ser", "Sex", "Sge", "Sgr", "Tau",
		"Tel", "Tri", "Tuc", "UMa", "UMi", "Vel", "Vir", "Vol", "Vul", "Xer",
	}
	suffixes := []string{
		"ia", "um", "or", "is", "an", "os", "us", "a", "e", "o", "i",
		"ae", "on", "ar", "en", "ir", "yn", "el",
		"io", "eo", "ua", "ea", "oi", "ou", "ie", "ei", "au", "ai",
		"ys", "em", "id", "ul", "am", "ed", "yn", "el", "ir", "en",
	}

	for attempt := 0; attempt < 50; attempt++ {
		prefix := prefixes[rng.Intn(len(prefixes))]
		suffix := suffixes[rng.Intn(len(suffixes))]
		name := prefix + suffix
		if !usedNames[name] {
			usedNames[name] = true
			return name
		}
	}
	return "Faction-" + randomSuffix(rng)
}

// GenerateProductName генерирует абстрактное название для товара
func GenerateProductName(category string, rng *rand.Rand, usedNames map[string]bool) string {
	roots := []string{
		"Atr", "Brol", "Vex", "Gron", "Drey", "Zorn", "Kvel", "Mrain", "Nox", "Prox",
		"Svar", "Trok", "Frin", "Hard", "Zvok", "Shrak", "Erd", "Alt", "Brim", "Vald",
		"Gart", "Dreyk", "Zelt", "Korn", "Lint", "Morn", "Norn", "Orn", "Parn", "Rork",
		"Sorn", "Thorn", "Urn", "Forn", "Horn", "Tsorn", "Shorn", "Yurn", "Yarn", "Ael",
		"Bern", "Virn", "Garn", "Dern", "Zern", "Kern", "Lern", "Mern", "Nern", "Pern",
		"Xyl", "Zin", "Cry", "Flo", "Glim", "Nyx", "Onyx", "Pyro", "Quar", "Rime",
		"Scor", "Temp", "Void", "Wisp", "Zeph", "Aura", "Blaz", "Chro", "Dusk", "Echo",
		"Astr", "Bore", "Cael", "Drac", "Elys", "Fen", "Gla", "Heli", "Ion", "Jov",
		"Kry", "Lyr", "Mira", "Neb", "Orion", "Peg", "Phae", "Qui", "Reg", "Sir",
		"Tau", "Urs", "Vega", "Xan", "Yor", "Zeph", "Aquil", "Ara", "Aur", "Boot",
		"Caelum", "Camel", "Canc", "Capr", "Carr", "Cassi", "Cen", "Ceph", "Cetus", "Col",
		"Com", "Corv", "Crat", "Cru", "Cyg", "Del", "Dor", "Dra", "Equ", "Eri",
		"For", "Gem", "Grus", "Her", "Hor", "Hyd", "Ind", "Lac", "Leo", "Lep",
		"Lib", "Lup", "Lyn", "Lyr", "Mic", "Mon", "Mus", "Nor", "Oct", "Oph",
		"Ori", "Pav", "Phe", "Pic", "Pis", "Pup", "Pyx", "Ret", "Sag", "Sco",
		"Scul", "Ser", "Sex", "Sge", "Sgr", "Tau", "Tel", "Tri", "Tuc", "UMa",
		"UMi", "Vel", "Vir", "Vol", "Vul", "Xer",
	}
	suffixes := map[string][]string{
		"food":         {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"clothing":     {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"construction": {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"electronics":  {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"weapons":      {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"medicine":     {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"tools":        {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"transport":    {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"fuel":         {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
		"furniture":    {"in", "an", "il", "on", "or", "ite", "um", "oz", "a", "ya", "ic", "ane", "ene", "ine", "one"},
	}
	sufList := suffixes[category]
	if len(sufList) == 0 {
		sufList = suffixes["food"]
	}

	for attempt := 0; attempt < 50; attempt++ {
		root := roots[rng.Intn(len(roots))]
		suf := sufList[rng.Intn(len(sufList))]
		name := root + suf
		if !usedNames[name] {
			usedNames[name] = true
			return name
		}
	}
	return "Product-" + randomSuffix(rng)
}

// randomSuffix — вспомогательная функция для создания уникального суффикса
func randomSuffix(rng *rand.Rand) string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 4)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}