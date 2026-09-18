package storage

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- FilesystemProvider Tests ---

func TestFilesystemProvider_Write(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewFilesystemProvider(tmpDir)
	ctx := context.Background()

	t.Run("writes content to file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "notes.md")
		err := p.Write(ctx, path, []byte("# Hello"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != "# Hello" {
			t.Errorf("content = %q, want %q", string(data), "# Hello")
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "deep", "nested", "file.md")
		err := p.Write(ctx, path, []byte("deep"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != "deep" {
			t.Errorf("content = %q, want %q", string(data), "deep")
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "overwrite.md")
		if err := p.Write(ctx, path, []byte("first")); err != nil {
			t.Fatal(err)
		}
		if err := p.Write(ctx, path, []byte("second")); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "second" {
			t.Errorf("content = %q, want %q", string(data), "second")
		}
	})
}

func TestFilesystemProvider_Read(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewFilesystemProvider(tmpDir)
	ctx := context.Background()

	t.Run("reads existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "read-test.md")
		content := []byte("read me")
		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}
		data, err := p.Read(ctx, path)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if string(data) != "read me" {
			t.Errorf("content = %q, want %q", string(data), "read me")
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := p.Read(ctx, filepath.Join(tmpDir, "kb", "nonexistent.md"))
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestFilesystemProvider_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewFilesystemProvider(tmpDir)
	ctx := context.Background()

	t.Run("deletes existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "delete-test.md")
		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("bye"), 0600); err != nil {
			t.Fatal(err)
		}
		err := p.Delete(ctx, path)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("file should not exist after delete")
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		err := p.Delete(ctx, filepath.Join(tmpDir, "kb", "nonexistent.md"))
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestFilesystemProvider_Exists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewFilesystemProvider(tmpDir)
	ctx := context.Background()

	t.Run("returns true for existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "exists.md")
		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("yes"), 0600); err != nil {
			t.Fatal(err)
		}
		exists, err := p.Exists(ctx, path)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("expected true for existing file")
		}
	})

	t.Run("returns false for missing file", func(t *testing.T) {
		exists, err := p.Exists(ctx, filepath.Join(tmpDir, "kb", "nope.md"))
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("expected false for missing file")
		}
	})
}

// --- GitProvider Tests ---

// initGitRepo creates a temporary git repo for testing.
func initGitRepo(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}

	gitInit := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "init")
	gitInit.Dir = tmpDir
	if out, err := gitInit.CombinedOutput(); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup on failure — secondary to init error
		t.Fatalf("git init: %s: %v", strings.TrimSpace(string(out)), err)
	}

	// Set git config for commits
	gitConfig := func(args ...string) {
		cmd := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", args...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			_ = os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup on failure — secondary to config error
			t.Fatalf("git %v: %s: %v", args, strings.TrimSpace(string(out)), err)
		}
	}
	gitConfig("config", "user.name", "test")
	gitConfig("config", "user.email", "test@test.com")

	// Create initial commit so HEAD exists
	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test KB\n"), 0600); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup on failure — secondary to write error
		t.Fatal(err)
	}
	gitAdd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "add", "README.md")
	gitAdd.Dir = tmpDir
	if out, err := gitAdd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup on failure — secondary to add error
		t.Fatalf("git add: %s: %v", strings.TrimSpace(string(out)), err)
	}
	gitCommit := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "commit", "-m", "initial")
	gitCommit.Dir = tmpDir
	if out, err := gitCommit.CombinedOutput(); err != nil {
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup on failure — secondary to commit error
		t.Fatalf("git commit: %s: %v", strings.TrimSpace(string(out)), err)
	}

	return tmpDir
}

