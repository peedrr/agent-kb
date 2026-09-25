// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package frontmatter

import (
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/template"
)

func TestParse(t *testing.T) {
	t.Run("valid YAML frontmatter with type + title + other fields", func(t *testing.T) {
		input := `---
type: note
title: My Note
tags:
  - go
  - testing
summary: A test note
---
`
		fm, body, err := Parse([]byte(input))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if fm.Type != "note" {
			t.Errorf("Type = %q, want %q", fm.Type, "note")
		}
		if fm.Title != "My Note" {
			t.Errorf("Title = %q, want %q", fm.Title, "My Note")
		}
		if len(fm.Fields) == 0 {
			t.Fatal("Fields is empty, want non-empty")
		}
		tags, ok := fm.Fields["tags"]
		if !ok {
			t.Error("Fields missing 'tags' key")
		}
		tagSlice, ok := tags.([]any)
		if !ok {
			t.Errorf("tags = %T, want []any", tags)
		} else if len(tagSlice) != 2 {
			t.Errorf("len(tags) = %d, want 2", len(tagSlice))
		}
		summary, ok := fm.Fields["summary"]
		if !ok {
			t.Error("Fields missing 'summary' key")
		}
		if s, ok := summary.(string); !ok || s != "A test note" {
			t.Errorf("summary = %v, want %q", summary, "A test note")
		}
		// type and title should NOT be in Fields
		if _, ok := fm.Fields["type"]; ok {
			t.Error("Fields should not contain 'type'")
		}
		if _, ok := fm.Fields["title"]; ok {
			t.Error("Fields should not contain 'title'")
		}
		_ = body // body tested separately
	})

	t.Run("body content preserved after delimiters", func(t *testing.T) {
		input := "---\ntype: note\ntitle: Test\n---\nThis is the body content.\n"
		fm, body, err := Parse([]byte(input))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		_ = fm
		if strings.TrimSpace(string(body)) != "This is the body content." {
			t.Errorf("body = %q, want %q", string(body), "This is the body content.")
		}
	})

	t.Run("multi-line body content preserved", func(t *testing.T) {
		input := "---\ntype: note\ntitle: Test\n---\nLine 1\nLine 2\nLine 3\n"
		fm, body, err := Parse([]byte(input))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		_ = fm
		expected := "Line 1\nLine 2\nLine 3\n"
		if string(body) != expected {
			t.Errorf("body = %q, want %q", string(body), expected)
		}
	})

	t.Run("error on missing frontmatter entirely", func(t *testing.T) {
		input := []byte("Just some content without frontmatter")
		_, _, err := Parse(input)
		if err == nil {
			t.Fatal("expected error for missing frontmatter, got nil")
		}
	})

	t.Run("error on empty frontmatter", func(t *testing.T) {
		input := []byte("---\n---\nSome body content")
		_, _, err := Parse(input)
		if err == nil {
			t.Fatal("expected error for empty frontmatter, got nil")
		}
	})

	t.Run("error on malformed YAML", func(t *testing.T) {
		input := []byte("---\ntype: [broken yaml {{{\n---\nbody")
		_, _, err := Parse(input)
		if err == nil {
			t.Fatal("expected error for malformed YAML, got nil")
		}
	})
}

func TestValidateType(t *testing.T) {
	templates := map[string]template.Template{
		"note": {Name: "note", Dir: "notes"},
		"adr":  {Name: "adr", Dir: "decisions"},
	}

	t.Run("error on empty type", func(t *testing.T) {
		fm := &ParsedFrontmatter{Type: "", Title: "Test", Fields: map[string]any{}}
		err := ValidateType(fm, templates)
		if err == nil {
			t.Fatal("expected error for empty type, got nil")
		}
		if err.Error() != "missing required field 'type' in frontmatter" {
			t.Errorf("error = %q, want %q", err.Error(), "missing required field 'type' in frontmatter")
		}
	})

	t.Run("error on unknown type", func(t *testing.T) {
		fm := &ParsedFrontmatter{Type: "unknown", Title: "Test", Fields: map[string]any{}}
		err := ValidateType(fm, templates)
		if err == nil {
			t.Fatal("expected error for unknown type, got nil")
		}
		expected := "unknown type 'unknown'. Author .agent-kb/templates/unknown.yaml (see the kb-management skill's TEMPLATE.md) or copy a showcase with `akb template list --examples`"
		if err.Error() != expected {
			t.Errorf("error = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("success on known type", func(t *testing.T) {
		fm := &ParsedFrontmatter{Type: "note", Title: "Test", Fields: map[string]any{}}
		err := ValidateType(fm, templates)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestValidateTitle(t *testing.T) {
	t.Run("error on empty title", func(t *testing.T) {
		fm := &ParsedFrontmatter{Type: "note", Title: "", Fields: map[string]any{}}
		err := ValidateTitle(fm)
		if err == nil {
			t.Fatal("expected error for empty title, got nil")
		}
		if err.Error() != "missing required field 'title' in frontmatter" {
			t.Errorf("error = %q, want %q", err.Error(), "missing required field 'title' in frontmatter")
		}
	})

	t.Run("success on present title", func(t *testing.T) {
		fm := &ParsedFrontmatter{Type: "note", Title: "My Note", Fields: map[string]any{}}
		err := ValidateTitle(fm)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestIsDraft(t *testing.T) {
	t.Run("nil fields returns true", func(t *testing.T) {
		if !IsDraft(nil) {
			t.Error("expected IsDraft(nil) = true")
		}
	})

	t.Run("missing key returns true", func(t *testing.T) {
		if !IsDraft(map[string]any{}) {
			t.Error("expected IsDraft(empty map) = true")
		}
	})

	t.Run("explicit true bool returns true", func(t *testing.T) {
		if !IsDraft(map[string]any{"is_draft": true}) {
			t.Error("expected IsDraft(true) = true")
		}
	})

	t.Run("explicit false bool returns false", func(t *testing.T) {
		if IsDraft(map[string]any{"is_draft": false}) {
			t.Error("expected IsDraft(false) = false")
		}
	})

	t.Run("string true returns true", func(t *testing.T) {
		if !IsDraft(map[string]any{"is_draft": "true"}) {
			t.Error("expected IsDraft('true') = true")
		}
	})

	t.Run("string false returns false", func(t *testing.T) {
		if IsDraft(map[string]any{"is_draft": "false"}) {
			t.Error("expected IsDraft('false') = false")
		}
	})

	t.Run("other fields ignored", func(t *testing.T) {
		if !IsDraft(map[string]any{"tags": []string{"go"}}) {
			t.Error("expected IsDraft with only tags = true")
		}
	})
}
