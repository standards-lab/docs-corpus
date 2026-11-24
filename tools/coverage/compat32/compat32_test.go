package compat32_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/compat32"
	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
)

func parse(lines ...string) []mail.Part {
	return compat32.Parser{}.Parse([]byte(strings.Join(lines, "\n")))
}

func types(parts []mail.Part) []string {
	var out []string
	for _, p := range parts {
		out = append(out, p.ContentType())
	}
	return out
}

func TestParseWalksPartsDepthFirst(t *testing.T) {
	parts := parse(
		"Content-Type: multipart/mixed; boundary=outer",
		"",
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
		"Subject: inner",
		"",
		"body",
		"--outer--",
		"")
	want := []string{"multipart/mixed", "multipart/alternative", "text/plain", "text/html", "message/rfc822", "text/plain"}
	if got := types(parts); !slices.Equal(got, want) {
		t.Fatalf("parts = %v, want %v", got, want)
	}
	if parts[0].Payload() != nil {
		t.Errorf("a multipart part's payload = %q, want nil", parts[0].Payload())
	}
	if got := string(parts[2].Payload()); got != "plain" {
		t.Errorf("the text/plain payload = %q, want its line without the boundary's newline", got)
	}
	if got, _ := parts[5].Header("Subject"); got != "inner" {
		t.Errorf("the attached message's Subject = %q", got)
	}
}

func TestParseKeepsPartsWithoutAClosingBoundary(t *testing.T) {
	parts := parse(
		"Content-Type: multipart/mixed; boundary=b",
		"",
		"--b",
		"Content-Type: text/plain",
		"",
		"one",
		"--b",
		"",
		"two")
	if got := types(parts); !slices.Equal(got, []string{"multipart/mixed", "text/plain", "text/plain"}) {
		t.Errorf("parts = %v", got)
	}
}

func TestContentTypeDefaultsAsCompat32Does(t *testing.T) {
	if got := parse("Subject: x", "", "body")[0].ContentType(); got != "text/plain" {
		t.Errorf("no Content-Type gives %q, want text/plain", got)
	}
	if got := parse("Content-Type: text", "", "body")[0].ContentType(); got != "text/plain" {
		t.Errorf("a malformed Content-Type gives %q, want text/plain", got)
	}
	digest := parse(
		"Content-Type: multipart/digest; boundary=d",
		"",
		"--d",
		"",
		"Subject: one",
		"",
		"body",
		"--d--")
	if got := digest[1].ContentType(); got != "message/rfc822" {
		t.Errorf("a digest's part gives %q, want message/rfc822", got)
	}
}

func TestPayloadDecodesTransferEncodings(t *testing.T) {
	for _, c := range []struct {
		cte, body, want string
	}{
		{"quoted-printable", "caf=C3=A9 =\nau lait=3D", "café au lait="},
		{"base64", "Y2Fm\nw6k=", "café"},
		{"base64", "Y2Fmw6k", "café"},
		{"x-uuencode", "begin 644 a.txt\n#86)C\n`\nend", "abc"},
		{"7bit", "as is", "as is"},
	} {
		parts := parse("Content-Transfer-Encoding: "+c.cte, "", c.body)
		if got := string(parts[0].Payload()); got != c.want {
			t.Errorf("%s payload %q = %q, want %q", c.cte, c.body, got, c.want)
		}
	}
}

func TestParamsAndFilename(t *testing.T) {
	p := parse(
		`Content-Type: multipart/signed; protocol="application/pkcs7-signature"; charset="UTF-8"; boundary=s`,
		`Content-Disposition: attachment; filename*=utf-8''%E2%82%AC%20rates.pdf`,
		"",
		"--s--")[0]
	if got, ok := p.Param("content-type", "protocol"); !ok || got != "application/pkcs7-signature" {
		t.Errorf("Param(protocol) = %q, %v", got, ok)
	}
	if _, ok := p.Param("content-type", "missing"); ok {
		t.Error("Param found a parameter the header doesn't have")
	}
	if got := p.Charset(); got != "utf-8" {
		t.Errorf("Charset = %q, want utf-8", got)
	}
	if got := p.Filename(); got != "€ rates.pdf" {
		t.Errorf("Filename = %q, want the RFC 2231 value decoded", got)
	}
	if got := parse("Content-Type: image/png; name=\" logo.png \"", "", "")[0].Filename(); got != "logo.png" {
		t.Errorf("Filename from the Content-Type's name = %q", got)
	}
}

func TestHeaderReplacesNonASCIIAndRawHeaderKeepsIt(t *testing.T) {
	p := parse("subject: caf\xc3\xa9", "", "")[0]
	if got, ok := p.Header("Subject"); !ok || got != "caf��" {
		t.Errorf("Header = %q, %v", got, ok)
	}
	if got, _ := p.RawHeader("SUBJECT"); got != "caf\xc3\xa9" {
		t.Errorf("RawHeader = %q", got)
	}
	if _, ok := p.Header("To"); ok {
		t.Error("Header found a field the message doesn't have")
	}
}

func TestDecodeHeaderJoinsEncodedWords(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"=?utf-8?q?caf=C3=A9?= =?utf-8?b?w6k=?=", "caféé"},
		{"Re: =?iso-8859-1?q?caf=E9_noir?=", "Re: café noir"},
		{"plain subject", "plain subject"},
		{"caf\xc3\xa9", "caf��"},
	} {
		if got := (compat32.Parser{}).DecodeHeader(c.in); got != c.want {
			t.Errorf("DecodeHeader(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