func TestGitProvider_Write(t *testing.T) {
	tmpDir := initGitRepo(t)
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewGitProvider(tmpDir, false)
	ctx := context.Background()

	t.Run("writes file and commits", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "notes.md")
		err := p.Write(ctx, path, []byte("# My Notes"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		// Verify file exists on disk
		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != "# My Notes" {
			t.Errorf("content = %q, want %q", string(data), "# My Notes")
		}

		// Verify git commit was made
		gitLog := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline", "-1")
		gitLog.Dir = tmpDir
		out, err := gitLog.CombinedOutput()
		if err != nil {
			t.Fatalf("git log failed: %v", err)
		}
		logMsg := strings.TrimSpace(string(out))
		if !strings.Contains(logMsg, "akb: write") {
			t.Errorf("commit message should contain 'akb: write', got %q", logMsg)
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "deep", "nested", "file.md")
		err := p.Write(ctx, path, []byte("nested content"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != "nested content" {
			t.Errorf("content = %q, want %q", string(data), "nested content")
		}
	})

	t.Run("commit message includes relative path with kb/ prefix", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "my-note.md")
		err := p.Write(ctx, path, []byte("note content"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		gitLog := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline", "-1")
		gitLog.Dir = tmpDir
		out, err := gitLog.CombinedOutput()
		if err != nil {
			t.Fatalf("git log failed: %v", err)
		}
		logMsg := strings.TrimSpace(string(out))
		if !strings.Contains(logMsg, "kb/my-note.md") {
			t.Errorf("commit message should contain 'kb/my-note.md', got %q", logMsg)
		}
	})
}

func TestGitProvider_Write_NoCommit(t *testing.T) {
	tmpDir := initGitRepo(t)
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewGitProvider(tmpDir, true) // noCommit = true
	ctx := context.Background()

	t.Run("writes file but skips commit", func(t *testing.T) {
		// Get current commit count
		gitLog := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline")
		gitLog.Dir = tmpDir
		out, _ := gitLog.CombinedOutput()
		commitCountBefore := len(strings.Split(strings.TrimSpace(string(out)), "\n"))

		path := filepath.Join(tmpDir, "kb", "no-commit.md")
		err := p.Write(ctx, path, []byte("no commit"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		// Verify file exists on disk
		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != "no commit" {
			t.Errorf("content = %q, want %q", string(data), "no commit")
		}

		// Verify file is staged (git add was run)
		gitStatus := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "status", "--porcelain")
		gitStatus.Dir = tmpDir
		out, err = gitStatus.CombinedOutput()
		if err != nil {
			t.Fatalf("git status failed: %v", err)
		}
		statusLines := strings.TrimSpace(string(out))
		if !strings.Contains(statusLines, "kb/no-commit.md") {
			t.Errorf("file should be staged, got status: %q", statusLines)
		}

		// Verify no new commit was made
		gitLog = exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline")
		gitLog.Dir = tmpDir
		out, _ = gitLog.CombinedOutput()
		commitCountAfter := len(strings.Split(strings.TrimSpace(string(out)), "\n"))
		if commitCountAfter != commitCountBefore {
			t.Errorf("commit count changed: before=%d, after=%d", commitCountBefore, commitCountAfter)
		}
	})
}

func TestGitProvider_Read(t *testing.T) {
	tmpDir := initGitRepo(t)
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewGitProvider(tmpDir, false)
	ctx := context.Background()

	t.Run("reads existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "read-test.md")
		if err := p.Write(ctx, path, []byte("read me")); err != nil {
			t.Fatal(err)
		}
		data, err := p.Read(ctx, path)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if string(data) != "read me" {
			t.Errorf("content = %q, want %q", string(data), "read me")
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := p.Read(ctx, filepath.Join(tmpDir, "kb", "nonexistent.md"))
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestGitProvider_Delete(t *testing.T) {
	tmpDir := initGitRepo(t)
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewGitProvider(tmpDir, false)
	ctx := context.Background()

	t.Run("deletes file and commits", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "to-delete.md")
		if err := p.Write(ctx, path, []byte("delete me")); err != nil {
			t.Fatal(err)
		}

		err := p.Delete(ctx, path)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify file is gone
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("file should not exist after delete")
		}

		// Verify git commit was made
		gitLog := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline", "-1")
		gitLog.Dir = tmpDir
		out, err := gitLog.CombinedOutput()
		if err != nil {
			t.Fatalf("git log failed: %v", err)
		}
		logMsg := strings.TrimSpace(string(out))
		if !strings.Contains(logMsg, "akb: delete") {
			t.Errorf("commit message should contain 'akb: delete', got %q", logMsg)
		}
	})

	t.Run("delete commit message includes relative path", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "del-note.md")
		if err := p.Write(ctx, path, []byte("will delete")); err != nil {
			t.Fatal(err)
		}

		err := p.Delete(ctx, path)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		gitLog := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline", "-1")
		gitLog.Dir = tmpDir
		out, err := gitLog.CombinedOutput()
		if err != nil {
			t.Fatalf("git log failed: %v", err)
		}
		logMsg := strings.TrimSpace(string(out))
		if !strings.Contains(logMsg, "kb/del-note.md") {
			t.Errorf("commit message should contain 'kb/del-note.md', got %q", logMsg)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		err := p.Delete(ctx, filepath.Join(tmpDir, "kb", "nonexistent.md"))
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestGitProvider_Delete_NoCommit(t *testing.T) {
	tmpDir := initGitRepo(t)
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewGitProvider(tmpDir, true) // noCommit = true
	ctx := context.Background()

	t.Run("deletes file but skips commit", func(t *testing.T) {
		// First write a file with commit enabled
		pCommit := NewGitProvider(tmpDir, false)
		path := filepath.Join(tmpDir, "kb", "no-commit-del.md")
		if err := pCommit.Write(ctx, path, []byte("will delete")); err != nil {
			t.Fatal(err)
		}

		// Get current commit count
		gitLog := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline")
		gitLog.Dir = tmpDir
		out, _ := gitLog.CombinedOutput()
		commitCountBefore := len(strings.Split(strings.TrimSpace(string(out)), "\n"))

		err := p.Delete(ctx, path)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify file is gone
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("file should not exist after delete")
		}

		// Verify no new commit was made
		gitLog = exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline")
		gitLog.Dir = tmpDir
		out, _ = gitLog.CombinedOutput()
		commitCountAfter := len(strings.Split(strings.TrimSpace(string(out)), "\n"))
		if commitCountAfter != commitCountBefore {
			t.Errorf("commit count changed: before=%d, after=%d", commitCountBefore, commitCountAfter)
		}
	})
}

func TestGitProvider_WriteWithCommitMsg(t *testing.T) {
	tmpDir := initGitRepo(t)
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewGitProvider(tmpDir, false)
	ctx := context.Background()

	t.Run("uses custom commit message", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "custom-msg.md")
		err := p.WriteWithCommitMsg(ctx, path, []byte("custom content"), "akb: append kb/custom-msg.md")
		if err != nil {
			t.Fatalf("WriteWithCommitMsg failed: %v", err)
		}

		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != "custom content" {
			t.Errorf("content = %q, want %q", string(data), "custom content")
		}

		gitLog := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline", "-1")
		gitLog.Dir = tmpDir
		out, err := gitLog.CombinedOutput()
		if err != nil {
			t.Fatalf("git log failed: %v", err)
		}
		logMsg := strings.TrimSpace(string(out))
		if !strings.Contains(logMsg, "akb: append kb/custom-msg.md") {
			t.Errorf("commit message should contain 'akb: append kb/custom-msg.md', got %q", logMsg)
		}
	})

	t.Run("skips commit when noCommit is true", func(t *testing.T) {
		pNoCommit := NewGitProvider(tmpDir, true)
		path := filepath.Join(tmpDir, "kb", "no-commit-msg.md")

		gitLog := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline")
		gitLog.Dir = tmpDir
		out, _ := gitLog.CombinedOutput()
		commitCountBefore := len(strings.Split(strings.TrimSpace(string(out)), "\n"))

		err := pNoCommit.WriteWithCommitMsg(ctx, path, []byte("no commit custom"), "akb: append kb/no-commit-msg.md")
		if err != nil {
			t.Fatalf("WriteWithCommitMsg with noCommit failed: %v", err)
		}

		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != "no commit custom" {
			t.Errorf("content = %q, want %q", string(data), "no commit custom")
		}

		gitLog = exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "--oneline")
		gitLog.Dir = tmpDir
		out, _ = gitLog.CombinedOutput()
		commitCountAfter := len(strings.Split(strings.TrimSpace(string(out)), "\n"))
		if commitCountAfter != commitCountBefore {
			t.Errorf("commit count changed: before=%d, after=%d", commitCountBefore, commitCountAfter)
		}
	})
}

func TestGitProvider_Exists(t *testing.T) {
	tmpDir := initGitRepo(t)
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewGitProvider(tmpDir, false)
	ctx := context.Background()

	t.Run("returns true for existing file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "exists.md")
		if err := p.Write(ctx, path, []byte("yes")); err != nil {
			t.Fatal(err)
		}
		exists, err := p.Exists(ctx, path)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("expected true for existing file")
		}
	})

	t.Run("returns false for missing file", func(t *testing.T) {
		exists, err := p.Exists(ctx, filepath.Join(tmpDir, "kb", "nope.md"))
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("expected false for missing file")
		}
	})
}

