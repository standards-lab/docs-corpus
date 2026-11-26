package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
	"github.com/standards-lab/docs-corpus/tools/coverage/manifest"
)

// Manifest reads docs/coverage.json, docs/license-readings.json, docs/candidates.json,
// docs/corpus-placement.json, docs/measurements.json, and docs/sourced.json, relative to the
// working directory, and the reference manifest under $DOCS_REFERENCE, default ~/docs, and writes
// MANIFEST.json, as manifest.Build builds it, to stdout, in dispose's JSON form. args are the
// command line's arguments without the program's name: --copy first copies every file it lists
// from the reference docs to its path under the working directory, as manifest.Copy does, leaving
// the candidates to mise run fetch-candidates and the sourced inputs and their truth files out. An
// error writes nothing to stdout.
func Manifest(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("manifest", flag.ContinueOnError)
	flags.SetOutput(stderr)
	copyFiles := flags.Bool("copy", false, "copy every listed file from the reference docs into place")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("usage: manifest [--copy]")
	}
	var coverage disposition.Coverage
	if err := decodeStrict("docs/coverage.json", &coverage); err != nil {
		return err
	}
	var readings []manifest.Reading
	if err := decodeStrict("docs/license-readings.json", &readings); err != nil {
		return err
	}
	var candidates []manifest.Candidate
	if err := decodeStrict("docs/candidates.json", &candidates); err != nil {
		return err
	}
	var placed []manifest.Placed
	if err := decodeStrict("docs/corpus-placement.json", &placed); err != nil {
		return err
	}
	var measurements []disposition.Measurement
	if err := decodeStrict("docs/measurements.json", &measurements); err != nil {
		return err
	}
	var sourced []manifest.Sourced
	if err := decodeStrict("docs/sourced.json", &sourced); err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := envOr("DOCS_REFERENCE", filepath.Join(home, "docs"))
	referencePath := filepath.Join(dir, "MANIFEST.json")
	data, err := os.ReadFile(referencePath)
	if err != nil {
		return fmt.Errorf("%w; set DOCS_REFERENCE to the reference docs", err)
	}
	var reference manifest.Reference
	if err := json.Unmarshal(data, &reference); err != nil {
		return fmt.Errorf("%s: %w", referencePath, err)
	}
	m, err := manifest.Build(coverage, reference, readings, candidates, placed, measurements, sourced)
	if err != nil {
		return err
	}
	if *copyFiles {
		if err := manifest.Copy(referenceFiles(m, sourced), dir, "."); err != nil {
			return err
		}
	}
	return encode(stdout, m)
}

func referenceFiles(m manifest.Manifest, sourced []manifest.Sourced) manifest.Manifest {
	outside := map[string]bool{}
	for _, s := range sourced {
		outside[s.Path] = true
		for _, t := range s.Truth {
			outside[t.Path] = true
		}
	}
	out := manifest.Manifest{Version: m.Version}
	for _, f := range m.Files {
		if f.Gap == "" && !outside[f.Path] {
			out.Files = append(out.Files, f)
		}
	}
	return out
}
