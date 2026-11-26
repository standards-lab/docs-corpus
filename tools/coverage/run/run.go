package run

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/text"
)

// Output is an output event's output: what docproc wrote for one part of an input.
type Output struct {
	Kind      string                     `json:"kind"`
	Name      string                     `json:"name"`
	Sha256    string                     `json:"sha256"`
	MediaType string                     `json:"media_type"`
	Meta      map[string]json.RawMessage `json:"meta"`
}

// MetaRaw returns the JSON text of the output's meta key, or "null" when it has none.
func (o *Output) MetaRaw(key string) string {
	if raw, ok := o.Meta[key]; ok {
		return string(raw)
	}
	return "null"
}

// Event is one line of a run's record.jsonl. For a part the run replayed from its run
// directory's index, whose outcome is seen, Load sets Outcome to Recorded, the outcome the part
// first had.
type Event struct {
	Event    string          `json:"event"`
	Input    string          `json:"input"`
	Outcome  json.RawMessage `json:"outcome"`
	Recorded json.RawMessage `json:"recorded"`
	Part     Part            `json:"part"`
	Output   *Output         `json:"output"`
}

// Part is the part of an input an event is about: its depth, the input at 0, its content type,
// and the flags docproc's checks raised on it.
type Part struct {
	Depth       *float64    `json:"depth"`
	ContentType ContentType `json:"content_type"`
	Flags       []Flag      `json:"flags"`
}

// ContentType is a part's kind and the route docproc dispatched it on; Route is nil when the
// record has none.
type ContentType struct {
	Kind  string  `json:"kind"`
	Route *string `json:"route"`
}

// Flag is one flag a docproc check raised on a part.
type Flag struct {
	Check string `json:"check"`
}

// Row is one line of a record output: the output's name and the line's JSON object.
type Row struct {
	Name   string
	Fields json.RawMessage
}

// Record is a docproc run's events, by input path relative to the run's input, and the content
// store its outputs' blobs are in.
type Record struct {
	content string
	events  map[string][]Event
}

const schema = "2"

// Load reads the run id under runs: its manifest's inputs, whose deepest common directory is the
// root the events' input paths are taken relative to, the one input itself when there is one, and
// its record.jsonl, a replayed part's event taking its recorded outcome. The run directory is
// runs' parent: Load refuses one whose rundir.json is missing or names a schema other than "2",
// and its content store is runs' sibling, content.
func Load(runs, id string) (_ *Record, err error) {
	dir := filepath.Dir(runs)
	if err := checkSchema(dir); err != nil {
		return nil, err
	}
	r := &Record{content: filepath.Join(dir, "content"), events: map[string][]Event{}}
	data, err := os.ReadFile(filepath.Join(runs, id, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var manifest struct {
		Inputs []string `json:"inputs"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	if len(manifest.Inputs) == 0 {
		return nil, errors.New("run: " + id + ": its manifest names no input")
	}
	root := strings.TrimSuffix(commonRoot(manifest.Inputs), "/") + "/"
	f, err := os.Open(filepath.Join(runs, id, "record.jsonl"))
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(nil, 1<<30)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return nil, err
		}
		if len(e.Recorded) > 0 {
			e.Outcome = e.Recorded
		}
		if rel, ok := strings.CutPrefix(e.Input, root); ok {
			r.events[rel] = append(r.events[rel], e)
		}
	}
	return r, scanner.Err()
}

func commonRoot(inputs []string) string {
	root := inputs[0]
	for _, in := range inputs[1:] {
		for root != in && !strings.HasPrefix(in, strings.TrimSuffix(root, "/")+"/") && root != filepath.Dir(root) {
			root = filepath.Dir(root)
		}
	}
	return root
}

func checkSchema(dir string) error {
	path := filepath.Join(dir, "rundir.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("run: %s holds no rundir.json, so it is no schema %q run directory; run docproc v0.9.0 again to write one: %w",
			dir, schema, err)
	}
	if err != nil {
		return err
	}
	var f struct {
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("run: %s: %w", path, err)
	}
	if f.Schema != schema {
		return fmt.Errorf("run: %s is schema %q; run reads schema %q", dir, f.Schema, schema)
	}
	return nil
}

// Events returns the events of the input at path, in the record's order.
func (r *Record) Events(path string) []Event {
	return r.events[path]
}

func (r *Record) blob(sha string) []byte {
	data, err := os.ReadFile(filepath.Join(r.content, sha[:min(2, len(sha))], sha))
	if err != nil {
		return nil
	}
	return data
}

// Outputs returns the output events of the input at path, in the record's order.
func (r *Record) Outputs(path string) []Event {
	var out []Event
	for _, e := range r.events[path] {
		if e.Event == "output" {
			out = append(out, e)
		}
	}
	return out
}

// Text returns the text the run extracted from the input at path, joined by newlines: each text
// and markdown output read as UTF-8, as docproc writes them whatever their encoding and
// encoding_by meta say of the source, and the markdown, subject, text, and title strings of each
// record output's lines.
func (r *Record) Text(path string) string {
	var parts []string
	for _, e := range r.Outputs(path) {
		out := e.Output
		if out.Kind != "text" && out.Kind != "markdown" && out.Kind != "record" {
			continue
		}
		body := string(r.blob(out.Sha256))
		if out.Kind == "record" {
			for _, line := range text.SplitLines(body) {
				if found, err := stringsUnder([]byte(line), recordTextKeys); err == nil {
					parts = append(parts, found...)
				}
			}
		} else {
			parts = append(parts, body)
		}
	}
	return strings.Join(parts, "\n")
}

var recordTextKeys = map[string]bool{"markdown": true, "subject": true, "text": true, "title": true}

// Rows returns the object lines of the input's record outputs whose media type contains media,
// such as "mail" or "onenote", in order.
func (r *Record) Rows(path, media string) []Row {
	var rows []Row
	for _, e := range r.Outputs(path) {
		out := e.Output
		if out.Kind != "record" || !strings.Contains(out.MediaType, media) {
			continue
		}
		for _, line := range text.SplitLines(string(r.blob(out.Sha256))) {
			fields := bytes.TrimSpace([]byte(line))
			if bytes.HasPrefix(fields, []byte("{")) && json.Valid(fields) {
				rows = append(rows, Row{out.Name, fields})
			}
		}
	}
	return rows
}

func stringsUnder(line []byte, keys map[string]bool) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(line))
	var out []string
	if err := walk(dec, "", keys, &out); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("run: data after a record line's value")
	}
	return out, nil
}

func walk(dec *json.Decoder, key string, keys map[string]bool, out *[]string) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	switch t := tok.(type) {
	case json.Delim:
		for dec.More() {
			child := ""
			if t == '{' {
				k, err := dec.Token()
				if err != nil {
					return err
				}
				child, _ = k.(string)
			}
			if err := walk(dec, child, keys, out); err != nil {
				return err
			}
		}
		_, err := dec.Token()
		return err
	case string:
		if keys[key] {
			*out = append(*out, t)
		}
	}
	return nil
}
