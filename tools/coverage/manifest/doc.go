// Package manifest builds MANIFEST.json, the corpus inventory, from docs/coverage.json's adopted
// samples, the reference docs' manifest, and the license readings in docs/license-readings.json,
// and copies the files it lists into place.
//
// Its exported identifiers:
//
//   - [Manifest] is the inventory, schema [Version], of [File] entries, a sample's with its
//     [Truth] files.
//   - [Reference] is the reference docs' manifest, of [ReferenceFile] entries.
//   - [Reading] is a sample's, truth file's, or doc file's license, read at its source.
//   - [Build] builds the inventory.
//   - [Copy] copies its files from the reference docs and checks each copy.
//
// See README.md
// See docs/coverage.md
package manifest
