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

// PlaceTraits builds every probe adapter, then writes the covers of every sample the repository's
// MANIFEST.json lists to stdout, as placement.Placer.Run writes them. It reads the samples from the
// working directory, docproc's run over the corpus from $RUNS, default
// ~/runs/docs-corpus/baseline/v0.2.0/docproc/runs, and the judged IDs docs/judged.json adds. args
// are the command line's arguments without the program's name: --corpus is required, and --facts
// adds the probes' findings. Any other argument, or none, is refused, with flag's usage on stderr
// for an unknown flag, before any probe adapter is built. An error writes nothing to stdout.
func PlaceTraits(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("place-traits", flag.ContinueOnError)
	flags.SetOutput(stderr)
	facts := flags.Bool("facts", false, "add the probes' findings to each element")
	corpus := flags.Bool("corpus", false, "place the samples MANIFEST.json lists, from docproc's run over the corpus, and write their covers")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 || !*corpus {
		return errors.New("usage: place-traits --corpus [--facts]")
	}
	tools, err := probe.New(ctx)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return placement.Placer{
		Root:     ".",
		Runs:     envOr("RUNS", filepath.Join(home, "runs/docs-corpus/baseline/v0.2.0/docproc/runs")),
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
		Judged:  "docs/judged.json",
	}.Run(ctx, stdout)
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
