// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplates(t *testing.T) {
	t.Run("loads valid new-format templates", func(t *testing.T) {
		dir := t.TempDir()
		noteYaml := `name: note
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
    tags:
      type: list
      required: true
    created:
      type: string
      required: false
validations:
  - id: v1
    rule: "size(title) > 0"
    expect: "title must not be empty"
lint_rules:
  - id: l1
    rule: "summary.length > 10"
    severity: warning
    expect: "summary should be longer than 10 characters"
`
		err := os.WriteFile(filepath.Join(dir, "note.yaml"), []byte(noteYaml), 0600)
		if err != nil {
			t.Fatal(err)
		}

		templates, err := LoadTemplates(dir)
		if err != nil {
			t.Fatalf("LoadTemplates failed: %v", err)
		}
		if len(templates) != 1 {
			t.Fatalf("len(templates) = %d, want 1", len(templates))
		}
		note, ok := templates["note"]
		if !ok {
			t.Fatal("expected template with name 'note'")
		}
		if note.Name != "note" {
			t.Errorf("Name = %q, want %q", note.Name, "note")
		}
		if note.Description != "General knowledge note" {
			t.Errorf("Description = %q, want %q", note.Description, "General knowledge note")
		}
		if note.Dir != "notes" {
			t.Errorf("Dir = %q, want %q", note.Dir, "notes")
		}
		if len(note.Schema.Frontmatter) != 5 {
			t.Errorf("Schema.Frontmatter len = %d, want 5", len(note.Schema.Frontmatter))
		}
		titleSchema, ok := note.Schema.Frontmatter["title"]
		if !ok {
			t.Fatal("expected 'title' in schema")
		}
		if titleSchema.Type != "string" {
			t.Errorf("title.Type = %q, want %q", titleSchema.Type, "string")
		}
		if !titleSchema.Required {
			t.Error("title.Required = false, want true")
		}
		createdSchema, ok := note.Schema.Frontmatter["created"]
		if !ok {
			t.Fatal("expected 'created' in schema")
		}
		if createdSchema.Required {
			t.Error("created.Required = true, want false")
		}
		if len(note.Validations) != 1 {
			t.Errorf("Validations len = %d, want 1", len(note.Validations))
		}
		if note.Validations[0].ID != "v1" {
			t.Errorf("Validation[0].ID = %q, want %q", note.Validations[0].ID, "v1")
		}
		if len(note.LintRules) != 1 {
			t.Errorf("LintRules len = %d, want 1", len(note.LintRules))
		}
		if note.LintRules[0].Severity != "warning" {
			t.Errorf("LintRule[0].Severity = %q, want %q", note.LintRules[0].Severity, "warning")
		}
	})

	t.Run("loads template with enum field", func(t *testing.T) {
		dir := t.TempDir()
		adrYaml := `name: adr
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
        - deprecated
`
		err := os.WriteFile(filepath.Join(dir, "adr.yaml"), []byte(adrYaml), 0600)
		if err != nil {
			t.Fatal(err)
		}

		templates, err := LoadTemplates(dir)
		if err != nil {
			t.Fatalf("LoadTemplates failed: %v", err)
		}
		adr, ok := templates["adr"]
		if !ok {
			t.Fatal("expected template with name 'adr'")
		}
		statusSchema, ok := adr.Schema.Frontmatter["status"]
		if !ok {
			t.Fatal("expected 'status' in schema")
		}
		if len(statusSchema.Enum) != 3 {
			t.Fatalf("Enum length = %d, want 3", len(statusSchema.Enum))
		}
		if statusSchema.Enum[0] != "proposed" || statusSchema.Enum[1] != "accepted" || statusSchema.Enum[2] != "deprecated" {
			t.Errorf("Enum = %v, want [proposed accepted deprecated]", statusSchema.Enum)
		}
	})

	t.Run("rejects old-format template with required", func(t *testing.T) {
		dir := t.TempDir()
		oldYaml := `name: note
dir: notes
required:
  - title
`
		err := os.WriteFile(filepath.Join(dir, "old.yaml"), []byte(oldYaml), 0600)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadTemplates(dir)
		if err == nil {
			t.Fatal("expected error for old-format template")
		}
		if !contains(err.Error(), "Template format has changed") {
			t.Errorf("error should mention format change, got: %v", err)
		}
	})

	t.Run("rejects old-format template with optional", func(t *testing.T) {
		dir := t.TempDir()
		oldYaml := `name: note
dir: notes
optional:
  - created
`
		err := os.WriteFile(filepath.Join(dir, "old.yaml"), []byte(oldYaml), 0600)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadTemplates(dir)
		if err == nil {
			t.Fatal("expected error for old-format template")
		}
		if !contains(err.Error(), "Template format has changed") {
			t.Errorf("error should mention format change, got: %v", err)
		}
	})

	t.Run("rejects old-format template with body", func(t *testing.T) {
		dir := t.TempDir()
		oldYaml := `name: note
dir: notes
body: |
  # {{.Title}}
`
		err := os.WriteFile(filepath.Join(dir, "old.yaml"), []byte(oldYaml), 0600)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadTemplates(dir)
		if err == nil {
			t.Fatal("expected error for old-format template")
		}
		if !contains(err.Error(), "Template format has changed") {
			t.Errorf("error should mention format change, got: %v", err)
		}
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		dir := t.TempDir()
		content := `name: note
dir: notes
schema:
  frontmatter:
    title:
      type: string
      required: true
`
		err := os.WriteFile(filepath.Join(dir, "note.yaml"), []byte(content), 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(filepath.Join(dir, "note2.yaml"), []byte(content), 0600)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadTemplates(dir)
		if err == nil {
			t.Fatal("expected error for duplicate name")
		}
		if !contains(err.Error(), "note.yaml") || !contains(err.Error(), "note2.yaml") {
			t.Errorf("error should mention both filenames, got: %v", err)
		}
	})

	t.Run("rejects missing name", func(t *testing.T) {
		dir := t.TempDir()
		content := `dir: notes
schema:
  frontmatter: {}
`
		err := os.WriteFile(filepath.Join(dir, "noname.yaml"), []byte(content), 0600)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadTemplates(dir)
		if err == nil {
			t.Fatal("expected error for missing name")
		}
		if !contains(err.Error(), "noname.yaml") {
			t.Errorf("error should mention filename, got: %v", err)
		}
	})

	t.Run("rejects malformed YAML", func(t *testing.T) {
		dir := t.TempDir()
		content := `name: [broken yaml {{{`
		err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte(content), 0600)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadTemplates(dir)
		if err == nil {
			t.Fatal("expected error for malformed YAML")
		}
		if !contains(err.Error(), "bad.yaml") {
			t.Errorf("error should mention filename, got: %v", err)
		}
	})

	t.Run("returns empty map for nonexistent directory", func(t *testing.T) {
		result, err := LoadTemplates("/nonexistent/path/templates")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty map, got: %v", result)
		}
	})
}

