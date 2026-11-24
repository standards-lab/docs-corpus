# Coverage

Coverage is the record of which library scenarios the corpus exercises and which sample exercises
each one. It decides which samples the corpus keeps and names a candidate for every scenario no
sample covers. It has two parts:

- [`coverage.json`](coverage.json): the data. It holds the scenario vocabulary, each sample's
  library, trait, and feature placement, its disposition, and the gaps. `scripts/check.sh` reads it.
- This page, which explains the data and how the check uses it.

## The vocabulary

The vocabulary in [`coverage.json`](coverage.json), `.vocabulary`, defines every scenario ID once,
on three axes:

- **Library** (`.vocabulary.library`): an identifier that docproc or docinfer defines. These are
  docproc's file kinds, routes, and inspection checks, and docinfer's capabilities. Each entry
  names the identifier and links to its definition, pinned to a commit. A capability's entry also
  has `fed_by`: the routes whose outputs docinfer sends to that capability, with links to the
  docproc stages that emit those outputs and the docinfer lines that take them.
- **Trait** (`.vocabulary.trait`): how a sample's content departs from the plain case. The traits
  are script, text layer, scan, charset, audio, and scale. Each value has a one-line `rule` saying
  when a sample exhibits it.
- **Feature** (`.vocabulary.feature`): a content structure that docproc's readers handle
  separately or list under their known limits, such as a table, a tracked change, or a message
  attachment. Each one has a one-line `rule`.

### IDs

An ID is its group, a colon, and a value:

- A library ID's group is `kind`, `route`, `check`, or `capability`. Its value is the identifier's
  string value, exactly as docproc's run record or docinfer emits it, spaces included, as in
  `kind:utf-8 bom`. So a run record's `kind`, `route`, and `check` values map to IDs with no
  translation.
- A trait ID is `trait:<trait>.<value>`, as in `trait:script.cyrillic`.
- A feature ID is `feature:<name>`, as in `feature:tracked-changes`.

A sample whose content has no departure on a trait carries no ID for that trait. For example, a
sample with only Latin script carries no `trait:script.*` ID.

Some things are not scenarios and get no ID:

- a format alone, a manifest role, or a password;
- a docproc outcome, such as `seen`;
- docproc's `Unknown` kind, the empty kind. Content with no signature is covered by its route.

go-ai, Cerebro, and Athena add no IDs.

## Dispositions

Every sample in `~/docs`, matched by path and sha256 against `~/docs/MANIFEST.json`, gets exactly
one disposition in `.samples`, its `disposition` field:

- **adopt**: the sample joins the corpus. Its covers set is its `library`, `traits`, and
  `features` IDs together.
- **trim**: an adopted sample already covers this sample's scenarios. `trimmed_for` names that
  sample.
- **replace**: the sample can't be adopted, and a gap entry takes over its scenarios. `gap` names
  that gap.

A sample stays only while it exercises a scenario no other sample covers. So a sample is trimmed
when one adopted sample's covers contain all of its covers, whatever that sample's format. No two
adopted samples share a covers set, and so none share a format and a covers set. Where several
samples could stay, the one that stays, and the one a trim names, is chosen by these tie-breaks,
in order:

1. It has truth.
2. It has a verified license.
3. It is redistributable.
4. It is the smallest.

