package app_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/standards-lab/docs-corpus/tools/coverage/internal/app"
	"github.com/standards-lab/docs-corpus/tools/coverage/internal/docstest"
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

func noCandidates(t *testing.T, dir string) {
	t.Helper()
	write(t, filepath.Join(dir, "docs", "candidates.json"), `[]`)
	write(t, filepath.Join(dir, "docs", "corpus-placement.json"), `[]`)
	write(t, filepath.Join(dir, "docs", "measurements.json"), `[]`)
	write(t, filepath.Join(dir, "docs", "sourced.json"), `[]`)
}

func TestDisposeWritesTheDisposedData(t *testing.T) {
	dir := t.TempDir()
	noCandidates(t, dir)
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [{"path": "a", "library": ["kind:x"], "traits": [], "features": ["feature:f"], `+
		`"bytes": 1, "license_status": "n/a", "redistributable": true, "truth": [], "docs": []}], `+
		`"gaps": [{"id": "spent", "covers": ["kind:x"], "candidate": null}]}`)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Dispose(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"feature:f"`) || !strings.HasSuffix(stdout.String(), "}\n") || !strings.Contains(stdout.String(), `"disposition": "adopt"`) {
		t.Errorf("stdout = %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "gap spent") {
		t.Errorf("stderr = %q, want gap spent named", stderr.String())
	}
	stdout.Reset()
	for _, args := range [][]string{{"extra"}, {"--placement", "p.json"}} {
		if err := app.Dispose(args, &stdout, &stderr); err == nil {
			t.Errorf("Dispose(%v) = nil, want it refused", args)
		}
	}
	if stdout.Len() != 0 {
		t.Errorf("Dispose wrote %q despite its error", stdout.String())
	}
}

func TestDisposeWithCandidatesTrimsForThemAndSplitsTheirGaps(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [{"path": "a", "library": ["kind:x"], "traits": [], "features": [], `+
		`"bytes": 1, "license_status": "n/a", "redistributable": true, "truth": [], "docs": []}], `+
		`"gaps": [{"id": "g1", "covers": ["kind:y", "kind:z"], "candidate": {"kind": "sourced", "license_status": "verified"}}]}`)
	write(t, filepath.Join(dir, "docs", "candidates.json"), `[{"gap": "g1", "path": "x/c", "source": "s", "sha256": "h", "bytes": 2, "format": "f", "format_by": "file-5"},
		{"gap": "g9", "path": "x/unplaced", "source": "s", "sha256": "h", "bytes": 2, "format": "f", "format_by": "file-5"}]`)
	write(t, filepath.Join(dir, "docs", "corpus-placement.json"), `[{"path": "x/c", "covers": ["kind:x", "kind:y"]}]`)
	write(t, filepath.Join(dir, "docs", "measurements.json"), `[]`)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Dispose(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"trimmed_for": "x/c"`, `"id": "g1-open"`, `"split_from": "g1"`} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout = %s\nwant it to contain %s", stdout.String(), want)
		}
	}
}

func TestDisposeMeasuresTheListedSamples(t *testing.T) {
	dir := t.TempDir()
	noCandidates(t, dir)
	sample := func(path string, bytes int) string {
		return `{"path": "` + path + `", "library": ["kind:x"], "traits": [], "features": [], "bytes": ` + strconv.Itoa(bytes) +
			`, "license_status": "n/a", "redistributable": true, "truth": [], "docs": []}`
	}
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [`+sample("a", 1)+`, `+sample("b", 2)+`], "gaps": []}`)
	measurement := `{"id": "m", "repository": "docproc", "acceptance": "A", "summary": "s", "needs_truth": false, "inputs": ["b"]`
	write(t, filepath.Join(dir, "docs", "measurements.json"), `[`+measurement+`}]`)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Dispose(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"disposition": "measure"`) {
		t.Errorf("stdout = %s\nwant b disposed measure", stdout.String())
	}
	write(t, filepath.Join(dir, "docs", "measurements.json"), `[`+measurement+`, "extra": 1}]`)
	stdout.Reset()
	if err := app.Dispose(nil, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "extra") {
		t.Errorf("Dispose = %v, want an error naming the unknown field", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("Dispose wrote %q despite its error", stdout.String())
	}
}

