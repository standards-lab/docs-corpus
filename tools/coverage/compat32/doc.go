// Package compat32 parses RFC 5322 messages as CPython's email package parses them under its
// compat32 policy: the feed parser's headers, multipart boundaries, and defects tolerated as
// CPython tolerates them, the parameter and RFC 2231 handling of email.message, RFC 2047 header
// decoding, and the binascii decoders for quoted-printable, base64, and uuencode. It is ported
// from CPython and carries the PSF license's notice in the file that holds the port.
//
// It is one implementation of mail's Parser, and mail never imports it, so another parser can
// replace it without a change to the placement rules.
//
// Its exported identifiers:
//
//   - [Parser] parses a message into its parts, [Parser.Parse], and decodes a header's encoded
//     words, [Parser.DecodeHeader].
package compat32
