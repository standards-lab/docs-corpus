// Package disposition derives every sample's disposition in docs/coverage.json from its
// placement and the facts the data records, by the rules docs/coverage.md states.
//
// Its exported identifiers:
//
//   - [Dispose] derives the dispositions in the parsed coverage data.
//   - [Place] sets the data's traits and features from place-traits' output.
package disposition
