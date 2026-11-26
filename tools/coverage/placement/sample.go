package placement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
	"github.com/standards-lab/docs-corpus/tools/coverage/media"
	"github.com/standards-lab/docs-corpus/tools/coverage/office"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
	"github.com/standards-lab/docs-corpus/tools/coverage/text"
)

const mib = 1024 * 1024

type sample struct {
	Path    string                     `json:"path"`
	Library []string                   `json:"library"`
	Judged  map[string]json.RawMessage `json:"judged"`
}

type manifestEntry struct {
	Format string `json:"format"`
}

type facts struct {
	Pages                *int           `json:"pages,omitempty"`
	RenderedPages        *int           `json:"rendered_pages,omitempty"`
	TextPages            *int           `json:"text_pages,omitempty"`
	DPIs                 []string       `json:"dpis,omitzero"`
	PagesWithLargeImages *int           `json:"pages_with_large_images,omitempty"`
	LargeImagePPI        []int          `json:"large_image_ppi,omitzero"`
	Fonts                *int           `json:"fonts,omitempty"`
	FontsWithoutUnicode  *int           `json:"fonts_without_unicode,omitempty"`
	InvisibleTextRuns    *int           `json:"invisible_text_runs,omitempty"`
	Duration             *float64       `json:"duration,omitempty"`
	AudioStreams         *int           `json:"audio_streams,omitempty"`
	VideoStreams         *int           `json:"video_streams,omitempty"`
	AttachedPictures     *int           `json:"attached_pictures,omitempty"`
	SampleRates          []int          `json:"sample_rates,omitzero"`
	Channels             []int          `json:"channels,omitzero"`
	SubtitleStreams      *int           `json:"subtitle_streams,omitempty"`
	Frames               *int           `json:"frames,omitempty"`
	Format               *string        `json:"format,omitempty"`
	Width                *int           `json:"width,omitempty"`
	Height               *int           `json:"height,omitempty"`
	Alpha                *string        `json:"alpha,omitempty"`
	Orientation          *string        `json:"orientation,omitempty"`
	Opaque               *string        `json:"opaque,omitempty"`
	Parts                *int           `json:"parts,omitempty"`
	Messages             *int           `json:"messages,omitempty"`
	FoldersWithMessages  *int           `json:"folders_with_messages,omitempty"`
	BodyProperties       []string       `json:"body_properties,omitzero"`
	Sections             *int           `json:"sections,omitempty"`
	Entries              *int           `json:"entries,omitempty"`
	Tables               *int           `json:"tables,omitempty"`
	Media                *int           `json:"media,omitempty"`
	Converted            *bool          `json:"converted,omitempty"`
	Charset              *string        `json:"charset,omitempty"`
	Scripts              map[string]int `json:"scripts"`
}

func (f *facts) pdf(p media.PDFFacts) {
	f.Pages, f.RenderedPages, f.TextPages = new(p.Pages), new(p.RenderedPages), new(p.TextPages)
	f.DPIs, f.PagesWithLargeImages, f.LargeImagePPI = p.DPIs, new(p.PagesWithLargeImages), p.LargeImagePPI
	f.Fonts, f.FontsWithoutUnicode, f.InvisibleTextRuns = new(p.Fonts), new(p.FontsWithoutUnicode), new(p.InvisibleTextRuns)
}

func (f *facts) audioVideo(a media.AudioVideoFacts) {
	f.Duration, f.AudioStreams, f.VideoStreams = new(a.Duration), new(a.AudioStreams), new(a.VideoStreams)
	f.AttachedPictures, f.SampleRates, f.Channels = new(a.AttachedPictures), a.SampleRates, a.Channels
	f.SubtitleStreams = new(a.SubtitleStreams)
}

func (f *facts) image(i media.ImageFacts) {
	f.Frames = new(i.Frames)
	if i.Frames > 0 {
		f.Format, f.Width, f.Height = new(i.Format), new(i.Width), new(i.Height)
		f.Alpha, f.Orientation, f.Opaque = new(i.Alpha), new(i.Orientation), new(i.Opaque)
	}
}

func (f *facts) records(r mail.RecordsFacts) {
	if r.Messages > 0 {
		f.Messages = new(r.Messages)
	}
	f.FoldersWithMessages = new(r.FoldersWithMessages)
}

func (f *facts) onenote(o office.OneNoteFacts) {
	if o.Sections > 0 {
		f.Sections = new(o.Sections)
	}
	if o.Pages > 0 {
		f.Pages = new(o.Pages)
	}
}

var (
	audioVideoExts = idset.New(".wav", ".mp3", ".m4a", ".m4b", ".ogg", ".opus", ".flac", ".mp4", ".mkv", ".webm", ".avi", ".mpeg", ".ogv")
	imageExts      = idset.New(".emf", ".wmf", ".tif", ".heic", ".avif", ".webp", ".svg")
	textExts       = idset.New(".txt", ".csv", ".tsv", ".md", ".json", ".ndjson", ".xml", ".srt", ".html")
	archiveExts    = idset.New(".zip", ".7z", ".gz", ".zst")
	nestedArchive  = regexp.MustCompile(`(?i)\.(zip|7z|tar|gz|tgz|zst|rar)$`)
	xmlDeclaration = regexp.MustCompile(`<\?xml[^>]*encoding=["']([\w.-]+)`)
)

