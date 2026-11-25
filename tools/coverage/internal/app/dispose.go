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
)

// Dispose reads docs/coverage.json, relative to the working directory, and writes it to stdout
// with every disposition derived by disposition.Dispose, indented by two spaces, with no HTML
// escaped, and a newline. A key disposition.Coverage does not define is an error. args are the
// command line's arguments without the program's name: --placement <file> first sets the traits
// and features from that place-traits output, as disposition.Place does. It names each gap that
// may be dropped on stderr. An error writes nothing to stdout.
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
	droppable, err := disposition.Dispose(&coverage)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(coverage); err != nil {
		return err
	}
	if _, err := stdout.Write(buf.Bytes()); err != nil {
		return err
	}
	for _, id := range droppable {
		if _, err := fmt.Fprintf(stderr, "dispose: gap %s: no replace names it and adopted samples exercise all its covers; it may be dropped\n", id); err != nil {
			return err
		}
	}
	return nil
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
