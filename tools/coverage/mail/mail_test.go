package mail_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/standards-lab/docs-corpus/tools/coverage/compat32"
	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/internal/docstest"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
)

type part struct {
	headers  map[string]string
	ctype    string
	params   map[string]string
	filename string
	payload  string
}

func (p part) Header(name string) (string, bool)    { v, ok := p.headers[name]; return v, ok }
func (p part) RawHeader(name string) (string, bool) { return p.Header(name) }
func (p part) ContentType() string                  { return p.ctype }
func (p part) Charset() string                      { return "" }
func (p part) Filename() string                     { return p.filename }
func (p part) Payload() []byte                      { return []byte(p.payload) }
func (p part) Param(header, name string) (string, bool) {
	v, ok := p.params[header+";"+name]
	return v, ok
}

type parser []mail.Part

func (f parser) Parse([]byte) []mail.Part       { return f }
func (parser) DecodeHeader(value string) string { return strings.ToUpper(value) }

func message(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n"))
}

func sorted(s idset.Set) []string {
	return s.Sorted("")
}

func TestMessageReadsOnlyThroughItsParser(t *testing.T) {
	p := parser{
		part{headers: map[string]string{"Subject": "signed"}, ctype: "multipart/signed",
			params: map[string]string{"content-type;protocol": "application/pkcs7-signature"}},
		part{ctype: "text/html", payload: `<a href="https://example.org">x</a>`},
		part{ctype: "application/ms-tnef", filename: "winmail.dat"},
	}
	traits, features, body, facts := mail.Message(p, nil)
	want := []string{"feature:attachments", "feature:html-body", "feature:hyperlinks", "feature:smime-signed", "feature:tnef-part"}
	if got := sorted(features); !slices.Equal(got, want) {
		t.Errorf("features = %v, want %v", got, want)
	}
	if len(traits) != 0 {
		t.Errorf("traits = %v, want none", sorted(traits))
	}
	if body != "<a href=\"https://example.org\">x</a>\nSIGNED" {
		t.Errorf("body = %q, want the HTML body and the subject the parser decoded", body)
	}
	if got := string(jsonvalue.Marshal(facts)); got != "{\n  \"parts\": 3\n}" {
		t.Errorf("facts = %s", got)
	}
}

func TestMessagePlacesWhatItsPartsShow(t *testing.T) {
	traits, features, body, _ := mail.Message(compat32.Parser{}, message(
		"Subject: =?iso-8859-1?q?caf=E9?=",
		"Content-Type: multipart/related; boundary=b",
		"",
		"--b",
		"Content-Type: text/plain; charset=iso-8859-1",
		"Content-Transfer-Encoding: quoted-printable",
		"",
		"see http://example.org caf=E9",
		"--b",
		"Content-Type: image/png",
		"Content-ID: <logo>",
		"",
		"png",
		"--b",
		"Content-Type: application/octet-stream",
		"Content-Disposition: attachment; filename=\"message.rpmsg\"",
		"",
		"rpmsg",
		"--b--",
		""))
	if got := sorted(traits); !slices.Equal(got, []string{"trait:charset.declared"}) {
		t.Errorf("traits = %v, want charset.declared", got)
	}
	want := []string{"feature:attachments", "feature:hyperlinks", "feature:inline-images", "feature:rights-managed"}
	if got := sorted(features); !slices.Equal(got, want) {
		t.Errorf("features = %v, want %v", got, want)
	}
	if body != "see http://example.org café\ncafé" {
		t.Errorf("body = %q", body)
	}
}

func TestMboxFindsThreadsAcrossMessages(t *testing.T) {
	_, features, body, facts := mail.Mbox(compat32.Parser{}, message(
		"From a@example.org Mon Jan  1 00:00:00 2024",
		"Message-ID: <1@example.org>",
		"Subject: first",
		"",
		"hello",
		"From b@example.org Mon Jan  1 00:00:01 2024",
		"In-Reply-To: <1@example.org>",
		"",
		"reply",
		""))
	if !features["feature:threads"] {
		t.Errorf("features = %v, want threads", sorted(features))
	}
	if body != "hello\nfirst\nreply\n" {
		t.Errorf("body = %q", body)
	}
	if got := string(jsonvalue.Marshal(facts)); got != "{\n  \"messages\": 2\n}" {
		t.Errorf("facts = %s", got)
	}
}

