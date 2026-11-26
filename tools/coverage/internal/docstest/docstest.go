package docstest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

// Path returns the path of rel in the repository's checkout, skipping t, naming the fetch, when
// the file isn't fetched, and failing it when MANIFEST.json doesn't list rel.
func Path(t testing.TB, rel string) string {
	t.Helper()
	if !listed(t)[rel] {
		t.Fatalf("%s: not a file MANIFEST.json lists", rel)
	}
	path := filepath.Join(Root(t), filepath.FromSlash(rel))
	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s is not fetched; fetch the samples with mise run fetch: %v", rel, err)
	}
	return path
}

func listed(t testing.TB) map[string]bool {
	t.Helper()
	path := filepath.Join(Root(t), "MANIFEST.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	var manifest struct {
		Files []struct {
			Path string `json:"path"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	paths := make(map[string]bool, len(manifest.Files))
	for _, f := range manifest.Files {
		paths[f.Path] = true
	}
	return paths
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

// Root returns the repository's root, the directory docs/coverage.json is in, skipping t when
// that file is absent, as it is outside the module.
func Root(t testing.TB) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("docstest: no source path to find docs/coverage.json from")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
	if _, err := os.Stat(filepath.Join(root, "docs", "coverage.json")); err != nil {
		t.Skipf("%s is absent: %v", filepath.Join(root, "docs", "coverage.json"), err)
	}
	return root
}

// Coverage returns the bytes of the repository's docs/coverage.json, skipping t when the file is
// absent, as Root does.
func Coverage(t testing.TB) []byte {
	t.Helper()
	path := filepath.Join(Root(t), "docs", "coverage.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return data
}

// Samples returns the samples of the repository's docs/coverage.json that MANIFEST.json lists,
// skipping t when docs/coverage.json is absent, as Coverage does, and failing it when either file
// doesn't parse.
func Samples(t testing.TB) []Sample {
	t.Helper()
	var coverage struct {
		Samples []Sample `json:"samples"`
	}
	if err := json.Unmarshal(Coverage(t), &coverage); err != nil {
		t.Fatalf("docs/coverage.json: %v", err)
	}
	paths := listed(t)
	return slices.DeleteFunc(coverage.Samples, func(s Sample) bool { return !paths[s.Path] })
}
