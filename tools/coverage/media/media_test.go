package media_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
	"github.com/standards-lab/docs-corpus/tools/coverage/media"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
)

type tool string

func (t tool) Output(context.Context, string) []byte {
	if t == "" {
		return nil
	}
	return []byte(t)
}

func lines(l ...string) tool {
	return tool(strings.Join(l, "\n") + "\n")
}

func output(kind, name, meta string) run.Event {
	return run.Event{Event: "output", Output: &run.Output{Kind: kind, Name: name, Meta: map[string]json.RawMessage{
		strings.Split(meta, "=")[0]: json.RawMessage(strings.Split(meta, "=")[1]),
	}}}
}

func TestPDFPlacesTextLayersScaleAndForms(t *testing.T) {
	pr := media.Prober{
		PDFInfo: lines("Title:          x", "Pages:          120"),
		PDFImages: lines(
			"page   num  type   width height color comp bpc  enc interp  object ID x-ppi y-ppi size ratio",
			"--------------------------------------------------------------------------------------------",
			"   1     0 image    1200   800  rgb     3   8  jpeg   no        12  0   300   300  100K 3.5%",
			"   2     1 image     100   100  rgb     3   8  jpeg   no        13  0    72    72    1K 3.5%"),
		PDFFonts: lines(
			"name                                 type              encoding         emb sub uni object ID",
			"------------------------------------ ----------------- ---------------- --- --- --- ---------",
			"ABCDEF+Arial                         TrueType          WinAnsi          yes yes no      12  0",
			"Helvetica                            Type 1            Standard         no  no  yes     13  0"),
		QPDF: lines("BT 3 Tr ET", "/AcroForm << >>", "/FT /Tx", "/FT /Sig"),
	}
	nested := output("text", "page 2", "page=2")
	depth := 1.0
	nested.Part.Depth = &depth
	outputs := []run.Event{
		output("text", "page 1", "page=1"),
		nested,
		output("image", "page 1", "dpi=150"),
		output("image", "page 2", "dpi=150"),
		output("image", "page 3", "dpi=300"),
	}
	traits, features, facts := pr.PDF(context.Background(), "/a.pdf", outputs)
	want := []string{"trait:scale.many-pages", "trait:scale.oversized-page", "trait:text-layer.ocr", "trait:text-layer.partial"}
	if got := traits.Sorted(""); !slices.Equal(got, want) {
		t.Errorf("traits = %v, want %v", got, want)
	}
	if got := features.Sorted(""); !slices.Equal(got, []string{"feature:forms", "feature:signature-fields"}) {
		t.Errorf("features = %v", got)
	}
	wantFacts := `{
  "pages": 120,
  "rendered_pages": 3,
  "text_pages": 1,
  "dpis": [
    150,
    300
  ],
  "pages_with_large_images": 1,
  "large_image_ppi": [
    300
  ],
  "fonts": 2,
  "fonts_without_unicode": 1,
  "invisible_text_runs": 1
}`
	if got := string(jsonvalue.Marshal(facts)); got != wantFacts {
		t.Errorf("facts =\n%s\nwant\n%s", got, wantFacts)
	}
}

func TestPDFReadsTheFileWhenQPDFWritesNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.7\n/Collection << /View /D >>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pr := media.Prober{PDFInfo: tool(""), PDFImages: tool(""), PDFFonts: tool(""), QPDF: tool("")}
	traits, features, _ := pr.PDF(context.Background(), path, nil)
	if len(traits) != 0 {
		t.Errorf("traits = %v, want none for a PDF with no pages", traits.Sorted(""))
	}
	if got := features.Sorted(""); !slices.Equal(got, []string{"feature:portfolio"}) {
		t.Errorf("features = %v, want portfolio", got)
	}
}

