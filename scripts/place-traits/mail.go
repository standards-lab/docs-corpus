package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

type set map[string]bool

func (s set) add(ids ...string) {
	for _, id := range ids {
		s[id] = true
	}
}

func (s set) merge(other set) {
	for id := range other {
		s[id] = true
	}
}

var (
	declaredWord = regexp.MustCompile(`=\?[^?]+\?[bBqQ]\?`)
	htmlInline   = regexp.MustCompile(`(?i)<img[^>]+src=["']?(data:|cid:)`)
	htmlLink     = regexp.MustCompile(`(?i)<a[^>]+href=["']?https?://`)
)

func hasPlainLink(text string) bool {
	for _, scheme := range []string{"http://", "https://"} {
		rest := text
		for {
			i := strings.Index(rest, scheme)
			if i < 0 {
				break
			}
			rest = rest[i+len(scheme):]
			if r, _ := utf8.DecodeRuneInString(rest); rest != "" && !isPySpace(r) {
				return true
			}
		}
	}
	return false
}

func message(data []byte) (set, set, string, *object) {
	msg := parseMessage(string(data))
	traits, features := set{}, set{}
	var texts []string
	parts := 0
	head, _, _ := bytes.Cut(data, []byte("\n\n"))
	if declaredWord.Match(head) {
		traits.add("trait:charset.declared")
	}
	bodies := set{}
	msg.walk(func(p *part) {
		parts++
		ctype := p.contentType()
		cs := p.contentCharset()
		if !utf8Names[cs] {
			traits.add("trait:charset.declared")
		}
		disp, _ := p.get("Content-Disposition")
		disp = strings.ToLower(disp)
		filename := strings.ToLower(p.filename())
		if ctype == "message/rfc822" || (strings.HasPrefix(ctype, "message/") && p != msg && ctype != "message/delivery-status") {
			features.add("feature:embedded-messages")
		}
		if ctype == "application/ms-tnef" || filename == "winmail.dat" {
			features.add("feature:tnef-part")
		}
		if ctype == "application/pkcs7-mime" || ctype == "application/x-pkcs7-mime" {
			features.add("feature:smime-opaque")
		}
		if protocol, _ := p.param("protocol", "content-type"); ctype == "multipart/signed" && strings.Contains(protocol.text, "pkcs7") {
			features.add("feature:smime-signed")
		}
		if filename == "message.rpmsg" {
			features.add("feature:rights-managed")
		}
		contentID, hasContentID := p.get("Content-ID")
		if contentID != "" && strings.HasPrefix(ctype, "image/") {
			features.add("feature:inline-images")
		}
		if strings.Contains(disp, "attachment") || (p.filename() != "" && !strings.HasPrefix(ctype, "multipart/") && !hasContentID) {
			features.add("feature:attachments")
		}
		if (ctype == "text/plain" || ctype == "text/html") && !strings.Contains(disp, "attachment") {
			bodies.add(ctype)
			name := cs
			if name == "" {
				name = "utf-8"
			}
			text := decodeOrUTF8(p.decodedPayload(), name)
			texts = append(texts, text)
			if ctype == "text/html" {
				if htmlInline.MatchString(text) {
					features.add("feature:inline-images")
				}
				if htmlLink.MatchString(text) {
					features.add("feature:hyperlinks")
				}
			} else if hasPlainLink(text) {
				features.add("feature:hyperlinks")
			}
		}
	})
	if len(bodies) == 1 && bodies["text/html"] {
		features.add("feature:html-body")
	}
	if subject, ok := msg.rawGet("Subject"); ok && subject != "" {
		texts = append(texts, headerText(subject))
	}
	facts := newObject()
	facts.set("parts", parts)
	return traits, features, strings.Join(texts, "\n"), facts
}

var (
	nonNoteItem = regexp.MustCompile(`(?i)^IPM\.(Contact|Appointment|Task|StickyNote|Activity|Schedule|DistList)`)
	httpLink    = regexp.MustCompile(`https?://`)
)

