package media

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
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

// PDF places the traits and features of the PDF at path, from pdfinfo, pdfimages, pdffonts, and
// qpdf, and from outputs, the output events docproc's run recorded for it. It returns its facts:
// pages, rendered and text pages, DPIs, large images, fonts, and invisible text runs.
func (pr Prober) PDF(ctx context.Context, path string, outputs []run.Event) (traits, features idset.Set, facts *jsonvalue.Object) {
	need("PDF", "PDFInfo", pr.PDFInfo)
	need("PDF", "PDFImages", pr.PDFImages)
	need("PDF", "PDFFonts", pr.PDFFonts)
	need("PDF", "QPDF", pr.QPDF)
	facts = jsonvalue.NewObject()
	traits, features = idset.Set{}, idset.Set{}
	info := string(pr.PDFInfo.Output(ctx, path))
	pages := 0
	if m := pagesLine.FindStringSubmatch(info); m != nil {
		pages, _ = strconv.Atoi(m[1])
	}
	textPages := idset.Set{}
	imagePages := 0
	dpis := idset.Set{}
	var dpiValues []any
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
				if v, err := jsonvalue.Parse([]byte(raw)); err == nil && jsonvalue.Truthy(v) {
					dpiValues = append(dpiValues, v)
				}
			}
		}
	}
	rendered := imagePages
	if rendered == 0 {
		rendered = pages
	}
	slices.SortFunc(dpiValues, jsonvalue.Compare)
	facts.Set("pages", pages)
	facts.Set("rendered_pages", imagePages)
	facts.Set("text_pages", len(textPages))
	facts.Set("dpis", dpiValues)
	if rendered > 0 {
		if len(textPages) == 0 {
			traits.Add("trait:text-layer.none")
		} else if len(textPages) < rendered {
			traits.Add("trait:text-layer.partial")
		}
	}
	if len(dpis) > 1 {
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
	facts.Set("pages_with_large_images", len(largeImagePages))
	var ppis []int
	for p := range ppiSet {
		ppis = append(ppis, p)
	}
	slices.Sort(ppis)
	if len(ppis) > 3 {
		ppis = ppis[:3]
	}
	facts.Set("large_image_ppi", ppis)
	fonts := text.SplitLines(string(pr.PDFFonts.Output(ctx, path)))
	if len(fonts) > 2 {
		fonts = fonts[2:]
	} else {
		fonts = nil
	}
	facts.Set("fonts", len(fonts))
	withoutUnicode := 0
	for _, l := range fonts {
		if fontNoUnicode.MatchString(l) {
			withoutUnicode++
		}
	}
	facts.Set("fonts_without_unicode", withoutUnicode)
	raw := pr.QPDF.Output(ctx, path)
	if len(raw) == 0 {
		raw, _ = os.ReadFile(path)
	}
	invisible := len(invisibleText.FindAllIndex(raw, -1))
	facts.Set("invisible_text_runs", invisible)
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
	Disposition map[string]any  `json:"disposition"`
}

func (s ffStream) attachedPicture() bool {
	return jsonvalue.Truthy(s.Disposition["attached_pic"])
}

// AudioVideo places the traits and features of the audio or video file at path, from ffprobe's
// streams and format, and returns its facts: duration, streams, sample rates, and channels.
func (pr Prober) AudioVideo(ctx context.Context, path string) (traits, features idset.Set, facts *jsonvalue.Object) {
	need("AudioVideo", "FFprobe", pr.FFprobe)
	data := pr.FFprobe.Output(ctx, path)
	if len(data) == 0 {
		data = []byte("{}")
	}
	var probe struct {
		Streams []ffStream `json:"streams"`
		Format  struct {
			Duration any `json:"duration"`
		} `json:"format"`
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&probe); err != nil {
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
	if jsonvalue.Truthy(probe.Format.Duration) {
		switch d := probe.Format.Duration.(type) {
		case string:
			duration, _ = strconv.ParseFloat(strings.TrimSpace(d), 64)
		case json.Number:
			duration, _ = d.Float64()
		}
	}
	rateSet := map[int]bool{}
	for _, s := range audio {
		rateSet[jsonInt(s.SampleRate)] = true
	}
	var rates []int
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
	facts = jsonvalue.NewObject()
	facts.Set("duration", jsonvalue.Float(rounded))
	facts.Set("audio_streams", len(audio))
	facts.Set("video_streams", len(video))
	facts.Set("attached_pictures", len(pics))
	facts.Set("sample_rates", rates)
	facts.Set("channels", channels)
	facts.Set("subtitle_streams", subtitles)
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

// Image places the features of the image at path, from magick identify's frames, and returns its
// facts: frames, and the first frame's format, size, alpha, orientation, and opacity.
func (pr Prober) Image(ctx context.Context, path string) (traits, features idset.Set, facts *jsonvalue.Object) {
	need("Image", "Magick", pr.Magick)
	out := string(pr.Magick.Output(ctx, path))
	var frames [][]string
	for _, l := range text.SplitLines(out) {
		if strings.Count(l, "|") == 5 {
			frames = append(frames, strings.Split(l, "|"))
		}
	}
	facts = jsonvalue.NewObject()
	facts.Set("frames", len(frames))
	traits, features = idset.Set{}, idset.Set{}
	if len(frames) == 0 {
		return traits, features, facts
	}
	f := frames[0]
	format, alpha, orient, opaque := f[0], f[3], f[4], f[5]
	width, _ := strconv.Atoi(strings.TrimSpace(f[1]))
	height, _ := strconv.Atoi(strings.TrimSpace(f[2]))
	facts.Set("format", format)
	facts.Set("width", width)
	facts.Set("height", height)
	facts.Set("alpha", alpha)
	facts.Set("orientation", orient)
	facts.Set("opaque", opaque)
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
