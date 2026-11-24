// Package generate makes the corpus's generated gap candidates from adopted samples, as
// docs/coverage.json's generated candidates record them.
//
// Its exported identifiers:
//
//   - [RightsManaged] makes a rights-managed message: an rpmsg.message Content-Class header and a
//     message.rpmsg attachment added to a multipart/mixed message.
//   - [Mbox] makes an mbox of 1,000 messages.
//   - [Run] makes one candidate from command-line arguments, reading and writing its files.
package generate
