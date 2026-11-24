// From CPython (v3.14.8), PSF-2.0: the compat32 parsing of Lib/email (feedparser, message, utils, header, _encoded_words) and the a2b decoders of Modules/binascii.c, ported to Go and reduced to what place-traits reads.
//
// Copyright (c) 2001 Python Software Foundation; All Rights Reserved
//
// PYTHON SOFTWARE FOUNDATION LICENSE VERSION 2
// --------------------------------------------
//
// 1. This LICENSE AGREEMENT is between the Python Software Foundation
// ("PSF"), and the Individual or Organization ("Licensee") accessing and
// otherwise using this software ("Python") in source or binary form and
// its associated documentation.
//
// 2. Subject to the terms and conditions of this License Agreement, PSF hereby
// grants Licensee a nonexclusive, royalty-free, world-wide license to reproduce,
// analyze, test, perform and/or display publicly, prepare derivative works,
// distribute, and otherwise use Python alone or in any derivative version,
// provided, however, that PSF's License Agreement and PSF's notice of copyright,
// i.e., "Copyright (c) 2001 Python Software Foundation; All Rights Reserved"
// are retained in Python alone or in any derivative version prepared by Licensee.
//
// 3. In the event Licensee prepares a derivative work that is based on
// or incorporates Python or any part thereof, and wants to make
// the derivative work available to others as provided herein, then
// Licensee hereby agrees to include in any such work a brief summary of
// the changes made to Python.
//
// 4. PSF is making Python available to Licensee on an "AS IS"
// basis.  PSF MAKES NO REPRESENTATIONS OR WARRANTIES, EXPRESS OR
// IMPLIED.  BY WAY OF EXAMPLE, BUT NOT LIMITATION, PSF MAKES NO AND
// DISCLAIMS ANY REPRESENTATION OR WARRANTY OF MERCHANTABILITY OR FITNESS
// FOR ANY PARTICULAR PURPOSE OR THAT THE USE OF PYTHON WILL NOT
// INFRINGE ANY THIRD PARTY RIGHTS.
//
// 5. PSF SHALL NOT BE LIABLE TO LICENSEE OR ANY OTHER USERS OF PYTHON
// FOR ANY INCIDENTAL, SPECIAL, OR CONSEQUENTIAL DAMAGES OR LOSS AS
// A RESULT OF MODIFYING, DISTRIBUTING, OR OTHERWISE USING PYTHON,
// OR ANY DERIVATIVE THEREOF, EVEN IF ADVISED OF THE POSSIBILITY THEREOF.
//
// 6. This License Agreement will automatically terminate upon a material
// breach of its terms and conditions.
//
// 7. Nothing in this License Agreement shall be deemed to create any
// relationship of agency, partnership, or joint venture between PSF and
// Licensee.  This License Agreement does not grant permission to use PSF
// trademarks or trade name in a trademark sense to endorse or promote
// products or services of Licensee, or any third party.
//
// 8. By copying, installing or otherwise using Python, Licensee
// agrees to be bound by the terms and conditions of this License
// Agreement.

package main

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

type header struct {
	name, value string
}

type part struct {
	headers     []header
	defaultType string
	payload     string
	parts       []*part
	multipart   bool
}

func (p *part) attach(child *part) {
	p.parts = append(p.parts, child)
	p.multipart = true
}

func (p *part) walk(visit func(*part)) {
	visit(p)
	for _, c := range p.parts {
		c.walk(visit)
	}
}

func sanitizeHeader(v string) string {
	var b strings.Builder
	for i := 0; i < len(v); i++ {
		if v[i] >= utf8.RuneSelf {
			b.WriteRune(utf8.RuneError)
		} else {
			b.WriteByte(v[i])
		}
	}
	return b.String()
}

func (p *part) rawGet(name string) (string, bool) {
	for _, h := range p.headers {
		if strings.EqualFold(h.name, name) {
			return h.value, true
		}
	}
	return "", false
}

func (p *part) get(name string) (string, bool) {
	v, ok := p.rawGet(name)
	return sanitizeHeader(v), ok
}

func pyStrip(s string) string {
	return strings.TrimFunc(s, isPySpace)
}

