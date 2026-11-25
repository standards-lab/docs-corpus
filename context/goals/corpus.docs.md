# goal · corpus.docs

- **State:** idle
- **Task:** none
- **Branch:** none

## Tasks

1. [x] review
2. [x] curate
3. [ ] docproc-final
4. [ ] snapshot
5. [ ] baseline

## Task brief · corpus.docs · docproc-final

**Problem**

curate placed the corpus with docproc v0.5.0, and docproc-coverage's later releases change
routing, kinds, checks, and outputs for samples in it: HTML, EPUB, `#!` scripts, and LICENSE
text route Text; RAR, xz, bzip2, MPEG-PS, and Ogg video gain kinds; ODF, WordPerfect, and .xlsb go
through soffice; docproc-ooxml reads text boxes, comments, and cell notes; TNEF is read; PDF
column padding goes. v1 and baseline should describe the docproc the capstone starts from, so the
set is re-curated once, against docproc-coverage's final release.

**Behaviors**
1. Both placements come from docproc built at docproc-coverage's final release and its require
   tags: the review's run over ~/docs, recorded in `.library_placement`, and the corpus run,
   recorded in `.corpus_placement`. The run reader reads that release's run-directory schema and
   refuses any other.
2. The review placement is refreshed: place-library and place-traits over the new ~/docs run, then
   dispose re-derives every disposition. A ~/docs sample newly adopted joins the corpus with a
   license reading; one newly trimmed leaves it. A trim that loses truth no other kept sample
   carries is escalated.
3. IDs the new runs observe that the vocabulary lacks join it, each linked at the release that
   defines it; existing links stay at `48216e1`.
4. A gap whose covers the release changes by design (a route or kind renamed, such as exec-4's
   `route:unrouted` → `route:text`) is edited to match, with the reason recorded. Dispose
   open-splits only what the release should exercise but doesn't. exec-4-open closes once the
   release flags its polyglot; onenote-2-open and onenote-3-open are rechecked.
5. Every candidate is placed again from the new corpus run; its bytes stay as recorded (nothing
   regenerates), and its covers contain its gap's.
6. The check passes; dispose, manifest, and place-traits (both modes) reproduce their files; the
   docs and the goal record's pin name the release.

**Test seams**
- The check's verdict over the working copy; dispose, manifest, and place-traits output.

**Slices**
1. The release: docproc built at it, the run reader for its schema, and the vocabulary's new IDs.
   It needs docproc-coverage synced; until then BUILD escalates and hands off.
2. The review refresh over ~/docs: new run, placements, dispositions, adopted and trimmed samples
   moved in or out with their readings (behaviors 2, 3).
3. The corpus re-placement: new run, gap covers edited to the release, open gaps rechecked,
   dispose, manifest (behaviors 4, 5).
4. Docs and the record (behavior 6).

**Out of scope**
- Tagging, publishing, going public: snapshot's.
- Searching for new candidates beyond the gaps the refresh opens; a scenario no ~/docs sample or
  candidate exercises stays an open gap.
- docproc releases after docproc-coverage's last.

**Door**

Two-way: nothing is tagged or published, and the repository stays private.

## Decisions

- plan: docproc-final re-curates the set once, against docproc-coverage's final release, after
  that goal syncs; releases after it go to a later corpus goal. Until then the corpus placement
  pins docproc v0.5.0 (cmd/docproc v0.2.3).
- plan: docproc-final refreshes the ~/docs review placement at that release, not only the
  corpus run, and edits gap covers the release changes by design, recording the reason.
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
