package probe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	probeTimeout   = 600 * time.Second
	convertTimeout = 300 * time.Second
	versionTimeout = 60 * time.Second
	waitDelay      = 10 * time.Second
)

// ErrMissing is the cause of an adapter's refusal when its program is not on PATH.
var ErrMissing = errors.New("not on PATH")

// ErrTooOld is the cause of an adapter's refusal when its program is older than its minimum.
var ErrTooOld = errors.New("older than the minimum")

// ErrNoVersion is the cause of an adapter's refusal when its program's version command printed
// no version.
var ErrNoVersion = errors.New("no version in its output")

type spec struct {
	name    string
	args    []string
	pattern *regexp.Regexp
	minimum string
	fix     string
}

const miseFix = "run mise install"

type program struct {
	name    string
	path    string
	version string
}

func find(ctx context.Context, s spec) (program, error) {
	path, err := exec.LookPath(s.name)
	if err != nil {
		return program{}, fmt.Errorf("probe: %s: %w; %s", s.name, ErrMissing, s.fix)
	}
	ctx, cancel := context.WithTimeout(ctx, versionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, s.args...)
	cmd.WaitDelay = waitDelay
	out, err := cmd.CombinedOutput()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return program{}, fmt.Errorf("probe: %s: %w", s.name, ctxErr)
	}
	m := s.pattern.FindSubmatch(out)
	if m == nil {
		if err != nil {
			return program{}, fmt.Errorf("probe: %s: %w (%w); %s", s.name, ErrNoVersion, err, s.fix)
		}
		return program{}, fmt.Errorf("probe: %s: %w; %s", s.name, ErrNoVersion, s.fix)
	}
	found := string(m[1])
	if compareVersions(found, s.minimum) < 0 {
		return program{}, fmt.Errorf("probe: %s: version %s is %w %s; %s", s.name, found, ErrTooOld, s.minimum, s.fix)
	}
	return program{name: s.name, path: path, version: found}, nil
}

func compareVersions(a, b string) int {
	x, y := versionParts(a), versionParts(b)
	for i := range max(len(x), len(y)) {
		var m, n int
		if i < len(x) {
			m = x[i]
		}
		if i < len(y) {
			n = y[i]
		}
		if m != n {
			if m < n {
				return -1
			}
			return 1
		}
	}
	return 0
}

func versionParts(v string) []int {
	var parts []int
	for _, f := range strings.FieldsFunc(v, func(r rune) bool { return r == '.' || r == '-' }) {
		n, _ := strconv.Atoi(f)
		parts = append(parts, n)
	}
	return parts
}

func (p program) output(ctx context.Context, args []string, input []byte, timeout time.Duration) []byte {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.path, args...)
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.WaitDelay = waitDelay
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil
	}
	if _, ok := errors.AsType[*exec.Error](err); ok {
		return nil
	}
	if _, ok := errors.AsType[*os.PathError](err); ok {
		return nil
	}
	return out.Bytes()
}
