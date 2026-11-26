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
- **measure**: a measurement in [`measurements.json`](measurements.json) lists the sample, and an
  adopted sample already covers its scenarios. It stays in the corpus for that measurement, as
  [Measurements](#measurements) states, and names no `trimmed_for` or `gap`.

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
adopted sample covers all of its scenarios, and replaced otherwise. The gap candidates in the
corpus count as adopted samples for trims, so a trim can name a candidate, as
[Candidates in the corpus](#candidates-in-the-corpus) states.

A listed sample that would be trimmed is measured instead, and a listed sample with a scenario no
other sample covers stays adopted. A measured sample doesn't count as adopted: no trim or gap's
`covered_by` names it, and the check counts it toward no vocabulary ID's coverage. Every other
sample's disposition, and every gap, is therefore the same as without the list. A listed sample
that isn't redistributable, is sensitive, or has a break can't be measured, and `dispose` fails
naming it.

### Breaks

A sample that breaks a standard in [`STANDARDS.md`](../STANDARDS.md) records the break in its
entry's `breaks`, a list with one object per standard it breaks:

- `standard`: the opening of the standard's bullet, verbatim, such as `A sample's content is
  realistic`;
- `reason`: how the sample breaks it, from what its source record or the file shows.

`dispose` trims a sample with a break, whatever its other facts, so its disposition is never set
by hand. `trimmed_for` names the sample the tie-breaks prefer among those in the corpus whose
covers contain its own, and is absent when none does. Each of its covers must still be another
corpus sample's or a gap's; `dispose` stops, naming the covers, when one isn't, so a break never
loses a scenario silently. For example,
`onenote/large/Large OneDrive.one` records that its pages are printouts of generated test PDFs,
and `image/vector/script-bearing.svg` that it was written by hand.

The [`dispose`](../tools/coverage/cmd/dispose/doc.go) command of the Go module `tools/coverage`
derives every disposition by these rules, the covers of each gap a replace names, as
[Gaps](#gaps) states, and each candidate's place in the corpus, as
[Candidates in the corpus](#candidates-in-the-corpus) states. To rerun it and compare its output
with the data, run this from the repository root:

```sh
mise run dispose | cmp - docs/coverage.json
```

To re-derive the placements and the dispositions together, run `mise run place-traits > "$tmp"`,
then `mise run dispose --placement "$tmp"`, where `$tmp` names a temporary file.

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

The placements come from docproc's run over `~/docs`, made by the docproc release the corpus
targets. `.library_placement` in [`coverage.json`](coverage.json) records the run's ID, the
`docproc` release that made it, the `docproc_command` that ran it, and the `command` that placed
the IDs. The run reads `~/docs` and writes only its run directory, outside the repository. To make
it, put docproc's tools on `PATH` as [Corpus placement](#corpus-placement) states, install
`cmd/docproc` v0.2.7, the command of docproc v0.9.0, and run:

```sh
GOBIN=<dir> go install github.com/standards-lab/docproc/cmd/docproc@v0.2.7
<dir>/docproc run --out ~/runs/docs-corpus/docproc-final/docs ~/docs
```

Record the run's ID, from the summary's last line, in `.library_placement`. The placing command
is a script that runs one jq program over the run's `record.jsonl`. A sample's library IDs are
what the run observed for that file:

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

The script reads the run from `${RUNS:-$HOME/runs/docs-corpus/docproc-final/docs/runs}` and the
manifest from `${DOCS_REFERENCE:-$HOME/docs}`. It writes nothing.

## Trait and feature placement

Each entry's `traits` and `features` are the trait and feature IDs the sample exercises, placed by
reading the file and its manifest entry, never from the manifest's `covers` tags. Each value's
rule in the vocabulary is the test. The [`place-traits`](../tools/coverage/cmd/place-traits/doc.go)
command of the Go module `tools/coverage` probes every sample and applies the rules it can decide
from the file or from what docproc's kept run recorded for it:

- PDF text layers from the run's page outputs, OCR layers from invisible text over page images,
  page counts, forms, signature fields, and portfolios, read with pdfinfo, pdfimages, pdffonts, and
  qpdf, and an oversized page from a page image the run rendered below the 150 dpi that
  cmd/docproc requests;
- audio and video streams, duration, and sample rate, read with ffprobe;
- image frames, alpha, and stored orientation, read with ImageMagick's identify;
- OOXML structures read from the package's parts, with legacy and RTF documents converted to OOXML
  by LibreOffice first, as docproc converts them;
- message structure, parsed from each `.eml` and `.mbox` with Go's standard library, as the
  [`stdmail`](../tools/coverage/stdmail/doc.go) package documents, and taken from the run's
  records for `.msg` and `.pst` files;
- OneNote structures from the run's page records;
- charsets from byte order marks, UTF-8 validity, and uchardet;
- scripts by counting each script's letters in the file's text, or in the text the run extracted.
  A letter counts toward the script its Unicode character name starts with: CJK for Han, and
  Hiragana or Katakana for kana. Modifier letters, fullwidth and halfwidth forms, and letterlike
  symbols such as the Kelvin sign have names that start with no script, so they count toward none.

Where a value takes judgement, such as handwriting, a degraded or simulated scan, spontaneous or
noisy speech, or a layout in columns, the entry's `judged` object holds the ID and the evidence
for it: what was seen in the rendered file, or what its truth or source record shows. The command
adds those IDs to what it probes. To rerun it and compare its output with the data, run this from
the repository root:

```sh
mise run place-traits | cmp - <(jq '.samples | map({path, traits, features})' docs/coverage.json)
```

It reads the same run and manifest as `scripts/place-library.sh`, and writes only to a temporary
directory. With `--facts`, as in `mise run place-traits --facts`, it also prints each sample's
probe findings.

[`mise.toml`](../mise.toml) pins Go and every probe `place-traits` calls except LibreOffice's
`soffice`, the one unpinned host requirement: it is in neither mise's registry nor conda-forge, so
`place-traits` and ImageMagick's EMF and WMF delegate use the host's. `place-traits` stops at
startup, naming the tool and the fix, when a probe is not on `PATH` or is older than the version
its [adapter](../tools/coverage/probe/doc.go) in the `probe` package accepts. Each adapter accepts
nothing older than the version `mise.toml` pins, so on a host whose `PATH` puts older system tools
ahead of mise's installs, run `place-traits` with mise's tool paths first on `PATH`.

## Corpus placement

The manifest's covers are placed from a docproc run over the corpus itself: the format
directories of a working copy whose sample files are in place. `.corpus_placement` records that
run's ID, the docproc release that made it, the `docproc_command` that ran it from the repository
root, and the `command` that places the covers.

`mise run place-traits --corpus` places every sample `MANIFEST.json` lists, reading its file from
the working copy and the run from `${RUNS:-$HOME/runs/docs-corpus/docproc-final/corpus/runs}`. For
each sample, it takes the library IDs from that run by the rules of
[Library placement](#library-placement), adds the trait and feature IDs
[`place-traits`](#trait-and-feature-placement) probes and the `judged` IDs of the sample's entry in
`.samples`, and prints the sample's `covers`. A candidate has no entry in `.samples`, so its judged
IDs, each with its evidence, are in [`judged.json`](judged.json), keyed by its path. The command's
output is committed as [`corpus-placement.json`](corpus-placement.json), and the manifest takes
each candidate's covers from it.

To make the run, put docproc's tools on `PATH`, as docproc's README states for its release:
`docproc-ooxml`, `docproc-mail`, and `docproc-onenote` from their release tags, the real `7zz`
binary rather than mise's shim, and docker for `soffice`'s pinned image. Then install
`cmd/docproc` v0.2.7, the command of docproc v0.9.0, outside the repository, and run it from the
repository root with the run directory outside it:

```sh
GOBIN=<dir> go install github.com/standards-lab/docproc/cmd/docproc@v0.2.7
<dir>/docproc run archive audio email exec image office onenote pdf text video --out ~/runs/docs-corpus/docproc-final/corpus
```

Record the run's ID, from the summary's last line, in `.corpus_placement`. A run into the same
run directory replays the parts its index already holds, with outcome `seen` and the first
outcome under `recorded`; the placement reads the recorded one. To place the covers and compare
them with the committed placement and the manifest, run this from the repository root:

```sh
mise run place-traits --corpus | cmp - docs/corpus-placement.json
cmp docs/corpus-placement.json <(jq '[.files[] | select(.role == "sample") | {path, covers}] | sort_by(.path)' MANIFEST.json)
```

## Gaps

Each scenario lists the adopted samples that exercise it. A scenario that no adopted sample
exercises is a gap in `.gaps`. Each gap has an `id`, the `covers` a candidate must exercise, and
a `candidate`. A replaced sample names the gap whose covers contain its own. Replaced samples that
need the same covers, or covers inside another replaced sample's, share one gap; where several
replaced samples' covers contain a sample's, it shares the gap of the one the tie-breaks prefer.
A gap for scenarios no sample exercised groups the IDs one candidate file can exercise together.
A gap also holds the covers of a sample that breaks [`STANDARDS.md`](../STANDARDS.md) and that
no other sample covers, so that its candidate takes them over. For example, image-8's gap holds
the covers of the hand-written `image/vector/script-bearing.svg`, whose `breaks` record it, and
image-8's sourced SVG is in the corpus in its place.

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
  `generator_license`. The inputs are adopted samples or sourced candidates in the corpus, and
  a command may read its own earlier output, so nothing in a candidate is typed by hand. The
  candidate's own license follows from its inputs': an archive of adopted samples carries their
  licenses.
- **unavailable**: no file can be sourced or generated for the gap's covers. It records
  `searched`, the suites, files, and searches tried, and `reason`, why none exists, such as a
  structure only a live mail server makes. Such a gap stays open: the corpus doesn't exercise its
  scenarios until a candidate turns up.

Each sourced or generated candidate also has a `fit` line: a rating, Good or Fair, then its
quality and how it exercises the gap's covers. The fit line also shows that the candidate meets
[`STANDARDS.md`](../STANDARDS.md)'s realistic-content standard; a generated candidate's names the
sourced real text, images, speech, or scans its inputs give its content.

A candidate that breaks a standard leaves the corpus: its rows leave
[`candidates.json`](candidates.json), [`judged.json`](judged.json), and, through the tools,
[`corpus-placement.json`](corpus-placement.json) and the manifest, and its file leaves the working
copy. Its gap stays open, with an unavailable candidate whose `searched` names the candidate that
left and how it was made or fetched, and whose `reason` names the standard it broke and why, as
archive-11, email-5, and office-8 record. No replacement is searched for until a later task takes the gap up.

`dispose` writes two fields to record what the corpus placement found, as
[Candidates in the corpus](#candidates-in-the-corpus) states:

- `split_from` marks an open gap, id `<gap>-open`, and names the gap it splits from. The open gap
  holds the covers that gap's candidate doesn't confirm, and its own candidate is unavailable.
- `covered_by` names the sample whose covers contain the gap's candidate's covers. That candidate
  leaves the corpus.

A suspicious scenario, such as an archive bomb, a path traversal, an executable, or a polyglot, has
a harmless candidate: a parser fixture from an established test suite, or a file built by a
recorded command from benign inputs.

A candidate is a proposal until it is fetched or generated into the corpus and confirmed with
docproc. A gap whose `candidate` is `null` has none yet, and fails the check.

### Candidates in the corpus

`mise run fetch-candidates` fetches each sourced candidate from its pinned URL and places it at
`<area>/<source>/<name>`. The area is the gap id's prefix, the source is the GitHub repository's
name in lower case, or `wikimedia-commons`, and the name is the URL's last segment, decoded. A
trailing `.txt` is dropped from a name that already has an extension, so `jwz.mbox.txt` becomes
`jwz.mbox`. [`candidates.json`](candidates.json) records what was fetched for each gap: its
`path`, `source`, `sha256`, `bytes`, and `format`, as `file -b --mime-type` reports it, with
`format_by`, the `file` version that ran. A later fetch must match the record, so a pinned
candidate can't change. The fetch skips a gap whose candidate is `covered_by` another sample.

`mise run generate-candidates` makes each generated candidate the record doesn't list yet, once,
by running its recorded command from the repository root. It records the candidate with `source:
"generated"`, at the path the command writes. Before a command runs, each of its inputs must match
`MANIFEST.json`, and the script creates the directories of the corpus paths the command names.
The command must write exactly one new file, and if any command fails, nothing is recorded.
Nothing regenerates a recorded candidate, and its bytes are never edited, so the record and the
manifest's sha256 pin what was made. Generation is not byte-reproducible: archives and the mbox
carry the time they were made.

Run it with mise's pinned tools first on `PATH`, since the host's may be older: bsdtar from
`conda:libarchive`, ffmpeg from `conda:ffmpeg`, magick from ImageMagick, qpdf, and Go, each from
`mise where <tool>`'s `bin`. pandoc, Ghostscript, LibreOffice, iconv, and coreutils are unpinned
and run as the host installs them. The candidate's `generator` in `.gaps` records the versions
that ran, and the manifest carries it.

The manifest gives each candidate its gap's candidate's provenance, its covers from the corpus
placement, and its `gap`. A sourced candidate has no truth. A generated candidate has the truth of
the input its record's `truth_from` names, unchanged, set only where the command leaves that
input's content as it was, as audio-1's MP3 keeps the AMI headset mix's speech; otherwise it has
none. Its license is its inputs' distinct licenses, joined by `; `. Its license is verified only
when every input's is, and it is redistributable only when every input is.

`dispose` reads the record and the placement, and settles each placed candidate's place:

- **Open gap**: the covers of a gap that its candidate's placed covers lack become its open gap,
  and the gap keeps the covers they confirm. `dispose` writes the open gap with an unavailable
  candidate. The record of what was searched and why none is found is written once, as every
  candidate's research is, and `dispose` never changes it. `dispose` removes the open gap once a
  placement confirms its covers.
- **Trim**: an adopted sample whose covers lie inside a candidate's is trimmed, and
  `trimmed_for` names the sample the tie-breaks prefer among those that stay, candidate or not.
  A sample or sourced candidate a generated candidate lists in its `inputs` stays, as
  [`STANDARDS.md`](../STANDARDS.md) allows.
- **Covered**: a candidate whose covers lie inside another sample's leaves the corpus, and its
  gap's `covered_by` names that sample.

A candidate the placement doesn't list, as one that left the corpus, keeps its gap's record as it
stands, but for a `covered_by` that names a sample no longer in the corpus: `dispose` clears it,
so `mise run fetch-candidates` fetches the candidate again and the next corpus run places it.

To count the dispositions and gaps, run:

```sh
jq '(.samples | group_by(.disposition) | map({(.[0].disposition): length}) | add) + {gaps: (.gaps | length)}' docs/coverage.json
```

## Measurements

[`measurements.json`](measurements.json) is the measurement list: the targeted acceptance
measurements in docproc and docinfer that rerun over corpus inputs. It is written by hand. A measurement input is a `~/docs` sample, or a truth file, that a listed measurement reads.
As [`STANDARDS.md`](../STANDARDS.md) allows, it stays in the corpus while a measurement lists it,
even when another sample covers its scenarios. The list is the only way a sample enters the corpus
for a measurement: `dispose` gives a listed sample it would otherwise trim the `measure`
disposition.

The list is a JSON array, `[]` when it is empty, with one object per measurement:

- `id`: the measurement's own id.
- `repository`: `docproc` or `docinfer`, the repository whose acceptance it reruns.
- `acceptance`: the heading of the section it reruns in that repository's `docs/acceptance.md`,
  verbatim, such as `Word error rate`.
- `summary`: one line saying what it measures.
- `needs_truth`: `true` when its inputs need truth, else `false`.
- `inputs`: the corpus paths it reads, each a sample's `path` in [`coverage.json`](coverage.json)
  or a truth file a sample there lists in its `truth`.

An entry names its inputs by path and holds no other facts about them: the manifest and
`coverage.json` hold those. `scripts/check.sh` reads `docs/measurements.json`, or the copy the
`MEASUREMENTS` environment variable names, after the coverage data passes. It fails when:

- the file is missing, isn't JSON, or isn't an array, or an entry isn't an object;
- an entry lacks its id, or two entries share an id;
- an entry has a field other than these six;
- an entry lacks its repository or its acceptance section, its summary, or a boolean
  `needs_truth`;
- an entry lists no inputs, or an input that isn't a path;
- an input is neither a sample nor a truth file in `coverage.json`;
- an input isn't redistributable, is sensitive, or carries a break. A truth file takes these facts
  from the samples that list it: it isn't redistributable when none of them is, and it is
  sensitive when one of them is;
- a measurement that needs truth lists a sample without truth;
- a sample is measured but no measurement lists it, or a measurement lists a sample that is
  trimmed though it could be measured.

Each failure message names its fix.

## The manifest

[`MANIFEST.json`](../MANIFEST.json) lists each adopted or measured sample and each candidate in
the corpus, with its `covers`: its `library`, `traits`, and `features` IDs together, sorted. It
also lists the truth and doc files the samples link, as [README.md](../README.md#the-manifest)
describes. Each sample or truth file a measurement reads names, in `measures`, the ids of the
measurements in [`measurements.json`](measurements.json) that list it. `mise run manifest` builds
it from this data and the measurement list, so the inventory and the scenario index agree.
`scripts/check.sh` checks the manifest before the data. It fails when:

- a listed file is missing, or its size or sha256 differs from its entry;
- a file under a format directory isn't listed.

After the data and the measurement list pass, it checks the manifest against them and git. It
fails when:

- the manifest's `version` isn't 1, or an entry's role isn't sample, truth, or doc;
- an adopted or measured sample lacks `license_url` or `license_read`, the reading of its license
  at its source that [`docs/license-readings.json`](license-readings.json) records;
- an entry's `covers` ID isn't in the vocabulary;
- git tracks a sample file, or doesn't track a truth or doc file;
- a manifest sample that isn't a candidate is neither adopted nor measured in this data, or an
  adopted or measured sample here isn't a sample in the manifest;
- an adopted or measured sample's covers differ from its covers in this data;
- two samples that aren't measured share a format and covers;
- an entry's `measures` isn't a non-empty list of ids in the measurement list, or differs from the
  sorted ids of the measurements that list its path;
- a file a measurement lists has no `measures`, or isn't in the manifest;
- a sample that isn't measured has covers that lie inside another such sample's, whatever the
  format, unless a generated candidate lists it in its `inputs`, as
  [`STANDARDS.md`](../STANDARDS.md) allows;
- a candidate's gap has no sourced or generated candidate, or is `covered_by` another sample;
- a candidate's provenance differs from its gap's candidate's;
- a sourced candidate has truth, or a generated one has truth that isn't its inputs';
- a generated candidate's license, license status, or redistributability differs from what its
  inputs give;
- a candidate's covers lack some of its gap's covers;
- a sourced or generated gap that no sample covers has no candidate in the manifest;
- `corpus-placement.json`'s rows differ from the manifest's samples' paths and covers;
- a path in `judged.json` isn't a sample, or an ID there isn't in the vocabulary;
- with `~/docs` present, the manifest differs from what `mise run manifest` prints.

The payload must be in place: without it, every sample is a missing file. Each failure message
names its fix.

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
- a break's `standard` isn't the opening of a [`STANDARDS.md`](../STANDARDS.md) bullet, or it
  has no `reason`;
- a sample with a break isn't trimmed, its trim leaves out a sample in the corpus whose covers
  contain its own, or one of its covers is no other adopted sample's, candidate's, or gap's;
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
- an adopted or measured sample isn't redistributable or is sensitive;
- two adopted samples share a format and a covers set, or one adopted sample's covers lie inside
  another's or equal another's that wins the tie-breaks. The candidates in the corpus count as
  adopted samples here, a generated candidate's inputs are exempt, and measured samples are
  skipped;
- a gap's `covered_by` names neither an adopted sample nor a candidate in the corpus, or one whose
  covers lack the gap's;
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
