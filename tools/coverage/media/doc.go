// Package media places the traits and features of PDFs, audio and video, images, and archives
// from what the system tools print about them, and, for a PDF, from the page outputs docproc's
// run recorded. A Prober holds the tools as Tool interfaces, which probe's adapters satisfy.
//
// Its exported identifiers:
//
//   - [Tool] runs one system tool over a file.
//   - [Prober] holds the tools; [Prober.PDF], [Prober.AudioVideo], and [Prober.Image] place a
//     file's traits and features and return what they found as [PDFFacts], [AudioVideoFacts],
//     and [ImageFacts], and [Prober.ArchiveEntries] lists an archive's files.
//   - [RequestDPI] is the DPI docproc's run requests a PDF's pages at.
package media
