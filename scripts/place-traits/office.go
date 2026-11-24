package main

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type zipEntry struct {
	name string
	data []byte
}

func zipTexts(path string) []zipEntry {
	z, err := zip.OpenReader(path)
	if err != nil && !errors.Is(err, zip.ErrInsecurePath) {
		return nil
	}
	defer z.Close()
	var entries []zipEntry
	index := map[string]int{}
	for _, f := range z.File {
		rc, err := f.Open()
		if err != nil {
			return nil
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil
		}
		if i, ok := index[f.Name]; ok {
			entries[i].data = data
			continue
		}
		index[f.Name] = len(entries)
		entries = append(entries, zipEntry{f.Name, data})
	}
	return entries
}

var (
	sheetPart      = regexp.MustCompile(`^xl/worksheets/sheet\d+\.xml$`)
	slidePart      = regexp.MustCompile(`^ppt/slides/slide\d+\.xml$`)
	wordPart       = regexp.MustCompile(`^word/(document|header\d*|footer\d*|footnotes|endnotes)\.xml$`)
	wordTable      = regexp.MustCompile(`<w:tbl[ >]`)
	slideTable     = regexp.MustCompile(`<a:tbl[ >]`)
	wordMerge      = regexp.MustCompile(`<w:gridSpan |<w:vMerge`)
	slideMerge     = regexp.MustCompile(`gridSpan="|rowSpan="|hMerge="1"|vMerge="1"`)
	chartPart      = regexp.MustCompile(`(charts|chartsheets)/(chart|chartEx)\d*\.xml$`)
	commentPart    = regexp.MustCompile(`^(word/comments\.xml|ppt/comments/|ppt/modernComment|word/commentsExtended)`)
	commentElement = regexp.MustCompile(`<w:comment |<p:cm |<p188:cm `)
	cellNotePart   = regexp.MustCompile(`^xl/comments\d*\.xml$`)
	threadedPart   = regexp.MustCompile(`^xl/threadedComments/`)
	revision       = regexp.MustCompile(`<w:(ins|del) `)
	notesSlide     = regexp.MustCompile(`^ppt/notesSlides/notesSlide\d+\.xml$`)
	notesText      = regexp.MustCompile(`<a:t>[^<]*\S`)
	wordTextBox    = regexp.MustCompile(`<w:txbxContent|<v:textbox`)
	drawingPart    = regexp.MustCompile(`^xl/drawings/drawing\d+\.xml$`)
	smartArtPart   = regexp.MustCompile(`/diagrams/data\d*\.xml$`)
	formula        = regexp.MustCompile(`<f[ >]`)
	externalLink   = regexp.MustCompile(`Type="[^"]*/hyperlink"[^>]*TargetMode="External"|TargetMode="External"[^>]*Type="[^"]*/hyperlink"`)
	cropped        = regexp.MustCompile(`<a:srcRect [^>]*[lrtb]="[1-9]`)
	numFmt         = regexp.MustCompile(`<numFmt numFmtId="(\d+)" formatCode="([^"]*)"`)
	numFmtLiteral  = regexp.MustCompile(`&quot;.*?&quot;|"[^"]*"|\[[^\]]*\]|\\.`)
	dateCode       = regexp.MustCompile(`[dmyhs]`)
	cellXfs        = regexp.MustCompile(`(?s)<cellXfs[^>]*>(.*?)</cellXfs>`)
	xfElement      = regexp.MustCompile(`<xf ([^>]*)`)
	numFmtID       = regexp.MustCompile(`numFmtId="(\d+)"`)
	cell           = regexp.MustCompile(`(?s)<c ([^>]*)>(.*?)</c>`)
	cellStyle      = regexp.MustCompile(`\bs="(\d+)"`)
	slideSize      = regexp.MustCompile(`<p:sldSz cx="(\d+)" cy="(\d+)"`)
	picture        = regexp.MustCompile(`(?s)<p:pic>.*?</p:pic>`)
	pictureFrame   = regexp.MustCompile(`<a:off x="(-?\d+)" y="(-?\d+)"/><a:ext cx="(\d+)" cy="(\d+)"`)
	placeholder    = regexp.MustCompile(`<p:ph[ />]`)
	markup         = regexp.MustCompile(`<[^>]+>`)
)

func containsBefore(text, inner, closing string, start int) bool {
	rest := text[start:]
	end := strings.Index(rest, closing)
	if end < 0 {
		end = len(rest)
	}
	i := strings.Index(rest, inner)
	return i >= 0 && i < end
}

func nestedTable(v string) bool {
	for _, m := range regexp.MustCompile(`<w:tc[ >]`).FindAllStringIndex(v, -1) {
		rest := v[m[1]:]
		end := strings.Index(rest, "</w:tc>")
		if end < 0 {
			end = len(rest)
		}
		if loc := wordTable.FindStringIndex(rest); loc != nil && loc[0] < end {
			return true
		}
	}
	return false
}

