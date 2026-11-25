package app_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/standards-lab/docs-corpus/tools/coverage/internal/app"
)

func write(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDisposeWritesTheDisposedData(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [{"path": "a", "library": ["kind:x"], "traits": [], "features": [], `+
		`"bytes": 1, "license_status": "n/a", "redistributable": true, "truth": [], "docs": []}], `+
		`"gaps": [{"id": "spent", "covers": ["kind:x"], "candidate": null}]}`)
	placement := filepath.Join(dir, "placement.json")
	write(t, placement, `[{"path": "a", "traits": [], "features": ["feature:f"]}]`)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Dispose([]string{"--placement", placement}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"feature:f"`) || !strings.HasSuffix(stdout.String(), "}\n") || !strings.Contains(stdout.String(), `"disposition": "adopt"`) {
		t.Errorf("stdout = %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "gap spent") {
		t.Errorf("stderr = %q, want gap spent named", stderr.String())
	}
	stdout.Reset()
	if err := app.Dispose([]string{"extra"}, &stdout, &stderr); err == nil {
		t.Error("Dispose with an argument = nil, want its usage")
	}
	if stdout.Len() != 0 {
		t.Errorf("Dispose wrote %q despite its error", stdout.String())
	}
}

func TestPlaceTraitsRefusesOtherArguments(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var stdout, stderr bytes.Buffer
	if err := app.PlaceTraits(ctx, []string{"--bogus"}, &stdout, &stderr); err == nil || !strings.Contains(stderr.String(), "-bogus") {
		t.Errorf("PlaceTraits(--bogus) = %v with stderr %q, want the flag refused", err, stderr.String())
	}
	stderr.Reset()
	if err := app.PlaceTraits(ctx, []string{"extra"}, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "usage: place-traits") {
		t.Errorf("PlaceTraits(extra) = %v, want its usage", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("PlaceTraits wrote %q despite its error", stdout.String())
	}
}

func TestPlaceTraitsAcceptsFacts(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var stdout, stderr bytes.Buffer
	err := app.PlaceTraits(ctx, []string{"--facts"}, &stdout, &stderr)
	if stderr.Len() != 0 || err != nil && strings.Contains(err.Error(), "usage") {
		t.Errorf("PlaceTraits(--facts) = %v with stderr %q, want --facts accepted", err, stderr.String())
	}
}

const mixed = "MIME-Version: 1.0\nContent-Type: multipart/mixed; boundary=\"b\"\n\n--b\nContent-Type: text/plain\n\nhi\n--b--\n"

func TestGenerateWritesTheCandidate(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "a.eml")
	payload := filepath.Join(dir, "p.txt")
	out := filepath.Join(dir, "out.eml")
	write(t, in, mixed)
	write(t, payload, "abc")
	if err := app.Generate([]string{"rights-managed", in, payload, out}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(out); err != nil || !bytes.Contains(data, []byte("YWJj")) {
		t.Errorf("out.eml = %q, %v", data, err)
	}
	mbox := filepath.Join(dir, "out.mbox")
	if err := app.Generate([]string{"mbox", mbox, in}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(mbox); err != nil || bytes.Count(data, []byte("\nFrom ")) != 999 {
		t.Errorf("out.mbox has %d From lines after the first, want 999: %v", bytes.Count(data, []byte("\nFrom ")), err)
	}
}

func TestGenerateRefusesBadArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"rights-managed", "a"}, {"mbox", "out"}, {"zip"}} {
		if err := app.Generate(args, time.Now()); err == nil {
			t.Errorf("Generate(%q) succeeded", args)
		}
	}
}
