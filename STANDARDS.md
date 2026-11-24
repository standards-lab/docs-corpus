# docs-corpus standards

The judgement calls the standards-reviewer applies, beyond what `scripts/check.sh` enforces.

- A sample stays only while it exercises a scenario in the underlying libraries that no other
  sample covers.
- A sample is sourced or generated, with its source or generating command recorded; none is
  constructed by hand.
- A sample is in the repository only when its license, read at its source, lets it be
  redistributed, because the repository is public from v1 ([`README.md`](README.md)).
- A sample file is a release asset, never committed to git, as [`README.md`](README.md) states.
- `architecture/principles/context-architecture.md`: the manifest is the inventory's one home, and
  [`docs/coverage.json`](docs/coverage.json) is the scenario vocabulary's.
- `architecture/principles/comments.md`: the comments in every script, Go source, and
  configuration file.
- `architecture/standards/go-elemental/principles/topology-and-naming.md`: `tools/coverage` is a
  nested module of its own, its `cmd/*` initialization only, each one call into its composition
  root, `internal/app`, and its packages split by growth or dependency weight.
- `architecture/standards/go-elemental/principles/cli-applications.md` does not apply:
  `tools/coverage` is a repository tool on the standard library alone, not a go-cli-sdk
  application, so it takes no App/Nodes graph and no SDK exit codes. Its `internal/app` keeps
  only topology-and-naming's import direction.
- `architecture/standards/go-elemental/principles/system-tools.md`: every probe runs through a
  `tools/coverage/probe` adapter that finds its tool and checks its version at startup, and runs
  it under the caller's context.
- `architecture/standards/go-elemental/principles/tests-and-docs.md`: `tools/coverage`'s unit
  tier is black-box and hermetic, a test that reads `~/docs` or the coverage data skips when it
  is absent, and each package's `doc.go` documents its API.
- `architecture/standards/go-elemental/principles/release-and-ci.md`: `scripts/check.sh` runs the
  Go gates for `tools/coverage`, with Go and golangci-lint pinned in [`mise.toml`](mise.toml).
- `architecture/standards/go-elemental/principles/dependencies.md`: `tools/coverage` takes the
  standard library and `golang.org/x/text` alone.