func declaredCharsetHeader(hdr string) bool {
	if declaredWord.MatchString(hdr) {
		return true
	}
	for i := 0; i+len("charset=") <= len(hdr); i++ {
		if !strings.EqualFold(hdr[i:i+len("charset=")], "charset=") {
			continue
		}
		rest := strings.TrimPrefix(hdr[i+len("charset="):], `"`)
		lower := strings.ToLower(rest)
		if strings.HasPrefix(lower, "utf-8") || strings.HasPrefix(lower, "us-ascii") {
			continue
		}
		if r, _ := utf8.DecodeRuneInString(rest); rest != "" && (r == '-' || isWordRune(r)) {
			return true
		}
	}
	return false
}

func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
}

func rawKey(v any) string {
	return fmt.Sprintf("%T:%v", v, v)
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func mailRecords(rows []record) (set, set, *object) {
	traits, features := set{}, set{}
	facts := newObject()
	messages := 0
	ids := set{}
	folders := set{}
	var refs []string
	top := 0
	for _, rec := range rows {
		r := rec.row
		if !strings.HasPrefix(rec.name, "message") || !(r.has("subject") || r.has("sender") || r.has("transport_headers")) {
			continue
		}
		messages++
		facts.set("messages", messages)
		if r.has("id") {
			top++
		}
		if folder := r.get("folder"); folder != nil {
			folders.add(rawKey(folder))
		}
		cls := str(r.get("class"))
		if !truthy(r.get("class")) {
			cls = "IPM.Note"
		}
		upper := strings.ToUpper(cls)
		if !(strings.HasPrefix(upper, "IPM.NOTE") || strings.HasPrefix(upper, "REPORT.IPM.NOTE")) || strings.HasPrefix(upper, "IPM.NOTE.RULES") {
			if nonNoteItem.MatchString(cls) {
				features.add("feature:non-note-items")
			}
		}
		if truthy(r.get("message_id")) {
			ids.add(str(r.get("message_id")))
		}
		refs = append(refs, str(r.get("in_reply_to"))+" "+str(r.get("references")))
		if declaredCharsetHeader(str(r.get("transport_headers"))) {
			traits.add("trait:charset.declared")
		}
		attachments, _ := r.get("attachments").([]any)
		for _, item := range attachments {
			a, ok := item.(*object)
			if !ok {
				continue
			}
			method, mime := a.str("method"), a.str("mime_type")
			switch {
			case method == "embedded_message" || mime == "message/rfc822":
				features.add("feature:embedded-messages")
			case truthy(a.get("inline")):
				features.add("feature:inline-images")
			case strings.HasPrefix(method, "by_ref"):
				features.add("feature:attachment-by-reference")
			default:
				features.add("feature:attachments")
			}
			if mime == "application/ms-tnef" {
				features.add("feature:tnef-part")
			}
			if strings.Contains(mime, "pkcs7-mime") {
				features.add("feature:smime-opaque")
			}
			name := str(a.get("name"))
			if !truthy(a.get("name")) {
				name = str(a.get("filename"))
			}
			if strings.ToLower(name) == "message.rpmsg" {
				features.add("feature:rights-managed")
			}
		}
		if httpLink.MatchString(str(r.get("markdown"))) {
			features.add("feature:hyperlinks")
		}
	}
	if top > 1 && slices.ContainsFunc(refs, func(ref string) bool {
		for id := range ids {
			if id != "" && strings.Contains(ref, id) {
				return true
			}
		}
		return false
	}) {
		features.add("feature:threads")
	}
	if len(folders) > 1 {
		features.add("feature:folders")
	}
	if messages >= 1000 {
		traits.add("trait:scale.many-messages")
	}
	facts.set("folders_with_messages", len(folders))
	return traits, features, facts
}

func msgBodies(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	prefix := []byte("_\x00_\x00s\x00u\x00b\x00s\x00t\x00g\x001\x00.\x000\x00_\x00")
	found := set{}
	for rest := data; ; {
		i := bytes.Index(rest, prefix)
		if i < 0 {
			break
		}
		rest = rest[i+len(prefix):]
		if len(rest) < 8 {
			continue
		}
		var id []byte
		for k := 0; k < 4; k++ {
			c := rest[2*k]
			if rest[2*k+1] != 0 || !(c >= '0' && c <= '9' || c >= 'A' && c <= 'F') {
				id = nil
				break
			}
			id = append(id, c)
		}
		if id != nil {
			found.add(string(id))
			rest = rest[8:]
		}
	}
	bodies := []string{}
	for _, id := range []string{"1000", "1009", "1013"} {
		if found[id] {
			bodies = append(bodies, id)
		}
	}
	return bodies, nil
}

func splitMbox(data []byte) [][]byte {
	var chunks [][]byte
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' || !bytes.HasPrefix(data[i+1:], []byte("From ")) {
			continue
		}
		if bytes.IndexByte(data[i+1:], '\n') < 0 {
			continue
		}
		chunks = append(chunks, data[start:i])
		start = i + 1
	}
	return append(chunks, data[start:])
}

