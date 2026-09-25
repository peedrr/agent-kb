// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cel

import (
	"context"
	"testing"
	"time"

	"github.com/google/cel-go/common/types"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/markdown"
)

type mockStore struct {
	data   map[string][]byte
	exists map[string]bool
}

func (m *mockStore) Read(_ context.Context, path string) ([]byte, error) {
	return m.data[path], nil
}

func (m *mockStore) Exists(_ context.Context, path string) (bool, error) {
	return m.exists[path], nil
}

func (m *mockStore) Write(_ context.Context, _ string, _ []byte) error     { return nil }
func (m *mockStore) Delete(_ context.Context, _ string) error              { return nil }
func (m *mockStore) List(_ context.Context, _, _ string) ([]string, error) { return nil, nil }

func TestBuildPage(t *testing.T) {
	mdSource := []byte(`---
type: note
title: Test Page
tags:
  - go
  - testing
---

# Heading One

Some body text with a [link](https://example.com).

` + "```go\nfmt.Println(\"hello\")\n```\n" + `

## Heading Two

^[inferred]

<!-- olw-auto: status=draft -->
`)

	fm, body, err := frontmatter.Parse(mdSource)
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(mdSource))

	page := BuildPage("kb/notes/test.md", fm, body, doc, mdSource)

	fileMap, ok := page["file"].(map[string]any)
	if !ok {
		t.Fatalf("file map missing")
	}
	if fileMap["path"] != "kb/notes/test.md" {
		t.Errorf("file.path = %v, want %v", fileMap["path"], "kb/notes/test.md")
	}
	if fileMap["name"] != "test.md" {
		t.Errorf("file.name = %v, want %v", fileMap["name"], "test.md")
	}
	if fileMap["dir"] != "kb/notes" {
		t.Errorf("file.dir = %v, want %v", fileMap["dir"], "kb/notes")
	}

	fmMap, ok := page["frontmatter"].(map[string]any)
	if !ok {
		t.Fatalf("frontmatter map missing")
	}
	if fmMap["type"] != "note" {
		t.Errorf("frontmatter.type = %v, want %v", fmMap["type"], "note")
	}
	if fmMap["title"] != "Test Page" {
		t.Errorf("frontmatter.title = %v, want %v", fmMap["title"], "Test Page")
	}

	tags, ok := fmMap["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Errorf("frontmatter.tags = %v, want 2 items", fmMap["tags"])
	}

	contentMap, ok := page["content"].(map[string]any)
	if !ok {
		t.Fatalf("content map missing")
	}
	if contentMap["raw"] != string(body) {
		t.Errorf("content.raw mismatch")
	}
	if contentMap["char_count"] != len(body) {
		t.Errorf("content.char_count = %v, want %v", contentMap["char_count"], len(body))
	}
	wordCount := len(contentMap["raw"].(string))
	_ = wordCount

	astMap, ok := page["ast"].(map[string]any)
	if !ok {
		t.Fatalf("ast map missing")
	}
	headings, ok := astMap["headings"].([]map[string]any)
	if !ok {
		t.Fatalf("ast.headings type mismatch")
	}
	if len(headings) != 2 {
		t.Errorf("ast.headings len = %d, want 2", len(headings))
	}
	if headings[0]["level"] != 1 || headings[0]["text"] != "Heading One" {
		t.Errorf("ast.headings[0] = %+v, want level=1 text=Heading One", headings[0])
	}
	if headings[1]["level"] != 2 || headings[1]["text"] != "Heading Two" {
		t.Errorf("ast.headings[1] = %+v, want level=2 text=Heading Two", headings[1])
	}

	links, ok := astMap["links"].([]map[string]any)
	if !ok {
		t.Fatalf("ast.links type mismatch")
	}
	if len(links) != 1 {
		t.Errorf("ast.links len = %d, want 1", len(links))
	}
	if links[0]["target"] != "https://example.com" {
		t.Errorf("ast.links[0].target = %v, want %v", links[0]["target"], "https://example.com")
	}

	codeBlocks, ok := astMap["code_blocks"].([]map[string]any)
	if !ok {
		t.Fatalf("ast.code_blocks type mismatch")
	}
	if len(codeBlocks) != 1 {
		t.Errorf("ast.code_blocks len = %d, want 1", len(codeBlocks))
	}
	if codeBlocks[0]["language"] != "go" {
		t.Errorf("ast.code_blocks[0].language = %v, want %v", codeBlocks[0]["language"], "go")
	}

	akbMap, ok := page["akb"].(map[string]any)
	if !ok {
		t.Fatalf("akb map missing")
	}
	markers, ok := akbMap["provenance_markers"].([]any)
	if !ok {
		t.Fatalf("akb.provenance_markers type mismatch, got %T", akbMap["provenance_markers"])
	}
	if len(markers) != 1 {
		t.Errorf("akb.provenance_markers len = %d, want 1", len(markers))
	}

	annotations, ok := akbMap["annotations"].([]any)
	if !ok {
		t.Fatalf("akb.annotations type mismatch, got %T", akbMap["annotations"])
	}
	if len(annotations) != 1 {
		t.Errorf("akb.annotations len = %d, want 1", len(annotations))
	}
}

