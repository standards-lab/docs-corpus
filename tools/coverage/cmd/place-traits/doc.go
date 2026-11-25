// Command place-traits prints the trait and feature placement of every ~/docs sample as a JSON
// array, sorted by path. Run it from the repository root:
//
//	go run ./tools/coverage/cmd/place-traits [--facts] [--corpus]
//
// It builds every probe adapter first, so a missing or outdated tool fails before any work,
// naming the tool and the fix. It reads the samples and MANIFEST.json from $DOCS_REFERENCE,
// default ~/docs, and docproc's run from $RUNS, default ~/runs/docproc/final/runs. With --corpus,
// it places the samples the repository's MANIFEST.json lists instead, from docproc's run over the
// corpus, read from $RUNS, default ~/runs/docs-corpus/curate/runs, and prints each one's covers.
// With --facts, each element also has the probes' findings. It writes nothing outside a temporary
// directory.
//
// See docs/coverage.md
package main
