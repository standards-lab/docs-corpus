# corpus-docs context

The corpus the workspace's real workloads run over: docproc's and docinfer's acceptance, and
capstone.data-intelligence's. It starts from the local pool at `~/docs` (592 files, 1.7 GB,
`MANIFEST.json` dated 2026-11-06), which docproc's acceptance ran over, and becomes a curated,
versioned corpus in this repository.

## Capability map

- **Inventory** (planned): every file's source, license, truth, and redistribution flag in
  `MANIFEST.json`, with `version` naming the corpus version, and the layout by format directory.
- **Scenario index** (planned): each sample's `covers` tags in a fixed vocabulary of the
  scenarios the underlying libraries exercise, with no two samples sharing a format and scenario
  set.
- **Coverage page** (planned): each library scenario mapped to the samples that exercise it, with
  what was trimmed and what was sourced or generated to fill a gap.
- **Sharing** (planned): the corpus files in Git LFS on this private repository, one git tag per
  version, cloned to the model hosts and to `~/docs`.
- **Baseline** (planned): docproc's and docinfer's run over each version, the reference timings
  capstone.data-intelligence measures against.
- **Check** (built): the manifest and scenario checks once `MANIFEST.json` is present, and
  `lint-context` and `lint-references` ([scripts/check.sh](../scripts/check.sh)), with the tools
  pinned in [mise.toml](../mise.toml). No CI runs it.

## Goal records

`goals/` holds one record per active goal rooted here; `standards-lab/roadmap/index.toml` lists
the goals.
