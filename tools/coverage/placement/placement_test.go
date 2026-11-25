package placement_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/placement"
	"github.com/standards-lab/docs-corpus/tools/coverage/stdmail"
)

type guesser string

func (g guesser) Guess(context.Context, []byte) []byte { return []byte(g) }

type soffice struct{}

func (soffice) Convert(context.Context, string, string, string, string) {}

func write(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func placer(t *testing.T, samples string) placement.Placer {
	t.Helper()
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	write(t, filepath.Join(docs, "MANIFEST.json"), `{"files": [
		{"path": "text/a.txt", "format": "text/plain"},
		{"path": "text/b.txt", "format": "text/plain"},
		{"path": "email/c.eml", "format": "message/rfc822"}]}`)
	write(t, filepath.Join(docs, "text", "a.txt"), strings.Repeat("ж", 25))
	write(t, filepath.Join(docs, "text", "b.txt"), "caf\xe9 au lait")
	write(t, filepath.Join(docs, "email", "c.eml"), "Subject: x\nContent-Type: text/html\n\n<p>hi</p>\n")
	runs := filepath.Join(root, "runs")
	write(t, filepath.Join(root, "rundir.json"), `{"schema": "2"}`)
	write(t, filepath.Join(runs, "r1", "manifest.json"), `{"inputs": ["/docs"]}`)
	write(t, filepath.Join(runs, "r1", "record.jsonl"), "")
	coverage := filepath.Join(root, "coverage.json")
	write(t, coverage, `{"library_placement": {"run": "r1"}, "samples": [`+samples+`]}`)
	t.Setenv("TMPDIR", filepath.Join(root, "tmp"))
	if err := os.Mkdir(filepath.Join(root, "tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	return placement.Placer{
		Docs:     docs,
		Runs:     runs,
		Coverage: coverage,
		Charset:  guesser("WINDOWS-1252\n"),
		Soffice:  soffice{},
		Parser:   stdmail.Parser{},
	}
}

const samples = `
	{"path": "text/b.txt", "library": [], "judged": {"trait:scan.handwriting": "seen in the scan"}},
	{"path": "email/c.eml", "library": []},
	{"path": "text/a.txt", "library": []}`

func TestRunPlacesEverySampleSortedByPath(t *testing.T) {
	var out bytes.Buffer
	if err := placer(t, samples).Run(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	want := `[
  {
    "path": "email/c.eml",
    "traits": [],
    "features": [
      "feature:html-body"
    ]
  },
  {
    "path": "text/a.txt",
    "traits": [
      "trait:script.cyrillic"
    ],
    "features": []
  },
  {
    "path": "text/b.txt",
    "traits": [
      "trait:charset.windows-1252",
      "trait:scan.handwriting"
    ],
    "features": []
  }
]
`
	if got := out.String(); got != want {
		t.Errorf("Run wrote\n%s\nwant\n%s", got, want)
	}
}

func TestRunWithFactsAddsTheProbesFindings(t *testing.T) {
	p := placer(t, `{"path": "text/b.txt", "library": [], "judged": {"trait:scan.handwriting": "seen"}}`)
	p.Facts = true
	var out bytes.Buffer
	if err := p.Run(context.Background(), &out); err != nil {
		t.Fatal(err)
	}
	want := `[
  {
    "path": "text/b.txt",
    "traits": [
      "trait:charset.windows-1252",
      "trait:scan.handwriting"
    ],
    "features": [],
    "facts": {
      "charset": "WINDOWS-1252",
      "scripts": {
        "latin": 10
      }
    },
    "probed": [
      "trait:charset.windows-1252"
    ]
  }
]
`
	if got := out.String(); got != want {
		t.Errorf("Run wrote\n%s\nwant\n%s", got, want)
	}
}

func TestRunLeavesNoTemporaryDirectory(t *testing.T) {
	p := placer(t, samples)
	if err := p.Run(context.Background(), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	left, err := os.ReadDir(os.Getenv("TMPDIR"))
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Errorf("Run left %v in its temporary directory", left)
	}
}

func TestRunRefusesASampleMissingFromTheManifest(t *testing.T) {
	var out bytes.Buffer
	err := placer(t, `{"path": "text/z.txt", "library": []}`).Run(context.Background(), &out)
	if err == nil || !strings.Contains(err.Error(), "text/z.txt: not in") {
		t.Errorf("Run = %v, want an error naming text/z.txt", err)
	}
	if out.Len() != 0 {
		t.Errorf("Run wrote %q despite its error", out.String())
	}
}

func TestRunStopsWhenItsContextEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	if err := placer(t, samples).Run(ctx, &out); !errors.Is(err, context.Canceled) {
		t.Errorf("Run = %v, want context.Canceled", err)
	}
	if out.Len() != 0 {
		t.Errorf("Run wrote %q despite its error", out.String())
	}
}

func TestRunWithoutAParserPanicsNamingIt(t *testing.T) {
	p := placer(t, samples)
	p.Parser = nil
	defer func() {
		if r := recover(); r != "placement: Placer.Run requires a Parser" {
			t.Errorf("recovered %v", r)
		}
	}()
	_ = p.Run(context.Background(), &bytes.Buffer{})
}
