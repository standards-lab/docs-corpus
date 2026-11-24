// Command generate writes the corpus's generated candidates from adopted samples. Run it from
// the repository root:
//
//	go run ./tools/coverage/cmd/generate rights-managed <message.eml> <payload> <out.eml>
//	go run ./tools/coverage/cmd/generate mbox <out.mbox> <message.eml> [<message.eml> ...]
//
// rights-managed adds a Content-Class: rpmsg.message header to a multipart/mixed message and
// attaches the payload as message.rpmsg. mbox writes 1,000 messages, the given ones in turn.
//
// See docs/coverage.md
package main
