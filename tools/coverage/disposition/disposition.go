package disposition

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
)

type sample struct {
	entry  *Sample
	covers []string
	set    idset.Set
	gap    string
}

func readSample(e *Sample) *sample {
	s := &sample{entry: e, gap: e.Gap, set: idset.Set{}}
	s.set.Add(e.Library...)
	s.set.Add(e.Traits...)
	s.set.Add(e.Features...)
	s.covers = s.set.Sorted("")
	return s
}

func lacks(has bool) int {
	if has {
		return 0
	}
	return 1
}

func rank(a, b *sample) int {
	x, y := a.entry, b.entry
	return cmp.Or(
		cmp.Compare(lacks(len(x.Truth) > 0), lacks(len(y.Truth) > 0)),
		cmp.Compare(lacks(x.LicenseStatus == "verified"), lacks(y.LicenseStatus == "verified")),
		cmp.Compare(lacks(x.Redistributable), lacks(y.Redistributable)),
		cmp.Compare(x.Bytes, y.Bytes),
		strings.Compare(x.Path, y.Path),
	)
}

func within(inner, outer *sample) bool {
	for _, id := range inner.covers {
		if !outer.set[id] {
			return false
		}
	}
	return true
}

func best(candidates []*sample) *sample {
	if len(candidates) == 0 {
		return nil
	}
	return slices.MinFunc(candidates, rank)
}

func (s *sample) dispose(disposition, trimmedFor, gap string) {
	s.entry.Disposition, s.entry.TrimmedFor, s.entry.Gap = disposition, trimmedFor, gap
}

