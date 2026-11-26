# goal · corpus.docs

- **State:** building
- **Task:** docproc-final
- **Branch:** docproc-final

## Tasks

1. [x] review
2. [x] curate
3. [ ] docproc-final
4. [ ] snapshot
5. [ ] baseline

## Task brief · corpus.docs · docproc-final

**Problem**

curate placed the corpus with docproc v0.5.0, and docproc-coverage's final release (v0.9.0,
cmd/docproc/v0.2.7) changes routing, kinds, checks, and outputs for samples in it: HTML, EPUB,
`#!` scripts, and LICENSE text route Text; MPEG-PS and Ogg video gain kinds; ODF, WordPerfect, and
.xlsb go through soffice; docproc-ooxml reads text boxes, comments, and cell notes; TNEF is read; PDF
column padding goes. v1 and baseline should describe the docproc the capstone starts from, so the
set is re-curated once, against that release. And docinfer's acceptance (escalation 7) found
onenote.rs's generated timing sections, whose pseudo-word pages were noise to a model; the corpus
measures accuracy, so filler and timing or stress inputs leave it, under a written standard.

**Behaviors**
1. Both placements come from docproc built at docproc-coverage's final release and its require
   tags: the review's run over ~/docs, recorded in `.library_placement`, and the corpus run,
   recorded in `.corpus_placement`. The run reader reads that release's run-directory schema and
   refuses any other.
2. STANDARDS.md holds the realistic-content standard: "A sample's content is realistic:
   real-language text and images at the density and count a real document of its kind holds.
   Filler (pseudo-words, random characters, repeated or templated pages) and timing or stress
   inputs stay out of the corpus, which measures accuracy, where a model's output on filler is
   noise. A generated sample draws its content from sourced real text and images, never from a
   random-word generator." The generated-candidate `fit` in docs/coverage.md checks it.
3. A sample that breaks a standard records the break, with its reason, in the coverage data, and
   dispose trims it. `Large Desktop.one` and `Large OneDrive.one` record the realistic-content
   break; `Large Desktop.one` leaves the corpus files, MANIFEST.json, and both placements. ~/docs
   and its runs stay unchanged.
4. Every sample and candidate is read against the standard. A sample that fails records the break
   and is trimmed; a candidate that fails leaves the corpus and its gap stays open (archive-11's
   and email-5's are expected to).
5. The review placement is refreshed: place-library and place-traits over the new ~/docs run, then
   dispose re-derives every disposition. A ~/docs sample newly adopted joins the corpus with a
   license reading; one newly trimmed leaves it. A trim that loses truth no other kept sample
   carries is escalated.
6. IDs the new runs observe that the vocabulary lacks join it, each linked at the release that
   defines it; existing links stay at `48216e1`.
7. A gap whose covers the release changes by design (a route or kind renamed, such as exec-4's
   `route:unrouted` → `route:text`) is edited to match, with the reason recorded. Dispose
   open-splits only what the release should exercise but doesn't. exec-4-open closes once the
   release flags its polyglot; onenote-2-open and onenote-3-open are rechecked.
8. Every candidate is placed again from the new corpus run; its bytes stay as recorded (nothing
   regenerates), and its covers contain its gap's.
9. The check passes; dispose, manifest, and place-traits (both modes) reproduce their files; the
   docs and the goal record's pin name the release. The goal record's pending edits remove both
   timing sections from docproc's acceptance citations and the onenote spike's sources.

**Test seams**
- The check's verdict over the working copy; dispose, manifest, and place-traits output.

**Slices**
1. The release: docproc built at its tags, the run reader for its schema (committed), and the
   vocabulary's new IDs, which commit with the re-placement that covers them.
2. The realistic-content standard: STANDARDS.md, the `fit` check, the recorded standard break that
   dispose trims, both timing sections out, and the sweep of every sample and candidate
   (behaviors 2–4).
3. The review refresh over ~/docs: new run, placements, dispositions, adopted and trimmed samples
   moved in or out with their readings (behaviors 5, 6).
4. The corpus re-placement: new run, gap covers edited to the release, open gaps rechecked,
   dispose, manifest (behaviors 7, 8).
5. Docs, the record, and the pending edits (behavior 9).

**Out of scope**
- Tagging, publishing, going public: snapshot's.
- Searching for new candidates beyond the gaps the refresh opens; a scenario no ~/docs sample or
  candidate exercises stays an open gap, including one a failed candidate leaves.
- Editing docproc or the onenote spike: the goal's sync applies the pending edits.
- docproc releases after docproc-coverage's last.

**Door**

Two-way: nothing is tagged or published, and the repository stays private.

## Progress

slices 1/5 committed, slice 2 in part (93a803b: the standard, `breaks`, lorem-ipsum.epub trimmed
to new open gap archive-12, archive-11's, email-5's, and office-8's candidates out); slice 1's two
kinds wait on the re-placement; Large Desktop.one's and LargeData.one's breaks wait on slices 3–4,
since trimming each brings back a candidate or sample no current run placed (Printout.one,
OneWithFileData.one) · standards — · spec — · editor —
- escalation, slice 1: v1.ai.inference's sync (standards-lab #222) added the OneNote timing-section
  removal and the realistic-content standard to docproc-final. Fold both in? yes; trim through a
  recorded standard break dispose reads? yes; docproc's and the onenote spike's edits as pending
  edits? yes; sweep every sample and candidate now? yes. The amended brief was approved.
- escalation, slice 2: does text/csv/tab-separated.tsv (real city names, random values, gen_text.py)
  pass the realistic-content standard? yes; office-1, office-2, and pdf-1 stay.

## Decisions

- docproc-final: a sample that breaks a standard records the break in the coverage data, and
  dispose trims it; dispositions still never change by hand.
- docproc-final: onenote.rs's timing sections, `Large Desktop.one` and `Large OneDrive.one`, are
  out of the corpus under the realistic-content standard; ~/docs keeps them.
- docproc-final: started on docproc-coverage's released tags before that goal's coordinator
  sync; only the coordinator note waits on the sync.
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
- docproc · docs/acceptance.md: remove the timing rows and citations of `Large Desktop.one` and
  `Large OneDrive.one` (docproc-final, realistic-content standard).
- spike-ai-pipeline-onenote · README, bench/findings.rs, scripts/findings.sh, and
  docs/findings/onenote-cli: remove both timing sections from the sources.
