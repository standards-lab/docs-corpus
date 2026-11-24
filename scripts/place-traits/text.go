// From CPython (v3.14.8), PSF-2.0: the codec aliases of Lib/encodings/aliases.py and normalize_encoding of Lib/encodings/__init__.py, ported to Go and reduced to the codecs golang.org/x/text provides.
//
// Copyright (c) 2001 Python Software Foundation; All Rights Reserved

package main

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	xunicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

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

type scriptCount struct {
	name  string
	count int
}

func scriptsIn(text string) (map[string]bool, []scriptCount) {
	var counts []scriptCount
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
			counts = append(counts, scriptCount{s, 0})
		}
		counts[i].count++
	}
	found := map[string]bool{}
	for _, c := range counts {
		threshold, ok := scriptThreshold[c.name]
		if !ok {
			threshold = sentence
		}
		if c.count >= threshold {
			found[c.name] = true
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

func isPySpace(r rune) bool {
	return unicode.IsSpace(r) || (r >= 0x1C && r <= 0x1F)
}

func pyNormalizeEncoding(name string) string {
	var b strings.Builder
	punct := false
	for _, c := range strings.ToLower(name) {
		if unicode.IsLetter(c) || unicode.IsNumber(c) || c == '.' {
			if punct && b.Len() > 0 {
				b.WriteByte('_')
			}
			if c < utf8.RuneSelf {
				b.WriteRune(c)
			}
			punct = false
		} else {
			punct = true
		}
	}
	return b.String()
}

const (
	codecASCII   = "ascii"
	codecUTF8    = "utf_8"
	codecUTF16   = "utf_16"
	codecUTF16LE = "utf_16_le"
	codecUTF16BE = "utf_16_be"
	codecUTF32   = "utf_32"
	codecUTF32LE = "utf_32_le"
	codecUTF32BE = "utf_32_be"
)

var pyCodecAliases = map[string]string{
	"646": codecASCII, "ansi_x3.4_1968": codecASCII, "ansi_x3_4_1968": codecASCII, "ascii": codecASCII,
	"cp367": codecASCII, "csascii": codecASCII, "ibm367": codecASCII, "iso646_us": codecASCII,
	"iso_646.irv_1991": codecASCII, "iso_ir_6": codecASCII, "us": codecASCII, "us_ascii": codecASCII,
	"u8": codecUTF8, "utf": codecUTF8, "utf8": codecUTF8, "utf8_ucs2": codecUTF8, "utf8_ucs4": codecUTF8,
	"utf_8": codecUTF8, "cp65001": codecUTF8,
	"u16": codecUTF16, "utf16": codecUTF16, "utf_16": codecUTF16, "unicodelittleunmarked": codecUTF16LE,
	"utf_16le": codecUTF16LE, "utf_16_le": codecUTF16LE, "unicodebigunmarked": codecUTF16BE,
	"utf_16be": codecUTF16BE, "utf_16_be": codecUTF16BE,
	"u32": codecUTF32, "utf32": codecUTF32, "utf_32": codecUTF32, "utf_32le": codecUTF32LE,
	"utf_32_le": codecUTF32LE, "utf_32be": codecUTF32BE, "utf_32_be": codecUTF32BE,
	"8859": "latin_1", "cp819": "latin_1", "csisolatin1": "latin_1", "ibm819": "latin_1",
	"iso8859": "latin_1", "iso8859_1": "latin_1", "iso_8859_1": "latin_1", "iso_8859_1_1987": "latin_1",
	"iso_ir_100": "latin_1", "l1": "latin_1", "latin": "latin_1", "latin1": "latin_1", "latin_1": "latin_1",
	"csisolatin2": "iso8859_2", "iso_8859_2": "iso8859_2", "iso_8859_2_1987": "iso8859_2", "iso_ir_101": "iso8859_2",
	"l2": "iso8859_2", "latin2": "iso8859_2",
	"csisolatin3": "iso8859_3", "iso_8859_3": "iso8859_3", "iso_8859_3_1988": "iso8859_3", "iso_ir_109": "iso8859_3",
	"l3": "iso8859_3", "latin3": "iso8859_3",
	"csisolatin4": "iso8859_4", "iso_8859_4": "iso8859_4", "iso_8859_4_1988": "iso8859_4", "iso_ir_110": "iso8859_4",
	"l4": "iso8859_4", "latin4": "iso8859_4",
	"csisolatincyrillic": "iso8859_5", "cyrillic": "iso8859_5", "iso_8859_5": "iso8859_5",
	"iso_8859_5_1988": "iso8859_5", "iso_ir_144": "iso8859_5",
	"arabic": "iso8859_6", "asmo_708": "iso8859_6", "csisolatinarabic": "iso8859_6", "ecma_114": "iso8859_6",
	"iso_8859_6": "iso8859_6", "iso_8859_6_1987": "iso8859_6", "iso_ir_127": "iso8859_6",
	"csisolatingreek": "iso8859_7", "ecma_118": "iso8859_7", "elot_928": "iso8859_7", "greek": "iso8859_7",
	"greek8": "iso8859_7", "iso_8859_7": "iso8859_7", "iso_8859_7_1987": "iso8859_7", "iso_ir_126": "iso8859_7",
	"csisolatinhebrew": "iso8859_8", "hebrew": "iso8859_8", "iso_8859_8": "iso8859_8",
	"iso_8859_8_1988": "iso8859_8", "iso_ir_138": "iso8859_8",
	"csisolatin5": "iso8859_9", "iso_8859_9": "iso8859_9", "iso_8859_9_1989": "iso8859_9", "iso_ir_148": "iso8859_9",
	"l5": "iso8859_9", "latin5": "iso8859_9",
	"csisolatin6": "iso8859_10", "iso_8859_10": "iso8859_10", "iso_8859_10_1992": "iso8859_10",
	"iso_ir_157": "iso8859_10", "l6": "iso8859_10", "latin6": "iso8859_10",
	"iso_8859_13": "iso8859_13", "l7": "iso8859_13", "latin7": "iso8859_13",
	"iso_8859_14": "iso8859_14", "iso_8859_14_1998": "iso8859_14", "iso_celtic": "iso8859_14",
	"iso_ir_199": "iso8859_14", "l8": "iso8859_14", "latin8": "iso8859_14",
	"iso_8859_15": "iso8859_15", "l9": "iso8859_15", "latin9": "iso8859_15",
	"iso_8859_16": "iso8859_16", "iso_8859_16_2001": "iso8859_16", "iso_ir_226": "iso8859_16",
	"l10": "iso8859_16", "latin10": "iso8859_16",
	"windows_1250": "cp1250", "windows_1251": "cp1251", "windows_1252": "cp1252", "windows_1253": "cp1253",
	"windows_1254": "cp1254", "windows_1255": "cp1255", "windows_1256": "cp1256", "windows_1257": "cp1257",
	"windows_1258": "cp1258",
	"437":          "cp437", "ibm437": "cp437", "850": "cp850", "ibm850": "cp850", "866": "cp866", "ibm866": "cp866",
	"cskoi8r": "koi8_r", "macroman": "mac_roman", "macintosh": "mac_roman", "maccyrillic": "mac_cyrillic",
	"chinese": "gb2312", "csiso58gb231280": "gb2312", "euc_cn": "gb2312", "euccn": "gb2312",
	"eucgb2312_cn": "gb2312", "gb2312_1980": "gb2312", "gb2312_80": "gb2312", "iso_ir_58": "gb2312",
	"936": "gbk", "cp936": "gbk", "ms936": "gbk", "gb18030_2000": "gb18030",
	"big5_tw": "big5", "csbig5": "big5",
	"eucjp": "euc_jp", "ujis": "euc_jp", "u_jis": "euc_jp",
	"csshiftjis": "shift_jis", "shiftjis": "shift_jis", "sjis": "shift_jis", "s_jis": "shift_jis",
	"csiso2022jp": "iso2022_jp", "iso2022jp": "iso2022_jp", "iso_2022_jp": "iso2022_jp",
	"euckr": "euc_kr", "korean": "euc_kr", "ksc5601": "euc_kr", "ks_c_5601": "euc_kr",
	"ks_c_5601_1987": "euc_kr", "ksx1001": "euc_kr", "ks_x_1001": "euc_kr",
	"949": "cp949", "ms949": "cp949", "uhc": "cp949",
}

var pyCodecs = map[string]encoding.Encoding{
	"latin_1": charmap.ISO8859_1, "iso8859_2": charmap.ISO8859_2, "iso8859_3": charmap.ISO8859_3,
	"iso8859_4": charmap.ISO8859_4, "iso8859_5": charmap.ISO8859_5, "iso8859_6": charmap.ISO8859_6,
	"iso8859_7": charmap.ISO8859_7, "iso8859_8": charmap.ISO8859_8, "iso8859_9": charmap.ISO8859_9,
	"iso8859_10": charmap.ISO8859_10, "iso8859_13": charmap.ISO8859_13, "iso8859_14": charmap.ISO8859_14,
	"iso8859_15": charmap.ISO8859_15, "iso8859_16": charmap.ISO8859_16,
	"cp1250": charmap.Windows1250, "cp1251": charmap.Windows1251, "cp1252": charmap.Windows1252,
	"cp1253": charmap.Windows1253, "cp1254": charmap.Windows1254, "cp1255": charmap.Windows1255,
	"cp1256": charmap.Windows1256, "cp1257": charmap.Windows1257, "cp1258": charmap.Windows1258,
	"cp437": charmap.CodePage437, "cp850": charmap.CodePage850, "cp866": charmap.CodePage866,
	"koi8_r": charmap.KOI8R, "koi8_u": charmap.KOI8U, "mac_roman": charmap.Macintosh,
	"mac_cyrillic": charmap.MacintoshCyrillic,
	"gb2312":       simplifiedchinese.GBK, "gbk": simplifiedchinese.GBK, "gb18030": simplifiedchinese.GB18030,
	"big5": traditionalchinese.Big5, "euc_jp": japanese.EUCJP, "shift_jis": japanese.ShiftJIS,
	"iso2022_jp": japanese.ISO2022JP, "euc_kr": korean.EUCKR, "cp949": korean.EUCKR,
	codecUTF16:   xunicode.UTF16(xunicode.LittleEndian, xunicode.UseBOM),
	codecUTF16LE: xunicode.UTF16(xunicode.LittleEndian, xunicode.IgnoreBOM),
	codecUTF16BE: xunicode.UTF16(xunicode.BigEndian, xunicode.IgnoreBOM),
	codecUTF32:   utf32.UTF32(utf32.LittleEndian, utf32.UseBOM),
	codecUTF32LE: utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM),
	codecUTF32BE: utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM),
}

