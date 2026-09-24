package cel

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/markdown"
	"github.com/peedrr/agent-kb/internal/storage"
)

// convertDateField converts ISO-8601 date strings to time.Time for CEL compatibility.
// It handles both RFC3339 ("2024-01-01T00:00:00Z") and date-only ("2024-01-01") formats.
// Any string that is not one of those formats is returned unchanged.
func convertDateField(value any) any {
	str, ok := value.(string)
	if !ok {
		return value
	}

	// Try RFC3339 first (includes time)
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		return t
	}

	// Fall back to date-only format (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", str); err == nil {
		return t
	}

	// Keep original string if parsing fails
	return value
}

// BuildPage assembles the page map used by CEL expressions from a page's
// constituent parts.
func BuildPage(relPath string, fm *frontmatter.ParsedFrontmatter, body []byte, astDoc ast.Node, source []byte) map[string]any {
	page := make(map[string]any)

	page["file"] = map[string]any{
		"path": relPath,
		"name": filepath.Base(relPath),
		"dir":  filepath.Dir(relPath),
	}

	fmMap := make(map[string]any, len(fm.Fields)+2)
	for k, v := range fm.Fields {
		fmMap[k] = convertDateField(v)
	}
	fmMap["type"] = fm.Type
	fmMap["title"] = fm.Title
	page["frontmatter"] = fmMap

	bodyStr := string(body)
	page["content"] = map[string]any{
		"raw":        bodyStr,
		"word_count": len(strings.Fields(bodyStr)),
		"char_count": len(body),
	}

	page["ast"] = map[string]any{
		"headings":    flattenHeadings(astDoc, source),
		"links":       flattenLinks(astDoc, source),
		"code_blocks": flattenCodeBlocks(astDoc, source),
	}

	page["akb"] = map[string]any{
		"provenance_markers": parseProvenanceMarkers(bodyStr),
		"annotations":        parseAnnotations(bodyStr),
	}

	return page
}

// BuildOldPage reads the on-disk version of a page and builds its page map
// from the pre-modification state. If the file does not exist, it returns nil.
func BuildOldPage(relPath string, store storage.Provider) (map[string]any, error) {
	ctx := context.Background()

	exists, err := store.Exists(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("check existing page: %w", err)
	}
	if !exists {
		return nil, nil
	}

	data, err := store.Read(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("read existing page: %w", err)
	}

	fm, body, err := frontmatter.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse existing frontmatter: %w", err)
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(body))

	return BuildPage(relPath, fm, body, doc, body), nil
}

// The exclusion helpers below duplicate internal/markdown's unexported
// equivalents (fenced code block, inline code, and HTML comment ranges) for
// the provenance-marker and annotation parsers.

func offsetToLine(source []byte, offset int) int {
	if offset < 0 || offset > len(source) {
		return 1
	}
	return bytes.Count(source[:offset], []byte("\n")) + 1
}

func extractText(n ast.Node, source []byte) string {
	var text []byte
	_ = ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := child.(*ast.Text); ok {
			text = append(text, t.Value(source)...)
		}
		return ast.WalkContinue, nil
	})
	return string(text)
}

func flattenHeadings(doc ast.Node, source []byte) []map[string]any {
	var headings []map[string]any
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		headings = append(headings, map[string]any{
			"level": h.Level,
			"text":  extractText(h, source),
			"line":  offsetToLine(source, h.Pos()),
		})
		return ast.WalkContinue, nil
	})
	return headings
}

