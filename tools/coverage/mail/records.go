package mail

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
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

type mailRow struct {
	ID               json.RawMessage `json:"id"`
	Subject          json.RawMessage `json:"subject"`
	Sender           json.RawMessage `json:"sender"`
	TransportHeaders json.RawMessage `json:"transport_headers"`
	Folder           json.RawMessage `json:"folder"`
	Class            string          `json:"class"`
	MessageID        string          `json:"message_id"`
	InReplyTo        string          `json:"in_reply_to"`
	References       string          `json:"references"`
	Markdown         string          `json:"markdown"`
	Attachments      []struct {
		Method   string `json:"method"`
		MimeType string `json:"mime_type"`
		Inline   bool   `json:"inline"`
		Name     string `json:"name"`
		Filename string `json:"filename"`
	} `json:"attachments"`
}

func headers(raw json.RawMessage) string {
	var s string
	_ = json.Unmarshal(raw, &s)
	return s
}

// RecordsFacts is what Records found in the run's mail.
type RecordsFacts struct {
	// Messages is the number of message records.
	Messages int
	// FoldersWithMessages is the number of folders the messages are in.
	FoldersWithMessages int
}

// Records places the traits and features of the mail docproc's run recorded for a .msg or .pst
// file, from its record rows, and returns its facts. A row field of another JSON type than
// docproc's mail records give it reads as absent.
func Records(rows []run.Row) (traits, features idset.Set, facts RecordsFacts) {
	traits, features = idset.Set{}, idset.Set{}
	ids := idset.Set{}
	folders := idset.Set{}
	var refs []string
	top := 0
	for _, rec := range rows {
		var r mailRow
		_ = json.Unmarshal(rec.Fields, &r)
		if !strings.HasPrefix(rec.Name, "message") || r.Subject == nil && r.Sender == nil && r.TransportHeaders == nil {
			continue
		}
		facts.Messages++
		if r.ID != nil {
			top++
		}
		if r.Folder != nil && string(r.Folder) != "null" {
			folders.Add(string(r.Folder))
		}
		cls := r.Class
		if cls == "" {
			cls = "IPM.Note"
		}
		upper := strings.ToUpper(cls)
		if !strings.HasPrefix(upper, "IPM.NOTE") && !strings.HasPrefix(upper, "REPORT.IPM.NOTE") || strings.HasPrefix(upper, "IPM.NOTE.RULES") {
			if nonNoteItem.MatchString(cls) {
				features.Add("feature:non-note-items")
			}
		}
		if r.MessageID != "" {
			ids.Add(r.MessageID)
		}
		refs = append(refs, r.InReplyTo+" "+r.References)
		if declaredCharsetHeader(headers(r.TransportHeaders)) {
			traits.Add("trait:charset.declared")
		}
		for _, a := range r.Attachments {
			switch {
			case a.Method == "embedded_message" || a.MimeType == "message/rfc822":
				features.Add("feature:embedded-messages")
			case a.Inline:
				features.Add("feature:inline-images")
			case strings.HasPrefix(a.Method, "by_ref"):
				features.Add("feature:attachment-by-reference")
			default:
				features.Add("feature:attachments")
			}
			if a.MimeType == "application/ms-tnef" {
				features.Add("feature:tnef-part")
			}
			if strings.Contains(a.MimeType, "pkcs7-mime") {
				features.Add("feature:smime-opaque")
			}
			name := a.Name
			if name == "" {
				name = a.Filename
			}
			if strings.ToLower(name) == "message.rpmsg" {
				features.Add("feature:rights-managed")
			}
		}
		if httpLink.MatchString(r.Markdown) {
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
	if facts.Messages >= 1000 {
		traits.Add("trait:scale.many-messages")
	}
	facts.FoldersWithMessages = len(folders)
	return traits, features, facts
}
