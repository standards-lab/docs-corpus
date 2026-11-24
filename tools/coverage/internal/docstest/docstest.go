package docstest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

// Dir returns the reference docs directory, $DOCS_REFERENCE or else ~/docs, and skips t, naming
// the directory, when it is absent.
func Dir(t testing.TB) string {
	t.Helper()
	dir := os.Getenv("DOCS_REFERENCE")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no home directory to find ~/docs in: %v", err)
		}
		dir = filepath.Join(home, "docs")
	}
	if _, err := os.Stat(filepath.Join(dir, "MANIFEST.json")); err != nil {
		t.Skipf("%s/MANIFEST.json is absent; set DOCS_REFERENCE to the reference docs", dir)
	}
	return dir
}

// Path returns the path of rel under Dir, skipping t when Dir is absent and failing it when Dir
// is present but rel is not.
func Path(t testing.TB, rel string) string {
	t.Helper()
	path := filepath.Join(Dir(t), rel)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return path
}

// Sample is one sample's entry in docs/coverage.json.
type Sample struct {
	Path     string                     `json:"path"`
	Traits   []string                   `json:"traits"`
	Features []string                   `json:"features"`
	Judged   map[string]json.RawMessage `json:"judged"`
}

// Probed returns the sample's feature IDs that are not judged, sorted, the ones a probe places.
func (s Sample) Probed() []string {
	out := []string{}
	for _, id := range s.Features {
		if _, judged := s.Judged[id]; !judged {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out
}

// Coverage returns the bytes of the repository's docs/coverage.json, skipping t when the file is
// absent, as it is outside the module.
func Coverage(t testing.TB) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("docstest: no source path to find docs/coverage.json from")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "docs", "coverage.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("%s is absent: %v", path, err)
	}
	return data
}

// Samples returns the samples of the repository's docs/coverage.json, skipping t when the file is
// absent, as Coverage does, and failing it when the file doesn't parse.
func Samples(t testing.TB) []Sample {
	t.Helper()
	var coverage struct {
		Samples []Sample `json:"samples"`
	}
	if err := json.Unmarshal(Coverage(t), &coverage); err != nil {
		t.Fatalf("docs/coverage.json: %v", err)
	}
	return coverage.Samples
}
