package office

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
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

// OneNote places the traits and features of the OneNote section or notebook docproc's run
// recorded, from its record rows and the run's outcomes for it, each as its JSON text, such as
// "encrypted" quoted. It returns its facts: the sections and pages recorded.
func OneNote(rows []run.Row, outcomes idset.Set) (traits, features idset.Set, facts *jsonvalue.Object) {
	traits, features = idset.Set{}, idset.Set{}
	facts = jsonvalue.NewObject()
	sections, pages := 0, 0
	if outcomes[`"encrypted"`] {
		features.Add("feature:protected-section")
	}
	for _, rec := range rows {
		r := rec.Fields
		if strings.HasPrefix(rec.Name, "section") {
			sections++
			facts.Set("sections", sections)
			if path := r.Get("path"); len(asList(path)) > 1 || utf8.RuneCountInString(jsonvalue.String(path)) > 1 {
				features.Add("feature:section-groups")
			}
		}
		if !strings.HasPrefix(rec.Name, "page") {
			continue
		}
		pages++
		facts.Set("pages", pages)
		md := r.Str("markdown")
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
		if jsonvalue.Truthy(r.Get("ink_alternatives")) {
			traits.Add("trait:scan.handwriting")
		}
		if jsonvalue.Truthy(r.Get("revisions")) {
			features.Add("feature:page-revisions")
		}
		media, _ := r.Get("media").([]any)
		for _, item := range media {
			m, ok := item.(*jsonvalue.Object)
			if !ok {
				continue
			}
			switch m.Str("kind") {
			case "image":
				features.Add("feature:pictures")
			case "ink":
				features.Add("feature:ink")
			case "embedded_file":
				features.Add("feature:embedded-objects")
			}
			if jsonvalue.Truthy(m.Get("ocr")) || jsonvalue.Truthy(m.Get("text")) {
				features.Add("feature:image-ocr")
			}
		}
	}
	return traits, features, facts
}

func asList(v any) []any {
	list, _ := v.([]any)
	return list
}
