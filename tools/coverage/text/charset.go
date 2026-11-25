package text

import (
	"bytes"
	"context"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var latinGuesses = map[string]bool{
	"WINDOWS-1252": true, "WINDOWS-1250": true, "WINDOWS-1254": true, "WINDOWS-1257": true,
	"ISO-8859-1": true, "ISO-8859-2": true, "ISO-8859-3": true, "ISO-8859-4": true, "ISO-8859-9": true,
	"ISO-8859-15": true, "ISO-8859-16": true,
}

var utf8Names = map[string]bool{"utf-8": true, "utf8": true, "us-ascii": true, "ascii": true, "": true}

// UTF8Name reports whether a declared charset, lower-cased, names UTF-8 or its ASCII subset, or
// is empty.
func UTF8Name(name string) bool {
	return utf8Names[name]
}

// Guesser guesses the charset of bytes that are neither UTF-8 nor UTF-16, as uchardet does.
type Guesser interface {
	// Guess returns the guesser's output for data, the charset's name on a line, or nil when it
	// could not run.
	Guess(ctx context.Context, data []byte) []byte
}

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

// Charset returns the trait:charset ID data places, or "" for none, and the charset it found:
// "bom" for a byte order mark, "utf-8" for valid UTF-8, UTF-16-BE or UTF-16-LE for text that
// decodes as UTF-16 without a BOM, and otherwise g's guess, upper-cased.
func Charset(ctx context.Context, data []byte, g Guesser) (string, string) {
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
				if unicode.IsPrint(r) || IsSpace(r) {
					printable++
				}
			}
			if float64(printable) >= 0.98*float64(len(decoded)) {
				return "trait:charset.utf-16-no-bom", strings.ToUpper(enc)
			}
		}
	}
	guess := strings.ToUpper(strings.TrimFunc(string(g.Guess(ctx, data)), IsSpace))
	if strings.HasPrefix(guess, "UTF-16") || strings.HasPrefix(guess, "UTF-32") {
		return "trait:charset.utf-16-no-bom", guess
	}
	if latinGuesses[guess] && cp1252Strict(data) {
		return "trait:charset.windows-1252", guess
	}
	return "trait:charset.legacy-other", guess
}

// DecodeText decodes a text file's data by its byte order mark, or else by guess, Charset's
// second result lower-cased unless it is "bom" or "utf-8".
func DecodeText(data []byte, guess string) string {
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		return string(data[3:])
	}
	if bytes.HasPrefix(data, []byte{0xFF, 0xFE}) || bytes.HasPrefix(data, []byte{0xFE, 0xFF}) {
		return DecodeOrUTF8(data, "utf-16")
	}
	if strings.HasPrefix(guess, "UTF-16") {
		if len(data) > 0 && data[0] == 0 {
			return DecodeOrUTF8(data, "utf-16-be")
		}
		return DecodeOrUTF8(data, "utf-16-le")
	}
	if guess == "utf-8" || guess == "bom" || guess == "" {
		return string(data)
	}
	return DecodeOrUTF8(data, guess)
}
