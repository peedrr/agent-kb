package template

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFieldUnmarshalYAML(t *testing.T) {
	t.Run("plain string field", func(t *testing.T) {
		input := `"title"`
		var f Field
		if err := yaml.Unmarshal([]byte(input), &f); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if f.Name != "title" {
			t.Errorf("Name = %q, want %q", f.Name, "title")
		}
		if len(f.Enum) != 0 {
			t.Errorf("Enum = %v, want empty", f.Enum)
		}
	})

	t.Run("map with enum", func(t *testing.T) {
		input := `
status:
  enum:
    - proposed
    - accepted
    - deprecated
`
		var f Field
		if err := yaml.Unmarshal([]byte(input), &f); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if f.Name != "status" {
			t.Errorf("Name = %q, want %q", f.Name, "status")
		}
		if len(f.Enum) != 3 {
			t.Fatalf("Enum length = %d, want 3", len(f.Enum))
		}
		if f.Enum[0] != "proposed" || f.Enum[1] != "accepted" || f.Enum[2] != "deprecated" {
			t.Errorf("Enum = %v, want [proposed accepted deprecated]", f.Enum)
		}
	})
}

func TestLoadTemplates(t *testing.T) {
	t.Run("loads valid templates", func(t *testing.T) {
		dir := t.TempDir()
		noteYaml := `name: note
dir: notes
required:
  - title
  - type
  - summary
  - tags
optional:
  - created
  - updated
  - sources
body: |
  # {{.Title}}
`
		err := os.WriteFile(filepath.Join(dir, "note.yaml"), []byte(noteYaml), 0644)
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
		if note.Dir != "notes" {
			t.Errorf("Dir = %q, want %q", note.Dir, "notes")
		}
		if len(note.Required) != 4 || note.Required[0].Name != "title" || note.Required[1].Name != "type" || note.Required[2].Name != "summary" || note.Required[3].Name != "tags" {
			t.Errorf("Required = %v, want [{Name:title},{Name:type},{Name:summary},{Name:tags}]", note.Required)
		}
		if len(note.Optional) != 3 || note.Optional[0].Name != "created" || note.Optional[1].Name != "updated" || note.Optional[2].Name != "sources" {
			t.Errorf("Optional = %v, want [{Name:created},{Name:updated},{Name:sources}]", note.Optional)
		}
	})

	t.Run("loads template with enum field", func(t *testing.T) {
		dir := t.TempDir()
		adrYaml := `name: adr
dir: decisions
required:
  - title
  - status:
      enum:
        - proposed
        - accepted
        - deprecated
optional:
  - context
body: |
  # {{.Title}}
`
		err := os.WriteFile(filepath.Join(dir, "adr.yaml"), []byte(adrYaml), 0644)
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
		var statusField *Field
		for i := range adr.Required {
			if adr.Required[i].Name == "status" {
				statusField = &adr.Required[i]
				break
			}
		}
		if statusField == nil {
			t.Fatal("expected required field 'status'")
		}
		if len(statusField.Enum) != 3 {
			t.Errorf("Enum length = %d, want 3", len(statusField.Enum))
		}
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		dir := t.TempDir()
		content := `name: note
dir: notes
required:
  - title
`
		err := os.WriteFile(filepath.Join(dir, "note.yaml"), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(filepath.Join(dir, "note2.yaml"), []byte(content), 0644)
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
required:
  - title
`
		err := os.WriteFile(filepath.Join(dir, "noname.yaml"), []byte(content), 0644)
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
		err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte(content), 0644)
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
		data, err := os.ReadFile(notePath)
		if err != nil {
			t.Fatalf("note.yaml not found: %v", err)
		}
		if len(data) == 0 {
			t.Error("note.yaml is empty")
		}

		adrPath := filepath.Join(dir, "adr.yaml")
		data, err = os.ReadFile(adrPath)
		if err != nil {
			t.Fatalf("adr.yaml not found: %v", err)
		}
		if len(data) == 0 {
			t.Error("adr.yaml is empty")
		}
	})

	t.Run("skips existing files", func(t *testing.T) {
		dir := t.TempDir()
		existingContent := "name: my-custom-note\ndir: custom\n"
		err := os.WriteFile(filepath.Join(dir, "note.yaml"), []byte(existingContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		err = CopyDefaults(dir)
		if err != nil {
			t.Fatalf("CopyDefaults failed: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, "note.yaml"))
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
}

func TestEmbeddedTemplates(t *testing.T) {
	t.Run("embedded templates are valid", func(t *testing.T) {
		templates, err := LoadTemplatesFromFS(DefaultTemplates)
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

		adr, ok := templates["adr"]
		if !ok {
			t.Fatal("expected 'adr' template")
		}
		if adr.Dir != "decisions" {
			t.Errorf("adr.Dir = %q, want %q", adr.Dir, "decisions")
		}

		var statusField *Field
		for i := range adr.Required {
			if adr.Required[i].Name == "status" {
				statusField = &adr.Required[i]
				break
			}
		}
		if statusField == nil {
			t.Fatal("expected 'status' required field in adr template")
		}
		if len(statusField.Enum) == 0 {
			t.Error("status field should have enum values")
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
