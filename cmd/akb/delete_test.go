package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDelete(t *testing.T) {
	// Create a temp KB for testing
	kbRoot := t.TempDir()
	setupTestKBWithGit(t, kbRoot)

	// Save original cwd and restore after
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origCwd)

	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	t.Run("successful delete of existing page", func(t *testing.T) {
		testPage := filepath.Join(kbRoot, "kb", "notes", "test-page.md")
		if err := os.MkdirAll(filepath.Dir(testPage), 0755); err != nil {
			t.Fatal(err)
		}
		// Write with proper frontmatter for production code path
		content := "---\ntype: note\ntitle: Test Page\n---\nContent here.\n"
		if err := os.WriteFile(testPage, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		if err := addToGit(kbRoot, "kb/notes/test-page.md"); err != nil {
			t.Fatal(err)
		}

		// Save and reset noCommit after test
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"notes/test-page.md"})
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		// Verify file is deleted
		if _, err := os.Stat(testPage); !os.IsNotExist(err) {
			t.Errorf("expected file to be deleted, got error: %v", err)
		}
	})

	t.Run("error on non-existent page", func(t *testing.T) {
		// Save and reset noCommit after test
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"notes/non-existent.md"})
		if err == nil {
			t.Error("expected error for non-existent page, got nil")
		}
		// Check error message contains "page not found"
		if err != nil && !strings.Contains(err.Error(), "page not found") {
			t.Errorf("expected 'page not found' in error, got: %v", err)
		}
	})

	t.Run("error on raw/ prefix path", func(t *testing.T) {
		// Save and reset noCommit after test
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"raw/some-file.txt"})
		if err == nil {
			t.Error("expected error for raw/ prefix, got nil")
		}
		// Production code returns "use `akb raw delete`"
		if err != nil && !strings.Contains(err.Error(), "akb raw delete") {
			t.Errorf("expected 'akb raw delete' in error, got: %v", err)
		}
	})

	t.Run("successful delete with no-commit flag", func(t *testing.T) {
		testPage := filepath.Join(kbRoot, "kb", "notes", "no-commit-test.md")
		if err := os.MkdirAll(filepath.Dir(testPage), 0755); err != nil {
			t.Fatal(err)
		}
		// Write with proper frontmatter for production code path
		content := "---\ntype: note\ntitle: No Commit Test\n---\nContent here.\n"
		if err := os.WriteFile(testPage, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		if err := addToGit(kbRoot, "kb/notes/no-commit-test.md"); err != nil {
			t.Fatal(err)
		}

		// Save and reset noCommit after test
		origNoCommit := noCommit
		noCommit = true
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"notes/no-commit-test.md"})
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		if _, err := os.Stat(testPage); !os.IsNotExist(err) {
			t.Errorf("expected file to be deleted, got error: %v", err)
		}
	})

	t.Run("delete with kb/ prefix", func(t *testing.T) {
		testPage := filepath.Join(kbRoot, "kb", "notes", "prefix-test.md")
		if err := os.MkdirAll(filepath.Dir(testPage), 0755); err != nil {
			t.Fatal(err)
		}
		// Write with proper frontmatter for production code path
		content := "---\ntype: note\ntitle: Prefix Test\n---\nContent here.\n"
		if err := os.WriteFile(testPage, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		if err := addToGit(kbRoot, "kb/notes/prefix-test.md"); err != nil {
			t.Fatal(err)
		}

		commitInitial(kbRoot)

		// Save and reset noCommit after test
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"kb/notes/prefix-test.md"})
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		if _, err := os.Stat(testPage); !os.IsNotExist(err) {
			t.Errorf("expected file to be deleted, got error: %v", err)
		}
	})

	t.Run("error on delete index.md", func(t *testing.T) {
		indexPath := filepath.Join(kbRoot, "kb", "index.md")
		if err := os.WriteFile(indexPath, []byte("# Index\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Save and reset noCommit after test
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"kb/index.md"})
		if err == nil {
			t.Error("expected error for index.md delete, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot delete index.md") {
			t.Errorf("expected 'cannot delete index.md' in error, got: %v", err)
		}
	})

	t.Run("error on delete log.md", func(t *testing.T) {
		logPath := filepath.Join(kbRoot, "kb", "log.md")
		if err := os.WriteFile(logPath, []byte("# Log\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Save and reset noCommit after test
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"kb/log.md"})
		if err == nil {
			t.Error("expected error for log.md delete, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot delete log.md") {
			t.Errorf("expected 'cannot delete log.md' in error, got: %v", err)
		}
	})

	t.Run("successful delete page not in index", func(t *testing.T) {
		testPage := filepath.Join(kbRoot, "kb", "notes", "not-in-index.md")
		if err := os.MkdirAll(filepath.Dir(testPage), 0755); err != nil {
			t.Fatal(err)
		}
		// Write with proper frontmatter for production code path
		content := "---\ntype: note\ntitle: Not In Index\n---\nContent here.\n"
		if err := os.WriteFile(testPage, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		if err := addToGit(kbRoot, "kb/notes/not-in-index.md"); err != nil {
			t.Fatal(err)
		}

		// Save and reset noCommit after test
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"notes/not-in-index.md"})
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		if _, err := os.Stat(testPage); !os.IsNotExist(err) {
			t.Errorf("expected file to be deleted, got error: %v", err)
		}
	})
}

// Helper to setup minimal test KB structure
func setupTestKBWithGit(t *testing.T, kbRoot string) {
	dirs := []string{
		filepath.Join(kbRoot, "kb"),
		filepath.Join(kbRoot, "raw"),
		filepath.Join(kbRoot, ".akb"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	configContent := `name: test-kb
created: "2024-01-01T00:00:00Z"
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", ".akb.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "index.md"), []byte("# Index\n\n"), 0644); err != nil {
		t.Fatal(err)
	}

	initTestSearchDB(t, kbRoot)

	if err := initGitRepo(kbRoot); err != nil {
		t.Logf("git not available, skipping git-related tests: %v", err)
	}

	if err := addToGit(kbRoot, "."); err != nil {
		t.Logf("git add initial files failed: %v", err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", kbRoot)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	regPath := filepath.Join(kbRoot, ".config", "agent-kb", "registry.yaml")
	os.MkdirAll(filepath.Dir(regPath), 0755)
	regContent := `default: test-kb
entries:
  - name: test-kb
    path: ` + kbRoot + `
    created: "2024-01-01T00:00:00Z"
`
	os.WriteFile(regPath, []byte(regContent), 0644)
}

func initGitRepo(kbRoot string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = kbRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %s: %w", strings.TrimSpace(string(out)), err)
	}

	setGitConfig := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = kbRoot
		_, err := cmd.CombinedOutput()
		return err
	}

	if err := setGitConfig("user.name", "akb-test"); err != nil {
	}
	if err := setGitConfig("user.email", "akb-test@local"); err != nil {
	}

	return nil
}

func addToGit(kbRoot, path string) error {
	cmd := exec.Command("git", "add", path)
	cmd.Dir = kbRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func commitInitial(kbRoot string) {
	cmd := exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = kbRoot
	cmd.CombinedOutput()
}
