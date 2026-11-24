package disposition

import (
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
)

type sample struct {
	obj    *jsonvalue.Object
	path   string
	covers []string
	set    idset.Set
	truth  bool
	verify bool
	redist bool
	sens   bool
	bytes  float64
	gap    string
}

func ids(v any) []string {
	list, _ := v.([]any)
	out := []string{}
	for _, item := range list {
		out = append(out, jsonvalue.String(item))
	}
	return out
}

func readSample(v any) (*sample, error) {
	o, ok := v.(*jsonvalue.Object)
	if !ok {
		return nil, errors.New("samples: an entry is not an object")
	}
	s := &sample{obj: o, path: o.Str("path"), gap: o.Str("gap"), set: idset.Set{}}
	for _, field := range []string{"library", "traits", "features"} {
		s.set.Add(ids(o.Get(field))...)
	}
	s.covers = s.set.Sorted("")
	truth, _ := o.Get("truth").([]any)
	s.truth = len(truth) > 0
	s.verify = o.Str("license_status") == "verified"
	s.redist = o.Get("redistributable") == true
	s.sens = o.Get("sensitive") == true
	n, ok := o.Get("bytes").(json.Number)
	if !ok {
		return nil, fmt.Errorf("samples: %s has no bytes", s.path)
	}
	s.bytes, _ = n.Float64()
	return s, nil
}

func lacks(has bool) int {
	if has {
		return 0
	}
	return 1
}

