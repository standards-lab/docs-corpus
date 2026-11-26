package run_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	write(t, filepath.Join(root, "rundir.json"), `{"schema": "2"}`)
	write(t, filepath.Join(runs, "r1", "manifest.json"), `{"inputs": ["/in"]}`)
	write(t, filepath.Join(runs, "r1", "record.jsonl"), strings.Join([]string{
		`{"event": "start", "input": "/in/a.pdf"}`,
		`{"event": "output", "input": "/in/a.pdf", "output": {"kind": "text", "name": "page 1", "sha256": "ab01", "media_type": "text/plain", "meta": {"page": 1, "encoding": "windows-1252", "encoding_by": "fallback"}}}`,
		`{"event": "output", "input": "/in/b.txt", "output": {"kind": "text", "name": "text", "sha256": "ef01", "media_type": "text/plain", "meta": {"encoding": "windows-1252 (fallback)"}}}`,
		`{"event": "output", "input": "/in/b.txt", "output": {"kind": "markdown", "name": "text", "sha256": "ef02", "media_type": "text/markdown", "meta": {"encoding": "utf-16le (byte order mark)"}}}`,
		`{"event": "output", "input": "/in/c.txt", "output": {"kind": "text", "name": "text", "sha256": "ef03", "media_type": "text/plain", "meta": {"encoding": "utf-16le", "encoding_by": "bom"}}}`,
		`{"event": "output", "input": "/in/a.pdf", "part": {"depth": 1}, "output": {"kind": "image", "name": "page 1", "sha256": "ab02", "media_type": "image/png"}}`,
		`{"event": "output", "input": "/in/m.pst", "output": {"kind": "record", "name": "message 1", "sha256": "cd01", "media_type": "application/x-docproc-mail+jsonl"}}`,
		`{"event": "output", "input": "/elsewhere/a.pdf", "output": {"kind": "text", "name": "page 1", "sha256": "ab01"}}`,
		`{"event": "end", "input": "/in/m.pst", "outcome": "encrypted"}`,
		`{"event": "part", "input": "/in/n.one", "outcome": "seen", "recorded": "encrypted"}`,
		`{"event": "part", "input": "/in/n.one", "outcome": "seen"}`,
	}, "\n")+"\n")
	write(t, filepath.Join(root, "content", "ab", "ab01"), "café")
	write(t, filepath.Join(root, "content", "ef", "ef01"), "naïve")
	write(t, filepath.Join(root, "content", "ef", "ef02"), "# Ünïcode")
	write(t, filepath.Join(root, "content", "ef", "ef03"), "żółw")
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