func TestBuildPageAnnotationExclusions(t *testing.T) {
	mdSource := []byte(`---
type: note
title: Annotation Exclusions
---

` + "```\n<!-- olw-auto: in_fence=true -->\n```\n" + `
Use ` + "`<!-- olw-auto: in_inline_code=true -->`" + ` in prose.

<!-- olw-auto: real=true -->
`)

	fm, body, err := frontmatter.Parse(mdSource)
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(mdSource))

	page := BuildPage("kb/notes/annotations.md", fm, body, doc, mdSource)

	akbMap, ok := page["akb"].(map[string]any)
	if !ok {
		t.Fatalf("akb map missing")
	}
	annotations, ok := akbMap["annotations"].([]any)
	if !ok {
		t.Fatalf("akb.annotations type mismatch, got %T", akbMap["annotations"])
	}
	if len(annotations) != 1 {
		t.Fatalf("akb.annotations len = %d, want 1", len(annotations))
	}

	ann, ok := annotations[0].(map[string]any)
	if !ok {
		t.Fatalf("annotation type mismatch, got %T", annotations[0])
	}
	fields, ok := ann["fields"].(map[string]string)
	if !ok {
		t.Fatalf("fields type mismatch, got %T", ann["fields"])
	}
	if fields["real"] != "true" {
		t.Errorf("fields = %v, want real=true", fields)
	}
}

func TestConvertDateFieldNonDateStringsUnchanged(t *testing.T) {
	for _, value := range []string{"note", "draft", "not a date", "2026-13-40", "10-05-2026", "2026-05-10T08:30", ""} {
		if got := convertDateField(value); got != value {
			t.Errorf("convertDateField(%q) = %v (%T), want unchanged", value, got, got)
		}
	}

	if got := convertDateField(42); got != 42 {
		t.Errorf("convertDateField(42) = %v (%T), want 42 unchanged", got, got)
	}
	if got := convertDateField(nil); got != nil {
		t.Errorf("convertDateField(nil) = %v, want nil", got)
	}
}

