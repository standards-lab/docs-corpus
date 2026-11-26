# goal · corpus.docs

- **State:** building
- **Task:** truth
- **Branch:** truth

## Tasks

1. [x] review
2. [x] curate
3. [x] docproc-final
4. [x] snapshot
5. [x] measurement-set
6. [ ] truth
7. [ ] snapshot-2
8. [ ] baseline

## Task brief · truth

**Problem**

The measurement list is empty. Eight targeted acceptance measurements in docproc and docinfer ran
over `~/docs` inputs, and baseline removes `~/docs`. Most of the inputs they read are trimmed
today, because an adopted sample already covers their scenarios. Unless this task lists them,
copies them into the corpus with their truth and license readings, and places them from a docproc
run over the corpus, the measurements can't be rerun once `~/docs` is gone.

**Behaviors**
1. The measurement list holds eight entries, each naming its acceptance section by its heading,
   verbatim: word error rate (docproc, "Word error rate", needs truth); transcription (docinfer,
   "Transcription", needs truth); PDF page text (docproc, "PDF page text", needs truth); vision
   re-probe (docinfer, "Re-probe at 1,120 image tokens"); vision confirmation (docinfer,
   "Confirmation"); shared gs (docproc, "PDF pages in shared gs processes"); peak memory (docproc,
   "Peak memory"); office-and-mail (docproc, "The office-and-mail ~/docs run"). The last five need
   no truth.
2. Each entry lists every redistributable input its section read, adopted samples included:
   - Word error rate: the AMI es2002a mix-headset recording and its eight truth files, and the
     LibriVox recording and its approximate truth.
   - Transcription: the 28 single-window clips with truth: FLEURS ar/hi/ja/ru/zh, LibriSpeech
     dev-clean, MINDS-14 es/fr/ko.
   - PDF page text: the six olmOCR-bench multi-column pages, their truth file, and the CFR volume.
   - Vision re-probe: the 13 PDFs its pages came from.
   - Vision confirmation: the CFR volume and old_scans-69, -70, and -71.
   - Shared gs: all 59 redistributable PDFs.
   - Peak memory: the redistributable inputs of its peak calls; `archive/mixed-bundle.tar.zst` is
     not listed (not redistributable, and archive-3's candidate holds that path).
   - Office-and-mail: the four ODF 1.3 files, both POI `.xlsb` files, the WordPerfect file, the two
     duck-and-cover videos, and the four eurlex decision HTML truth files, whose PDFs shared gs
     lists.
3. No input breaks a standard, is sensitive, or isn't redistributable. Not listed: Earnings-21/22
   audio, `multilang-rtl.pdf`, `lists.libreoffice.rtf`, the `.libreoffice` ODFs (not
   redistributable); `LargeData.one`, `lorem-ipsum.epub` (realistic-content).
4. Every listed sample an adopted sample covers is disposed `measure`; no listed sample stays
   `trim`. A listed sample with a scenario no other sample covers stays adopted.
5. Every measured sample and every truth or doc file it links has a license reading, read at its
   source, with its URL and date: 101 new readings (80 samples, 21 truth files).
6. The manifest lists every measured sample with its truth and doc files, and each file a
   measurement reads names in `measures` the ids of the measurements that list it.
7. The 80 measured sample files are in place in the corpus, each matching its manifest size and
   sha256; their truth files are committed; `~/docs` is unchanged.
8. Corpus placement comes from a docproc run over the corpus at docproc v0.9.0 (command v0.2.7)
   with gotenberg 9.10.1-libreoffice, replaying the parts the earlier corpus run recorded; the
   coverage data records its run ID. Every manifest sample has a placement row whose covers equal
   its manifest covers.
9. `place-traits --corpus`, `dispose`, and `manifest` reproduce the committed placement, coverage
   data, and manifest, and the check passes.
10. The coverage page and the working context describe the measurement list as filled and point to
    it without restating it; the changelog's Unreleased section gains one Added bullet naming the
    eight measurements and the number of measured samples.

**Test seams**
- The check, over the repository, with `~/docs` present.
- `place-traits --corpus`, `dispose`, and `manifest`, each compared with its committed output.

**Slices**
1. The measurement inputs, landing whole because the check fails between its parts: the eight
   entries, the `measure` dispositions, the license readings, the copied samples and truth files,
   the docproc run over the corpus and its run ID, and the regenerated placement, coverage data,
   and manifest. Escalate if a measured sample's corpus-run covers differ from its `~/docs` covers
   in a way `dispose` can't settle.
2. The prose: the coverage page's Measurements section, the working context's measurement list,
   and the changelog's Unreleased bullet.

**Out of scope**

Rerunning any measurement; changing tools/coverage or the check's rules; tagging or publishing
(snapshot-2 releases v0.2.0); deleting `~/docs` or `~/runs/docs-corpus/docproc-final` (baseline);
inputs no redistributable file can stand in for (Earnings audio, the seed-2 vision sample's
OneNote pages).

**Door**

Two-way. Nothing is tagged, and the change stays unreleased until snapshot-2.

## Progress

slices 2/2 committed · standards done · spec done (no gaps) · editor done
- escalation, standards: dodm-5200-01's license reading was a statute, not its source page. The
  architect read the issuance at esd.whs.mil (520001m_vol2.pdf); its Releasability paragraph is
  the reading (92bb03c).

## Decisions

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
- baseline: `mise run place-traits` fails on the laptop, because mise puts the older system
  ffprobe, magick, and bsdtar first on PATH; run, as truth did,
  `PATH=$(mise bin-paths | tr '\n' ':')$PATH go run ./tools/coverage/cmd/place-traits --corpus`.
- baseline: held gotenberg at 9.10.1-libreoffice: 9.10.2 changes nothing docproc uses, and
  docproc releases after v0.9.0 go to a later corpus goal.
- baseline: supplies the v0.2.0 times that the data-intelligence pending edit states inline.
- baseline: before removing ~/docs, commits the ~/docs manifest rows the corpus's samples need,
  so `mise run manifest` regenerates without ~/docs, and retires `place-library`, `place-traits`
  without `--corpus`, and the check's ~/docs comparisons (round 1, question 9).
- snapshot: gb10-2 holds the verified v0.1.0 checkout at `~/docs-corpus` (tag commit `8873b89`),
  cloned anonymously and fetched from the release; its check ended "All checks passed."
- snapshot: the release assets are uploaded while private; docs-corpus goes public with its
  history unchanged, then gb10-2 fetches anonymously into `~/docs-corpus`, kept for baseline.
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
- curate: 24 samples and 26 truth and doc entries record `license_url`
  https://github.com/standards-lab/docs-corpus/blob/main/LICENSE, which resolves only once curate
  merges and LICENSE is on main; snapshot relies on it.
- review: a copyleft fixture, such as onenote-1's AGPL-3.0-or-later test.onepkg from Joplin, is
  acceptable in the public corpus when it ships unmodified with its license.
- review: `mise run place-traits` then `mise run dispose` rebuild coverage.json; a later task
  changes placements or dispositions through them, never by hand.
- plan: ~/docs is never changed before baseline removes it. It is the reference for a corpus
  built from scratch in docs-corpus.
- plan: snapshot publishes one release, v0.1.0, from the set docproc-final re-curated.
  docs-corpus stays private until then, and snapshot makes it public.
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
  inputs, and cite the docs-corpus tag in each. `roadmap/index.toml`: add it to `backlog`.
