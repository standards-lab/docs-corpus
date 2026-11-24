package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type runOutput struct {
	Kind      string                     `json:"kind"`
	Name      string                     `json:"name"`
	Sha256    string                     `json:"sha256"`
	MediaType string                     `json:"media_type"`
	Meta      map[string]json.RawMessage `json:"meta"`
}

type runEvent struct {
	Event   string          `json:"event"`
	Input   string          `json:"input"`
	Outcome json.RawMessage `json:"outcome"`
	Part    struct {
		Depth *float64 `json:"depth"`
	} `json:"part"`
	Output *runOutput `json:"output"`
}

type record struct {
	name string
	row  *object
}

type runRecord struct {
	content string
	events  map[string][]runEvent
}

func loadRun(runs, id string) (_ *runRecord, err error) {
	r := &runRecord{content: filepath.Join(filepath.Dir(runs), "content"), events: map[string][]runEvent{}}
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
	root := manifest.Inputs[0] + "/"
	f, err := os.Open(filepath.Join(runs, id, "record.jsonl"))
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(nil, 1<<30)
	for scanner.Scan() {
		var e runEvent
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return nil, err
		}
		if rel, ok := strings.CutPrefix(e.Input, root); ok {
			r.events[rel] = append(r.events[rel], e)
		}
	}
	return r, scanner.Err()
}

func (r *runRecord) blob(sha string) []byte {
	data, err := os.ReadFile(filepath.Join(r.content, sha[:min(2, len(sha))], sha))
	if err != nil {
		return nil
	}
	return data
}

func (r *runRecord) outputs(path string) []runEvent {
	var out []runEvent
	for _, e := range r.events[path] {
		if e.Event == "output" {
			out = append(out, e)
		}
	}
	return out
}

func metaString(meta map[string]json.RawMessage, key string) string {
	var s string
	if raw, ok := meta[key]; ok {
		_ = json.Unmarshal(raw, &s)
	}
	return s
}

func metaRaw(meta map[string]json.RawMessage, key string) string {
	if raw, ok := meta[key]; ok {
		return string(raw)
	}
	return "null"
}

func (r *runRecord) text(path string) string {
	var parts []string
	for _, e := range r.outputs(path) {
		out := e.Output
		if out.Kind != "text" && out.Kind != "markdown" && out.Kind != "record" {
			continue
		}
		data := r.blob(out.Sha256)
		enc := metaString(out.Meta, "encoding")
		if enc == "" {
			enc = "utf-8"
		}
		body := decodeOrUTF8(data, enc)
		if out.Kind == "record" {
			for _, line := range pySplitLines(body) {
				if v, err := parseJSON([]byte(line)); err == nil {
					parts = append(parts, stringsUnder(v, recordTextKeys)...)
				}
			}
		} else {
			parts = append(parts, body)
		}
	}
	return strings.Join(parts, "\n")
}

var recordTextKeys = map[string]bool{"markdown": true, "subject": true, "text": true, "title": true}

func (r *runRecord) records(path, media string) []record {
	var rows []record
	for _, e := range r.outputs(path) {
		out := e.Output
		if out.Kind != "record" || !strings.Contains(out.MediaType, media) {
			continue
		}
		for _, line := range pySplitLines(string(r.blob(out.Sha256))) {
			v, err := parseJSON([]byte(line))
			if err != nil {
				continue
			}
			if o, ok := v.(*object); ok {
				rows = append(rows, record{out.Name, o})
			}
		}
	}
	return rows
}

func stringsUnder(v any, keys map[string]bool) []string {
	var out []string
	switch t := v.(type) {
	case *object:
		for _, k := range t.keys {
			if s, ok := t.values[k].(string); ok && keys[k] {
				out = append(out, s)
			} else {
				out = append(out, stringsUnder(t.values[k], keys)...)
			}
		}
	case []any:
		for _, item := range t {
			out = append(out, stringsUnder(item, keys)...)
		}
	}
	return out
}

func pySplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); {
		r, size := rune(s[i]), 1
		if r >= 0x80 {
			r, size = utf8.DecodeRuneInString(s[i:])
		}
		switch r {
		case '\n', '\v', '\f', 0x1C, 0x1D, 0x1E, 0x85, 0x2028, 0x2029:
			lines = append(lines, s[start:i])
			start = i + size
		case '\r':
			lines = append(lines, s[start:i])
			if i+1 < len(s) && s[i+1] == '\n' {
				size = 2
			}
			start = i + size
		}
		i += size
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
