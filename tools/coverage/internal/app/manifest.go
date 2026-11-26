package app

import (
	"errors"
	"flag"
	"io"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
	"github.com/standards-lab/docs-corpus/tools/coverage/manifest"
)

// Manifest reads docs/coverage.json, docs/reference-rows.json, docs/license-readings.json,
// docs/candidates.json, docs/corpus-placement.json, docs/measurements.json, and docs/sourced.json,
// relative to the working directory, and writes MANIFEST.json, as manifest.Build builds it, to
// stdout, in dispose's JSON form. args are the command line's arguments without the program's
// name; it takes none. An error writes nothing to stdout.
func Manifest(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("manifest", flag.ContinueOnError)
	flags.SetOutput(stderr)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("usage: manifest")
	}
	var coverage disposition.Coverage
	if err := decodeStrict("docs/coverage.json", &coverage); err != nil {
		return err
	}
	var reference manifest.Reference
	if err := decodeStrict("docs/reference-rows.json", &reference.Files); err != nil {
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
	m, err := manifest.Build(coverage, reference, readings, candidates, placed, measurements, sourced)
	if err != nil {
		return err
	}
	return encode(stdout, m)
}