// flattenLinks returns one entry per link in source. Wikilink tokens are
// merged in from markdown.ParseWikilinks because goldmark yields no ast.Link
// node for a bare [[...]]. Entries are sorted by source offset so CEL
// evaluation order is deterministic.
//
// A wikilink-form entry reports the target in the one normalized spelling the
// link graph records: the parser's destination when it has one, and otherwise
// goldmark's destination run through normalizeWikilinkTarget. A plain markdown
// link keeps goldmark's destination as written, because the link graph ignores
// those links entirely.
//
// Bracket runs nested inside a quoted title or inside a path destination are
// part of that destination, so the parser records no token for them and they
// emit no entry of their own.
func flattenLinks(doc ast.Node, source []byte) []map[string]any {
	type mergedLink struct {
		offset int
		entry  map[string]any
	}

	var goldmarkLinks []*ast.Link
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if l, ok := n.(*ast.Link); ok {
			goldmarkLinks = append(goldmarkLinks, l)
		}
		return ast.WalkContinue, nil
	})

	// Goldmark sets Link.Pos() to the opening '[' of the link text, which for
	// the [[display]](dest) form is the wikilink token start itself, so the two
	// spans describe the same token.
	byOffset := make(map[int]int, len(goldmarkLinks))
	for i, l := range goldmarkLinks {
		byOffset[l.Pos()] = i
	}

	var merged []mergedLink
	consumed := make([]bool, len(goldmarkLinks))

	wikilinks := markdown.ParseWikilinks(string(source))

	// A wikilink token whose parenthesized destination is itself a bracket
	// token — [[a]]([[b]]) — is malformed: its destination names a wikilink
	// rather than a path. The bracket parts of tokens never overlap, so such an
	// inner token can only sit inside the outer token's destination part. Drop
	// the outer token and the goldmark link describing it, so the surviving
	// entries neither overlap nor carry the bracket-shaped target. A real path
	// destination (x.md, notes/[[weird]].md) is not a bracket token: the outer
	// token stays a valid link that keeps its merge with the goldmark link, and
	// the parser has already suppressed the bracket runs nested in its title or
	// path, so those emit no entry of their own.
	containsAnother := make([]bool, len(wikilinks))
	for i, outer := range wikilinks {
		for j, inner := range wikilinks {
			if i == j {
				continue
			}
			if outer.Start <= inner.Start && inner.End <= outer.End &&
				(outer.Start < inner.Start || inner.End < outer.End) &&
				strings.HasPrefix(outer.Destination, "[[") {
				containsAnother[i] = true
			}
		}
	}

	for k, wl := range wikilinks {
		if containsAnother[k] {
			if i, matched := byOffset[wl.Start]; matched {
				consumed[i] = true
			}
			continue
		}

		i, matched := byOffset[wl.Start]
		if matched {
			consumed[i] = true
			l := goldmarkLinks[i]
			// The parser's destination is the link graph's spelling, so it wins
			// whenever it exists. For degenerate bracket nesting such as
			// [[[a]]](notes/x.md) the parser keeps the token plain while goldmark
			// still parses a real inline link; there the goldmark destination is
			// normalized into that same spelling rather than reported as written.
			// A destination that normalizes away (".md", "./") leaves the
			// token a plain wikilink, which is how the link graph records it.
			target := wl.Destination
			if target == "" && len(l.Destination) > 0 {
				target = normalizeWikilinkTarget(string(l.Destination))
			}
			if target != "" {
				merged = append(merged, mergedLink{offset: wl.Start, entry: map[string]any{
					"target":      target,
					"text":        extractText(l, source),
					"is_wikilink": true,
					"line":        offsetToLine(source, wl.Start),
				}})
				continue
			}
		}
		// Plain wikilink: bare, or carrying parens the parser ignores — empty,
		// whitespace-only, angle-only, unclosed, or a destination that normalizes
		// away. Goldmark's entry for such a token is dropped, so the token still
		// yields exactly one entry, carrying the bracket target.
		merged = append(merged, mergedLink{offset: wl.Start, entry: map[string]any{
			"target":      wl.Target,
			"text":        wl.Display,
			"is_wikilink": true,
			"line":        offsetToLine(source, wl.Start),
		}})
	}

	for i, l := range goldmarkLinks {
		if consumed[i] {
			continue
		}
		pos := l.Pos()
		isWikilink := pos+1 < len(source) && source[pos] == '[' && source[pos+1] == '['
		target := string(l.Destination)
		if isWikilink {
			// A destination that normalizes away is not a spelling the link
			// graph would record, so goldmark's spelling is kept as written.
			if normalized := normalizeWikilinkTarget(target); normalized != "" {
				target = normalized
			}
		}
		merged = append(merged, mergedLink{offset: pos, entry: map[string]any{
			"target":      target,
			"text":        extractText(l, source),
			"is_wikilink": isWikilink,
			"line":        offsetToLine(source, pos),
		}})
	}

	if len(merged) == 0 {
		return nil
	}

	sort.SliceStable(merged, func(i, j int) bool { return merged[i].offset < merged[j].offset })

	links := make([]map[string]any, 0, len(merged))
	for _, m := range merged {
		links = append(links, m.entry)
	}
	return links
}

// normalizeWikilinkTarget normalizes a wikilink-form destination reported by
// goldmark, which carries no link title (goldmark keeps a title in a field of
// its own). A scheme-qualified destination is a URL rather than a KB path, so
// it is left as written.
func normalizeWikilinkTarget(dest string) string {
	if strings.Contains(dest, "://") {
		return dest
	}
	return markdown.NormalizeBareDestination(dest)
}

func flattenCodeBlocks(doc ast.Node, source []byte) []map[string]any {
	var blocks []map[string]any
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		cb, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}
		lang := ""
		if cb.Info != nil {
			lang = string(cb.Info.Value(source))
		}
		blocks = append(blocks, map[string]any{
			"language": lang,
			"line":     offsetToLine(source, cb.Pos()),
		})
		return ast.WalkContinue, nil
	})
	return blocks
}

