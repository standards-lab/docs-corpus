package docstest_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/internal/docstest"
)

func TestProbedLeavesOutJudgedFeatures(t *testing.T) {
	s := docstest.Sample{
		Features: []string{"feature:tables", "feature:charts", "feature:ink"},
		Judged:   map[string]json.RawMessage{"feature:ink": json.RawMessage(`"seen"`)},
	}
	if got := s.Probed(); !slices.Equal(got, []string{"feature:charts", "feature:tables"}) {
		t.Errorf("Probed = %v", got)
	}
}

func TestSamplesReadsTheCoverageData(t *testing.T) {
	samples := docstest.Samples(t)
	if len(samples) == 0 || samples[0].Path == "" {
		t.Errorf("Samples = %d samples, the first %+v", len(samples), samples)
	}
}

func TestCoverageReadsTheCoverageData(t *testing.T) {
	if data := docstest.Coverage(t); !json.Valid(data) {
		t.Errorf("Coverage = %d bytes that are not JSON", len(data))
	}
}

func TestSamplesAreTheCoverageDataSamplesTheManifestLists(t *testing.T) {
	paths := map[string]bool{}
	for _, s := range docstest.Samples(t) {
		paths[s.Path] = true
	}
	if !paths["email/eml/cpython/msg_36.eml"] {
		t.Error("Samples lacks email/eml/cpython/msg_36.eml, an adopted sample")
	}
	if paths["archive/high-ratio-200mb.zip"] {
		t.Error("Samples has archive/high-ratio-200mb.zip, a replaced sample MANIFEST.json doesn't list")
	}
}

func TestPathFindsAFetchedSampleInTheCheckout(t *testing.T) {
	path := docstest.Path(t, "text/encodings/windows-1252.txt")
	if !strings.HasSuffix(filepath.ToSlash(path), "/text/encodings/windows-1252.txt") {
		t.Errorf("Path = %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Error(err)
	}
}
