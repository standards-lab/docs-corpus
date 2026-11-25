package text

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/ianaindex"
	xunicode "golang.org/x/text/encoding/unicode"
)

// Encoding returns the encoding golang.org/x/text's IANA index finds for the charset name, or
// else the one its WHATWG index finds, or nil when neither supports name.
func Encoding(name string) encoding.Encoding {
	if e, err := ianaindex.IANA.Encoding(name); err == nil && e != nil {
		return e
	}
	if e, err := htmlindex.Get(name); err == nil {
		return e
	}
	return nil
}

var utf16Names = map[string]encoding.Encoding{
	"utf-16":    xunicode.UTF16(xunicode.LittleEndian, xunicode.UseBOM),
	"utf-16-le": xunicode.UTF16(xunicode.LittleEndian, xunicode.IgnoreBOM),
	"utf-16-be": xunicode.UTF16(xunicode.BigEndian, xunicode.IgnoreBOM),
}

// DecodeOrUTF8 decodes data with the encoding [Encoding] finds for name, replacing what the
// encoding cannot decode, and returns data as it is when name names no encoding. "utf-16" reads a
// byte order mark and defaults to little-endian, "utf-16-le" and "utf-16-be" are little- and
// big-endian, ASCII replaces each byte above 0x7F with U+FFFD, and UTF-8 keeps data's bytes as they
// are.
func DecodeOrUTF8(data []byte, name string) string {
	e, ok := utf16Names[name]
	if !ok {
		e = Encoding(name)
	}
	if e == nil || e == xunicode.UTF8 {
		return string(data)
	}
	out, err := e.NewDecoder().Bytes(data)
	if err != nil {
		return string(data)
	}
	return string(out)
}
