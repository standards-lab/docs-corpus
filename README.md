# docs-corpus

The versioned, curated corpus that the Standards Lab workspace's real workloads run over: sample
documents, messages, images, audio, and video, each with its source, license, ground truth, and
the scenarios it exercises recorded in `MANIFEST.json`. The repository is public from v1, and
[`STANDARDS.md`](STANDARDS.md) sets which samples that admits.
[`docs/coverage.md`](docs/coverage.md) records which scenarios the samples cover and why each one
is in the corpus.

Each version is a git tag. The payload, the sample files themselves, is the tag's GitHub release
assets: one archive per format area, each under 2 GiB. Git holds the manifest, the truth files,
the docs, the coverage, and the check. A fetch step downloads the archives and verifies each
file's sha256 against the manifest. The repository uses no Git LFS.

## The manifest

`MANIFEST.json` is the inventory. It has `version: 1` and one entry per corpus file, by `role`: a
`sample`, or a `truth` or `doc` file a sample links. Files sit under the format directories
`audio/`, `email/`, `image/`, `office/`, `onenote/`, `pdf/`, `text/`, and `video/`, at the paths
they had in `~/docs`. Each entry records the file's `sha256`, `bytes`, and `format`, as
`file -b --mime-type` reports it, and its provenance: `source`, `url`, `ref`, `license`,
`license_url`, `license_read`, `license_status`, `redistributable`, `derived_from`, and `note`,
each present where its source records it. A sample also records its `covers`, its scenario IDs in
[`docs/coverage.json`](docs/coverage.json)'s vocabulary; its `truth`, a list of `path` and `note`
pairs; a `truth_note` giving the reason when it has no truth; and its `docs`. A generated file
adds `command`, `inputs`, `generator`, and `generator_license`, and a candidate its `gap`.

The manifest is generated, never edited by hand. To regenerate it from the adopted samples in
`docs/coverage.json`, `~/docs`' manifest, and
[`docs/license-readings.json`](docs/license-readings.json), the licenses read at their source
for samples `~/docs` records only as reported, run:

```sh
mise run manifest > MANIFEST.json
```

With `--copy`, it first copies every file it lists from `~/docs`, or `$DOCS_REFERENCE`, into
place and checks each copy's size and sha256. Git ignores the format directories, so sample files
are never committed; truth and doc files are committed with `git add -f`, and
[`scripts/check.sh`](scripts/check.sh) enforces both. The working context,
[`context/README.md`](context/README.md), states what is still planned.
