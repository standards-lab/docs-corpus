package idset_test

import (
	"slices"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
)

func TestSortedListsTheIDsWithAPrefix(t *testing.T) {
	s := idset.New("trait:b", "feature:x")
	s.Add("trait:a", "trait:b")
	s.Merge(idset.New("feature:w"))
	if got := s.Sorted("trait:"); !slices.Equal(got, []string{"trait:a", "trait:b"}) {
		t.Errorf("Sorted(trait:) = %v", got)
	}
	if got := s.Sorted(""); !slices.Equal(got, []string{"feature:w", "feature:x", "trait:a", "trait:b"}) {
		t.Errorf("Sorted() = %v", got)
	}
	if got := s.Sorted("kind:"); got == nil || len(got) != 0 {
		t.Errorf("Sorted(kind:) = %#v, want an empty, non-nil list", got)
	}
}
