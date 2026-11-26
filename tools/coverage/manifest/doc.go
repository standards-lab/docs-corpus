// Package manifest builds MANIFEST.json, the corpus inventory, from docs/coverage.json's adopted
// and measured samples, the reference docs' manifest, the license readings in
// docs/license-readings.json, the gap candidates fetched into docs/candidates.json with their
// covers from docs/corpus-placement.json, the measurements in docs/measurements.json, which name
// the files they read, and the sourced inputs in docs/sourced.json, and copies the files it lists
// into place.
//
// Its exported identifiers:
//
//   - [Manifest] is the inventory, schema [Version], of [File] entries, a sample's with its
//     [Truth] files, and each file a measurement reads with the measurements' ids.
//   - [Reference] is the reference docs' manifest, of [ReferenceFile] entries.
//   - [Reading] is a sample's, truth file's, or doc file's license, read at its source.
//   - [Candidate] is a gap candidate's fetched file, as docs/candidates.json records it, and
//     [Placed] one sample's covers, as place-traits --corpus prints them.
//   - [Sourced] is a measurement input sourced outside the reference docs, with its
//     [SourcedTruth] files, as docs/sourced.json records it.
//   - [Build] builds the inventory.
//   - [Copy] copies its files from the reference docs and checks each copy.
//
// See README.md
// See docs/coverage.md
package manifest
