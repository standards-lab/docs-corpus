// Package probe runs the system tools place-traits reads samples with, one adapter per tool. Each
// adapter's constructor finds its program on PATH and runs its version command, so a missing or
// outdated tool fails at startup: the refusal names the tool and the fix, mise install for a tool
// mise.toml pins and LibreOffice on the host for soffice, and wraps ErrMissing, ErrTooOld, or
// ErrNoVersion. A version is dotted numbers, a hyphen read as a dot, compared as numbers with a
// missing part counting as zero, and the check is a minimum, the version mise.toml pins.
//
// An adapter runs its program as a process under the caller's context, so cancelling the context
// stops it, within a bound of its own: 300 s for a conversion and 600 s for any other run. A run
// returns the program's standard output whatever its exit code, and nil when the program could
// not start, the context ended, or the bound passed; its standard error is discarded.
//
// Its exported identifiers:
//
//   - [PDFInfo], [PDFImages], and [PDFFonts] are poppler's tools, built by [NewPDFInfo],
//     [NewPDFImages], and [NewPDFFonts], each against [PopplerMinimum].
//   - [QPDF] is qpdf, built by [NewQPDF] against [QPDFMinimum].
//   - [FFprobe] is ffprobe, built by [NewFFprobe] against [FFmpegMinimum].
//   - [Magick] is ImageMagick's magick, built by [NewMagick] against [ImageMagickMinimum].
//   - [Bsdtar] is libarchive's bsdtar, built by [NewBsdtar] against [LibarchiveMinimum].
//   - [Uchardet] is uchardet, built by [NewUchardet] against [UchardetMinimum].
//   - [Soffice] is LibreOffice's soffice, built by [NewSoffice] against [LibreOfficeMinimum].
//   - Each path adapter's Output runs its tool over one file; [Uchardet.Guess] reads bytes on
//     standard input, and [Soffice.Convert] converts a document into a directory.
//   - [Tools] holds every adapter, and [New] builds them all, joining every refusal.
//   - [ErrMissing], [ErrTooOld], and [ErrNoVersion] are a refusal's causes.
package probe
