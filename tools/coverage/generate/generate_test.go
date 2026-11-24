package generate_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/standards-lab/docs-corpus/tools/coverage/generate"
	"github.com/standards-lab/docs-corpus/tools/coverage/internal/docstest"
	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
	"github.com/standards-lab/docs-corpus/tools/coverage/stdmail"
)

const mixed = "MIME-Version: 1.0\nContent-Type: multipart/mixed; boundary=\"b\"\n\n--b\nContent-Type: text/plain\n\nhi\n--b--\n"

func TestRightsManagedAddsTheHeaderAndAttachment(t *testing.T) {
	got, err := generate.RightsManaged([]byte(mixed), []byte("abc"))
	if err != nil {
		t.Fatal(err)
	}
	want := "MIME-Version: 1.0\nContent-Type: multipart/mixed; boundary=\"b\"\nContent-Class: rpmsg.message\n" +
		"\n--b\nContent-Type: text/plain\n\nhi\n" +
		"--b\nContent-Type: application/x-microsoft-rpmsg-message\nContent-Transfer-Encoding: base64\n" +
		"Content-Disposition: attachment; filename=\"message.rpmsg\"\nMIME-Version: 1.0\n\nYWJj\n" +
		"--b--\n"
	if string(got) != want {
		t.Errorf("RightsManaged =\n%q\nwant\n%q", got, want)
	}
}

func TestRightsManagedKeepsCRLFAndWrapsBase64(t *testing.T) {
	got, err := generate.RightsManaged([]byte(strings.ReplaceAll(mixed, "\n", "\r\n")), bytes.Repeat([]byte("x"), 100))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(bytes.ReplaceAll(got, []byte("\r\n"), nil), []byte("\n")) {
		t.Errorf("a CRLF message gained a bare LF: %q", got)
	}
	var widest int
	for line := range strings.SplitSeq(string(got), "\r\n") {
		widest = max(widest, len(line))
	}
	if widest != 76 {
		t.Errorf("the widest line is %d bytes, want base64 wrapped at 76", widest)
	}
}

func TestRightsManagedRefusesAMessageItCannotExtend(t *testing.T) {
	for name, msg := range map[string]string{
		"not multipart/mixed": "Content-Type: text/plain\n\nhi\n",
		"no closing boundary": "Content-Type: multipart/mixed; boundary=b\n\n--b\n\nhi\n",
		"no header":           "",
	} {
		if _, err := generate.RightsManaged([]byte(msg), nil); err == nil {
			t.Errorf("%s: RightsManaged succeeded", name)
		}
	}
}

func TestMboxRepeatsTheMessagesToAThousand(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("east", 5*3600))
	out := string(generate.Mbox([][]byte{
		[]byte("From kept@example.org Mon Jan  1 00:00:00 2024\nSubject: a\n\nFrom here\n"),
		[]byte("Subject: b\n\nno newline"),
	}, now))
	if got := strings.Count(out, "\nFrom ") + 1; got != 1000 {
		t.Errorf("the mbox has %d From lines, want 1000", got)
	}
	if got := strings.Count(out, "From kept@example.org Mon Jan  1 00:00:00 2024\n"); got != 500 {
		t.Errorf("the first message's own From line appears %d times, want 500", got)
	}
	if got := strings.Count(out, "From MAILER-DAEMON Thu Jan  1 22:04:05 2026\n"); got != 500 {
		t.Errorf("the generated From line, in UTC, appears %d times, want 500", got)
	}
	if !strings.Contains(out, "\n>From here\n") || !strings.Contains(out, "no newline\n\n") {
		t.Errorf("the mbox doesn't escape From or end each message with a blank line: %.300q", out)
	}
}

func TestRunWritesTheCandidate(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "a.eml")
	payload := filepath.Join(dir, "p.txt")
	out := filepath.Join(dir, "out.eml")
	for path, data := range map[string]string{in: mixed, payload: "abc"} {
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := generate.Run([]string{"rights-managed", in, payload, out}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(out); err != nil || !bytes.Contains(data, []byte("YWJj")) {
		t.Errorf("out.eml = %q, %v", data, err)
	}
	mbox := filepath.Join(dir, "out.mbox")
	if err := generate.Run([]string{"mbox", mbox, in}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(mbox); err != nil {
		t.Error(err)
	}
}

func TestRunRefusesBadArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"rights-managed", "a"}, {"mbox", "out"}, {"zip"}} {
		if err := generate.Run(args, time.Now()); err == nil {
			t.Errorf("Run(%q) succeeded", args)
		}
	}
}

func TestTheRightsManagedCandidatePlacesItsFeature(t *testing.T) {
	message, err := os.ReadFile(docstest.Path(t, "email/eml/cpython/msg_36.eml"))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(docstest.Path(t, "text/encodings/windows-1252.txt"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := generate.RightsManaged(message, payload)
	if err != nil {
		t.Fatal(err)
	}
	_, features, _, _ := mail.Message(stdmail.Parser{}, out)
	if !features["feature:rights-managed"] {
		t.Errorf("the candidate's features = %v, want rights-managed", features.Sorted(""))
	}
	if _, before, _, _ := mail.Message(stdmail.Parser{}, message); before["feature:rights-managed"] {
		t.Error("msg_36.eml is rights-managed before the generator runs")
	}
}
