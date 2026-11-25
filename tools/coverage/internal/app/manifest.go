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

// Manifest reads docs/coverage.json and docs/license-readings.json, relative to the working
// directory, and the reference manifest under $DOCS_REFERENCE, default ~/docs, and writes
// MANIFEST.json, as manifest.Build builds it, to stdout, in dispose's JSON form. args are the
// command line's arguments without the program's name: --copy first copies every file it lists
// from the reference docs to its path under the working directory, as manifest.Copy does. An error
// writes nothing to stdout.
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
	dir := os.Getenv("DOCS_REFERENCE")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(home, "docs")
	}
	data, err := os.ReadFile(filepath.Join(dir, "MANIFEST.json"))
	if err != nil {
		return fmt.Errorf("%w; set DOCS_REFERENCE to the reference docs", err)
	}
	var reference manifest.Reference
	if err := json.Unmarshal(data, &reference); err != nil {
		return fmt.Errorf("%s: %w", filepath.Join(dir, "MANIFEST.json"), err)
	}
	m, err := manifest.Build(coverage, reference, readings)
	if err != nil {
		return err
	}
	if *copyFiles {
		if err := manifest.Copy(m, dir, "."); err != nil {
			return err
		}
	}
	return encode(stdout, m)
}