func (p *part) contentType() string {
	v, ok := p.get("content-type")
	if !ok {
		return p.defaultType
	}
	ctype, _, _ := strings.Cut(v, ";")
	ctype = strings.ToLower(pyStrip(ctype))
	if strings.Count(ctype, "/") != 1 {
		return "text/plain"
	}
	return ctype
}

func (p *part) mainType() string {
	t, _, _ := strings.Cut(p.contentType(), "/")
	return t
}

func parseParamList(s string) []string {
	s = ";" + s
	var plist []string
	for strings.HasPrefix(s, ";") {
		s = s[1:]
		end := strings.Index(s, ";")
		for end > 0 && (strings.Count(s[:end], `"`)-strings.Count(s[:end], `\"`))%2 != 0 {
			next := strings.Index(s[end+1:], ";")
			if next < 0 {
				end = -1
			} else {
				end += 1 + next
			}
		}
		if end < 0 {
			end = len(s)
		}
		f := s[:end]
		if i := strings.Index(f, "="); i >= 0 {
			f = strings.ToLower(pyStrip(f[:i])) + "=" + pyStrip(f[i+1:])
		}
		plist = append(plist, pyStrip(f))
		s = s[end:]
	}
	return plist
}

type paramValue struct {
	text       string
	extended   bool
	hasCharset bool
	charset    string
}

type param struct {
	name  string
	value paramValue
}

