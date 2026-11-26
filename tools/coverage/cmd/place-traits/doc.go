// Command place-traits prints, as a JSON array sorted by path, the covers of every sample the
// repository's MANIFEST.json lists. Run it from the repository root:
//
//	go run ./tools/coverage/cmd/place-traits --corpus [--facts]
//
// It builds every probe adapter first, so a missing or outdated tool fails before any work,
// naming the tool and the fix. It reads the samples from the repository, docproc's run over the
// corpus from $RUNS, default ~/runs/docs-corpus/docproc-final/corpus/runs, and the judged IDs
// docs/judged.json records for samples docs/coverage.json doesn't list. --corpus is required. With
// --facts, each element also has the probes' findings. It writes nothing outside a temporary
// directory.
//
// See docs/coverage.md
package main
