package text

import "unicode"

var scriptThreshold = map[string]int{"latin": 60, "han": 10, "kana": 10, "hangul": 10}

const sentence = 20

var scriptTables = []struct {
	name  string
	table *unicode.RangeTable
}{
	{"latin", unicode.Latin}, {"cyrillic", unicode.Cyrillic}, {"greek", unicode.Greek},
	{"arabic", unicode.Arabic}, {"hebrew", unicode.Hebrew}, {"devanagari", unicode.Devanagari},
	{"kana", unicode.Hiragana}, {"kana", unicode.Katakana}, {"hangul", unicode.Hangul}, {"han", unicode.Han},
}

var lettersOutsideTheirScriptName = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x00AA, 0x00BA, 0x10}, {0x02B0, 0x02B8, 1}, {0x02E0, 0x02E4, 1}, {0x1D2C, 0x1D61, 1},
		{0x1D78, 0x1D78, 1}, {0x1D9B, 0x1DBF, 1}, {0x2071, 0x2071, 1}, {0x207F, 0x207F, 1},
		{0x2126, 0x2126, 1}, {0x212A, 0x212B, 1}, {0x2132, 0x2132, 1}, {0x214E, 0x214E, 1},
		{0x2183, 0x2183, 1}, {0x2C7D, 0x2C7D, 1}, {0x3005, 0x3005, 1}, {0x303B, 0x303B, 1},
		{0xA69C, 0xA69D, 1}, {0xA770, 0xA770, 1}, {0xA7F2, 0xA7F4, 1}, {0xA7F8, 0xA7F9, 1},
		{0xAB5C, 0xAB5F, 1}, {0xAB69, 0xAB69, 1}, {0xFF21, 0xFF3A, 1}, {0xFF41, 0xFF5A, 1},
		{0xFF66, 0xFF6F, 1}, {0xFF71, 0xFF9D, 1}, {0xFFA0, 0xFFBE, 1}, {0xFFC2, 0xFFC7, 1},
		{0xFFCA, 0xFFCF, 1}, {0xFFD2, 0xFFD7, 1}, {0xFFDA, 0xFFDC, 1},
	},
	R32: []unicode.Range32{
		{0x10780, 0x10785, 1}, {0x10787, 0x107B0, 1}, {0x107B2, 0x107BA, 1}, {0x16FE3, 0x16FE3, 1},
		{0x1B002, 0x1B11E, 1}, {0x1E030, 0x1E050, 1}, {0x1E06B, 0x1E06D, 1},
	},
}

var commonLettersNamedForScript = map[rune]string{0x0374: "greek", 0x0640: "arabic", 0x30FC: "kana"}

func scriptOf(r rune) string {
	if !unicode.IsLetter(r) {
		return ""
	}
	if s, ok := commonLettersNamedForScript[r]; ok {
		return s
	}
	if unicode.Is(lettersOutsideTheirScriptName, r) {
		return ""
	}
	for _, t := range scriptTables {
		if unicode.Is(t.table, r) {
			return t.name
		}
	}
	return ""
}

// ScriptCount is how many letters of one script a text holds.
type ScriptCount struct {
	// Name is the script's name in its trait ID, such as "cyrillic", or "kana" for Hiragana and
	// Katakana alike.
	Name string
	// Count is the number of its letters.
	Count int
}

// Scripts counts each script's letters in text and returns the trait:script IDs the counts place,
// with the counts in the order each script first appears. A script places a trait at 20 letters,
// Latin at 60, and Han, kana, and Hangul at 10; Latin places none of its own, and two writing
// systems, Han with kana counting as one, place trait:script.mixed.
func Scripts(text string) (map[string]bool, []ScriptCount) {
	var counts []ScriptCount
	index := map[string]int{}
	for _, r := range text {
		s := scriptOf(r)
		if s == "" {
			continue
		}
		i, ok := index[s]
		if !ok {
			i = len(counts)
			index[s] = i
			counts = append(counts, ScriptCount{s, 0})
		}
		counts[i].Count++
	}
	found := map[string]bool{}
	for _, c := range counts {
		threshold, ok := scriptThreshold[c.Name]
		if !ok {
			threshold = sentence
		}
		if c.Count >= threshold {
			found[c.Name] = true
		}
	}
	traits := map[string]bool{}
	systems := map[string]bool{}
	for s := range found {
		if s != "latin" {
			traits["trait:script."+s] = true
		}
		if (s == "han" || s == "kana") && found["kana"] {
			systems["japanese"] = true
		} else {
			systems[s] = true
		}
	}
	if len(systems) >= 2 {
		traits["trait:script.mixed"] = true
	}
	return traits, counts
}