func TestBuildPageCustomDateFieldsInTimestampRule(t *testing.T) {
	mdSource := []byte(`---
type: note
title: Custom Date Fields
valid_from: 2026-05-10
starts_at: 2026-05-10T08:30:00Z
status: draft
summary: not a date at all
---

Body text.
`)

	fm, body, err := frontmatter.Parse(mdSource)
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(mdSource))

	page := BuildPage("kb/notes/dates.md", fm, body, doc, mdSource)

	fmMap, ok := page["frontmatter"].(map[string]any)
	if !ok {
		t.Fatalf("frontmatter map missing")
	}

	validFrom, ok := fmMap["valid_from"].(time.Time)
	if !ok {
		t.Fatalf("frontmatter.valid_from = %T, want time.Time", fmMap["valid_from"])
	}
	if want := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC); !validFrom.Equal(want) {
		t.Errorf("frontmatter.valid_from = %v, want %v", validFrom, want)
	}

	startsAt, ok := fmMap["starts_at"].(time.Time)
	if !ok {
		t.Fatalf("frontmatter.starts_at = %T, want time.Time", fmMap["starts_at"])
	}
	if want := time.Date(2026, 5, 10, 8, 30, 0, 0, time.UTC); !startsAt.Equal(want) {
		t.Errorf("frontmatter.starts_at = %v, want %v", startsAt, want)
	}

	if fmMap["status"] != "draft" {
		t.Errorf("frontmatter.status = %v, want draft", fmMap["status"])
	}
	if fmMap["summary"] != "not a date at all" {
		t.Errorf("frontmatter.summary = %v, want unchanged", fmMap["summary"])
	}

	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	vars := map[string]any{
		"page": page,
		"now":  time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC),
	}

	rules := []struct {
		expr string
		want bool
	}{
		{`timestamp(page.frontmatter.valid_from) <= now`, true},
		{`timestamp(page.frontmatter.starts_at) > timestamp(page.frontmatter.valid_from)`, true},
		{`timestamp(page.frontmatter.valid_from) > now`, false},
	}
	for _, rule := range rules {
		prg, err := CompileRule(env, rule.expr)
		if err != nil {
			t.Errorf("CompileRule(%s): %v", rule.expr, err)
			continue
		}
		result, err := Evaluate(context.Background(), prg, vars)
		if err != nil {
			t.Errorf("Evaluate(%s): %v", rule.expr, err)
			continue
		}
		if result != types.Bool(rule.want) {
			t.Errorf("%s = %v, want %v", rule.expr, result, rule.want)
		}
	}
}

func TestBuildPageWordCount(t *testing.T) {
	mdSource := []byte(`---
type: note
title: Word Count Test
---

one two   three
four	five
`)

	fm, body, err := frontmatter.Parse(mdSource)
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(mdSource))

	page := BuildPage("kb/test.md", fm, body, doc, mdSource)
	contentMap := page["content"].(map[string]any)

	if contentMap["word_count"] != 5 {
		t.Errorf("word_count = %v, want 5", contentMap["word_count"])
	}
	if contentMap["char_count"] != len(body) {
		t.Errorf("char_count = %v, want %d", contentMap["char_count"], len(body))
	}
}

func TestBuildPageHasOnNil(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	prg, err := CompileRule(env, "old_page != null && has(old_page.frontmatter) && has(old_page.frontmatter.type)")
	if err != nil {
		t.Fatalf("CompileRule: %v", err)
	}

	vars := map[string]any{
		"page": map[string]any{
			"frontmatter": map[string]any{"type": "note"},
		},
		"old_page": nil,
	}

	result, err := Evaluate(context.Background(), prg, vars)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result != types.False {
		t.Errorf("has(nil.frontmatter.type) = %v, want false", result)
	}
}

