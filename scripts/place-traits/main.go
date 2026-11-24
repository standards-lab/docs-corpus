// Command place-traits prints the trait and feature placement of every ~/docs sample as a JSON
// array, sorted by path. Run it from the repository root.
//
// With --facts, each element also has the probes' findings. It writes nothing outside a
// temporary directory.
//
// See docs/coverage.md
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	mib            = 1024 * 1024
	probeTimeout   = 600 * time.Second
	convertTimeout = 300 * time.Second
)

var probeTools = []string{"pdfinfo", "pdfimages", "pdffonts", "qpdf", "ffprobe", "magick", "uchardet", "bsdtar", "soffice"}

func requireProbeTools() error {
	for _, tool := range probeTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s is not on PATH; install it", tool)
		}
	}
	return nil
}

func run(args []string, input []byte, timeout time.Duration) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.WaitDelay = 10 * time.Second
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil
	}
	var execErr *exec.Error
	var pathErr *os.PathError
	if errors.As(err, &execErr) || errors.As(err, &pathErr) {
		return nil
	}
	return out.Bytes()
}

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

type sample struct {
	Path    string                     `json:"path"`
	Library []string                   `json:"library"`
	Judged  map[string]json.RawMessage `json:"judged"`
}

type manifestEntry struct {
	Format string `json:"format"`
}

var (
	audioVideoExts = newSet(".wav", ".mp3", ".m4a", ".m4b", ".ogg", ".opus", ".flac", ".mp4", ".mkv", ".webm", ".avi", ".mpeg", ".ogv")
	imageExts      = newSet(".emf", ".wmf", ".tif", ".heic", ".avif", ".webp", ".svg")
	textExts       = newSet(".txt", ".csv", ".tsv", ".md", ".json", ".ndjson", ".xml", ".srt", ".html")
	archiveExts    = newSet(".zip", ".7z", ".gz", ".zst")
	nestedArchive  = regexp.MustCompile(`(?i)\.(zip|7z|tar|gz|tgz|zst|rar)$`)
	xmlDeclaration = regexp.MustCompile(`<\?xml[^>]*encoding=["']([\w.-]+)`)
)

type prober struct {
	docs string
	rec  *runRecord
	tmp  string
	n    int
}

func (p *prober) probe(s sample, entry manifestEntry) (set, set, *object, error) {
	rel := s.Path
	path := filepath.Join(p.docs, rel)
	library := newSet(s.Library...)
	ext := pyExt(rel)
	traits, features := set{}, set{}
	facts := newObject()
	var text *string
	setText := func(t string) { text = &t }
	merge := func(t, f set, fs *object) {
		traits.merge(t)
		features.merge(f)
		facts = fs
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if info.Size() >= 100*mib {
		traits.add("trait:scale.large-file")
	}
	head, err := readHead(path, 8)
	if err != nil {
		return nil, nil, nil, err
	}
	switch {
	case bytes.HasPrefix(head, []byte("%PDF")) || ext == ".pdf":
		merge(pdf(path, rel, p.rec))
	case audioVideoExts[ext]:
		merge(audioVideo(path))
	case strings.HasPrefix(entry.Format, "image/") || imageExts[ext]:
		if ext != ".svg" {
			merge(image(path))
		}
	case ext == ".eml":
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		t, fe, body, fs := message(data)
		merge(t, fe, fs)
		setText(body)
	case ext == ".mbox":
		t, fe, body, fs, err := mbox(path)
		if err != nil {
			return nil, nil, nil, err
		}
		merge(t, fe, fs)
		setText(body)
	case ext == ".msg" || ext == ".pst":
		merge(mailRecords(p.rec.records(rel, "mail")))
		if ext == ".msg" {
			bodies, err := msgBodies(path)
			if err != nil {
				return nil, nil, nil, err
			}
			facts.set("body_properties", bodies)
			if slices.Equal(bodies, []string{"1013"}) {
				features.add("feature:html-body")
			}
		}
	case ext == ".one" || ext == ".onetoc2":
		outcomes := set{}
		for _, e := range p.rec.events[rel] {
			outcomes.add(string(e.Outcome))
		}
		merge(onenote(p.rec.records(rel, "onenote"), outcomes))
	case ext == ".odt" || ext == ".ods" || ext == ".odp" || ext == ".odg":
		setText(odfText(path))
	case archiveExts[ext]:
		entries := archiveEntries(path)
		facts.set("entries", len(entries))
		if len(entries) >= 1000 {
			traits.add("trait:scale.many-entries")
		}
		if library["kind:zip"] || library["kind:7z"] || library["kind:gzip"] || library["kind:zstd"] || library["kind:tar"] {
			if slices.ContainsFunc(entries, nestedArchive.MatchString) {
				features.add("feature:nested-archives")
			}
		}
	}
	kind := strings.TrimPrefix(ext, ".")
	if ooxmlKinds[kind] || legacyTarget[ext] != "" {
		target := path
		if !ooxmlKinds[kind] {
			p.n++
			target = converted(path, p.tmp, p.n)
		}
		if target != "" {
			f, ofacts := ooxml(target)
			features.merge(f)
			for _, k := range ofacts.keys {
				facts.set(k, ofacts.values[k])
			}
			facts.set("converted", target != path)
		}
	}
	if textExts[ext] {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, nil, err
		}
		trait, guess := charsetOf(data)
		if trait != "" {
			traits.add(trait)
		}
		facts.set("charset", guess)
		if guess != "bom" && guess != "utf-8" {
			guess = strings.ToLower(guess)
		}
		setText(decodeText(data, guess))
		if m := xmlDeclaration.FindSubmatch(data[:min(200, len(data))]); m != nil && !utf8Names[strings.ToLower(string(m[1]))] {
			traits.add("trait:charset.declared")
		}
	}
	switch {
	case text == nil:
		setText(p.rec.text(rel))
	case ext == ".eml" || ext == ".mbox":
		setText(*text + "\n" + p.rec.text(rel))
	}
	scripts, counts := scriptsIn(*text)
	traits.merge(scripts)
	found := newObject()
	for _, c := range counts {
		if c.count >= 5 {
			found.set(c.name, c.count)
		}
	}
	facts.set("scripts", found)
	return traits, features, facts, nil
}

