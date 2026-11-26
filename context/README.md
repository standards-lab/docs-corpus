# docs-corpus context

The corpus the workspace's real workloads run over: docproc's and docinfer's acceptance, and
capstone.data-intelligence's. It starts from the local pool at `~/docs` (592 files, 1.7 GB,
`MANIFEST.json` dated 2026-11-06), which docproc's acceptance ran over, and becomes a curated,
versioned corpus in this repository, public from v0.1.0 ([README.md](../README.md)).

## Capability map

- **Inventory** (built): `MANIFEST.json`, schema `version` 1, records each corpus file's source,
  license, truth, and redistribution flag: the adopted and measured samples, the gap candidates,
  and the truth and doc files they link, laid out by format directory. A file a measurement reads
  names the measurements in `measures`. `mise run manifest` generates it,
  `mise run fetch-candidates` fetches the sourced candidates, and `mise run generate-candidates`
  makes the generated ones once ([README.md](../README.md#the-manifest)).
- **Scenario index** (built): each sample's `covers` in `MANIFEST.json`, in the coverage
  vocabulary. An adopted sample's equal its covers in `docs/coverage.json`; a candidate's are
  placed from docproc's run over the corpus and contain its gap's covers. No sample's covers lie
  inside another's, except a generated sample's inputs and a measured sample
  ([docs/coverage.md](../docs/coverage.md#the-manifest)).
- **Coverage page** (built): the scenario vocabulary, each `~/docs` sample's disposition and the
  standards it breaks, and each gap's candidate, the open gap a candidate leaves, and the sample
  that covers a dropped candidate's gap ([docs/coverage.md](../docs/coverage.md)).
- **Measurement list** (built): `docs/measurements.json`, written by hand, lists the targeted
  acceptance measurements in docproc and docinfer that rerun over corpus inputs, each naming its
  acceptance section by its heading and the inputs it reads. A `~/docs` sample that a measurement
  lists, and that would otherwise be trimmed, is disposed `measure` and stays in the corpus ([docs/coverage.md](../docs/coverage.md#measurements)).
- **Distribution** (built): `mise run pack` writes one uncompressed tar of sample files per format
  area, and `mise run fetch` reads the tars from a release URL or a directory and places the files
  once every one matches the manifest's size and sha256
  ([README.md](../README.md#distribution)). From v0.1.0, the tars are each version tag's release
  assets ([CHANGELOG.md](../CHANGELOG.md)).
- **Baseline** (planned): docproc's and docinfer's run over each version, the reference timings
  capstone.data-intelligence measures against.
- **Check** (built): the Go gates for [tools/coverage](../tools/coverage), from build to
  golangci-lint, the [manifest checks](../docs/coverage.md#the-manifest), the
  [coverage checks](../docs/coverage.md#the-check), and `lint-context` and `lint-references`
  ([scripts/check.sh](../scripts/check.sh)), with the tools pinned in [mise.toml](../mise.toml). No
  CI runs it.

## Goal records

`goals/` holds one record per active goal rooted here; `standards-lab/roadmap/index.toml` lists
the goals.
