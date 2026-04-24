package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/index"
)

func setupIndexTestKB(t *testing.T) string {
	t.Helper()
	kbRoot := t.TempDir()
	setupTestKBWithGit(t, kbRoot)

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origCwd) }) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	return kbRoot
}

func TestIndexShow_FreshKB(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = runIndexShow(nil, nil)

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	if err != nil {
		t.Fatalf("index show failed: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	expected := "# Index\n\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}

	indexPath := filepath.Join(kbRoot, "kb", "index.md")
	data, err := os.ReadFile(indexPath) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatalf("read index file: %v", err)
	}
	if string(data) != expected {
		t.Errorf("file content: expected %q, got %q", expected, string(data))
	}
}

func TestIndexAdd_CreatesHeadingAndEntry(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	notesDir := filepath.Join(kbRoot, "kb", "notes")
	if err := os.MkdirAll(notesDir, 0750); err != nil {
		t.Fatal(err)
	}

	pageContent := "---\ntype: note\ntitle: My Note\n---\nNote body.\n"
	pagePath := filepath.Join(notesDir, "my-note.md")
	if err := os.WriteFile(pagePath, []byte(pageContent), 0600); err != nil {
		t.Fatal(err)
	}

	if err := addToGit(kbRoot, "kb/notes/my-note.md"); err != nil {
		t.Fatal(err)
	}

	err := runIndexAdd(nil, []string{"kb/notes/my-note.md", "A test note"})
	if err != nil {
		t.Fatalf("index add failed: %v", err)
	}

	entries, err := index.ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "kb/notes/my-note.md" {
		t.Errorf("path = %q, want %q", entries[0].Path, "kb/notes/my-note.md")
	}
	if entries[0].Title != "My Note" {
		t.Errorf("title = %q, want %q", entries[0].Title, "My Note")
	}
	if entries[0].Summary != "A test note" {
		t.Errorf("summary = %q, want %q", entries[0].Summary, "A test note")
	}
	if entries[0].Type != "note" {
		t.Errorf("type = %q, want %q", entries[0].Type, "note")
	}
}

func TestIndexAdd_UpdatesExistingPath(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	notesDir := filepath.Join(kbRoot, "kb", "notes")
	if err := os.MkdirAll(notesDir, 0750); err != nil {
		t.Fatal(err)
	}

	pageContent := "---\ntype: note\ntitle: My Note\n---\nNote body.\n"
	pagePath := filepath.Join(notesDir, "my-note.md")
	if err := os.WriteFile(pagePath, []byte(pageContent), 0600); err != nil {
		t.Fatal(err)
	}

	if err := addToGit(kbRoot, "kb/notes/my-note.md"); err != nil {
		t.Fatal(err)
	}

	err := runIndexAdd(nil, []string{"kb/notes/my-note.md", "Original summary"})
	if err != nil {
		t.Fatalf("first index add failed: %v", err)
	}

	err = runIndexAdd(nil, []string{"kb/notes/my-note.md", "Updated summary"})
	if err != nil {
		t.Fatalf("second index add failed: %v", err)
	}

	entries, err := index.ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Summary != "Updated summary" {
		t.Errorf("summary = %q, want %q", entries[0].Summary, "Updated summary")
	}
}

func TestIndexAdd_RejectsIndexMd(t *testing.T) {
	setupIndexTestKB(t)

	err := runIndexAdd(nil, []string{"kb/index.md", "should fail"})
	if err == nil {
		t.Error("expected error for index.md, got nil")
	}
}

func TestIndexAdd_RejectsLogMd(t *testing.T) {
	setupIndexTestKB(t)

	err := runIndexAdd(nil, []string{"kb/log.md", "should fail"})
	if err == nil {
		t.Error("expected error for log.md, got nil")
	}
}

func TestIndexAdd_RejectsParentDir(t *testing.T) {
	setupIndexTestKB(t)

	err := runIndexAdd(nil, []string{"../escape.md", "should fail"})
	if err == nil {
		t.Fatal("expected error for .. path, got nil")
	}
	if !strings.Contains(err.Error(), "..") {
		t.Errorf("expected error to contain '..', got: %v", err)
	}
}

func TestIndexAdd_RejectsAbsolutePath(t *testing.T) {
	setupIndexTestKB(t)

	err := runIndexAdd(nil, []string{"/tmp/evil.md", "should fail"})
	if err == nil {
		t.Fatal("expected error for absolute path, got nil")
	}
	if !strings.Contains(err.Error(), "absolute") && !strings.Contains(err.Error(), "relative") {
		t.Errorf("expected error about absolute/relative path, got: %v", err)
	}
}

func TestIndexRemove_RemovesEntry(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	entry := index.IndexEntry{
		Path:    "kb/notes/my-note.md",
		Title:   "My Note",
		Summary: "A test note",
		Type:    "note",
	}
	if err := index.AddEntry(kbRoot, entry); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	err := runIndexRemove(nil, []string{"kb/notes/my-note.md"})
	if err != nil {
		t.Fatalf("index remove failed: %v", err)
	}

	entries, err := index.ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestIndexRemove_NonexistentPath(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	entry := index.IndexEntry{
		Path:    "kb/notes/my-note.md",
		Title:   "My Note",
		Summary: "A test note",
		Type:    "note",
	}
	if err := index.AddEntry(kbRoot, entry); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	err := runIndexRemove(nil, []string{"kb/notes/nonexistent.md"})
	if err != nil {
		t.Fatalf("expected remove of nonexistent path to succeed, got: %v", err)
	}

	entries, err := index.ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after removing nonexistent path, got %d", len(entries))
	}
}

func TestIndexRemove_RejectsParentDir(t *testing.T) {
	setupIndexTestKB(t)

	err := runIndexRemove(nil, []string{"../escape.md"})
	if err == nil {
		t.Fatal("expected error for .. path, got nil")
	}
	if !strings.Contains(err.Error(), "..") {
		t.Errorf("expected error to contain '..', got: %v", err)
	}
}

func TestIndexRemove_RejectsAbsolutePath(t *testing.T) {
	setupIndexTestKB(t)

	err := runIndexRemove(nil, []string{"/tmp/evil.md"})
	if err == nil {
		t.Fatal("expected error for absolute path, got nil")
	}
	if !strings.Contains(err.Error(), "absolute") && !strings.Contains(err.Error(), "relative") {
		t.Errorf("expected error about absolute/relative path, got: %v", err)
	}
}

func TestIndexRebuild_RegeneratesFromFilesystem(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	notesDir := filepath.Join(kbRoot, "kb", "notes")
	decisionsDir := filepath.Join(kbRoot, "kb", "decisions")

	if err := os.MkdirAll(notesDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(decisionsDir, 0750); err != nil {
		t.Fatal(err)
	}

	noteContent := "---\ntype: note\ntitle: My Note\nsummary: A test note\n---\nNote body.\n"
	if err := os.WriteFile(filepath.Join(notesDir, "my-note.md"), []byte(noteContent), 0600); err != nil {
		t.Fatal(err)
	}

	adrContent := "---\ntype: adr\ntitle: My ADR\nsummary: A test ADR\n---\nADR body.\n"
	if err := os.WriteFile(filepath.Join(decisionsDir, "my-adr.md"), []byte(adrContent), 0600); err != nil {
		t.Fatal(err)
	}

	err := runIndexRebuild(nil, nil)
	if err != nil {
		t.Fatalf("index rebuild failed: %v", err)
	}

	entries, err := index.ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	foundNote := false
	foundADR := false
	for _, e := range entries {
		if e.Type == "note" && e.Title == "My Note" {
			foundNote = true
		}
		if e.Type == "adr" && e.Title == "My ADR" {
			foundADR = true
		}
	}
	if !foundNote {
		t.Error("expected to find note entry")
	}
	if !foundADR {
		t.Error("expected to find ADR entry")
	}
}

func TestIndexRebuild_RepopulatesSQLite(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	notesDir := filepath.Join(kbRoot, "kb", "notes")
	if err := os.MkdirAll(notesDir, 0750); err != nil {
		t.Fatal(err)
	}

	pageContent := "---\ntype: note\ntitle: My Note\ntags: testing\nsummary: A test note\n---\nNote body.\n"
	pagePath := filepath.Join(notesDir, "my-note.md")
	if err := os.WriteFile(pagePath, []byte(pageContent), 0600); err != nil {
		t.Fatal(err)
	}

	if err := addToGit(kbRoot, "kb/notes/my-note.md"); err != nil {
		t.Fatal(err)
	}

	err := runIndexRebuild(nil, nil)
	if err != nil {
		t.Fatalf("index rebuild failed: %v", err)
	}

	sqlDB, err := db.OpenKB(kbRoot)
	if err != nil {
		t.Fatalf("OpenKB failed: %v", err)
	}
	defer sqlDB.Close() //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	var docCount int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 1 {
		t.Errorf("expected 1 document, got %d", docCount)
	}

	var pageCount int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM pages").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("expected 1 page, got %d", pageCount)
	}

	var title, summary string
	if err := sqlDB.QueryRow("SELECT title, summary FROM pages WHERE path = ?", "kb/notes/my-note.md").Scan(&title, &summary); err != nil {
		t.Fatalf("query page: %v", err)
	}
	if title != "My Note" {
		t.Errorf("expected title %q, got %q", "My Note", title)
	}
	if summary != "A test note" {
		t.Errorf("expected summary %q, got %q", "A test note", summary)
	}
}

