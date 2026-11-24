package office_test

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/internal/docstest"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
	"github.com/standards-lab/docs-corpus/tools/coverage/office"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
)

func TestExtensionsClassifyAsOOXMLOrLegacy(t *testing.T) {
	for ext, want := range map[string][2]bool{
		".docx": {true, false}, ".pptm": {true, false}, ".xltx": {true, false},
		".doc": {false, true}, ".rtf": {false, true}, ".wpd": {false, true}, ".xls": {false, true}, ".ppt": {false, true},
		".odt": {false, false}, ".pdf": {false, false}, "": {false, false},
	} {
		if got := [2]bool{office.IsOOXML(ext), office.IsLegacy(ext)}; got != want {
			t.Errorf("%q: IsOOXML, IsLegacy = %v, want %v", ext, got, want)
		}
	}
}

func TestOOXMLOfAFileThatIsNotAZipHasNoFeatures(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.docx")
	if err := os.WriteFile(path, []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	features, facts := office.OOXML(path)
	if len(features) != 0 {
		t.Errorf("features = %v, want none", features.Sorted(""))
	}
	if got := string(jsonvalue.Marshal(facts)); got != "{\n  \"parts\": 0,\n  \"tables\": 0,\n  \"media\": 0\n}" {
		t.Errorf("facts = %s", got)
	}
}

func TestRealOOXMLPlacesItsRecordedFeatures(t *testing.T) {
	checked := 0
	for _, s := range docstest.Samples(t) {
		if !office.IsOOXML(strings.ToLower(filepath.Ext(s.Path))) {
			continue
		}
		features, _ := office.OOXML(docstest.Path(t, s.Path))
		if got, want := features.Sorted(""), s.Probed(); !slices.Equal(got, want) {
			t.Errorf("%s: features = %v, want %v", s.Path, got, want)
		}
		checked++
	}
	if checked == 0 {
		t.Error("docs/coverage.json lists no OOXML sample")
	}
}

func TestODFTextIsTheContentWithoutMarkup(t *testing.T) {
	got := office.ODFText(docstest.Path(t, "office/odf/writer-odf13.odt"))
	if strings.TrimSpace(got) == "" || strings.ContainsAny(got, "<>") {
		t.Errorf("ODFText = %.200q, want text with no markup", got)
	}
	if got := office.ODFText(filepath.Join(t.TempDir(), "missing.odt")); got != "" {
		t.Errorf("ODFText of a missing file = %q", got)
	}
}

type soffice struct {
	calls [][]string
	write bool
}

func (s *soffice) Convert(_ context.Context, path, format, outDir, profile string) {
	s.calls = append(s.calls, []string{path, format, outDir, profile})
	if s.write {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + "." + format
		_ = os.WriteFile(filepath.Join(outDir, name), nil, 0o644)
	}
}

func TestConversionsNumberEachDocumentsDirectory(t *testing.T) {
	dir := t.TempDir()
	s := &soffice{write: true}
	c := &office.Conversions{Soffice: s, Dir: dir}
	ctx := context.Background()
	if got := c.Convert(ctx, "/in/a.DOC"); got != filepath.Join(dir, "1", "a.docx") {
		t.Errorf("first conversion = %q", got)
	}
	if got := c.Convert(ctx, "/in/notes.txt"); got != "" {
		t.Errorf("a document that isn't legacy converts to %q, want none", got)
	}
	if got := c.Convert(ctx, "/in/b.xls"); got != filepath.Join(dir, "3", "b.xlsx") {
		t.Errorf("third conversion = %q, want its own numbered directory", got)
	}
	want := [][]string{
		{"/in/a.DOC", "docx", filepath.Join(dir, "1"), dir + "/profile"},
		{"/in/b.xls", "xlsx", filepath.Join(dir, "3"), dir + "/profile"},
	}
	if !slices.EqualFunc(s.calls, want, slices.Equal) {
		t.Errorf("soffice ran %q, want %q", s.calls, want)
	}
}

func TestConversionThatWritesNothingHasNoPath(t *testing.T) {
	c := &office.Conversions{Soffice: &soffice{}, Dir: t.TempDir()}
	if got := c.Convert(context.Background(), "/in/a.ppt"); got != "" {
		t.Errorf("Convert = %q, want none", got)
	}
}

func TestConversionsWithoutSofficePanicsNamingIt(t *testing.T) {
	defer func() {
		if r := recover(); r != "office: Conversions.Convert requires a Soffice" {
			t.Errorf("recovered %v", r)
		}
	}()
	(&office.Conversions{Dir: t.TempDir()}).Convert(context.Background(), "/in/a.doc")
}

func onenoteRows(t *testing.T, rows map[string]string) []run.Row {
	t.Helper()
	var out []run.Row
	for _, name := range slices.Sorted(maps.Keys(rows)) {
		v, err := jsonvalue.Parse([]byte(rows[name]))
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, run.Row{Name: name, Fields: v.(*jsonvalue.Object)})
	}
	return out
}

func TestOneNotePlacesItsRecordsPages(t *testing.T) {
	traits, features, facts := office.OneNote(onenoteRows(t, map[string]string{
		"page 1":    `{"markdown": "| a | b |\n- [ ] todo\n[l](https://example.org)\n$$x^2$$", "ink_alternatives": ["x"], "revisions": [1]}`,
		"page 2":    `{"markdown": "", "media": [{"kind": "image", "ocr": "text"}, {"kind": "ink"}, {"kind": "embedded_file"}]}`,
		"section 1": `{"path": ["Notebook", "Group"]}`,
	}), idset.New(`"encrypted"`))
	if got := traits.Sorted(""); !slices.Equal(got, []string{"trait:scan.handwriting"}) {
		t.Errorf("traits = %v", got)
	}
	want := []string{"feature:embedded-objects", "feature:equations", "feature:hyperlinks", "feature:image-ocr",
		"feature:ink", "feature:note-tags", "feature:page-revisions", "feature:pictures", "feature:protected-section",
		"feature:section-groups", "feature:tables"}
	if got := features.Sorted(""); !slices.Equal(got, want) {
		t.Errorf("features = %v, want %v", got, want)
	}
	if got := string(jsonvalue.Marshal(facts)); got != "{\n  \"pages\": 2,\n  \"sections\": 1\n}" {
		t.Errorf("facts = %s", got)
	}
}

func TestOneNoteWithAPlainSectionHasNoFeatures(t *testing.T) {
	_, features, _ := office.OneNote(onenoteRows(t, map[string]string{
		"page 1":    `{"markdown": "just text | not a row"}`,
		"section 1": `{"path": "S"}`,
	}), idset.Set{})
	if len(features) != 0 {
		t.Errorf("features = %v, want none", features.Sorted(""))
	}
}
