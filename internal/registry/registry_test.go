package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("returns nil for non-existent file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "nonexistent.yaml")

		entries, err := Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if entries != nil {
			t.Fatalf("expected nil for non-existent file, got %v", entries)
		}
	})

	t.Run("parses existing registry file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "registry.yaml")

		content := `- name: test-kb
  path: /home/user/kb/test
  created: "2024-01-01"
`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write test file: %v", err)
		}

		entries, err := Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].Name != "test-kb" {
			t.Fatalf("expected name 'test-kb', got %q", entries[0].Name)
		}
	})
}

func TestSave(t *testing.T) {
	t.Run("creates file and parent directories", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "registry.yaml")

		entries := []Entry{
			{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"},
		}

		err := Save(path, entries)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("registry file was not created")
		}
	})

	t.Run("writes valid YAML", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "registry.yaml")

		entries := []Entry{
			{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"},
			{Name: "kb2", Path: "/path/kb2", Created: "2024-01-02"},
		}

		err := Save(path, entries)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := Load(path)
		if err != nil {
			t.Fatalf("Load after Save failed: %v", err)
		}
		if len(loaded) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(loaded))
		}
	})
}

func TestAddEntry(t *testing.T) {
	t.Run("adds entry to registry", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)
		defer os.Setenv("HOME", origHome)

		regPath, err := RegistryPath()
		if err != nil {
			t.Fatalf("RegistryPath failed: %v", err)
		}

		entry := Entry{Name: "new-kb", Path: "/path/new", Created: "2024-01-01"}

		err = AddEntry(entry)
		if err != nil {
			t.Fatalf("AddEntry failed: %v", err)
		}

		entries, err := Load(regPath)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].Name != "new-kb" {
			t.Fatalf("expected name 'new-kb', got %q", entries[0].Name)
		}
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)
		defer os.Setenv("HOME", origHome)

		regPath, err := RegistryPath()
		if err != nil {
			t.Fatalf("RegistryPath failed: %v", err)
		}

		existing := []Entry{
			{Name: "existing-kb", Path: "/path/existing", Created: "2024-01-01"},
		}
		if err := Save(regPath, existing); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		err = AddEntry(Entry{Name: "existing-kb", Path: "/path/new", Created: "2024-01-02"})
		if err == nil {
			t.Fatal("expected error for duplicate name")
		}
	})
}

func TestFindByName(t *testing.T) {
	t.Run("finds existing entry", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)
		defer os.Setenv("HOME", origHome)

		regPath, err := RegistryPath()
		if err != nil {
			t.Fatalf("RegistryPath failed: %v", err)
		}

		entries := []Entry{
			{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"},
			{Name: "kb2", Path: "/path/kb2", Created: "2024-01-02"},
		}
		if err := Save(regPath, entries); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		found, err := FindByName("kb2")
		if err != nil {
			t.Fatalf("FindByName failed: %v", err)
		}
		if found == nil {
			t.Fatal("expected to find kb2")
		}
		if found.Path != "/path/kb2" {
			t.Fatalf("expected path '/path/kb2', got %q", found.Path)
		}
	})

	t.Run("returns nil for non-existent name", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)
		defer os.Setenv("HOME", origHome)

		regPath, err := RegistryPath()
		if err != nil {
			t.Fatalf("RegistryPath failed: %v", err)
		}

		entries := []Entry{
			{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"},
		}
		if err := Save(regPath, entries); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		found, err := FindByName("nonexistent")
		if err != nil {
			t.Fatalf("FindByName failed: %v", err)
		}
		if found != nil {
			t.Fatalf("expected nil for non-existent name, got %v", found)
		}
	})
}
