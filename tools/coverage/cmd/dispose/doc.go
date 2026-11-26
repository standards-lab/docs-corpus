// Command dispose prints docs/coverage.json with every sample's disposition derived from its
// placement, and the covers of each gap a replace names. Run it from the repository root:
//
//	go run ./tools/coverage/cmd/dispose [--placement <file>]
//
// With --placement, the traits and features come from that place-traits output first. The gap
// candidates docs/candidates.json records and docs/corpus-placement.json places join the adopted
// samples: it trims an adopted sample whose covers lie inside a candidate's, records a candidate
// inside another sample's as its gap's covered_by, clears a covered_by that names a sample no
// longer in the corpus, and splits off a gap's covers its candidate doesn't confirm as an open
// gap. A sample whose breaks record a standard it breaks is trimmed, and a sample a measurement in
// docs/measurements.json lists is disposed measure where it would be trimmed. It names on standard
// error each gap that may be dropped, and writes nothing else.
//
// See docs/coverage.md
package main
