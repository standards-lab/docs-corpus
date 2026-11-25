package mail

// Parser parses an RFC 5322 message into its parts. The placement rules read a message only
// through it, so one parser can replace another.
type Parser interface {
	// Parse parses data and returns its parts depth first, the message itself first.
	Parse(data []byte) []Part
	// DecodeHeader decodes a header value's RFC 2047 encoded words.
	DecodeHeader(value string) string
}

// Part is one part of a parsed message, or the message itself.
type Part interface {
	// Header returns the value of the part's first header field named name, case-insensitively,
	// unfolded, and whether the part has one.
	Header(name string) (string, bool)
	// ContentType returns the part's media type, lower-cased.
	ContentType() string
	// Charset returns the Content-Type's charset parameter, lower-cased, or "".
	Charset() string
	// Param returns the parameter name of the header field named header, unquoted, and whether
	// the field has it.
	Param(header, name string) (string, bool)
	// Filename returns the part's file name, from its Content-Disposition or Content-Type, or "".
	Filename() string
	// Payload returns the part's body decoded from its Content-Transfer-Encoding, or nil for a
	// part that holds others.
	Payload() []byte
	// Text returns the payload decoded from its charset.
	Text() string
}
