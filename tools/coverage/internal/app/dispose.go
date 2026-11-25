package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
)

// Dispose reads docs/coverage.json, relative to the working directory, and writes it to stdout
// with every disposition derived by disposition.Dispose, as jsonvalue.Marshal writes it and a
// newline. args are the command line's arguments without the program's name: --placement <file>
// first sets the traits and features from that place-traits output, as disposition.Place does.
// It names each gap that may be dropped on stderr. An error writes nothing to stdout.
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
		if err := disposition.Place(coverage, placement); err != nil {
			return fmt.Errorf("%s: %w", *placementPath, err)
		}
	}
	droppable, err := disposition.Dispose(coverage)
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
