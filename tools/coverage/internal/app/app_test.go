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

func TestDisposeWritesSortedIDsEmptyListsAndUnescapedText(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"vocabulary": {
		"library": {"kind:z": {"identifier": "Z", "link": "l"}, "kind:a": {"identifier": "A", "link": "l"}},
		"trait": {}, "feature": {"feature:b": {"rule": "x"}, "feature:a": {"rule": "Q&A > 1"}}},
		"library_placement": {"run": "r", "command": "c"},
		"corpus_placement": {"run": "r2", "docproc": "v", "docproc_command": "d", "command": "c2"},
		"samples": [{"path": "a", "sha256": "s", "library": ["kind:a"], "traits": [], "features": [],
			"judged": {"feature:b": "seen", "feature:a": "Q&A <b>"}, "format": "f", "bytes": 1,
			"license_status": "n/a", "redistributable": true, "truth": [], "docs": [], "disposition": "trim", "trimmed_for": "z"}],
		"gaps": [], "docs_without_sample": []}`)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Dispose(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	want := `{
  "vocabulary": {
    "library": {
      "kind:a": {
        "identifier": "A",
        "link": "l"
      },
      "kind:z": {
        "identifier": "Z",
        "link": "l"
      }
    },
    "trait": {},
    "feature": {
      "feature:a": {
        "rule": "Q&A > 1"
      },
      "feature:b": {
        "rule": "x"
      }
    }
  },
  "library_placement": {
    "run": "r",
    "command": "c"
  },
  "corpus_placement": {
    "run": "r2",
    "docproc": "v",
    "docproc_command": "d",
    "command": "c2"
  },
  "samples": [
    {
      "path": "a",
      "sha256": "s",
      "library": [
        "kind:a"
      ],
      "traits": [],
      "features": [],
      "judged": {
        "feature:a": "Q&A <b>",
        "feature:b": "seen"
      },
      "format": "f",
      "bytes": 1,
      "license_status": "n/a",
      "redistributable": true,
      "truth": [],
      "docs": [],
      "disposition": "adopt"
    }
  ],
  "gaps": [],
  "docs_without_sample": []
}
`
	if got := stdout.String(); got != want {
		t.Errorf("Dispose wrote\n%s\nwant\n%s", got, want)
	}
}

func TestDisposeRefusesAKeyItDoesNotKnow(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [], "gaps": [], "extra": 1}`)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Dispose(nil, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "extra") {
		t.Errorf("Dispose = %v, want an error naming the unknown key", err)
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

func TestPlaceTraitsAcceptsFactsAndCorpus(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for _, args := range [][]string{{"--facts"}, {"--corpus"}, {"--corpus", "--facts"}} {
		var stdout, stderr bytes.Buffer
		err := app.PlaceTraits(ctx, args, &stdout, &stderr)
		if stderr.Len() != 0 || err != nil && strings.Contains(err.Error(), "usage") {
			t.Errorf("PlaceTraits(%v) = %v with stderr %q, want it accepted", args, err, stderr.String())
		}
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

func TestManifestWritesTheInventoryAndCopiesItsFiles(t *testing.T) {
	dir, docs := t.TempDir(), t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [{"path": "a/s.txt", "sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"library": ["kind:x"], "traits": [], "features": [], "format": "text/plain", "bytes": 3, "license_status": "reported",
		"redistributable": true, "truth": [], "docs": [], "disposition": "adopt"}], "gaps": []}`)
	write(t, filepath.Join(dir, "docs", "license-readings.json"), `[{"path": "a/s.txt", "license_url": "https://example.org/l", "license_read": "2026-11-25"}]`)
	write(t, filepath.Join(docs, "MANIFEST.json"), `{"files": [{"path": "a/s.txt", "sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"bytes": 3, "format": "text/plain", "source": "s", "source_url": "https://example.org/s?a&b", "license": "CC0-1.0",
		"license_status": "reported", "redistributable": true, "role": "sample", "truth": "none: a timing input"}]}`)
	write(t, filepath.Join(docs, "a", "s.txt"), "abc")
	t.Setenv("DOCS_REFERENCE", docs)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Manifest([]string{"--copy"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	want := `{
  "version": 1,
  "files": [
    {
      "path": "a/s.txt",
      "role": "sample",
      "sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
      "bytes": 3,
      "format": "text/plain",
      "covers": [
        "kind:x"
      ],
      "truth": [],
      "truth_note": "a timing input",
      "docs": [],
      "source": "s",
      "url": "https://example.org/s?a&b",
      "license": "CC0-1.0",
      "license_url": "https://example.org/l",
      "license_read": "2026-11-25",
      "license_status": "verified",
      "redistributable": true
    }
  ]
}
`
	if got := stdout.String(); got != want {
		t.Errorf("Manifest wrote\n%s\nwant\n%s", got, want)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "a", "s.txt")); err != nil || string(data) != "abc" {
		t.Errorf("a/s.txt = %q, %v, want it copied", data, err)
	}
}

func TestManifestRefusesArgumentsAndAMissingReference(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [], "gaps": []}`)
	write(t, filepath.Join(dir, "docs", "license-readings.json"), `[]`)
	t.Setenv("DOCS_REFERENCE", filepath.Join(dir, "absent"))
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Manifest([]string{"extra"}, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "usage: manifest") {
		t.Errorf("Manifest(extra) = %v, want its usage", err)
	}
	if err := app.Manifest(nil, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "DOCS_REFERENCE") {
		t.Errorf("Manifest without a reference = %v, want DOCS_REFERENCE named", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("Manifest wrote %q despite its error", stdout.String())
	}
}