func TestCopyDefaults(t *testing.T) {
	t.Run("copies embedded templates to target", func(t *testing.T) {
		dir := t.TempDir()
		err := CopyDefaults(dir)
		if err != nil {
			t.Fatalf("CopyDefaults failed: %v", err)
		}

		notePath := filepath.Join(dir, "note.yaml")
		data, err := os.ReadFile(notePath) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("note.yaml not found: %v", err)
		}
		if len(data) == 0 {
			t.Error("note.yaml is empty")
		}

		adrPath := filepath.Join(dir, "adr.yaml")
		data, err = os.ReadFile(adrPath) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("adr.yaml not found: %v", err)
		}
		if len(data) == 0 {
			t.Error("adr.yaml is empty")
		}
	})

	t.Run("skips existing files", func(t *testing.T) {
		dir := t.TempDir()
		existingContent := "name: my-custom-note\nschema:\n  frontmatter: {}\n"
		err := os.WriteFile(filepath.Join(dir, "note.yaml"), []byte(existingContent), 0600)
		if err != nil {
			t.Fatal(err)
		}

		err = CopyDefaults(dir)
		if err != nil {
			t.Fatalf("CopyDefaults failed: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, "note.yaml")) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != existingContent {
			t.Error("existing file should not be overwritten")
		}

		_, err = os.Stat(filepath.Join(dir, "adr.yaml"))
		if err != nil {
			t.Error("adr.yaml should have been copied")
		}
	})

	t.Run("creates target directory if needed", func(t *testing.T) {
		dir := t.TempDir()
		nestedDir := filepath.Join(dir, "templates", "defaults")
		err := CopyDefaults(nestedDir)
		if err != nil {
			t.Fatalf("CopyDefaults failed: %v", err)
		}
		if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
			t.Error("target directory should have been created")
		}
	})

	t.Run("copies .md files alongside .yaml files", func(t *testing.T) {
		dir := t.TempDir()
		err := CopyDefaults(dir)
		if err != nil {
			t.Fatalf("CopyDefaults failed: %v", err)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}
		if len(entries) == 0 {
			t.Error("expected files to be copied")
		}
	})
}

func TestEmbeddedTemplates(t *testing.T) {
	t.Run("embedded templates are valid new-format", func(t *testing.T) {
		dir := t.TempDir()
		if err := CopyDefaults(dir); err != nil {
			t.Fatalf("CopyDefaults failed: %v", err)
		}

		templates, err := LoadTemplates(dir)
		if err != nil {
			t.Fatalf("embedded templates invalid: %v", err)
		}
		if len(templates) != 2 {
			t.Fatalf("expected 2 embedded templates, got %d", len(templates))
		}

		note, ok := templates["note"]
		if !ok {
			t.Fatal("expected 'note' template")
		}
		if note.Dir != "notes" {
			t.Errorf("note.Dir = %q, want %q", note.Dir, "notes")
		}
		if len(note.Schema.Frontmatter) == 0 {
			t.Error("note schema should have frontmatter fields")
		}

		adr, ok := templates["adr"]
		if !ok {
			t.Fatal("expected 'adr' template")
		}
		if adr.Dir != "decisions" {
			t.Errorf("adr.Dir = %q, want %q", adr.Dir, "decisions")
		}

		statusSchema, ok := adr.Schema.Frontmatter["status"]
		if !ok {
			t.Fatal("expected 'status' in adr schema")
		}
		if len(statusSchema.Enum) == 0 {
			t.Error("status field should have enum values")
		}
		if statusSchema.Enum[0] != "proposed" {
			t.Errorf("status.Enum[0] = %q, want %q", statusSchema.Enum[0], "proposed")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
