# docs-corpus

The versioned, curated corpus that the Standards Lab workspace's real workloads run over: sample
documents, messages, images, audio, and video, each with its source, license, ground truth, and
the scenarios it exercises recorded in `MANIFEST.json`. The repository is public from v1, so it
holds only samples that can be redistributed. [`docs/coverage.md`](docs/coverage.md) records which
scenarios the samples cover and why each one is in the corpus.

Each version is a git tag. The payload, the sample files themselves, is the tag's GitHub release
assets: one archive per format area, each under 2 GiB. Git holds the manifest, the truth files,
the docs, the coverage, and the check. A fetch step downloads the archives and verifies each
file's sha256 against the manifest. The repository uses no Git LFS.

The corpus arrives with its first version; until then the working context,
[`context/README.md`](context/README.md), states what is planned.
