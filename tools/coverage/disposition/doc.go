// Package disposition derives every sample's disposition in docs/coverage.json from its
// placement and the facts the data records, by the rules docs/coverage.md states: adopt, trim
// with trimmed_for, or replace with gap, and the covers of each gap a replace names. Gap
// candidates are researched data, so it keeps them as they are and adds no gap.
//
// Its exported identifiers:
//
//   - [Dispose] derives the dispositions in the parsed coverage data.
//   - [Place] sets the data's traits and features from place-traits' output.
//   - [Run] reads docs/coverage.json and writes it disposed, from command-line arguments.
package disposition
