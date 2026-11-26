// Package docstest gives the unit tier the corpus's real data read in place, never copied into
// the module: the repository's docs/coverage.json and MANIFEST.json, and the samples fetched into
// the checkout at their MANIFEST.json paths. All are outside the module, so a test that needs one
// skips when it is absent, naming what is missing and, for a sample, the fetch; a file a test
// names that MANIFEST.json doesn't list fails the test.
//
// Its exported identifiers:
//
//   - [Path] returns a sample's path in the checkout.
//   - [Root] returns the repository's root, and [Coverage] the coverage data's bytes.
//   - [Samples] returns the coverage data's samples MANIFEST.json lists; [Sample] is one of them,
//     and [Sample.Probed] its feature IDs less the judged ones.
package docstest
