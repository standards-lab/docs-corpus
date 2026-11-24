package probe_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/standards-lab/docs-corpus/tools/coverage/probe"
)

const standin = "PROBE_STANDIN"

var versions = map[string]string{
	"pdfinfo":   "pdfinfo version " + probe.PopplerMinimum,
	"pdfimages": "pdfimages version " + probe.PopplerMinimum,
	"pdffonts":  "pdffonts version " + probe.PopplerMinimum,
	"qpdf":      "qpdf version " + probe.QPDFMinimum,
	"ffprobe":   "ffprobe version " + probe.FFmpegMinimum + " Copyright (c) 2007-2026 the FFmpeg developers",
	"magick":    "Version: ImageMagick " + probe.ImageMagickMinimum + " Q16-HDRI x86_64",
	"uchardet":  "\nuchardet Command Line Tool\nVersion " + probe.UchardetMinimum,
	"bsdtar":    "bsdtar " + probe.LibarchiveMinimum + " - libarchive " + probe.LibarchiveMinimum,
	"soffice":   "LibreOffice 26.8.1.1 680(Build:1)",
}

var versionArgs = map[string]bool{"-v": true, "--version": true, "-version": true}

func TestMain(m *testing.M) {
	if os.Getenv(standin) != "" {
		os.Exit(act(filepath.Base(os.Args[0]), os.Args[1:]))
	}
	os.Exit(m.Run())
}

func act(name string, args []string) int {
	if len(args) == 1 && versionArgs[args[0]] {
		if v, ok := os.LookupEnv("STANDIN_VERSION_" + name); ok {
			fmt.Fprintln(os.Stderr, v)
		} else {
			fmt.Fprintln(os.Stderr, versions[name])
		}
		return 0
	}
	if log := os.Getenv("STANDIN_LOG"); log != "" {
		f, err := os.OpenFile(log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return 2
		}
		_, err = fmt.Fprintln(f, strings.Join(append([]string{name}, args...), " "))
		if errors.Join(err, f.Close()) != nil {
			return 2
		}
	}
	if d, err := time.ParseDuration(os.Getenv("STANDIN_SLEEP")); err == nil {
		time.Sleep(d)
	}
	for _, a := range args {
		fmt.Println(a)
	}
	if !strings.Contains(os.Getenv("STANDIN_STDIN"), name) {
		return exitCode()
	}
	in, err := io.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		return 2
	}
	fmt.Printf("stdin:%s\n", in)
	return exitCode()
}

func exitCode() int {
	var code int
	if _, err := fmt.Sscan(os.Getenv("STANDIN_EXIT"), &code); err != nil {
		return 0
	}
	return code
}

