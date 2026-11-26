# goal · corpus.docs

- **State:** building
- **Task:** baseline
- **Branch:** baseline

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

## Task brief · baseline

**Problem**

The capstone has no baseline over a corpus it can share. docproc's and docinfer's acceptance pages
still cite runs over ~/docs, a local pool. docs-corpus's tools can't regenerate the manifest without
~/docs. ai-hosting's bench is built from a ~/docs run. Until docproc and docinfer run over v0.2.0,
nothing reads the public corpus as the capstone's reference, and neither ~/docs nor the runs built
from it can be removed.

**Behaviors**
1. `mise run manifest` prints the committed MANIFEST.json byte for byte with no ~/docs on the
   machine. It reads committed copies of the ~/docs manifest rows that the corpus's samples, truth
   files and doc files need. In those rows, dodm-5200-01's url is its issuance PDF, the URL its
   license reading names.
2. The check reproduces MANIFEST.json every time, not only when ~/docs exists, and no longer
   compares the coverage data with ~/docs.
3. The check fails a manifest url that is empty, is only whitespace, or says "not recorded", and
   names the fix: record the URL. The committed corpus passes.
4. place-library and `place-traits` without `--corpus` are gone, with their docs, task descriptions
   and tests. `place-traits --corpus` remains.
5. The tests that read ~/docs samples read the corpus checkout's own files at their manifest paths
   instead. When a file isn't fetched, they skip and name the fetch. In a fetched checkout they run
   and pass.
6. docproc v0.9.0 runs over every file in v0.2.0's areas on the laptop, and its raw run is kept in
   the baseline run directory for v0.2.0.
7. `place-traits --corpus` reads baseline's v0.2.0 docproc run by default and reproduces the
   committed corpus placement. The coverage data's corpus placement cites that run. If the
   placement doesn't reproduce, BUILD escalates.
8. docinfer runs over that docproc run on the laptop through the gateway's `docinfer` key, at the
   row caps acceptance left: vision 32, audio 48, embedding 16. Its rerun follows. A run with any
   request that Azure answered, read from the gateway's request log, is run again. Its raw runs
   are kept beside docproc's.
9. docs-corpus's baseline page gains a v0.2.0 section. It holds each run's ID, build, machine,
   configuration, wall time, and counts by outcome and capability, plus the rerun's time. It names
   the run directory and is the page the capstone measures against.
10. ai-hosting's bench builder defaults to baseline's v0.2.0 docproc run. Its dense pages and
    28–30 s audio windows are rebuilt from that run. The committed page list, input manifest, and
    the README's page and window counts match the rebuilt inputs, and the builder's check passes.
11. docproc's acceptance page replaces "The ~/docs run", "Failures by class", and the intro's
    kept-run paragraph with the v0.2.0 run, citing the tag and run ID. "PDF page text" is rerun
    over v0.2.0's inputs and cites the tag. The other targeted sections stay, dated to the retired
    ~/docs pool.
12. docproc's STANDARDS.md no longer glosses ~/docs as live, and its testdata READMEs give their
    ~/docs origins in the past tense. sevenzip's testdata README no longer says the integration
    tier reads ~/docs. The payload measurement is rerun over v0.2.0 and cites the tag.
13. docinfer commits its acceptance recipes: the run checker, the vision sweep and
    transcription-eval tools with their scripts and test, and the acceptance run script. Each takes
    its run directory, input and gateway as arguments, not ~/docs paths, and docinfer's check
    covers them.
14. docinfer's acceptance page replaces "What ran", "Run 1 and its rerun" and "The fresh run
    (runs 3–8)" with the v0.2.0 runs. The earlier runs' Finding/Fix rows become one table, dated to
    the retired ~/docs runs, with their fix commits and no run paths. "Re-probe at 1,120 image
    tokens" and "Confirmation" are rerun over v0.2.0's inputs at 4 in flight through the gateway,
    citing the tag. "Transcription" stays, dated.
15. Each measurement's `acceptance` heading matches the heading of the section it names, and the
    check passes.
16. The goal record's data-intelligence pending edit states v0.2.0's docinfer time and rerun time
    inline, taken from the baseline page.
17. A workspace grep finds no live reader of ~/docs or of any run being deleted. Live means code,
    mise tasks, scripts, tests, and run defaults across every repository, ~/experiments and ~/runs.
    Only then are these deleted, with no archive:
    - ~/docs
    - docproc's final, final-no-synthetic and office-and-mail runs
    - docs-corpus's docproc-final runs, over both ~/docs and the corpus
    - docs-corpus's curate runs
    - docinfer's ~/docs runs, logs, configs and recipe results
    - ~/docinfer-finding

    Afterwards, every touched repository's check passes on the laptop.

**Test seams**
- docs-corpus's check, and `place-traits --corpus` and `manifest` compared with their committed files.
- docproc's, docinfer's and ai-hosting's checks, and the bench builder's check.
- The baseline run directory and the gateway's request log for the runs.
- The workspace grep.

