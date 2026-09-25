// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListPages(t *testing.T) {
	t.Run("empty KB returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		pages, err := listPages(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pages != nil {
			t.Errorf("expected nil for empty KB, got %v", pages)
		}
	})

	t.Run("single page", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(kbDir, "note.md"), []byte("content"), 0600); err != nil {
			t.Fatal(err)
		}

		pages, err := listPages(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pages) != 1 {
			t.Fatalf("expected 1 page, got %d", len(pages))
		}
		if pages[0].Path != "note.md" {
			t.Errorf("expected 'note.md', got %q", pages[0].Path)
		}
	})

	t.Run("alphabetical order", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		files := []string{"zebra.md", "apple.md", "mango.md"}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(kbDir, f), []byte("content"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		pages, err := listPages(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pages) != 3 {
			t.Fatalf("expected 3 pages, got %d", len(pages))
		}
		if pages[0].Path != "apple.md" || pages[1].Path != "mango.md" || pages[2].Path != "zebra.md" {
			t.Errorf("expected alphabetical order, got %v", pages)
		}
	})

	t.Run("excludes index.md and log.md", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		files := []string{"note.md", "index.md", "log.md", "other.md"}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(kbDir, f), []byte("content"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		pages, err := listPages(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pages) != 2 {
			t.Fatalf("expected 2 pages, got %d", len(pages))
		}
		for _, p := range pages {
			if p.Path == "index.md" || p.Path == "log.md" {
				t.Errorf("index.md or log.md should be excluded, got %q", p.Path)
			}
		}
	})

	t.Run("strips kb/ prefix from subdirectory pages", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		subDir := filepath.Join(kbDir, "notes")
		if err := os.MkdirAll(subDir, 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "note1.md"), []byte("content"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "note2.md"), []byte("content"), 0600); err != nil {
			t.Fatal(err)
		}

		pages, err := listPages(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pages) != 2 {
			t.Fatalf("expected 2 pages, got %d", len(pages))
		}
		for _, p := range pages {
			if strings.HasPrefix(p.Path, "kb/") {
				t.Errorf("path should not have kb/ prefix, got %q", p.Path)
			}
			if !strings.HasPrefix(p.Path, "notes/") {
				t.Errorf("expected notes/ prefix, got %q", p.Path)
			}
		}
	})

	t.Run("all features together", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatal(err)
		}

		sub1 := filepath.Join(kbDir, "aaa")
		sub2 := filepath.Join(kbDir, "zzz")
		if err := os.MkdirAll(sub1, 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(sub2, 0750); err != nil {
			t.Fatal(err)
		}

		files := []string{
			"index.md",
			"log.md",
			"aaa/first.md",
			"aaa/second.md",
			"zzz/top.md",
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(kbDir, f), []byte("content"), 0600); err != nil {
				t.Fatal(err)
			}
		}

		pages, err := listPages(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(pages) != 3 {
			t.Fatalf("expected 3 pages, got %d: %v", len(pages), pages)
		}

		if pages[0].Path != "aaa/first.md" || pages[1].Path != "aaa/second.md" || pages[2].Path != "zzz/top.md" {
			t.Errorf("expected alphabetical sorted, got %v", pages)
		}

		for _, p := range pages {
			if strings.HasPrefix(p.Path, "kb/") {
				t.Errorf("path should not have kb/ prefix, got %q", p.Path)
			}
		}
	})
}