func TestBuildOldPageMissing(t *testing.T) {
	store := &mockStore{
		exists: map[string]bool{"kb/missing.md": false},
	}

	page, err := BuildOldPage("kb/missing.md", store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page != nil {
		t.Errorf("page = %v, want nil", page)
	}
}

func TestBuildOldPageExisting(t *testing.T) {
	raw := []byte(`---
type: adr
title: Old Page
---

# Old Heading

Body text.
`)
	store := &mockStore{
		data:   map[string][]byte{"kb/old.md": raw},
		exists: map[string]bool{"kb/old.md": true},
	}

	page, err := BuildOldPage("kb/old.md", store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page == nil {
		t.Fatalf("page is nil")
	}

	fmMap, ok := page["frontmatter"].(map[string]any)
	if !ok {
		t.Fatalf("frontmatter map missing")
	}
	if fmMap["type"] != "adr" {
		t.Errorf("frontmatter.type = %v, want adr", fmMap["type"])
	}
	if fmMap["title"] != "Old Page" {
		t.Errorf("frontmatter.title = %v, want Old Page", fmMap["title"])
	}

	fileMap := page["file"].(map[string]any)
	if fileMap["name"] != "old.md" {
		t.Errorf("file.name = %v, want old.md", fileMap["name"])
	}
	if fileMap["dir"] != "kb" {
		t.Errorf("file.dir = %v, want kb", fileMap["dir"])
	}

	astMap := page["ast"].(map[string]any)
	headings := astMap["headings"].([]map[string]any)
	if len(headings) != 1 || headings[0]["text"] != "Old Heading" {
		t.Errorf("headings = %+v, want 1 heading with text Old Heading", headings)
	}
}

func flattenLinksForTest(t *testing.T, source string) []map[string]any {
	t.Helper()

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader([]byte(source)))

	return flattenLinks(doc, []byte(source))
}

func assertLinkEntry(t *testing.T, entry map[string]any, target, text string, isWikilink bool, line int) {
	t.Helper()

	if entry["target"] != target {
		t.Errorf("target = %v, want %q", entry["target"], target)
	}
	if entry["text"] != text {
		t.Errorf("text = %v, want %q", entry["text"], text)
	}
	if entry["is_wikilink"] != isWikilink {
		t.Errorf("is_wikilink = %v, want %v", entry["is_wikilink"], isWikilink)
	}
	if entry["line"] != line {
		t.Errorf("line = %v, want %d", entry["line"], line)
	}
}

func TestFlattenLinksWikilinkForms(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantTarget string
		wantText   string
		wantLine   int
	}{
		{
			name:       "bare wikilink",
			source:     "See [[notes/page]] for context.\n",
			wantTarget: "notes/page",
			wantText:   "notes/page",
			wantLine:   1,
		},
		{
			name:       "wikilink with display text",
			source:     "See [[notes/page|the page]].\n",
			wantTarget: "notes/page",
			wantText:   "the page",
			wantLine:   1,
		},
		{
			name:       "wikilink with heading",
			source:     "See [[notes/page#Details]].\n",
			wantTarget: "notes/page",
			wantText:   "Details",
			wantLine:   1,
		},
		{
			name:       "explicit destination target is normalized",
			source:     "See [[the page]](notes/page.md).\n",
			wantTarget: "notes/page",
			wantText:   "[the page]",
			wantLine:   1,
		},
		{
			name:       "explicit destination wins over heading",
			source:     "See [[notes/page#Details]](other/page.md).\n",
			wantTarget: "other/page",
			wantText:   "[notes/page#Details]",
			wantLine:   1,
		},
		{
			name:       "angle-bracketed destination is normalized",
			source:     "See [[the page]](<notes/page.md>).\n",
			wantTarget: "notes/page",
			wantText:   "[the page]",
			wantLine:   1,
		},
		{
			name:       "relative destination is normalized",
			source:     "See [[the page]](./notes/page.md).\n",
			wantTarget: "notes/page",
			wantText:   "[the page]",
			wantLine:   1,
		},
		{
			name:       "titled destination drops the title",
			source:     "See [[the page]](notes/page.md \"title\").\n",
			wantTarget: "notes/page",
			wantText:   "[the page]",
			wantLine:   1,
		},
		{
			name:       "repeated .md affixes collapse",
			source:     "See [[the page]](notes/page.md.md).\n",
			wantTarget: "notes/page",
			wantText:   "[the page]",
			wantLine:   1,
		},
		{
			name:       "degenerate nested brackets normalize the destination",
			source:     "See [[[a]]](notes/x.md).\n",
			wantTarget: "notes/x",
			wantText:   "[[a]]",
			wantLine:   1,
		},
		{
			name:       "deeper degenerate nesting at offset 0 normalizes the destination",
			source:     "[[[[a]]]](x.md)\n",
			wantTarget: "x",
			wantText:   "[[[a]]]",
			wantLine:   1,
		},
		{
			name:       "empty destination parens stay a plain wikilink",
			source:     "See [[notes/page]]().\n",
			wantTarget: "notes/page",
			wantText:   "notes/page",
			wantLine:   1,
		},
		{
			name:       "whitespace-only destination parens stay a plain wikilink",
			source:     "See [[notes/page]](   ).\n",
			wantTarget: "notes/page",
			wantText:   "notes/page",
			wantLine:   1,
		},
		{
			name:       "angle-only destination parens stay a plain wikilink",
			source:     "See [[notes/page]](<>).\n",
			wantTarget: "notes/page",
			wantText:   "notes/page",
			wantLine:   1,
		},
		{
			name:       "unclosed destination parens stay a plain wikilink",
			source:     "See [[notes/page]](unclosed.\n",
			wantTarget: "notes/page",
			wantText:   "notes/page",
			wantLine:   1,
		},
		{
			name:       "wikilink on a later line",
			source:     "Intro text.\n\nSee [[notes/page]].\n",
			wantTarget: "notes/page",
			wantText:   "notes/page",
			wantLine:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links := flattenLinksForTest(t, tt.source)
			if len(links) != 1 {
				t.Fatalf("links = %+v, want exactly 1 entry", links)
			}
			assertLinkEntry(t, links[0], tt.wantTarget, tt.wantText, true, tt.wantLine)
		})
	}
}