func TestLoadTakesAReplayedPartsRecordedOutcome(t *testing.T) {
	events := load(t).Events("n.one")
	var got []string
	for _, e := range events {
		got = append(got, string(e.Outcome))
	}
	if want := []string{`"encrypted"`, `"seen"`}; strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("n.one's outcomes = %q, want %q", got, want)
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

func TestTextReadsOutputsAsUTF8AndReadsRecordStrings(t *testing.T) {
	rec := load(t)
	for path, want := range map[string]string{"a.pdf": "café", "b.txt": "naïve\n# Ünïcode", "c.txt": "żółw"} {
		if got := rec.Text(path); got != want {
			t.Errorf("Text(%s) = %q, want the UTF-8 output unchanged, %q, whatever its encoding meta names", path, got, want)
		}
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
	if got := rows[0].Name + " " + string(rows[0].Fields); got != `message 1 {"subject": "Hi", "body": {"text": "T", "n": 1}}` {
		t.Errorf("first row = %s", got)
	}
	if got := rec.Rows("m.pst", "onenote"); len(got) != 0 {
		t.Errorf("Rows for another media type = %d, want none", len(got))
	}
}

func TestLoadFailsOnAMissingOrEmptyManifest(t *testing.T) {
	root := t.TempDir()
	runs := filepath.Join(root, "runs")
	write(t, filepath.Join(root, "rundir.json"), `{"schema": "2"}`)
	if _, err := run.Load(runs, "r1"); err == nil {
		t.Error("Load of a missing run succeeded")
	}
	write(t, filepath.Join(runs, "r2", "manifest.json"), `{"inputs": []}`)
	write(t, filepath.Join(runs, "r2", "record.jsonl"), "")
	if _, err := run.Load(runs, "r2"); err == nil {
		t.Error("Load of a run with no input succeeded")
	}
}

func TestLoadRefusesARunDirectoryOfAnotherSchema(t *testing.T) {
	root := t.TempDir()
	runs := filepath.Join(root, "runs")
	write(t, filepath.Join(runs, "r1", "manifest.json"), `{"inputs": ["/in"]}`)
	write(t, filepath.Join(runs, "r1", "record.jsonl"), "")
	write(t, filepath.Join(root, "rundir.json"), `{"schema": "3"}`)
	_, err := run.Load(runs, "r1")
	if err == nil || !strings.Contains(err.Error(), `schema "3"`) || !strings.Contains(err.Error(), `reads schema "2"`) {
		t.Errorf("Load of a schema 3 run directory = %v, want an error naming schema \"3\" and the \"2\" it reads", err)
	}
}

func TestLoadRefusesARunDirectoryWithNoRundirJSON(t *testing.T) {
	root := t.TempDir()
	runs := filepath.Join(root, "runs")
	write(t, filepath.Join(runs, "r1", "manifest.json"), `{"inputs": ["/in"]}`)
	write(t, filepath.Join(runs, "r1", "record.jsonl"), "")
	_, err := run.Load(runs, "r1")
	if err == nil || !strings.Contains(err.Error(), "no rundir.json") || !strings.Contains(err.Error(), "docproc v0.9.0") {
		t.Errorf("Load of a run directory with no rundir.json = %v, want an error naming it and the docproc to run again", err)
	}
}

func TestLoadTakesPathsRelativeToTheInputsCommonDirectory(t *testing.T) {
	root := t.TempDir()
	runs := filepath.Join(root, "runs")
	write(t, filepath.Join(root, "rundir.json"), `{"schema": "2"}`)
	write(t, filepath.Join(runs, "r1", "manifest.json"), `{"inputs": ["/corpus/audio", "/corpus/email", "/corpus/pdf"]}`)
	write(t, filepath.Join(runs, "r1", "record.jsonl"), strings.Join([]string{
		`{"event": "start", "input": "/corpus/audio/a.wav"}`,
		`{"event": "start", "input": "/corpus/pdf/edge/b.pdf"}`,
		`{"event": "start", "input": "/corpusx/c.pdf"}`,
	}, "\n")+"\n")
	rec, err := run.Load(runs, "r1")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"audio/a.wav", "pdf/edge/b.pdf"} {
		if got := len(rec.Events(path)); got != 1 {
			t.Errorf("%s has %d events, want 1", path, got)
		}
	}
	if got := rec.Events("../corpusx/c.pdf"); got != nil {
		t.Errorf("an input outside the inputs' directory has events: %v", got)
	}
}

func TestEventPartHasItsContentTypeAndFlags(t *testing.T) {
	root := t.TempDir()
	runs := filepath.Join(root, "runs")
	write(t, filepath.Join(root, "rundir.json"), `{"schema": "2"}`)
	write(t, filepath.Join(runs, "r1", "manifest.json"), `{"inputs": ["/in"]}`)
	write(t, filepath.Join(runs, "r1", "record.jsonl"),
		`{"event": "part", "input": "/in/a.zip", "part": {"depth": 0, "content_type": {"kind": "zip", "route": "archive"}, "flags": [{"check": "bomb"}]}}`+"\n"+
			`{"event": "part", "input": "/in/a.zip", "part": {"depth": 1, "content_type": {"kind": ""}}}`+"\n")
	rec, err := run.Load(runs, "r1")
	if err != nil {
		t.Fatal(err)
	}
	events := rec.Events("a.zip")
	first, second := events[0].Part, events[1].Part
	if first.ContentType.Kind != "zip" || first.ContentType.Route == nil || *first.ContentType.Route != "archive" {
		t.Errorf("the first part's content type = %+v, want zip on archive", first.ContentType)
	}
	if len(first.Flags) != 1 || first.Flags[0].Check != "bomb" {
		t.Errorf("the first part's flags = %+v, want bomb", first.Flags)
	}
	if second.ContentType.Route != nil || second.Flags != nil {
		t.Errorf("the second part = %+v, want no route and no flags", second)
	}
}
