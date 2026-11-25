package manifest

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
)

// Version is the manifest schema's version, written as Manifest.Version.
const Version = 1

// Manifest is MANIFEST.json, as README.md states it.
type Manifest struct {
	// Version is the schema's version, Version.
	Version int `json:"version"`
	// Files are the corpus files, sorted by path.
	Files []File `json:"files"`
}

// File is one corpus file's entry. A sample carries Covers, Truth, and Docs, and TruthNote when
// it has no truth; a truth or doc file carries none of them. A generated file carries Command,
// Inputs, Generator, and GeneratorLicense, and a candidate its Gap. Every other field is absent
// where its source records nothing.
type File struct {
	// Path is its path from the repository root.
	Path string `json:"path"`
	// Role is sample, truth, or doc.
	Role string `json:"role"`
	// Sha256 is its content's SHA-256, in hex.
	Sha256 string `json:"sha256"`
	// Bytes is its size.
	Bytes int64 `json:"bytes"`
	// Format is its media type, as file -b --mime-type reports it.
	Format string `json:"format"`
	// Covers are a sample's library, trait, and feature IDs, sorted.
	Covers []string `json:"covers,omitzero"`
	// Truth are a sample's truth files.
	Truth []Truth `json:"truth,omitzero"`
	// TruthNote is why a sample has no truth.
	TruthNote string `json:"truth_note,omitempty"`
	// Docs are a sample's doc files' paths.
	Docs []string `json:"docs,omitzero"`
	// Source names the collection it comes from.
	Source string `json:"source,omitempty"`
	// URL is where it was fetched.
	URL string `json:"url,omitempty"`
	// Ref is the commit or revision URL is pinned to.
	Ref string `json:"ref,omitempty"`
	// License is its license.
	License string `json:"license"`
	// LicenseURL is where the license was read.
	LicenseURL string `json:"license_url,omitempty"`
	// LicenseRead is the date the license was read, YYYY-MM-DD.
	LicenseRead string `json:"license_read,omitempty"`
	// LicenseStatus is how its license is known, such as verified.
	LicenseStatus string `json:"license_status"`
	// Redistributable is whether it may be redistributed.
	Redistributable bool `json:"redistributable"`
	// DerivedFrom is the file it was cut or converted from.
	DerivedFrom string `json:"derived_from,omitempty"`
	// Command is the command that makes a generated file, run from the repository root.
	Command string `json:"command,omitempty"`
	// Inputs are the samples Command reads.
	Inputs []string `json:"inputs,omitzero"`
	// Generator is the generating tool, with its version.
	Generator string `json:"generator,omitempty"`
	// GeneratorLicense is the generator's license.
	GeneratorLicense string `json:"generator_license,omitempty"`
	// Gap is the id of the gap a candidate fills.
	Gap string `json:"gap,omitempty"`
	// Note is what its source records about it.
	Note string `json:"note,omitempty"`
}

// Truth is one of a sample's truth files.
type Truth struct {
	// Path is the truth file's path.
	Path string `json:"path"`
	// Note is the reference's qualifier on it, such as the span it covers.
	Note string `json:"note,omitempty"`
}

// Reference is the reference docs' MANIFEST.json, ~/docs', in the fields Build reads.
type Reference struct {
	// Files are its entries.
	Files []ReferenceFile `json:"files"`
}

// ReferenceFile is one entry of the reference manifest.
type ReferenceFile struct {
	// Path is its path under the reference docs.
	Path string `json:"path"`
	// Sha256 is its content's SHA-256, in hex.
	Sha256 string `json:"sha256"`
	// Bytes is its size.
	Bytes int64 `json:"bytes"`
	// Format is its media type.
	Format string `json:"format"`
	// Source names the collection it comes from.
	Source string `json:"source"`
	// SourceURL is where it was fetched, becoming File.URL.
	SourceURL string `json:"source_url"`
	// License is its license.
	License string `json:"license"`
	// LicenseStatus is how its license is known.
	LicenseStatus string `json:"license_status"`
	// Redistributable is whether it may be redistributed.
	Redistributable bool `json:"redistributable"`
	// Role is sample, truth, or doc in the reference.
	Role string `json:"role"`
	// Truth is free text: a path or a folder, either with a parenthesized qualifier, clauses
	// joined by "; ", or "none: " and the reason there is no truth.
	Truth string `json:"truth"`
	// DerivedFrom is the file it was cut or converted from.
	DerivedFrom string `json:"derived_from"`
	// Note is what the reference records about it.
	Note string `json:"note"`
}

// Reading is a sample's license, read at its source, which makes its status verified.
type Reading struct {
	// Path is the sample's path.
	Path string `json:"path"`
	// LicenseURL is where the license was read.
	LicenseURL string `json:"license_url"`
	// LicenseRead is the date it was read, YYYY-MM-DD.
	LicenseRead string `json:"license_read"`
}