func lookupCodec(name string) (string, bool) {
	norm := pyNormalizeEncoding(name)
	for _, n := range []string{norm, strings.ReplaceAll(norm, ".", "_")} {
		if c, ok := pyCodecAliases[n]; ok {
			return c, true
		}
	}
	if norm == codecASCII || norm == codecUTF8 {
		return norm, true
	}
	if _, ok := pyCodecs[norm]; ok {
		return norm, true
	}
	return "", false
}

func decodeLossy(data []byte, name string) (string, bool) {
	codec, ok := lookupCodec(name)
	if !ok {
		return "", false
	}
	switch codec {
	case codecUTF8:
		return string(data), true
	case codecASCII:
		var b strings.Builder
		for _, c := range data {
			if c < utf8.RuneSelf {
				b.WriteByte(c)
			} else {
				b.WriteRune(utf8.RuneError)
			}
		}
		return b.String(), true
	}
	out, err := pyCodecs[codec].NewDecoder().Bytes(data)
	if err != nil {
		return string(data), true
	}
	return string(out), true
}

func decodeOrUTF8(data []byte, name string) string {
	if s, ok := decodeLossy(data, name); ok {
		return s
	}
	return string(data)
}

var latinGuesses = map[string]bool{
	"WINDOWS-1252": true, "WINDOWS-1250": true, "WINDOWS-1254": true, "WINDOWS-1257": true,
	"ISO-8859-1": true, "ISO-8859-2": true, "ISO-8859-3": true, "ISO-8859-4": true, "ISO-8859-9": true,
	"ISO-8859-15": true, "ISO-8859-16": true,
}

