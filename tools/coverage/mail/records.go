package mail

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
)

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

// Records places the traits and features of the mail docproc's run recorded for a .msg or .pst
// file, from its record rows, and returns its facts: the messages and the folders holding them.
func Records(rows []run.Row) (traits, features idset.Set, facts *jsonvalue.Object) {
	traits, features = idset.Set{}, idset.Set{}
	facts = jsonvalue.NewObject()
	messages := 0
	ids := idset.Set{}
	folders := idset.Set{}
	var refs []string
	top := 0
	for _, rec := range rows {
		r := rec.Fields
		if !strings.HasPrefix(rec.Name, "message") || !r.Has("subject") && !r.Has("sender") && !r.Has("transport_headers") {
			continue
		}
		messages++
		facts.Set("messages", messages)
		if r.Has("id") {
			top++
		}
		if folder := r.Get("folder"); folder != nil {
			folders.Add(rawKey(folder))
		}
		cls := jsonvalue.String(r.Get("class"))
		if !jsonvalue.Truthy(r.Get("class")) {
			cls = "IPM.Note"
		}
		upper := strings.ToUpper(cls)
		if !strings.HasPrefix(upper, "IPM.NOTE") && !strings.HasPrefix(upper, "REPORT.IPM.NOTE") || strings.HasPrefix(upper, "IPM.NOTE.RULES") {
			if nonNoteItem.MatchString(cls) {
				features.Add("feature:non-note-items")
			}
		}
		if jsonvalue.Truthy(r.Get("message_id")) {
			ids.Add(jsonvalue.String(r.Get("message_id")))
		}
		refs = append(refs, jsonvalue.String(r.Get("in_reply_to"))+" "+jsonvalue.String(r.Get("references")))
		if declaredCharsetHeader(jsonvalue.String(r.Get("transport_headers"))) {
			traits.Add("trait:charset.declared")
		}
		attachments, _ := r.Get("attachments").([]any)
		for _, item := range attachments {
			a, ok := item.(*jsonvalue.Object)
			if !ok {
				continue
			}
			method, mime := a.Str("method"), a.Str("mime_type")
			switch {
			case method == "embedded_message" || mime == "message/rfc822":
				features.Add("feature:embedded-messages")
			case jsonvalue.Truthy(a.Get("inline")):
				features.Add("feature:inline-images")
			case strings.HasPrefix(method, "by_ref"):
				features.Add("feature:attachment-by-reference")
			default:
				features.Add("feature:attachments")
			}
			if mime == "application/ms-tnef" {
				features.Add("feature:tnef-part")
			}
			if strings.Contains(mime, "pkcs7-mime") {
				features.Add("feature:smime-opaque")
			}
			name := jsonvalue.String(a.Get("name"))
			if !jsonvalue.Truthy(a.Get("name")) {
				name = jsonvalue.String(a.Get("filename"))
			}
			if strings.ToLower(name) == "message.rpmsg" {
				features.Add("feature:rights-managed")
			}
		}
		if httpLink.MatchString(jsonvalue.String(r.Get("markdown"))) {
			features.Add("feature:hyperlinks")
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
		features.Add("feature:threads")
	}
	if len(folders) > 1 {
		features.Add("feature:folders")
	}
	if messages >= 1000 {
		traits.Add("trait:scale.many-messages")
	}
	facts.Set("folders_with_messages", len(folders))
	return traits, features, facts
}
