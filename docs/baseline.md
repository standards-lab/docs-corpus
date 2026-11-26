# Baseline

The reference that capstone.data-intelligence measures against: the runs made over each corpus
tag's files, one section per tag. A tag's files are those its `MANIFEST.json` lists, and
[coverage.md](coverage.md) records the scenarios they cover. Each tag's runs are kept outside the
repository, in the run directory `~/runs/docs-corpus/baseline/<tag>`. Every figure below comes from
the run's own records: its manifest, its record, and the summary docproc printed.

## v0.2.0

The run directory is `~/runs/docs-corpus/baseline/v0.2.0`. It holds docproc's run directory,
`docproc/`, the summary docproc printed, `docproc-run.stdout`, its empty `docproc-run.stderr`, and
the binaries that ran, `bin/`.

### docproc

| | |
| --- | --- |
| Run | `01a218c9-8847-7ec5-b98e-cf5803c6a989` |
| Record | `~/runs/docs-corpus/baseline/v0.2.0/docproc/runs/01a218c9-8847-7ec5-b98e-cf5803c6a989/` |
| docproc | v0.9.0 (`46aebf1a0e61c2dd8ce054ec12b558d8fda3477a`) |
| cmd/docproc | v0.2.7 |
| Started | 2026-11-26T16:56:30.535926687Z |
| Ended | 2026-11-26T16:58:33.998798715Z |
| Wall time | 123.5 s |
| Files collected | 299 |
| Workers | 22 |

The run started with its run directory empty, so it replayed nothing from an earlier run. It
exited 0 with nothing on stderr; the summary's last line reads `failed parts 5  error inputs 0`.
[`coverage.json`](coverage.json)'s `.corpus_placement` cites this run, and
`mise run place-traits --corpus` reproduces [`corpus-placement.json`](corpus-placement.json) from
it, as [coverage.md](coverage.md#corpus-placement) states.

**Files.** The 299 files are every file `MANIFEST.json` lists at v0.2.0: its samples and the truth
and doc files they link. Before the run, each file in the checkout matched its manifest entry's
sha256 and size, and the format directories held no other file. `MANIFEST.json` on this branch
differs from the tag's only in one sample's `url`.

**Machine.** The laptop:

| | |
| --- | --- |
| Hostname | `terminal` |
| CPU | Intel Core Ultra 9 185H, 16 cores, 22 threads |
| RAM | 33,058,959,360 bytes |
| OS | Omarchy, an Arch Linux derivative, kernel 7.2.15-1-omarchy |
| Docker | 29.9.2 |

**Configuration.** No `DOCPROC_` variable was set, so every bound was its starting value, and the
workers were the default, the host's 22 CPUs. cmd/docproc was installed with `go install` from the
module proxy. docproc-ooxml 0.2.0 and docproc-mail 0.2.0 came from their `tools/ooxml/v0.2.0` and
`tools/mail/v0.2.0` releases and docproc-onenote 0.1.2 from its `tools/onenote/v0.1.2` release,
each the linux-x64 archive, matched against the release's `SHA256SUMS`. The four binaries in
`bin/`, then the real `7zz`, were first on `PATH`, and the command ran from the repository root:

```sh
GOBIN=~/runs/docs-corpus/baseline/v0.2.0/bin go install github.com/standards-lab/docproc/cmd/docproc@v0.2.7
docproc run archive audio email exec image office onenote pdf text video --out ~/runs/docs-corpus/baseline/v0.2.0/docproc
```

The stage table wired all ten routes. The tools the run's manifest records:

| Tool | Version | From |
| --- | --- | --- |
| gs | 10.08.0 | `/usr/bin/gs` |
| qpdf | 12.4.2 | `/usr/bin/qpdf` |
| pdfdetach, pdftotext | 26.10.0 | `/usr/bin` |
| magick | 7.1.2-34 Q16-HDRI | `/usr/bin/magick` |
| soffice | LibreOffice 26.8.2.2 | `gotenberg/gotenberg:9.10.1-libreoffice@sha256:79aec465636fef089f6ed3da69b75c7d3c7155fcecd20b3d80563868c54b5a90`, in its container |
| docproc-ooxml | 0.2.0 | `bin/` |
| docproc-mail | 0.2.0 | `bin/` |
| docproc-onenote | 0.1.2 | `bin/` |
| 7zz | 26.06 | mise, `github:ip7z/7zip` |
| ffmpeg, ffprobe | n9.0.2 | `/usr/bin` |

**Outcomes.** Each file's outcome, by format area, the top directory of its path. A `seen` file's
content was already recorded in this run, from another file or from a part inside one.

| Area | prepared | seen | encrypted | no processor yet | failed | quarantined | Total |
| --- | --- | --- | --- | --- | --- | --- | --- |
| archive | 6 | 0 | 1 | 0 | 0 | 3 | 10 |
| audio | 51 | 0 | 0 | 0 | 0 | 0 | 51 |
| email | 34 | 2 | 0 | 1 | 0 | 0 | 37 |
| exec | 1 | 0 | 0 | 3 | 0 | 0 | 4 |
| image | 10 | 0 | 0 | 0 | 2 | 0 | 12 |
| office | 53 | 3 | 1 | 0 | 1 | 0 | 58 |
| onenote | 13 | 0 | 1 | 0 | 0 | 0 | 14 |
| pdf | 76 | 0 | 3 | 0 | 1 | 0 | 80 |
| text | 8 | 2 | 0 | 1 | 0 | 0 | 11 |
| video | 22 | 0 | 0 | 0 | 0 | 0 | 22 |
| **Total** | 274 | 7 | 6 | 5 | 4 | 3 | 299 |