func dropPlaceholderShapes(v string) string {
	var b strings.Builder
	pos := 0
	for {
		i := strings.Index(v[pos:], "<p:sp>")
		if i < 0 {
			break
		}
		start := pos + i
		body := v[start+len("<p:sp>"):]
		end := strings.Index(body, "</p:sp>")
		ph := placeholder.FindStringIndex(body)
		if end >= 0 && ph != nil && ph[0] < end {
			b.WriteString(v[pos:start])
			pos = start + len("<p:sp>") + end + len("</p:sp>")
			continue
		}
		b.WriteString(v[pos : start+1])
		pos = start + 1
	}
	b.WriteString(v[pos:])
	return b.String()
}

func shapeWithText(v string) bool {
	for i := 0; ; {
		j := strings.Index(v[i:], "<p:sp>")
		if j < 0 {
			return false
		}
		start := i + j + len("<p:sp>")
		rest := v[start:]
		end := strings.Index(rest, "</p:sp>")
		if end < 0 {
			end = len(rest)
		}
		if tb := strings.Index(rest[:end], "<p:txBody>"); tb >= 0 && strings.Contains(rest[tb+len("<p:txBody>"):end], "<a:t>") {
			return true
		}
		i = i + j + 1
	}
}

func drawingWithText(v string) bool {
	for i := 0; ; {
		j := strings.Index(v[i:], "<xdr:txBody>")
		if j < 0 {
			return false
		}
		if containsBefore(v, "<a:t>", "</xdr:txBody>", i+j+len("<xdr:txBody>")) {
			return true
		}
		i = i + j + 1
	}
}

func hasFootnote(notes string) bool {
	for _, tag := range []string{"<w:footnote", "<w:endnote"} {
		for i := 0; ; {
			j := strings.Index(notes[i:], tag)
			if j < 0 {
				break
			}
			rest := notes[i+j+len(tag):]
			if k := strings.IndexByte(rest, '>'); k >= 0 {
				rest = rest[:k]
			}
			separator := strings.Contains(rest, `w:type="separator"`) || strings.Contains(rest, `w:type="continuationSeparator"`) ||
				strings.Contains(rest, `w:type="continuationNotice"`)
			if !separator && strings.Contains(rest, "w:id=") {
				return true
			}
			i = i + j + 1
		}
	}
	return false
}

type xmlPart struct {
	name, text string
}

func ooxml(path string) (set, *object) {
	files := zipTexts(path)
	var parts []xmlPart
	byName := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f.name, ".xml") || strings.HasSuffix(f.name, ".rels") {
			parts = append(parts, xmlPart{f.name, string(f.data)})
			byName[f.name] = string(f.data)
		}
	}
	var bodyParts, relParts, sheets, slides, word []string
	for _, p := range parts {
		if strings.HasSuffix(p.name, ".rels") {
			relParts = append(relParts, p.text)
		} else {
			bodyParts = append(bodyParts, p.text)
		}
		if sheetPart.MatchString(p.name) {
			sheets = append(sheets, p.text)
		}
		if slidePart.MatchString(p.name) {
			slides = append(slides, p.text)
		}
		if wordPart.MatchString(p.name) {
			word = append(word, p.text)
		}
	}
	body, rels := strings.Join(bodyParts, "\n"), strings.Join(relParts, "\n")
	anyOf := func(texts []string, test func(string) bool) bool {
		for _, t := range texts {
			if test(t) {
				return true
			}
		}
		return false
	}
	anyName := func(test func(string) bool) bool {
		for _, f := range files {
			if test(f.name) {
				return true
			}
		}
		return false
	}
	anyPart := func(test func(xmlPart) bool) bool {
		for _, p := range parts {
			if test(p) {
				return true
			}
		}
		return false
	}
	features := set{}
	facts := newObject()
	facts.set("parts", len(files))
	tables := 0
	for _, v := range word {
		tables += len(wordTable.FindAllStringIndex(v, -1))
	}
	for _, v := range slides {
		tables += len(slideTable.FindAllStringIndex(v, -1))
	}
	if tables > 0 {
		features.add("feature:tables")
	}
	if anyOf(word, nestedTable) {
		features.add("feature:nested-tables")
	}
	if anyOf(word, wordMerge.MatchString) || anyOf(slides, slideMerge.MatchString) ||
		anyOf(sheets, func(v string) bool { return strings.Contains(v, "<mergeCell ") }) {
		features.add("feature:merged-cells")
	}
	if anyName(chartPart.MatchString) {
		features.add("feature:charts")
	}
	if anyName(commentPart.MatchString) && commentElement.MatchString(body) {
		features.add("feature:comments")
	}
	if anyName(cellNotePart.MatchString) || anyName(threadedPart.MatchString) {
		features.add("feature:cell-notes")
	}
	if anyOf(word, revision.MatchString) {
		features.add("feature:tracked-changes")
	}
	if hasFootnote(byName["word/footnotes.xml"] + byName["word/endnotes.xml"]) {
		features.add("feature:footnotes")
	}
	if anyPart(func(p xmlPart) bool { return notesSlide.MatchString(p.name) && notesText.MatchString(p.text) }) {
		features.add("feature:speaker-notes")
	}
	if wordTextBox.MatchString(body) ||
		anyOf(slides, func(v string) bool { return shapeWithText(dropPlaceholderShapes(v)) }) ||
		anyPart(func(p xmlPart) bool { return drawingPart.MatchString(p.name) && drawingWithText(p.text) }) {
		features.add("feature:text-boxes")
	}
	if strings.Contains(body, "<m:oMath") {
		features.add("feature:equations")
	}
	if anyName(smartArtPart.MatchString) {
		features.add("feature:smartart")
	}
	if anyOf(sheets, formula.MatchString) {
		features.add("feature:formulas")
	}
	if len(sheets) > 0 && xlsxDates(parts, byName) {
		features.add("feature:dates")
	}
	if externalLink.MatchString(rels) {
		features.add("feature:hyperlinks")
	}
	media := 0
	for _, f := range files {
		if strings.Contains(f.name, "/media/") {
			media++
		}
	}
	if media > 0 {
		features.add("feature:pictures")
	}
	if anyPart(func(p xmlPart) bool { return cropped.MatchString(p.text) }) || slideOffpage(parts, byName) {
		features.add("feature:picture-layout")
	}
	if anyName(func(n string) bool { return strings.Contains(n, "/embeddings/") }) ||
		strings.Contains(body, "<o:OLEObject") || strings.Contains(body, "<p:oleObj") {
		features.add("feature:embedded-objects")
	}
	facts.set("tables", tables)
	facts.set("media", media)
	return features, facts
}