func rows(t *testing.T, lines ...string) []run.Row {
	t.Helper()
	var out []run.Row
	for _, l := range lines {
		v, err := jsonvalue.Parse([]byte(l))
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, run.Row{Name: "message 1", Fields: v.(*jsonvalue.Object)})
	}
	return out
}

func TestRecordsPlaceTheRunsMail(t *testing.T) {
	traits, features, facts := mail.Records(rows(t,
		`{"id": 1, "subject": "a", "folder": "Inbox", "message_id": "<1@x>", "transport_headers": "Content-Type: text/plain; charset=\"iso-8859-1\""}`,
		`{"id": 2, "subject": "b", "folder": "Sent", "in_reply_to": "<1@x>", "markdown": "see https://example.org", "transport_headers": "Content-Type: text/plain; charset=utf-8",
		  "attachments": [{"method": "by_reference_only", "name": "a.pdf"}, {"method": "by_value", "mime_type": "application/ms-tnef"}]}`,
		`{"subject": "c", "class": "IPM.Contact", "attachments": [{"method": "embedded_message"}, {"inline": true}, {"filename": "message.rpmsg"}]}`,
		`{"note": "not a message"}`,
	))
	if got := sorted(traits); !slices.Equal(got, []string{"trait:charset.declared"}) {
		t.Errorf("traits = %v", got)
	}
	want := []string{"feature:attachment-by-reference", "feature:attachments", "feature:embedded-messages",
		"feature:folders", "feature:hyperlinks", "feature:inline-images", "feature:non-note-items",
		"feature:rights-managed", "feature:threads", "feature:tnef-part"}
	if got := sorted(features); !slices.Equal(got, want) {
		t.Errorf("features = %v, want %v", got, want)
	}
	if got := string(jsonvalue.Marshal(facts)); got != "{\n  \"messages\": 3,\n  \"folders_with_messages\": 2\n}" {
		t.Errorf("facts = %s", got)
	}
}

func utf16le(s string) string {
	var b strings.Builder
	for _, u := range utf16.Encode([]rune(s)) {
		b.WriteByte(byte(u))
		b.WriteByte(byte(u >> 8))
	}
	return b.String()
}

func TestMSGBodiesListsTheBodyPropertiesInOrder(t *testing.T) {
	data := utf16le("__substg1.0_1013001F") + "\x00junk" + utf16le("__substg1.0_1000001F") + utf16le("__substg1.0_0037001F") + utf16le("__substg1.0_zz")
	if got := mail.MSGBodies([]byte(data)); !slices.Equal(got, []string{"1000", "1013"}) {
		t.Errorf("MSGBodies = %v, want [1000 1013]", got)
	}
	if got := mail.MSGBodies(nil); got == nil || len(got) != 0 {
		t.Errorf("MSGBodies(nil) = %#v, want an empty list", got)
	}
}

func TestRealMessagesPlaceTheirRecordedFeatures(t *testing.T) {
	samples := docstest.Samples(t)
	checked := 0
	for _, s := range samples {
		ext := filepath.Ext(s.Path)
		if ext != ".eml" && ext != ".mbox" {
			continue
		}
		data, err := os.ReadFile(docstest.Path(t, s.Path))
		if err != nil {
			t.Fatal(err)
		}
		var features idset.Set
		if ext == ".eml" {
			_, features, _, _ = mail.Message(compat32.Parser{}, data)
		} else {
			_, features, _, _ = mail.Mbox(compat32.Parser{}, data)
		}
		if got, want := sorted(features), s.Probed(); !slices.Equal(got, want) {
			t.Errorf("%s: features = %v, want %v", s.Path, got, want)
		}
		checked++
	}
	if checked == 0 {
		t.Error("docs/coverage.json lists no .eml or .mbox sample")
	}
}
