package office

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
	"github.com/standards-lab/docs-corpus/tools/coverage/text"
)

var (
	noteTag     = regexp.MustCompile(`\[(To Do|Important|Question|Idea|Critical)\]`)
	markdownURL = regexp.MustCompile(`\]\(https?://`)
	equation    = regexp.MustCompile(`\$\$[^$]+\$\$`)
)

func hasTableRow(md string) bool {
	for line := range strings.SplitSeq(md, "\n") {
		trimmed := strings.TrimRightFunc(line, text.IsSpace)
		if strings.HasPrefix(line, "|") && len(trimmed) > 1 && strings.HasSuffix(trimmed, "|") {
			return true
		}
	}
	return false
}

func hasNoteTag(md string) bool {
	for line := range strings.SplitSeq(md, "\n") {
		trimmed := strings.TrimLeftFunc(line, text.IsSpace)
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ") {
			return true
		}
	}
	return noteTag.MatchString(md)
}

type onenoteRow struct {
	Path            json.RawMessage   `json:"path"`
	Markdown        string            `json:"markdown"`
	InkAlternatives []json.RawMessage `json:"ink_alternatives"`
	Revisions       []json.RawMessage `json:"revisions"`
	Media           []struct {
		Kind string `json:"kind"`
		OCR  string `json:"ocr"`
		Text string `json:"text"`
	} `json:"media"`
}

func inGroup(path json.RawMessage) bool {
	var list []json.RawMessage
	if json.Unmarshal(path, &list) == nil {
		return len(list) > 1
	}
	var name string
	return json.Unmarshal(path, &name) == nil && utf8.RuneCountInString(name) > 1
}

// OneNoteFacts is what OneNote found in the records.
type OneNoteFacts struct {
	// Sections is the number of section records.
	Sections int
	// Pages is the number of page records.
	Pages int
}

// OneNote places the traits and features of the OneNote section or notebook docproc's run
// recorded, from its record rows and the run's outcomes for it, each as its JSON text, such as
// "encrypted" quoted. It returns its facts. A row field of another JSON type than OneNote's
// records give it reads as absent.
func OneNote(rows []run.Row, outcomes idset.Set) (traits, features idset.Set, facts OneNoteFacts) {
	traits, features = idset.Set{}, idset.Set{}
	if outcomes[`"encrypted"`] {
		features.Add("feature:protected-section")
	}
	for _, rec := range rows {
		var r onenoteRow
		_ = json.Unmarshal(rec.Fields, &r)
		if strings.HasPrefix(rec.Name, "section") {
			facts.Sections++
			if inGroup(r.Path) {
				features.Add("feature:section-groups")
			}
		}
		if !strings.HasPrefix(rec.Name, "page") {
			continue
		}
		facts.Pages++
		md := r.Markdown
		if hasTableRow(md) {
			features.Add("feature:tables")
		}
		if hasNoteTag(md) {
			features.Add("feature:note-tags")
		}
		if markdownURL.MatchString(md) {
			features.Add("feature:hyperlinks")
		}
		if equation.MatchString(md) {
			features.Add("feature:equations")
		}
		if len(r.InkAlternatives) > 0 {
			traits.Add("trait:scan.handwriting")
		}
		if len(r.Revisions) > 0 {
			features.Add("feature:page-revisions")
		}
		for _, m := range r.Media {
			switch m.Kind {
			case "image":
				features.Add("feature:pictures")
			case "ink":
				features.Add("feature:ink")
			case "embedded_file":
				features.Add("feature:embedded-objects")
			}
			if m.OCR != "" || m.Text != "" {
				features.Add("feature:image-ocr")
			}
		}
	}
	return traits, features, facts
}
