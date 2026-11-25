package media

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
	"github.com/standards-lab/docs-corpus/tools/coverage/text"
)

var (
	pagesLine      = regexp.MustCompile(`(?m)^Pages:\s+(\d+)`)
	fontNoUnicode  = regexp.MustCompile(`[\t\n\v\f\r ](yes|no)[\t\n\v\f\r ]+(yes|no)[\t\n\v\f\r ]+no[\t\n\v\f\r ]+\d+[\t\n\v\f\r ]+\d+[\t\n\v\f\r ]*$`)
	invisibleText  = regexp.MustCompile(`(?m)(^|[^0-9.])3[\t\n\v\f\r ]+Tr\b`)
	acroForm       = regexp.MustCompile(`/AcroForm[\t\n\v\f\r ]*(<<|\d+ 0 R)`)
	formField      = regexp.MustCompile(`/FT[\t\n\v\f\r ]*/(Tx|Btn|Ch)`)
	signatureField = regexp.MustCompile(`/FT[\t\n\v\f\r ]*/Sig`)
	portfolio      = regexp.MustCompile(`/Collection[\t\n\v\f\r ]*(<<|\d+ 0 R)`)
)

// RequestDPI is the DPI docproc's run requests a PDF's pages at, cmd/docproc's document request;
// a page rendered below it passed the request's MaxPixels.
const RequestDPI = 150

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// PDFFacts is what PDF found in a PDF.
type PDFFacts struct {
	// Pages is the page count pdfinfo prints.
	Pages int
	// RenderedPages is the number of page images the run rendered.
	RenderedPages int
	// TextPages is the number of pages the run extracted text from.
	TextPages int
	// DPIs are the DPIs the run's page images name in their meta, sorted, never nil.
	DPIs []string
	// PagesWithLargeImages is the number of pages with an image at least 500 pixels wide.
	PagesWithLargeImages int
	// LargeImagePPI are the lowest three resolutions of those images, ascending, never nil.
	LargeImagePPI []int
	// Fonts is the number of fonts pdffonts lists.
	Fonts int
	// FontsWithoutUnicode is the number of those fonts with no Unicode map.
	FontsWithoutUnicode int
	// InvisibleTextRuns is the number of text runs in render mode 3, invisible.
	InvisibleTextRuns int
}

// PDF places the traits and features of the PDF at path, from pdfinfo, pdfimages, pdffonts, and
// qpdf, and from outputs, the output events docproc's run recorded for it. It returns its facts.
func (pr Prober) PDF(ctx context.Context, path string, outputs []run.Event) (traits, features idset.Set, facts PDFFacts) {
	need("PDF", "PDFInfo", pr.PDFInfo)
	need("PDF", "PDFImages", pr.PDFImages)
	need("PDF", "PDFFonts", pr.PDFFonts)
	need("PDF", "QPDF", pr.QPDF)
	traits, features = idset.Set{}, idset.Set{}
	info := string(pr.PDFInfo.Output(ctx, path))
	pages := 0
	if m := pagesLine.FindStringSubmatch(info); m != nil {
		pages, _ = strconv.Atoi(m[1])
	}
	textPages := idset.Set{}
	imagePages := 0
	dpis := idset.Set{}
	dpiValues := []string{}
	for _, e := range outputs {
		if e.Part.Depth != nil && *e.Part.Depth != 0 {
			continue
		}
		out := e.Output
		if out.Kind == "text" && strings.HasPrefix(out.Name, "page ") {
			textPages.Add(out.MetaRaw("page"))
		}
		if out.Kind == "image" && strings.HasPrefix(out.Name, "page ") {
			imagePages++
			raw := out.MetaRaw("dpi")
			if !dpis[raw] {
				dpis.Add(raw)
				var dpi string
				if json.Unmarshal([]byte(raw), &dpi) == nil && dpi != "" {
					dpiValues = append(dpiValues, dpi)
				}
			}
		}
	}
	rendered := imagePages
	if rendered == 0 {
		rendered = pages
	}
	slices.Sort(dpiValues)
	facts.Pages = pages
	facts.RenderedPages = imagePages
	facts.TextPages = len(textPages)
	facts.DPIs = dpiValues
	if rendered > 0 {
		if len(textPages) == 0 {
			traits.Add("trait:text-layer.none")
		} else if len(textPages) < rendered {
			traits.Add("trait:text-layer.partial")
		}
	}
	if slices.ContainsFunc(dpiValues, func(dpi string) bool {
		n, err := strconv.ParseFloat(dpi, 64)
		return err == nil && n < RequestDPI
	}) {
		traits.Add("trait:scale.oversized-page")
	}
	if max(pages, rendered) >= 100 {
		traits.Add("trait:scale.many-pages")
	}
	images := text.SplitLines(string(pr.PDFImages.Output(ctx, path)))
	if len(images) > 2 {
		images = images[2:]
	} else {
		images = nil
	}
	largeImagePages := idset.Set{}
	ppiSet := map[int]bool{}
	for _, l := range images {
		f := text.Fields(l)
		if len(f) > 0 && isDigits(f[0]) {
			if w, _ := strconv.Atoi(f[3]); w >= 500 {
				page, _ := strconv.Atoi(f[0])
				largeImagePages.Add(strconv.Itoa(page))
			}
		}
		if len(f) > 12 && isDigits(f[12]) {
			if w, _ := strconv.Atoi(f[3]); w >= 500 {
				ppi, _ := strconv.Atoi(f[12])
				ppiSet[ppi] = true
			}
		}
	}
	facts.PagesWithLargeImages = len(largeImagePages)
	ppis := []int{}
	for p := range ppiSet {
		ppis = append(ppis, p)
	}
	slices.Sort(ppis)
	if len(ppis) > 3 {
		ppis = ppis[:3]
	}
	facts.LargeImagePPI = ppis
	fonts := text.SplitLines(string(pr.PDFFonts.Output(ctx, path)))
	if len(fonts) > 2 {
		fonts = fonts[2:]
	} else {
		fonts = nil
	}
	facts.Fonts = len(fonts)
	withoutUnicode := 0
	for _, l := range fonts {
		if fontNoUnicode.MatchString(l) {
			withoutUnicode++
		}
	}
	facts.FontsWithoutUnicode = withoutUnicode
	raw := pr.QPDF.Output(ctx, path)
	if len(raw) == 0 {
		raw, _ = os.ReadFile(path)
	}
	invisible := len(invisibleText.FindAllIndex(raw, -1))
	facts.InvisibleTextRuns = invisible
	if len(textPages) > 0 && len(largeImagePages) > 0 && invisible > 0 && len(largeImagePages) >= max(1, len(textPages)/2) {
		traits.Add("trait:text-layer.ocr")
	}
	if acroForm.Match(raw) && formField.Match(raw) {
		features.Add("feature:forms")
	}
	if signatureField.Match(raw) {
		features.Add("feature:signature-fields")
	}
	if portfolio.Match(raw) {
		features.Add("feature:portfolio")
	}
	return traits, features, facts
}

