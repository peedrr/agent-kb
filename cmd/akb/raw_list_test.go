// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRawList_EmptyRawDir(t *testing.T) {
	_ = setupLinksTestKB(t)

	output, err := captureOutput(func() error {
		return runRawList(nil, nil)
	})
	if err != nil {
		t.Fatalf("runRawList: %v", err)
	}

	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output for empty raw dir, got %q", output)
	}
}

func TestRawList_MissingRawDir(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	_ = os.RemoveAll(filepath.Join(kbRoot, "raw")) //nolint:errcheck // test cleanup — failure is non-fatal

	output, err := captureOutput(func() error {
		return runRawList(nil, nil)
	})
	if err != nil {
		t.Fatalf("runRawList: %v", err)
	}

	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output for missing raw dir, got %q", output)
	}
}

func TestRawList_WithFiles(t *testing.T) {
	kbRoot := setupLinksTestKB(t)

	rawDir := filepath.Join(kbRoot, "raw")
	if err := os.WriteFile(filepath.Join(rawDir, "config.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rawDir, "data.csv"), []byte("a,b\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(rawDir, "nested"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rawDir, "nested", "file.txt"), []byte("nested"), 0600); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = runRawList(nil, nil)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	if err != nil {
		t.Fatalf("runRawList: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 files, got %d: %q", len(lines), output)
	}

	expected := map[string]bool{
		"config.json":     false,
		"data.csv":        false,
		"nested/file.txt": false,
	}
	for _, line := range lines {
		if _, ok := expected[line]; ok {
			expected[line] = true
		} else {
			t.Errorf("unexpected file %q in output", line)
		}
	}
	for file, found := range expected {
		if !found {
			t.Errorf("expected file %q in output, not found", file)
		}
	}
}

func TestRawList_ExcludesFilesLog(t *testing.T) {
	kbRoot := setupLinksTestKB(t)

	rawDir := filepath.Join(kbRoot, "raw")
	if err := os.WriteFile(filepath.Join(rawDir, "config.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = runRawList(nil, nil)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	if err != nil {
		t.Fatalf("runRawList: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if strings.Contains(output, "files.log") {
		t.Errorf("files.log should be excluded, got: %q", output)
	}
	if !strings.Contains(output, "config.json") {
		t.Errorf("expected config.json in output, got: %q", output)
	}
}
