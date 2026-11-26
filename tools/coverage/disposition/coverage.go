package disposition

// Coverage is docs/coverage.json, as docs/coverage.md states it. Its fields are in the file's key
// order, and its ID-keyed objects are maps, written sorted by ID.
type Coverage struct {
	// Vocabulary defines every ID the data uses.
	Vocabulary Vocabulary `json:"vocabulary"`
	// LibraryPlacement names the docproc run the library IDs were placed from.
	LibraryPlacement LibraryPlacement `json:"library_placement"`
	// CorpusPlacement names the docproc run over the corpus the manifest's covers were placed
	// from.
	CorpusPlacement *CorpusPlacement `json:"corpus_placement,omitempty"`
	// Samples are the ~/docs samples, sorted by path.
	Samples []Sample `json:"samples"`
	// Gaps are the scenarios no adopted sample exercises, each with its candidate.
	Gaps []Gap `json:"gaps"`
	// DocsWithoutSample are the ~/docs truth and doc files no sample links.
	DocsWithoutSample []DocWithoutSample `json:"docs_without_sample"`
}

// Vocabulary is the data's three axes, each keyed by ID.
type Vocabulary struct {
	// Library is the kind:, route:, check:, and capability: IDs, with their links.
	Library map[string]Library `json:"library"`
	// Trait is the trait: IDs, with their rules.
	Trait map[string]Rule `json:"trait"`
	// Feature is the feature: IDs, with their rules.
	Feature map[string]Rule `json:"feature"`
}

// Library is a library ID's definition in docproc or docinfer.
type Library struct {
	// Identifier is the Go identifier that defines it.
	Identifier string `json:"identifier"`
	// Link is the pinned source line that defines it.
	Link string `json:"link"`
	// FedBy, for a capability, names the routes that feed it.
	FedBy *FedBy `json:"fed_by,omitempty"`
}

// FedBy is the routes that feed a capability, and the source lines that show it.
type FedBy struct {
	// Routes are the route: IDs.
	Routes []string `json:"routes"`
	// Links are the pinned source lines.
	Links []string `json:"links"`
}

// Rule is a trait or feature ID's one-line rule.
type Rule struct {
	// Rule is the test a sample meets to carry the ID.
	Rule string `json:"rule"`
}

// LibraryPlacement is the docproc run the library IDs were placed from.
type LibraryPlacement struct {
	// Run is the run's ID.
	Run string `json:"run"`
	// Command is the command that placed them.
	Command string `json:"command"`
}

// CorpusPlacement is the docproc run over the corpus's format directories that place-traits
// --corpus places the manifest's samples from.
type CorpusPlacement struct {
	// Run is the run's ID.
	Run string `json:"run"`
	// Docproc is the docproc release that ran, with its commit and its command module's version.
	Docproc string `json:"docproc"`
	// DocprocCommand is the docproc command that made the run, run from the repository root.
	DocprocCommand string `json:"docproc_command"`
	// Command is the command that places the covers.
	Command string `json:"command"`
}

// Sample is one sample's entry.
type Sample struct {
	// Path is its path under ~/docs.
	Path string `json:"path"`
	// Sha256 is its content's SHA-256, in hex.
	Sha256 string `json:"sha256"`
	// Library are its library IDs.
	Library []string `json:"library"`
	// Traits are its trait IDs.
	Traits []string `json:"traits"`
	// Features are its feature IDs.
	Features []string `json:"features"`
	// Judged holds the evidence for each ID placed by judgement, keyed by ID.
	Judged map[string]string `json:"judged,omitempty"`
	// Format is its media type.
	Format string `json:"format"`
	// Bytes is its size.
	Bytes int64 `json:"bytes"`
	// LicenseStatus is how its license is known, such as verified.
	LicenseStatus string `json:"license_status"`
	// Redistributable is whether it may be redistributed.
	Redistributable bool `json:"redistributable"`
	// Sensitive is whether the manifest marks it sensitive or local_only.
	Sensitive bool `json:"sensitive,omitempty"`
	// Breaks are the STANDARDS.md standards it breaks, each with its reason.
	Breaks []Break `json:"breaks,omitempty"`
	// Truth are its truth files' paths.
	Truth []string `json:"truth"`
	// Docs are its doc files' paths.
	Docs []string `json:"docs"`
	// Disposition is adopt, trim, or replace.
	Disposition string `json:"disposition"`
	// TrimmedFor, for a trim, is the adopted sample that covers it.
	TrimmedFor string `json:"trimmed_for,omitempty"`
	// Gap, for a replace, is the gap that replaces it.
	Gap string `json:"gap,omitempty"`
}

