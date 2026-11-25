package docstest_test

import (
	"encoding/json"
	"slices"
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