// A wikilink token nested inside a markdown link label yields two entries:
// the outer goldmark link, whose start offset opens the label, and the
// wikilink token parsed from that label. Start-equality dedupe does not merge
// them because their offsets differ.
func TestFlattenLinksLinkContainingWikilink(t *testing.T) {
	source := "See [text [[f]]](dest).\n"

	links := flattenLinksForTest(t, source)
	if len(links) != 2 {
		t.Fatalf("links = %+v, want exactly 2 entries", links)
	}
	assertLinkEntry(t, links[0], "dest", "text [[f]]", false, 1)
	assertLinkEntry(t, links[1], "f", "f", true, 1)
}

// A markdown link nested inside a wikilink-shaped token yields only the inner
// goldmark link: the wikilink token regex cannot span the inner ']' of [b], so
// the token never parses as a wikilink and no wikilink entry is emitted.
func TestFlattenLinksWikilinkShapedTokenContainingMarkdownLink(t *testing.T) {
	source := "See [[a [b](c)]].\n"

	links := flattenLinksForTest(t, source)
	if len(links) != 1 {
		t.Fatalf("links = %+v, want exactly 1 entry", links)
	}
	assertLinkEntry(t, links[0], "c", "b", false, 1)
}

func TestFlattenLinksNonWikilinkLinksUnchanged(t *testing.T) {
	source := "A [plain link](https://example.com) and `[[inline code]]` and:\n\n" +
		"```\n[[fenced code]]\n```\n"

	links := flattenLinksForTest(t, source)
	if len(links) != 1 {
		t.Fatalf("links = %+v, want only the plain markdown link", links)
	}
	assertLinkEntry(t, links[0], "https://example.com", "plain link", false, 1)
}

// Goldmark sets Link.Pos() to the opening '[' of the link text, so for
// [[display]](dest) it is the wikilink token start. The merge relies on that
// to recognise the goldmark link and the wikilink as one token; this pins the
// behaviour of goldmark v1.8.2 that the containment check is built on.
func TestFlattenLinksExplicitDestinationSharesGoldmarkLinkStart(t *testing.T) {
	source := "See [[the page]](notes/page.md) for context.\n"

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader([]byte(source)))

	var goldmarkPositions []int
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if l, ok := n.(*ast.Link); ok {
			goldmarkPositions = append(goldmarkPositions, l.Pos())
		}
		return ast.WalkContinue, nil
	})

	wikilinks := markdown.ParseWikilinks(source)
	if len(goldmarkPositions) != 1 || len(wikilinks) != 1 {
		t.Fatalf("goldmark links = %v, wikilinks = %+v, want one of each", goldmarkPositions, wikilinks)
	}
	if goldmarkPositions[0] != wikilinks[0].Start {
		t.Fatalf("goldmark Link.Pos() = %d, wikilink Start = %d, want equal", goldmarkPositions[0], wikilinks[0].Start)
	}
}