type ffStream struct {
	CodecType   string          `json:"codec_type"`
	SampleRate  json.RawMessage `json:"sample_rate"`
	Channels    json.RawMessage `json:"channels"`
	Disposition struct {
		AttachedPic int `json:"attached_pic"`
	} `json:"disposition"`
}

func (s ffStream) attachedPicture() bool {
	return s.Disposition.AttachedPic != 0
}

// AudioVideoFacts is what AudioVideo found in an audio or video file.
type AudioVideoFacts struct {
	// Duration is its duration in seconds, rounded to a tenth.
	Duration float64
	// AudioStreams is the number of its audio streams.
	AudioStreams int
	// VideoStreams is the number of its video streams that are not attached pictures.
	VideoStreams int
	// AttachedPictures is the number of its attached pictures, such as cover art.
	AttachedPictures int
	// SampleRates are its audio streams' sample rates, ascending, never nil.
	SampleRates []int
	// Channels are its audio streams' nonzero channel counts, ascending, never nil.
	Channels []int
	// SubtitleStreams is the number of its subtitle streams.
	SubtitleStreams int
}

// AudioVideo places the traits and features of the audio or video file at path, from ffprobe's
// streams and format, and returns its facts.
func (pr Prober) AudioVideo(ctx context.Context, path string) (traits, features idset.Set, facts AudioVideoFacts) {
	need("AudioVideo", "FFprobe", pr.FFprobe)
	data := pr.FFprobe.Output(ctx, path)
	if len(data) == 0 {
		data = []byte("{}")
	}
	var probe struct {
		Streams []ffStream `json:"streams"`
		Format  struct {
			Duration json.RawMessage `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		probe.Streams = nil
		probe.Format.Duration = nil
	}
	var audio, video, pics []ffStream
	subtitles := 0
	for _, s := range probe.Streams {
		if s.CodecType == "audio" {
			audio = append(audio, s)
		}
		if s.CodecType == "video" && !s.attachedPicture() {
			video = append(video, s)
		}
		if s.attachedPicture() {
			pics = append(pics, s)
		}
		if s.CodecType == "subtitle" {
			subtitles++
		}
	}
	duration := 0.0
	var durationText string
	if json.Unmarshal(probe.Format.Duration, &durationText) == nil {
		duration, _ = strconv.ParseFloat(strings.TrimSpace(durationText), 64)
	} else {
		_ = json.Unmarshal(probe.Format.Duration, &duration)
	}
	rateSet := map[int]bool{}
	for _, s := range audio {
		rateSet[jsonInt(s.SampleRate)] = true
	}
	rates := []int{}
	for r := range rateSet {
		rates = append(rates, r)
	}
	slices.Sort(rates)
	channelSet := map[int]bool{}
	for _, s := range audio {
		if c := jsonInt(s.Channels); c != 0 {
			channelSet[c] = true
		}
	}
	channels := []int{}
	for c := range channelSet {
		channels = append(channels, c)
	}
	slices.Sort(channels)
	rounded, _ := strconv.ParseFloat(strconv.FormatFloat(duration, 'f', 1, 64), 64)
	facts = AudioVideoFacts{
		Duration:         rounded,
		AudioStreams:     len(audio),
		VideoStreams:     len(video),
		AttachedPictures: len(pics),
		SampleRates:      rates,
		Channels:         channels,
		SubtitleStreams:  subtitles,
	}
	traits, features = idset.Set{}, idset.Set{}
	if len(video) > 0 && len(audio) == 0 {
		traits.Add("trait:audio.none")
	}
	if len(audio) > 0 && slices.Max(rates) <= 8000 {
		traits.Add("trait:audio.narrowband")
	}
	if len(audio) > 0 && duration > 30 {
		traits.Add("trait:audio.multi-window")
	}
	if duration >= 1800 {
		traits.Add("trait:scale.long-media")
	}
	if len(pics) > 0 {
		features.Add("feature:attached-picture")
	}
	if len(audio) > 1 || len(video) > 1 {
		features.Add("feature:multiple-streams")
	}
	return traits, features, facts
}

func jsonInt(raw json.RawMessage) int {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0
	}
	switch t := v.(type) {
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(t))
		return n
	case float64:
		return int(t)
	}
	return 0
}

// ImageFacts is what Image found in an image: its frames, and its first frame's properties as
// magick identify prints them, all zero when it has no frame.
type ImageFacts struct {
	// Frames is the number of its frames.
	Frames int
	// Format is the first frame's format, such as PNG.
	Format string
	// Width is the first frame's width in pixels.
	Width int
	// Height is the first frame's height in pixels.
	Height int
	// Alpha is the first frame's alpha channel, such as Blend or Undefined.
	Alpha string
	// Orientation is the first frame's orientation, such as TopLeft.
	Orientation string
	// Opaque is whether the first frame is opaque, True or False.
	Opaque string
}

// Image places the features of the image at path, from magick identify's frames, and returns its
// facts.
func (pr Prober) Image(ctx context.Context, path string) (traits, features idset.Set, facts ImageFacts) {
	need("Image", "Magick", pr.Magick)
	out := string(pr.Magick.Output(ctx, path))
	var frames [][]string
	for _, l := range text.SplitLines(out) {
		if strings.Count(l, "|") == 5 {
			frames = append(frames, strings.Split(l, "|"))
		}
	}
	facts.Frames = len(frames)
	traits, features = idset.Set{}, idset.Set{}
	if len(frames) == 0 {
		return traits, features, facts
	}
	f := frames[0]
	format, alpha, orient, opaque := f[0], f[3], f[4], f[5]
	width, _ := strconv.Atoi(strings.TrimSpace(f[1]))
	height, _ := strconv.Atoi(strings.TrimSpace(f[2]))
	facts.Format, facts.Width, facts.Height = format, width, height
	facts.Alpha, facts.Orientation, facts.Opaque = alpha, orient, opaque
	if len(frames) > 1 && format == "TIFF" {
		features.Add("feature:multi-page-image")
	}
	if len(frames) > 1 && (format == "GIF" || format == "WEBP" || format == "AVIF") {
		features.Add("feature:animation")
	}
	if orient != "" && orient != "Undefined" && orient != "TopLeft" {
		features.Add("feature:orientation")
	}
	if alpha != "Undefined" && alpha != "False" && alpha != "" && strings.ToLower(opaque) == "false" {
		features.Add("feature:transparency")
	}
	return traits, features, facts
}

// ArchiveEntries returns the file entries bsdtar lists in the archive at path, directories left
// out.
func (pr Prober) ArchiveEntries(ctx context.Context, path string) []string {
	need("ArchiveEntries", "Bsdtar", pr.Bsdtar)
	var entries []string
	for _, l := range text.SplitLines(string(pr.Bsdtar.Output(ctx, path))) {
		if l != "" && !strings.HasSuffix(l, "/") {
			entries = append(entries, l)
		}
	}
	return entries
}
