# Changelog

All notable changes to docs-corpus, the corpus the Standards Lab workspace's real workloads run
over, are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Each release's section is headed by the
version its `vX.Y.Z` tag carries, and is the release's notes.

## [Unreleased]

## [v0.1.0] - 2026-11-26

The first release of the corpus: 129 samples in ten format areas, `archive`, `audio`, `email`,
`exec`, `image`, `office`, `onenote`, `pdf`, `text`, and `video`, each with its source, license,
truth, and covers recorded in [`MANIFEST.json`](MANIFEST.json), whose schema `version` is 1.
[`docs/coverage.md`](docs/coverage.md) records the scenarios the samples cover and why each one is
in the corpus. The sample files are the release's assets, one uncompressed `<area>.tar` per format
area, which `mise run fetch` downloads and verifies against the manifest before it places them.

### Added

- `MANIFEST.json`, the inventory of the samples and the truth and doc files they link, generated
  by `mise run manifest`.
- The scenario vocabulary, `docs/coverage.json`, and the coverage page, `docs/coverage.md`.
- One `<area>.tar` release asset per format area, written by `mise run pack` and placed by
  `mise run fetch`.
