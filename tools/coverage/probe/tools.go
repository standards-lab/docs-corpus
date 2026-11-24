package probe

import (
	"context"
	"errors"
	"regexp"
)

// PopplerMinimum is the oldest poppler, for pdfinfo, pdfimages, and pdffonts, the adapters
// accept: the version mise.toml pins.
const PopplerMinimum = "26.11.0"

// QPDFMinimum is the oldest qpdf the adapter accepts: the version mise.toml pins.
const QPDFMinimum = "12.4.2"

// FFmpegMinimum is the oldest ffprobe the adapter accepts: the version mise.toml pins.
const FFmpegMinimum = "9.1.0"

// ImageMagickMinimum is the oldest magick the adapter accepts: the version mise.toml pins, its
// patch level after the hyphen.
const ImageMagickMinimum = "7.1.2-36"

// UchardetMinimum is the oldest uchardet the adapter accepts: the version mise.toml pins.
const UchardetMinimum = "0.0.8"

// LibarchiveMinimum is the oldest bsdtar the adapter accepts: the version mise.toml pins.
const LibarchiveMinimum = "3.8.11"

// LibreOfficeMinimum is the oldest soffice the adapter accepts: 24.2, the oldest LibreOffice a
// current long-term distribution ships. mise.toml can't pin it, so it is a host requirement.
const LibreOfficeMinimum = "24.2"

const libreOfficeFix = "install LibreOffice " + LibreOfficeMinimum + " or later on the host"

func popplerSpec(name string) spec {
	return spec{
		name:    name,
		args:    []string{"-v"},
		pattern: regexp.MustCompile(name + ` version (\d+(?:\.\d+)*)`),
		minimum: PopplerMinimum,
		fix:     miseFix,
	}
}

var (
	qpdfSpec = spec{
		name:    "qpdf",
		args:    []string{"--version"},
		pattern: regexp.MustCompile(`qpdf version (\d+(?:\.\d+)*)`),
		minimum: QPDFMinimum,
		fix:     miseFix,
	}
	ffprobeSpec = spec{
		name:    "ffprobe",
		args:    []string{"-version"},
		pattern: regexp.MustCompile(`ffprobe version n?(\d+(?:\.\d+)*)`),
		minimum: FFmpegMinimum,
		fix:     miseFix,
	}
	magickSpec = spec{
		name:    "magick",
		args:    []string{"-version"},
		pattern: regexp.MustCompile(`Version: ImageMagick (\d+(?:\.\d+)*(?:-\d+)?)`),
		minimum: ImageMagickMinimum,
		fix:     miseFix,
	}
	uchardetSpec = spec{
		name:    "uchardet",
		args:    []string{"--version"},
		pattern: regexp.MustCompile(`Version (\d+(?:\.\d+)*)`),
		minimum: UchardetMinimum,
		fix:     miseFix,
	}
	bsdtarSpec = spec{
		name:    "bsdtar",
		args:    []string{"--version"},
		pattern: regexp.MustCompile(`bsdtar (\d+(?:\.\d+)*)`),
		minimum: LibarchiveMinimum,
		fix:     miseFix,
	}
	sofficeSpec = spec{
		name:    "soffice",
		args:    []string{"--version"},
		pattern: regexp.MustCompile(`LibreOffice (\d+(?:\.\d+)*)`),
		minimum: LibreOfficeMinimum,
		fix:     libreOfficeFix,
	}
)

// PDFInfo is poppler's pdfinfo.
type PDFInfo struct{ p program }

// NewPDFInfo finds pdfinfo and checks it is PopplerMinimum or newer.
func NewPDFInfo(ctx context.Context) (*PDFInfo, error) {
	p, err := find(ctx, popplerSpec("pdfinfo"))
	if err != nil {
		return nil, err
	}
	return &PDFInfo{p}, nil
}

// Output returns what pdfinfo prints for the PDF at path, or nil when it could not run.
func (t *PDFInfo) Output(ctx context.Context, path string) []byte {
	return t.p.output(ctx, []string{path}, nil, probeTimeout)
}

// PDFImages is poppler's pdfimages.
type PDFImages struct{ p program }

// NewPDFImages finds pdfimages and checks it is PopplerMinimum or newer.
func NewPDFImages(ctx context.Context) (*PDFImages, error) {
	p, err := find(ctx, popplerSpec("pdfimages"))
	if err != nil {
		return nil, err
	}
	return &PDFImages{p}, nil
}

// Output returns the image list pdfimages -list prints for the PDF at path, or nil when it could
// not run.
func (t *PDFImages) Output(ctx context.Context, path string) []byte {
	return t.p.output(ctx, []string{"-list", path}, nil, probeTimeout)
}

// PDFFonts is poppler's pdffonts.
type PDFFonts struct{ p program }

// NewPDFFonts finds pdffonts and checks it is PopplerMinimum or newer.
func NewPDFFonts(ctx context.Context) (*PDFFonts, error) {
	p, err := find(ctx, popplerSpec("pdffonts"))
	if err != nil {
		return nil, err
	}
	return &PDFFonts{p}, nil
}

// Output returns the font table pdffonts prints for the PDF at path, or nil when it could not
// run.
func (t *PDFFonts) Output(ctx context.Context, path string) []byte {
	return t.p.output(ctx, []string{path}, nil, probeTimeout)
}

// QPDF is qpdf.
type QPDF struct{ p program }

