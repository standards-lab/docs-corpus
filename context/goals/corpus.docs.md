# goal · corpus.docs

- **State:** idle
- **Task:** none
- **Branch:** none

## Tasks

1. [x] review
2. [x] curate
3. [x] docproc-final
4. [x] snapshot
5. [x] measurement-set
6. [x] truth
7. [x] sourcing
8. [x] snapshot-2
9. [ ] baseline

## Decisions

- sourcing: DVIDS 882019 is in the corpus as the input of `video-interpretation`, alongside the
  NASA ISS mp4, to measure the interpretation of subtext. Its stock footage and Danica Dora song
  are unflagged by DVIDS and stand under its unqualified release.
- sourcing: the video's interpretation key is sourced, never authored: a revision adds or rewords
  published readings only, and neither the architect nor a session adds its own interpretation.
- truth: baseline updates a measurement's `acceptance` when it renames or replaces that heading,
  since the check doesn't match headings.
- baseline: when it commits the ~/docs manifest rows, corrects dodm-5200-01's `url` from the
  placeholder `https://www.esd.whs.mil/Directives/issuances/ (not recorded)` to its issuance PDF,
  the reading's `license_url`; a reading can't set `url`.
- truth: shared gs lists all 60 redistributable PDFs in ~/docs, not the brief's 59, which missed
  `header-no-version-info.pdf` (recorded text/plain); the vision entries list the PDFs the runs
  actually read, not old_scans-4/15/17/30/31/46/63/64.
- truth: peak memory doesn't list `archive/mixed-bundle.tar.zst`; the reruns backlog goal reruns
  peak memory without that 7zz input.
- baseline: waits for v1.ai.hosting's acceptance to merge; runs on the laptop over every file in
  v0.2.0's areas; docinfer goes through the gateway's `docinfer` key at the caps acceptance leaves;
  a run with any Azure fallback is rerun.
- baseline: records `docs/baseline.md`, one section per tag, the raw runs in
  `~/runs/docs-corpus/baseline/<tag>`; it releases no tag of its own.
- baseline: replaces both acceptance pages' whole-corpus run sections with the v0.2.0 runs and
  reruns PDF page text, the vision re-probe, and its confirmation; the other targeted sections
  stay, dated to the retired ~/docs pool, their reruns a backlog goal. This overrides
  docproc-final's "as they closed" for those sections.
- baseline: rewrites docproc's STANDARDS.md ~/docs gloss, puts the testdata READMEs' origins in the
  past tense, corrects sevenzip's testdata README, reruns payload-measurement; commits docinfer's
  `~/runs/docinfer/tools` scripts; rebuilds ai-hosting's bench dense pages from its docproc run.
- baseline: deletes ~/docs outright, with no archive, and every run built from it, including
  `~/runs/docs-corpus/docproc-final/{docs,corpus}`, only after truth merges and a workspace grep
  finds nothing live reading them; the closed spikes stay as they closed.
- baseline: held gotenberg at 9.10.1-libreoffice: 9.10.2 changes nothing docproc uses, and
  docproc releases after v0.9.0 go to a later corpus goal.
- baseline: supplies the v0.2.0 times that the data-intelligence pending edit states inline.
- baseline: before removing ~/docs, commits the ~/docs manifest rows the corpus's samples need,
  so `mise run manifest` regenerates without ~/docs, and retires `place-library`, `place-traits`
  without `--corpus`, and the check's ~/docs comparisons (round 1, question 9).
- snapshot-2: gb10-2 holds the verified v0.2.0 checkout at `~/docs-corpus` (tag commit
  `ceddd30`), a fresh anonymous clone fetched from the release (210 samples); its check ended "All
  checks passed." without ~/docs.
- docproc-final: docproc's acceptance and the onenote spike's findings stay as they closed; both
  are closed out, so their citations of the timing sections are not edited.
- docproc-final: the realistic-content standard admits limit checks whose content is real
  (archive-8, pdf-7, `archive/high-ratio-200mb.zip`, `text/json/deeply-nested-64.json`) and
  `text/csv/tab-separated.tsv` (real city names, random values); it rejects scale, stress, and
  timing inputs whose content is templated, random, or pseudo-words.
- plan: docproc releases after v0.9.0, docproc-coverage's last, go to a later corpus goal.
- curate: the vocabulary's links stay pinned to docproc `48216e1`; an ID a newer run adds links to
  that run's release.
