package stdmail_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
	"github.com/standards-lab/docs-corpus/tools/coverage/stdmail"
)

func message(lines ...string) []byte {
	return []byte(strings.Join(lines, "\r\n"))
}

func types(parts []mail.Part) []string {
	var out []string
	for _, p := range parts {
		out = append(out, p.ContentType())
	}
	return out
}

func TestParseWalksNestedMultipartsDepthFirst(t *testing.T) {
	parts := stdmail.Parser{}.Parse(message(
		"Subject: nested",
		"Content-Type: multipart/mixed; boundary=outer",
		"",
		"preamble",
		"--outer",
		"Content-Type: multipart/alternative; boundary=\"inner\"",
		"",
		"--inner",
		"Content-Type: text/plain",
		"",
		"plain",
		"--inner",
		"Content-Type: text/html",
		"",
		"<p>html</p>",
		"--inner--",
		"--outer",
		"Content-Type: message/rfc822",
		"",
		"Subject: attached",
		"",
		"inner body",
		"--outer",
		"Content-Type: multipart/digest; boundary=d",
		"",
		"--d",
		"",
		"Subject: digested",
		"",
		"digest body",
		"--d--",
		"--outer--",
		"epilogue"))
	want := []string{"multipart/mixed", "multipart/alternative", "text/plain", "text/html",
		"message/rfc822", "text/plain", "multipart/digest", "message/rfc822", "text/plain"}
	if got := types(parts); !slices.Equal(got, want) {
		t.Fatalf("types = %q, want %q", got, want)
	}
	if got := string(parts[2].Payload()); got != "plain" {
		t.Errorf("text/plain payload = %q, want the body without its delimiter's line break", got)
	}
	if parts[0].Payload() != nil || parts[4].Payload() != nil {
		t.Error("a part that holds others has a payload")
	}
	if subject, ok := parts[5].Header("subject"); !ok || subject != "attached" {
		t.Errorf("the embedded message's Subject = %q, %v", subject, ok)
	}
	if got := string(parts[8].Payload()); got != "digest body" {
		t.Errorf("the digested message's body = %q", got)
	}
}

func TestParseKeepsAMultipartWithoutItsBoundaryWhole(t *testing.T) {
	for name, data := range map[string][]byte{
		"no boundary parameter": message("Content-Type: multipart/mixed", "", "--x", "", "body", "--x--"),
		"boundary never seen":   message("Content-Type: multipart/mixed; boundary=b", "", "--x", "", "body", "--x--"),
	} {
		parts := stdmail.Parser{}.Parse(data)
		if len(parts) != 1 || parts[0].ContentType() != "multipart/mixed" {
			t.Errorf("%s: types = %q, want the multipart alone", name, types(parts))
			continue
		}
		if got := string(parts[0].Payload()); !strings.Contains(got, "body") {
			t.Errorf("%s: payload = %q, want its body", name, got)
		}
	}
}

func TestParseReadsAnUnclosedMultipart(t *testing.T) {
	parts := stdmail.Parser{}.Parse(message(
		"Content-Type: multipart/mixed; boundary=b", "", "--b", "Content-Type: text/plain", "", "one",
		"--b", "Content-Type: image/png", "Content-ID: <x>", "", "png"))
	if got := types(parts); !slices.Equal(got, []string{"multipart/mixed", "text/plain", "image/png"}) {
		t.Errorf("types = %q, want both parts", got)
	}
}

func TestDecodeHeaderDecodesEncodedWords(t *testing.T) {
	for value, want := range map[string]string{
		"=?utf-8?b?0J/RgNC40LLQtdGC?=":                      "Привет",
		"=?iso-8859-1?q?caf=E9?= =?ISO-8859-1?Q?_cr=E8me?=": "café crème",
		"=?windows-1251?b?z/Do4uXy?=":                       "Привет",
		"=?koi8-r?q?=F0=D2=C9=D7=C5=D4?=":                   "Привет",
		"=?shift_jis?b?k/qWe4zq?=":                          "日本語",
		"plain =?no-such-charset?q?x?=":                     "plain =?no-such-charset?q?x?=",
	} {
		if got := (stdmail.Parser{}).DecodeHeader(value); got != want {
			t.Errorf("DecodeHeader(%q) = %q, want %q", value, got, want)
		}
	}
}

func TestTextDecodesTheCharsetAndTransferEncoding(t *testing.T) {
	parts := stdmail.Parser{}.Parse(message(
		"Content-Type: multipart/mixed; boundary=b", "",
		"--b", "Content-Type: text/plain; charset=ISO-8859-1", "Content-Transfer-Encoding: quoted-printable", "",
		"caf=E9 =", "cr=E8me",
		"--b", "Content-Type: text/plain; charset=\"windows-1251\"", "Content-Transfer-Encoding: base64", "",
		"z/Do4uXy",
		"--b", "Content-Type: text/plain; charset=gb2312", "Content-Transfer-Encoding: base64", "",
		"1tDOxA",
		"--b", "Content-Type: text/plain; charset=x-unknown", "", "as is",
		"--b--"))
	want := []string{"", "café crème", "Привет", "中文", "as is"}
	for i, p := range parts {
		if got := p.Text(); got != want[i] {
			t.Errorf("part %d Text = %q, want %q", i, got, want[i])
		}
	}
	if got := parts[2].Charset(); got != "windows-1251" {
		t.Errorf("Charset = %q, want windows-1251", got)
	}
}

func TestParamsAndFilenames(t *testing.T) {
	parts := stdmail.Parser{}.Parse(message(
		"Content-Type: multipart/signed; protocol=\"application/pkcs7-signature\"; boundary=b", "",
		"--b", "Content-Type: application/octet-stream; name=\"from-type.bin\"", "", "x",
		"--b", "Content-Type: application/octet-stream", "Content-Disposition: attachment;",
		" filename*0*=utf-8''r%C3%A9; filename*1*=sum%C3%A9.txt", "", "y",
		"--b--"))
	if protocol, ok := parts[0].Param("content-type", "protocol"); !ok || protocol != "application/pkcs7-signature" {
		t.Errorf("protocol = %q, %v", protocol, ok)
	}
	if got := parts[1].Filename(); got != "from-type.bin" {
		t.Errorf("Filename = %q, want the Content-Type's name", got)
	}
	if got := parts[2].Filename(); got != "résumé.txt" {
		t.Errorf("Filename = %q, want the RFC 2231 continuation decoded", got)
	}
}

func TestParseSkipsTheMboxFromLineAndEndsTheHeaderAtABadLine(t *testing.T) {
	parts := stdmail.Parser{}.Parse([]byte("From a@example.org Mon Jan  1 00:00:00 2024\nSubject: one\n two\nnot a header\nbody\n"))
	if subject, ok := parts[0].Header("Subject"); !ok || subject != "one two" {
		t.Errorf("Subject = %q, %v, want the folded field unfolded", subject, ok)
	}
	if got := string(parts[0].Payload()); got != "not a header\nbody\n" {
		t.Errorf("payload = %q, want the body from the first line that is not a field", got)
	}
	if got := parts[0].ContentType(); got != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain by default", got)
	}
}

func TestContentTypeFallsBackToTextPlainWhenInvalid(t *testing.T) {
	for _, ctype := range []string{"not a type", "text; charset=us-ascii", "text/"} {
		parts := stdmail.Parser{}.Parse(message("Content-Type: "+ctype, "", "x"))
		if got := parts[0].ContentType(); got != "text/plain" {
			t.Errorf("Content-Type: %s gives %q, want text/plain", ctype, got)
		}
	}
}
