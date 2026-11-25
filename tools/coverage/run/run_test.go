package run_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
)

func write(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func load(t *testing.T) *run.Record {
	t.Helper()
	root := t.TempDir()
	runs := filepath.Join(root, "runs")
	write(t, filepath.Join(runs, "r1", "manifest.json"), `{"inputs": ["/in"]}`)
	write(t, filepath.Join(runs, "r1", "record.jsonl"), strings.Join([]string{
		`{"event": "start", "input": "/in/a.pdf"}`,
		`{"event": "output", "input": "/in/a.pdf", "output": {"kind": "text", "name": "page 1", "sha256": "ab01", "media_type": "text/plain", "meta": {"page": 1, "encoding": "ISO-8859-1"}}}`,
		`{"event": "output", "input": "/in/a.pdf", "part": {"depth": 1}, "output": {"kind": "image", "name": "page 1", "sha256": "ab02", "media_type": "image/png"}}`,
		`{"event": "output", "input": "/in/m.pst", "output": {"kind": "record", "name": "message 1", "sha256": "cd01", "media_type": "application/x-docproc-mail+jsonl"}}`,
		`{"event": "output", "input": "/elsewhere/a.pdf", "output": {"kind": "text", "name": "page 1", "sha256": "ab01"}}`,
		`{"event": "end", "input": "/in/m.pst", "outcome": "encrypted"}`,
	}, "\n")+"\n")
	write(t, filepath.Join(root, "content", "ab", "ab01"), "caf\xe9")
	write(t, filepath.Join(root, "content", "cd", "cd01"),
		`{"subject": "Hi", "body": {"text": "T", "n": 1}}`+"\n"+`{"x": [{"title": "Z"}]}`+"\n"+`not json`+"\n"+`["a list"]`+"\n")
	rec, err := run.Load(runs, "r1")
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

func TestLoadGroupsEventsByPathUnderTheRunsInput(t *testing.T) {
	rec := load(t)
	if got := len(rec.Events("a.pdf")); got != 3 {
		t.Errorf("a.pdf has %d events, want 3", got)
	}
	if got := len(rec.Outputs("a.pdf")); got != 2 {
		t.Errorf("a.pdf has %d outputs, want 2", got)
	}
	if got := rec.Events("/elsewhere/a.pdf"); got != nil {
		t.Errorf("an input outside the run's input has events: %v", got)
	}
	ends := rec.Events("m.pst")
	if got := string(ends[len(ends)-1].Outcome); got != `"encrypted"` {
		t.Errorf("m.pst's outcome = %s", got)
	}
}

func TestOutputMetaRawIsTheValuesJSON(t *testing.T) {
	out := load(t).Outputs("a.pdf")
	if got := out[0].Output.MetaRaw("page"); got != "1" {
		t.Errorf("MetaRaw(page) = %q, want 1", got)
	}
	if got := out[0].Output.MetaRaw("dpi"); got != "null" {
		t.Errorf("MetaRaw(dpi) = %q, want null", got)
	}
	if d := out[1].Part.Depth; d == nil || *d != 1 {
		t.Errorf("the second output's depth = %v, want 1", d)
	}
}

func TestTextDecodesOutputsAndReadsRecordStrings(t *testing.T) {
	rec := load(t)
	if got := rec.Text("a.pdf"); got != "café" {
		t.Errorf("Text(a.pdf) = %q, want the ISO-8859-1 text decoded", got)
	}
	if got := rec.Text("m.pst"); got != "Hi\nT\nZ" {
		t.Errorf("Text(m.pst) = %q, want the record's subject, text, and title", got)
	}
	if got := rec.Text("missing"); got != "" {
		t.Errorf("Text(missing) = %q", got)
	}
}

func TestRowsAreTheRecordOutputsObjectLines(t *testing.T) {
	rec := load(t)
	rows := rec.Rows("m.pst", "mail")
	if len(rows) != 2 {
		t.Fatalf("Rows = %d, want the 2 object lines", len(rows))
	}
	if rows[0].Name != "message 1" || rows[0].Fields.Str("subject") != "Hi" {
		t.Errorf("first row = %s %s", rows[0].Name, jsonvalue.Marshal(rows[0].Fields))
	}
	if got := rec.Rows("m.pst", "onenote"); len(got) != 0 {
		t.Errorf("Rows for another media type = %d, want none", len(got))
	}
}

func TestLoadFailsOnAMissingOrEmptyManifest(t *testing.T) {
	runs := t.TempDir()
	if _, err := run.Load(runs, "r1"); err == nil {
		t.Error("Load of a missing run succeeded")
	}
	write(t, filepath.Join(runs, "r2", "manifest.json"), `{"inputs": []}`)
	write(t, filepath.Join(runs, "r2", "record.jsonl"), "")
	if _, err := run.Load(runs, "r2"); err == nil {
		t.Error("Load of a run with no input succeeded")
	}
}
