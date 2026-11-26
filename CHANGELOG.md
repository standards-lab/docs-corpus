# Changelog

All notable changes to docs-corpus, the corpus the Standards Lab workspace's real workloads run
over, are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Each release's section is headed by its
`vX.Y.Z` tag and holds the release's notes.

## [Unreleased]

### Added

- [`docs/baseline.md`](docs/baseline.md), the reference capstone.data-intelligence measures
  against: docproc's and docinfer's runs over each tag's files, starting with v0.2.0.
- [`docs/reference-rows.json`](docs/reference-rows.json), the rows of `~/docs`' manifest for the
  adopted and measured samples and the truth and doc files they link, which `mise run manifest`
  builds their entries from.
- `scripts/check.sh` fails a manifest `url` that is empty, only whitespace, or contains "not
  recorded".

### Changed

- `mise run manifest` reads nothing outside the repository, and `scripts/check.sh` compares
  `MANIFEST.json` with its output on every run.
- `pdf/born-digital/dodm-5200-01.pdf`'s `url` is its issuance PDF, the URL its license reading
  names.
- `mise run place-traits --corpus` reads the baseline's v0.2.0 docproc run by default, and the
  corpus placement cites that run.
- The `vision-reprobe` and `vision-confirmation` measurements list the inputs of their reruns over
  v0.2.0.
- The tests that read samples read the checkout's fetched files, and skip when they aren't
  fetched.

### Removed

- `scripts/place-library.sh`, `place-traits` without `--corpus`, `manifest --copy`,
  `dispose --placement`, and the check's comparisons with `~/docs`.

## [v0.2.0] - 2026-11-26

The second release holds 210 samples in the same ten format areas, 81 of them new: 80 measured
samples, which docproc's and docinfer's targeted acceptance measurements rerun over, and DVIDS
882019, the first sourced input. It adds 22 truth files, for 58 in all, and 101 license readings,
for 250 in all.
No v0.1.0 sample file is removed or changed. The sample files are again the release's assets, one
uncompressed `<area>.tar` per format area.

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
- [`docs/sourced.json`](docs/sourced.json), the hand-written list of sourced inputs: files a
  measurement reads that come from outside `~/docs`, each with its provenance and reading.
- The `video-interpretation` measurement over DVIDS 882019, "Ghosts In The Machine", with an
  interpretation key that collects the official account's and three outlets' readings.
- `STANDARDS.md` defines a reading at the source, and directs sourcing to government publications
  first.

### Changed

- `mise run manifest` lists the measured samples, with their truth and doc files, as it lists the
  adopted ones.
- `scripts/check.sh` checks the measurement list, accepts measured samples, and checks each
  entry's `measures` against the list. Two samples that share a format and covers are now caught
  after the coverage data passes, and a measured sample is exempt, like a generated candidate's
  inputs, from the rule that no sample's covers lie inside another's.
- `mise.toml` sets `activate_aggressive`, so its tasks run mise's pinned tools ahead of older
  host tools.
- The LibriSpeech, LibriVox, and Bush speech readings are taken at their publishers rather than at
  a mirror or aggregator.
- `docs/license-readings.json` gains 101 readings, for the measured samples' and their truth's
  `~/docs` paths in `audio`, `office`, and `pdf`.

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

[Unreleased]: https://github.com/standards-lab/docs-corpus/compare/v0.2.0...HEAD
[v0.2.0]: https://github.com/standards-lab/docs-corpus/releases/tag/v0.2.0
[v0.1.0]: https://github.com/standards-lab/docs-corpus/releases/tag/v0.1.0
