# docs-corpus context

The corpus the workspace's real workloads run over: docproc's and docinfer's acceptance, and
capstone.data-intelligence's. It starts from the local pool at `~/docs` (592 files, 1.7 GB,
`MANIFEST.json` dated 2026-11-06), which docproc's acceptance ran over, and becomes a curated,
versioned corpus in this repository, public from v1 ([README.md](../README.md)).

## Capability map

- **Inventory** (planned): every file's source, license, truth, and redistribution flag in
  `MANIFEST.json`, with `version` naming the corpus version, and the layout by format directory.
- **Scenario index** (planned): each sample's `covers` tags in the coverage vocabulary, with no
  two samples sharing a format and scenario set.
- **Coverage page** (built): the scenario vocabulary, each `~/docs` sample's disposition, and
  each gap's candidate ([docs/coverage.md](../docs/coverage.md)).
- **Distribution** (planned): the payload as release assets on each version tag, and the fetch
  step that verifies it against the manifest ([README.md](../README.md)).
- **Baseline** (planned): docproc's and docinfer's run over each version, the reference timings
  capstone.data-intelligence measures against.
- **Check** (built): the Go gates for [tools/coverage](../tools/coverage), from build to
  golangci-lint, the manifest and scenario checks once `MANIFEST.json` is present, the
  [coverage checks](../docs/coverage.md#the-check), and `lint-context` and `lint-references`
  ([scripts/check.sh](../scripts/check.sh)), with the tools pinned in [mise.toml](../mise.toml). No
  CI runs it.

## Goal records

`goals/` holds one record per active goal rooted here; `standards-lab/roadmap/index.toml` lists
the goals.