func TestIndexRebuild_EmptyKBSucceeds(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	err := runIndexRebuild(nil, nil)
	if err != nil {
		t.Fatalf("index rebuild on empty KB failed: %v", err)
	}

	sqlDB, err := db.OpenKB(kbRoot)
	if err != nil {
		t.Fatalf("OpenKB failed: %v", err)
	}
	defer sqlDB.Close() //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	var docCount int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 0 {
		t.Errorf("expected 0 documents after rebuild, got %d", docCount)
	}

	var pageCount int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM pages").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 0 {
		t.Errorf("expected 0 pages after rebuild, got %d", pageCount)
	}
}

func TestIndexRebuild_CreatesSearchDB(t *testing.T) {
	kbRoot := setupIndexTestKB(t)

	origNoCommit := noCommit
	noCommit = true
	t.Cleanup(func() { noCommit = origNoCommit })

	dbPath := filepath.Join(kbRoot, ".akb", "search.db")
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	notesDir := filepath.Join(kbRoot, "kb", "notes")
	if err := os.MkdirAll(notesDir, 0750); err != nil {
		t.Fatal(err)
	}

	pageContent := "---\ntype: note\ntitle: Test Note\nsummary: A test\n---\nContent.\n"
	pagePath := filepath.Join(notesDir, "test.md")
	if err := os.WriteFile(pagePath, []byte(pageContent), 0600); err != nil {
		t.Fatal(err)
	}

	if err := addToGit(kbRoot, "kb/notes/test.md"); err != nil {
		t.Fatal(err)
	}

	if err := runIndexRebuild(nil, nil); err != nil {
		t.Fatalf("index rebuild failed: %v", err)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected search.db to be created after rebuild")
	}

	sqlDB, err := db.OpenKB(kbRoot)
	if err != nil {
		t.Fatalf("OpenKB failed: %v", err)
	}
	defer sqlDB.Close() //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	var pageCount int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM pages").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("expected 1 page in new DB, got %d", pageCount)
	}
}
