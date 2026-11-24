// Package text reads text as the Python scripts the coverage tools replaced read it: Python's
// codecs and their aliases over golang.org/x/text, its whitespace and line splitting, a text
// file's charset from its byte order mark, UTF-8 validity, UTF-16 without a BOM, or a Guesser,
// and the scripts a text's letters are written in. It carries the module's one sourced
// dependency, golang.org/x/text.
//
// Its exported identifiers:
//
//   - [DecodeOrUTF8] decodes bytes with a Python codec name, or keeps them as they are.
//   - [Guesser] guesses a charset, as uchardet does.
//   - [Charset] places a text file's trait:charset ID and names its charset.
//   - [DecodeText] decodes a text file by its BOM or Charset's guess.
//   - [UTF8Name] reports whether a declared charset names UTF-8.
//   - [Scripts] counts a text's letters by script and places its trait:script IDs;
//     [ScriptCount] is one script's count.
//   - [IsSpace], [Strip], [Fields], and [SplitLines] are Python's whitespace and line rules.
package text