// The link graph reads markdown.ParseWikilinks while page.ast.links merges that
// output with goldmark. Both readers must name the destination of
// [[[a]]](notes/x.md) in the one normalized spelling: the link graph records
// "notes/x" and page.ast.links reports that same target — never the bracket
// fragment "[a" and never goldmark's unnormalized "notes/x.md".
func TestFlattenLinksDegenerateBracketRunNamesTheDestination(t *testing.T) {
	source := "See [[[a]]](notes/x.md).\n"

	wikilinks := markdown.ParseWikilinks(source)
	if len(wikilinks) != 1 {
		t.Fatalf("wikilinks = %+v, want exactly 1", wikilinks)
	}
	if wikilinks[0].Target != "notes/x" {
		t.Errorf("link-graph target = %q, want %q", wikilinks[0].Target, "notes/x")
	}
	if wikilinks[0].Destination != "notes/x" {
		t.Errorf("link-graph destination = %q, want %q", wikilinks[0].Destination, "notes/x")
	}

	links := flattenLinksForTest(t, source)
	if len(links) != 1 {
		t.Fatalf("links = %+v, want exactly 1 entry", links)
	}
	assertLinkEntry(t, links[0], "notes/x", "[[a]]", true, 1)
}

// page.ast.links must spell a wikilink target exactly as the link graph does.
// The graph records markdown.ParseWikilinks' Target, so the CEL target for the
// same token must equal it — including .md-bearing destinations, the
// angle-bracketed form, and a destination whose title-like suffix is part of
// the destination rather than a link title.
func TestFlattenLinksMatchesLinkGraphTargetSpelling(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "explicit destination with .md", source: "[[a]](notes/x.md).\n"},
		{name: "angle-bracketed destination with .md", source: "[[a]](<notes/x.md>).\n"},
		{name: "relative destination with .md", source: "[[a]](./notes/x.md).\n"},
		{name: "titled destination with .md", source: "[[a]](notes/x.md \"title\").\n"},
		{name: "repeated .md affixes", source: "[[a]](notes/x.md.md).\n"},
		{name: "angle-bracketed destination holding a title-like suffix", source: "[[a]](<foo (bar)>).\n"},
		{name: "destination that normalizes away", source: "[[a]](.md).\n"},
		{name: "relative destination that normalizes away", source: "[[a]](./).\n"},
		{name: "empty angle-bracketed destination", source: "[[a]](<>).\n"},
		{name: "degenerate nested brackets", source: "[[[a]]](notes/x.md).\n"},
		{name: "deeper degenerate nesting at offset 0", source: "[[[[a]]]](x.md)\n"},
		{name: "pipe display with an explicit destination", source: "[[a|A]](notes/x.md).\n"},
		{name: "heading in the bracket part with an explicit destination", source: "[[a#h]](notes/x.md).\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wikilinks := markdown.ParseWikilinks(tt.source)
			if len(wikilinks) != 1 {
				t.Fatalf("wikilinks = %+v, want exactly 1", wikilinks)
			}

			links := flattenLinksForTest(t, tt.source)
			if len(links) != 1 {
				t.Fatalf("links = %+v, want exactly 1 entry", links)
			}
			if links[0]["target"] != wikilinks[0].Target {
				t.Errorf("page.ast.links target = %v, want the link graph's %q", links[0]["target"], wikilinks[0].Target)
			}
		})
	}
}

