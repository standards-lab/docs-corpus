# docs-corpus standards

The judgement calls the standards-reviewer applies, beyond what `scripts/check.sh` enforces.

- A sample stays only while it exercises a scenario in the underlying libraries that no other
  sample covers. There are two exceptions. A generated sample's recorded inputs stay beside it, as
  msg_36.eml does beside email-8 and old_scans-69.pdf beside pdf-2, because a generated sample must
  be rebuildable from v0.1.0. A measurement input, a redistributable file that a measurement in
  [`docs/measurements.json`](docs/measurements.json) reads, stays while that measurement lists it,
  even when another sample covers its scenarios, because the measurement reruns over it. It is
  still realistic and sourced, and its license is still read at its source, as the standards below
  require.
- A sample's content is realistic: real-language text and images at the density and count a real
  document of its kind holds. Filler (pseudo-words, random characters, repeated or templated
  pages) and timing or stress inputs stay out of the corpus, which measures accuracy, where a
  model's output on filler is noise. A generated sample draws its content from sourced real text
  and images, never from a random-word generator.
- A sample is sourced or generated, with its source or generating command recorded; none is
  constructed by hand.
- A sample is in the repository only when its license, read at its source, lets it be
  redistributed, because the repository is public from v0.1.0 ([`README.md`](README.md)). A
  reading at the source is the publisher's own statement at the source's own URL: its license,
  its reuse policy, or the document's releasability paragraph. A statute, a mirror, or an
  aggregator's page is not a reading at the source.
- A sourced file is sought first in government publications, which are realistic and free to
  redistribute: US federal works (17 USC 105), from govinfo.gov first for documents and DVIDS
  first for video, and EU documents under Decision 2011/833/EU. A US government document is
  redistributable only when it is cleared for public release (Distribution Statement A). In a
  government document, a contractor's work, or third-party material the publisher credits or
  flags, has its license read like any other; the publisher's unqualified release of the finished
  work stands as the reading for the rest.
- A sample file is a release asset, never committed to git, as [`README.md`](README.md) states.
- `architecture/principles/context-architecture.md`: the manifest is the inventory's one home, and
  [`docs/coverage.json`](docs/coverage.json) is the scenario vocabulary's.
- `architecture/principles/comments.md`: the comments in every script, Go source, and
  configuration file.
- `architecture/standards/go-elemental/principles/topology-and-naming.md`: `tools/coverage` is a
  nested module of its own. Each `cmd/*` main only initializes and makes one call into the
  composition root, `internal/app`, and the packages are split by growth or dependency weight.
- `architecture/standards/go-elemental/principles/cli-applications.md` does not apply:
  `tools/coverage` is a repository tool built on the standard library and `golang.org/x/text`,
  not a go-cli-sdk application, so it takes no App/Nodes graph and no SDK exit codes. Its
  `internal/app` keeps only topology-and-naming's import direction.
- `architecture/standards/go-elemental/principles/system-tools.md`: every probe runs through a
  `tools/coverage/probe` adapter that finds its tool and checks its version at startup, and runs
  it under the caller's context.
- `architecture/standards/go-elemental/principles/tests-and-docs.md`: `tools/coverage`'s unit
  tier is black-box and hermetic. A test that reads the coverage data skips when that data is
  absent, a test that reads samples reads the checkout's fetched samples and skips when they
  aren't fetched, and each package's `doc.go` documents its API.
- `architecture/standards/go-elemental/principles/release-and-ci.md`: `scripts/check.sh` runs the
  Go gates for `tools/coverage`, with Go and golangci-lint pinned in [`mise.toml`](mise.toml).
- `architecture/standards/go-elemental/principles/dependencies.md`: `tools/coverage` takes the
  standard library and `golang.org/x/text` alone.
