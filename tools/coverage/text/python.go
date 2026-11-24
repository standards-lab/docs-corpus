package text

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// IsSpace reports whether r is whitespace as Python's str.isspace reads it, which adds the
// separators U+001C to U+001F to Unicode's white space.
func IsSpace(r rune) bool {
	return unicode.IsSpace(r) || (r >= 0x1C && r <= 0x1F)
}

// Strip trims s's leading and trailing whitespace, as Python's str.strip does.
func Strip(s string) string {
	return strings.TrimFunc(s, IsSpace)
}

// Fields splits s at runs of whitespace, as Python's str.split does with no separator.
func Fields(s string) []string {
	return strings.FieldsFunc(s, IsSpace)
}

// SplitLines splits s into lines, without their ends, as Python's str.splitlines does: at \n,
// \r, \r\n, \v, \f, U+001C to U+001E, U+0085, U+2028, and U+2029.
func SplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); {
		r, size := rune(s[i]), 1
		if r >= 0x80 {
			r, size = utf8.DecodeRuneInString(s[i:])
		}
		switch r {
		case '\n', '\v', '\f', 0x1C, 0x1D, 0x1E, 0x85, 0x2028, 0x2029:
			lines = append(lines, s[start:i])
			start = i + size
		case '\r':
			lines = append(lines, s[start:i])
			if i+1 < len(s) && s[i+1] == '\n' {
				size = 2
			}
			start = i + size
		}
		i += size
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
