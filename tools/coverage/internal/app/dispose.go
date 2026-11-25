package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
	"github.com/standards-lab/docs-corpus/tools/coverage/manifest"
)

// Dispose reads docs/coverage.json, relative to the working directory, and writes it to stdout
// with every disposition derived by disposition.Dispose, indented by two spaces, with no HTML
// escaped, and a newline. The candidates it places are those docs/candidates.json records that
// docs/corpus-placement.json places, with the covers placed there. A key disposition.Coverage does
// not define is an error. args are the command line's arguments without the program's name:
// --placement <file> first sets the traits and features from that place-traits output, as
// disposition.Place does. It names each gap that may be dropped on stderr. An error writes nothing
// to stdout.
func Dispose(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("dispose", flag.ContinueOnError)
	flags.SetOutput(stderr)
	placementPath := flags.String("placement", "", "a place-traits output whose traits and features replace the data's")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("usage: dispose [--placement <file>]")
	}
	var coverage disposition.Coverage
	if err := decodeStrict("docs/coverage.json", &coverage); err != nil {
		return err
	}
	if *placementPath != "" {
		data, err := os.ReadFile(*placementPath)
		if err != nil {
			return err
		}
		var placement []disposition.Placement
		if err := json.Unmarshal(data, &placement); err != nil {
			return fmt.Errorf("%s: %w", *placementPath, err)
		}
		if err := disposition.Place(&coverage, placement); err != nil {
			return fmt.Errorf("%s: %w", *placementPath, err)
		}
	}
	placed, err := corpusPlaced()
	if err != nil {
		return err
	}
	droppable, err := disposition.Dispose(&coverage, placed)
	if err != nil {
		return err
	}
	if err := encode(stdout, coverage); err != nil {
		return err
	}
	for _, id := range droppable {
		if _, err := fmt.Fprintf(stderr, "dispose: gap %s: no replace names it and adopted samples exercise all its covers; it may be dropped\n", id); err != nil {
			return err
		}
	}
	return nil
}

func corpusPlaced() ([]disposition.Placed, error) {
	var candidates []manifest.Candidate
	if err := decodeStrict("docs/candidates.json", &candidates); err != nil {
		return nil, err
	}
	var rows []manifest.Placed
	if err := decodeStrict("docs/corpus-placement.json", &rows); err != nil {
		return nil, err
	}
	covers := map[string][]string{}
	for _, r := range rows {
		covers[r.Path] = r.Covers
	}
	var placed []disposition.Placed
	for _, c := range candidates {
		if ids, ok := covers[c.Path]; ok {
			placed = append(placed, disposition.Placed{Gap: c.Gap, Path: c.Path, Bytes: c.Bytes, Covers: ids})
		}
	}
	return placed, nil
}

func encode(w io.Writer, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func decodeStrict(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if dec.More() {
		return fmt.Errorf("%s: data after its JSON value", path)
	}
	return nil
}
