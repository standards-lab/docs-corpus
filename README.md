# docs-corpus

The versioned, curated corpus that the Standards Lab workspace's real workloads run over: sample
documents, messages, images, audio, and video, each with its source, license, ground truth, and
the scenarios it exercises recorded in `MANIFEST.json`. The repository is public from v0.1.0, and
[`STANDARDS.md`](STANDARDS.md) sets which samples that admits.
[`docs/coverage.md`](docs/coverage.md) records which scenarios the samples cover and why each one
is in the corpus.

Each version is a git tag, such as v0.1.0, and [`CHANGELOG.md`](CHANGELOG.md) holds each
version's notes. The payload, the sample files themselves, is the tag's GitHub release assets, one
archive per format area, as [Distribution](#distribution) describes. Git holds the manifest, the
truth files, the docs, the coverage, and the check. The repository uses no Git LFS.

## The manifest

`MANIFEST.json` is the inventory. Its `version`, 1, is the manifest's schema version, not a
release. It has one entry per corpus file, whose `role` is `sample`, or `truth` or `doc` for a file
a sample links. Files sit under the format directories `archive/`, `audio/`, `email/`, `exec/`,
`image/`, `office/`, `onenote/`, `pdf/`, `text/`, and `video/`. An adopted sample keeps the path it had in `~/docs`, a sourced gap candidate sits at
`<area>/<source>/<name>`, and a generated one at the path its command writes, as
[docs/coverage.md](docs/coverage.md#candidates-in-the-corpus) states.

Each entry records the file's `sha256`, `bytes`, and `format`, as `file -b --mime-type` reports
it, and its provenance: `source`, `url`, `ref`, `license`, `license_url`, `license_read`,
`license_status`, `redistributable`, `derived_from`, and `note`, each present where its source
records it. A sample also records:

- `covers`: its scenario IDs in [`docs/coverage.json`](docs/coverage.json)'s vocabulary;
- `truth`: a list of `path` and `note` pairs, or, when it has no truth, a `truth_note` giving the
  reason;
- `docs`: the doc files it links.

A generated file adds `command`, `inputs`, `generator`, and `generator_license`, and a candidate
adds its `gap`.

The manifest is generated, never edited by hand. `mise run manifest` builds it from three sources
for the adopted samples and their files:

- the adopted samples in `docs/coverage.json`;
- `~/docs`' manifest;
- [`docs/license-readings.json`](docs/license-readings.json): the license of every adopted
  sample, truth file, and doc file, read at its source. Each of those entries records its
  reading's `license_url` and `license_read`, and a reading's `license`, where it has one,
  replaces the license `~/docs` records.

The candidates come from [`docs/candidates.json`](docs/candidates.json), which records what was
fetched or generated for each gap. A candidate takes its provenance from its gap's candidate and
its covers from [`docs/corpus-placement.json`](docs/corpus-placement.json), and a generated
candidate's license follows from its inputs'. To regenerate the manifest, run:

```sh
mise run manifest > MANIFEST.json
```

With `--copy`, it first copies every adopted file it lists from `~/docs`, or `$DOCS_REFERENCE`,
into place and checks each copy's size and sha256. `mise run fetch-candidates` places the sourced
candidates, and `mise run generate-candidates` makes the generated ones once.

Git ignores the format directories, so sample files are never committed. Truth and doc files are
committed with `git add -f`, and [`scripts/check.sh`](scripts/check.sh) enforces both. The
working context, [`context/README.md`](context/README.md), states what is still planned.

## Distribution

A format area is the top directory of a sample's path. Each area with samples has one release
asset, `<area>.tar`, under 2 GiB: an uncompressed tar of that area's sample files, at their
manifest paths in sorted order. Truth and doc files aren't packed, since git carries them. To
place the sample files in a checkout, from a version's release assets or a local directory, run:

```sh
mise run fetch -- <release-url|dir>
```

The release URL is the base the assets download from, such as
`https://github.com/standards-lab/docs-corpus/releases/download/<tag>`. The fetch extracts every
area's tar into a staging directory and verifies each sample file's size and sha256 against
`MANIFEST.json`. It moves the files into place only when every one matches; otherwise it names
each file that is missing or differs, and places nothing.

To write the tars into `<dir>` from a working copy whose sample files are in place, run
`mise run pack -- <dir>`. It fails when a file differs from the manifest or an area's tar would
reach 2 GiB.
