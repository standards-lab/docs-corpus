package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/standards-lab/docs-corpus/tools/coverage/media"
	"github.com/standards-lab/docs-corpus/tools/coverage/placement"
	"github.com/standards-lab/docs-corpus/tools/coverage/probe"
	"github.com/standards-lab/docs-corpus/tools/coverage/stdmail"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := placeTraits(ctx, os.Args[1:])
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "place-traits:", err)
		os.Exit(1)
	}
}

func placeTraits(ctx context.Context, args []string) error {
	tools, err := probe.New(ctx)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return placement.Placer{
		Docs:     envOr("DOCS_REFERENCE", filepath.Join(home, "docs")),
		Runs:     envOr("RUNS", filepath.Join(home, "runs/docproc/final/runs")),
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
		Facts:   slices.Contains(args, "--facts"),
	}.Run(ctx, os.Stdout)
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