func xlsxDates(parts []xmlPart, byName map[string]string) bool {
	styles := byName["xl/styles.xml"]
	dateFormats := set{}
	for i := 14; i < 23; i++ {
		dateFormats.add(strconv.Itoa(i))
	}
	for i := 45; i < 48; i++ {
		dateFormats.add(strconv.Itoa(i))
	}
	for _, m := range numFmt.FindAllStringSubmatch(styles, -1) {
		code := strings.ToLower(numFmtLiteral.ReplaceAllString(m[2], ""))
		if dateCode.MatchString(code) {
			dateFormats.add(m[1])
		}
	}
	xfs := ""
	if m := cellXfs.FindStringSubmatch(styles); m != nil {
		xfs = m[1]
	}
	styled := set{}
	for i, m := range xfElement.FindAllStringSubmatch(xfs, -1) {
		if f := numFmtID.FindStringSubmatch(m[1]); f != nil && dateFormats[f[1]] {
			styled.add(strconv.Itoa(i))
		}
	}
	var sheetTexts []string
	for _, p := range parts {
		if strings.HasPrefix(p.name, "xl/worksheets/sheet") {
			sheetTexts = append(sheetTexts, p.text)
		}
	}
	for _, m := range cell.FindAllStringSubmatch(strings.Join(sheetTexts, "\n"), -1) {
		style := cellStyle.FindStringSubmatch(m[1])
		if style != nil && styled[style[1]] && strings.Contains(m[2], "<v>") && !strings.Contains(m[1], `t="s"`) {
			return true
		}
	}
	return false
}

func slideOffpage(parts []xmlPart, byName map[string]string) bool {
	m := slideSize.FindStringSubmatch(byName["ppt/presentation.xml"])
	if m == nil {
		return false
	}
	cx, _ := strconv.ParseInt(m[1], 10, 64)
	cy, _ := strconv.ParseInt(m[2], 10, 64)
	for _, p := range parts {
		if !slidePart.MatchString(p.name) {
			continue
		}
		for _, pic := range picture.FindAllString(p.text, -1) {
			off := pictureFrame.FindStringSubmatch(pic)
			if off == nil {
				continue
			}
			var n [4]int64
			for i := range n {
				n[i], _ = strconv.ParseInt(off[i+1], 10, 64)
			}
			x, y, w, h := n[0], n[1], n[2], n[3]
			if x < 0 || y < 0 || x+w > cx || y+h > cy {
				return true
			}
		}
	}
	return false
}

var legacyTarget = map[string]string{".doc": "docx", ".rtf": "docx", ".wpd": "docx", ".xls": "xlsx", ".ppt": "pptx"}

var ooxmlKinds = set{}

func init() {
	ooxmlKinds.add("docx", "xlsx", "pptx", "docm", "dotm", "xlsm", "xltm", "pptm", "dotx", "xltx", "potx", "ppsx")
}

func converted(path, tmp string, n int) string {
	target, ok := legacyTarget[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return ""
	}
	out := filepath.Join(tmp, strconv.Itoa(n))
	if err := os.MkdirAll(out, 0o755); err != nil {
		return ""
	}
	run([]string{"soffice", "-env:UserInstallation=file://" + tmp + "/profile", "--headless", "--convert-to", target, "--outdir", out, path}, nil, convertTimeout)
	entries, err := os.ReadDir(out)
	if err != nil || len(entries) == 0 {
		return ""
	}
	return filepath.Join(out, entries[0].Name())
}

func odfText(path string) string {
	for _, f := range zipTexts(path) {
		if f.name == "content.xml" {
			return markup.ReplaceAllString(string(f.data), " ")
		}
	}
	return ""
}
