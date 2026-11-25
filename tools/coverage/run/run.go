package run

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

func (o *Output) metaString(key string) string {
	var s string
	if raw, ok := o.Meta[key]; ok {
		_ = json.Unmarshal(raw, &s)
	}
	return s
}

// Event is one line of a run's record.jsonl.
type Event struct {
	Event   string          `json:"event"`
	Input   string          `json:"input"`
	Outcome json.RawMessage `json:"outcome"`
	Part    struct {
		Depth *float64 `json:"depth"`
	} `json:"part"`
	Output *Output `json:"output"`
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

// Load reads the run id under runs: its manifest's first input, the root the events' input paths
// are taken relative to, and its record.jsonl. The content store is runs' sibling, content.
func Load(runs, id string) (_ *Record, err error) {
	r := &Record{content: filepath.Join(filepath.Dir(runs), "content"), events: map[string][]Event{}}
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
	root := manifest.Inputs[0] + "/"
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
		if rel, ok := strings.CutPrefix(e.Input, root); ok {
			r.events[rel] = append(r.events[rel], e)
		}
	}
	return r, scanner.Err()
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
// and markdown output decoded by its meta encoding, UTF-8 when it names none, and the markdown,
// subject, text, and title strings of each record output's lines.
func (r *Record) Text(path string) string {
	var parts []string
	for _, e := range r.Outputs(path) {
		out := e.Output
		if out.Kind != "text" && out.Kind != "markdown" && out.Kind != "record" {
			continue
		}
		data := r.blob(out.Sha256)
		enc := out.metaString("encoding")
		if enc == "" {
			enc = "utf-8"
		}
		body := text.DecodeOrUTF8(data, enc)
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
