// Package placement places every ~/docs sample the coverage data lists on the trait and feature
// axes, as docs/coverage.md states: it probes each file by its extension and header with media,
// office, mail, and text, adds what docproc's run recorded for it and the IDs the coverage data
// judged for it, and writes the placement as JSON. In its corpus mode it places every sample the
// repository's MANIFEST.json lists instead, from docproc's run over the corpus, with the library
// IDs that run observed, and writes each sample's covers.
//
// Its exported identifiers:
//
//   - [Placer] holds what placement reads and the probes it runs, and [Placer.Run] writes the
//     placement.
package placement