var utf8Names = map[string]bool{"utf-8": true, "utf8": true, "us-ascii": true, "ascii": true, "": true}

func hasBOM(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) || bytes.HasPrefix(data, []byte{0xFF, 0xFE}) ||
		bytes.HasPrefix(data, []byte{0xFE, 0xFF})
}

func strictUTF16(data []byte, bigEndian bool) ([]rune, bool) {
	units := make([]uint16, len(data)/2)
	for i := range units {
		if bigEndian {
			units[i] = uint16(data[2*i])<<8 | uint16(data[2*i+1])
		} else {
			units[i] = uint16(data[2*i+1])<<8 | uint16(data[2*i])
		}
	}
	var out []rune
	for i := 0; i < len(units); i++ {
		u := rune(units[i])
		switch {
		case u >= 0xD800 && u < 0xDC00:
			if i+1 >= len(units) || units[i+1] < 0xDC00 || units[i+1] >= 0xE000 {
				return nil, false
			}
			out = append(out, utf16.DecodeRune(u, rune(units[i+1])))
			i++
		case u >= 0xDC00 && u < 0xE000:
			return nil, false
		default:
			out = append(out, u)
		}
	}
	return out, true
}

func cp1252Strict(data []byte) bool {
	for _, c := range data {
		switch c {
		case 0x81, 0x8D, 0x8F, 0x90, 0x9D:
			return false
		}
	}
	return true
}

