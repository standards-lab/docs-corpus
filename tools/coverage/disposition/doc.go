// Package disposition derives every sample's disposition in docs/coverage.json from its
// placement and the facts the data records, by the rules docs/coverage.md states.
//
// Its exported identifiers:
//
//   - [Coverage] is the coverage data, with its [Vocabulary] of [Library] and [Rule] entries,
//     [LibraryPlacement], [CorpusPlacement], [Sample]s, [Gap]s and their [Candidate]s, and
//     [DocWithoutSample]s; [FedBy] is a capability's routes.
//   - [Dispose] derives the dispositions in the coverage data, with the gap candidates the
//     corpus placement [Placed], and splits each gap a candidate's covers don't confirm.
//   - [Placement] is one element of place-traits' output, and [Place] sets the data's traits and
//     features from it.
package disposition
