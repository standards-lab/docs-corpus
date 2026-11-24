# corpus-docs standards

The judgement calls the standards-reviewer applies, beyond what `scripts/check.sh` enforces: every
file matches its `MANIFEST.json` entry, no two samples share a format and scenario set, and
`lint-context` and `lint-references` pass.

- A sample stays only while it exercises a scenario in the underlying libraries that no other
  sample covers.
- A sample is sourced or generated, with its source or generating command recorded; none is
  constructed by hand.
- `architecture/principles/context-architecture.md`: the manifest is the inventory's one home.
- `architecture/principles/comments.md`: the comments in every script and configuration file.
