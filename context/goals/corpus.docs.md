# goal · corpus.docs

- **State:** idle
- **Task:** none
- **Branch:** none

## Tasks

1. [x] review
2. [ ] curate
3. [ ] snapshot
4. [ ] baseline

## Decisions

- review: once curate writes `MANIFEST.json`, the check validates its `covers` against the
  vocabulary in `docs/coverage.json`.
- review: email-7 (attachment-by-reference) is an accepted open gap; its candidate is
  `unavailable`, and curate leaves it open.
- review: a copyleft fixture, such as onenote-1's AGPL-3.0-or-later test.onepkg from Joplin, is
  acceptable in the public corpus when it ships unmodified with its license.
- review: a generated candidate may transform adopted samples by its recorded command, as email-8
  adds a rights-managed header and attachment to msg_36.eml.
- review: every candidate whose `fit` is rated Fair (8 of them, among them onenote-2's image-ocr and
  onenote-3's page-revisions) must be confirmed to exercise its covers when curate builds it.
- review: email-8 is generated from the adopted msg_36.eml, whose covers lie inside email-8's, so
  the subset rule and the adopted-inputs rule conflict there; curate settles which gives way.
- review: Cerebro and Athena add no scenarios until their design tasks close.
- review: docs-corpus takes the workspace's LICENSE, Apache-2.0, identical to the other
  repositories', before snapshot makes it public; the generated candidates' `generator_license`
  then names Apache-2.0, since their generator, `tools/coverage/cmd/generate`, is this
  repository's own code. Samples keep their own licenses, recorded per sample.
- review: coverage.json is rebuilt by `mise run place-traits` then `mise run dispose`; curate
  changes placements or dispositions through them, never by hand.
- review: tools/coverage's probe minimums equal the mise.toml pins; a host whose PATH puts older
  system tools ahead of mise's installs must run the programs with mise's paths first.
- review: docproc output encodings such as `windows-1252 (fallback)` and `utf-8 (byte order mark)`
  are not decoded by tools/coverage; fixing it may move placements, so it is curate's to take or
  leave.
- review: dispose never edits gap candidates; a disposition move that trims a generated
  candidate's input fails the check until the candidate is revised.
- plan: ~/docs is never changed. It is the reference for a corpus built from scratch in
  docs-corpus.
- plan: `MANIFEST.json` carries `version`, and acceptance runs cite it.
- plan: curate first and publish one v1.
- plan: docs-corpus stays private until snapshot publishes v1; snapshot makes it public.
- plan: once docs-corpus v1 is final, the baseline task reruns docproc and docinfer over it,
  repoints the acceptance citations, and removes ~/docs.
- plan: the corpus has no size cap.
- plan: the corpus is laid out by format directory, and each sample's `covers` indexes its
  scenarios in the fixed vocabulary.

## Pending edits

- coordinator · `context/docs-corpus.md`: the corpus note (what the corpus holds, how it changes,
  how it is shared), cited from `context/README.md`'s Notes.
