package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	kblog "github.com/peedrr/agent-kb/internal/log"
)

func TestLog(t *testing.T) {
	kbRoot := t.TempDir()
	setupLogTestKB(t, kbRoot)

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origCwd) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	t.Run("show on fresh KB shows nothing exit 0", func(t *testing.T) {
		cmd := logShowCmd
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetArgs([]string{})
		logShowLast = 0
		logShowType = ""

		err := runLogShow(cmd, []string{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if buf.String() != "" {
			t.Errorf("expected no output, got: %q", buf.String())
		}
	})

	t.Run("append creates log entry", func(t *testing.T) {
		cmd := logAppendCmd
		logAppendTitle = ""

		err := runLogAppend(cmd, []string{"ingest", "Added notes"})
		if err != nil {
			t.Fatalf("append failed: %v", err)
		}

		logPath := filepath.Join(kbRoot, "kb", "log.md")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Fatal("log.md should exist after append")
		}
	})

	t.Run("show after append shows entry", func(t *testing.T) {
		cmd := logShowCmd
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		logShowLast = 0
		logShowType = ""

		err := runLogShow(cmd, []string{})
		if err != nil {
			t.Fatalf("show failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "ingest") {
			t.Errorf("expected output to contain 'ingest', got: %s", output)
		}
		if !strings.Contains(output, "Added notes") {
			t.Errorf("expected output to contain 'Added notes', got: %s", output)
		}
	})

	t.Run("show --last 1 shows most recent entry only", func(t *testing.T) {
		logAppendTitle = ""
		err := runLogAppend(logAppendCmd, []string{"delete", "Removed something"})
		if err != nil {
			t.Fatalf("append second entry failed: %v", err)
		}

		cmd := logShowCmd
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		logShowLast = 1
		logShowType = ""

		err = runLogShow(cmd, []string{})
		if err != nil {
			t.Fatalf("show --last 1 failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "delete") {
			t.Errorf("expected output to contain 'delete', got: %s", output)
		}
		if strings.Contains(output, "Added notes") {
			t.Errorf("expected output NOT to contain first entry description, got: %s", output)
		}
	})

	t.Run("show --type ingest filters correctly", func(t *testing.T) {
		cmd := logShowCmd
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		logShowLast = 0
		logShowType = "ingest"

		err := runLogShow(cmd, []string{})
		if err != nil {
			t.Fatalf("show --type ingest failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "ingest") {
			t.Errorf("expected output to contain 'ingest', got: %s", output)
		}
		if strings.Contains(output, "delete") {
			t.Errorf("expected output NOT to contain 'delete', got: %s", output)
		}
	})

	t.Run("show --type nonexistent shows nothing exit 0", func(t *testing.T) {
		cmd := logShowCmd
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		logShowLast = 0
		logShowType = "nonexistent"

		err := runLogShow(cmd, []string{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if buf.String() != "" {
			t.Errorf("expected no output for nonexistent type, got: %q", buf.String())
		}
	})

	t.Run("invalid operation type rejected", func(t *testing.T) {
		cmd := logAppendCmd
		logAppendTitle = ""

		err := runLogAppend(cmd, []string{"invalid_op", "Bad operation"})
		if err == nil {
			t.Fatal("expected error for invalid operation, got nil")
		}
		if !strings.Contains(err.Error(), "invalid operation") {
			t.Errorf("expected 'invalid operation' in error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "ingest") || !strings.Contains(err.Error(), "delete") {
			t.Errorf("expected valid operations listed in error, got: %v", err)
		}
	})

	t.Run("append with --title flag", func(t *testing.T) {
		cmd := logAppendCmd
		logAppendTitle = "My Title"

		err := runLogAppend(cmd, []string{"update", "Updated content"})
		if err != nil {
			t.Fatalf("append with title failed: %v", err)
		}

		entries, err := kblog.ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}

		found := false
		for _, e := range entries {
			if e.Operation == "update" && e.Title == "My Title" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find update entry with title 'My Title'")
		}
	})

	t.Run("append with --no-commit skips git commit", func(t *testing.T) {
		origNoCommit := noCommit
		noCommit = true
		defer func() { noCommit = origNoCommit }()

		cmd := logAppendCmd
		logAppendTitle = ""

		err := runLogAppend(cmd, []string{"lint", "Ran linter"})
		if err != nil {
			t.Fatalf("append with --no-commit failed: %v", err)
		}

		entries, err := kblog.ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		found := false
		for _, e := range entries {
			if e.Operation == "lint" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find lint entry after --no-commit append")
		}
	})
}

func setupLogTestKB(t *testing.T, kbRoot string) {
	t.Helper()
	dirs := []string{
		filepath.Join(kbRoot, "kb"),
		filepath.Join(kbRoot, ".akb"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatal(err)
		}
	}

	configContent := "name: test-kb\ncreated: \"2024-01-01T00:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(kbRoot, ".akb", ".akb.yaml"), []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "index.md"), []byte("# Index\n\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := initGitRepo(kbRoot); err != nil {
		t.Logf("git not available: %v", err)
	}
	if err := addToGit(kbRoot, "."); err != nil {
		t.Logf("git add initial files failed: %v", err)
	}
	commitInitial(kbRoot)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", kbRoot)                         //nolint:errcheck,gosec // test setup — failure is non-fatal
	t.Cleanup(func() { os.Setenv("HOME", origHome) }) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	regPath := filepath.Join(kbRoot, ".config", "agent-kb", "registry.yaml")
	if err := os.MkdirAll(filepath.Dir(regPath), 0750); err != nil {
		t.Fatal(err)
	}
	regContent := `default: test-kb
entries:
  - name: test-kb
    path: ` + kbRoot + `
    created: "2024-01-01T00:00:00Z"
`
	if err := os.WriteFile(regPath, []byte(regContent), 0600); err != nil {
		t.Fatal(err)
	}
}
