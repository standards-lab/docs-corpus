// Command manifest prints MANIFEST.json: one entry per adopted sample in docs/coverage.json and
// per truth and doc file it links, with the reference docs' provenance, each sample's covers, and
// the license readings in docs/license-readings.json. Run it from the repository root:
//
//	go run ./tools/coverage/cmd/manifest [--copy] > MANIFEST.json
//
// It reads the reference docs from $DOCS_REFERENCE, default ~/docs. With --copy, it first copies
// every file it lists from there to its path in the repository, checking each copy's size and
// sha256.
//
// See README.md
package main
