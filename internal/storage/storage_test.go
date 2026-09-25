// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package storage

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
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

// isolateGitConfig keeps the developer machine's global and system git config
// out of the test, so an ambient user.name cannot pick the identity branch
// under test.
func isolateGitConfig(t *testing.T) {
	t.Helper()

	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("HOME", t.TempDir())
}

// forceUnsetUser forces git's merged config to carry no user.name. Some git
// builds (nix) ignore GIT_CONFIG_GLOBAL and read a compiled-in global config,
// so the identity resolution is given an empty user.name instead, which it
// treats as unset.
func forceUnsetUser(t *testing.T) {
	t.Helper()

	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "user.name")
	t.Setenv("GIT_CONFIG_VALUE_0", "")
}

// gitIn runs one git command in repo and returns its combined output.
func gitIn(t *testing.T, repo string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // test helper launching the git binary
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// mustGitIn runs one git command in repo and fails the test when it fails.
func mustGitIn(t *testing.T, repo string, args ...string) string {
	t.Helper()

	out, err := gitIn(t, repo, args...)
	if err != nil {
		t.Fatalf("git %s in %s: %s: %v", strings.Join(args, " "), repo, out, err)
	}
	return out
}

// initRepoWithoutIdentity creates a temporary git repository with one commit
// and no configured user identity; the initial commit is authored through
// one-off overrides, so the repository config stays empty.
func initRepoWithoutIdentity(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	mustGitIn(t, repo, "init")
	readme := filepath.Join(repo, "README.md")
	if err := os.WriteFile(readme, []byte("# repo\n"), 0600); err != nil { //nolint:gosec // test writing into its temp repository
		t.Fatal(err)
	}
	mustGitIn(t, repo, "add", "--", "README.md")
	mustGitIn(t, repo, "-c", "user.name=seed", "-c", "user.email=seed@test", "commit", "-m", "initial")
	return repo
}

// commitAuthor returns the author of HEAD.
func commitAuthor(t *testing.T, repo string) string {
	t.Helper()

	return mustGitIn(t, repo, "log", "-1", "--format=%an <%ae>")
}

// assertRepoIdentityUnset fails the test when the repository config carries a
// user identity.
func assertRepoIdentityUnset(t *testing.T, repo string) {
	t.Helper()

	for _, key := range []string{"user.name", "user.email"} {
		if out, err := gitIn(t, repo, "config", "--local", key); err == nil {
			t.Errorf("repository config %s = %q, want it unset", key, out)
		}
	}
}

func TestCommitIdentityArgs(t *testing.T) {
	isolateGitConfig(t)

	t.Run("falls back to the akb identity without a configured user", func(t *testing.T) {
		forceUnsetUser(t)
		repo := initRepoWithoutIdentity(t)

		identity, err := CommitIdentityArgs(repo)
		if err != nil {
			t.Fatalf("CommitIdentityArgs: %v", err)
		}
		if want := []string{"-c", "user.name=akb", "-c", "user.email=akb@local"}; !reflect.DeepEqual(identity, want) {
			t.Errorf("identity = %v, want %v", identity, want)
		}
	})

	t.Run("keeps the configured identity", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")

		identity, err := CommitIdentityArgs(repo)
		if err != nil {
			t.Fatalf("CommitIdentityArgs: %v", err)
		}
		if len(identity) != 0 {
			t.Errorf("identity = %v, want no overrides", identity)
		}
	})
}

func TestGitProvider_CommitWithoutRepoIdentity(t *testing.T) {
	isolateGitConfig(t)
	forceUnsetUser(t)
	repo := initRepoWithoutIdentity(t)

	provider := NewGitProvider(repo, false)
	ctx := context.Background()

	t.Run("commits as the fallback identity", func(t *testing.T) {
		path := filepath.Join(repo, "kb", "auto-config.md")
		if err := provider.Write(ctx, path, []byte("auto config test")); err != nil {
			t.Fatalf("write without repository identity: %v", err)
		}

		data, err := os.ReadFile(path) //nolint:gosec // test reading a file in its temp repository
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(data) != "auto config test" {
			t.Errorf("content = %q, want %q", string(data), "auto config test")
		}

		if author := commitAuthor(t, repo); author != "akb <akb@local>" {
			t.Errorf("author = %q, want %q", author, "akb <akb@local>")
		}
		assertRepoIdentityUnset(t, repo)
	})
}