func unquote(s string) string {
	if len(s) > 1 {
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			return strings.ReplaceAll(strings.ReplaceAll(s[1:len(s)-1], `\\`, `\`), `\"`, `"`)
		}
		if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func quote(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`)
}

var rfc2231Continuation = regexp.MustCompile(`^(\w+)\*(([0-9]+)\*?)?$`)

func percentDecodeLatin1(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			if n, err := strconv.ParseUint(s[i+1:i+3], 16, 8); err == nil {
				b.WriteRune(rune(n))
				i += 2
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func (p *part) params(headerName string) ([]param, bool) {
	v, ok := p.get(headerName)
	if !ok {
		return nil, false
	}
	var raw [][2]string
	for _, item := range parseParamList(v) {
		name, val, found := strings.Cut(item, "=")
		if !found {
			raw = append(raw, [2]string{pyStrip(item), ""})
		} else {
			raw = append(raw, [2]string{pyStrip(name), pyStrip(val)})
		}
	}
	out := []param{{raw[0][0], paramValue{text: raw[0][1]}}}
	type continuation struct {
		num     int
		hasNum  bool
		text    string
		encoded bool
	}
	var order []string
	groups := map[string][]continuation{}
	for _, nv := range raw[1:] {
		name, value := nv[0], unquote(nv[1])
		encoded := strings.HasSuffix(name, "*")
		if m := rfc2231Continuation.FindStringSubmatch(name); m != nil {
			c := continuation{text: value, encoded: encoded}
			if m[3] != "" {
				c.num, _ = strconv.Atoi(m[3])
				c.hasNum = true
			}
			if _, seen := groups[m[1]]; !seen {
				order = append(order, m[1])
			}
			groups[m[1]] = append(groups[m[1]], c)
			continue
		}
		out = append(out, param{name, paramValue{text: `"` + quote(value) + `"`}})
	}
	for _, name := range order {
		cs := groups[name]
		slices.SortStableFunc(cs, func(a, b continuation) int {
			if a.num != b.num {
				return a.num - b.num
			}
			return strings.Compare(a.text, b.text)
		})
		var value strings.Builder
		extended := false
		for _, c := range cs {
			s := c.text
			if c.encoded {
				s = percentDecodeLatin1(s)
				extended = true
			}
			value.WriteString(s)
		}
		joined := quote(value.String())
		if extended {
			pieces := strings.SplitN(joined, "'", 3)
			if len(pieces) <= 2 {
				out = append(out, param{name, paramValue{text: `"` + joined + `"`, extended: true}})
			} else {
				out = append(out, param{name, paramValue{text: `"` + pieces[2] + `"`, extended: true, hasCharset: true, charset: pieces[0]}})
			}
		} else {
			out = append(out, param{name, paramValue{text: `"` + joined + `"`}})
		}
	}
	return out, true
}

func (p *part) param(name, headerName string) (paramValue, bool) {
	params, ok := p.params(headerName)
	if !ok {
		return paramValue{}, false
	}
	for _, pr := range params {
		if strings.EqualFold(pr.name, name) {
			v := pr.value
			v.text = unquote(v.text)
			return v, true
		}
	}
	return paramValue{}, false
}

func latin1Bytes(s string) []byte {
	var out []byte
	for _, r := range s {
		out = append(out, byte(r))
	}
	return out
}

func collapseRFC2231(v paramValue) string {
	if !v.extended {
		return unquote(v.text)
	}
	charset := v.charset
	if !v.hasCharset {
		charset = "us-ascii"
	}
	if s, ok := decodeLossy(latin1Bytes(v.text), charset); ok {
		return s
	}
	return unquote(v.text)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func (p *part) contentCharset() string {
	v, ok := p.param("charset", "content-type")
	if !ok {
		return ""
	}
	if !isASCII(v.text) {
		return ""
	}
	return strings.ToLower(v.text)
}

func (p *part) filename() string {
	v, ok := p.param("filename", "content-disposition")
	if !ok {
		if v, ok = p.param("name", "content-type"); !ok {
			return ""
		}
	}
	return pyStrip(collapseRFC2231(v))
}

func (p *part) boundary() (string, bool) {
	v, ok := p.param("boundary", "content-type")
	if !ok {
		return "", false
	}
	return strings.TrimRightFunc(collapseRFC2231(v), isPySpace), true
}

func bytesSplitLines(b string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(b); i++ {
		switch b[i] {
		case '\n':
			lines = append(lines, b[start:i])
			start = i + 1
		case '\r':
			lines = append(lines, b[start:i])
			if i+1 < len(b) && b[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}
	if start < len(b) {
		lines = append(lines, b[start:])
	}
	return lines
}

func (p *part) decodedPayload() []byte {
	if p.multipart {
		return nil
	}
	v, _ := p.get("content-transfer-encoding")
	cte := strings.ToLower(pyStrip(v))
	switch cte {
	case "quoted-printable":
		return a2bQP(p.payload)
	case "base64":
		return decodeB(strings.Join(bytesSplitLines(p.payload), ""))
	case "x-uuencode", "uuencode", "uue", "x-uue":
		if out, ok := decodeUU(p.payload); ok {
			return out
		}
	}
	return []byte(p.payload)
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func hexValue(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	}
	return c - 'A' + 10
}

func a2bQP(data string) []byte {
	var out []byte
	for in := 0; in < len(data); {
		if data[in] != '=' {
			out = append(out, data[in])
			in++
			continue
		}
		in++
		if in >= len(data) {
			break
		}
		switch {
		case data[in] == '\n' || data[in] == '\r':
			if data[in] != '\n' {
				for in < len(data) && data[in] != '\n' {
					in++
				}
			}
			if in < len(data) {
				in++
			}
		case data[in] == '=':
			out = append(out, '=')
			in++
		case in+1 < len(data) && isHex(data[in]) && isHex(data[in+1]):
			out = append(out, hexValue(data[in])<<4|hexValue(data[in+1]))
			in += 2
		default:
			out = append(out, '=')
		}
	}
	return out
}

func base64Value(c byte) (byte, bool) {
	switch {
	case c >= 'A' && c <= 'Z':
		return c - 'A', true
	case c >= 'a' && c <= 'z':
		return c - 'a' + 26, true
	case c >= '0' && c <= '9':
		return c - '0' + 52, true
	case c == '+':
		return 62, true
	case c == '/':
		return 63, true
	}
	return 0, false
}

func a2bBase64(data string, strict bool) ([]byte, bool) {
	var out []byte
	quadPos, pads := 0, 0
	var left byte
	paddingStarted := false
	for i := 0; i < len(data); i++ {
		c := data[i]
		if c == '=' {
			paddingStarted = true
			if strict && quadPos == 0 {
				return nil, false
			}
			if quadPos >= 2 {
				pads++
				if quadPos+pads >= 4 {
					if strict && i+1 < len(data) {
						return nil, false
					}
					return out, true
				}
			}
			continue
		}
		v, ok := base64Value(c)
		if !ok {
			if strict {
				return nil, false
			}
			continue
		}
		if strict && paddingStarted {
			return nil, false
		}
		pads = 0
		switch quadPos {
		case 0:
			quadPos, left = 1, v
		case 1:
			out = append(out, left<<2|v>>4)
			quadPos, left = 2, v&0x0f
		case 2:
			out = append(out, left<<4|v>>2)
			quadPos, left = 3, v&0x03
		case 3:
			out = append(out, left<<6|v)
			quadPos, left = 0, 0
		}
	}
	if quadPos != 0 {
		return nil, false
	}
	return out, true
}

func decodeB(encoded string) []byte {
	padding := ""
	if r := len(encoded) % 4; r != 0 {
		padding = "==="[:4-r]
	}
	if out, ok := a2bBase64(encoded+padding, true); ok {
		return out
	}
	if out, ok := a2bBase64(encoded, false); ok {
		return out
	}
	if out, ok := a2bBase64(encoded+"==", false); ok {
		return out
	}
	return []byte(encoded)
}

func a2bUU(line string) ([]byte, bool) {
	line = strings.TrimRight(line, " \t\r\n")
	if line == "" {
		return make([]byte, 32), true
	}
	n := int((line[0] - ' ') & 0x3f)
	var out []byte
	var acc uint32
	bits := 0
	for i := 1; i < len(line) && len(out) < n; i++ {
		c := line[i]
		if c == '\n' || c == '\r' {
			break
		}
		if c < ' ' || c > ' '+64 {
			return nil, false
		}
		acc = acc<<6 | uint32((c-' ')&0x3f)
		bits += 6
		if bits >= 8 {
			bits -= 8
			out = append(out, byte(acc>>bits))
		}
	}
	for len(out) < n {
		out = append(out, 0)
	}
	return out, true
}

func decodeUU(payload string) ([]byte, bool) {
	lines := bytesSplitLines(payload)
	i := 0
	for ; i < len(lines); i++ {
		if rest, ok := strings.CutPrefix(lines[i], "begin "); ok {
			mode, _, _ := strings.Cut(rest, " ")
			if _, err := strconv.ParseInt(strings.TrimSpace(mode), 8, 64); err == nil {
				break
			}
		}
	}
	if i == len(lines) {
		return nil, false
	}
	var out []byte
	for _, line := range lines[i+1:] {
		if line == "" {
			return nil, false
		}
		if strings.Trim(line, " \t\r\n\f") == "end" {
			break
		}
		decoded, ok := a2bUU(line)
		if !ok {
			n := (int((line[0]-32)&63)*4 + 5) / 3
			if decoded, ok = a2bUU(line[:min(n, len(line))]); !ok {
				return nil, false
			}
		}
		out = append(out, decoded...)
	}
	return out, true
}

type feedInput struct {
	lines    []string
	pos      int
	pushback []string
	eof      []func(string) bool
}

func splitKeepEnds(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\n':
			lines = append(lines, s[start:i+1])
			start = i + 1
		case '\r':
			if i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
			lines = append(lines, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func (in *feedInput) readline() string {
	var line string
	switch {
	case len(in.pushback) > 0:
		line = in.pushback[len(in.pushback)-1]
		in.pushback = in.pushback[:len(in.pushback)-1]
	case in.pos < len(in.lines):
		line = in.lines[in.pos]
		in.pos++
	default:
		return ""
	}
	for i := len(in.eof) - 1; i >= 0; i-- {
		if in.eof[i](line) {
			in.pushback = append(in.pushback, line)
			return ""
		}
	}
	return line
}

func (in *feedInput) unread(line string) {
	in.pushback = append(in.pushback, line)
}

func (in *feedInput) rest() []string {
	var lines []string
	for line := in.readline(); line != ""; line = in.readline() {
		lines = append(lines, line)
	}
	return lines
}

func startsWithNewline(line string) bool {
	return strings.HasPrefix(line, "\r") || strings.HasPrefix(line, "\n")
}

func trimOneNewline(s string) string {
	switch {
	case strings.HasSuffix(s, "\r\n"):
		return s[:len(s)-2]
	case strings.HasSuffix(s, "\r"), strings.HasSuffix(s, "\n"):
		return s[:len(s)-1]
	}
	return s
}

var headerLine = regexp.MustCompile(`^(From |[\x21-\x39\x3b-\x7e]*:|[\t ])`)

type feedParser struct {
	in    *feedInput
	stack []*part
	cur   *part
	last  *part
}

func parseMessage(data string) *part {
	p := &feedParser{in: &feedInput{lines: splitKeepEnds(data)}}
	p.parse()
	return p.stack[0]
}

func (p *feedParser) newMessage() {
	msg := &part{defaultType: "text/plain"}
	if p.cur != nil && p.cur.contentType() == "multipart/digest" {
		msg.defaultType = "message/rfc822"
	}
	if len(p.stack) > 0 {
		p.stack[len(p.stack)-1].attach(msg)
	}
	p.stack = append(p.stack, msg)
	p.cur = msg
	p.last = msg
}

func (p *feedParser) popMessage() {
	p.stack = p.stack[:len(p.stack)-1]
	if len(p.stack) > 0 {
		p.cur = p.stack[len(p.stack)-1]
	} else {
		p.cur = nil
	}
}

func (p *feedParser) parseHeaders(lines []string) {
	lastHeader := ""
	var lastValue []string
	flush := func() {
		name, value, _ := strings.Cut(lastValue[0], ":")
		value = strings.TrimLeft(value, " \t") + strings.Join(lastValue[1:], "")
		p.cur.headers = append(p.cur.headers, header{name, strings.TrimRight(value, "\r\n")})
	}
	for lineno, line := range lines {
		if line[0] == ' ' || line[0] == '\t' {
			if lastHeader != "" {
				lastValue = append(lastValue, line)
			}
			continue
		}
		if lastHeader != "" {
			flush()
			lastHeader, lastValue = "", nil
		}
		if strings.HasPrefix(line, "From ") {
			if lineno == 0 {
				continue
			}
			if lineno == len(lines)-1 {
				p.in.unread(line)
				return
			}
			continue
		}
		i := strings.Index(line, ":")
		if i == 0 {
			continue
		}
		lastHeader = line[:i]
		lastValue = []string{line}
	}
	if lastHeader != "" {
		flush()
	}
}

var boundaryEnd = regexp.MustCompile(`^(--)?[ \t]*(\r\n|\r|\n)?$`)

func (p *feedParser) parse() {
	p.newMessage()
	var headers []string
	for {
		line := p.in.readline()
		if line == "" {
			break
		}
		if !headerLine.MatchString(line) {
			if !startsWithNewline(line) {
				p.in.unread(line)
			}
			break
		}
		headers = append(headers, line)
	}
	p.parseHeaders(headers)
	switch {
	case p.cur.contentType() == "message/delivery-status":
		for {
			p.in.eof = append(p.in.eof, startsWithNewline)
			p.parse()
			p.popMessage()
			p.in.eof = p.in.eof[:len(p.in.eof)-1]
			p.in.readline()
			line := p.in.readline()
			if line == "" {
				break
			}
			p.in.unread(line)
		}
	case p.cur.mainType() == "message":
		p.parse()
		p.popMessage()
	case p.cur.mainType() == "multipart":
		p.parseMultipart()
	default:
		p.cur.payload = strings.Join(p.in.rest(), "")
	}
}

func (p *feedParser) parseMultipart() {
	boundary, ok := p.cur.boundary()
	if !ok {
		p.cur.payload = strings.Join(p.in.rest(), "")
		return
	}
	separator := "--" + boundary
	match := func(line string) (bool, bool) {
		rest, ok := strings.CutPrefix(line, separator)
		if !ok {
			return false, false
		}
		m := boundaryEnd.FindStringSubmatch(rest)
		if m == nil && strings.HasSuffix(rest, "\n") {
			m = boundaryEnd.FindStringSubmatch(rest[:len(rest)-1])
		}
		if m == nil {
			return false, false
		}
		return true, m[1] != ""
	}
	isBoundary := func(line string) bool {
		ok, _ := match(line)
		return ok
	}
	capturingPreamble := true
	closeSeen := false
	var preamble []string
	for {
		line := p.in.readline()
		if line == "" {
			break
		}
		ok, end := match(line)
		if !ok {
			preamble = append(preamble, line)
			continue
		}
		if end {
			closeSeen = true
			break
		}
		if capturingPreamble {
			capturingPreamble = false
			p.in.unread(line)
			continue
		}
		for {
			line = p.in.readline()
			if !isBoundary(line) {
				p.in.unread(line)
				break
			}
		}
		p.in.eof = append(p.in.eof, isBoundary)
		p.parse()
		if !p.last.multipart && p.last.mainType() != "multipart" {
			p.last.payload = trimOneNewline(p.last.payload)
		}
		p.in.eof = p.in.eof[:len(p.in.eof)-1]
		p.popMessage()
		p.last = p.cur
	}
	if capturingPreamble {
		p.cur.payload = strings.Join(preamble, "")
		p.in.rest()
		return
	}
	if closeSeen {
		p.in.rest()
	}
}

var encodedWord = regexp.MustCompile(`=\?([^?]*?)\?([qQbB])\?(.*?)\?=`)

var qHex = regexp.MustCompile(`=[a-fA-F0-9]{2}`)

func headerText(value string) string {
	if !isASCII(value) {
		return sanitizeHeader(value)
	}
	if !encodedWord.MatchString(value) {
		return value
	}
	type word struct {
		text     string
		encoding string
		charset  string
	}
	var words []word
	for _, line := range pySplitLines(value) {
		first := true
		pos := 0
		pieces := encodedWord.FindAllStringSubmatchIndex(line, -1)
		for i := 0; i <= len(pieces); i++ {
			end := len(line)
			if i < len(pieces) {
				end = pieces[i][0]
			}
			unencoded := line[pos:end]
			if first {
				unencoded = strings.TrimLeftFunc(unencoded, isPySpace)
				first = false
			}
			if unencoded != "" {
				words = append(words, word{text: unencoded})
			}
			if i < len(pieces) {
				m := pieces[i]
				words = append(words, word{
					text:     line[m[6]:m[7]],
					encoding: strings.ToLower(line[m[4]:m[5]]),
					charset:  strings.ToLower(line[m[2]:m[3]]),
				})
				pos = m[1]
			}
		}
	}
	var kept []word
	for n, w := range words {
		if n+1 < len(words) && n > 0 && words[n+1].encoding != "" && words[n-1].encoding != "" &&
			w.text != "" && strings.TrimFunc(w.text, isPySpace) == "" {
			continue
		}
		kept = append(kept, w)
	}
	type chunk struct {
		data    []byte
		charset string
		encoded bool
	}
	var chunks []chunk
	for _, w := range kept {
		var data []byte
		switch w.encoding {
		case "q":
			s := strings.ReplaceAll(w.text, "_", " ")
			s = qHex.ReplaceAllStringFunc(s, func(m string) string {
				return string([]byte{hexValue(m[1])<<4 | hexValue(m[2])})
			})
			data = []byte(s)
		case "b":
			s := w.text
			if r := len(s) % 4; r != 0 {
				s += "==="[:4-r]
			}
			data, _ = a2bBase64(s, false)
		default:
			data = []byte(w.text)
		}
		encoded := w.encoding != ""
		if n := len(chunks); n > 0 && chunks[n-1].encoded == encoded && chunks[n-1].charset == w.charset {
			if !encoded {
				chunks[n-1].data = append(chunks[n-1].data, ' ')
			}
			chunks[n-1].data = append(chunks[n-1].data, data...)
			continue
		}
		chunks = append(chunks, chunk{data, w.charset, encoded})
	}
	var b strings.Builder
	for _, c := range chunks {
		if !c.encoded {
			b.Write(c.data)
			continue
		}
		b.WriteString(decodeOrUTF8(c.data, c.charset))
	}
	return b.String()
}
