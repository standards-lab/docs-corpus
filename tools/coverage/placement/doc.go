// Package placement places every sample the repository's MANIFEST.json lists on the library,
// trait, and feature axes, as docs/coverage.md states: it takes the library IDs docproc's run over
// the corpus observed for each file, probes it by its extension and header with media, office,
// mail, and text, adds what that run recorded for it and the IDs judged for it, and writes each
// sample's covers as JSON.
//
// Its exported identifiers:
//
//   - [Placer] holds what placement reads and the probes it runs, and [Placer.Run] writes the
//     placement.
package placement