var provenanceRe = regexp.MustCompile(`\^\[([^\]]+?)\]`)

func parseProvenanceMarkers(content string) []any {
	if content == "" {
		return nil
	}
	excluded := computeExclusions(content)
	matches := provenanceRe.FindAllStringSubmatchIndex(content, -1)
	var markers []any
	for _, m := range matches {
		innerStart := m[2]
		innerEnd := m[3]
		if excluded.isExcluded(innerStart, innerEnd) {
			continue
		}
		markerType := content[innerStart:innerEnd]
		markers = append(markers, map[string]any{
			"type":     markerType,
			"position": m[0],
		})
	}
	return markers
}

var annotationRe = regexp.MustCompile(`<!--\s*olw-auto:\s*(.+?)\s*-->`)

func parseAnnotations(content string) []any {
	if content == "" {
		return nil
	}
	// Only code ranges are excluded: an annotation is itself an HTML comment,
	// so including comment ranges here would exclude every real annotation.
	var excluded exclusionSet
	excluded.ranges = append(excluded.ranges, fencedCodeBlockRanges(content)...)
	excluded.ranges = append(excluded.ranges, inlineCodeRanges(content)...)

	matches := annotationRe.FindAllStringSubmatchIndex(content, -1)
	var annotations []any
	for _, m := range matches {
		if excluded.isExcluded(m[0], m[1]) {
			continue
		}
		fieldsStr := content[m[2]:m[3]]
		fields := parseAnnotationFields(fieldsStr)
		annotations = append(annotations, map[string]any{
			"type":     "olw-auto",
			"fields":   fields,
			"position": m[0],
		})
	}
	return annotations
}

func parseAnnotationFields(s string) map[string]string {
	fields := make(map[string]string)
	s = strings.TrimSpace(s)
	if s == "" {
		return fields
	}
	tokens := strings.Fields(s)
	for _, token := range tokens {
		key, val, ok := strings.Cut(token, "=")
		if !ok {
			continue
		}
		fields[key] = val
	}
	return fields
}

type exclusion struct {
	start int
	end   int
}

type exclusionSet struct {
	ranges []exclusion
}

func (es *exclusionSet) isExcluded(start, end int) bool {
	for _, r := range es.ranges {
		if start >= r.start && end <= r.end {
			return true
		}
	}
	return false
}

func computeExclusions(content string) exclusionSet {
	var es exclusionSet
	es.ranges = append(es.ranges, fencedCodeBlockRanges(content)...)
	es.ranges = append(es.ranges, inlineCodeRanges(content)...)
	es.ranges = append(es.ranges, htmlCommentRanges(content)...)
	return es
}

func fencedCodeBlockRanges(content string) []exclusion {
	var ranges []exclusion
	inBlock := false
	blockStart := 0
	for i := 0; i < len(content); {
		if strings.HasPrefix(content[i:], "```") {
			if !inBlock {
				inBlock = true
				blockStart = i
				i += 3
			} else {
				ranges = append(ranges, exclusion{start: blockStart, end: i + 3})
				inBlock = false
				i += 3
			}
		} else {
			i++
		}
	}
	return ranges
}

func inlineCodeRanges(content string) []exclusion {
	var ranges []exclusion
	i := 0
	for i < len(content) {
		backtickRun := countBackticks(content, i)
		if backtickRun == 0 {
			i++
			continue
		}
		opener := i
		i += backtickRun
		closeIdx := findClosingBackticks(content, i, backtickRun)
		if closeIdx < 0 {
			continue
		}
		ranges = append(ranges, exclusion{start: opener, end: closeIdx + backtickRun})
		i = closeIdx + backtickRun
	}
	return ranges
}

func countBackticks(s string, pos int) int {
	count := 0
	for pos+count < len(s) && s[pos+count] == '`' {
		count++
	}
	return count
}

func findClosingBackticks(s string, start int, runLen int) int {
	for i := start; i <= len(s)-runLen; i++ {
		if s[i] == '`' {
			found := countBackticks(s, i)
			if found == runLen {
				return i
			}
		}
	}
	return -1
}

func htmlCommentRanges(content string) []exclusion {
	var ranges []exclusion
	commentOpen := "<!--"
	commentClose := "-->"
	i := 0
	for i <= len(content)-len(commentOpen) {
		if strings.HasPrefix(content[i:], commentOpen) {
			start := i
			closeIdx := strings.Index(content[i:], commentClose)
			if closeIdx < 0 {
				break
			}
			end := i + closeIdx + len(commentClose)
			ranges = append(ranges, exclusion{start: start, end: end})
			i = end
		} else {
			i++
		}
	}
	return ranges
}
