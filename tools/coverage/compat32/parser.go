package compat32

import "github.com/standards-lab/docs-corpus/tools/coverage/mail"

// Parser parses messages as CPython's email package does under its compat32 policy. Its zero
// value is ready to use.
type Parser struct{}

// Parse parses data and returns its parts depth first, the message itself first.
func (Parser) Parse(data []byte) []mail.Part {
	var parts []mail.Part
	parse(string(data)).walk(func(p *part) { parts = append(parts, p) })
	return parts
}

// DecodeHeader decodes a header value's RFC 2047 encoded words, as email.header's decode_header
// and make_header do, replacing what a charset cannot decode. A value that is not ASCII has each
// byte above 0x7F replaced with U+FFFD instead.
func (Parser) DecodeHeader(value string) string {
	return decodeHeader(value)
}
