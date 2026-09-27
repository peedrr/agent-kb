// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("loads valid config", func(t *testing.T) {
		content := "name: test-agent\ncreated: \"2024-01-15T10:30:00Z\"\n"
		f, err := os.CreateTemp("", "*akb.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name()) //nolint:errcheck // test cleanup — failure is non-fatal
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load(f.Name())
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Name != "test-agent" {
			t.Errorf("Name = %q, want %q", cfg.Name, "test-agent")
		}
		if cfg.Created != "2024-01-15T10:30:00Z" {
			t.Errorf("Created = %q, want %q", cfg.Created, "2024-01-15T10:30:00Z")
		}
	})

	t.Run("detects merge conflict", func(t *testing.T) {
		content := `name: test-agent
created: "2024-01-15T10:30:00Z"
<<<<<<< HEAD
name: other-agent
=======
name: test-agent
>>>>>>> branch
`
		f, err := os.CreateTemp("", "*akb.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name()) //nolint:errcheck // test cleanup — failure is non-fatal
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		_, err = Load(f.Name())
		if err == nil {
			t.Fatal("Expected error for merge conflict")
		}
		if err.Error() != "resolve merge conflicts in akb.yaml before proceeding" {
			t.Errorf("Error = %q, want %q", err.Error(), "resolve merge conflicts in akb.yaml before proceeding")
		}
	})

	t.Run("rejects missing name", func(t *testing.T) {
		content := "created: \"2024-01-15T10:30:00Z\"\n"
		f, err := os.CreateTemp("", "*akb.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name()) //nolint:errcheck // test cleanup — failure is non-fatal
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		_, err = Load(f.Name())
		if err == nil {
			t.Fatal("Expected error for missing name")
		}
		if err.Error() != "missing required field 'name' in akb.yaml" {
			t.Errorf("Error = %q, want %q", err.Error(), "missing required field 'name' in akb.yaml")
		}
	})

	t.Run("ignores unknown fields", func(t *testing.T) {
		content := "name: test-agent\ncreated: \"2024-01-15T10:30:00Z\"\nunknown: value\n"
		f, err := os.CreateTemp("", "*akb.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name()) //nolint:errcheck // test cleanup — failure is non-fatal
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		cfg, err := Load(f.Name())
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Name != "test-agent" {
			t.Errorf("Name = %q, want %q", cfg.Name, "test-agent")
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := Load("/nonexistent/path/akb.yaml")
		if err == nil {
			t.Fatal("Expected error for missing file")
		}
	})

	t.Run("defaults a missing versioning key to git", func(t *testing.T) {
		path := writeTempConfig(t, "name: test-agent\ncreated: \"2024-01-15T10:30:00Z\"\n")

		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Versioning != VersioningGit {
			t.Errorf("Versioning = %q, want %q", cfg.Versioning, VersioningGit)
		}
	})

	t.Run("loads versioning and commit identity", func(t *testing.T) {
		path := writeTempConfig(t, `name: test-agent
created: "2024-01-15T10:30:00Z"
versioning: none
git-author: agent-kb
git-email: agent@agent-kb
`)

		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Versioning != VersioningNone {
			t.Errorf("Versioning = %q, want %q", cfg.Versioning, VersioningNone)
		}
		if cfg.GitAuthor != "agent-kb" {
			t.Errorf("GitAuthor = %q, want %q", cfg.GitAuthor, "agent-kb")
		}
		if cfg.GitEmail != "agent@agent-kb" {
			t.Errorf("GitEmail = %q, want %q", cfg.GitEmail, "agent@agent-kb")
		}
	})

	t.Run("treats an empty versioning value as git", func(t *testing.T) {
		path := writeTempConfig(t, "name: test-agent\ncreated: \"2024-01-15T10:30:00Z\"\nversioning: \"\"\n")

		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Versioning != VersioningGit {
			t.Errorf("Versioning = %q, want %q", cfg.Versioning, VersioningGit)
		}
	})
}

// writeTempConfig writes content to a temporary akb.yaml and returns its path.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "akb.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSave(t *testing.T) {
	t.Run("saves config to file", func(t *testing.T) {
		f, err := os.CreateTemp("", "*akb.yaml")
		if err != nil {
			t.Fatal(err)
		}
		path := f.Name()
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path) //nolint:errcheck // test cleanup — failure is non-fatal

		cfg := &Config{
			Name:    "test-agent",
			Created: "2024-01-15T10:30:00Z",
		}

		err = Save(path, cfg)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := Load(path)
		if err != nil {
			t.Fatalf("Load after save failed: %v", err)
		}
		if loaded.Name != cfg.Name {
			t.Errorf("Name = %q, want %q", loaded.Name, cfg.Name)
		}
		if loaded.Created != cfg.Created {
			t.Errorf("Created = %q, want %q", loaded.Created, cfg.Created)
		}
	})

	t.Run("round-trips description", func(t *testing.T) {
		f, err := os.CreateTemp("", "*akb.yaml")
		if err != nil {
			t.Fatal(err)
		}
		path := f.Name()
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path) //nolint:errcheck // test cleanup — failure is non-fatal

		cfg := &Config{
			Name:        "described",
			Created:     "2024-01-15T10:30:00Z",
			Description: "notes on the project",
		}
		if err := Save(path, cfg); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := Load(path)
		if err != nil {
			t.Fatalf("Load after save failed: %v", err)
		}
		if loaded.Description != cfg.Description {
			t.Errorf("Description = %q, want %q", loaded.Description, cfg.Description)
		}
	})

	t.Run("omits empty description", func(t *testing.T) {
		f, err := os.CreateTemp("", "*akb.yaml")
		if err != nil {
			t.Fatal(err)
		}
		path := f.Name()
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path) //nolint:errcheck // test cleanup — failure is non-fatal

		if err := Save(path, &Config{Name: "bare", Created: "2024-01-15T10:30:00Z"}); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
		data, err := os.ReadFile(path) //nolint:gosec // test temp file
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "description") {
			t.Errorf("saved config contains a description key: %s", data)
		}
	})

	t.Run("round-trips versioning and commit identity", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "akb.yaml")
		cfg := &Config{
			Name:       "versioned",
			Created:    "2024-01-15T10:30:00Z",
			Versioning: VersioningNone,
			GitAuthor:  "agent-kb",
			GitEmail:   "agent@agent-kb",
		}
		if err := Save(path, cfg); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := Load(path)
		if err != nil {
			t.Fatalf("Load after save failed: %v", err)
		}
		if loaded.Versioning != cfg.Versioning {
			t.Errorf("Versioning = %q, want %q", loaded.Versioning, cfg.Versioning)
		}
		if loaded.GitAuthor != cfg.GitAuthor {
			t.Errorf("GitAuthor = %q, want %q", loaded.GitAuthor, cfg.GitAuthor)
		}
		if loaded.GitEmail != cfg.GitEmail {
			t.Errorf("GitEmail = %q, want %q", loaded.GitEmail, cfg.GitEmail)
		}
	})

	t.Run("omits an unset versioning mode and identity", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "akb.yaml")
		if err := Save(path, &Config{Name: "bare", Created: "2024-01-15T10:30:00Z"}); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
		data, err := os.ReadFile(path) //nolint:gosec // test temp file
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"versioning", "git-author", "git-email"} {
			if strings.Contains(string(data), key) {
				t.Errorf("saved config contains %q: %s", key, data)
			}
		}
	})
}
