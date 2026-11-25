// Package app is tools/coverage's composition root: it builds each command from the module's
// packages and runs it over the command line's arguments, so each cmd/* entrypoint is one call.
//
// Its exported identifiers:
//
//   - [PlaceTraits] runs place-traits.
//   - [Dispose] runs dispose.
//   - [Generate] runs generate.
//
// See architecture/standards/go-elemental/principles/topology-and-naming.md
package app