func TestGitProvider_CommitWithConfiguredIdentity(t *testing.T) {
	isolateGitConfig(t)
	repo := initRepoWithoutIdentity(t)
	mustGitIn(t, repo, "config", "user.name", "ada")
	mustGitIn(t, repo, "config", "user.email", "ada@example.com")

	provider := NewGitProvider(repo, false)
	path := filepath.Join(repo, "kb", "configured.md")
	if err := provider.Write(context.Background(), path, []byte("configured identity test")); err != nil {
		t.Fatalf("write with configured repository identity: %v", err)
	}

	if author := commitAuthor(t, repo); author != "ada <ada@example.com>" {
		t.Errorf("author = %q, want %q", author, "ada <ada@example.com>")
	}
	if name := mustGitIn(t, repo, "config", "--local", "user.name"); name != "ada" {
		t.Errorf("repository user.name = %q, want %q", name, "ada")
	}
	if email := mustGitIn(t, repo, "config", "--local", "user.email"); email != "ada@example.com" {
		t.Errorf("repository user.email = %q, want %q", email, "ada@example.com")
	}
}

func TestCommitFilesRecordsDeletionInNestedKB(t *testing.T) {
	isolateGitConfig(t)

	repo := t.TempDir()
	mustGitIn(t, repo, "init")
	mustGitIn(t, repo, "config", "user.name", "ada")
	mustGitIn(t, repo, "config", "user.email", "ada@example.com")

	// The KB lives in a subdirectory of the repository, so every path the
	// commit machinery resolves is relative to the KB root.
	kbRoot := filepath.Join(repo, "product")
	rawDir := filepath.Join(kbRoot, "raw")
	if err := os.MkdirAll(rawDir, 0750); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"data.csv", "files.log"} {
		if err := os.WriteFile(filepath.Join(rawDir, name), []byte(name+"\n"), 0600); err != nil { //nolint:gosec // test writing into its temp repository
			t.Fatal(err)
		}
	}
	mustGitIn(t, repo, "add", "--", "product")
	mustGitIn(t, repo, "commit", "-m", "add raw files")

	if err := os.Remove(filepath.Join(rawDir, "data.csv")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rawDir, "files.log"), []byte("# empty\n"), 0600); err != nil { //nolint:gosec // test writing into its temp repository
		t.Fatal(err)
	}

	if err := CommitFiles(kbRoot, "akb: raw delete data.csv", "raw/data.csv", "raw/files.log"); err != nil {
		t.Fatalf("CommitFiles: %v", err)
	}

	var recorded []string
	for _, line := range strings.Split(mustGitIn(t, repo, "show", "--name-only", "--format=", "HEAD"), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			recorded = append(recorded, line)
		}
	}
	sort.Strings(recorded)
	want := []string{"product/raw/data.csv", "product/raw/files.log"}
	if !reflect.DeepEqual(recorded, want) {
		t.Errorf("commit recorded %v, want %v", recorded, want)
	}
	if status := mustGitIn(t, repo, "status", "--porcelain"); status != "" {
		t.Errorf("repository is not clean after the commit:\n%s", status)
	}
}

func TestFilesystemProvider_ImplementsStorageProvider(_ *testing.T) {
	// Compile-time interface check
	var _ Provider = (*FilesystemProvider)(nil)
}

func TestGitProvider_ImplementsStorageProvider(_ *testing.T) {
	// Compile-time interface check
	var _ Provider = (*GitProvider)(nil)
}