// Build returns the manifest of coverage's adopted samples and the truth and doc files they link,
// each with the reference's facts and provenance, the sample's covers, and its readings applied.
// A file the reference lacks, a sample whose sha256 differs from the reference's, a path that is
// both a sample and a truth or doc file, or a reading for a path that is not an adopted sample is
// an error.
func Build(coverage disposition.Coverage, reference Reference, readings []Reading) (Manifest, error) {
	byPath := make(map[string]ReferenceFile, len(reference.Files))
	for _, f := range reference.Files {
		byPath[f.Path] = f
	}
	readingFor := make(map[string]Reading, len(readings))
	for _, r := range readings {
		readingFor[r.Path] = r
	}
	files := map[string]File{}
	linked := map[string]string{}
	var errs []error
	for _, s := range coverage.Samples {
		if s.Disposition != "adopt" {
			continue
		}
		ref, ok := byPath[s.Path]
		if !ok {
			errs = append(errs, fmt.Errorf("%s: an adopted sample the reference manifest lacks", s.Path))
			continue
		}
		if ref.Sha256 != s.Sha256 {
			errs = append(errs, fmt.Errorf("%s: sha256 %s differs from the reference's %s", s.Path, s.Sha256, ref.Sha256))
			continue
		}
		f := entry(ref, "sample")
		f.Covers = covers(s)
		f.Truth, f.TruthNote = truth(ref.Truth, s.Truth)
		f.Docs = append([]string{}, s.Docs...)
		if r, ok := readingFor[s.Path]; ok {
			f.LicenseURL, f.LicenseRead, f.LicenseStatus = r.LicenseURL, r.LicenseRead, "verified"
			delete(readingFor, s.Path)
		}
		files[s.Path] = f
		for _, p := range s.Truth {
			linked[p] = "truth"
		}
		for _, p := range s.Docs {
			if linked[p] == "truth" {
				errs = append(errs, fmt.Errorf("%s: linked as both truth and doc", p))
			}
			linked[p] = "doc"
		}
	}
	for path := range readingFor {
		errs = append(errs, fmt.Errorf("%s: a license reading for a path that is not an adopted sample", path))
	}
	for path, role := range linked {
		if _, ok := files[path]; ok {
			errs = append(errs, fmt.Errorf("%s: both an adopted sample and a %s file", path, role))
			continue
		}
		ref, ok := byPath[path]
		if !ok {
			errs = append(errs, fmt.Errorf("%s: a %s file the reference manifest lacks", path, role))
			continue
		}
		files[path] = entry(ref, role)
	}
	if err := errors.Join(errs...); err != nil {
		return Manifest{}, err
	}
	m := Manifest{Version: Version, Files: make([]File, 0, len(files))}
	for _, f := range files {
		m.Files = append(m.Files, f)
	}
	slices.SortFunc(m.Files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return m, nil
}

func entry(ref ReferenceFile, role string) File {
	return File{
		Path:            ref.Path,
		Role:            role,
		Sha256:          ref.Sha256,
		Bytes:           ref.Bytes,
		Format:          ref.Format,
		Source:          ref.Source,
		URL:             ref.SourceURL,
		License:         ref.License,
		LicenseStatus:   ref.LicenseStatus,
		Redistributable: ref.Redistributable,
		DerivedFrom:     ref.DerivedFrom,
		Note:            ref.Note,
	}
}

func covers(s disposition.Sample) []string {
	ids := slices.Concat(s.Library, s.Traits, s.Features)
	slices.Sort(ids)
	return slices.Compact(ids)
}

func truth(text string, paths []string) ([]Truth, string) {
	if reason, ok := strings.CutPrefix(text, "none:"); ok {
		return []Truth{}, strings.TrimSpace(reason)
	}
	clauses := splitTopLevel(text)
	out := make([]Truth, 0, len(paths))
	for _, p := range paths {
		out = append(out, Truth{Path: p, Note: note(clauses, p)})
	}
	return out, ""
}

func splitTopLevel(text string) []string {
	var clauses []string
	depth, start := 0, 0
	for i := 0; i < len(text); i++ {
		switch {
		case text[i] == '(':
			depth++
		case text[i] == ')' && depth > 0:
			depth--
		case depth == 0 && strings.HasPrefix(text[i:], "; "):
			clauses = append(clauses, text[start:i])
			start = i + 2
		}
	}
	return append(clauses, text[start:])
}

func note(clauses []string, path string) string {
	names := []string{path}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			names = append(names, path[:i+1])
		}
	}
	for _, name := range names {
		for _, clause := range clauses {
			i := wordIndex(clause, name)
			if i < 0 {
				continue
			}
			before := strings.TrimSpace(clause[:i])
			after := strings.TrimSpace(clause[i+len(name):])
			switch {
			case before == "" && after == "":
				return ""
			case before == "" && enclosed(after):
				return after[1 : len(after)-1]
			default:
				return clause
			}
		}
	}
	return ""
}

func wordIndex(s, word string) int {
	for from := 0; ; {
		i := strings.Index(s[from:], word)
		if i < 0 {
			return -1
		}
		i += from
		end := i + len(word)
		if (i == 0 || s[i-1] == ' ') && (end == len(s) || s[end] == ' ') {
			return i
		}
		from = i + 1
	}
}

func enclosed(s string) bool {
	if len(s) < 2 || s[0] != '(' || s[len(s)-1] != ')' {
		return false
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 && i != len(s)-1 {
				return false
			}
		}
	}
	return true
}
