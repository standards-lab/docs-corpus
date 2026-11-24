# Coverage

Coverage is the record of which library scenarios the corpus exercises and which sample exercises
each one. It decides which samples the corpus keeps and names a candidate for every scenario no
sample covers. It has two parts:

- [`coverage.json`](coverage.json): the data. It holds the scenario vocabulary, each sample's
  disposition, and the gaps. `scripts/check.sh` reads it.
- This page, which explains the data and how the check uses it.

## The vocabulary

The vocabulary in [`coverage.json`](coverage.json), `.vocabulary`, defines every scenario ID once,
on three axes:

- **Library** (`.vocabulary.library`): an identifier that docproc or docinfer defines. These are
  docproc's file kinds, routes, and inspection checks, and docinfer's capabilities. Each entry
  names the identifier and links to its definition, pinned to a commit.
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
one disposition in `.samples`:

- **adopt**: the sample joins the corpus with its covers set, the IDs it exercises.
- **trim**: an adopted sample already covers this sample's scenarios. The entry names that sample.
- **replace**: the sample can't be adopted, and a gap entry takes over its scenarios. The entry
  names that gap. A sample that isn't redistributable is replaced. It is trimmed only when an
  adopted sample already covers its scenarios.

A truth file or doc file follows its sample's disposition.

No two adopted samples share a format and a covers set. When samples collide, the one that stays
is chosen by these tie-breaks, in order:

1. It has truth.
2. It has a verified license.
3. It is redistributable.
4. It is the smallest.

The library placements come from docproc's kept run over `~/docs`. The data records the run ID
and the jq command that produced them.

## Gaps

Each scenario lists the adopted samples that exercise it. A scenario that no adopted sample
exercises is a gap in `.gaps`. Each gap names one candidate:

- **sourced**: the candidate's URL, pinned to a commit where one exists, and its license, read
  at the source and recorded as verified.
- **generated**: the command that makes the candidate, the generator, and the generator's license.

Each candidate also gets one line on its quality and fit. None is hand-built.

## The check

`scripts/check.sh` validates the data whenever `docs/coverage.json` exists. The `COVERAGE`
environment variable points it at another copy. It fails when:

- a vocabulary entry doesn't have its axis's ID shape;
- a library entry lacks its identifier or its pinned link;
- a trait or feature entry lacks its rule;
- a `covers` ID anywhere in the file isn't in the vocabulary.

To fix an unknown ID, use an ID the vocabulary defines. If the scenario is new, define it in the
vocabulary, with its link or rule.