// Break is a STANDARDS.md standard a sample breaks.
type Break struct {
	// Standard is the opening of the standard's STANDARDS.md bullet, verbatim.
	Standard string `json:"standard"`
	// Reason is how the sample breaks it.
	Reason string `json:"reason"`
}

// Gap is a set of scenarios no adopted sample exercises.
type Gap struct {
	// ID names it.
	ID string `json:"id"`
	// Covers are the IDs a sample filling it must exercise.
	Covers []string `json:"covers"`
	// SplitFrom, for an open gap, is the gap whose candidate's covers lack this one's covers.
	SplitFrom string `json:"split_from,omitempty"`
	// CoveredBy is the sample whose covers contain its candidate's, which leaves the corpus.
	CoveredBy string `json:"covered_by,omitempty"`
	// Candidate is the file researched to fill it.
	Candidate *Candidate `json:"candidate"`
}

// Candidate is a gap's candidate, by its Kind: sourced, generated, or unavailable. Each kind
// carries its own fields, and the others are absent.
type Candidate struct {
	// Kind is sourced, generated, or unavailable.
	Kind string `json:"kind"`
	// URL is where a sourced file is fetched.
	URL string `json:"url,omitempty"`
	// Ref is the commit or revision URL is pinned to.
	Ref string `json:"ref,omitempty"`
	// License is the license read at the source.
	License string `json:"license,omitempty"`
	// LicenseURL is where the license was read.
	LicenseURL string `json:"license_url,omitempty"`
	// LicenseStatus is verified once the license is read.
	LicenseStatus string `json:"license_status,omitempty"`
	// LicenseRead is the date the license was read, YYYY-MM-DD.
	LicenseRead string `json:"license_read,omitempty"`
	// Command is the command that makes a generated file.
	Command string `json:"command,omitempty"`
	// Inputs are the adopted samples and sourced candidates Command reads.
	Inputs []string `json:"inputs,omitempty"`
	// Generator is the generating tool and its version.
	Generator string `json:"generator,omitempty"`
	// GeneratorLicense is the generator's license.
	GeneratorLicense string `json:"generator_license,omitempty"`
	// Searched is what was searched for an unavailable candidate.
	Searched string `json:"searched,omitempty"`
	// Reason is why no file can fill the gap.
	Reason string `json:"reason,omitempty"`
	// Fit is the candidate's quality and how it exercises the covers.
	Fit string `json:"fit,omitempty"`
}

// DocWithoutSample is a ~/docs truth or doc file no sample links.
type DocWithoutSample struct {
	// Path is its path under ~/docs.
	Path string `json:"path"`
	// Handling is what is done with it.
	Handling string `json:"handling"`
}

// Placement is one element of place-traits' output: a sample's traits and features.
type Placement struct {
	// Path is the sample's path.
	Path string `json:"path"`
	// Traits are its trait IDs.
	Traits []string `json:"traits"`
	// Features are its feature IDs.
	Features []string `json:"features"`
}

// Placed is a gap candidate as the corpus placement placed it: its gap, path, size, and the input
// whose truth it inherits, as docs/candidates.json records them, and its covers, as place-traits
// --corpus prints them.
type Placed struct {
	// Gap is the id of the gap it fills.
	Gap string
	// Path is its path from the repository root.
	Path string
	// Bytes is its size.
	Bytes int64
	// Covers are its library, trait, and feature IDs.
	Covers []string
	// TruthFrom, for a generated candidate, is the input whose truth it inherits unchanged.
	TruthFrom string
}
