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
	defer os.Chdir(origCwd) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	t.Run("successful delete of existing page", func(t *testing.T) {
		testPage := filepath.Join(kbRoot, "kb", "notes", "test-page.md")
		if err := os.MkdirAll(filepath.Dir(testPage), 0750); err != nil {
			t.Fatal(err)
		}
		// Write with proper frontmatter for production code path
		content := "---\ntype: note\ntitle: Test Page\n---\nContent here.\n"
		if err := os.WriteFile(testPage, []byte(content), 0600); err != nil {
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
		if err := os.MkdirAll(filepath.Dir(testPage), 0750); err != nil {
			t.Fatal(err)
		}
		// Write with proper frontmatter for production code path
		content := "---\ntype: note\ntitle: No Commit Test\n---\nContent here.\n"
		if err := os.WriteFile(testPage, []byte(content), 0600); err != nil {
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
		if err := os.MkdirAll(filepath.Dir(testPage), 0750); err != nil {
			t.Fatal(err)
		}
		// Write with proper frontmatter for production code path
		content := "---\ntype: note\ntitle: Prefix Test\n---\nContent here.\n"
		if err := os.WriteFile(testPage, []byte(content), 0600); err != nil {
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
		if err := os.WriteFile(indexPath, []byte("# Index\n"), 0600); err != nil {
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
		if err := os.WriteFile(logPath, []byte("# Log\n"), 0600); err != nil {
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

	t.Run("error on delete of managed file via dot-segment paths", func(t *testing.T) {
		indexPath := filepath.Join(kbRoot, "kb", "index.md")
		if err := os.WriteFile(indexPath, []byte("# Index\n"), 0600); err != nil {
			t.Fatal(err)
		}

		origNoCommit := noCommit
		noCommit = true
		t.Cleanup(func() { noCommit = origNoCommit })

		cases := []struct {
			input   string
			wantErr string
		}{
			{"./index.md", "cannot delete index.md"},
			{"kb/./log.md", "cannot delete log.md"},
		}
		for _, tc := range cases {
			err := runDeleteCmd(&cobra.Command{}, []string{tc.input})
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.input)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("delete %q: expected %q in error, got: %v", tc.input, tc.wantErr, err)
			}
		}

		if _, err := os.Stat(indexPath); err != nil {
			t.Errorf("index.md should not be deleted: %v", err)
		}
	})

	t.Run("error on parent dir traversal", func(t *testing.T) {
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"../escape.md"})
		if err == nil {
			t.Fatal("expected error for .. path, got nil")
		}
		if !strings.Contains(err.Error(), "..") {
			t.Errorf("expected error to contain '..', got: %v", err)
		}
	})

	t.Run("error on absolute path", func(t *testing.T) {
		origNoCommit := noCommit
		noCommit = false
		t.Cleanup(func() { noCommit = origNoCommit })

		err := runDeleteCmd(&cobra.Command{}, []string{"/tmp/evil.md"})
		if err == nil {
			t.Fatal("expected error for absolute path, got nil")
		}
		if !strings.Contains(err.Error(), "absolute") && !strings.Contains(err.Error(), "relative") {
			t.Errorf("expected error about absolute/relative path, got: %v", err)
		}
	})

	t.Run("successful delete page not in index", func(t *testing.T) {
		testPage := filepath.Join(kbRoot, "kb", "notes", "not-in-index.md")
		if err := os.MkdirAll(filepath.Dir(testPage), 0750); err != nil {
			t.Fatal(err)
		}
		// Write with proper frontmatter for production code path
		content := "---\ntype: note\ntitle: Not In Index\n---\nContent here.\n"
		if err := os.WriteFile(testPage, []byte(content), 0600); err != nil {
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
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatal(err)
		}
	}

	configContent := `name: test-kb
created: "2024-01-01T00:00:00Z"
`
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", ".akb.yaml"), []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "index.md"), []byte("# Index\n\n"), 0600); err != nil {
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
	os.Setenv("HOME", kbRoot)                         //nolint:errcheck,gosec // test setup — failure is non-fatal
	t.Cleanup(func() { os.Setenv("HOME", origHome) }) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	useTestKBSelection(t, kbRoot)
}

func initGitRepo(kbRoot string) error {
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "init")
	cmd.Dir = kbRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %s: %w", strings.TrimSpace(string(out)), err)
	}

	setGitConfig := func(args ...string) error {
		cmd := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", args...)
		cmd.Dir = kbRoot
		_, err := cmd.CombinedOutput()
		return err
	}

	_ = setGitConfig("user.name", "akb-test")
	_ = setGitConfig("user.email", "akb-test@local")

	return nil
}

func addToGit(kbRoot, path string) error {
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "add", path)
	cmd.Dir = kbRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func commitInitial(kbRoot string) {
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "commit", "-m", "initial")
	cmd.Dir = kbRoot
	_, _ = cmd.CombinedOutput() //nolint:errcheck,gosec // best-effort git commit in test helper
}

