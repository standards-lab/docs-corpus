package media

import "context"

// Tool runs one system tool over a file, as probe's adapters do.
type Tool interface {
	// Output returns what the tool prints for the file at path, or nil when it could not run.
	Output(ctx context.Context, path string) []byte
}

// Prober probes PDFs, audio and video, images, and archives with the system tools. A method
// panics, naming the field, when a tool it runs is nil.
type Prober struct {
	// PDFInfo is poppler's pdfinfo.
	PDFInfo Tool
	// PDFImages is poppler's pdfimages, listing a PDF's images.
	PDFImages Tool
	// PDFFonts is poppler's pdffonts.
	PDFFonts Tool
	// QPDF is qpdf, writing a PDF decrypted and uncompressed.
	QPDF Tool
	// FFprobe is FFmpeg's ffprobe, printing streams and format as JSON.
	FFprobe Tool
	// Magick is ImageMagick's magick identify, printing one line per frame.
	Magick Tool
	// Bsdtar is libarchive's bsdtar, listing an archive's entries.
	Bsdtar Tool
}

func need(method, field string, t Tool) {
	if t == nil {
		panic("media: Prober." + method + " requires a " + field)
	}
}
