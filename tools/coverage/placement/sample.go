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
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
	"github.com/standards-lab/docs-corpus/tools/coverage/mail"
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

func (p *prober) probe(ctx context.Context, s sample, entry manifestEntry) (idset.Set, idset.Set, *jsonvalue.Object, error) {
	rel := s.Path
	path := filepath.Join(p.Docs, rel)
	library := idset.New(s.Library...)
	ext := pyExt(rel)
	traits, features := idset.Set{}, idset.Set{}
	facts := jsonvalue.NewObject()
	var body *string
	setText := func(t string) { body = &t }
	merge := func(t, f idset.Set, fs *jsonvalue.Object) {
		traits.Merge(t)
		features.Merge(f)
		facts = fs
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
		merge(p.Media.PDF(ctx, path, p.rec.Outputs(rel)))
	case audioVideoExts[ext]:
		merge(p.Media.AudioVideo(ctx, path))
	case strings.HasPrefix(entry.Format, "image/") || imageExts[ext]:
		if ext != ".svg" {
			merge(p.Media.Image(ctx, path))
		}
	case ext == ".eml":
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		t, fe, b, fs := mail.Message(p.Parser, data)
		merge(t, fe, fs)
		setText(b)
	case ext == ".mbox":
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		t, fe, b, fs := mail.Mbox(p.Parser, data)
		merge(t, fe, fs)
		setText(b)
	case ext == ".msg" || ext == ".pst":
		merge(mail.Records(p.rec.Rows(rel, "mail")))
		if ext == ".msg" {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, nil, err
			}
			bodies := mail.MSGBodies(data)
			facts.Set("body_properties", bodies)
			if slices.Equal(bodies, []string{"1013"}) {
				features.Add("feature:html-body")
			}
		}
	case ext == ".one" || ext == ".onetoc2":
		outcomes := idset.Set{}
		for _, e := range p.rec.Events(rel) {
			outcomes.Add(string(e.Outcome))
		}
		merge(office.OneNote(p.rec.Rows(rel, "onenote"), outcomes))
	case ext == ".odt" || ext == ".ods" || ext == ".odp" || ext == ".odg":
		setText(office.ODFText(path))
	case archiveExts[ext]:
		entries := p.Media.ArchiveEntries(ctx, path)
		facts.Set("entries", len(entries))
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
			f, ofacts := office.OOXML(target)
			features.Merge(f)
			for _, k := range ofacts.Keys() {
				facts.Set(k, ofacts.Get(k))
			}
			facts.Set("converted", target != path)
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
		facts.Set("charset", guess)
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
	found := jsonvalue.NewObject()
	for _, c := range counts {
		if c.Count >= 5 {
			found.Set(c.Name, c.Count)
		}
	}
	facts.Set("scripts", found)
	return traits, features, facts, nil
}