- curate: the candidates for archive-1, archive-2, archive-3, and archive-5 reuse `~/docs` paths
  (`archive/mixed-bundle.7z`, `.tar.gz`, `.tar.zst`, `archive/zip-in-zip.zip`) with different
  content, so baseline must not compare v0.1.0 runs with `~/docs` runs by path alone.
- review: a copyleft fixture, such as onenote-1's AGPL-3.0-or-later test.onepkg from Joplin, is
  acceptable in the public corpus when it ships unmodified with its license.
- review: `mise run place-traits` then `mise run dispose` rebuild coverage.json; a later task
  changes placements or dispositions through them, never by hand.
- plan: ~/docs is never changed before baseline removes it. It is the reference for a corpus
  built from scratch in docs-corpus.
- plan: acceptance runs cite the release tag, such as v0.1.0; `MANIFEST.json`'s `version` is the
  schema version, 1, not a release.
- plan: the corpus has no size cap.

## Pending edits

- coordinator · `context/docs-corpus.md`: add the corpus note. It says that docs-corpus holds the
  real workloads' corpus: redistributable samples in ten format areas, each with its source,
  license, truth, and scenarios in `MANIFEST.json`. It says how the corpus changes: a goal's task
  adds, replaces, or drops samples under docs-corpus's STANDARDS.md, and releases the change under
  a new version tag, since a released tag is never re-cut. A sample stays when it exercises a
  scenario no other sample covers, or when a measurement in `docs/measurements.json` reads it. The
  manifest and coverage are never edited by hand: its `mise` tasks regenerate them, without
  `~/docs`, which baseline removed. It says how the corpus is shared: the repository is public,
  each version tag's release assets hold the sample files, one tar per format area, from v0.1.0,
  and `mise run fetch` places them in a checkout once each matches the manifest. It says that gb10-2
  holds the verified v0.2.0 checkout at `~/docs-corpus`, and that docs-corpus's `docs/baseline.md`
  records docproc's and docinfer's baseline runs over it. It links to docs-corpus's
  [README.md](https://github.com/standards-lab/docs-corpus/blob/main/README.md),
  [docs/coverage.md](https://github.com/standards-lab/docs-corpus/blob/main/docs/coverage.md), and
  [docs/baseline.md](https://github.com/standards-lab/docs-corpus/blob/main/docs/baseline.md)
  rather than restating the manifest, the fetch, the scenarios, the gaps, the measurements, or the
  times. `context/README.md`'s Notes gains, after `data-intelligence.md`: "- `docs-corpus.md` —
  the real workloads' corpus: what it holds, how it changes, how it is shared, the verified
  checkout on gb10-2, and the baseline."
- coordinator · `references.md`, the docs-corpus entry: replace its last sentence, "It is built
  from scratch, with the local pool at `~/docs` as its reference, and becomes public when v1 is
  published.", with "It was built from scratch, with the local pool at `~/docs` as its reference.
  The repository is public, and its files are release assets from v0.1.0."
- coordinator · `context/data-intelligence.md`: replace the integration root's "the acceptance run
  over ~/docs (592 files, 1.7 GB), measured against docinfer's baseline: 6.0 h, and a 3.1 s rerun"
  with the acceptance run over docs-corpus v0.2.0, measured against the v0.2.0 times stated inline
  and linking docs-corpus's
  [docs/baseline.md](https://github.com/standards-lab/docs-corpus/blob/main/docs/baseline.md);
  replace v1.ai.hosting's "a ~/docs-scale run" with "a run over docs-corpus v0.2.0".
  `roadmap/capstone.data-intelligence.toml`: replace the summary's "the ~/docs acceptance run" with
  "the acceptance run over docs-corpus v0.2.0, measured against its baseline".
- coordinator · `roadmap/corpus-measurement-reruns.toml`: create the backlog goal (repos docproc
  and docinfer): rerun the targeted acceptance sections that baseline leaves dated to the retired
  `~/docs` pool (docproc's "Word error rate", "PDF pages in shared gs processes", "Peak memory",
  and "The office-and-mail ~/docs run"; docinfer's "Transcription") over docs-corpus's measurement
  inputs, and cite the docs-corpus tag in each; and write docinfer's "Video interpretation"
  acceptance section, which runs the `video-interpretation` measurement over DVIDS 882019 and
  judges the run against its interpretation key. `roadmap/index.toml`: add it to `backlog`.
- coordinator · `roadmap/corpus-olmocr-provenance.toml`: create the backlog goal (repo
  docs-corpus): review the third-party documents inside olmOCR-bench, read today under the
  dataset's ODC-BY license, against STANDARDS.md's reading rule. `roadmap/index.toml`: add it to
  `backlog`.
