// Command dispose prints docs/coverage.json with every sample's disposition derived from its
// placement, and the covers of each gap a replace names. Run it from the repository root:
//
//	go run ./tools/coverage/cmd/dispose [--placement <file>]
//
// With --placement, the traits and features come from that place-traits output first. It names
// on standard error each gap that may be dropped, and writes nothing else.
//
// See docs/coverage.md
package main