func rank(a, b *sample) int {
	return cmp.Or(
		cmp.Compare(lacks(a.truth), lacks(b.truth)),
		cmp.Compare(lacks(a.verify), lacks(b.verify)),
		cmp.Compare(lacks(a.redist), lacks(b.redist)),
		cmp.Compare(a.bytes, b.bytes),
		strings.Compare(a.path, b.path),
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

func set(o *jsonvalue.Object, disposition, field, value string) *jsonvalue.Object {
	out := jsonvalue.NewObject()
	for _, k := range o.Keys() {
		if k != "disposition" && k != "trimmed_for" && k != "gap" {
			out.Set(k, o.Get(k))
		}
	}
	out.Set("disposition", disposition)
	if field != "" {
		out.Set(field, value)
	}
	return out
}

// Dispose derives every sample's disposition in coverage, the parsed coverage data, as
// docs/coverage.md states, and sets it in place with its trimmed_for or gap, and each gap a
// replace names its covers. A replaced sample takes the gap of the replaced sample whose covers
// contain its own, the one the tie-break prefers among those no other replaced sample's covers
// contain, and that gap is the one those samples name in coverage; candidates are left as they
// are. It returns the gaps no replace names whose covers adopted samples all exercise, which may
// be dropped. A group of replaced samples that names no gap, or names two, is an error.
func Dispose(coverage *jsonvalue.Object) ([]string, error) {
	entries, ok := coverage.Get("samples").([]any)
	if !ok {
		return nil, errors.New("samples: is not an array")
	}
	gapList, ok := coverage.Get("gaps").([]any)
	if !ok {
		return nil, errors.New("gaps: is not an array")
	}
	samples := make([]*sample, len(entries))
	for i, e := range entries {
		s, err := readSample(e)
		if err != nil {
			return nil, err
		}
		samples[i] = s
	}

	var eligible, adopted []*sample
	for _, s := range samples {
		if s.redist && !s.sens {
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
	for i, s := range samples {
		holder := best(filter(adopted, func(t *sample) bool { return within(s, t) }))
		switch {
		case holder == s:
			entries[i] = set(s.obj, "adopt", "", "")
		case holder != nil:
			entries[i] = set(s.obj, "trim", "trimmed_for", holder.path)
		default:
			replaced = append(replaced, s)
		}
	}

	gaps := map[string]*jsonvalue.Object{}
	for _, g := range gapList {
		if o, ok := g.(*jsonvalue.Object); ok {
			gaps[o.Str("id")] = o
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
	for i, s := range samples {
		if !slices.Contains(replaced, s) {
			continue
		}
		holder := best(filter(outer, func(t *sample) bool { return within(s, t) }))
		key := strings.Join(holder.covers, " ")
		id, ok := groupGap[key]
		if !ok {
			return nil, fmt.Errorf("%s: replaced, and no gap holds the covers %s; research a candidate and add the gap", holder.path, key)
		}
		named.Add(id)
		gaps[id].Set("covers", holder.covers)
		entries[i] = set(s.obj, "replace", "gap", id)
	}

	exercised := idset.Set{}
	for _, s := range adopted {
		exercised.Merge(s.set)
	}
	var droppable []string
	for _, g := range gapList {
		o, ok := g.(*jsonvalue.Object)
		if !ok || named[o.Str("id")] {
			continue
		}
		if !slices.ContainsFunc(ids(o.Get("covers")), func(id string) bool { return !exercised[id] }) {
			droppable = append(droppable, o.Str("id"))
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

// Place sets each sample's traits and features in coverage to its element in placement, the JSON
// array place-traits prints. A sample placement lacks, or an element no sample has, is an error.
func Place(coverage *jsonvalue.Object, placement []any) error {
	entries, ok := coverage.Get("samples").([]any)
	if !ok {
		return errors.New("samples: is not an array")
	}
	placed := map[string]*jsonvalue.Object{}
	for _, p := range placement {
		o, ok := p.(*jsonvalue.Object)
		if !ok {
			return errors.New("placement: an element is not an object")
		}
		placed[o.Str("path")] = o
	}
	for _, e := range entries {
		o, ok := e.(*jsonvalue.Object)
		if !ok {
			return errors.New("samples: an entry is not an object")
		}
		p, ok := placed[o.Str("path")]
		if !ok {
			return fmt.Errorf("placement: has no element for %s", o.Str("path"))
		}
		o.Set("traits", p.Get("traits"))
		o.Set("features", p.Get("features"))
		delete(placed, o.Str("path"))
	}
	if extra := slices.Sorted(maps.Keys(placed)); len(extra) > 0 {
		return fmt.Errorf("placement: %s is not a sample", extra[0])
	}
	return nil
}

// Run reads docs/coverage.json, relative to the working directory, and writes it to stdout with
// every disposition derived by Dispose, as jsonvalue.Marshal writes it and a newline. args are
// the command line's arguments without the program's name: --placement <file> first sets the
// traits and features from that place-traits output, as Place does. It names each gap Dispose
// finds may be dropped on stderr. An error writes nothing to stdout.
func Run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("dispose", flag.ContinueOnError)
	flags.SetOutput(stderr)
	placementPath := flags.String("placement", "", "a place-traits output whose traits and features replace the data's")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("usage: dispose [--placement <file>]")
	}
	coverage, err := readObject("docs/coverage.json")
	if err != nil {
		return err
	}
	if *placementPath != "" {
		data, err := os.ReadFile(*placementPath)
		if err != nil {
			return err
		}
		v, err := jsonvalue.Parse(data)
		if err != nil {
			return fmt.Errorf("%s: %w", *placementPath, err)
		}
		placement, ok := v.([]any)
		if !ok {
			return fmt.Errorf("%s: is not a JSON array", *placementPath)
		}
		if err := Place(coverage, placement); err != nil {
			return fmt.Errorf("%s: %w", *placementPath, err)
		}
	}
	droppable, err := Dispose(coverage)
	if err != nil {
		return err
	}
	if _, err := stdout.Write(append(jsonvalue.Marshal(coverage), '\n')); err != nil {
		return err
	}
	for _, id := range droppable {
		if _, err := fmt.Fprintf(stderr, "dispose: gap %s: no replace names it and adopted samples exercise all its covers; it may be dropped\n", id); err != nil {
			return err
		}
	}
	return nil
}

func readObject(path string) (*jsonvalue.Object, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	v, err := jsonvalue.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	o, ok := v.(*jsonvalue.Object)
	if !ok {
		return nil, fmt.Errorf("%s: is not a JSON object", path)
	}
	return o, nil
}