// normalizeWikilinkTarget serves the goldmark-only path, where goldmark
// reports the destination without a link title. It must not strip a
// title-like suffix, which is part of the destination, and it must leave a
// scheme-qualified destination alone.
func TestNormalizeWikilinkTarget(t *testing.T) {
	tests := []struct {
		name string
		dest string
		want string
	}{
		{name: "trailing .md is stripped", dest: "notes/x.md", want: "notes/x"},
		{name: "angle brackets and .md are stripped", dest: "<notes/x.md>", want: "notes/x"},
		{name: "leading ./ is stripped", dest: "./notes/x.md", want: "notes/x"},
		{name: "repeated affixes collapse", dest: "././notes/x.md.md", want: "notes/x"},
		{name: "a title-like suffix survives", dest: "foo (bar)", want: "foo (bar)"},
		{name: "a destination that normalizes away is empty", dest: ".md", want: ""},
		{name: "a scheme-qualified destination is left as written", dest: "https://example.com/x.md", want: "https://example.com/x.md"},
		{name: "a scheme-qualified destination without an extension is left as written", dest: "https://example.com/x", want: "https://example.com/x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeWikilinkTarget(tt.dest); got != tt.want {
				t.Errorf("normalizeWikilinkTarget(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

// A wikilink whose parenthesized destination is itself a wikilink token —
// [[a]]([[b]]) — yields exactly one entry: the inner token. The outer token is
// malformed because its destination is a bracket token rather than a path, so
// it and its goldmark link are dropped instead of overlapping the inner token
// with the bogus target "[[b]]".
func TestFlattenLinksWikilinkInsideDestination(t *testing.T) {
	t.Run("the outer token is dropped", func(t *testing.T) {
		source := "See [[a]]([[b]]) here.\n"

		links := flattenLinksForTest(t, source)
		if len(links) != 1 {
			t.Fatalf("links = %+v, want exactly 1 entry", links)
		}
		assertLinkEntry(t, links[0], "b", "b", true, 1)
	})

	t.Run("an angle-bracketed bracket-token destination drops the outer token", func(t *testing.T) {
		source := "See [[a]](<[[b]]>) here.\n"

		links := flattenLinksForTest(t, source)
		if len(links) != 1 {
			t.Fatalf("links = %+v, want exactly 1 entry", links)
		}
		assertLinkEntry(t, links[0], "b", "b", true, 1)
	})

	t.Run("later tokens are unaffected", func(t *testing.T) {
		source := "[[a]]([[b]]) and [[c]].\n"

		links := flattenLinksForTest(t, source)
		if len(links) != 2 {
			t.Fatalf("links = %+v, want exactly 2 entries", links)
		}
		assertLinkEntry(t, links[0], "b", "b", true, 1)
		assertLinkEntry(t, links[1], "c", "c", true, 1)
	})
}

// A wikilink token nested in an outer token's parenthesized destination is
// only recorded when the destination is itself a bracket token. A bracket run
// inside a quoted title or inside a path destination is part of that
// destination, so the parser records no token for it and the outer token is
// the only entry — carrying the parser-normalized spelling, the same spelling
// the link graph records.
func TestFlattenLinksWikilinkInsidePathDestination(t *testing.T) {
	t.Run("an inner token in a quoted title is not recorded", func(t *testing.T) {
		source := "See [[a]](x.md \"see [[b]]\") here.\n"

		links := flattenLinksForTest(t, source)
		if len(links) != 1 {
			t.Fatalf("links = %+v, want exactly 1 entry", links)
		}
		assertLinkEntry(t, links[0], "x", "[a]", true, 1)
	})

	t.Run("an inner token inside the destination path is not recorded", func(t *testing.T) {
		source := "See [[a]](notes/[[weird]].md) here.\n"

		links := flattenLinksForTest(t, source)
		if len(links) != 1 {
			t.Fatalf("links = %+v, want exactly 1 entry", links)
		}
		assertLinkEntry(t, links[0], "notes/[[weird]]", "[a]", true, 1)
	})

	t.Run("a bracket-token destination still drops the outer token", func(t *testing.T) {
		source := "See [[a]]([[b]]) here.\n"

		links := flattenLinksForTest(t, source)
		if len(links) != 1 {
			t.Fatalf("links = %+v, want exactly 1 entry", links)
		}
		assertLinkEntry(t, links[0], "b", "b", true, 1)
	})
}

func TestFlattenLinksSortedBySourceOffset(t *testing.T) {
	source := "Bare [[alpha]] then a [plain](https://example.com) then [[beta|B]](notes/beta.md).\n"

	links := flattenLinksForTest(t, source)
	if len(links) != 3 {
		t.Fatalf("links = %+v, want 3 entries", links)
	}

	want := []struct {
		target     string
		text       string
		isWikilink bool
	}{
		{"alpha", "alpha", true},
		{"https://example.com", "plain", false},
		{"notes/beta", "[beta|B]", true},
	}
	for i, w := range want {
		if links[i]["target"] != w.target {
			t.Errorf("links[%d].target = %v, want %q", i, links[i]["target"], w.target)
		}
		if links[i]["text"] != w.text {
			t.Errorf("links[%d].text = %v, want %q", i, links[i]["text"], w.text)
		}
		if links[i]["is_wikilink"] != w.isWikilink {
			t.Errorf("links[%d].is_wikilink = %v, want %v", i, links[i]["is_wikilink"], w.isWikilink)
		}
	}
}

func TestBuildPageWikilinkLinks(t *testing.T) {
	mdSource := []byte(`---
type: note
title: Wikilink Links
---

Bare [[alpha]] and [[beta|Beta]] and [[gamma#Section]] and [[display]](notes/delta.md).

Also a [plain link](https://example.com).
`)

	fm, body, err := frontmatter.Parse(mdSource)
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(mdSource))

	page := BuildPage("kb/notes/wikilinks.md", fm, body, doc, mdSource)

	astMap, ok := page["ast"].(map[string]any)
	if !ok {
		t.Fatalf("ast map missing")
	}
	links, ok := astMap["links"].([]map[string]any)
	if !ok {
		t.Fatalf("ast.links type mismatch")
	}
	if len(links) != 5 {
		t.Fatalf("ast.links = %+v, want 5 entries", links)
	}

	want := []struct {
		target     string
		text       string
		isWikilink bool
		line       int
	}{
		{"alpha", "alpha", true, 6},
		{"beta", "Beta", true, 6},
		{"gamma", "Section", true, 6},
		{"notes/delta", "[display]", true, 6},
		{"https://example.com", "plain link", false, 8},
	}
	for i, w := range want {
		assertLinkEntry(t, links[i], w.target, w.text, w.isWikilink, w.line)
	}
}

func TestBuildPageWikilinkLinksInCELRule(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	prg, err := CompileRule(env, `page.ast.links.exists(l, l.is_wikilink && l.target == "notes/alpha")`)
	if err != nil {
		t.Fatalf("CompileRule: %v", err)
	}

	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "bare wikilink satisfies the rule",
			body: "See [[notes/alpha]].\n",
			want: true,
		},
		{
			name: "wikilink with display text satisfies the rule",
			body: "See [[notes/alpha|alpha]].\n",
			want: true,
		},
		{
			name: "plain markdown link does not",
			body: "See [alpha](notes/alpha.md).\n",
			want: false,
		},
		{
			name: "explicit destination with .md satisfies the rule",
			body: "See [[alpha]](notes/alpha.md).\n",
			want: true,
		},
		{
			name: "angle-bracketed explicit destination satisfies the rule",
			body: "See [[alpha]](<notes/alpha.md>).\n",
			want: true,
		},
		{
			name: "explicit destination naming another page does not",
			body: "See [[alpha]](notes/beta.md).\n",
			want: false,
		},
		{
			name: "destination that normalizes away keeps the bracket target",
			body: "See [[alpha]](.md).\n",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mdSource := []byte("---\ntype: note\ntitle: Rule Page\n---\n\n" + tt.body)

			fm, body, err := frontmatter.Parse(mdSource)
			if err != nil {
				t.Fatalf("parse frontmatter: %v", err)
			}

			md := goldmark.New()
			doc := md.Parser().Parse(text.NewReader(mdSource))

			page := BuildPage("kb/notes/rule.md", fm, body, doc, mdSource)

			result, err := Evaluate(context.Background(), prg, map[string]any{"page": page})
			if err != nil {
				t.Fatalf("Evaluate: %v", err)
			}
			if result != types.Bool(tt.want) {
				t.Errorf("rule = %v, want %v", result, tt.want)
			}
		})
	}
}