func link(t *testing.T, names ...string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, name := range names {
		if err := os.Symlink(exe, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
	t.Setenv(standin, "1")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
}

func allTools() []string {
	var names []string
	for name := range versions {
		names = append(names, name)
	}
	return names
}

func TestNewBuildsEveryAdapterAtItsMinimum(t *testing.T) {
	link(t, allTools()...)
	tools, err := probe.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tools.PDFInfo == nil || tools.Soffice == nil || tools.Uchardet == nil {
		t.Errorf("New left an adapter nil: %+v", tools)
	}
}

func TestNewNamesEveryMissingToolAndItsFix(t *testing.T) {
	link(t, "pdfinfo")
	_, err := probe.New(context.Background())
	if !errors.Is(err, probe.ErrMissing) {
		t.Fatalf("New = %v, want ErrMissing", err)
	}
	for _, name := range allTools() {
		if name == "pdfinfo" {
			if strings.Contains(err.Error(), "probe: pdfinfo:") {
				t.Errorf("New refused pdfinfo, which is on PATH: %v", err)
			}
			continue
		}
		if !strings.Contains(err.Error(), "probe: "+name+": not on PATH") {
			t.Errorf("New's error doesn't name %s: %v", name, err)
		}
	}
	if !strings.Contains(err.Error(), "qpdf: not on PATH; run mise install") {
		t.Errorf("a pinned tool's refusal doesn't name mise install: %v", err)
	}
	if !strings.Contains(err.Error(), "soffice: not on PATH; install LibreOffice 24.2 or later on the host") {
		t.Errorf("soffice's refusal doesn't name the host requirement: %v", err)
	}
}

func TestAdaptersCheckTheVersionAsMinimum(t *testing.T) {
	for _, c := range []struct {
		name, version string
		build         func(context.Context) error
		want          error
	}{
		{"pdfinfo", "pdfinfo version 26.10.0", newErr(probe.NewPDFInfo), probe.ErrTooOld},
		{"pdfinfo", "pdfinfo version 27.0", newErr(probe.NewPDFInfo), nil},
		{"ffprobe", "ffprobe version 9.1 Copyright", newErr(probe.NewFFprobe), nil},
		{"ffprobe", "ffprobe version n9.0.2-1", newErr(probe.NewFFprobe), probe.ErrTooOld},
		{"magick", "Version: ImageMagick 7.1.2-34 Q16", newErr(probe.NewMagick), probe.ErrTooOld},
		{"magick", "Version: ImageMagick 7.1.10-1 Q16", newErr(probe.NewMagick), nil},
		{"bsdtar", "bsdtar 3.8.10 - libarchive 3.8.10", newErr(probe.NewBsdtar), probe.ErrTooOld},
		{"soffice", "LibreOffice 7.6.4.1 60(Build:1)", newErr(probe.NewSoffice), probe.ErrTooOld},
		{"soffice", "LibreOffice 24.2 40(Build:2)", newErr(probe.NewSoffice), nil},
		{"qpdf", "qpdf, a build of no version", newErr(probe.NewQPDF), probe.ErrNoVersion},
		{"uchardet", "uchardet Command Line Tool", newErr(probe.NewUchardet), probe.ErrNoVersion},
	} {
		t.Run(c.name+" "+c.version, func(t *testing.T) {
			link(t, c.name)
			t.Setenv("STANDIN_VERSION_"+c.name, c.version)
			err := c.build(context.Background())
			if !errors.Is(err, c.want) || (c.want == nil) != (err == nil) {
				t.Fatalf("build = %v, want %v", err, c.want)
			}
			if err != nil && !strings.Contains(err.Error(), "probe: "+c.name+":") {
				t.Errorf("the refusal doesn't name %s: %v", c.name, err)
			}
		})
	}
}

func newErr[T any](build func(context.Context) (T, error)) func(context.Context) error {
	return func(ctx context.Context) error {
		_, err := build(ctx)
		return err
	}
}

func TestOutputRunsTheToolWithItsArguments(t *testing.T) {
	link(t, allTools()...)
	tools, err := probe.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for name, c := range map[string]struct {
		got  []byte
		want string
	}{
		"pdfinfo":   {tools.PDFInfo.Output(ctx, "/a.pdf"), "/a.pdf\n"},
		"pdfimages": {tools.PDFImages.Output(ctx, "/a.pdf"), "-list\n/a.pdf\n"},
		"pdffonts":  {tools.PDFFonts.Output(ctx, "/a.pdf"), "/a.pdf\n"},
		"qpdf":      {tools.QPDF.Output(ctx, "/a.pdf"), "--qdf\n--object-streams=disable\n--decrypt\n/a.pdf\n-\n"},
		"ffprobe":   {tools.FFprobe.Output(ctx, "/a.wav"), "-v\nerror\n-show_streams\n-show_format\n-of\njson\n/a.wav\n"},
		"magick":    {tools.Magick.Output(ctx, "/a.gif"), "identify\n-format\n%m|%w|%h|%A|%[orientation]|%[opaque]\n\n/a.gif\n"},
		"bsdtar":    {tools.Bsdtar.Output(ctx, "/a.zip"), "-tf\n/a.zip\n"},
	} {
		if string(c.got) != c.want {
			t.Errorf("%s ran with %q, want %q", name, c.got, c.want)
		}
	}
}

func TestGuessPassesTheDataOnStandardInput(t *testing.T) {
	link(t, "uchardet")
	t.Setenv("STANDIN_STDIN", "uchardet")
	u, err := probe.NewUchardet(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := string(u.Guess(context.Background(), []byte("caf\xe9"))); got != "stdin:caf\xe9\n" {
		t.Errorf("Guess ran with %q", got)
	}
}

func TestConvertRunsSofficeHeadlessWithItsProfile(t *testing.T) {
	link(t, "soffice")
	log := filepath.Join(t.TempDir(), "runs")
	t.Setenv("STANDIN_LOG", log)
	s, err := probe.NewSoffice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	s.Convert(context.Background(), "/in/a.doc", "docx", "/tmp/x/1", "/tmp/x/profile")
	got, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	want := "soffice -env:UserInstallation=file:///tmp/x/profile --headless --convert-to docx --outdir /tmp/x/1 /in/a.doc\n"
	if string(got) != want {
		t.Errorf("soffice ran as %q, want %q", got, want)
	}
}

func TestOutputKeepsTheStandardOutputOfAFailedRun(t *testing.T) {
	link(t, "pdfinfo")
	t.Setenv("STANDIN_EXIT", "3")
	p, err := probe.NewPDFInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := string(p.Output(context.Background(), "/a.pdf")); got != "/a.pdf\n" {
		t.Errorf("Output of a run that exited 3 = %q, want its standard output", got)
	}
}

func TestCancellingTheContextStopsTheTool(t *testing.T) {
	link(t, "bsdtar")
	b, err := probe.NewBsdtar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("STANDIN_SLEEP", "30s")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	if got := b.Output(ctx, "/a.zip"); got != nil {
		t.Errorf("Output under an ended context = %q, want nil", got)
	}
	if elapsed := time.Since(start); elapsed > 15*time.Second {
		t.Errorf("Output returned %v after its context ended", elapsed)
	}
}

func TestMinimumsAreTheVersionsMisePins(t *testing.T) {
	path := filepath.Join("..", "..", "..", "mise.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("%s is absent, as it is outside the repository: %v", path, err)
	}
	for key, minimum := range map[string]string{
		`"conda:poppler"`:                  probe.PopplerMinimum,
		`"conda:qpdf"`:                     probe.QPDFMinimum,
		`"conda:ffmpeg"`:                   probe.FFmpegMinimum,
		`"github:ImageMagick/ImageMagick"`: probe.ImageMagickMinimum,
		`"conda:uchardet"`:                 probe.UchardetMinimum,
		`"conda:libarchive"`:               probe.LibarchiveMinimum,
	} {
		m := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + ` = (?:\{ version = )?"([^"]+)"`).FindSubmatch(data)
		if m == nil {
			t.Errorf("%s pins no %s", path, key)
			continue
		}
		if string(m[1]) != minimum {
			t.Errorf("%s pins %s at %s, but its adapter's minimum is %s", path, key, m[1], minimum)
		}
	}
}
