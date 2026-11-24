package stdmail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	netmail "net/mail"
	"net/textproto"
	"regexp"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
)

// Parser parses messages with Go's standard library. Its zero value is ready to use.
type Parser struct{}

type part struct {
	header      textproto.MIMEHeader
	defaultType string
	body        []byte
	parts       []*part
}

// Parse parses data as an RFC 5322 message, its MIME parts as RFC 2045 and 2046 define them, and
// returns them depth first, the message itself first. A leading mbox From line is not part of
// the message. The header section ends at its empty line or at the first line that is neither a
// field nor a field's continuation. A message/rfc822 or message/global part holds the message
// its body carries, and a multipart part the parts its boundary delimits; a multipart part with
// no boundary, or whose boundary its body never has, holds none and keeps its body.
//
// A part without a Content-Type has the type RFC 2045 and 2046 give it, text/plain, or
// message/rfc822 in a multipart/digest, and one whose Content-Type is not a valid media type is
// text/plain, as RFC 2045 recommends. Its parameters are RFC 2231 decoded, and its file name is
// the Content-Disposition's filename, or else the Content-Type's name, trimmed of white space. Its
// payload is decoded from quoted-printable or base64 as far as it decodes, and its text from its
// charset with golang.org/x/text, kept as it is when the charset is UTF-8, US-ASCII, absent, or
// one golang.org/x/text does not know.
func (Parser) Parse(data []byte) []mail.Part {
	var parts []mail.Part
	parseMessage(data, "text/plain").walk(func(p *part) { parts = append(parts, p) })
	return parts
}

var wordDecoder = &mime.WordDecoder{CharsetReader: charsetReader}

// DecodeHeader decodes a header value's RFC 2047 encoded words, decoding each word's charset with
// golang.org/x/text. A value with a word it cannot decode is returned as it is.
func (Parser) DecodeHeader(value string) string {
	decoded, err := wordDecoder.DecodeHeader(value)
	if err != nil {
		return value
	}
	return decoded
}

func (p *part) walk(visit func(*part)) {
	visit(p)
	for _, c := range p.parts {
		c.walk(visit)
	}
}

var fieldLine = regexp.MustCompile(`^[\x21-\x39\x3b-\x7e]+[ \t]*:`)

func splitHeader(data []byte) (head, body []byte) {
	for i := 0; i < len(data); {
		end := bytes.IndexByte(data[i:], '\n')
		next := len(data)
		if end >= 0 {
			next = i + end + 1
		}
		line := bytes.TrimRight(data[i:next], "\r\n")
		switch {
		case len(line) == 0:
			return data[:i], data[next:]
		case fieldLine.Match(line), i > 0 && (line[0] == ' ' || line[0] == '\t'):
			i = next
		default:
			return data[:i], data[i:]
		}
	}
	return data, nil
}

func parseMessage(data []byte, defaultType string) *part {
	if bytes.HasPrefix(data, []byte("From ")) {
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			data = data[i+1:]
		} else {
			data = nil
		}
	}
	head, body := splitHeader(data)
	header := textproto.MIMEHeader{}
	if msg, err := netmail.ReadMessage(bytes.NewReader(append(bytes.Clone(head), '\n'))); err == nil {
		header = textproto.MIMEHeader(msg.Header)
	}
	return parseEntity(header, body, defaultType)
}

func parseEntity(header textproto.MIMEHeader, body []byte, defaultType string) *part {
	p := &part{header: header, defaultType: defaultType, body: body}
	ctype := p.ContentType()
	switch {
	case ctype == "message/rfc822" || ctype == "message/global":
		p.parts = []*part{parseMessage(p.Payload(), "text/plain")}
	case strings.HasPrefix(ctype, "multipart/"):
		boundary, ok := p.Param("Content-Type", "boundary")
		if !ok || boundary == "" {
			return p
		}
		childType := "text/plain"
		if ctype == "multipart/digest" {
			childType = "message/rfc822"
		}
		r := multipart.NewReader(bytes.NewReader(body), boundary)
		for {
			raw, err := r.NextRawPart()
			if err != nil {
				break
			}
			content, err := io.ReadAll(raw)
			p.parts = append(p.parts, parseEntity(raw.Header, content, childType))
			if err != nil {
				break
			}
		}
	}
	return p
}

func (p *part) Header(name string) (string, bool) {
	values := p.header.Values(name)
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func (p *part) params(header string) (string, map[string]string) {
	v, ok := p.Header(header)
	if !ok {
		return "", nil
	}
	mediatype, params, err := mime.ParseMediaType(v)
	if err != nil && !errors.Is(err, mime.ErrInvalidMediaParameter) {
		return "", nil
	}
	return mediatype, params
}

func (p *part) ContentType() string {
	if _, ok := p.Header("Content-Type"); !ok {
		return p.defaultType
	}
	mediatype, _ := p.params("Content-Type")
	if kind, subtype, ok := strings.Cut(mediatype, "/"); !ok || kind == "" || subtype == "" {
		return "text/plain"
	}
	return mediatype
}

func (p *part) Param(header, name string) (string, bool) {
	_, params := p.params(header)
	v, ok := params[strings.ToLower(name)]
	return v, ok
}

func (p *part) Charset() string {
	cs, _ := p.Param("Content-Type", "charset")
	return strings.ToLower(cs)
}

func (p *part) Filename() string {
	name, ok := p.Param("Content-Disposition", "filename")
	if !ok {
		name, _ = p.Param("Content-Type", "name")
	}
	return strings.TrimSpace(name)
}

func (p *part) Payload() []byte {
	if len(p.parts) > 0 {
		return nil
	}
	cte, _ := p.Header("Content-Transfer-Encoding")
	switch strings.ToLower(strings.TrimSpace(cte)) {
	case "quoted-printable":
		out, _ := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(p.body)))
		return out
	case "base64":
		clean := bytes.Map(func(r rune) rune {
			if r == '=' || r == ' ' || r == '\t' || r == '\r' || r == '\n' {
				return -1
			}
			return r
		}, p.body)
		out := make([]byte, base64.RawStdEncoding.DecodedLen(len(clean)))
		n, _ := base64.RawStdEncoding.Decode(out, clean)
		return out[:n]
	}
	return p.body
}

func (p *part) Text() string {
	payload := p.Payload()
	if e := lookup(p.Charset()); e != nil {
		if s, err := e.NewDecoder().Bytes(payload); err == nil {
			return string(s)
		}
	}
	return string(payload)
}

func lookup(charset string) encoding.Encoding {
	switch charset {
	case "", "utf-8", "utf8", "us-ascii", "ascii":
		return nil
	}
	if e, err := ianaindex.MIME.Encoding(charset); err == nil && e != nil {
		return e
	}
	if e, err := htmlindex.Get(charset); err == nil {
		return e
	}
	return nil
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	e := lookup(strings.ToLower(charset))
	if e == nil {
		return nil, errors.New("unknown charset " + charset)
	}
	return e.NewDecoder().Reader(input), nil
}
