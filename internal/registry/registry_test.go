package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("returns empty registry for non-existent file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "nonexistent.yaml")

		reg, err := Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if reg == nil {
			t.Fatal("expected non-nil registry for non-existent file")
		}
		if reg.Default != "" {
			t.Fatalf("expected empty default, got %q", reg.Default)
		}
		if reg.Entries != nil {
			t.Fatalf("expected nil entries, got %v", reg.Entries)
		}
	})

	t.Run("parses new format registry file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "registry.yaml")

		content := `default: test-kb
entries:
  - name: test-kb
    path: /home/user/kb/test
    created: "2024-01-01"
`
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatalf("write test file: %v", err)
		}

		reg, err := Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if reg.Default != "test-kb" {
			t.Fatalf("expected default 'test-kb', got %q", reg.Default)
		}
		if len(reg.Entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(reg.Entries))
		}
		if reg.Entries[0].Name != "test-kb" {
			t.Fatalf("expected name 'test-kb', got %q", reg.Entries[0].Name)
		}
	})

	t.Run("migrates old flat format", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "registry.yaml")

		content := `- name: test-kb
  path: /home/user/kb/test
  created: "2024-01-01"
`
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatalf("write test file: %v", err)
		}

		reg, err := Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if reg.Default != "" {
			t.Fatalf("expected empty default after migration, got %q", reg.Default)
		}
		if len(reg.Entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(reg.Entries))
		}
		if reg.Entries[0].Name != "test-kb" {
			t.Fatalf("expected name 'test-kb', got %q", reg.Entries[0].Name)
		}
	})
}

func TestSave(t *testing.T) {
	t.Run("creates file and parent directories", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "registry.yaml")

		reg := &Registry{
			Default: "kb1",
			Entries: []Entry{{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"}},
		}

		err := Save(path, reg)
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

		reg := &Registry{
			Entries: []Entry{
				{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"},
				{Name: "kb2", Path: "/path/kb2", Created: "2024-01-02"},
			},
		}

		err := Save(path, reg)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := Load(path)
		if err != nil {
			t.Fatalf("Load after Save failed: %v", err)
		}
		if len(loaded.Entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(loaded.Entries))
		}
	})
}

func TestAddEntry(t *testing.T) {
	t.Run("adds entry to registry", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)            //nolint:errcheck,gosec // test setup — failure is non-fatal
		defer os.Setenv("HOME", origHome) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

		regPath, err := Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		entry := Entry{Name: "new-kb", Path: "/path/new", Created: "2024-01-01"}

		err = AddEntry(entry)
		if err != nil {
			t.Fatalf("AddEntry failed: %v", err)
		}

		reg, err := Load(regPath)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if len(reg.Entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(reg.Entries))
		}
		if reg.Entries[0].Name != "new-kb" {
			t.Fatalf("expected name 'new-kb', got %q", reg.Entries[0].Name)
		}
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)            //nolint:errcheck,gosec // test setup — failure is non-fatal
		defer os.Setenv("HOME", origHome) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

		regPath, err := Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		existing := &Registry{
			Entries: []Entry{{Name: "existing-kb", Path: "/path/existing", Created: "2024-01-01"}},
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
		os.Setenv("HOME", dir)            //nolint:errcheck,gosec // test setup — failure is non-fatal
		defer os.Setenv("HOME", origHome) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

		regPath, err := Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		reg := &Registry{
			Entries: []Entry{
				{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"},
				{Name: "kb2", Path: "/path/kb2", Created: "2024-01-02"},
			},
		}
		if err := Save(regPath, reg); err != nil {
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
		os.Setenv("HOME", dir)            //nolint:errcheck,gosec // test setup — failure is non-fatal
		defer os.Setenv("HOME", origHome) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

		regPath, err := Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		reg := &Registry{
			Entries: []Entry{{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"}},
		}
		if err := Save(regPath, reg); err != nil {
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

func TestSetDefault(t *testing.T) {
	t.Run("sets default to existing entry", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)            //nolint:errcheck,gosec // test setup — failure is non-fatal
		defer os.Setenv("HOME", origHome) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

		regPath, err := Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		reg := &Registry{
			Entries: []Entry{
				{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"},
			},
		}
		if err := Save(regPath, reg); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		if err := SetDefault("kb1"); err != nil {
			t.Fatalf("SetDefault failed: %v", err)
		}

		loaded, err := Load(regPath)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.Default != "kb1" {
			t.Fatalf("expected default 'kb1', got %q", loaded.Default)
		}
	})

	t.Run("rejects non-existent name", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)            //nolint:errcheck,gosec // test setup — failure is non-fatal
		defer os.Setenv("HOME", origHome) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

		regPath, err := Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		reg := &Registry{
			Entries: []Entry{{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"}},
		}
		if err := Save(regPath, reg); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		err = SetDefault("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-existent name")
		}
	})
}

func TestGetDefault(t *testing.T) {
	t.Run("returns default entry", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)            //nolint:errcheck,gosec // test setup — failure is non-fatal
		defer os.Setenv("HOME", origHome) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

		regPath, err := Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		reg := &Registry{
			Default: "kb1",
			Entries: []Entry{
				{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"},
				{Name: "kb2", Path: "/path/kb2", Created: "2024-01-02"},
			},
		}
		if err := Save(regPath, reg); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		entry, err := GetDefault()
		if err != nil {
			t.Fatalf("GetDefault failed: %v", err)
		}
		if entry.Name != "kb1" {
			t.Fatalf("expected name 'kb1', got %q", entry.Name)
		}
		if entry.Path != "/path/kb1" {
			t.Fatalf("expected path '/path/kb1', got %q", entry.Path)
		}
	})

	t.Run("returns error when no default set", func(t *testing.T) {
		dir := t.TempDir()
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", dir)            //nolint:errcheck,gosec // test setup — failure is non-fatal
		defer os.Setenv("HOME", origHome) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

		regPath, err := Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		reg := &Registry{
			Entries: []Entry{{Name: "kb1", Path: "/path/kb1", Created: "2024-01-01"}},
		}
		if err := Save(regPath, reg); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		_, err = GetDefault()
		if err == nil {
			t.Fatal("expected error when no default set")
		}
	})
}