func charsetOf(data []byte) (string, string) {
	if hasBOM(data) {
		return "", "bom"
	}
	if utf8.Valid(data) {
		return "", "utf-8"
	}
	for _, enc := range []string{"utf-16-be", "utf-16-le"} {
		if bytes.IndexByte(data, 0) >= 0 && len(data)%2 == 0 {
			decoded, ok := strictUTF16(data, enc == "utf-16-be")
			if !ok {
				continue
			}
			printable := 0
			for _, r := range decoded {
				if unicode.IsPrint(r) || isPySpace(r) {
					printable++
				}
			}
			if float64(printable) >= 0.98*float64(len(decoded)) {
				return "trait:charset.utf-16-no-bom", strings.ToUpper(enc)
			}
		}
	}
	guess := strings.ToUpper(strings.TrimFunc(string(run([]string{"uchardet"}, data, probeTimeout)), isPySpace))
	if strings.HasPrefix(guess, "UTF-16") || strings.HasPrefix(guess, "UTF-32") {
		return "trait:charset.utf-16-no-bom", guess
	}
	if latinGuesses[guess] && cp1252Strict(data) {
		return "trait:charset.windows-1252", guess
	}
	return "trait:charset.legacy-other", guess
}

func decodeText(data []byte, guess string) string {
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		return string(data[3:])
	}
	if bytes.HasPrefix(data, []byte{0xFF, 0xFE}) || bytes.HasPrefix(data, []byte{0xFE, 0xFF}) {
		return decodeOrUTF8(data, "utf-16")
	}
	if strings.HasPrefix(guess, "UTF-16") {
		if len(data) > 0 && data[0] == 0 {
			return decodeOrUTF8(data, "utf-16-be")
		}
		return decodeOrUTF8(data, "utf-16-le")
	}
	if guess == "utf-8" || guess == "bom" || guess == "" {
		return string(data)
	}
	return decodeOrUTF8(data, guess)
}
