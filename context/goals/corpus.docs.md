# goal · corpus.docs

- **State:** idle
- **Task:** none
- **Branch:** none

## Tasks

1. [x] review
2. [x] curate
3. [x] docproc-final
4. [x] snapshot
5. [ ] baseline

## Decisions

- snapshot: the release assets are uploaded while private; docs-corpus goes public with its
  history unchanged, then gb10-2 fetches anonymously into `~/docs-corpus`, kept for baseline.
- docproc-final: docproc's acceptance and the onenote spike's findings stay as they closed; both
  are closed out, so their citations of the timing sections are not edited.
- docproc-final: the realistic-content standard admits limit checks whose content is real
  (archive-8, pdf-7, `archive/high-ratio-200mb.zip`, `text/json/deeply-nested-64.json`) and
  `text/csv/tab-separated.tsv` (real city names, random values); it rejects scale, stress, and
  timing inputs whose content is templated, random, or pseudo-words.
- docproc-final: place-library and `place-traits --corpus` reread the placements' runs, in
  `~/runs/docs-corpus/docproc-final/docs` and `.../corpus`, so baseline keeps them until v0.1.0
  is final.
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
- plan: ~/docs is never changed. It is the reference for a corpus built from scratch in
  docs-corpus.
- plan: snapshot publishes one release, v0.1.0, from the set docproc-final re-curated.
  docs-corpus stays private until then, and snapshot makes it public.
- plan: acceptance runs cite the release tag, such as v0.1.0; `MANIFEST.json`'s `version` is the
  schema version, 1, not a release.
- plan: once docs-corpus v0.1.0 is final, the baseline task reruns docproc and docinfer over it,
  repoints the acceptance citations, and removes ~/docs.
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
- coordinator · `roadmap/corpus.docs.toml`: replace "v1" with v0.1.0 in `criteria` ("docs-corpus
  v0.1.0 is tagged with its release assets, and its check passes over it.", "Every sample in
  v0.1.0 exercises a library scenario no other sample covers.", "gb10-2 holds a verified fetch of
  v0.1.0.", "docproc's and docinfer's runs over v0.1.0 are the capstone's baseline, their
  acceptance cites v0.1.0, and ~/docs is removed."), in `[tasks.snapshot]`'s summary ("Tags
  v0.1.0, publishes …") and proof ("A fresh fetch at v0.1.0 on gb10-2 passes the check."), and in
  `[tasks.baseline]`'s summary ("Runs `docproc` and `docinfer` over v0.1.0 …, repoints docproc's
  and docinfer's acceptance citations to v0.1.0, and removes ~/docs."); also the goal's `summary`
  ("Once v0.1.0's reruns repoint …") and `[tasks.snapshot]`'s name ("Snapshot v0.1.0").
