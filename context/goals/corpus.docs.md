# goal · corpus.docs

- **State:** idle
- **Task:** none
- **Branch:** none

## Tasks

1. [x] review
2. [x] curate
3. [x] docproc-final
4. [ ] snapshot
5. [ ] baseline

## Decisions

- docproc-final: the realistic-content standard admits limit checks whose content is real
  (archive-8, pdf-7, `archive/high-ratio-200mb.zip`, `text/json/deeply-nested-64.json`) and
  `text/csv/tab-separated.tsv` (real city names, random values); it rejects scale, stress, and
  timing inputs whose content is templated, random, or pseudo-words.
- docproc-final: place-library and `place-traits --corpus` reread the placements' runs, in
  `~/runs/docs-corpus/docproc-final/docs` and `.../corpus`, so baseline keeps them until v1 is final.
- plan: docproc releases after v0.9.0, docproc-coverage's last, go to a later corpus goal.
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
- plan: snapshot publishes one v1, from the set docproc-final re-curated. docs-corpus stays
  private until then, and snapshot makes it public.
- plan: acceptance runs cite `MANIFEST.json`'s `version`.
- plan: once docs-corpus v1 is final, the baseline task reruns docproc and docinfer over it,
  repoints the acceptance citations, and removes ~/docs.
- plan: the corpus has no size cap.

## Pending edits

- coordinator · `context/docs-corpus.md`: the corpus note, cited from `context/README.md`'s Notes.
  It says what the corpus holds, how it changes, and how it is shared, and links to docs-corpus's
  README.md (the manifest, and the release tars `mise run fetch` places) and docs/coverage.md
  (the scenarios, gaps, and candidates) rather than restating them.
- docproc · docs/acceptance.md: remove the timing rows and citations of `Large Desktop.one` and
  `Large OneDrive.one` (docproc-final, realistic-content standard).
- spike-ai-pipeline-onenote · README, bench/findings.rs, scripts/findings.sh, and
  docs/findings/onenote-cli: remove both timing sections from the sources.
