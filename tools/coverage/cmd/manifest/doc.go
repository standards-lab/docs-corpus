// Command manifest prints MANIFEST.json: one entry per adopted or measured sample in
// docs/coverage.json and per truth and doc file it links, with the provenance of its reference row
// in docs/reference-rows.json, each sample's covers, and the license readings in
// docs/license-readings.json, and one per gap candidate docs/candidates.json records, with its
// gap's provenance and its covers from docs/corpus-placement.json, and one per sourced input and
// truth file docs/sourced.json records, a sourced input with its covers from
// docs/corpus-placement.json. Each entry a measurement in docs/measurements.json reads names the
// measurements that read it. Run it from the repository root:
//
//	go run ./tools/coverage/cmd/manifest > MANIFEST.json
//
// It reads nothing outside the repository.
//
// See README.md
package main
