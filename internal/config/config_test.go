// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"os"
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
}
