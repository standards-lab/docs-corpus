// Command place-traits prints, as a JSON array sorted by path, the trait and feature placement of
// every ~/docs sample, or with --corpus the covers of every sample MANIFEST.json lists. Run it from
// the repository root:
//
//	go run ./tools/coverage/cmd/place-traits [--facts] [--corpus]
//
// It builds every probe adapter first, so a missing or outdated tool fails before any work,
// naming the tool and the fix. It reads the samples and MANIFEST.json from $DOCS_REFERENCE,
// default ~/docs, and docproc's run from $RUNS, default ~/runs/docs-corpus/docproc-final/docs/runs.
// With --corpus, it places the samples the repository's MANIFEST.json lists instead, from
// docproc's run over the corpus, read from $RUNS, default
// ~/runs/docs-corpus/docproc-final/corpus/runs, with the judged IDs docs/judged.json records for
// samples docs/coverage.json doesn't list, and prints each one's covers. With --facts, each element also has the probes' findings. It writes nothing
// outside a temporary directory.
//
// See docs/coverage.md
package main
