package app

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/standards-lab/docs-corpus/tools/coverage/media"
	"github.com/standards-lab/docs-corpus/tools/coverage/placement"
	"github.com/standards-lab/docs-corpus/tools/coverage/probe"
	"github.com/standards-lab/docs-corpus/tools/coverage/stdmail"
)

// PlaceTraits builds every probe adapter, then writes the placement of every sample
// docs/coverage.json lists to stdout, as placement.Placer.Run writes it. It reads the samples
// from $DOCS_REFERENCE, default ~/docs, and docproc's run from $RUNS, default
// ~/runs/docproc/final/runs. args are the command line's arguments without the program's name:
// --facts adds the probes' findings, and --corpus places the samples the repository's
// MANIFEST.json lists instead, from the working directory, with docproc's run from $RUNS, default
// ~/runs/docs-corpus/curate/runs, writing each one's covers. Any other argument is refused, with
// flag's usage on stderr for an unknown flag, before any probe adapter is built. An error writes
// nothing to stdout.
func PlaceTraits(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("place-traits", flag.ContinueOnError)
	flags.SetOutput(stderr)
	facts := flags.Bool("facts", false, "add the probes' findings to each element")
	corpus := flags.Bool("corpus", false, "place the samples MANIFEST.json lists, from docproc's run over the corpus, and write their covers")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("usage: place-traits [--facts] [--corpus]")
	}
	tools, err := probe.New(ctx)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	docs, runs := envOr("DOCS_REFERENCE", filepath.Join(home, "docs")), envOr("RUNS", filepath.Join(home, "runs/docproc/final/runs"))
	if *corpus {
		docs, runs = ".", envOr("RUNS", filepath.Join(home, "runs/docs-corpus/curate/runs"))
	}
	return placement.Placer{
		Docs:     docs,
		Runs:     runs,
		Coverage: "docs/coverage.json",
		Media: media.Prober{
			PDFInfo:   tools.PDFInfo,
			PDFImages: tools.PDFImages,
			PDFFonts:  tools.PDFFonts,
			QPDF:      tools.QPDF,
			FFprobe:   tools.FFprobe,
			Magick:    tools.Magick,
			Bsdtar:    tools.Bsdtar,
		},
		Charset: tools.Uchardet,
		Soffice: tools.Soffice,
		Parser:  stdmail.Parser{},
		Facts:   *facts,
		Corpus:  *corpus,
	}.Run(ctx, stdout)
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
