package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func setupTemplateTestKB(t *testing.T) string {
	t.Helper()
	kbRoot := t.TempDir()
	setupTestKBWithGit(t, kbRoot)

	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := os.MkdirAll(templatesDir, 0750); err != nil {
		t.Fatal(err)
	}

	noteYAML := `name: note
description: General knowledge note
dir: notes
schema:
  frontmatter:
    title:
      type: string
      required: true
    type:
      type: string
      required: true
    summary:
      type: string
      required: true
validations:
  - id: require_title
    rule: 'page.frontmatter.title != ""'
    requirement: Title must not be empty
    expect: frontmatter title must be non-empty
lint_rules:
  - id: note_stale
    rule: 'now - timestamp(page.frontmatter.updated) < duration("2160h")'
    severity: warning
    expect: Note has not been updated in 90 days
`
	if err := os.WriteFile(filepath.Join(templatesDir, "note.yaml"), []byte(noteYAML), 0600); err != nil {
		t.Fatal(err)
	}

	notePass := `---
type: note
title: Test Note
summary: Test summary
tags: [test]
---

# Test Note

This is a valid note.
`
	if err := os.WriteFile(filepath.Join(templatesDir, "note_pass.md"), []byte(notePass), 0600); err != nil {
		t.Fatal(err)
	}

	adrYAML := `name: adr
description: Architecture Decision Record
dir: decisions
schema:
  frontmatter:
    title:
      type: string
      required: true
    status:
      type: string
      required: true
      enum:
        - proposed
        - accepted
validations:
  - id: require_context
    rule: 'page.ast.headings.exists(h, h.level == 2 && h.text == "Context")'
    requirement: Must include H2 'Context' section
    expect: page must have a '## Context' heading
`
	if err := os.WriteFile(filepath.Join(templatesDir, "adr.yaml"), []byte(adrYAML), 0600); err != nil {
		t.Fatal(err)
	}

	adrPass := `---
type: adr
title: Test ADR
summary: Test summary
tags: [test]
status: proposed
---

# Test ADR

## Context

This is the context.
`
	if err := os.WriteFile(filepath.Join(templatesDir, "adr_pass.md"), []byte(adrPass), 0600); err != nil {
		t.Fatal(err)
	}

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origCwd) }) //nolint:errcheck,gosec // test cleanup

	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	return kbRoot
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestTemplateGet_WriterView(t *testing.T) {
	setupTemplateTestKB(t)

	output := captureStdout(t, func() {
		err := runTemplateGet(nil, []string{"note"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result map[string]any
	if err := yaml.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse writer view YAML: %v\noutput: %q", err, output)
	}

	if result["name"] != "note" {
		t.Errorf("name = %v, want %q", result["name"], "note")
	}
	if result["description"] != "General knowledge note" {
		t.Errorf("description = %v, want %q", result["description"], "General knowledge note")
	}

	schema, ok := result["schema"].(map[string]any)
	if !ok {
		t.Fatalf("schema is not a map, got %T", result["schema"])
	}
	frontmatter, ok := schema["frontmatter"].(map[string]any)
	if !ok {
		t.Fatalf("schema.frontmatter is not a map, got %T", schema["frontmatter"])
	}
	if _, ok := frontmatter["title"]; !ok {
		t.Error("schema.frontmatter should contain 'title'")
	}

	requirements, ok := result["requirements"].([]any)
	if !ok {
		t.Fatalf("requirements is not a list, got %T", result["requirements"])
	}
	if len(requirements) != 1 {
		t.Errorf("expected 1 requirement, got %d", len(requirements))
	} else if requirements[0] != "Title must not be empty" {
		t.Errorf("requirement = %v, want %q", requirements[0], "Title must not be empty")
	}

	if _, ok := result["lint_rules"]; ok {
		t.Error("writer view should not contain lint_rules")
	}
	if _, ok := result["validations"]; ok {
		t.Error("writer view should not contain validations")
	}
}

func TestTemplateGet_WriterView_NoRequirements(t *testing.T) {
	kbRoot := setupTemplateTestKB(t)

	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	yamlNoReq := `name: minimal
description: Minimal template
schema:
  frontmatter:
    title:
      type: string
      required: true
validations:
  - id: some_rule
    rule: 'true'
    expect: always passes
`
	if err := os.WriteFile(filepath.Join(templatesDir, "minimal.yaml"), []byte(yamlNoReq), 0600); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		err := runTemplateGet(nil, []string{"minimal"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var result map[string]any
	if err := yaml.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse writer view YAML: %v", err)
	}

	if _, ok := result["requirements"]; ok {
		t.Error("writer view should not contain requirements when none have requirement text")
	}
}

func TestTemplateGet_Example(t *testing.T) {
	setupTemplateTestKB(t)

	orig := templateExample
	templateExample = true
	t.Cleanup(func() { templateExample = orig })

	output := captureStdout(t, func() {
		err := runTemplateGet(nil, []string{"note"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Test Note") {
		t.Errorf("expected output to contain 'Test Note', got: %q", output)
	}
	if !strings.Contains(output, "type: note") {
		t.Errorf("expected output to contain 'type: note', got: %q", output)
	}
}

func TestTemplateGet_Full(t *testing.T) {
	setupTemplateTestKB(t)

	orig := templateFull
	templateFull = true
	t.Cleanup(func() { templateFull = orig })

	output := captureStdout(t, func() {
		err := runTemplateGet(nil, []string{"note"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "name: note") {
		t.Errorf("expected output to contain 'name: note', got: %q", output)
	}
	if !strings.Contains(output, "lint_rules:") {
		t.Errorf("expected full view to contain lint_rules, got: %q", output)
	}
	if !strings.Contains(output, "validations:") {
		t.Errorf("expected full view to contain validations, got: %q", output)
	}
	if !strings.Contains(output, "rule: '") {
		t.Errorf("expected full view to contain CEL rule expressions, got: %q", output)
	}
}

func TestTemplateGet_MissingTemplate(t *testing.T) {
	setupTemplateTestKB(t)

	err := runTemplateGet(nil, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("expected error to contain template name, got: %v", err)
	}
	if !strings.Contains(err.Error(), "akb template list") {
		t.Errorf("expected error to suggest 'akb template list', got: %v", err)
	}
}

func TestTemplateGet_MissingPassMockup(t *testing.T) {
	kbRoot := setupTemplateTestKB(t)

	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	yamlNoPass := `name: nopass
description: Template without pass mockup
schema:
  frontmatter:
    title:
      type: string
      required: true
`
	if err := os.WriteFile(filepath.Join(templatesDir, "nopass.yaml"), []byte(yamlNoPass), 0600); err != nil {
		t.Fatal(err)
	}

	orig := templateExample
	templateExample = true
	t.Cleanup(func() { templateExample = orig })

	err := runTemplateGet(nil, []string{"nopass"})
	if err == nil {
		t.Fatal("expected error for missing pass mockup, got nil")
	}
	if !strings.Contains(err.Error(), "pass mockup not found") {
		t.Errorf("expected 'pass mockup not found' error, got: %v", err)
	}
}

func TestTemplateList(t *testing.T) {
	setupTemplateTestKB(t)

	output := captureStdout(t, func() {
		err := runTemplateList(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "note:") {
		t.Errorf("expected output to contain 'note:', got: %q", output)
	}
	if !strings.Contains(output, "adr:") {
		t.Errorf("expected output to contain 'adr:', got: %q", output)
	}
	if !strings.Contains(output, "General knowledge note") {
		t.Errorf("expected output to contain note description, got: %q", output)
	}
	if !strings.Contains(output, "Architecture Decision Record") {
		t.Errorf("expected output to contain adr description, got: %q", output)
	}
}

func TestTemplateList_Empty(t *testing.T) {
	kbRoot := setupTemplateTestKB(t)

	entries, err := os.ReadDir(filepath.Join(kbRoot, ".akb", "templates"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if err := os.Remove(filepath.Join(kbRoot, ".akb", "templates", entry.Name())); err != nil {
			t.Fatal(err)
		}
	}

	output := captureStdout(t, func() {
		err := runTemplateList(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "No templates found") {
		t.Errorf("expected 'No templates found', got: %q", output)
	}
}
