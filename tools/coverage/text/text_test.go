package text_test

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/text"
)

type guesser struct {
	out   string
	calls int
}

func (g *guesser) Guess(_ context.Context, _ []byte) []byte {
	g.calls++
	if g.out == "" {
		return nil
	}
	return []byte(g.out)
}

func TestSplitLinesSplitsWherePythonDoes(t *testing.T) {
	for _, c := range []struct {
		in   string
		want []string
	}{
		{"a\r\nb\rc\nd", []string{"a", "b", "c", "d"}},
		{"a\x1cb\x1dc\x1ed\x1fe", []string{"a", "b", "c", "d\x1fe"}},
		{"a b\u0085c\n", []string{"a", "b", "c"}},
		{"a\v\fb", []string{"a", "", "b"}},
		{"\n", []string{""}},
	} {
		if got := text.SplitLines(c.in); !slices.Equal(got, c.want) {
			t.Errorf("SplitLines(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	if got := text.SplitLines(""); len(got) != 0 {
		t.Errorf("SplitLines(\"\") = %q, want none", got)
	}
}

func TestWhitespaceIncludesPythonsSeparators(t *testing.T) {
	if !text.IsSpace('\x1f') || !text.IsSpace(' ') || text.IsSpace('a') {
		t.Error("IsSpace disagrees with str.isspace")
	}
	if got := text.Strip("\x1c a b\t\x1f"); got != "a b" {
		t.Errorf("Strip = %q", got)
	}
	if got := text.Fields(" 1  0\x1fimage "); !slices.Equal(got, []string{"1", "0", "image"}) {
		t.Errorf("Fields = %q", got)
	}
}

func TestDecodeOrUTF8FindsCodecsByPythonsAliases(t *testing.T) {
	for _, c := range []struct {
		data       string
		name, want string
	}{
		{"caf\xe9", "latin-1", "café"},
		{"caf\xe9", "ISO_8859-1:1987", "café"},
		{"\x93q\x94", "Windows-1252", "“q”"},
		{"\xc1", "koi8_r", "а"},
		{"caf\xc3\xa9", "UTF8", "café"},
		{"a\xe9", "us-ascii", "a�"},
		{"\x00h\x00i", "utf-16-be", "hi"},
	} {
		if got := text.DecodeOrUTF8([]byte(c.data), c.name); got != c.want {
			t.Errorf("DecodeOrUTF8(%q, %q) = %q; want %q", c.data, c.name, got, c.want)
		}
	}
	if got := text.DecodeOrUTF8([]byte("caf\xe9"), "no-such-codec"); got != "caf\xe9" {
		t.Errorf("DecodeOrUTF8 with an unknown codec = %q, want the bytes as they are", got)
	}
}

func TestCharsetReadsBOMAndUTF8WithoutGuessing(t *testing.T) {
	g := &guesser{out: "WINDOWS-1252\n"}
	for _, c := range []struct{ data, want string }{
		{"\xef\xbb\xbfhi", "bom"},
		{"\xff\xfeh\x00", "bom"},
		{"plain ascii", "utf-8"},
		{"caf\xc3\xa9", "utf-8"},
	} {
		trait, guess := text.Charset(context.Background(), []byte(c.data), g)
		if trait != "" || guess != c.want {
			t.Errorf("Charset(%q) = %q, %q; want no trait and %q", c.data, trait, guess, c.want)
		}
	}
	if g.calls != 0 {
		t.Errorf("the guesser ran %d times, want none", g.calls)
	}
}

func TestCharsetFindsUTF16WithoutABOM(t *testing.T) {
	g := &guesser{}
	trait, guess := text.Charset(context.Background(), []byte("\x00c\x00a\x00f\x00\xe9\x00 \x00w"), g)
	if trait != "trait:charset.utf-16-no-bom" || guess != "UTF-16-BE" {
		t.Errorf("Charset = %q, %q; want utf-16-no-bom, UTF-16-BE", trait, guess)
	}
	if g.calls != 0 {
		t.Errorf("the guesser ran %d times, want none", g.calls)
	}
}

func TestCharsetPlacesTheGuessersCharset(t *testing.T) {
	for _, c := range []struct {
		data, out, trait, guess string
	}{
		{"caf\xe9 au lait", "windows-1252\n", "trait:charset.windows-1252", "WINDOWS-1252"},
		{"caf\xe9 \x81", "windows-1252\n", "trait:charset.legacy-other", "WINDOWS-1252"},
		{"caf\xe9", "ISO-8859-2", "trait:charset.windows-1252", "ISO-8859-2"},
		{"\xa4\xa4\xa4\xe5", "BIG5\n", "trait:charset.legacy-other", "BIG5"},
		{"\xa4\xa4", "UTF-16LE", "trait:charset.utf-16-no-bom", "UTF-16LE"},
		{"\xa4\xa4", "", "trait:charset.legacy-other", ""},
	} {
		g := &guesser{out: c.out}
		trait, guess := text.Charset(context.Background(), []byte(c.data), g)
		if trait != c.trait || guess != c.guess || g.calls != 1 {
			t.Errorf("Charset(%q) with guess %q = %q, %q after %d guesses; want %q, %q after 1",
				c.data, c.out, trait, guess, g.calls, c.trait, c.guess)
		}
	}
}

func TestDecodeTextByBOMOrGuess(t *testing.T) {
	for _, c := range []struct{ data, guess, want string }{
		{"\xef\xbb\xbfhi", "bom", "hi"},
		{"\xff\xfeh\x00i\x00", "bom", "hi"},
		{"\x00h\x00i", "UTF-16-BE", "hi"},
		{"h\x00i\x00", "UTF-16-LE", "hi"},
		{"caf\xe9", "windows-1252", "café"},
		{"caf\xc3\xa9", "utf-8", "café"},
	} {
		if got := text.DecodeText([]byte(c.data), c.guess); got != c.want {
			t.Errorf("DecodeText(%q, %q) = %q, want %q", c.data, c.guess, got, c.want)
		}
	}
}

func TestUTF8NameAcceptsUTF8AndASCIIOnly(t *testing.T) {
	for name, want := range map[string]bool{"utf-8": true, "us-ascii": true, "": true, "iso-8859-1": false, "UTF-8": false} {
		if got := text.UTF8Name(name); got != want {
			t.Errorf("UTF8Name(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestScriptsPlaceAScriptAtItsThreshold(t *testing.T) {
	for _, c := range []struct {
		name string
		in   string
		want []string
	}{
		{"19 Cyrillic letters", strings.Repeat("ж", 19), nil},
		{"20 Cyrillic letters", strings.Repeat("ж", 20), []string{"trait:script.cyrillic"}},
		{"Latin alone", strings.Repeat("a", 100), nil},
		{"Han and kana are one system", strings.Repeat("漢", 10) + strings.Repeat("か", 10),
			[]string{"trait:script.han", "trait:script.kana"}},
		{"two systems", strings.Repeat("a", 60) + strings.Repeat("ж", 20),
			[]string{"trait:script.cyrillic", "trait:script.mixed"}},
		{"the Kelvin sign counts for no script", strings.Repeat("K", 100), nil},
	} {
		traits, _ := text.Scripts(c.in)
		got := slices.Sorted(maps.Keys(traits))
		if !slices.Equal(got, c.want) {
			t.Errorf("%s: Scripts = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestScriptsCountInFirstAppearanceOrder(t *testing.T) {
	_, counts := text.Scripts("жa ж1 aaカ")
	want := []text.ScriptCount{{Name: "cyrillic", Count: 2}, {Name: "latin", Count: 3}, {Name: "kana", Count: 1}}
	if !slices.Equal(counts, want) {
		t.Errorf("counts = %+v, want %+v", counts, want)
	}
}