// NewQPDF finds qpdf and checks it is QPDFMinimum or newer.
func NewQPDF(ctx context.Context) (*QPDF, error) {
	p, err := find(ctx, qpdfSpec)
	if err != nil {
		return nil, err
	}
	return &QPDF{p}, nil
}

// Output returns the PDF at path decrypted, with its object streams written out uncompressed in
// qpdf's QDF form, or nil when qpdf could not run.
func (t *QPDF) Output(ctx context.Context, path string) []byte {
	return t.p.output(ctx, []string{"--qdf", "--object-streams=disable", "--decrypt", path, "-"}, nil, probeTimeout)
}

// FFprobe is FFmpeg's ffprobe.
type FFprobe struct{ p program }

// NewFFprobe finds ffprobe and checks it is FFmpegMinimum or newer.
func NewFFprobe(ctx context.Context) (*FFprobe, error) {
	p, err := find(ctx, ffprobeSpec)
	if err != nil {
		return nil, err
	}
	return &FFprobe{p}, nil
}

// Output returns the streams and format of the media file at path as ffprobe's JSON, or nil when
// it could not run.
func (t *FFprobe) Output(ctx context.Context, path string) []byte {
	return t.p.output(ctx, []string{"-v", "error", "-show_streams", "-show_format", "-of", "json", path}, nil, probeTimeout)
}

// Magick is ImageMagick's magick.
type Magick struct{ p program }

// NewMagick finds magick and checks it is ImageMagickMinimum or newer.
func NewMagick(ctx context.Context) (*Magick, error) {
	p, err := find(ctx, magickSpec)
	if err != nil {
		return nil, err
	}
	return &Magick{p}, nil
}

// Output returns one line per frame of the image at path, its format, width, height, alpha,
// orientation, and opacity joined by |, or nil when magick identify could not run.
func (t *Magick) Output(ctx context.Context, path string) []byte {
	return t.p.output(ctx, []string{"identify", "-format", "%m|%w|%h|%A|%[orientation]|%[opaque]\n", path}, nil, probeTimeout)
}

// Bsdtar is libarchive's bsdtar.
type Bsdtar struct{ p program }

// NewBsdtar finds bsdtar and checks it is LibarchiveMinimum or newer.
func NewBsdtar(ctx context.Context) (*Bsdtar, error) {
	p, err := find(ctx, bsdtarSpec)
	if err != nil {
		return nil, err
	}
	return &Bsdtar{p}, nil
}

// Output returns the entry list bsdtar -tf prints for the archive at path, or nil when it could
// not run.
func (t *Bsdtar) Output(ctx context.Context, path string) []byte {
	return t.p.output(ctx, []string{"-tf", path}, nil, probeTimeout)
}

// Uchardet is uchardet.
type Uchardet struct{ p program }

// NewUchardet finds uchardet and checks it is UchardetMinimum or newer.
func NewUchardet(ctx context.Context) (*Uchardet, error) {
	p, err := find(ctx, uchardetSpec)
	if err != nil {
		return nil, err
	}
	return &Uchardet{p}, nil
}

// Guess returns the charset uchardet names for data, on a line, or nil when it could not run.
func (t *Uchardet) Guess(ctx context.Context, data []byte) []byte {
	return t.p.output(ctx, nil, data, probeTimeout)
}

// Soffice is LibreOffice's soffice.
type Soffice struct{ p program }

// NewSoffice finds soffice and checks it is LibreOfficeMinimum or newer.
func NewSoffice(ctx context.Context) (*Soffice, error) {
	p, err := find(ctx, sofficeSpec)
	if err != nil {
		return nil, err
	}
	return &Soffice{p}, nil
}

// Convert converts the document at path to format, such as docx, writing the result into outDir,
// headless, with its user profile in the directory profile. A conversion that fails writes
// nothing; the caller reads outDir.
func (t *Soffice) Convert(ctx context.Context, path, format, outDir, profile string) {
	t.p.output(ctx, []string{"-env:UserInstallation=file://" + profile, "--headless", "--convert-to", format, "--outdir", outDir, path}, nil, convertTimeout)
}

// Tools is every adapter place-traits runs.
type Tools struct {
	PDFInfo   *PDFInfo
	PDFImages *PDFImages
	PDFFonts  *PDFFonts
	QPDF      *QPDF
	FFprobe   *FFprobe
	Magick    *Magick
	Bsdtar    *Bsdtar
	Uchardet  *Uchardet
	Soffice   *Soffice
}

// New builds every adapter, and returns every refusal joined, each naming its tool and the fix.
func New(ctx context.Context) (*Tools, error) {
	var errs []error
	keep := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	var t Tools
	var err error
	t.PDFInfo, err = NewPDFInfo(ctx)
	keep(err)
	t.PDFImages, err = NewPDFImages(ctx)
	keep(err)
	t.PDFFonts, err = NewPDFFonts(ctx)
	keep(err)
	t.QPDF, err = NewQPDF(ctx)
	keep(err)
	t.FFprobe, err = NewFFprobe(ctx)
	keep(err)
	t.Magick, err = NewMagick(ctx)
	keep(err)
	t.Bsdtar, err = NewBsdtar(ctx)
	keep(err)
	t.Uchardet, err = NewUchardet(ctx)
	keep(err)
	t.Soffice, err = NewSoffice(ctx)
	keep(err)
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return &t, nil
}