// Dispose derives every sample's disposition in coverage as docs/coverage.md states, and sets it
// in place with its trimmed_for or gap, and each gap a replace names its covers. A replaced sample
// takes the gap of the replaced sample whose covers contain its own, the one the tie-break prefers
// among those no other replaced sample's covers contain, and that gap is the one those samples
// name in coverage; candidates are left as they are.
//
// placed are the gap candidates the corpus placement placed. A sourced candidate's license status
// is its candidate's; a generated one's inputs are adopted samples or placed sourced candidates,
// and its license status is verified when every input's is, otherwise the first input's that
// isn't, and it is redistributable when every input is, with the truth of the input its TruthFrom
// names. They join the adopted samples: an adopted sample whose covers lie inside
// one's is trimmed for it, and one whose covers lie inside another sample's leaves the corpus, its
// gap's covered_by naming that sample. An adopted sample or sourced candidate a generated
// candidate lists in its inputs stays whatever covers it. A gap whose candidate's covers lack some of its covers keeps the ones
// they confirm, and the rest become its open gap, id <gap>-open, split_from naming it, whose
// candidate is unavailable until its searched and reason are recorded; the open gap goes once the
// candidate confirms them all. A gap whose candidate placed doesn't list keeps its covers and
// open gap as they are, and its covered_by while that names a sample still in the corpus; a
// covered_by naming a sample that left the corpus is cleared, so its candidate is fetched again.
//
// A sample whose Breaks record a standard it breaks is trimmed, never adopted or replaced:
// trimmed_for names the sample the tie-break prefers among those in the corpus whose covers
// contain its own, and is absent when none does. Each of its covers must be another corpus
// sample's or a gap's.
//
// It returns the gaps no replace names whose covers adopted samples all exercise, which may be
// dropped. A group of replaced samples that names no gap, or names two, a breaking sample with a
// cover no other corpus sample or gap holds, a placed candidate whose gap has no sourced or
// generated candidate, and an open gap's id taken by another gap are errors.
func Dispose(coverage *Coverage, placed []Placed) ([]string, error) {
	samples := make([]*sample, len(coverage.Samples))
	for i := range coverage.Samples {
		samples[i] = readSample(&coverage.Samples[i])
	}
	inputs := idset.Set{}
	for _, g := range coverage.Gaps {
		if g.Candidate != nil && g.Candidate.Kind == "generated" {
			inputs.Add(g.Candidate.Inputs...)
		}
	}
	candidates, err := readCandidates(coverage.Gaps, placed, coverage.Samples)
	if err != nil {
		return nil, err
	}

	var eligible []*sample
	for _, s := range samples {
		if s.entry.Redistributable && !s.entry.Sensitive && len(s.entry.Breaks) == 0 {
			eligible = append(eligible, s)
		}
	}
	adopted := kept(eligible, inputs)
	corpus := kept(slices.Concat(eligible, candidates), inputs)
	var replaced []*sample
	held := idset.Set{}
	for _, s := range corpus {
		held.Merge(s.set)
	}
	for _, g := range coverage.Gaps {
		held.Add(g.Covers...)
	}
	for _, s := range samples {
		switch {
		case len(s.entry.Breaks) > 0:
			if unheld := slices.DeleteFunc(slices.Clone(s.covers), func(id string) bool { return held[id] }); len(unheld) > 0 {
				return nil, fmt.Errorf("%s: breaks %q, and no other sample or gap holds %s; research a candidate and add the gap", s.entry.Path, s.entry.Breaks[0].Standard, strings.Join(unheld, ", "))
			}
			trimmedFor := ""
			if h := holder(s, corpus); h != nil {
				trimmedFor = h.entry.Path
			}
			s.dispose("trim", trimmedFor, "")
		case !slices.ContainsFunc(adopted, func(t *sample) bool { return within(s, t) }):
			replaced = append(replaced, s)
		case slices.Contains(corpus, s):
			s.dispose("adopt", "", "")
		default:
			s.dispose("trim", holder(s, corpus).entry.Path, "")
		}
	}

	inCorpus := idset.Set{}
	for _, s := range corpus {
		inCorpus.Add(s.entry.Path)
	}
	gaps := map[string]*Gap{}
	for i := range coverage.Gaps {
		g := &coverage.Gaps[i]
		gaps[g.ID] = g
		if !inCorpus[g.CoveredBy] {
			g.CoveredBy = ""
		}
	}
	for _, c := range candidates {
		g := gaps[c.gap]
		g.CoveredBy = ""
		if !slices.Contains(corpus, c) {
			g.CoveredBy = holder(c, corpus).entry.Path
		}
	}
	outer := filter(replaced, func(s *sample) bool {
		return !slices.ContainsFunc(replaced, func(t *sample) bool { return t != s && within(s, t) && len(t.covers) > len(s.covers) })
	})
	groupGap := map[string]string{}
	for _, s := range outer {
		key := strings.Join(s.covers, " ")
		if s.gap == "" || gaps[s.gap] == nil {
			continue
		}
		if prior, ok := groupGap[key]; ok && prior != s.gap {
			return nil, fmt.Errorf("replaced samples with covers %s name the gaps %s and %s; keep one", key, prior, s.gap)
		}
		groupGap[key] = s.gap
	}
	named := idset.Set{}
	for _, s := range replaced {
		holder := best(filter(outer, func(t *sample) bool { return within(s, t) }))
		key := strings.Join(holder.covers, " ")
		id, ok := groupGap[key]
		if !ok {
			return nil, fmt.Errorf("%s: replaced, and no gap holds the covers %s; research a candidate and add the gap", holder.entry.Path, key)
		}
		named.Add(id)
		gaps[id].Covers = slices.Clone(holder.covers)
		s.dispose("replace", "", id)
	}

	exercised := idset.Set{}
	for _, s := range adopted {
		exercised.Merge(s.set)
	}
	var droppable []string
	for _, g := range coverage.Gaps {
		if named[g.ID] {
			continue
		}
		if !slices.ContainsFunc(g.Covers, func(id string) bool { return !exercised[id] }) {
			droppable = append(droppable, g.ID)
		}
	}
	return droppable, split(coverage, candidates)
}

func inherited(inputs []string, samples map[string]*Sample) (string, bool) {
	status, redistributable := "verified", len(inputs) > 0
	for _, path := range inputs {
		s, ok := samples[path]
		if !ok {
			return "", false
		}
		if status == "verified" {
			status = s.LicenseStatus
		}
		redistributable = redistributable && s.Redistributable
	}
	return status, redistributable
}

