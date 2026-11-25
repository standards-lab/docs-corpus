package disposition_test

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
)

func entry(path string, library []string, size int, redistributable bool, extra string) string {
	quoted := make([]string, len(library))
	for i, id := range library {
		quoted[i] = `"` + id + `"`
	}
	r := "false"
	if redistributable {
		r = "true"
	}
	return `{"path": "` + path + `", "library": [` + strings.Join(quoted, ", ") + `], "traits": [], "features": [], ` +
		`"bytes": ` + strconv.Itoa(size) + `, "license_status": "n/a", "redistributable": ` + r +
		`, "truth": [], "docs": []` + extra + `}`
}

func parse(t *testing.T, samples []string, gaps string) *disposition.Coverage {
	t.Helper()
	var coverage disposition.Coverage
	if err := json.Unmarshal([]byte(`{"samples": [`+strings.Join(samples, ", ")+`], "gaps": `+gaps+`}`), &coverage); err != nil {
		t.Fatal(err)
	}
	return &coverage
}

func dispositions(coverage *disposition.Coverage) []string {
	var out []string
	for _, s := range coverage.Samples {
		line := s.Path + " " + s.Disposition
		if s.TrimmedFor != "" {
			line += " " + s.TrimmedFor
		}
		if s.Gap != "" {
			line += " " + s.Gap
		}
		out = append(out, line)
	}
	return out
}

