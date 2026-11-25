// Package office places the features of office documents: an OOXML package's from its parts, a
// legacy or RTF document's once LibreOffice converts it to OOXML, an OpenDocument file's text,
// and a OneNote file's from the records docproc's run kept for it.
//
// Its exported identifiers:
//
//   - [OOXML] places an OOXML package's features, and [OOXMLFacts] is what it found.
//   - [IsOOXML] and [IsLegacy] classify a file extension.
//   - [Converter] converts a document with LibreOffice, as probe's soffice adapter does, and
//     [Conversions] converts legacy documents through it, [Conversions.Convert] one at a time.
//   - [ODFText] returns an OpenDocument file's text.
//   - [OneNote] places a OneNote file's traits and features, and [OneNoteFacts] is what it found.
package office
