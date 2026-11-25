// Package stdmail parses messages for mail's placement rules with Go's standard library: net/mail
// for the RFC 5322 header, mime and mime/multipart for RFC 2045 and 2046 media types, parameters,
// and parts, mime/quotedprintable and encoding/base64 for transfer encodings, mime's WordDecoder
// for RFC 2047 encoded words, and golang.org/x/text for the charsets the standard library lacks.
//
// It is one implementation of mail's Parser, and mail never imports it.
//
// Its exported identifiers:
//
//   - [Parser] parses a message into its parts, [Parser.Parse], and decodes a header's encoded
//     words, [Parser.DecodeHeader].
package stdmail
