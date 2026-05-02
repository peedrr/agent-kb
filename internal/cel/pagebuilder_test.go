package cel

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
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

func (m *mockStore) Write(_ context.Context, _ string, _ []byte) error   { return nil }
func (m *mockStore) Delete(_ context.Context, _ string) error            { return nil }
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
