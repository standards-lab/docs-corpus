# goal · corpus.docs

- **State:** idle
- **Task:** none
- **Branch:** none

## Tasks

1. [x] review
2. [x] curate
3. [ ] snapshot
4. [ ] baseline

## Decisions

- curate: the corpus placement pins docproc v0.5.0, with cmd/docproc v0.2.3. docproc v0.6.0
  (RAR, xz, bzip2, MPEG-PS, and Ogg-video kinds; `.ogv` moves from kind ogg to the Ogg-video kind)
  and docproc-coverage's "text as the last resort" release (#! scripts route to text and get the
  polyglot scan) move placements. A later task adopts them, and the second reopens
  exec-4-open's check:polyglot.
- curate: the vocabulary's links stay pinned to docproc `48216e1`; an ID a newer run adds links to
  that run's release.
- curate: the candidates for archive-1, archive-2, archive-3, and archive-5 reuse `~/docs` paths
  (`archive/mixed-bundle.7z`, `.tar.gz`, `.tar.zst`, `archive/zip-in-zip.zip`) with different
  content, so baseline must not compare v1 runs with `~/docs` runs by path alone.
- curate: 24 samples and 26 truth and doc entries record `license_url`
  https://github.com/standards-lab/docs-corpus/blob/main/LICENSE, which resolves only once curate
  merges and LICENSE is on main; snapshot relies on it.
- review: a copyleft fixture, such as onenote-1's AGPL-3.0-or-later test.onepkg from Joplin, is
  acceptable in the public corpus when it ships unmodified with its license.
- review: `mise run place-traits` then `mise run dispose` rebuild coverage.json; a later task
  changes placements or dispositions through them, never by hand.
- plan: ~/docs is never changed. It is the reference for a corpus built from scratch in
  docs-corpus.
- plan: snapshot publishes one v1, from curate's set. docs-corpus stays private until then, and
  snapshot makes it public.
- plan: acceptance runs cite `MANIFEST.json`'s `version`.
- plan: once docs-corpus v1 is final, the baseline task reruns docproc and docinfer over it,
  repoints the acceptance citations, and removes ~/docs.
- plan: the corpus has no size cap.

## Pending edits

- coordinator · `context/docs-corpus.md`: the corpus note, cited from `context/README.md`'s Notes.
  It says what the corpus holds, how it changes, and how it is shared, and links to docs-corpus's
  README.md (the manifest, and the release tars `mise run fetch` places) and docs/coverage.md
  (the scenarios, gaps, and candidates) rather than restating them.
