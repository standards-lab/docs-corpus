// From CPython (v3.14.8), PSF-2.0: the codec aliases of Lib/encodings/aliases.py and normalize_encoding of Lib/encodings/__init__.py, ported to Go and reduced to the codecs golang.org/x/text provides.
//
// Copyright (c) 2001 Python Software Foundation; All Rights Reserved

package text

import (
	"strings"
	"unicode"
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

func normalizeEncoding(name string) string {
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
	norm := normalizeEncoding(name)
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

func decode(data []byte, name string) (string, bool) {
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

// DecodeOrUTF8 decodes data with the codec Python's codecs.lookup finds for name, replacing what
// the codec cannot decode, and returns data as it is when name names no codec. ASCII replaces each
// byte above 0x7F with U+FFFD, and UTF-8 keeps data's bytes as they are.
func DecodeOrUTF8(data []byte, name string) string {
	if s, ok := decode(data, name); ok {
		return s
	}
	return string(data)
}
