package mail

import (
	"bytes"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
	"github.com/standards-lab/docs-corpus/tools/coverage/text"
)

var (
	declaredWord = regexp.MustCompile(`=\?[^?]+\?[bBqQ]\?`)
	htmlInline   = regexp.MustCompile(`(?i)<img[^>]+src=["']?(data:|cid:)`)
	htmlLink     = regexp.MustCompile(`(?i)<a[^>]+href=["']?https?://`)
)

func hasPlainLink(s string) bool {
	for _, scheme := range []string{"http://", "https://"} {
		rest := s
		for {
			i := strings.Index(rest, scheme)
			if i < 0 {
				break
			}
			rest = rest[i+len(scheme):]
			if r, _ := utf8.DecodeRuneInString(rest); rest != "" && !text.IsSpace(r) {
				return true
			}
		}
	}
	return false
}

// Message places an RFC 5322 message's traits and features, parsed by parser, and returns its
// text, each text/plain and text/html body that is not an attachment and its decoded Subject, and
// its facts, the number of parts.
func Message(parser Parser, data []byte) (traits, features idset.Set, body string, facts *jsonvalue.Object) {
	all := parser.Parse(data)
	traits, features = idset.Set{}, idset.Set{}
	var texts []string
	parts := 0
	head, _, _ := bytes.Cut(data, []byte("\n\n"))
	if declaredWord.Match(head) {
		traits.Add("trait:charset.declared")
	}
	bodies := idset.Set{}
	for i, p := range all {
		parts++
		ctype := p.ContentType()
		cs := p.Charset()
		if !text.UTF8Name(cs) {
			traits.Add("trait:charset.declared")
		}
		disp, _ := p.Header("Content-Disposition")
		disp = strings.ToLower(disp)
		filename := strings.ToLower(p.Filename())
		if ctype == "message/rfc822" || (strings.HasPrefix(ctype, "message/") && i > 0 && ctype != "message/delivery-status") {
			features.Add("feature:embedded-messages")
		}
		if ctype == "application/ms-tnef" || filename == "winmail.dat" {
			features.Add("feature:tnef-part")
		}
		if ctype == "application/pkcs7-mime" || ctype == "application/x-pkcs7-mime" {
			features.Add("feature:smime-opaque")
		}
		if protocol, _ := p.Param("content-type", "protocol"); ctype == "multipart/signed" && strings.Contains(protocol, "pkcs7") {
			features.Add("feature:smime-signed")
		}
		if filename == "message.rpmsg" {
			features.Add("feature:rights-managed")
		}
		contentID, hasContentID := p.Header("Content-ID")
		if contentID != "" && strings.HasPrefix(ctype, "image/") {
			features.Add("feature:inline-images")
		}
		if strings.Contains(disp, "attachment") || (p.Filename() != "" && !strings.HasPrefix(ctype, "multipart/") && !hasContentID) {
			features.Add("feature:attachments")
		}
		if (ctype == "text/plain" || ctype == "text/html") && !strings.Contains(disp, "attachment") {
			bodies.Add(ctype)
			decoded := p.Text()
			texts = append(texts, decoded)
			if ctype == "text/html" {
				if htmlInline.MatchString(decoded) {
					features.Add("feature:inline-images")
				}
				if htmlLink.MatchString(decoded) {
					features.Add("feature:hyperlinks")
				}
			} else if hasPlainLink(decoded) {
				features.Add("feature:hyperlinks")
			}
		}
	}
	if len(bodies) == 1 && bodies["text/html"] {
		features.Add("feature:html-body")
	}
	if subject, ok := all[0].Header("Subject"); ok && subject != "" {
		texts = append(texts, parser.DecodeHeader(subject))
	}
	facts = jsonvalue.NewObject()
	facts.Set("parts", parts)
	return traits, features, strings.Join(texts, "\n"), facts
}

func splitMbox(data []byte) [][]byte {
	var chunks [][]byte
	start := 0
	for i := range data {
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

// Mbox places an mbox's traits and features: each message's, as Message places them, and threads
// and scale across them. It returns the messages' text, joined by newlines, and its facts, the
// number of messages.
func Mbox(parser Parser, data []byte) (traits, features idset.Set, body string, facts *jsonvalue.Object) {
	chunks := splitMbox(data)
	var ids []string
	var refs []string
	traits, features = idset.Set{}, idset.Set{}
	var texts []string
	for _, c := range chunks {
		msg := parser.Parse(c)[0]
		if id, ok := msg.Header("Message-ID"); ok && id != "" {
			ids = append(ids, text.Strip(id))
		}
		inReplyTo, _ := msg.Header("In-Reply-To")
		references, _ := msg.Header("References")
		refs = append(refs, inReplyTo+" "+references)
		t, f, body, _ := Message(parser, c)
		traits.Merge(t)
		features.Merge(f)
		texts = append(texts, body)
	}
	if slices.ContainsFunc(refs, func(ref string) bool {
		return slices.ContainsFunc(ids, func(id string) bool { return strings.Contains(ref, id) })
	}) {
		features.Add("feature:threads")
	}
	if len(chunks) >= 1000 {
		traits.Add("trait:scale.many-messages")
	}
	facts = jsonvalue.NewObject()
	facts.Set("messages", len(chunks))
	return traits, features, strings.Join(texts, "\n"), facts
}
