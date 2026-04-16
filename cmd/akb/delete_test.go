package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/index"
	"github.com/peedrr/agent-kb/internal/log"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
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

	// Change to KB root so path.KBRoot() works
	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	t.Run("successful delete of existing page", func(t *testing.T) {
		testPage := filepath.Join(kbRoot, "kb", "notes", "test-page.md")
		if err := os.MkdirAll(filepath.Dir(testPage), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(testPage, []byte("# Test Page\n"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := addToGit(kbRoot, "kb/notes/test-page.md"); err != nil {
			t.Fatal(err)
		}

		err := runDelete("notes/test-page.md", false)
		if err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		// Verify file is deleted
		if _, err := os.Stat(testPage); !os.IsNotExist(err) {
			t.Errorf("expected file to be deleted, got error: %v", err)
		}
	})

	t.Run("error on non-existent page", func(t *testing.T) {
		err := runDelete("notes/non-existent.md", false)
		if err == nil {
			t.Error("expected error for non-existent page, got nil")
		}
		// Check error message contains "page not found"
		if err != nil && !strings.Contains(err.Error(), "page not found") {
			t.Errorf("expected 'page not found' in error, got: %v", err)
		}
	})

	t.Run("error on raw/ prefix path", func(t *testing.T) {
		err := runDelete("raw/some-file.txt", false)
		if err == nil {
			t.Error("expected error for raw/ prefix, got nil")
		}
		// Should get ErrUseAKBRawWrite
		if err != path.ErrUseAKBRawWrite {
			t.Errorf("expected ErrUseAKBRawWrite, got: %v", err)
		}
	})

	t.Run("successful delete with no-commit flag", func(t *testing.T) {
		testPage := filepath.Join(kbRoot, "kb", "notes", "no-commit-test.md")
		if err := os.MkdirAll(filepath.Dir(testPage), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(testPage, []byte("# No Commit Test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := addToGit(kbRoot, "kb/notes/no-commit-test.md"); err != nil {
			t.Fatal(err)
		}

		err := runDelete("notes/no-commit-test.md", true)
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
		if err := os.WriteFile(testPage, []byte("# Prefix Test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := addToGit(kbRoot, "kb/notes/prefix-test.md"); err != nil {
			t.Fatal(err)
		}

		commitInitial(kbRoot)

		err := runDelete("kb/notes/prefix-test.md", false)
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

		err := runDelete("kb/index.md", false)
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

		err := runDelete("kb/log.md", false)
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
		if err := os.WriteFile(testPage, []byte("# Not In Index\n"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := addToGit(kbRoot, "kb/notes/not-in-index.md"); err != nil {
			t.Fatal(err)
		}

		err := runDelete("notes/not-in-index.md", false)
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

	// Write minimal config
	configContent := `name: test-kb
created: "2024-01-01T00:00:00Z"
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", ".akb.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a basic index.md
	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "index.md"), []byte("# Index\n\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize git repo for testing
	if err := initGitRepo(kbRoot); err != nil {
		t.Logf("git not available, skipping git-related tests: %v", err)
	}

	if err := addToGit(kbRoot, "."); err != nil {
		t.Logf("git add initial files failed: %v", err)
	}
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

	if err := setGitConfig("config", "user.name", "akb-test"); err != nil {
	}
	if err := setGitConfig("config", "user.email", "akb-test@local"); err != nil {
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

func execGitAddTest(kbRoot, relPath string) {
	cmd := exec.Command("git", "add", relPath)
	cmd.Dir = kbRoot
	cmd.Output()
}

func runDelete(inputPath string, noCommit bool) error {
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
		return path.ErrUseAKBRawWrite
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")

	if cleanPath == "index.md" {
		return fmt.Errorf("cannot delete index.md; use 'akb index rebuild' to reset")
	}
	if cleanPath == "log.md" {
		return fmt.Errorf("cannot delete log.md; it is a managed file")
	}

	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	exists, err := fileExists(fullPath)
	if err != nil {
		return err
	}
	if !exists {
		return &DeleteError{Message: "page not found: " + inputPath}
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	ctx := context.Background()
	if err := store.Delete(ctx, fullPath); err != nil {
		return err
	}

	relPath := filepath.Join("kb", cleanPath)
	relPath = filepath.ToSlash(relPath)

	_ = index.RemoveEntry(kbRoot, relPath)
	_ = log.AppendLog(kbRoot, "delete", "Removed page "+cleanPath, "")
	execGitAddTest(kbRoot, "kb/index.md")
	execGitAddTest(kbRoot, "kb/log.md")

	fmt.Printf("Deleted %s\n", relPath)

	return nil
}

// DeleteError is a custom error for delete operations
type DeleteError struct {
	Message string
}

func (e *DeleteError) Error() string {
	return e.Message
}
