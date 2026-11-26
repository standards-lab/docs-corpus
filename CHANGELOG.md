# Changelog

All notable changes to docs-corpus, the corpus the Standards Lab workspace's real workloads run
over, are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Each release's section is headed by its
`vX.Y.Z` tag and holds the release's notes.

## [Unreleased]

### Added

- [`docs/measurements.json`](docs/measurements.json), the hand-written measurement list:
  the targeted acceptance measurements in docproc and docinfer that rerun over corpus inputs.
- Eight measurements and their 80 measured samples, with their truth: docproc's word error rate,
  PDF page text, PDF pages in shared gs processes, peak memory, and office-and-mail run, and
  docinfer's transcription, vision re-probe, and the re-probe's confirmation.
- The `measure` disposition in `docs/coverage.json`: a `~/docs` sample a measurement lists, which
  would otherwise be trimmed, stays in the corpus.
- `measures` in `MANIFEST.json`: each sample or truth file a measurement reads names the ids of
  the measurements that list it. The field is optional, so the schema `version` stays 1.

### Changed

- `mise run manifest` lists the measured samples, with their truth and doc files, as it lists the
  adopted ones.
- `scripts/check.sh` checks the measurement list, accepts measured samples, and checks each
  entry's `measures` against the list. Two samples that share a format and covers are now caught
  after the coverage data passes, and a measured sample is exempt, like a generated candidate's
  inputs, from the rule that no sample's covers lie inside another's.

## [v0.1.0] - 2026-11-26

The first release holds 129 samples in ten format areas: `archive`, `audio`, `email`, `exec`,
`image`, `office`, `onenote`, `pdf`, `text`, and `video`. [`MANIFEST.json`](MANIFEST.json), whose
schema `version` is 1, records each sample's source, license, truth, and covers.
[`docs/coverage.md`](docs/coverage.md) records the scenarios the samples cover and why each one is
in the corpus. The sample files are the release's assets, one uncompressed `<area>.tar` per format
area, which `mise run fetch` downloads and verifies against the manifest before it places them.

### Added

- `MANIFEST.json`, the inventory of the samples and the truth and doc files they link, generated
  by `mise run manifest`.
- The scenario vocabulary, `docs/coverage.json`, and the coverage page, `docs/coverage.md`.
- One `<area>.tar` release asset per format area, written by `mise run pack` and placed by
  `mise run fetch`.

[Unreleased]: https://github.com/standards-lab/docs-corpus/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/standards-lab/docs-corpus/releases/tag/v0.1.0
