package placement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
	"github.com/standards-lab/docs-corpus/tools/coverage/media"
	"github.com/standards-lab/docs-corpus/tools/coverage/office"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
	"github.com/standards-lab/docs-corpus/tools/coverage/text"
)

// Placer places every sample the coverage data lists, or, with Corpus, every sample the manifest
// lists. A nil Charset, Soffice, or Parser, an empty Docs, Runs, or Coverage, or with Corpus an
// empty Judged, panics naming the field when Run is called.
type Placer struct {
	// Docs is the directory the samples and their MANIFEST.json are in: ~/docs, or with Corpus the
	// repository root.
	Docs string
	// Runs is the directory of docproc's runs; the content store is its sibling, content.
	Runs string
	// Coverage is the coverage data's path, docs/coverage.json; it names the samples, their
	// judged IDs, the vocabulary's capabilities, and the run: library_placement.run, or with
	// Corpus corpus_placement.run.
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
	// Corpus places the samples Docs' MANIFEST.json lists instead of the coverage data's, their
	// library IDs placed from corpus_placement.run as scripts/place-library.sh places them from
	// the kept run, and writes each one's covers: its library, trait, and feature IDs.
	Corpus bool
	// Judged is the path of the judged IDs of samples the coverage data doesn't list,
	// docs/judged.json: by corpus path, each ID with its evidence. Corpus reads it.
	Judged string
}

type coverageData struct {
	Vocabulary struct {
		Library map[string]libraryEntry `json:"library"`
	} `json:"vocabulary"`
	LibraryPlacement struct {
		Run string `json:"run"`
	} `json:"library_placement"`
	CorpusPlacement struct {
		Run string `json:"run"`
	} `json:"corpus_placement"`
	Samples []sample `json:"samples"`
}

type manifestFile struct {
	Path string `json:"path"`
	Role string `json:"role"`
	manifestEntry
}

// Run writes one JSON element per sample to w, sorted by path, with its path and its traits and
// features, the IDs probed and judged, each sorted; with Corpus, its covers, the library IDs the
// run observed with those, sorted, in their place; with Facts, its facts and probed IDs too. The
// output is an array indented by two spaces, with no HTML escaped, and a newline. It writes
// nothing outside a temporary directory, which it removes. With Corpus, a sample's judged IDs are
// the coverage data's for its path and Judged's. A sample missing from the manifest, a file it
// cannot read, coverage data with Corpus that names no corpus_placement.run, and a context that
// ends are errors, and an error writes nothing to w.
func (p Placer) Run(ctx context.Context, w io.Writer) (err error) {
	p.require()
	data, err := os.ReadFile(p.Coverage)
	if err != nil {
		return err
	}
	var coverage coverageData
	if err := json.Unmarshal(data, &coverage); err != nil {
		return fmt.Errorf("%s: %w", p.Coverage, err)
	}
	manifestPath := filepath.Join(p.Docs, "MANIFEST.json")
	data, err = os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var manifest struct {
		Files []manifestFile `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("%s: %w", manifestPath, err)
	}
	entries := map[string]manifestEntry{}
	for _, f := range manifest.Files {
		entries[f.Path] = f.manifestEntry
	}
	samples, runID := coverage.Samples, coverage.LibraryPlacement.Run
	if p.Corpus {
		runID = coverage.CorpusPlacement.Run
		if runID == "" {
			return fmt.Errorf("%s names no corpus_placement.run; run docproc over the corpus and record its run there, as docs/coverage.md states", p.Coverage)
		}
		data, err := os.ReadFile(p.Judged)
		if err != nil {
			return err
		}
		var judged map[string]map[string]json.RawMessage
		if err := json.Unmarshal(data, &judged); err != nil {
			return fmt.Errorf("%s: %w", p.Judged, err)
		}
		samples = corpusSamples(manifest.Files, coverage.Samples, judged)
	}
	rec, err := run.Load(p.Runs, runID)
	if err != nil {
		return err
	}
	fed := capabilities(coverage.Vocabulary.Library)
	tmp, err := os.MkdirTemp("", "place-traits-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	pr := &prober{Placer: p, rec: rec, conversions: &office.Conversions{Soffice: p.Soffice, Dir: tmp}}
	slices.SortFunc(samples, func(a, b sample) int { return strings.Compare(a.Path, b.Path) })
	out := []row{}
	for _, s := range samples {
		if err := ctx.Err(); err != nil {
			return err
		}
		entry, ok := entries[s.Path]
		if !ok {
			return fmt.Errorf("%s: not in %s", s.Path, manifestPath)
		}
		var library idset.Set
		if p.Corpus {
			library = placeLibrary(rec.Events(s.Path), fed)
			s.Library = library.Sorted("")
		}
		traits, features, found, err := pr.probe(ctx, s, entry)
		if err != nil {
			return err
		}
		placed := idset.Set{}
		placed.Merge(traits)
		placed.Merge(features)
		for id := range s.Judged {
			placed.Add(id)
		}
		r := row{Path: s.Path}
		if p.Corpus {
			placed.Merge(library)
			r.Covers = placed.Sorted("")
		} else {
			r.Traits, r.Features = placed.Sorted("trait:"), placed.Sorted("feature:")
		}
		if p.Facts {
			probed := idset.Set{}
			probed.Merge(traits)
			probed.Merge(features)
			r.Facts = found
			r.Probed = probed.Sorted("")
		}
		out = append(out, r)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	return err
}

func corpusSamples(files []manifestFile, placed []sample, judged map[string]map[string]json.RawMessage) []sample {
	byPath := map[string]map[string]json.RawMessage{}
	for _, s := range placed {
		byPath[s.Path] = s.Judged
	}
	var samples []sample
	for _, f := range files {
		if f.Role != "sample" {
			continue
		}
		ids := maps.Clone(byPath[f.Path])
		if ids == nil {
			ids = map[string]json.RawMessage{}
		}
		maps.Copy(ids, judged[f.Path])
		samples = append(samples, sample{Path: f.Path, Judged: ids})
	}
	return samples
}

type row struct {
	Path     string   `json:"path"`
	Traits   []string `json:"traits,omitzero"`
	Features []string `json:"features,omitzero"`
	Covers   []string `json:"covers,omitzero"`
	Facts    *facts   `json:"facts,omitempty"`
	Probed   []string `json:"probed,omitzero"`
}

func (p Placer) require() {
	for _, f := range []struct {
		name string
		ok   bool
	}{
		{"Docs", p.Docs != ""}, {"Runs", p.Runs != ""}, {"Coverage", p.Coverage != ""},
		{"Charset", p.Charset != nil}, {"Soffice", p.Soffice != nil}, {"Parser", p.Parser != nil},
		{"Judged", !p.Corpus || p.Judged != ""},
	} {
		if !f.ok {
			panic("placement: Placer.Run requires a " + f.name)
		}
	}
}