func TestPDFWithNoTextLayerIsNone(t *testing.T) {
	pr := media.Prober{PDFInfo: lines("Pages: 2"), PDFImages: tool(""), PDFFonts: tool(""), QPDF: tool("x")}
	traits, _, _ := pr.PDF(context.Background(), "/a.pdf", nil)
	if got := traits.Sorted(""); !slices.Equal(got, []string{"trait:text-layer.none"}) {
		t.Errorf("traits = %v, want text-layer.none", got)
	}
}

func TestAudioVideoPlacesStreamsAndDuration(t *testing.T) {
	for _, c := range []struct {
		name, probe     string
		traits, feature []string
		facts           string
	}{
		{
			"narrowband audio with cover art",
			`{"streams": [{"codec_type": "audio", "sample_rate": "8000", "channels": 1},
			              {"codec_type": "video", "disposition": {"attached_pic": 1}}],
			  "format": {"duration": "1800.04"}}`,
			[]string{"trait:audio.multi-window", "trait:audio.narrowband", "trait:scale.long-media"},
			[]string{"feature:attached-picture"},
			`"duration": 1800.0`,
		},
		{
			"silent video with two streams",
			`{"streams": [{"codec_type": "video"}, {"codec_type": "video", "disposition": {"attached_pic": 0}}, {"codec_type": "subtitle"}],
			  "format": {"duration": "12.5"}}`,
			[]string{"trait:audio.none"},
			[]string{"feature:multiple-streams"},
			`"subtitle_streams": 1`,
		},
		{
			"nothing ffprobe could read",
			``,
			nil, nil,
			`"sample_rates": [],`,
		},
	} {
		traits, features, facts := media.Prober{FFprobe: tool(c.probe)}.AudioVideo(context.Background(), "/a")
		if got := traits.Sorted(""); !slices.Equal(got, c.traits) {
			t.Errorf("%s: traits = %v, want %v", c.name, got, c.traits)
		}
		if got := features.Sorted(""); !slices.Equal(got, c.feature) {
			t.Errorf("%s: features = %v, want %v", c.name, got, c.feature)
		}
		if got := string(jsonvalue.Marshal(facts)); !strings.Contains(got, c.facts) {
			t.Errorf("%s: facts = %s, want %s in them", c.name, got, c.facts)
		}
	}
}

func TestImagePlacesFramesAlphaAndOrientation(t *testing.T) {
	for _, c := range []struct {
		name, identify string
		want           []string
	}{
		{"animated GIF with alpha", "GIF|10|20|Blend|Undefined|False\nGIF|10|20|Blend|Undefined|False\n",
			[]string{"feature:animation", "feature:transparency"}},
		{"two-page rotated TIFF", "TIFF|10|20|Undefined|RightTop|True\nTIFF|10|20|Undefined|RightTop|True\n",
			[]string{"feature:multi-page-image", "feature:orientation"}},
		{"opaque PNG", "PNG|10|20|True|TopLeft|True\n", []string{}},
		{"no frames", "", []string{}},
	} {
		_, features, facts := media.Prober{Magick: tool(c.identify)}.Image(context.Background(), "/a")
		if got := features.Sorted(""); !slices.Equal(got, c.want) {
			t.Errorf("%s: features = %v, want %v", c.name, got, c.want)
		}
		if !facts.Has("frames") {
			t.Errorf("%s: facts have no frames: %s", c.name, jsonvalue.Marshal(facts))
		}
	}
}

func TestArchiveEntriesLeaveOutDirectories(t *testing.T) {
	got := media.Prober{Bsdtar: lines("a/", "a/b.txt", "", "c.zip")}.ArchiveEntries(context.Background(), "/a.zip")
	if !slices.Equal(got, []string{"a/b.txt", "c.zip"}) {
		t.Errorf("ArchiveEntries = %q", got)
	}
}

func TestAMissingToolPanicsNamingIt(t *testing.T) {
	defer func() {
		if r := recover(); r != "media: Prober.PDF requires a PDFFonts" {
			t.Errorf("recovered %v", r)
		}
	}()
	media.Prober{PDFInfo: tool(""), PDFImages: tool(""), QPDF: tool("")}.PDF(context.Background(), "/a.pdf", nil)
}