func TestGitProvider_MergeConflictDetection(t *testing.T) {
	tmpDir := initGitRepo(t)
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	p := NewGitProvider(tmpDir, false)
	ctx := context.Background()

	t.Run("detects merge conflict and returns error", func(t *testing.T) {
		// Create a merge conflict scenario
		// Create a branch, make changes, then create conflict
		gitCheckout := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "checkout", "-b", "feature")
		gitCheckout.Dir = tmpDir
		if out, err := gitCheckout.CombinedOutput(); err != nil {
			t.Fatalf("git checkout -b: %s: %v", strings.TrimSpace(string(out)), err)
		}

		conflictPath := filepath.Join(tmpDir, "kb", "conflict.md")
		if err := os.MkdirAll(filepath.Dir(conflictPath), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(conflictPath, []byte("feature content"), 0600); err != nil {
			t.Fatal(err)
		}
		gitAdd := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "add", "-A")
		gitAdd.Dir = tmpDir
		if out, err := gitAdd.CombinedOutput(); err != nil {
			t.Fatalf("git add: %s: %v", strings.TrimSpace(string(out)), err)
		}
		gitCommit := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "commit", "-m", "feature change")
		gitCommit.Dir = tmpDir
		if out, err := gitCommit.CombinedOutput(); err != nil {
			t.Fatalf("git commit: %s: %v", strings.TrimSpace(string(out)), err)
		}

		gitCheckout = exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "checkout", "master")
		gitCheckout.Dir = tmpDir
		if _, err := gitCheckout.CombinedOutput(); err != nil {
			gitCheckout = exec.Command( //nolint:gosec // test helper launching akb binary
				"git", "checkout", "main")
			gitCheckout.Dir = tmpDir
			if out, err := gitCheckout.CombinedOutput(); err != nil {
				t.Fatalf("git checkout main: %s: %v", strings.TrimSpace(string(out)), err)
			}
		}

		if err := os.MkdirAll(filepath.Dir(conflictPath), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(conflictPath, []byte("main content"), 0600); err != nil {
			t.Fatal(err)
		}
		gitAdd = exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "add", "-A")
		gitAdd.Dir = tmpDir
		if out, err := gitAdd.CombinedOutput(); err != nil {
			t.Fatalf("git add: %s: %v", strings.TrimSpace(string(out)), err)
		}
		gitCommit = exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "commit", "-m", "main change")
		gitCommit.Dir = tmpDir
		if out, err := gitCommit.CombinedOutput(); err != nil {
			t.Fatalf("git commit: %s: %v", strings.TrimSpace(string(out)), err)
		}

		// Merge feature branch to create conflict
		gitMerge := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "merge", "feature")
		gitMerge.Dir = tmpDir
		_, _ = gitMerge.CombinedOutput() //nolint:errcheck // expected to fail — conflict is the test scenario

		// Now try to write - should detect conflict
		err := p.Write(ctx, filepath.Join(tmpDir, "kb", "new-file.md"), []byte("test"))
		if err == nil {
			t.Fatal("expected error for merge conflict")
		}
		if !strings.Contains(err.Error(), "merge conflict") && !strings.Contains(err.Error(), "unmerged") {
			t.Errorf("error should mention merge conflict or unmerged, got: %v", err)
		}
	})
}