**Slices**
1. **docs-corpus: the corpus stands alone** (behaviors 1–5): manifest rows with dodm-5200-01's url
   fix; retire place-library, `place-traits` without `--corpus`, and the ~/docs comparisons;
   repoint the tests; add the placeholder-URL rule; move the marathon pin from 0.24.0 to 0.24.1, a
   patch with nothing to adapt.
2. **docs-corpus: baseline's docproc run** (6, 7, 9): run docproc over v0.2.0, repoint the corpus
   placement, and write the page's v0.2.0 section with docproc's figures.
3. **ai-hosting: bench inputs from v0.2.0** (10).
4. **docproc: acceptance over v0.2.0 and its follow-ups** (11, 12).
5. **docinfer: recipes, run, acceptance** (8, 13, 14): commit the recipes first, then run docinfer
   and its rerun, and rerun the re-probe and confirmation through them.
6. **docs-corpus: the docinfer figures and headings** (9, 15, 16): add docinfer's figures to the
   baseline page and update the measurement headings. The editor writes the times into the pending
   edit.
7. **Removal** (17): grep first, and delete only on a clean grep; otherwise escalate.

**Out of scope**

- Reruns of the targeted sections left dated: word error rate, shared gs with Identity, peak
  memory, office-and-mail, and transcription. The video-interpretation section belongs to the
  backlog goal.
- docproc releases after v0.9.0.
- Upgrading gotenberg, which stays held.
- Editing the closed spikes.
- Retuning docinfer's concurrency beyond the row caps.
- Any release tag.
- Committing the bench's input files themselves.

**Door**

One-way for the removal: ~/docs and the runs built from it are deleted with no archive, so it is
the last slice and gated on the grep. Every other slice is two-way.

## Progress

slices 6/7 committed (1 ab8b982, 2 f4715ec, 3 ai-hosting a0dd635, 4 docproc f600c44, 5 docinfer bef66ff b731a2e, 6 docs-corpus 5af6c90 docinfer 36c73e2 docproc b63cf71; docinfer run 01a218e4-ddaa-7101-949b-0c9030a50690 8,330.5 s, rerun 01a21964-5204-75c7-9d78-7f13ce1279db 1.4 s, no Azure fallback; docproc run 01a218c9-8847-7ec5-b98e-cf5803c6a989, 123.5 s, placement reproduced) · standards ✓ (docs-corpus d59582a f4d8bac, docinfer ec74b62 dd9804f) · spec ✓ (1 gap, docproc 9d7ea44, closed) · editor ✓
- slice 1, decided without the architect: removed `manifest --copy` and `dispose --placement`, whose only inputs were ~/docs and the retired placement mode.
- slice 6, decided without the architect: `jwz.mbox` stays off `vision-reprobe`; as an adopted gap candidate it is outside the measurement contract (samples and sourced inputs), and it stays in the corpus regardless.
- order: the reviews and the editor run before slice 7, so the one-way removal comes last; the removal waits on the architect's go-ahead.

## Decisions

- baseline: ~/docinfer-finding and docs-corpus's curate runs are deleted with ~/docs's runs.
- sourcing: DVIDS 882019 is in the corpus as the input of `video-interpretation`, alongside the
  NASA ISS mp4, to measure the interpretation of subtext. Its stock footage and Danica Dora song
  are unflagged by DVIDS and stand under its unqualified release.
- sourcing: the video's interpretation key is sourced, never authored: a revision adds or rewords
  published readings only, and neither the architect nor a session adds its own interpretation.
- baseline: deletes ~/docs outright, with no archive, and every run built from it, including
  `~/runs/docs-corpus/docproc-final/{docs,corpus}`, only after truth merges and a workspace grep
  finds nothing live reading them; the closed spikes stay as they closed.
- baseline: held gotenberg at 9.10.1-libreoffice: 9.10.2 changes nothing docproc uses, and
  docproc releases after v0.9.0 go to a later corpus goal.
- snapshot-2: gb10-2 holds the verified v0.2.0 checkout at `~/docs-corpus` (tag commit
  `ceddd30`), a fresh anonymous clone fetched from the release (210 samples); its check ended "All
  checks passed." without ~/docs.
- docproc-final: the realistic-content standard admits limit checks whose content is real
  (archive-8, pdf-7, `archive/high-ratio-200mb.zip`, `text/json/deeply-nested-64.json`) and
  `text/csv/tab-separated.tsv` (real city names, random values); it rejects scale, stress, and
  timing inputs whose content is templated, random, or pseudo-words.
- plan: docproc releases after v0.9.0, docproc-coverage's last, go to a later corpus goal.
- curate: the vocabulary's links stay pinned to docproc `48216e1`; an ID a newer run adds links to
  that run's release.
- review: a copyleft fixture, such as onenote-1's AGPL-3.0-or-later test.onepkg from Joplin, is
  acceptable in the public corpus when it ships unmodified with its license.
