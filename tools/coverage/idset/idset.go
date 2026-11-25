package idset

import (
	"slices"
	"strings"
)

// Set is a set of vocabulary IDs, such as trait:charset.declared or feature:tables.
type Set map[string]bool

// New returns a Set holding ids.
func New(ids ...string) Set {
	s := Set{}
	s.Add(ids...)
	return s
}

// Add adds ids to the set.
func (s Set) Add(ids ...string) {
	for _, id := range ids {
		s[id] = true
	}
}

// Merge adds other's IDs to the set.
func (s Set) Merge(other Set) {
	for id := range other {
		s[id] = true
	}
}

// Sorted returns the set's IDs that start with prefix, sorted, and an empty slice, never nil, when
// there are none.
func (s Set) Sorted(prefix string) []string {
	out := []string{}
	for id := range s {
		if strings.HasPrefix(id, prefix) {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out
}
