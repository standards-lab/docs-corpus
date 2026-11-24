package office

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
)

var legacyTarget = map[string]string{".doc": "docx", ".rtf": "docx", ".wpd": "docx", ".xls": "xlsx", ".ppt": "pptx"}

var ooxmlKinds = idset.New("docx", "xlsx", "pptx", "docm", "dotm", "xlsm", "xltm", "pptm", "dotx", "xltx", "potx", "ppsx")

// IsOOXML reports whether ext, a lower-cased file extension with its dot, is an OOXML package's.
func IsOOXML(ext string) bool {
	return ooxmlKinds[strings.TrimPrefix(ext, ".")]
}

// IsLegacy reports whether ext, a lower-cased file extension with its dot, is a legacy or RTF
// document's that LibreOffice converts to OOXML: .doc, .rtf, .wpd, .xls, or .ppt.
func IsLegacy(ext string) bool {
	return legacyTarget[ext] != ""
}

// Converter converts a document with LibreOffice, as soffice --convert-to does.
type Converter interface {
	// Convert converts the document at path to format, such as docx, writing the result into
	// outDir, with LibreOffice's user profile in the directory profile.
	Convert(ctx context.Context, path, format, outDir, profile string)
}

// Conversions converts legacy documents to OOXML, as docproc converts them, each into a
// directory of its own under Dir, numbered in order, with one LibreOffice profile in Dir/profile.
type Conversions struct {
	// Soffice converts each document. It is required.
	Soffice Converter
	// Dir is the directory the conversions are written in. It is required.
	Dir string
	n   int
}

// Convert converts the legacy document at path to OOXML and returns the converted file's path, or
// "" when path is not a legacy document or the conversion wrote nothing.
func (c *Conversions) Convert(ctx context.Context, path string) string {
	if c.Soffice == nil {
		panic("office: Conversions.Convert requires a Soffice")
	}
	if c.Dir == "" {
		panic("office: Conversions.Convert requires a Dir")
	}
	c.n++
	target, ok := legacyTarget[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return ""
	}
	out := filepath.Join(c.Dir, strconv.Itoa(c.n))
	if err := os.MkdirAll(out, 0o755); err != nil {
		return ""
	}
	c.Soffice.Convert(ctx, path, target, out, c.Dir+"/profile")
	entries, err := os.ReadDir(out)
	if err != nil || len(entries) == 0 {
		return ""
	}
	return filepath.Join(out, entries[0].Name())
}

// ODFText returns the text of the OpenDocument file at path: its content.xml with each element
// replaced by a space, or "" when it has none.
func ODFText(path string) string {
	for _, f := range zipTexts(path) {
		if f.name == "content.xml" {
			return markup.ReplaceAllString(string(f.data), " ")
		}
	}
	return ""
}
