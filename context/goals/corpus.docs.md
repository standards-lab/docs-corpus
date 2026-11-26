# goal · corpus.docs

- **State:** building
- **Task:** measurement-set
- **Branch:** measurement-set

## Tasks

1. [x] review
2. [x] curate
3. [x] docproc-final
4. [x] snapshot
5. [ ] measurement-set
6. [ ] truth
7. [ ] snapshot-2
8. [ ] baseline

## Task brief · measurement-set

**Problem**

Acceptance measurements in docproc and docinfer, such as word error rate, PDF page text and the
vision re-probe, ran over ~/docs inputs that the coverage review trimmed only because they repeat an
adopted sample's scenarios. Baseline removes ~/docs, and after that those measurements can't be
rerun. Every rule that admits a sample today requires a scenario no other sample covers: the
standards, the disposition rules, the check and the goal criterion. The corpus needs a second,
narrower way in: an input that a named measurement reads.

**Behaviors**
1. The standards admit a measurement input: a redistributable, realistic, sourced file that a
   listed measurement reads. It stays while that measurement lists it, even when another sample
   covers its scenarios. Its license is still read at its source.
2. The measurement list is a hand-written file of its own, empty in this task. Each entry has:
   - a unique id;
   - the repository and the acceptance section it reruns;
   - a one-line summary of what it measures;
   - whether its inputs need truth;
   - its inputs, as corpus paths. An input may be a sample or a truth file.
3. The check fails on a measurement list that:
   - is malformed;
   - repeats an id;
   - has an entry without its acceptance section or inputs;
   - names an input the coverage data doesn't know;
   - names an input that is not redistributable, is sensitive, or carries a break;
   - names an input without truth in a measurement that needs truth.
4. `dispose` gives the `measure` disposition to a listed sample it would otherwise trim. A listed
   sample with a unique scenario stays `adopt`. A `measure` sample never trims another sample, is
   never trimmed, and counts toward no gap or vocabulary coverage. A listed sample that `dispose`
   can't keep is an error that names it.
5. The manifest lists every `measure` sample with its truth and doc files. Every entry a
   measurement reads names the ids of the measurements that read it in an optional `measures`
   field. An entry that no measurement reads has no `measures`. The schema version stays 1.
6. The check's covers rules skip `measure` samples. Those rules are: shared format and covers,
   covers inside another sample's, and equal covers that lose the tie-break. The license-reading,
   redistributability, git-tracking and payload rules apply to `measure` samples as to any other.
7. The coverage page, the README's manifest section, the working context and the changelog's
   Unreleased section describe the `measure` disposition, the list and `measures`.
8. With the list empty, `dispose` and `manifest` reproduce the committed coverage data and manifest
   unchanged.

**Test seams**
- `dispose` and the manifest builder, on fixture coverage data and fixture measurement lists.
- The check, over the repository.

**Slices**
1. The standard and the list: the standards bullet, the empty measurement list, the check's rules
   for the list, and the coverage page's section on the list.
2. The `measure` disposition:
   - `dispose` reads the list and derives `measure`, with tests.
   - The check accepts `measure` and exempts it from the covers rules for adopted samples.
   - The coverage page's Dispositions section describes it.
3. `measures` in the manifest:
   - The builder lists `measure` samples and writes `measures`, with tests.
   - The check's manifest rules handle the field.
   - The README, context and changelog describe it.

**Out of scope**

Adding any input, writing the seven measurements' entries, license readings, placement reruns and
the release. These belong to `truth` and `snapshot-2`.

**Door**

Two-way. Nothing is tagged, and the change stays unreleased until `snapshot-2`.

## Progress

slices 0/3 committed · standards — · spec — · editor —

## Decisions

- measurement-set: MANIFEST.json stays schema version 1; `measures` is an optional field.
- measurement-set: the hand-written list is `docs/measurements.json`; the `measure` disposition
  goes only to a listed input `dispose` would otherwise trim.
- truth: writes the seven measurements' entries first, then copies about 80 inputs with
  `mise run manifest --copy`; ~/docs and `~/runs/docs-corpus/docproc-final/{docs,corpus}` are
  deleted only after truth merges.
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
- baseline: deletes ~/docs outright, with no archive, and every run built from it, once a workspace
  grep finds nothing live reading them; the closed spikes stay as they closed.
- baseline: held gotenberg at 9.10.1-libreoffice: 9.10.2 changes nothing docproc uses, and
  docproc releases after v0.9.0 go to a later corpus goal.
- baseline: data-intelligence.md gives the v0.2.0 times inline, linking `docs/baseline.md`; the
  same edit updates its "~/docs-scale" line and capstone.data-intelligence.toml's summary.

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
  adds, replaces, or drops samples under docs-corpus's STANDARDS.md, regenerates the manifest and
  coverage with its `mise` tasks, never by hand, and releases the change under a new version tag,
  since a released tag is never re-cut. It says how the corpus is shared: the repository is public,
  each version tag's release assets hold the sample files, one tar per format area, from v0.1.0,
  and `mise run fetch` places them in a checkout once each matches the manifest. It says that
  gb10-2 holds the verified v0.1.0 checkout at `~/docs-corpus`. It links to docs-corpus's
  [README.md](https://github.com/standards-lab/docs-corpus/blob/main/README.md) and
  [docs/coverage.md](https://github.com/standards-lab/docs-corpus/blob/main/docs/coverage.md)
  rather than restating the manifest, the fetch, the scenarios, or the gaps. `context/README.md`'s
  Notes gains, after `data-intelligence.md`: "- `docs-corpus.md` — the real workloads' corpus: what
  it holds, how it changes, how it is shared, and the verified checkout on gb10-2."
- coordinator · `references.md`, the docs-corpus entry: replace its last sentence, "It is built
  from scratch, with the local pool at `~/docs` as its reference, and becomes public when v1 is
  published.", with "It is built from scratch, with the local pool at `~/docs` as its reference.
  The repository is public, and its files are release assets from v0.1.0."
