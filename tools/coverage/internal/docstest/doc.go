// Package docstest gives the unit tier the corpus's real data read in place, never copied into
// the module: the samples under the reference docs directory, $DOCS_REFERENCE or ~/docs, and the
// repository's docs/coverage.json. Both are outside the module, so a test that needs one skips
// when it is absent, naming what is missing; where it is present, a file a test names that isn't
// there fails the test.
//
// Its exported identifiers:
//
//   - [Dir] returns the reference docs directory.
//   - [Path] returns a sample's path under it.
//   - [Coverage] returns the coverage data's bytes.
//   - [Samples] returns the coverage data's samples; [Sample] is one of them, and
//     [Sample.Probed] its feature IDs less the judged ones.
package docstest
