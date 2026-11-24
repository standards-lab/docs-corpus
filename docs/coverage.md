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
- **Trait** (`.vocabulary.trait`): how a sample's content departs from the plain case, such as its
  script or its text layer. Each value has a one-line `rule` saying when a sample exhibits it.
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

Following [`STANDARDS.md`](../STANDARDS.md), a sample is trimmed when one adopted sample's covers
contain all of its covers, whatever that sample's format. No two
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

dispose, a command of [`tools/coverage`](../tools/coverage/cmd/dispose/doc.go), derives every
disposition by these rules, and the covers of each gap a replace names, as [Gaps](#gaps) states.
It keeps every gap's candidate as it is and adds no gap. To rerun it and compare its output with
the data, run this from the repository root:

```sh
mise run dispose | cmp - docs/coverage.json
```

It writes nothing but its output, and names on standard error each gap that may be dropped. With
`--placement <file>`, it first takes each sample's traits and features from that output of
place-traits, so `mise run place-traits > "$tmp"` followed by `mise run dispose --placement "$tmp"`
re-derives the placements and the dispositions together.

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
rule in the vocabulary is the test. place-traits, a command of the Go module
[`tools/coverage`](../tools/coverage), probes every sample and applies the rules it can decide from
the file or from what docproc's kept run recorded for it:

- PDF text layers from the run's page outputs, OCR layers from invisible text over page images,
  page counts, forms, signature fields, and portfolios, read with pdfinfo, pdfimages, pdffonts, and
  qpdf;
- audio and video streams, duration, and sample rate, read with ffprobe;
- image frames, alpha, and stored orientation, read with ImageMagick's identify;
- OOXML structures read from the package's parts, with legacy and RTF documents converted to OOXML
  by LibreOffice first, as docproc converts them;
- message structure, parsed from each `.eml` and `.mbox` as CPython's `email` package parses it,
  and taken from the run's records for `.msg` and `.pst` files;
- OneNote structures from the run's page records;
- charsets from byte order marks, UTF-8 validity, and uchardet;
- scripts by counting each script's letters in the file's text, or in the text the run extracted.
  A letter counts toward the script its Unicode character name starts with, Han's being CJK and
  kana's Hiragana or Katakana, so modifier letters, fullwidth and halfwidth forms, and letterlike
  symbols such as the Kelvin sign count toward none.

Where a value takes judgement, such as handwriting, a degraded or simulated scan, spontaneous or
noisy speech, or a layout in columns, the entry's `judged` object holds the ID and the evidence
for it: what was seen in the rendered file, or what its truth or source record shows. The program
adds those IDs to what it probes. To rerun it and compare its output with the data, run this from
the repository root:

```sh
mise run place-traits | cmp - <(jq '.samples | map({path, traits, features})' docs/coverage.json)
```

It reads the same run and manifest as `scripts/place-library.sh`, and writes only to a temporary
directory. With `--facts`, as in `mise run place-traits --facts`, it also prints each sample's
probe findings.

[`mise.toml`](../mise.toml) pins Go and every probe the program calls except LibreOffice's
`soffice`, the one unpinned host requirement: it is in neither mise's registry nor conda-forge, so
the program and ImageMagick's EMF and WMF delegate use the host's. The program stops at its start,
naming the tool and the fix, when a probe is not on `PATH` or is older than the version its
[adapter](../tools/coverage/probe/doc.go) accepts.

## Gaps

Each scenario lists the adopted samples that exercise it. A scenario that no adopted sample
exercises is a gap in `.gaps`. Each gap has an `id`, the `covers` a candidate must exercise, and
a `candidate`. A replaced sample names the gap whose covers contain its own. Replaced samples that
need the same covers, or covers inside another replaced sample's, share one gap; where several
replaced samples' covers contain a sample's, it shares the gap of the one the tie-break prefers. A gap for
scenarios no sample exercised groups the IDs one candidate file can exercise together.

The data stores no list of samples per scenario: it follows from the placements. To print each
scenario with the adopted samples that exercise it, or with the gaps that cover it, run:

```sh
jq -r '[.samples[] | select(.disposition == "adopt")] as $adopted | .gaps as $gaps
  | [.vocabulary.library, .vocabulary.trait, .vocabulary.feature] | map(keys[]) | .[] as $id
  | [$adopted[] | select(any(.library[], .traits[], .features[]; . == $id)) | .path] as $paths
  | if $paths != [] then "\($id)\t\($paths | join(" "))"
    else "\($id)\tgap \([$gaps[] | select(any(.covers[]; . == $id)) | .id] | join(" "))" end' \
  docs/coverage.json
```

### Candidates

Each gap names one candidate, its `candidate` object, of one of three kinds, its `kind`:

- **sourced**: a file fetched from a well-known test suite or an openly licensed collection. It
  records the `url` it is fetched from and its `ref`, the commit or revision the URL is pinned
  to; a GitHub URL carries its commit, so it can't drift. It records its `license`, the
  `license_url` of the license as read at the source, `license_read`, the date it was read, and
  `license_status: "verified"`.
- **generated**: a file a recorded command makes. It records the exact `command`, run from the
  repository root, its `inputs`, the `generator` that runs, with its version, and the
  `generator_license`. The inputs are adopted samples only, and a command may read its own
  earlier output, so nothing in a candidate is typed by hand. The candidate's own license
  follows from its inputs': an archive of adopted samples carries their licenses.
- **unavailable**: no file can be sourced or generated for the gap's covers. It records
  `searched`, the suites, files, and searches tried, and `reason`, why none exists, such as a
  structure only a live mail server makes. Curate leaves such a gap open: the corpus doesn't
  exercise its scenarios until a candidate turns up.

Each sourced or generated candidate also has a `fit` line: a rating, Good or Fair, then its
quality and how it exercises the gap's covers.

A suspicious scenario, such as an archive bomb, a path traversal, an executable, or a polyglot, has
a harmless candidate: a parser fixture from an established test suite, or a file built by a
recorded command from benign inputs.

A candidate is a proposal. Fetching or generating it into the corpus, and confirming it with
docproc, is the next step. A gap whose `candidate` is `null` has none yet, and fails the check.

To count the dispositions and gaps, run:

```sh
jq '(.samples | group_by(.disposition) | map({(.[0].disposition): length}) | add) + {gaps: (.gaps | length)}' docs/coverage.json
```

## The check

`scripts/check.sh` validates `docs/coverage.json`, or the copy the `COVERAGE` environment
variable names. It fails when:

- the file is missing, or its `.samples` or `.gaps` isn't an array;
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
- two gaps share an id, or a gap has no candidate;
- a candidate's kind is not sourced, generated, or unavailable, or a sourced or generated
  candidate lacks its fit line;
- a sourced candidate lacks its https URL, its ref, its license, its https license URL, or the
  date its license was read, its license status isn't verified, or its GitHub URL isn't pinned
  to the 40-hex commit its ref names;
- a generated candidate lacks its command, its generator, or the generator's license, or lists
  no inputs, or an input that isn't an adopted sample;
- an unavailable candidate lacks what was searched or the reason;
- a vocabulary ID is exercised by no adopted sample and is in no gap's covers;
- an adopted sample isn't redistributable or is sensitive;
- two adopted samples share a format and a covers set, or one adopted sample's covers lie inside
  another's or equal another's that wins the tie-breaks;
- a `~/docs` sample is missing from `.samples`, or its entry's sha256 differs from the manifest's;
- an entry in `.samples` isn't a `~/docs` sample;
- an entry's format, bytes, license status, redistributability, or sensitivity differs from the
  manifest's;
- an entry links a truth or doc file that isn't in the manifest;
- a truth or doc file in the manifest is linked by no sample and isn't in `.docs_without_sample`.

The last five compare against `${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json`, and run only when
that file exists; a `DOCS_REFERENCE` that names no manifest fails the check. `~/docs` is planned
to be removed once the corpus is final, so the check doesn't depend on it: without `~/docs`, it
still verifies every disposition from the facts each entry records. Each failure message names its fix.

To fix an unknown ID, use an ID the vocabulary defines. If the scenario is new, define it in the
vocabulary, with its link or rule.