func TestDisposeAdoptsTrimsAndReplaces(t *testing.T) {
	coverage := parse(t, []string{
		entry("a", []string{"kind:x", "kind:y"}, 10, true, ""),
		entry("b", []string{"kind:x"}, 1, true, `, "disposition": "adopt"`),
		entry("c", []string{"kind:x"}, 1, false, ""),
		entry("d", []string{"kind:z", "kind:w"}, 5, false, `, "disposition": "replace", "gap": "g1"`),
		entry("e", []string{"kind:z"}, 5, false, `, "disposition": "trim", "trimmed_for": "a"`),
		entry("f", []string{"kind:x"}, 1, true, `, "sensitive": true`),
	}, `[{"id": "g1", "covers": ["kind:w"], "candidate": {"kind": "unavailable"}}]`)
	droppable, err := disposition.Dispose(coverage, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a adopt", "b trim a", "c trim a", "d replace g1", "e replace g1", "f trim a"}
	if got := dispositions(coverage); !slices.Equal(got, want) {
		t.Errorf("dispositions = %q, want %q", got, want)
	}
	gap := coverage.Gaps[0]
	if !slices.Equal(gap.Covers, []string{"kind:w", "kind:z"}) {
		t.Errorf("g1 covers = %q, want kind:w and kind:z", gap.Covers)
	}
	if gap.Candidate == nil || gap.Candidate.Kind != "unavailable" {
		t.Error("Dispose dropped g1's candidate")
	}
	if len(droppable) != 0 {
		t.Errorf("droppable = %q, want none", droppable)
	}
}

func TestDisposeAppliesTheTieBreaks(t *testing.T) {
	coverage := parse(t, []string{
		entry("big-truth", []string{"kind:x"}, 100, true, `, "truth": ["t"]`),
		entry("small", []string{"kind:x"}, 1, true, ""),
		entry("verified", []string{"kind:x", "kind:y"}, 50, true, `, "license_status": "verified"`),
		entry("cheap", []string{"kind:x", "kind:y"}, 5, true, ""),
		entry("held", []string{"kind:x"}, 1, false, ""),
	}, `[]`)
	if _, err := disposition.Dispose(coverage, nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"big-truth trim verified", "small trim verified", "verified adopt", "cheap trim verified", "held trim verified"}
	if got := dispositions(coverage); !slices.Equal(got, want) {
		t.Errorf("dispositions = %q, want %q", got, want)
	}
}

func TestDisposeGroupsReplacesUnderThePreferredGap(t *testing.T) {
	coverage := parse(t, []string{
		entry("inner", []string{"kind:x"}, 1, false, ""),
		entry("large", []string{"kind:x", "kind:y"}, 9, false, `, "gap": "g-large"`),
		entry("small", []string{"kind:x", "kind:z"}, 2, false, `, "gap": "g-small"`),
		entry("twin", []string{"kind:x", "kind:z"}, 3, false, ""),
	}, `[{"id": "g-large", "covers": [], "candidate": null}, {"id": "g-small", "covers": [], "candidate": null},
		{"id": "spare", "covers": ["kind:x"], "candidate": null}, {"id": "open", "covers": ["kind:q"], "candidate": null}]`)
	droppable, err := disposition.Dispose(coverage, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"inner replace g-small", "large replace g-large", "small replace g-small", "twin replace g-small"}
	if got := dispositions(coverage); !slices.Equal(got, want) {
		t.Errorf("dispositions = %q, want %q", got, want)
	}
	if len(droppable) != 0 {
		t.Errorf("droppable = %q; spare's kind:x and open's kind:q are exercised by no adopted sample", droppable)
	}
}

func TestDisposeNamesTheGapsItMayDrop(t *testing.T) {
	coverage := parse(t, []string{entry("a", []string{"kind:x"}, 1, true, "")},
		`[{"id": "spent", "covers": ["kind:x"], "candidate": null}, {"id": "open", "covers": ["kind:x", "kind:q"], "candidate": null}]`)
	droppable, err := disposition.Dispose(coverage, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(droppable, []string{"spent"}) {
		t.Errorf("droppable = %q, want [spent]", droppable)
	}
}

func TestDisposeNeverInventsAGap(t *testing.T) {
	coverage := parse(t, []string{entry("lone", []string{"kind:x"}, 1, false, "")}, `[]`)
	if _, err := disposition.Dispose(coverage, nil); err == nil || !strings.Contains(err.Error(), "lone") {
		t.Errorf("Dispose = %v, want an error naming lone", err)
	}
}

func TestPlaceSetsTraitsAndFeatures(t *testing.T) {
	coverage := parse(t, []string{entry("a", []string{"kind:x"}, 1, true, "")}, `[]`)
	placement := []disposition.Placement{{Path: "a", Traits: []string{"trait:t.v"}, Features: []string{"feature:f"}}}
	if err := disposition.Place(coverage, placement); err != nil {
		t.Fatal(err)
	}
	a := coverage.Samples[0]
	if got := strings.Join(slices.Concat(a.Library, a.Traits, a.Features), " "); got != "kind:x trait:t.v feature:f" {
		t.Errorf("placed = %q, want the library kept and the traits and features set", got)
	}
}

func TestPlaceRefusesAMismatchedPlacement(t *testing.T) {
	for name, placement := range map[string][]disposition.Placement{
		"missing": {},
		"extra":   {{Path: "a"}, {Path: "z"}},
	} {
		coverage := parse(t, []string{entry("a", []string{"kind:x"}, 1, true, "")}, `[]`)
		if err := disposition.Place(coverage, placement); err == nil {
			t.Errorf("%s: Place = nil, want an error", name)
		}
	}
}

const sourced = `{"kind": "sourced", "license_status": "verified"}`

func TestDisposeTrimsForCandidatesAndDropsCandidatesInsideAnother(t *testing.T) {
	coverage := parse(t, []string{
		entry("a", []string{"kind:x"}, 1, true, ""),
		entry("b", []string{"kind:y"}, 1, true, ""),
		entry("c", []string{"kind:x"}, 2, false, `, "disposition": "trim", "trimmed_for": "a"`),
	}, `[{"id": "g1", "covers": ["kind:z"], "candidate": `+sourced+`},
		{"id": "g2", "covers": ["kind:z"], "candidate": `+sourced+`},
		{"id": "g3", "covers": ["kind:w"], "candidate": {"kind": "generated", "inputs": ["b"]}},
		{"id": "g4", "covers": ["kind:v"], "covered_by": "x/kept", "candidate": `+sourced+`}]`)
	_, err := disposition.Dispose(coverage, []disposition.Placed{
		{Gap: "g1", Path: "x/one", Bytes: 9, Covers: []string{"kind:x", "kind:y", "kind:z"}},
		{Gap: "g2", Path: "x/two", Bytes: 1, Covers: []string{"kind:z"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a trim x/one", "b adopt", "c trim x/one"}
	if got := dispositions(coverage); !slices.Equal(got, want) {
		t.Errorf("dispositions = %q, want %q", got, want)
	}
	var covered []string
	for _, g := range coverage.Gaps {
		covered = append(covered, g.ID+" "+g.CoveredBy)
	}
	if want := []string{"g1 ", "g2 x/one", "g3 ", "g4 x/kept"}; !slices.Equal(covered, want) {
		t.Errorf("covered_by = %q, want %q", covered, want)
	}
}

func gapLines(coverage *disposition.Coverage) []string {
	var out []string
	for _, g := range coverage.Gaps {
		line := g.ID + " " + strings.Join(g.Covers, ",")
		if g.SplitFrom != "" {
			line += " from " + g.SplitFrom + " " + g.Candidate.Kind + " " + g.Candidate.Reason
		}
		out = append(out, line)
	}
	return out
}

func TestDisposeSplitsTheCoversACandidateDoesNotConfirm(t *testing.T) {
	coverage := parse(t, nil, `[{"id": "g1", "covers": ["feature:f", "kind:x"], "candidate": `+sourced+`},
		{"id": "g2", "covers": ["kind:y"], "candidate": `+sourced+`}]`)
	placed := []disposition.Placed{{Gap: "g1", Path: "x/one", Covers: []string{"kind:x"}}}
	if _, err := disposition.Dispose(coverage, placed); err != nil {
		t.Fatal(err)
	}
	want := []string{"g1 kind:x", "g1-open feature:f from g1 unavailable ", "g2 kind:y"}
	if got := gapLines(coverage); !slices.Equal(got, want) {
		t.Errorf("gaps = %q, want %q", got, want)
	}
	coverage.Gaps[1].Candidate.Reason = "unconfirmed"
	if _, err := disposition.Dispose(coverage, placed); err != nil {
		t.Fatal(err)
	}
	want[1] = "g1-open feature:f from g1 unavailable unconfirmed"
	if got := gapLines(coverage); !slices.Equal(got, want) {
		t.Errorf("gaps after a second dispose = %q, want %q", got, want)
	}
	placed[0].Covers = []string{"feature:f", "kind:x"}
	if _, err := disposition.Dispose(coverage, placed); err != nil {
		t.Fatal(err)
	}
	if got, want := gapLines(coverage), []string{"g1 feature:f,kind:x", "g2 kind:y"}; !slices.Equal(got, want) {
		t.Errorf("gaps once confirmed = %q, want %q", got, want)
	}
}

func TestDisposeRefusesAPlacedCandidateWithoutOne(t *testing.T) {
	cases := map[string]string{
		"no gap":       `[]`,
		"unavailable":  `[{"id": "g1", "covers": ["kind:x"], "candidate": {"kind": "unavailable"}}]`,
		"open id used": `[{"id": "g1", "covers": ["kind:x", "kind:y"], "candidate": ` + sourced + `}, {"id": "g1-open", "covers": ["kind:z"], "candidate": {"kind": "unavailable"}}]`,
	}
	for name, gaps := range cases {
		_, err := disposition.Dispose(parse(t, nil, gaps), []disposition.Placed{{Gap: "g1", Path: "x/one", Covers: []string{"kind:x"}}})
		if err == nil {
			t.Errorf("%s: Dispose = nil, want an error", name)
		}
	}
}
