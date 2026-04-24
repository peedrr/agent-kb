package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/search"
)

func setupSearchTestKB(t *testing.T) string {
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

func setupSearchDB(t *testing.T, kbRoot string) *sql.DB {
	t.Helper()
	akbDir := filepath.Join(kbRoot, ".akb")
	dbPath := filepath.Join(akbDir, "search.db")

	sqlDB, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	if err := db.CreateSchema(sqlDB); err != nil {
		_ = sqlDB.Close() //nolint:errcheck // close error secondary to schema error
		t.Fatalf("create schema: %v", err)
	}

	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = sqlDB.Close() //nolint:errcheck // close error secondary to WAL error
		t.Fatalf("enable WAL: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	t.Cleanup(func() { _ = sqlDB.Close() }) //nolint:errcheck,gosec // test cleanup — failure is non-fatal
	return sqlDB
}

func indexTestPage(t *testing.T, sqlDB *sql.DB, path, title, content, tags, summary string) {
	t.Helper()
	searcher := search.NewSQLiteFTS5Searcher(sqlDB)
	if err := searcher.IndexPage(context.Background(), path, title, content, tags, summary); err != nil {
		t.Fatalf("index page %s: %v", path, err)
	}
}

func TestSearch_WithResults(t *testing.T) {
	kbRoot := setupSearchTestKB(t)
	sqlDB := setupSearchDB(t, kbRoot)

	indexTestPage(t, sqlDB, "kb/notes/test.md", "Test Note", "This is a test note about testing", "test", "A test note")
	indexTestPage(t, sqlDB, "kb/notes/other.md", "Other Note", "Something completely different", "other", "Another note")

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = runSearch(nil, []string{"test"})

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "kb/notes/test.md") {
		t.Errorf("expected output to contain 'kb/notes/test.md', got: %q", output)
	}
	if !strings.Contains(output, "Test Note") {
		t.Errorf("expected output to contain 'Test Note', got: %q", output)
	}
}

func TestSearch_NoResults(t *testing.T) {
	kbRoot := setupSearchTestKB(t)
	sqlDB := setupSearchDB(t, kbRoot)

	indexTestPage(t, sqlDB, "kb/notes/test.md", "Test Note", "This is a test note", "test", "A test note")

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = runSearch(nil, []string{"nonexistent"})

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if output != "" {
		t.Errorf("expected empty output for no results, got: %q", output)
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	kbRoot := setupSearchTestKB(t)
	setupSearchDB(t, kbRoot)

	err := runSearch(nil, []string{""})
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
	if !strings.Contains(err.Error(), "search query must not be empty") {
		t.Errorf("expected 'search query must not be empty' error, got: %v", err)
	}
}

func TestSearch_JSONOutput(t *testing.T) {
	kbRoot := setupSearchTestKB(t)
	sqlDB := setupSearchDB(t, kbRoot)

	indexTestPage(t, sqlDB, "kb/notes/test.md", "Test Note", "This is a test note about testing", "test", "A test note")

	origJSON := searchJSON
	searchJSON = true
	t.Cleanup(func() { searchJSON = origJSON })

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = runSearch(nil, []string{"test"})

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	var results []searchJSONResult
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %q", err, output)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	found := results[0]
	if found.Path != "kb/notes/test.md" {
		t.Errorf("path = %q, want %q", found.Path, "kb/notes/test.md")
	}
	if found.Title != "Test Note" {
		t.Errorf("title = %q, want %q", found.Title, "Test Note")
	}
	if found.Summary != "A test note" {
		t.Errorf("summary = %q, want %q", found.Summary, "A test note")
	}
	if found.Rank == 0 {
		t.Error("rank should not be zero")
	}
	if found.Snippet == "" {
		t.Error("snippet should not be empty")
	}
}

func TestSearch_MissingDB(t *testing.T) {
	kbRoot := t.TempDir()
	kbDir := filepath.Join(kbRoot, "kb")
	akbDir := filepath.Join(kbRoot, ".akb")
	if err := os.MkdirAll(kbDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(akbDir, 0750); err != nil {
		t.Fatal(err)
	}
	configContent := "name: test-kb\ncreated: \"2024-01-01T00:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(akbDir, ".akb.yaml"), []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kbDir, "index.md"), []byte("# Index\n\n"), 0600); err != nil {
		t.Fatal(err)
	}

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

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origCwd) }) //nolint:errcheck,gosec // test cleanup — failure is non-fatal
	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	err = runSearch(nil, []string{"test"})
	if err == nil {
		t.Fatal("expected error for missing DB, got nil")
	}
	if !strings.Contains(err.Error(), "akb index rebuild") {
		t.Errorf("expected error suggesting 'akb index rebuild', got: %v", err)
	}
}