- plan: ~/docs is never changed before baseline removes it. It is the reference for a corpus
  built from scratch in docs-corpus.
- plan: the corpus has no size cap.

## Pending edits

- coordinator · `context/docs-corpus.md`: add the corpus note. It says that docs-corpus holds the
  real workloads' corpus: redistributable samples in ten format areas, each with its source,
  license, truth, and scenarios in `MANIFEST.json`. It says how the corpus changes: a goal's task
  adds, replaces, or drops samples under docs-corpus's STANDARDS.md, and releases the change under
  a new version tag, since a released tag is never re-cut. A sample stays when it exercises a
  scenario no other sample covers, or when a measurement in `docs/measurements.json` reads it. The
  manifest and coverage are never edited by hand: its `mise` tasks regenerate them, without
  `~/docs`, which baseline removed. It adds the rules for later changes that docs-corpus doesn't
  state: docproc releases after v0.9.0 go to a later corpus goal; the corpus's docproc runs hold
  gotenberg at 9.10.1-libreoffice, since 9.10.2 changes nothing docproc uses; DVIDS 882019's
  interpretation key, `video-interpretation`'s, is sourced, never authored: a revision adds or
  rewords published readings only; the realistic-content standard admits limit checks whose
  content is real (archive-8, pdf-7, `archive/high-ratio-200mb.zip`,
  `text/json/deeply-nested-64.json`) and `text/csv/tab-separated.tsv`; the vocabulary's links stay
  pinned to docproc `48216e1`, and an ID a newer run adds links to that run's release; a copyleft
  fixture is acceptable when it ships unmodified with its license; and the corpus has no size cap.
  It links docs-corpus's STANDARDS.md for the realistic-content standard and DVIDS's unqualified
  release, `docs/coverage.md` for the vocabulary's pinned links, and `docs/measurements.json` for
  DVIDS 882019 as `video-interpretation`'s input, rather than restating them. It says how the
  corpus is shared: the repository is public, each version tag's release assets hold the sample
  files, one tar per format area, from v0.1.0, and `mise run fetch` places them in a checkout once
  each matches the manifest. It says that gb10-2 holds the verified v0.2.0 checkout at
  `~/docs-corpus`, and that docs-corpus's `docs/baseline.md` records docproc's and docinfer's
  baseline runs over it. It links to docs-corpus's
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
  with "the acceptance run over docs-corpus v0.2.0 (299 files), measured against its baseline:
  docproc's run took 123.5 s, docinfer's 8,330.5 s (2.3 h), and docinfer's cached rerun 1.4 s",
  linking "its baseline" to docs-corpus's
  [docs/baseline.md](https://github.com/standards-lab/docs-corpus/blob/main/docs/baseline.md);
  replace v1.ai.hosting's "a ~/docs-scale run" with "a run over docs-corpus v0.2.0".
  `roadmap/capstone.data-intelligence.toml`: replace the summary's "the ~/docs acceptance run" with
  "the acceptance run over docs-corpus v0.2.0, measured against its baseline".
- coordinator · `roadmap/corpus-measurement-reruns.toml`: create the backlog goal (repos docproc
  and docinfer): rerun the targeted acceptance sections that baseline leaves dated to the retired
  `~/docs` pool (docproc's "Word error rate", "PDF pages in shared gs processes", "Peak memory",
  and "The office-and-mail ~/docs run"; docinfer's "Transcription") over docs-corpus's measurement
  inputs, peak memory without `archive/mixed-bundle.tar.zst`, which its measurement doesn't list,
  and cite the docs-corpus tag in each; and write docinfer's "Video interpretation" acceptance
  section, which runs the `video-interpretation` measurement over DVIDS 882019 and judges the run
  against its interpretation key. `roadmap/index.toml`: add it to `backlog`.
- coordinator · `roadmap/docproc-backlog.toml`: the summary's list gains, after "docproc-ooxml's
  silent out-of-memory exit;", "a multi-page TIFF whose pages are numbered from 10 (corpus.docs
  baseline, 2026-11-26);". docs-corpus's `image/formats/tiff/stella-maberly-3page.tif`, its pages
  numbered 10 to 12, fails: magick names the `+adjoin` PNGs by the TIFF's page numbers, and the
  picture stage looks for `-0.png` (docproc `docs/acceptance.md`, "Failures by class").
- coordinator · `roadmap/corpus-olmocr-provenance.toml`: create the backlog goal (repo
  docs-corpus): review the third-party documents inside olmOCR-bench, read today under the
  dataset's ODC-BY license, against STANDARDS.md's reading rule. `roadmap/index.toml`: add it to
  `backlog`.
- coordinator · `.claude/marathon.toml`, `[workspace] exclusive`: add go-ai, docproc, and docinfer
  to the group, now that corpus.docs no longer touches docproc and docinfer beside the capstone.
  `context/waves.md` (wave C) and `context/ai-strategy.md`: "join … when corpus.docs syncs"
  becomes that they are in the group. If a second active goal then touches the grown group, the
  sync asks the architect first.
