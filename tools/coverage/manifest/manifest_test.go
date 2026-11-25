package manifest_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
	"github.com/standards-lab/docs-corpus/tools/coverage/manifest"
)

func adopted(path string, truth, docs []string) disposition.Sample {
	return disposition.Sample{
		Path: path, Sha256: "s-" + path, Disposition: "adopt",
		Library: []string{"route:b", "kind:a"}, Traits: []string{"trait:x.y"}, Features: []string{"feature:f", "kind:a"},
		Truth: truth, Docs: docs,
	}
}

func ref(path, role, truth string) manifest.ReferenceFile {
	return manifest.ReferenceFile{
		Path: path, Role: role, Sha256: "s-" + path, Bytes: 3, Format: "text/plain", Source: "src",
		SourceURL: "https://example.org/" + path, License: "CC0-1.0", LicenseStatus: "reported",
		Redistributable: true, Truth: truth,
	}
}

func build(t *testing.T, samples []disposition.Sample, files []manifest.ReferenceFile, readings []manifest.Reading) manifest.Manifest {
	t.Helper()
	m, err := manifest.Build(disposition.Coverage{Samples: samples}, manifest.Reference{Files: files}, readings)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func byPath(m manifest.Manifest) map[string]manifest.File {
	out := map[string]manifest.File{}
	for _, f := range m.Files {
		out[f.Path] = f
	}
	return out
}

func TestBuildListsAdoptedSamplesAndTheirFilesSorted(t *testing.T) {
	trimmed := adopted("z/trimmed.txt", []string{"z/trimmed.truth"}, nil)
	trimmed.Disposition = "trim"
	m := build(t,
		[]disposition.Sample{adopted("b/s.txt", []string{"b/s.truth"}, []string{"LICENSE"}), adopted("a/s.txt", nil, []string{"LICENSE"}), trimmed},
		[]manifest.ReferenceFile{
			ref("b/s.txt", "sample", "b/s.truth"), ref("a/s.txt", "sample", ""), ref("b/s.truth", "truth", ""),
			ref("LICENSE", "doc", ""), ref("z/trimmed.txt", "sample", ""), ref("z/trimmed.truth", "truth", ""),
		}, nil)
	if m.Version != 1 {
		t.Errorf("Version = %d, want 1", m.Version)
	}
	var got []string
	for _, f := range m.Files {
		got = append(got, f.Path+" "+f.Role)
	}
	want := []string{"LICENSE doc", "a/s.txt sample", "b/s.truth truth", "b/s.txt sample"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("files = %q, want %q", got, want)
	}
}

func TestBuildCarriesProvenanceAndCovers(t *testing.T) {
	r := ref("a/s.txt", "sample", "")
	r.DerivedFrom, r.Note = "https://example.org/orig", "a note"
	m := build(t, []disposition.Sample{adopted("a/s.txt", nil, []string{"a/LICENSE"})},
		[]manifest.ReferenceFile{r, ref("a/LICENSE", "doc", "")}, nil)
	want := manifest.File{
		Path: "a/s.txt", Role: "sample", Sha256: "s-a/s.txt", Bytes: 3, Format: "text/plain",
		Covers: []string{"feature:f", "kind:a", "route:b", "trait:x.y"},
		Truth:  []manifest.Truth{}, Docs: []string{"a/LICENSE"},
		Source: "src", URL: "https://example.org/a/s.txt", License: "CC0-1.0", LicenseStatus: "reported",
		Redistributable: true, DerivedFrom: "https://example.org/orig", Note: "a note",
	}
	if got := byPath(m)["a/s.txt"]; !reflect.DeepEqual(got, want) {
		t.Errorf("sample = %+v, want %+v", got, want)
	}
	doc := byPath(m)["a/LICENSE"]
	if doc.Covers != nil || doc.Truth != nil || doc.Docs != nil || doc.URL != "https://example.org/a/LICENSE" {
		t.Errorf("doc = %+v, want provenance and no covers, truth, or docs", doc)
	}
}

func TestBuildCarriesTruthQualifiersAsNotes(t *testing.T) {
	cases := []struct {
		text  string
		paths []string
		want  []manifest.Truth
		note  string
	}{
		{"t/a.srt", []string{"t/a.srt"}, []manifest.Truth{{Path: "t/a.srt"}}, ""},
		{"t/a.srt (first 180 s)", []string{"t/a.srt"}, []manifest.Truth{{Path: "t/a.srt", Note: "first 180 s"}}, ""},
		{"t/d/ (words; and (nested) segments)", []string{"t/d/1.xml", "t/d/2.xml"},
			[]manifest.Truth{{Path: "t/d/1.xml", Note: "words; and (nested) segments"}, {Path: "t/d/2.xml", Note: "words; and (nested) segments"}}, ""},
		{"t/d/", []string{"t/d/1.xml"}, []manifest.Truth{{Path: "t/d/1.xml"}}, ""},
		{`t/s.json (entry "x/y.one"); content fields in t/c.json`, []string{"t/c.json", "t/s.json"},
			[]manifest.Truth{{Path: "t/c.json", Note: "content fields in t/c.json"}, {Path: "t/s.json", Note: `entry "x/y.one"`}}, ""},
		{"t/a.srt (one) (two)", []string{"t/a.srt"}, []manifest.Truth{{Path: "t/a.srt", Note: "t/a.srt (one) (two)"}}, ""},
		{"none: an ASR output, not truth", nil, []manifest.Truth{}, "an ASR output, not truth"},
	}
	for _, c := range cases {
		files := []manifest.ReferenceFile{ref("s.mp4", "sample", c.text)}
		for _, p := range c.paths {
			files = append(files, ref(p, "truth", ""))
		}
		got := byPath(build(t, []disposition.Sample{adopted("s.mp4", c.paths, nil)}, files, nil))["s.mp4"]
		if !reflect.DeepEqual(got.Truth, c.want) || got.TruthNote != c.note {
			t.Errorf("truth %q = %+v with note %q, want %+v with note %q", c.text, got.Truth, got.TruthNote, c.want, c.note)
		}
	}
}

func TestBuildAppliesLicenseReadings(t *testing.T) {
	m := build(t, []disposition.Sample{adopted("a.pdf", nil, nil), adopted("b.pdf", nil, nil)},
		[]manifest.ReferenceFile{ref("a.pdf", "sample", ""), ref("b.pdf", "sample", "")},
		[]manifest.Reading{{Path: "a.pdf", LicenseURL: "https://example.org/license", LicenseRead: "2026-11-25"}})
	a, b := byPath(m)["a.pdf"], byPath(m)["b.pdf"]
	if a.LicenseURL != "https://example.org/license" || a.LicenseRead != "2026-11-25" || a.LicenseStatus != "verified" {
		t.Errorf("read sample = %+v, want the reading applied and verified", a)
	}
	if b.LicenseURL != "" || b.LicenseRead != "" || b.LicenseStatus != "reported" {
		t.Errorf("unread sample = %+v, want the reference's status and no reading", b)
	}
}

func TestBuildRefusesInconsistentInputs(t *testing.T) {
	stale := ref("a.pdf", "sample", "")
	stale.Sha256 = "other"
	cases := []struct {
		name     string
		samples  []disposition.Sample
		files    []manifest.ReferenceFile
		readings []manifest.Reading
		want     string
	}{
		{"missing sample", []disposition.Sample{adopted("a.pdf", nil, nil)}, nil, nil, "a.pdf: an adopted sample the reference manifest lacks"},
		{"stale sha256", []disposition.Sample{adopted("a.pdf", nil, nil)}, []manifest.ReferenceFile{stale}, nil, "differs from the reference's"},
		{"missing truth", []disposition.Sample{adopted("a.pdf", []string{"t.json"}, nil)}, []manifest.ReferenceFile{ref("a.pdf", "sample", "")}, nil, "t.json: a truth file the reference manifest lacks"},
		{"sample as truth", []disposition.Sample{adopted("a.pdf", []string{"b.pdf"}, nil), adopted("b.pdf", nil, nil)},
			[]manifest.ReferenceFile{ref("a.pdf", "sample", ""), ref("b.pdf", "sample", "")}, nil, "b.pdf: both an adopted sample and a truth file"},
		{"stray reading", nil, nil, []manifest.Reading{{Path: "x.pdf"}}, "x.pdf: a license reading for a path that is not an adopted sample"},
	}
	for _, c := range cases {
		_, err := manifest.Build(disposition.Coverage{Samples: c.samples}, manifest.Reference{Files: c.files}, c.readings)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Build = %v, want an error containing %q", c.name, err, c.want)
		}
	}
}