func readHead(path string, size int) (head []byte, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	head = make([]byte, size)
	n, err := io.ReadFull(f, head)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		err = nil
	}
	return head[:n], err
}

func pyExt(path string) string {
	base := filepath.Base(path)
	i := strings.LastIndex(base, ".")
	if i <= 0 || strings.Trim(base[:i], ".") == "" {
		return ""
	}
	return strings.ToLower(base[i:])
}

type prober struct {
	Placer
	rec         *run.Record
	conversions *office.Conversions
}

func (p *prober) probe(ctx context.Context, s sample, entry manifestEntry) (idset.Set, idset.Set, *facts, error) {
	rel := s.Path
	path := filepath.Join(p.Root, rel)
	library := idset.New(s.Library...)
	ext := pyExt(rel)
	traits, features := idset.Set{}, idset.Set{}
	found := &facts{}
	var body *string
	setText := func(t string) { body = &t }
	merge := func(t, f idset.Set) {
		traits.Merge(t)
		features.Merge(f)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if info.Size() >= 100*mib {
		traits.Add("trait:scale.large-file")
	}
	head, err := readHead(path, 8)
	if err != nil {
		return nil, nil, nil, err
	}
	switch {
	case bytes.HasPrefix(head, []byte("%PDF")) || ext == ".pdf":
		t, fe, fs := p.Media.PDF(ctx, path, p.rec.Outputs(rel))
		merge(t, fe)
		found.pdf(fs)
	case audioVideoExts[ext]:
		t, fe, fs := p.Media.AudioVideo(ctx, path)
		merge(t, fe)
		found.audioVideo(fs)
	case strings.HasPrefix(entry.Format, "image/") || imageExts[ext]:
		if ext != ".svg" {
			t, fe, fs := p.Media.Image(ctx, path)
			merge(t, fe)
			found.image(fs)
		}
	case ext == ".eml":
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		t, fe, b, fs := mail.Message(p.Parser, data)
		merge(t, fe)
		found.Parts = new(fs.Parts)
		setText(b)
	case ext == ".mbox":
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		t, fe, b, fs := mail.Mbox(p.Parser, data)
		merge(t, fe)
		found.Messages = new(fs.Messages)
		setText(b)
	case ext == ".msg" || ext == ".pst":
		t, fe, fs := mail.Records(p.rec.Rows(rel, "mail"))
		merge(t, fe)
		found.records(fs)
		if ext == ".msg" {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, nil, err
			}
			bodies := mail.MSGBodies(data)
			found.BodyProperties = bodies
			if slices.Equal(bodies, []string{"1013"}) {
				features.Add("feature:html-body")
			}
		}
	case ext == ".one" || ext == ".onetoc2" || ext == ".onepkg":
		outcomes := idset.Set{}
		for _, e := range p.rec.Events(rel) {
			outcomes.Add(string(e.Outcome))
		}
		t, fe, fs := office.OneNote(p.rec.Rows(rel, "onenote"), outcomes)
		merge(t, fe)
		found.onenote(fs)
	case ext == ".odt" || ext == ".ods" || ext == ".odp" || ext == ".odg":
		setText(office.ODFText(path))
	case archiveExts[ext]:
		entries := p.Media.ArchiveEntries(ctx, path)
		found.Entries = new(len(entries))
		if len(entries) >= 1000 {
			traits.Add("trait:scale.many-entries")
		}
		if library["kind:zip"] || library["kind:7z"] || library["kind:gzip"] || library["kind:zstd"] || library["kind:tar"] {
			if slices.ContainsFunc(entries, nestedArchive.MatchString) {
				features.Add("feature:nested-archives")
			}
		}
	}
	if office.IsOOXML(ext) || office.IsLegacy(ext) {
		target := path
		if !office.IsOOXML(ext) {
			target = p.conversions.Convert(ctx, path)
		}
		if target != "" {
			f, fs := office.OOXML(target)
			features.Merge(f)
			found.Parts, found.Tables, found.Media = new(fs.Parts), new(fs.Tables), new(fs.Media)
			found.Converted = new(target != path)
		}
	}
	if textExts[ext] {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		trait, guess := text.Charset(ctx, data, p.Charset)
		if trait != "" {
			traits.Add(trait)
		}
		found.Charset = new(guess)
		if guess != "bom" && guess != "utf-8" {
			guess = strings.ToLower(guess)
		}
		setText(text.DecodeText(data, guess))
		if m := xmlDeclaration.FindSubmatch(data[:min(200, len(data))]); m != nil && !text.UTF8Name(strings.ToLower(string(m[1]))) {
			traits.Add("trait:charset.declared")
		}
	}
	switch {
	case body == nil:
		setText(p.rec.Text(rel))
	case ext == ".eml" || ext == ".mbox":
		setText(*body + "\n" + p.rec.Text(rel))
	}
	scripts, counts := text.Scripts(*body)
	traits.Merge(scripts)
	found.Scripts = map[string]int{}
	for _, c := range counts {
		if c.Count >= 5 {
			found.Scripts[c.Name] = c.Count
		}
	}
	return traits, features, found, nil
}
