// Package disposition derives every sample's disposition in docs/coverage.json from its
// placement and the facts the data records, by the rules docs/coverage.md states.
//
// Its exported identifiers:
//
//   - [Coverage] is the coverage data, with its [Vocabulary] of [Library] and [Rule] entries,
//     [LibraryPlacement], [CorpusPlacement], [Sample]s and the [Break]s of a standard each
//     records, [Gap]s and their [Candidate]s, and [DocWithoutSample]s; [FedBy] is a capability's
//     routes.
//   - [Measurement] is one entry of docs/measurements.json, the measurement list.
//   - [Dispose] derives the dispositions in the coverage data, with the gap candidates the
//     corpus placement places, as [Placed] values, trims each sample that records a [Break],
//     gives the measure disposition to each sample a [Measurement] lists that it would trim, and
//     splits each gap a candidate's covers don't confirm.
package disposition
