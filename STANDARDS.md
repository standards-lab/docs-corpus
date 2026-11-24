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
- `architecture/principles/comments.md`: the comments in every script and configuration file.
