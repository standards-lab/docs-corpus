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

// Dispose derives every sample's disposition in coverage as
// docs/coverage.md states, and sets it in place with its trimmed_for or gap, and each gap a
// replace names its covers. A replaced sample takes the gap of the replaced sample whose covers
// contain its own, the one the tie-break prefers among those no other replaced sample's covers
// contain, and that gap is the one those samples name in coverage; candidates are left as they
// are. It returns the gaps no replace names whose covers adopted samples all exercise, which may
// be dropped. A group of replaced samples that names no gap, or names two, is an error.
func Dispose(coverage *Coverage) ([]string, error) {
	samples := make([]*sample, len(coverage.Samples))
	for i := range coverage.Samples {
		samples[i] = readSample(&coverage.Samples[i])
	}

	var eligible, adopted []*sample
	for _, s := range samples {
		if s.entry.Redistributable && !s.entry.Sensitive {
			eligible = append(eligible, s)
		}
	}
	for _, s := range eligible {
		if !slices.ContainsFunc(eligible, func(t *sample) bool {
			return t != s && within(s, t) && (len(t.covers) > len(s.covers) || rank(t, s) < 0)
		}) {
			adopted = append(adopted, s)
		}
	}
	var replaced []*sample
	for _, s := range samples {
		holder := best(filter(adopted, func(t *sample) bool { return within(s, t) }))
		switch {
		case holder == s:
			s.dispose("adopt", "", "")
		case holder != nil:
			s.dispose("trim", holder.entry.Path, "")
		default:
			replaced = append(replaced, s)
		}
	}

	gaps := map[string]*Gap{}
	for i := range coverage.Gaps {
		gaps[coverage.Gaps[i].ID] = &coverage.Gaps[i]
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
	for _, s := range samples {
		if !slices.Contains(replaced, s) {
			continue
		}
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
	return droppable, nil
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
