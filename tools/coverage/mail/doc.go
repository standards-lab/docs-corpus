// Package mail places the traits and features of mail samples: an .eml message and an mbox,
// read through a Parser, the mail docproc's run recorded for a .msg or .pst file, and the body
// properties an Outlook .msg names.
//
// Its exported identifiers:
//
//   - [Parser] parses a message into its [Part]s and decodes a header's encoded words.
//   - [Message] places one message's traits and features and returns its text and its
//     [MessageFacts].
//   - [Mbox] places an mbox's, across its messages, and returns its [MboxFacts].
//   - [Records] places those of the mail a run recorded, and returns its [RecordsFacts].
//   - [MSGBodies] lists the body properties a .msg names.
package mail