Each entry records the facts the tie-breaks use, copied from the manifest, so a reader can verify
a trim without `~/docs`: `truth` (it has truth when the list isn't empty), `license_status`,
`redistributable`, and `bytes`. It also records `format`, the manifest's MIME type, and
`sensitive: true` for a file the manifest marks `sensitive` or `local_only`.

Only a redistributable sample that isn't sensitive is adopted. Any other sample is trimmed when an
adopted sample covers all of its scenarios, and replaced otherwise.

### Truth and doc files

A truth file or doc file follows its sample's disposition: it stays while a sample that lists it
is adopted. Each entry lists its files in `truth` and `docs`, linked from the manifest this way:

- **truth**: the paths the sample's own manifest `truth` field names, a folder standing for the
  truth files in it, and `page-N` for each page's file; and any truth file named after the sample,
  as `x.truth.rttm` is beside `x.mp3`. A truth file can serve several samples, such as one
  dataset's transcript file, and a truth file can itself be a sample, as a subtitle file is.
- **docs**: a license, notice, README, or source record applies to the samples in its folder and
  below it. Where it names its subject, it applies to that subject's samples only: a
  `<dataset>.LICENSE.md` to that dataset's files, a license in `office/realworld/licenses/` to
  the office samples whose manifest `source` is that project, `pdf/edge/safedocs.*` and
  `pdf/edge/verapdf.*` to the samples from that source, `unicode-passwords.README.md` to the
  `unicode-password` files, `multilang-rtl.source.html` to the `multilang-rtl` files, and
  `pdf/scanned/`'s README and generator files to the three sets they describe.

A doc that belongs to no sample is listed in `.docs_without_sample` with its `handling`. These are
`~/docs`' own README and research report: they describe the pool as a whole, so they're dropped,
and the corpus's README, its manifest, and this record take their place.

## Library placement

Each entry in `.samples` has the sample's `path` and `sha256`, from `~/docs/MANIFEST.json`, and
`library`: the library IDs the sample exercises. Every `~/docs` sample has an entry, even one with
an empty `library`.

The placements come from docproc's kept run over `~/docs`. `.library_placement` in
[`coverage.json`](coverage.json) records the run's ID and the command that produced them. The
command is a script that runs one jq program over the run's `record.jsonl`. A sample's library IDs
are what the run observed for that file:

- Every event of the file's input counts. That includes the parts docproc found inside the file,
  such as archive members, attachments, and embedded files, so their IDs go to the top-level sample.
- A part's `content_type.kind` gives a `kind:` ID, its `content_type.route` a `route:` ID, and
  each of its `flags[].check` a `check:` ID. The empty kind gives none, and an outcome, such as
  `seen`, gives none.
- A sample gets a `capability:` ID when one of its routes is in that capability's `fed_by`.

To rerun the command and compare its output with the data, run this from the repository root:

```sh
scripts/place-library.sh | cmp - <(jq '.samples | map({path, sha256, library})' docs/coverage.json)
```

The script reads the run from `${RUNS:-$HOME/runs/docproc/final/runs}` and the manifest from
`${DOCS_REFERENCE:-$HOME/docs}`. It writes nothing.

## Trait and feature placement

Each entry's `traits` and `features` are the trait and feature IDs the sample exercises, placed by
reading the file and its manifest entry, never from the manifest's `covers` tags. Each value's
rule in the vocabulary is the test. `scripts/place-traits.py` probes every sample and applies the
rules it can decide from the file or from what docproc's kept run recorded for it:

- PDF text layers from the run's page outputs, OCR layers from invisible text over page images,
  page counts, forms, signature fields, and portfolios, read with pdfinfo, pdfimages, and qpdf;
- audio and video streams, duration, and sample rate, read with ffprobe;
- image frames, alpha, and stored orientation, read with ImageMagick's identify;
- OOXML structures read from the package's parts, with legacy and RTF documents converted to OOXML
  by LibreOffice first, as docproc converts them;
- message structure, parsed from each `.eml` and `.mbox` and taken from the run's records for
  `.msg` and `.pst` files;
- OneNote structures from the run's page records;
- charsets from byte order marks, UTF-8 validity, and uchardet;
- scripts by counting each script's letters in the file's text, or in the text the run extracted.

Where a value takes judgement, such as handwriting, a degraded or simulated scan, spontaneous or
noisy speech, or a layout in columns, the entry's `judged` object holds the ID and the evidence
for it: what was seen in the rendered file, or what its truth or source record shows. The script
adds those IDs to what it probes. To rerun it and compare its output with the data, run this from
the repository root:

```sh
scripts/place-traits.py | cmp - <(jq '.samples | map({path, traits, features})' docs/coverage.json)
```

It reads the same run and manifest as `scripts/place-library.sh`, and writes only to a temporary
directory. With `--facts`, it also prints each sample's probe findings.

## Gaps

Each scenario lists the adopted samples that exercise it. A scenario that no adopted sample
exercises is a gap in `.gaps`. Each gap has an `id`, the `covers` a candidate must exercise, and
a `candidate`. A replaced sample names the gap whose covers contain its own. Replaced samples that
need the same covers, or covers inside another replaced sample's, share one gap. Each gap names
one candidate:

- **sourced**: the candidate's URL, pinned to a commit where one exists, and its license, read
  at the source and recorded as verified.
- **generated**: the command that makes the candidate, the generator, and the generator's license.

Each candidate also gets one line on its quality and fit. None is hand-built. A gap whose
`candidate` is `null` has none yet.

To count the dispositions and gaps, run:

```sh
jq '(.samples | group_by(.disposition) | map({(.[0].disposition): length}) | add) + {gaps: (.gaps | length)}' docs/coverage.json
```

## The check

`scripts/check.sh` validates the data whenever `docs/coverage.json` exists. The `COVERAGE`
environment variable points it at another copy. It fails when:

- a vocabulary entry doesn't have its axis's ID shape;
- a library entry lacks its identifier or its pinned link;
- a trait or feature entry lacks its rule;
- a `covers` ID anywhere in the file, or a sample's trait, feature, or `judged` ID, isn't in the
  vocabulary;
- a capability's `fed_by` route, or a sample's `library` ID, isn't a library ID in the vocabulary;
- a sample lacks its path or sha256, or two entries in `.samples` share a path;
- a sample lacks a tie-break fact, `format`, `truth`, or `docs`;
- a sample has no disposition, or has fields of another disposition;
- a trim names a sample that isn't adopted, one whose covers don't contain its own, or one another
  adopted sample beats on the tie-breaks;
- a replace names a gap that doesn't exist, or one whose covers don't contain its own;
- an adopted sample isn't redistributable or is sensitive;
- two adopted samples share a format and a covers set, or one adopted sample's covers lie inside
  another's or equal another's that wins the tie-breaks;
- a `~/docs` sample is missing from `.samples`, or its entry's sha256 differs from the manifest's;
- an entry in `.samples` isn't a `~/docs` sample;
- an entry's format, bytes, license status, redistributability, or sensitivity differs from the
  manifest's;
- a truth or doc file in the manifest is linked by no sample and isn't in `.docs_without_sample`.

The last four compare against `${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json`, and run only when
that file exists. `~/docs` is removed when the corpus is done, so the check doesn't need it. Without
`~/docs`, the check still verifies every disposition from the facts each entry records. Each
failure message names its fix.

To fix an unknown ID, use an ID the vocabulary defines. If the scenario is new, define it in the
vocabulary, with its link or rule.
