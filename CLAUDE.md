# docs-corpus

The real workloads' corpus, a `context` project managed with the marathon workflow.

- [`README.md`](README.md): what the corpus is and how to get it.
- [`context/README.md`](context/README.md): the working context.
- [`docs/coverage.md`](docs/coverage.md): the scenarios the corpus covers, and the samples that
  cover them.
- [`STANDARDS.md`](STANDARDS.md): the judgement calls the standards-reviewer applies.
- [`tools/coverage`](tools/coverage): the Go module that places the samples' traits and features,
  derives their dispositions, and generates candidates; each package's `doc.go` documents it.
- `scripts/check.sh`: the check (`mise run check`).