func readCandidates(gaps []Gap, placed []Placed, entries []Sample) ([]*sample, error) {
	samples := map[string]*Sample{}
	for i := range entries {
		samples[entries[i].Path] = &entries[i]
	}
	byID := map[string]*Gap{}
	for i := range gaps {
		byID[gaps[i].ID] = &gaps[i]
	}
	for _, p := range placed {
		g := byID[p.Gap]
		if g == nil || g.Candidate == nil || (g.Candidate.Kind != "sourced" && g.Candidate.Kind != "generated") {
			return nil, fmt.Errorf("%s: placed as gap %s's candidate, but that gap has no sourced or generated candidate", p.Path, p.Gap)
		}
	}
	var out []*sample
	for _, kind := range []string{"sourced", "generated"} {
		for _, p := range placed {
			c := byID[p.Gap].Candidate
			if c.Kind != kind {
				continue
			}
			e := &Sample{
				Path: p.Path, Library: p.Covers, Bytes: p.Bytes, Truth: []string{},
				LicenseStatus: c.LicenseStatus, Redistributable: c.LicenseStatus == "verified",
			}
			if kind == "generated" {
				e.LicenseStatus, e.Redistributable = inherited(c.Inputs, samples)
				if from, ok := samples[p.TruthFrom]; ok && slices.Contains(c.Inputs, p.TruthFrom) {
					e.Truth = slices.Clone(from.Truth)
				}
			} else {
				samples[p.Path] = e
			}
			s := readSample(e)
			s.gap = p.Gap
			out = append(out, s)
		}
	}
	return out, nil
}

func kept(pool []*sample, inputs idset.Set) []*sample {
	return filter(pool, func(s *sample) bool {
		return inputs[s.entry.Path] || !slices.ContainsFunc(pool, func(t *sample) bool {
			return t != s && within(s, t) && (len(t.covers) > len(s.covers) || rank(t, s) < 0)
		})
	})
}

func holder(s *sample, corpus []*sample) *sample {
	return best(filter(corpus, func(t *sample) bool { return t != s && within(s, t) }))
}

func split(coverage *Coverage, candidates []*sample) error {
	placed := map[string]*sample{}
	for _, c := range candidates {
		placed[c.gap] = c
	}
	open := map[string]Gap{}
	ids := idset.Set{}
	for _, g := range coverage.Gaps {
		ids.Add(g.ID)
		if g.SplitFrom != "" {
			open[g.SplitFrom] = g
		}
	}
	out := make([]Gap, 0, len(coverage.Gaps))
	for _, g := range coverage.Gaps {
		c := placed[g.SplitFrom]
		if g.SplitFrom != "" && c != nil {
			continue
		}
		c = placed[g.ID]
		if c == nil || g.SplitFrom != "" {
			out = append(out, g)
			continue
		}
		o, hasOpen := open[g.ID]
		full := idset.New(g.Covers...)
		full.Add(o.Covers...)
		confirmed, unconfirmed := idset.Set{}, idset.Set{}
		for id := range full {
			if c.set[id] {
				confirmed.Add(id)
			} else {
				unconfirmed.Add(id)
			}
		}
		g.Covers = confirmed.Sorted("")
		out = append(out, g)
		if len(unconfirmed) == 0 {
			continue
		}
		if !hasOpen {
			o = Gap{ID: g.ID + "-open", SplitFrom: g.ID, Candidate: &Candidate{Kind: "unavailable"}}
			if ids[o.ID] {
				return fmt.Errorf("gap %s: its open gap's id %s is another gap's; rename that gap", g.ID, o.ID)
			}
		}
		o.Covers = unconfirmed.Sorted("")
		out = append(out, o)
	}
	coverage.Gaps = out
	return nil
}

func filter(samples []*sample, keep func(*sample) bool) []*sample {
	var out []*sample
	for _, s := range samples {
		if keep(s) {
			out = append(out, s)
		}
	}
	return out
}

// Place sets each sample's traits and features in coverage to its element in placement, the
// elements place-traits prints. A sample placement lacks, or an element no sample has, is an
// error.
func Place(coverage *Coverage, placement []Placement) error {
	placed := map[string]Placement{}
	for _, p := range placement {
		placed[p.Path] = p
	}
	for i := range coverage.Samples {
		s := &coverage.Samples[i]
		p, ok := placed[s.Path]
		if !ok {
			return fmt.Errorf("placement: has no element for %s", s.Path)
		}
		s.Traits, s.Features = p.Traits, p.Features
		delete(placed, s.Path)
	}
	if extra := slices.Sorted(maps.Keys(placed)); len(extra) > 0 {
		return fmt.Errorf("placement: %s is not a sample", extra[0])
	}
	return nil
}
