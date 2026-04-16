package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

// TestReadCmd_ExistingPage tests reading an existing page successfully
func TestReadCmd_ExistingPage(t *testing.T) {
	// Create temp KB structure
	tmpDir := t.TempDir()
	setupTestKBForRead(t, tmpDir)

	// Create a test page with frontmatter
	kbDir := filepath.Join(tmpDir, "kb")
	pagePath := filepath.Join(kbDir, "my-note.md")
	pageContent := `---
title: My Note
created: 2024-01-01
---

# My Note

This is the content of my note.`
	os.WriteFile(pagePath, []byte(pageContent), 0644)

	// Change to the KB directory
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Run the read command
	kbRoot, err := path.KBRoot()
	if err != nil {
		t.Fatalf("KBRoot failed: %v", err)
	}

	provider := storage.NewGitProvider(kbRoot, true)

	// Test reading the page
	testCases := []struct {
		name      string
		inputPath string
	}{
		{"my-note.md", "my-note.md"},
		{"kb/my-note.md", "kb/my-note.md"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanPath := strings.TrimPrefix(tc.inputPath, "kb/")
			relPath := filepath.Join(kbRoot, "kb", cleanPath)

			data, err := provider.Read(context.Background(), relPath)
			if err != nil {
				t.Fatalf("Read %s failed: %v", relPath, err)
			}

			if string(data) != pageContent {
				t.Errorf("expected %q, got %q", pageContent, string(data))
			}
		})
	}
}

// TestReadCmd_NonExistentPage tests error on non-existent page
func TestReadCmd_NonExistentPage(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestKBForRead(t, tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	kbRoot, _ := path.KBRoot()
	provider := storage.NewGitProvider(kbRoot, true)

	cleanPath := strings.TrimPrefix("notes/non-existent.md", "kb/")
	relPath := filepath.Join(kbRoot, "kb", cleanPath)
	_, err := provider.Read(context.Background(), relPath)

	if err == nil {
		t.Fatal("expected error for non-existent page")
	}
}

// TestReadCmd_RawPrefix tests error on raw/ prefix path
func TestReadCmd_RawPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestKBForRead(t, tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	kbRoot, _ := path.KBRoot()

	_, err := path.ResolveKBPath(kbRoot, "raw/some-file.md")

	if err == nil {
		t.Fatal("expected error for raw/ prefix")
	}
}

// TestReadCmd_ParentDir tests error on .. path
func TestReadCmd_ParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestKBForRead(t, tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	kbRoot, _ := path.KBRoot()

	_, err := path.ResolveKBPath(kbRoot, "../some-file.md")

	if err == nil {
		t.Fatal("expected error for .. path")
	}
}

// TestReadCmd_ContentOutput tests that full content including frontmatter is output
func TestReadCmd_ContentOutput(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestKBForRead(t, tmpDir)

	kbDir := filepath.Join(tmpDir, "kb")
	pagePath := filepath.Join(kbDir, "test.md")
	expectedContent := `---
title: Test
---

# Content here`

	os.WriteFile(pagePath, []byte(expectedContent), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	kbRoot, _ := path.KBRoot()
	provider := storage.NewGitProvider(kbRoot, true)

	cleanPath := strings.TrimPrefix("test.md", "kb/")
	relPath := filepath.Join(kbRoot, "kb", cleanPath)
	data, _ := provider.Read(context.Background(), relPath)

	if string(data) != expectedContent {
		t.Errorf("expected full content with frontmatter, got %q", string(data))
	}
}

// setupTestKB creates a minimal KB structure for testing
func setupTestKBForRead(t *testing.T, tmpDir string) {
	akbDir := filepath.Join(tmpDir, ".akb")
	os.MkdirAll(akbDir, 0755)

	configContent := `name: test-kb
created: "2024-01-01T00:00:00Z"`
	os.WriteFile(filepath.Join(akbDir, ".akb.yaml"), []byte(configContent), 0644)

	os.MkdirAll(filepath.Join(tmpDir, "kb"), 0755)

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = tmpDir
	cmd.Run()
}
