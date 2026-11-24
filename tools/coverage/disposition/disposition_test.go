package disposition_test

import (
	"bytes"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
	"github.com/standards-lab/docs-corpus/tools/coverage/internal/docstest"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
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

func parse(t *testing.T, samples []string, gaps string) *jsonvalue.Object {
	t.Helper()
	v, err := jsonvalue.Parse([]byte(`{"samples": [` + strings.Join(samples, ", ") + `], "gaps": ` + gaps + `}`))
	if err != nil {
		t.Fatal(err)
	}
	return v.(*jsonvalue.Object)
}

func dispositions(coverage *jsonvalue.Object) []string {
	var out []string
	for _, s := range coverage.Get("samples").([]any) {
		o := s.(*jsonvalue.Object)
		line := o.Str("path") + " " + o.Str("disposition")
		if o.Has("trimmed_for") {
			line += " " + o.Str("trimmed_for")
		}
		if o.Has("gap") {
			line += " " + o.Str("gap")
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
	droppable, err := disposition.Dispose(coverage)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a adopt", "b trim a", "c trim a", "d replace g1", "e replace g1", "f trim a"}
	if got := dispositions(coverage); !slices.Equal(got, want) {
		t.Errorf("dispositions = %q, want %q", got, want)
	}
	gap := coverage.Get("gaps").([]any)[0].(*jsonvalue.Object)
	if got := string(jsonvalue.Marshal(gap.Get("covers"))); got != "[\n  \"kind:w\",\n  \"kind:z\"\n]" {
		t.Errorf("g1 covers = %s, want kind:w and kind:z", got)
	}
	if gap.Get("candidate") == nil {
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
	if _, err := disposition.Dispose(coverage); err != nil {
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
	droppable, err := disposition.Dispose(coverage)
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
	droppable, err := disposition.Dispose(coverage)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(droppable, []string{"spent"}) {
		t.Errorf("droppable = %q, want [spent]", droppable)
	}
}

func TestDisposeNeverInventsAGap(t *testing.T) {
	coverage := parse(t, []string{entry("lone", []string{"kind:x"}, 1, false, "")}, `[]`)
	if _, err := disposition.Dispose(coverage); err == nil || !strings.Contains(err.Error(), "lone") {
		t.Errorf("Dispose = %v, want an error naming lone", err)
	}
}

func TestPlaceSetsTraitsAndFeatures(t *testing.T) {
	coverage := parse(t, []string{entry("a", []string{"kind:x"}, 1, true, "")}, `[]`)
	placement, err := jsonvalue.Parse([]byte(`[{"path": "a", "traits": ["trait:t.v"], "features": ["feature:f"]}]`))
	if err != nil {
		t.Fatal(err)
	}
	if err := disposition.Place(coverage, placement.([]any)); err != nil {
		t.Fatal(err)
	}
	a := coverage.Get("samples").([]any)[0].(*jsonvalue.Object)
	if got := jsonvalue.String(a.Get("traits").([]any)[0]) + " " + jsonvalue.String(a.Get("features").([]any)[0]); got != "trait:t.v feature:f" {
		t.Errorf("placed = %q", got)
	}
	if !slices.Equal(a.Keys()[:5], []string{"path", "library", "traits", "features", "bytes"}) {
		t.Errorf("keys = %q, want their order kept", a.Keys())
	}
}

func TestPlaceRefusesAMismatchedPlacement(t *testing.T) {
	for name, placement := range map[string]string{
		"missing": `[]`,
		"extra":   `[{"path": "a", "traits": [], "features": []}, {"path": "z", "traits": [], "features": []}]`,
	} {
		coverage := parse(t, []string{entry("a", []string{"kind:x"}, 1, true, "")}, `[]`)
		v, err := jsonvalue.Parse([]byte(placement))
		if err != nil {
			t.Fatal(err)
		}
		if err := disposition.Place(coverage, v.([]any)); err == nil {
			t.Errorf("%s: Place = nil, want an error", name)
		}
	}
}

func TestDisposeReproducesTheCoverageData(t *testing.T) {
	data := docstest.Coverage(t)
	v, err := jsonvalue.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	coverage := v.(*jsonvalue.Object)
	for _, s := range coverage.Get("samples").([]any) {
		o := s.(*jsonvalue.Object)
		o.Set("disposition", "unset")
		o.Set("trimmed_for", "unset")
	}
	if _, err := disposition.Dispose(coverage); err != nil {
		t.Fatal(err)
	}
	if got := append(jsonvalue.Marshal(coverage), '\n'); !bytes.Equal(got, data) {
		t.Error("Dispose does not reproduce docs/coverage.json; run mise run dispose and compare")
	}
}
