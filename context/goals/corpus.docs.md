# goal · corpus.docs

- **State:** idle
- **Task:** none
- **Branch:** none

## Tasks

1. [ ] review
2. [ ] curate
3. [ ] snapshot
4. [ ] baseline

## Task brief · review

**Problem**

~/docs is the reference that docs-corpus will be built from from scratch, and nothing yet says
which of its samples earn a place. Its `covers` tags are free-form: 191 tags, 88 of them used once,
and 147 samples have none. check.sh's duplicate rule already fails on 298 of its 418 samples. The
curate task needs one record that, for each scenario the libraries exercise, names the sample that
exercises it. That record decides what is adopted, trimmed, or replaced, and names a candidate for
every gap. docs-corpus is public, so only redistributable samples can be adopted.

**Behaviors**

1. docs-corpus holds its coverage as two parts: a data file the check reads, and a docs page that
   explains it and outlives v1. The vocabulary defines every scenario ID once, on three axes:
   - **Library**: docproc's file kinds, routes and inspection checks, and docinfer's capabilities.
     Each ID is named after the library identifier it stands for and links to where that
     identifier is defined.
   - **Trait**: script, text layer, scan, charset, audio, scale.
   - **Feature**: content structures that docproc's readers handle separately or list under their
     known limits, such as tables, charts, comments, tracked changes, footnotes, forms, embedded
     objects and attachments.

   Each trait or feature value has a one-line rule saying when a sample exhibits it. No ID names
   only a format, a role, or a password. docproc's "seen" outcome is not a scenario. go-ai, Cerebro
   and Athena add no IDs.
2. Every one of the 418 ~/docs samples, matched by path and sha256 against ~/docs' manifest, has
   exactly one disposition:
   - **adopt**, with its covers set;
   - **trim**, naming the adopted sample that covers its scenarios;
   - **replace**, naming the gap entry that takes over its scenarios.

   Each truth and doc file follows its sample's disposition.
3. Kind, Route and Check placements come from docproc's kept ~/docs run, `01a208e5`. The coverage
   records the run ID and the jq command that produced them, and rerunning that command gives the
   same placements. Each docinfer capability is placed by the route that feeds it.
4. Traits and features are placed by reading each file and its manifest entry, never from its old
   `covers` tags alone.
5. No two adopted samples share a format and a covers set. When samples collide, the one that stays
   is chosen in this order: has truth, then has a verified license, then is redistributable, then
   is smallest. Each trim names the sample it lost to.
6. Each of the 61 non-redistributable samples is replaced, not trimmed, unless an adopted sample
   already covers its scenarios. The two sensitive files, the Vortex manual and the SSH session
   log, are never adopted.
7. Each scenario lists the adopted samples that exercise it, or is a gap. Each gap names one
   candidate:
   - **sourced**: its URL, pinned to a commit where one exists, and its license read at the source
     and recorded as verified;
   - **generated**: its command, its generator, and the generator's license.

   Each candidate also gets a line on its quality and fit. None is hand-built.
8. The check validates the coverage data:
   - every covers ID is in the vocabulary;
   - every ~/docs sample has exactly one disposition;
   - every gap and every replace has a candidate with a verified license or a recorded command;
   - no two adopted samples are duplicates.

   lint-context and lint-references pass.
9. The repository describes itself as docs-corpus: public at v1, payload as release assets on
   each version tag, no Git LFS. Its README, CLAUDE.md, STANDARDS.md, context README and
   configuration comment say so, and no LFS attributes remain.

**Test seams**
- The coverage data file, which the check reads with jq. Every behavior is tested through it.

**Slices**
1. Vocabulary: the three axes and their definitions in the data file and the page, the check
   rejecting an unknown ID, the context README's capability map linking the page, and the
   repository's self-description (behavior 9).
2. Library placement: every ~/docs sample placed by Kind, Route, Check and Capability from run
   `01a208e5`, with the recorded command, and the check confirming every sample appears exactly
   once.
3. Traits, features and dispositions: each sample read and placed on the trait and feature axes;
   adopt, trim or replace decided by the tie-break and the redistributable rule; the check
   enforcing no duplicate among adopted samples.
4. Gaps: every scenario with no adopted sample, each with a verified candidate, and the check
   confirming every gap and replace names one.

**Out of scope**

Curate's work: building the set, writing docs-corpus's manifest and covers, and the fetch step.
Also out: renaming corpus-docs to docs-corpus (the `plan` pull request, before this task builds),
release assets, downloading or generating candidates, rerunning docproc or docinfer, and any change
to ~/docs.

**Door**

Two-way. The task writes a data file, a page and a check in docs-corpus, and releases nothing. Any
disposition or candidate can be revised before curate.

## Decisions

- review: the scenario vocabulary is library-anchored IDs (docproc's Kind, Route and Check;
  docinfer's Capability), plus a trait axis (script, text layer, scan, charset, audio, scale) and a
  feature axis (the content structures docproc's readers handle separately or list as limits).
  go-ai adds none.
- review: Cerebro and Athena add no scenarios until their design tasks close.
- review: Kind, Route and Check are placed from docproc's kept run `01a208e5` by a recorded jq
  command; traits and features are placed by reading each file and its manifest entry.
- review: the coverage is one JSON data file (vocabulary, dispositions, gaps) that the check
  validates, plus a docs page that explains it. After curate, the check validates the manifest's
  covers against the same vocabulary.
- review: when samples collide, the one kept is chosen by truth, then verified license, then
  redistributable, then smallest.
- review: a gap candidate's license is verified at its source during review, with its URL pinned
  to a commit where one exists; a generated candidate records its command and its generator's
  license.
- plan: ~/docs is never changed. It is the reference for a corpus built from scratch in
  docs-corpus (github.com/standards-lab/docs-corpus), renamed from corpus-docs by a
  `plan corpus.docs` pull request before review builds.
- plan: docs-corpus is public. Only redistributable samples are adopted; each non-redistributable
  ~/docs sample is replaced; the Vortex manual and the SSH session log are never adopted.
- plan: the payload is GitHub release assets on each version tag, one archive per format area,
  each under 2 GiB. Git holds the manifest, truth, docs, coverage and check; a fetch step verifies
  each sha256 against the manifest. No Git LFS. `MANIFEST.json` carries `version`, and acceptance
  runs cite it.
- plan: curate first and publish one v1.
- plan: docs-corpus stays private until snapshot publishes v1; snapshot makes it public.
- plan: once docs-corpus v1 is final, the baseline task reruns docproc and docinfer over it,
  repoints the acceptance citations, and removes ~/docs.
- plan: a sample stays when it exercises a library scenario no other sample covers; gaps are
  filled by sourced or generated samples alike, judged on quality and fit, never hand-built; no
  size cap.
- plan: layout by format directory, with `covers` as a fixed-vocabulary scenario index.

## Pending edits

- coordinator · `context/docs-corpus.md`: the corpus note (what the corpus holds, how it changes,
  how it is shared), cited from `context/README.md`'s Notes.