func TestGitProvider_CommitWithoutRepoIdentity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	gitInit := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "init")
	gitInit.Dir = tmpDir
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", strings.TrimSpace(string(out)), err)
	}

	gitConfig := func(args ...string) {
		cmd := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", args...)
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, strings.TrimSpace(string(out)), err)
		}
	}
	gitConfig("config", "user.name", "test")
	gitConfig("config", "user.email", "test@test.com")

	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	gitAdd := exec.Command( //nolint:gosec // test helper launching akb binary
		"git", "add", "README.md")
	gitAdd.Dir = tmpDir
	gitAdd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com", "GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com")
	_, _ = gitAdd.CombinedOutput() //nolint:errcheck // test setup — failure caught by subsequent test assertions
	gitCommit := exec.Command(     //nolint:gosec // test helper launching akb binary
		"git", "commit", "-m", "initial")
	gitCommit.Dir = tmpDir
	gitCommit.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com", "GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com")
	_, _ = gitCommit.CombinedOutput() //nolint:errcheck // test setup — failure caught by subsequent test assertions

	gitConfig("config", "--unset", "user.name")
	gitConfig("config", "--unset", "user.email")

	p := NewGitProvider(tmpDir, false)
	ctx := context.Background()

	t.Run("commits without writing repository identity", func(t *testing.T) {
		path := filepath.Join(tmpDir, "kb", "auto-config.md")
		err := p.Write(ctx, path, []byte("auto config test"))
		if err != nil {
			t.Fatalf("Write without repository identity failed: %v", err)
		}

		data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != "auto config test" {
			t.Errorf("content = %q, want %q", string(data), "auto config test")
		}

		author := exec.Command( //nolint:gosec // test helper launching akb binary
			"git", "log", "-1", "--format=%an <%ae>")
		author.Dir = tmpDir
		out, err := author.Output()
		if err != nil {
			t.Fatalf("git log: %v", err)
		}
		if strings.TrimSpace(string(out)) != "akb <akb@local>" {
			t.Errorf("author = %q, want %q", strings.TrimSpace(string(out)), "akb <akb@local>")
		}

		for _, key := range []string{"user.name", "user.email"} {
			cmd := exec.Command( //nolint:gosec // test helper launching akb binary
				"git", "config", "--local", key)
			cmd.Dir = tmpDir
			cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
			if out, err := cmd.Output(); err == nil {
				t.Errorf("repository config %s = %q, want it unset", key, strings.TrimSpace(string(out)))
			}
		}
	})
}

// --- Interface Compliance Tests ---

func TestFilesystemProvider_ImplementsStorageProvider(_ *testing.T) {
	// Compile-time interface check
	var _ Provider = (*FilesystemProvider)(nil)
}

func TestGitProvider_ImplementsStorageProvider(_ *testing.T) {
	// Compile-time interface check
	var _ Provider = (*GitProvider)(nil)
}
