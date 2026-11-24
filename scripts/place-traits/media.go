package main

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
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

func pyFields(line string) []string {
	return strings.FieldsFunc(line, isPySpace)
}

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

func pyLines(b []byte) []string {
	return pySplitLines(string(b))
}

func pdf(path, rel string, rec *runRecord) (set, set, *object) {
	facts := newObject()
	traits, features := set{}, set{}
	info := string(run([]string{"pdfinfo", path}, nil, probeTimeout))
	pages := 0
	if m := pagesLine.FindStringSubmatch(info); m != nil {
		pages, _ = strconv.Atoi(m[1])
	}
	textPages := set{}
	imagePages := 0
	dpis := set{}
	var dpiValues []any
	for _, e := range rec.outputs(rel) {
		if e.Part.Depth != nil && *e.Part.Depth != 0 {
			continue
		}
		out := e.Output
		if out.Kind == "text" && strings.HasPrefix(out.Name, "page ") {
			textPages.add(metaRaw(out.Meta, "page"))
		}
		if out.Kind == "image" && strings.HasPrefix(out.Name, "page ") {
			imagePages++
			raw := metaRaw(out.Meta, "dpi")
			if !dpis[raw] {
				dpis.add(raw)
				if v, err := parseJSON([]byte(raw)); err == nil && truthy(v) {
					dpiValues = append(dpiValues, v)
				}
			}
		}
	}
	rendered := imagePages
	if rendered == 0 {
		rendered = pages
	}
	slices.SortFunc(dpiValues, comparePy)
	dpiList := append([]any{}, dpiValues...)
	facts.set("pages", pages)
	facts.set("rendered_pages", imagePages)
	facts.set("text_pages", len(textPages))
	facts.set("dpis", dpiList)
	if rendered > 0 {
		if len(textPages) == 0 {
			traits.add("trait:text-layer.none")
		} else if len(textPages) < rendered {
			traits.add("trait:text-layer.partial")
		}
	}
	if len(dpis) > 1 {
		traits.add("trait:scale.oversized-page")
	}
	if max(pages, rendered) >= 100 {
		traits.add("trait:scale.many-pages")
	}
	images := pyLines(run([]string{"pdfimages", "-list", path}, nil, probeTimeout))
	if len(images) > 2 {
		images = images[2:]
	} else {
		images = nil
	}
	largeImagePages := set{}
	ppiSet := map[int]bool{}
	for _, l := range images {
		f := pyFields(l)
		if len(f) > 0 && isDigits(f[0]) {
			if w, _ := strconv.Atoi(f[3]); w >= 500 {
				page, _ := strconv.Atoi(f[0])
				largeImagePages.add(strconv.Itoa(page))
			}
		}
		if len(f) > 12 && isDigits(f[12]) {
			if w, _ := strconv.Atoi(f[3]); w >= 500 {
				ppi, _ := strconv.Atoi(f[12])
				ppiSet[ppi] = true
			}
		}
	}
	facts.set("pages_with_large_images", len(largeImagePages))
	var ppis []int
	for p := range ppiSet {
		ppis = append(ppis, p)
	}
	slices.Sort(ppis)
	if len(ppis) > 3 {
		ppis = ppis[:3]
	}
	if ppis == nil {
		ppis = []int{}
	}
	facts.set("large_image_ppi", ppis)
	fonts := pyLines(run([]string{"pdffonts", path}, nil, probeTimeout))
	if len(fonts) > 2 {
		fonts = fonts[2:]
	} else {
		fonts = nil
	}
	facts.set("fonts", len(fonts))
	withoutUnicode := 0
	for _, l := range fonts {
		if fontNoUnicode.MatchString(l) {
			withoutUnicode++
		}
	}
	facts.set("fonts_without_unicode", withoutUnicode)
	raw := run([]string{"qpdf", "--qdf", "--object-streams=disable", "--decrypt", path, "-"}, nil, probeTimeout)
	if len(raw) == 0 {
		raw, _ = os.ReadFile(path)
	}
	invisible := len(invisibleText.FindAllIndex(raw, -1))
	facts.set("invisible_text_runs", invisible)
	if len(textPages) > 0 && len(largeImagePages) > 0 && invisible > 0 && len(largeImagePages) >= max(1, len(textPages)/2) {
		traits.add("trait:text-layer.ocr")
	}
	if acroForm.Match(raw) && formField.Match(raw) {
		features.add("feature:forms")
	}
	if signatureField.Match(raw) {
		features.add("feature:signature-fields")
	}
	if portfolio.Match(raw) {
		features.add("feature:portfolio")
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
	return truthy(s.Disposition["attached_pic"])
}

func audioVideo(path string) (set, set, *object) {
	data := run([]string{"ffprobe", "-v", "error", "-show_streams", "-show_format", "-of", "json", path}, nil, probeTimeout)
	if len(data) == 0 {
		data = []byte("{}")
	}
	var probe struct {
		Streams []ffStream `json:"streams"`
		Format  struct {
			Duration any `json:"duration"`
		} `json:"format"`
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
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
	if truthy(probe.Format.Duration) {
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
	if rates == nil {
		rates = []int{}
	}
	rounded, _ := strconv.ParseFloat(strconv.FormatFloat(duration, 'f', 1, 64), 64)
	facts := newObject()
	facts.set("duration", pyFloat(rounded))
	facts.set("audio_streams", len(audio))
	facts.set("video_streams", len(video))
	facts.set("attached_pictures", len(pics))
	facts.set("sample_rates", rates)
	facts.set("channels", channels)
	facts.set("subtitle_streams", subtitles)
	traits, features := set{}, set{}
	if len(video) > 0 && len(audio) == 0 {
		traits.add("trait:audio.none")
	}
	if len(audio) > 0 && slices.Max(rates) <= 8000 {
		traits.add("trait:audio.narrowband")
	}
	if len(audio) > 0 && duration > 30 {
		traits.add("trait:audio.multi-window")
	}
	if duration >= 1800 {
		traits.add("trait:scale.long-media")
	}
	if len(pics) > 0 {
		features.add("feature:attached-picture")
	}
	if len(audio) > 1 || len(video) > 1 {
		features.add("feature:multiple-streams")
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

func image(path string) (set, set, *object) {
	out := string(run([]string{"magick", "identify", "-format", "%m|%w|%h|%A|%[orientation]|%[opaque]\n", path}, nil, probeTimeout))
	var frames [][]string
	for _, l := range pySplitLines(out) {
		if strings.Count(l, "|") == 5 {
			frames = append(frames, strings.Split(l, "|"))
		}
	}
	facts := newObject()
	facts.set("frames", len(frames))
	traits, features := set{}, set{}
	if len(frames) == 0 {
		return traits, features, facts
	}
	f := frames[0]
	format, alpha, orient, opaque := f[0], f[3], f[4], f[5]
	width, _ := strconv.Atoi(strings.TrimSpace(f[1]))
	height, _ := strconv.Atoi(strings.TrimSpace(f[2]))
	facts.set("format", format)
	facts.set("width", width)
	facts.set("height", height)
	facts.set("alpha", alpha)
	facts.set("orientation", orient)
	facts.set("opaque", opaque)
	if len(frames) > 1 && format == "TIFF" {
		features.add("feature:multi-page-image")
	}
	if len(frames) > 1 && (format == "GIF" || format == "WEBP" || format == "AVIF") {
		features.add("feature:animation")
	}
	if orient != "" && orient != "Undefined" && orient != "TopLeft" {
		features.add("feature:orientation")
	}
	if alpha != "Undefined" && alpha != "False" && alpha != "" && strings.ToLower(opaque) == "false" {
		features.add("feature:transparency")
	}
	return traits, features, facts
}

func archiveEntries(path string) []string {
	var entries []string
	for _, l := range pySplitLines(string(run([]string{"bsdtar", "-tf", path}, nil, probeTimeout))) {
		if l != "" && !strings.HasSuffix(l, "/") {
			entries = append(entries, l)
		}
	}
	return entries
}

func comparePy(a, b any) int {
	if sa, ok := a.(string); ok {
		return strings.Compare(sa, str(b))
	}
	fa, _ := a.(json.Number).Float64()
	fb, _ := b.(json.Number).Float64()
	switch {
	case fa < fb:
		return -1
	case fa > fb:
		return 1
	}
	return 0
}