// TestDeleteFileStepFailureAfterIndexRemovalReportsRemediation pins that a file
// step failure after the search index, link graph, and index.md steps reports
// the divergent state and the rebuild remediation, while a successful delete
// reports neither.
func TestDeleteFileStepFailureAfterIndexRemovalReportsRemediation(t *testing.T) {
	kbRoot := t.TempDir()
	setupTestKBWithGit(t, kbRoot)

	origNoCommit := noCommit
	noCommit = false
	t.Cleanup(func() { noCommit = origNoCommit })

	mustGitInDir(t, kbRoot, "commit", "-m", "initial")

	// Control: the same command succeeds and reports no remediation while the
	// file step can run.
	controlRel := "kb/notes/delete-control.md"
	controlFull := filepath.Join(kbRoot, filepath.FromSlash(controlRel))
	if err := os.MkdirAll(filepath.Dir(controlFull), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(controlFull, []byte("---\ntype: note\ntitle: Delete Control\n---\nBody.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := addToGit(kbRoot, controlRel); err != nil {
		t.Fatal(err)
	}

	controlOut, err := captureOutput(func() error {
		return runDeleteCmd(&cobra.Command{}, []string{"notes/delete-control.md"})
	})
	if err != nil {
		t.Fatalf("control delete failed: %s: %v", controlOut, err)
	}
	if strings.Contains(controlOut, "akb index rebuild") {
		t.Errorf("a successful delete must not carry rebuild remediation, got: %s", controlOut)
	}
	if _, err := os.Stat(controlFull); !os.IsNotExist(err) {
		t.Errorf("control page should be gone after delete, stat error: %v", err)
	}

	// The page path is a non-empty directory, so os.Remove fails in the file
	// step of the delete after the search index, link graph, and index.md
	// steps ran.
	failedRel := "kb/notes/delete-failure.md"
	failedFull := filepath.Join(kbRoot, filepath.FromSlash(failedRel))
	if err := os.MkdirAll(failedFull, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(failedFull, "blocker"), []byte("blocker"), 0600); err != nil {
		t.Fatal(err)
	}

	err = runDeleteCmd(&cobra.Command{}, []string{"notes/delete-failure.md"})
	if err == nil {
		t.Fatal("expected the delete to fail at the file step")
	}
	if !strings.Contains(err.Error(), "already removed from the search index and link graph") {
		t.Errorf("expected the delete error to surface the removed derived records, got: %v", err)
	}
	if !strings.Contains(err.Error(), "file removal or its commit did not complete") {
		t.Errorf("expected the delete error to stay truthful across both failure sub-cases, got: %v", err)
	}
	if !strings.Contains(err.Error(), "akb index rebuild") {
		t.Errorf("expected the delete error to carry the rebuild remediation, got: %v", err)
	}
	if _, statErr := os.Stat(failedFull); statErr != nil {
		t.Errorf("the page path should still exist after the failed file step, stat error: %v", statErr)
	}
}

// TestDeleteLinkFailureRollsBackSearchRemoval pins that the delete path removes
// the page from the search index and the link graph in one transaction: when
// the second step, the link-graph removal, fails, the first step, the
// search-index removal, does not persist and both indexes keep the page.
func TestDeleteLinkFailureRollsBackSearchRemoval(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	origNoCommit := noCommit
	noCommit = false
	t.Cleanup(func() { noCommit = origNoCommit })

	target := "---\ntype: note\ntitle: Rollback Target\nsummary: test\ntags: test\n---\nTarget body."
	if out, err := writeRun(kbRoot, "rollback-target.md", target); err != nil {
		t.Fatalf("write target: %s: %v", out, err)
	}
	source := "---\ntype: note\ntitle: Rollback Source\nsummary: test\ntags: test\n---\nSource body. See [[rollback-target]]."
	if out, err := writeRun(kbRoot, "rollback-source.md", source); err != nil {
		t.Fatalf("write source: %s: %v", out, err)
	}

	const sourceRel = "kb/notes/rollback-source.md"
	preBody, ok := searchDBDocumentBody(t, kbRoot, sourceRel)
	if !ok {
		t.Fatalf("expected the source page indexed at %s", sourceRel)
	}
	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM links WHERE source_page = ? AND resolved_to = ?", sourceRel, "kb/notes/rollback-target.md"); got != 1 {
		t.Fatalf("pre-command resolved link rows = %d, want 1", got)
	}

	installSearchDBStatement(t, kbRoot, abortLinkDeleteSQL)

	err := runDeleteCmd(&cobra.Command{}, []string{"notes/rollback-source.md"})
	if err == nil {
		t.Fatal("expected the delete to fail when the link removal is aborted")
	}
	if !strings.Contains(err.Error(), "remove from link graph") {
		t.Fatalf("expected the link-graph removal to fail, got: %v", err)
	}

	postBody, ok := searchDBDocumentBody(t, kbRoot, sourceRel)
	if !ok {
		t.Errorf("expected %s to stay in the search index after the failed link removal — the search-index removal did not roll back", sourceRel)
	} else if postBody != preBody {
		t.Errorf("indexed body after the failed link removal = %q, want the pre-command body %q", postBody, preBody)
	}
	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM pages WHERE path = ?", sourceRel); got != 1 {
		t.Errorf("pages rows for %s = %d after the failed link removal, want 1", sourceRel, got)
	}
	if got := searchDBRowCount(t, kbRoot, "SELECT COUNT(*) FROM links WHERE source_page = ? AND resolved_to = ?", sourceRel, "kb/notes/rollback-target.md"); got != 1 {
		t.Errorf("resolved link rows after the failed link removal = %d, want 1", got)
	}
	if _, statErr := os.Stat(filepath.Join(kbRoot, "kb", "notes", "rollback-source.md")); statErr != nil {
		t.Errorf("the page file should stay on disk after the failed index removal, stat error: %v", statErr)
	}
}
