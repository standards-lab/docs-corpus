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
// Inputs, Generator, and GeneratorLicense, and a candidate its Gap. A file a measurement reads
// carries Measures. Every other field is absent where its source records nothing.
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
	// Inputs are the adopted samples and sourced candidates Command reads.
	Inputs []string `json:"inputs,omitzero"`
	// Generator is the generating tool, with its version.
	Generator string `json:"generator,omitempty"`
	// GeneratorLicense is the generator's license.
	GeneratorLicense string `json:"generator_license,omitempty"`
	// Gap is the id of the gap a candidate fills.
	Gap string `json:"gap,omitempty"`
	// Measures are the ids of the measurements in docs/measurements.json that read it, sorted.
	Measures []string `json:"measures,omitempty"`
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

// Reading is a sample's, truth file's, or doc file's license, read at its source, which makes its
// status verified.
type Reading struct {
	// Path is the file's path.
	Path string `json:"path"`
	// License is the license read, when it differs from the reference's.
	License string `json:"license,omitempty"`
	// LicenseURL is where the license was read.
	LicenseURL string `json:"license_url"`
	// LicenseRead is the date it was read, YYYY-MM-DD.
	LicenseRead string `json:"license_read"`
}

// Candidate is a gap candidate's file as docs/candidates.json records it once fetched or
// generated.
type Candidate struct {
	// Gap is the id of the gap it fills.
	Gap string `json:"gap"`
	// Path is its path from the repository root.
	Path string `json:"path"`
	// Source names the collection it comes from.
	Source string `json:"source"`
	// Sha256 is its content's SHA-256, in hex.
	Sha256 string `json:"sha256"`
	// Bytes is its size.
	Bytes int64 `json:"bytes"`
	// Format is its media type, as file -b --mime-type reports it.
	Format string `json:"format"`
	// FormatBy is the file version that reported Format.
	FormatBy string `json:"format_by"`
	// TruthFrom, for a generated candidate, is the input whose truth it inherits unchanged.
	TruthFrom string `json:"truth_from,omitempty"`
}

// Sourced is a measurement input sourced outside the reference docs, as docs/sourced.json records
// it: the file a measurement reads, with its provenance, its license read at its source, and its
// truth files.
type Sourced struct {
	// Path is its path from the repository root.
	Path string `json:"path"`
	// Source names the collection or publisher it comes from.
	Source string `json:"source"`
	// URL is where it was fetched.
	URL string `json:"url"`
	// License is its license, as read at its source.
	License string `json:"license"`
	// LicenseURL is where the license was read.
	LicenseURL string `json:"license_url"`
	// LicenseRead is the date the license was read, YYYY-MM-DD.
	LicenseRead string `json:"license_read"`
	// Format is its media type, as file -b --mime-type reports it.
	Format string `json:"format"`
	// Bytes is its size.
	Bytes int64 `json:"bytes"`
	// Sha256 is its content's SHA-256, in hex.
	Sha256 string `json:"sha256"`
	// Truth are its truth files.
	Truth []SourcedTruth `json:"truth"`
}

// SourcedTruth is one of a sourced input's truth files, with its own facts and license reading.
type SourcedTruth struct {
	// Path is its path from the repository root.
	Path string `json:"path"`
	// Sha256 is its content's SHA-256, in hex.
	Sha256 string `json:"sha256"`
	// Bytes is its size.
	Bytes int64 `json:"bytes"`
	// Format is its media type, as file -b --mime-type reports it.
	Format string `json:"format"`
	// License is its license, as read at its source.
	License string `json:"license"`
	// LicenseURL is where the license was read.
	LicenseURL string `json:"license_url"`
	// LicenseRead is the date the license was read, YYYY-MM-DD.
	LicenseRead string `json:"license_read"`
}

// Placed is one sample's covers, an element of place-traits --corpus's output.
type Placed struct {
	// Path is the sample's path.
	Path string `json:"path"`
	// Covers are its library, trait, and feature IDs.
	Covers []string `json:"covers"`
}