func TestDisposeReproducesTheCoverageData(t *testing.T) {
	root := docstest.Root(t)
	want, err := os.ReadFile(filepath.Join(root, "docs", "coverage.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	var stdout, stderr bytes.Buffer
	if err := app.Dispose(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != string(want) {
		t.Error("dispose does not reproduce docs/coverage.json; run mise run dispose and compare")
	}
}

func TestDisposeWritesSortedIDsEmptyListsBreaksAndUnescapedText(t *testing.T) {
	dir := t.TempDir()
	noCandidates(t, dir)
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"vocabulary": {
		"library": {"kind:z": {"identifier": "Z", "link": "l"}, "kind:a": {"identifier": "A", "link": "l"}},
		"trait": {}, "feature": {"feature:b": {"rule": "x"}, "feature:a": {"rule": "Q&A > 1"}}},
		"library_placement": {"run": "r", "docproc": "v", "docproc_command": "d", "command": "c"},
		"corpus_placement": {"run": "r2", "docproc": "v", "docproc_command": "d", "command": "c2"},
		"samples": [{"path": "a", "sha256": "s", "library": ["kind:a"], "traits": [], "features": [],
			"judged": {"feature:b": "seen", "feature:a": "Q&A <b>"}, "format": "f", "bytes": 1,
			"license_status": "n/a", "redistributable": true, "truth": [], "docs": [], "disposition": "trim", "trimmed_for": "z"},
			{"path": "b", "sha256": "s", "library": ["kind:a"], "traits": [], "features": [], "format": "f", "bytes": 2,
			"license_status": "n/a", "redistributable": true, "breaks": [{"standard": "A sample's content is realistic", "reason": "filler"}],
			"truth": [], "docs": [], "disposition": "adopt"}],
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
    "docproc": "v",
    "docproc_command": "d",
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
    },
    {
      "path": "b",
      "sha256": "s",
      "library": [
        "kind:a"
      ],
      "traits": [],
      "features": [],
      "format": "f",
      "bytes": 2,
      "license_status": "n/a",
      "redistributable": true,
      "breaks": [
        {
          "standard": "A sample's content is realistic",
          "reason": "filler"
        }
      ],
      "truth": [],
      "docs": [],
      "disposition": "trim",
      "trimmed_for": "a"
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
	if err := app.PlaceTraits(ctx, []string{"--corpus", "extra"}, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "usage: place-traits") {
		t.Errorf("PlaceTraits(extra) = %v, want its usage", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("PlaceTraits wrote %q despite its error", stdout.String())
	}
}

func TestPlaceTraitsRequiresCorpus(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for _, args := range [][]string{nil, {"--facts"}} {
		var stdout, stderr bytes.Buffer
		if err := app.PlaceTraits(ctx, args, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "usage: place-traits --corpus") {
			t.Errorf("PlaceTraits(%v) = %v, want its usage", args, err)
		}
	}
}

func TestPlaceTraitsAcceptsCorpusAndFacts(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for _, args := range [][]string{{"--corpus"}, {"--corpus", "--facts"}} {
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

func TestManifestWritesTheInventoryFromTheReferenceRows(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [{"path": "a/s.txt", "sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"library": ["kind:x"], "traits": [], "features": [], "format": "text/plain", "bytes": 3, "license_status": "reported",
		"redistributable": true, "truth": [], "docs": [], "disposition": "adopt"}], "gaps": []}`)
	write(t, filepath.Join(dir, "docs", "license-readings.json"), `[{"path": "a/s.txt", "license_url": "https://example.org/l", "license_read": "2026-11-25"}]`)
	noCandidates(t, dir)
	write(t, filepath.Join(dir, "docs", "reference-rows.json"), `[{"path": "a/s.txt", "sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"bytes": 3, "format": "text/plain", "source": "s", "source_url": "https://example.org/s?a&b", "license": "CC0-1.0",
		"license_status": "reported", "redistributable": true, "role": "sample", "truth": "none: a timing input"}]`)
	t.Setenv("HOME", t.TempDir())
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Manifest(nil, &stdout, &stderr); err != nil {
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
	if _, err := os.Stat(filepath.Join(dir, "a", "s.txt")); !os.IsNotExist(err) {
		t.Errorf("Manifest placed a/s.txt: %v, want it to write nothing but stdout", err)
	}
}

func TestManifestNamesTheMeasurementsThatReadAFile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [{"path": "a/s.txt", "sha256": "h",
		"library": ["kind:x"], "traits": [], "features": [], "format": "text/plain", "bytes": 3, "license_status": "reported",
		"redistributable": true, "truth": [], "docs": [], "disposition": "measure"}], "gaps": []}`)
	write(t, filepath.Join(dir, "docs", "license-readings.json"), `[]`)
	noCandidates(t, dir)
	measurement := `{"id": "m", "repository": "docproc", "acceptance": "A", "summary": "s", "needs_truth": false, "inputs": ["a/s.txt"]`
	write(t, filepath.Join(dir, "docs", "measurements.json"), `[`+measurement+`}]`)
	write(t, filepath.Join(dir, "docs", "reference-rows.json"), `[{"path": "a/s.txt", "sha256": "h", "bytes": 3, "format": "text/plain",
		"role": "sample", "truth": "none: a timing input"}]`)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Manifest(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"measures": [
        "m"
      ]`) {
		t.Errorf("stdout = %s\nwant a/s.txt to name measurement m", stdout.String())
	}
	write(t, filepath.Join(dir, "docs", "measurements.json"), `[`+measurement+`, "extra": 1}]`)
	stdout.Reset()
	if err := app.Manifest(nil, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "extra") {
		t.Errorf("Manifest with an unknown measurement field = %v, want it refused", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("Manifest wrote %q despite its error", stdout.String())
	}
}

func TestManifestAdmitsAListedSourcedInput(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [], "gaps": []}`)
	write(t, filepath.Join(dir, "docs", "license-readings.json"), `[]`)
	noCandidates(t, dir)
	write(t, filepath.Join(dir, "docs", "corpus-placement.json"), `[{"path": "video/d/v.mp4", "covers": ["kind:x"]}]`)
	measurement := `[{"id": "m", "repository": "docinfer", "acceptance": "A", "summary": "s", "needs_truth": true, "inputs": ["video/d/v.mp4"]}]`
	write(t, filepath.Join(dir, "docs", "measurements.json"), measurement)
	sourced := `{"path": "video/d/v.mp4", "source": "dvids", "url": "https://example.org/v", "license": "Public domain",
		"license_url": "https://example.org/c", "license_read": "2026-11-26", "format": "video/mp4", "bytes": 3, "sha256": "h",
		"truth": [{"path": "video/d/v.key.json", "sha256": "k", "bytes": 2, "format": "application/json", "license": "Apache-2.0",
		"license_url": "https://example.org/l", "license_read": "2026-11-26"}]`
	write(t, filepath.Join(dir, "docs", "sourced.json"), `[`+sourced+`}]`)
	write(t, filepath.Join(dir, "docs", "reference-rows.json"), `[]`)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if err := app.Manifest(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"path": "video/d/v.mp4"`, `"path": "video/d/v.key.json"`, `"role": "truth"`, `"kind:x"`, `"measures": [
        "m"
      ]`} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout = %s\nwant it to contain %s", stdout.String(), want)
		}
	}
	for _, c := range []struct{ name, measurements, sourced, want string }{
		{"unlisted", `[]`, `[` + sourced + `}]`, "a sourced input no measurement lists"},
		{"unknown field", measurement, `[` + sourced + `, "extra": 1}]`, "extra"},
	} {
		write(t, filepath.Join(dir, "docs", "measurements.json"), c.measurements)
		write(t, filepath.Join(dir, "docs", "sourced.json"), c.sourced)
		stdout.Reset()
		if err := app.Manifest(nil, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Manifest = %v, want an error containing %q", c.name, err, c.want)
		}
		if stdout.Len() != 0 {
			t.Errorf("%s: Manifest wrote %q despite its error", c.name, stdout.String())
		}
	}
}

func TestManifestRefusesArgumentsAndUnknownReferenceFields(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "docs", "coverage.json"), `{"samples": [], "gaps": []}`)
	write(t, filepath.Join(dir, "docs", "license-readings.json"), `[]`)
	write(t, filepath.Join(dir, "docs", "reference-rows.json"), `[{"path": "a/s.txt", "covers": []}]`)
	noCandidates(t, dir)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	for _, args := range [][]string{{"extra"}, {"--copy"}} {
		if err := app.Manifest(args, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "usage: manifest") && !strings.Contains(stderr.String(), "-copy") {
			t.Errorf("Manifest(%v) = %v, want it refused", args, err)
		}
	}
	if err := app.Manifest(nil, &stdout, &stderr); err == nil || !strings.Contains(err.Error(), "docs/reference-rows.json") || !strings.Contains(err.Error(), "covers") {
		t.Errorf("Manifest with an unknown reference field = %v, want docs/reference-rows.json and the field named", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("Manifest wrote %q despite its error", stdout.String())
	}
}