func sum(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func TestCopyPutsEachFileInPlaceAndChecksIt(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	for path, data := range map[string]string{"a/b c.txt": "abc", "d.txt": "defg", "bad.txt": "xyz"} {
		if err := os.MkdirAll(filepath.Join(src, filepath.Dir(path)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(src, path), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	good := manifest.Manifest{Files: []manifest.File{
		{Path: "a/b c.txt", Sha256: sum("abc"), Bytes: 3},
		{Path: "d.txt", Sha256: sum("defg"), Bytes: 4},
	}}
	if err := manifest.Copy(good, src, dst); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(dst, "a", "b c.txt")); err != nil || string(data) != "abc" {
		t.Errorf("copied a/b c.txt = %q, %v", data, err)
	}
	if err := manifest.Copy(good, t.TempDir(), dst); err != nil {
		t.Errorf("Copy over files already in place = %v, want them left as they are", err)
	}
	bad := manifest.Manifest{Files: []manifest.File{{Path: "bad.txt", Sha256: sum("abc"), Bytes: 3}}}
	if err := manifest.Copy(bad, src, dst); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Errorf("Copy of a file whose sha256 differs = %v, want a sha256 error", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "bad.txt")); !os.IsNotExist(err) {
		t.Errorf("bad.txt stayed after its check failed: %v", err)
	}
}
