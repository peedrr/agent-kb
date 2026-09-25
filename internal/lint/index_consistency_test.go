// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestIndexConsistencyChecker(t *testing.T) {
	checker := NewIndexConsistencyChecker()

	if checker.Name() != "index_consistency" {
		t.Errorf("expected name 'index_consistency', got %q", checker.Name())
	}

	t.Run("page in index but not on disk", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		indexContent := `# Index

## Notes

- [Existing Page](kb/notes/existing.md)
- [Missing Page](kb/notes/missing.md)
`
		if err := os.WriteFile(filepath.Join(kbDir, "index.md"), []byte(indexContent), 0600); err != nil {
			t.Fatal(err)
		}

		if err := os.MkdirAll(filepath.Join(kbDir, "notes"), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(kbDir, "notes", "existing.md"), []byte("---\ntitle: Existing\ntype: note\n---\nContent"), 0600); err != nil {
			t.Fatal(err)
		}

		kb := &KB{
			Root: tmpDir,
			Pages: []PageData{
				{RelPath: "kb/notes/existing.md"},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}

		if issues[0].Type != "index_consistency" {
			t.Errorf("expected issue type 'index_consistency', got %q", issues[0].Type)
		}
		if issues[0].Path != "kb/notes/missing.md" {
			t.Errorf("expected path 'kb/notes/missing.md', got %q", issues[0].Path)
		}
		if issues[0].Severity != "error" {
			t.Errorf("expected severity 'error', got %q", issues[0].Severity)
		}
	})

	t.Run("page on disk but not in index", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		indexContent := `# Index

## Notes

- [Indexed Page](kb/notes/indexed.md)
`
		if err := os.WriteFile(filepath.Join(kbDir, "index.md"), []byte(indexContent), 0600); err != nil {
			t.Fatal(err)
		}

		if err := os.MkdirAll(filepath.Join(kbDir, "notes"), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(kbDir, "notes", "indexed.md"), []byte("---\ntitle: Indexed\ntype: note\n---\nContent"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(kbDir, "notes", "unindexed.md"), []byte("---\ntitle: Unindexed\ntype: note\n---\nContent"), 0600); err != nil {
			t.Fatal(err)
		}

		kb := &KB{
			Root: tmpDir,
			Pages: []PageData{
				{RelPath: "kb/notes/indexed.md"},
				{RelPath: "kb/notes/unindexed.md"},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}

		if issues[0].Type != "index_consistency" {
			t.Errorf("expected issue type 'index_consistency', got %q", issues[0].Type)
		}
		if issues[0].Path != "kb/notes/unindexed.md" {
			t.Errorf("expected path 'kb/notes/unindexed.md', got %q", issues[0].Path)
		}
		if issues[0].Severity != "warning" {
			t.Errorf("expected severity 'warning', got %q", issues[0].Severity)
		}
	})

	t.Run("both missing from index and missing from disk", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		indexContent := `# Index

## Notes

- [Missing from disk](kb/notes/missing.md)
- [Also missing](kb/notes/also.md)
`
		if err := os.WriteFile(filepath.Join(kbDir, "index.md"), []byte(indexContent), 0600); err != nil {
			t.Fatal(err)
		}

		if err := os.MkdirAll(filepath.Join(kbDir, "notes"), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(kbDir, "notes", "new.md"), []byte("---\ntitle: New\ntype: note\n---\nContent"), 0600); err != nil {
			t.Fatal(err)
		}

		kb := &KB{
			Root: tmpDir,
			Pages: []PageData{
				{RelPath: "kb/notes/new.md"},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(issues) != 3 {
			t.Fatalf("expected 3 issues, got %d", len(issues))
		}

		errorCount := 0
		warningCount := 0
		for _, issue := range issues {
			if issue.Severity == "error" {
				errorCount++
			}
			if issue.Severity == "warning" {
				warningCount++
			}
		}

		if errorCount != 2 {
			t.Errorf("expected 2 errors, got %d", errorCount)
		}
		if warningCount != 1 {
			t.Errorf("expected 1 warning, got %d", warningCount)
		}
	})

	t.Run("index.md does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		if err := os.MkdirAll(filepath.Join(kbDir, "notes"), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(kbDir, "notes", "test.md"), []byte("---\ntitle: Test\ntype: note\n---\nContent"), 0600); err != nil {
			t.Fatal(err)
		}

		kb := &KB{
			Root: tmpDir,
			Pages: []PageData{
				{RelPath: "kb/notes/test.md"},
			},
		}

		_, err := checker.Check(context.Background(), kb)
		if err == nil {
			t.Fatal("expected error when index.md doesn't exist")
		}
	})

	t.Run("consistent index and pages", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		indexContent := `# Index

## Notes

- [Page One](kb/notes/one.md)
- [Page Two](kb/notes/two.md)
`
		if err := os.WriteFile(filepath.Join(kbDir, "index.md"), []byte(indexContent), 0600); err != nil {
			t.Fatal(err)
		}

		if err := os.MkdirAll(filepath.Join(kbDir, "notes"), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(kbDir, "notes", "one.md"), []byte("---\ntitle: One\ntype: note\n---\nContent"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(kbDir, "notes", "two.md"), []byte("---\ntitle: Two\ntype: note\n---\nContent"), 0600); err != nil {
			t.Fatal(err)
		}

		kb := &KB{
			Root: tmpDir,
			Pages: []PageData{
				{RelPath: "kb/notes/one.md"},
				{RelPath: "kb/notes/two.md"},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(issues) != 0 {
			t.Errorf("expected no issues when index and pages are consistent, got %d", len(issues))
		}
	})
}