func sorted(s set, prefix string) []string {
	out := []string{}
	for id := range s {
		if strings.HasPrefix(id, prefix) {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func placeTraits(withFacts bool) (err error) {
	if err := requireProbeTools(); err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	docs := envOr("DOCS_REFERENCE", filepath.Join(home, "docs"))
	runs := envOr("RUNS", filepath.Join(home, "runs/docproc/final/runs"))
	data, err := os.ReadFile("docs/coverage.json")
	if err != nil {
		return err
	}
	var coverage struct {
		Samples          []sample `json:"samples"`
		LibraryPlacement struct {
			Run string `json:"run"`
		} `json:"library_placement"`
	}
	if err := json.Unmarshal(data, &coverage); err != nil {
		return err
	}
	data, err = os.ReadFile(filepath.Join(docs, "MANIFEST.json"))
	if err != nil {
		return err
	}
	var manifest struct {
		Files []struct {
			Path string `json:"path"`
			manifestEntry
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	entries := map[string]manifestEntry{}
	for _, f := range manifest.Files {
		entries[f.Path] = f.manifestEntry
	}
	rec, err := loadRun(runs, coverage.LibraryPlacement.Run)
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "place-traits-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	p := &prober{docs: docs, rec: rec, tmp: tmp}
	slices.SortFunc(coverage.Samples, func(a, b sample) int { return strings.Compare(a.Path, b.Path) })
	out := []any{}
	for _, s := range coverage.Samples {
		entry, ok := entries[s.Path]
		if !ok {
			return fmt.Errorf("%s: not in %s/MANIFEST.json", s.Path, docs)
		}
		traits, features, facts, err := p.probe(s, entry)
		if err != nil {
			return err
		}
		placed := set{}
		placed.merge(traits)
		placed.merge(features)
		for id := range s.Judged {
			placed.add(id)
		}
		row := newObject()
		row.set("path", s.Path)
		row.set("traits", sorted(placed, "trait:"))
		row.set("features", sorted(placed, "feature:"))
		if withFacts {
			probed := set{}
			probed.merge(traits)
			probed.merge(features)
			row.set("facts", facts)
			row.set("probed", sorted(probed, ""))
		}
		out = append(out, row)
	}
	var buf bytes.Buffer
	writeJSON(&buf, out, 0)
	buf.WriteByte('\n')
	_, err = os.Stdout.Write(buf.Bytes())
	return err
}

func main() {
	if err := placeTraits(slices.Contains(os.Args[1:], "--facts")); err != nil {
		fmt.Fprintln(os.Stderr, "place-traits:", err)
		os.Exit(1)
	}
}
