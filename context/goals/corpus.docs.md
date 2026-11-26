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
9. [x] baseline

## Decisions

- sourcing: DVIDS 882019 is in the corpus as the input of `video-interpretation`, alongside the
  NASA ISS mp4, to measure the interpretation of subtext. Its stock footage and Danica Dora song
  are unflagged by DVIDS and stand under its unqualified release.
- sourcing: the video's interpretation key is sourced, never authored: a revision adds or rewords
  published readings only, and neither the architect nor a session adds its own interpretation.
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
