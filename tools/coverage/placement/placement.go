package placement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
	"github.com/standards-lab/docs-corpus/tools/coverage/media"
	"github.com/standards-lab/docs-corpus/tools/coverage/office"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
	"github.com/standards-lab/docs-corpus/tools/coverage/text"
)

// Placer places every sample the coverage data lists. A nil Charset, Soffice, or Parser, or an
// empty Docs, Runs, or Coverage, panics naming the field when Run is called.
type Placer struct {
	// Docs is the directory the samples and their MANIFEST.json are in, such as ~/docs.
	Docs string
	// Runs is the directory of docproc's kept runs; the content store is its sibling, content.
	Runs string
	// Coverage is the coverage data's path, docs/coverage.json; it names the samples, their
	// judged IDs, and the run, library_placement.run.
	Coverage string
	// Media probes PDFs, audio and video, images, and archives.
	Media media.Prober
	// Charset guesses a text file's charset, as uchardet does.
	Charset text.Guesser
	// Soffice converts legacy documents to OOXML.
	Soffice office.Converter
	// Parser parses .eml and .mbox messages.
	Parser mail.Parser
	// Facts adds each sample's facts and probed IDs to its element.
	Facts bool
}

// Run writes one JSON element per sample to w, sorted by path, with its path and its traits and
// features, the IDs probed and judged, each sorted; with Facts, its facts and probed IDs too.
// The output is an array written as Python's json.dumps writes it with an indent of 2, and a
// newline. It writes nothing outside a temporary directory, which it removes. A sample missing
// from the manifest, a file it cannot read, and a context that ends are errors, and an error
// writes nothing to w.
func (p Placer) Run(ctx context.Context, w io.Writer) (err error) {
	p.require()
	data, err := os.ReadFile(p.Coverage)
	if err != nil {
		return err
	}
	var coverage struct {
		Samples          []sample `json:"samples"`
		LibraryPlacement struct {
			Run string `json:"run"`
		} `json:"library_placement"`
	}
	if err := json.Unmarshal(data, &coverage); err != nil {
		return fmt.Errorf("%s: %w", p.Coverage, err)
	}
	data, err = os.ReadFile(filepath.Join(p.Docs, "MANIFEST.json"))
	if err != nil {
		return err
	}
	var manifest struct {
		Files []struct {
			Path string `json:"path"`
			manifestEntry
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("%s: %w", filepath.Join(p.Docs, "MANIFEST.json"), err)
	}
	entries := map[string]manifestEntry{}
	for _, f := range manifest.Files {
		entries[f.Path] = f.manifestEntry
	}
	rec, err := run.Load(p.Runs, coverage.LibraryPlacement.Run)
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "place-traits-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	pr := &prober{Placer: p, rec: rec, conversions: &office.Conversions{Soffice: p.Soffice, Dir: tmp}}
	slices.SortFunc(coverage.Samples, func(a, b sample) int { return strings.Compare(a.Path, b.Path) })
	out := []any{}
	for _, s := range coverage.Samples {
		if err := ctx.Err(); err != nil {
			return err
		}
		entry, ok := entries[s.Path]
		if !ok {
			return fmt.Errorf("%s: not in %s/MANIFEST.json", s.Path, p.Docs)
		}
		traits, features, facts, err := pr.probe(ctx, s, entry)
		if err != nil {
			return err
		}
		placed := idset.Set{}
		placed.Merge(traits)
		placed.Merge(features)
		for id := range s.Judged {
			placed.Add(id)
		}
		row := jsonvalue.NewObject()
		row.Set("path", s.Path)
		row.Set("traits", placed.Sorted("trait:"))
		row.Set("features", placed.Sorted("feature:"))
		if p.Facts {
			probed := idset.Set{}
			probed.Merge(traits)
			probed.Merge(features)
			row.Set("facts", facts)
			row.Set("probed", probed.Sorted(""))
		}
		out = append(out, row)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err = w.Write(append(jsonvalue.Marshal(out), '\n'))
	return err
}

func (p Placer) require() {
	for _, f := range []struct {
		name string
		ok   bool
	}{
		{"Docs", p.Docs != ""}, {"Runs", p.Runs != ""}, {"Coverage", p.Coverage != ""},
		{"Charset", p.Charset != nil}, {"Soffice", p.Soffice != nil}, {"Parser", p.Parser != nil},
	} {
		if !f.ok {
			panic("placement: Placer.Run requires a " + f.name)
		}
	}
}