// Build returns the manifest of coverage's adopted and measured samples and the truth and doc files
// they link, each with the reference's facts and provenance, a sample's covers, and its reading
// applied, the reading's license replacing the reference's when it names one, and of the candidates
// recorded for coverage's gaps, but for those whose gap is covered_by another sample. A candidate
// carries its recorded facts, its gap's candidate's provenance, and its covers in placed, or none
// when placed lacks it. A sourced candidate has no truth. A generated one, whose inputs are adopted
// samples or sourced candidates, has its inputs' distinct licenses, in input order joined by "; ",
// a license status that is verified when every input's is, otherwise the first input's that isn't,
// is redistributable when every input is, and has the truth of the input its TruthFrom names, or
// none. Each sourced input in sourced is a sample with its recorded facts and provenance, its
// covers in placed, or none when placed lacks it, a verified license, redistributable, and its
// truth files, each a truth file with its own facts and reading. Each file a measurement in
// measurements lists in its inputs names that measurement's id in its Measures.
//
// A file the reference lacks, a sample whose sha256 differs from the reference's, a path that is
// both a sample and a truth or doc file, a reading for a path that is neither an adopted or
// measured sample nor a file one links, a sourced or generated candidate with no recorded file
// whose gap isn't covered_by another sample, a recorded file for a gap that has no sourced or
// generated candidate, a candidate at another file's path, a generated candidate whose input is
// neither an adopted sample nor a sourced candidate it lists, a TruthFrom on a sourced candidate or
// naming no input, a sourced input or one of its truth files at another file's path, a sourced
// input no measurement lists, and a measurement input that is no file of the manifest are errors.
func Build(coverage disposition.Coverage, reference Reference, readings []Reading, candidates []Candidate, placed []Placed, measurements []disposition.Measurement, sourced []Sourced) (Manifest, error) {
	byPath := make(map[string]ReferenceFile, len(reference.Files))
	for _, f := range reference.Files {
		byPath[f.Path] = f
	}
	files := map[string]File{}
	linked := map[string]string{}
	var errs []error
	for _, s := range coverage.Samples {
		kind, ok := inCorpus[s.Disposition]
		if !ok {
			continue
		}
		ref, ok := byPath[s.Path]
		if !ok {
			errs = append(errs, fmt.Errorf("%s: %s sample the reference manifest lacks", s.Path, kind))
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
	for path, role := range linked {
		if _, ok := files[path]; ok {
			errs = append(errs, fmt.Errorf("%s: both a sample and a %s file", path, role))
			continue
		}
		ref, ok := byPath[path]
		if !ok {
			errs = append(errs, fmt.Errorf("%s: a %s file the reference manifest lacks", path, role))
			continue
		}
		files[path] = entry(ref, role)
	}
	for _, r := range readings {
		f, ok := files[r.Path]
		if !ok {
			errs = append(errs, fmt.Errorf("%s: a license reading for a path that is neither an adopted or measured sample nor a file one links", r.Path))
			continue
		}
		f.LicenseURL, f.LicenseRead, f.LicenseStatus = r.LicenseURL, r.LicenseRead, "verified"
		if r.License != "" {
			f.License = r.License
		}
		files[r.Path] = f
	}
	placedCovers := make(map[string][]string, len(placed))
	for _, p := range placed {
		placedCovers[p.Path] = p.Covers
	}
	errs = append(errs, addCandidates(files, coverage.Gaps, candidates, placedCovers)...)
	errs = append(errs, addSourced(files, sourced, placedCovers, measurements)...)
	errs = append(errs, addMeasures(files, measurements)...)
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

var inCorpus = map[string]string{"adopt": "an adopted", "measure": "a measured"}

func addMeasures(files map[string]File, measurements []disposition.Measurement) []error {
	var errs []error
	for _, m := range measurements {
		for _, path := range m.Inputs {
			f, ok := files[path]
			if !ok {
				errs = append(errs, fmt.Errorf("%s: measurement %s reads it, but it is no file of the manifest; list a sample that owns it as well", path, m.ID))
				continue
			}
			if !slices.Contains(f.Measures, m.ID) {
				f.Measures = append(f.Measures, m.ID)
				slices.Sort(f.Measures)
			}
			files[path] = f
		}
	}
	return errs
}

func addSourced(files map[string]File, sourced []Sourced, covers map[string][]string, measurements []disposition.Measurement) []error {
	listed := map[string]bool{}
	for _, m := range measurements {
		for _, path := range m.Inputs {
			listed[path] = true
		}
	}
	var errs []error
	for _, s := range sourced {
		if !listed[s.Path] {
			errs = append(errs, fmt.Errorf("%s: a sourced input no measurement lists; list it in a measurement's inputs in docs/measurements.json, or remove it from docs/sourced.json", s.Path))
			continue
		}
		f := File{
			Path:            s.Path,
			Role:            "sample",
			Sha256:          s.Sha256,
			Bytes:           s.Bytes,
			Format:          s.Format,
			Truth:           make([]Truth, 0, len(s.Truth)),
			Docs:            []string{},
			Source:          s.Source,
			URL:             s.URL,
			License:         s.License,
			LicenseURL:      s.LicenseURL,
			LicenseRead:     s.LicenseRead,
			LicenseStatus:   "verified",
			Redistributable: true,
		}
		if ids, ok := covers[s.Path]; ok {
			f.Covers = slices.Compact(slices.Sorted(slices.Values(ids)))
		}
		if len(s.Truth) == 0 {
			f.TruthNote = "a sourced input, which records no truth"
		}
		rows := []File{f}
		for _, t := range s.Truth {
			rows[0].Truth = append(rows[0].Truth, Truth{Path: t.Path})
			rows = append(rows, File{
				Path:            t.Path,
				Role:            "truth",
				Sha256:          t.Sha256,
				Bytes:           t.Bytes,
				Format:          t.Format,
				License:         t.License,
				LicenseURL:      t.LicenseURL,
				LicenseRead:     t.LicenseRead,
				LicenseStatus:   "verified",
				Redistributable: true,
			})
		}
		for _, r := range rows {
			if _, ok := files[r.Path]; ok {
				errs = append(errs, fmt.Errorf("%s: a sourced input's %s file at another file's path; give it a path of its own in docs/sourced.json", r.Path, r.Role))
				continue
			}
			files[r.Path] = r
		}
	}
	return errs
}

func addCandidates(files map[string]File, gaps []disposition.Gap, candidates []Candidate, covers map[string][]string) []error {
	fetched := make(map[string]bool, len(candidates))
	byGap := make(map[string]disposition.Gap, len(gaps))
	for _, g := range gaps {
		byGap[g.ID] = g
	}
	sourcedFirst := make([]Candidate, 0, len(candidates))
	for _, first := range []bool{true, false} {
		for _, c := range candidates {
			if sourced(byGap, c.Gap) == first {
				sourcedFirst = append(sourcedFirst, c)
			}
		}
	}
	var errs []error
	for _, c := range sourcedFirst {
		fetched[c.Gap] = true
		g, ok := byGap[c.Gap]
		if !ok || g.Candidate == nil || (g.Candidate.Kind != "sourced" && g.Candidate.Kind != "generated") {
			errs = append(errs, fmt.Errorf("%s: a candidate for gap %s, which has no sourced or generated candidate", c.Path, c.Gap))
			continue
		}
		if g.CoveredBy != "" {
			continue
		}
		if _, ok := files[c.Path]; ok {
			errs = append(errs, fmt.Errorf("%s: gap %s's candidate is at another file's path", c.Path, c.Gap))
			continue
		}
		f := candidateEntry(c, *g.Candidate, covers[c.Path])
		if g.Candidate.Kind == "generated" {
			if err := inherit(&f, c.TruthFrom, files, byGap); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", c.Path, err))
				continue
			}
		} else if c.TruthFrom != "" {
			errs = append(errs, fmt.Errorf("%s: a sourced candidate inherits no truth, but names truth_from %s", c.Path, c.TruthFrom))
			continue
		}
		files[c.Path] = f
	}
	for _, g := range gaps {
		if g.Candidate == nil || fetched[g.ID] || g.CoveredBy != "" {
			continue
		}
		switch g.Candidate.Kind {
		case "sourced":
			errs = append(errs, fmt.Errorf("gap %s: its sourced candidate has no fetched file; run mise run fetch-candidates", g.ID))
		case "generated":
			errs = append(errs, fmt.Errorf("gap %s: its generated candidate has no generated file; run mise run generate-candidates", g.ID))
		}
	}
	return errs
}

func sourced(gaps map[string]disposition.Gap, id string) bool {
	g, ok := gaps[id]
	return ok && g.Candidate != nil && g.Candidate.Kind == "sourced"
}

func inherit(f *File, truthFrom string, files map[string]File, gaps map[string]disposition.Gap) error {
	var licenses []string
	f.LicenseStatus, f.Redistributable = "verified", true
	for _, path := range f.Inputs {
		in, ok := files[path]
		if !ok || in.Role != "sample" || (in.Gap != "" && !sourced(gaps, in.Gap)) {
			return fmt.Errorf("its input %s is neither an adopted sample nor a sourced candidate in the corpus", path)
		}
		if !slices.Contains(licenses, in.License) {
			licenses = append(licenses, in.License)
		}
		if f.LicenseStatus == "verified" {
			f.LicenseStatus = in.LicenseStatus
		}
		f.Redistributable = f.Redistributable && in.Redistributable
	}
	f.License = strings.Join(licenses, "; ")
	if truthFrom == "" {
		f.TruthNote = "a generated candidate, which inherits no input's truth unchanged"
		return nil
	}
	if !slices.Contains(f.Inputs, truthFrom) {
		return fmt.Errorf("truth_from %s is not one of its inputs", truthFrom)
	}
	f.Truth = slices.Clone(files[truthFrom].Truth)
	return nil
}

func candidateEntry(c Candidate, from disposition.Candidate, placed []string) File {
	f := File{
		Path:             c.Path,
		Role:             "sample",
		Sha256:           c.Sha256,
		Bytes:            c.Bytes,
		Format:           c.Format,
		Truth:            []Truth{},
		Docs:             []string{},
		Source:           c.Source,
		URL:              from.URL,
		Ref:              from.Ref,
		License:          from.License,
		LicenseURL:       from.LicenseURL,
		LicenseRead:      from.LicenseRead,
		LicenseStatus:    from.LicenseStatus,
		Redistributable:  from.LicenseStatus == "verified",
		Command:          from.Command,
		Inputs:           slices.Clone(from.Inputs),
		Generator:        from.Generator,
		GeneratorLicense: from.GeneratorLicense,
		Gap:              c.Gap,
	}
	if from.Kind == "sourced" {
		f.TruthNote = "a sourced candidate, which inherits no truth"
	}
	if placed != nil {
		f.Covers = slices.Compact(slices.Sorted(slices.Values(placed)))
	}
	return f
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