func mbox(path string) (set, set, string, *object, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, "", nil, err
	}
	chunks := splitMbox(data)
	var ids []string
	var refs []string
	traits, features := set{}, set{}
	var texts []string
	for _, c := range chunks {
		msg := parseMessage(string(c))
		if id, ok := msg.get("Message-ID"); ok && id != "" {
			ids = append(ids, pyStrip(id))
		}
		inReplyTo, _ := msg.get("In-Reply-To")
		references, _ := msg.get("References")
		refs = append(refs, inReplyTo+" "+references)
		t, f, text, _ := message(c)
		traits.merge(t)
		features.merge(f)
		texts = append(texts, text)
	}
	if slices.ContainsFunc(refs, func(ref string) bool {
		return slices.ContainsFunc(ids, func(id string) bool { return strings.Contains(ref, id) })
	}) {
		features.add("feature:threads")
	}
	if len(chunks) >= 1000 {
		traits.add("trait:scale.many-messages")
	}
	facts := newObject()
	facts.set("messages", len(chunks))
	return traits, features, strings.Join(texts, "\n"), facts, nil
}

var (
	noteTag     = regexp.MustCompile(`\[(To Do|Important|Question|Idea|Critical)\]`)
	markdownURL = regexp.MustCompile(`\]\(https?://`)
	equation    = regexp.MustCompile(`\$\$[^$]+\$\$`)
)

func hasTableRow(md string) bool {
	for _, line := range strings.Split(md, "\n") {
		trimmed := strings.TrimRightFunc(line, isPySpace)
		if strings.HasPrefix(line, "|") && len(trimmed) > 1 && strings.HasSuffix(trimmed, "|") {
			return true
		}
	}
	return false
}

func hasNoteTag(md string) bool {
	for _, line := range strings.Split(md, "\n") {
		trimmed := strings.TrimLeftFunc(line, isPySpace)
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ") {
			return true
		}
	}
	return noteTag.MatchString(md)
}

func onenote(rows []record, outcomes set) (set, set, *object) {
	traits, features := set{}, set{}
	facts := newObject()
	sections, pages := 0, 0
	if outcomes[`"encrypted"`] {
		features.add("feature:protected-section")
	}
	for _, rec := range rows {
		r := rec.row
		if strings.HasPrefix(rec.name, "section") {
			sections++
			facts.set("sections", sections)
			if path := r.get("path"); len(asList(path)) > 1 || utf8.RuneCountInString(str(path)) > 1 {
				features.add("feature:section-groups")
			}
		}
		if !strings.HasPrefix(rec.name, "page") {
			continue
		}
		pages++
		facts.set("pages", pages)
		md := str(r.get("markdown"))
		if hasTableRow(md) {
			features.add("feature:tables")
		}
		if hasNoteTag(md) {
			features.add("feature:note-tags")
		}
		if markdownURL.MatchString(md) {
			features.add("feature:hyperlinks")
		}
		if equation.MatchString(md) {
			features.add("feature:equations")
		}
		if truthy(r.get("ink_alternatives")) {
			traits.add("trait:scan.handwriting")
		}
		if truthy(r.get("revisions")) {
			features.add("feature:page-revisions")
		}
		media, _ := r.get("media").([]any)
		for _, item := range media {
			m, ok := item.(*object)
			if !ok {
				continue
			}
			switch m.str("kind") {
			case "image":
				features.add("feature:pictures")
			case "ink":
				features.add("feature:ink")
			case "embedded_file":
				features.add("feature:embedded-objects")
			}
			if truthy(m.get("ocr")) || truthy(m.get("text")) {
				features.add("feature:image-ocr")
			}
		}
	}
	return traits, features, facts
}

func asList(v any) []any {
	list, _ := v.([]any)
	return list
}
