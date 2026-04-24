package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTempKB(t *testing.T, logContent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	kbDir := filepath.Join(tmpDir, "kb")
	if err := os.MkdirAll(kbDir, 0750); err != nil {
		t.Fatalf("mkdir kb: %v", err)
	}
	logPath := filepath.Join(tmpDir, "kb", "log.md")
	if err := os.WriteFile(logPath, []byte(logContent), 0600); err != nil {
		t.Fatalf("write log.md: %v", err)
	}
	return tmpDir
}

func TestReadLog(t *testing.T) {
	t.Run("fresh log with only header returns empty entries", func(t *testing.T) {
		kbRoot := setupTempKB(t, "# Log\n\n")
		entries, err := ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("parses single entry with title", func(t *testing.T) {
		content := "# Log\n\n## 2026-01-15 ingest | Note One\n\nIngested test notes\n"
		kbRoot := setupTempKB(t, content)
		entries, err := ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		e := entries[0]
		if e.Date != "2026-01-15" {
			t.Errorf("Date = %q, want %q", e.Date, "2026-01-15")
		}
		if e.Operation != "ingest" {
			t.Errorf("Operation = %q, want %q", e.Operation, "ingest")
		}
		if e.Title != "Note One" {
			t.Errorf("Title = %q, want %q", e.Title, "Note One")
		}
		if e.Description != "Ingested test notes" {
			t.Errorf("Description = %q, want %q", e.Description, "Ingested test notes")
		}
	})

	t.Run("parses entry without title", func(t *testing.T) {
		content := "# Log\n\n## 2026-01-15 delete\n\nRemoved something\n"
		kbRoot := setupTempKB(t, content)
		entries, err := ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].Title != "" {
			t.Errorf("Title = %q, want empty", entries[0].Title)
		}
	})

	t.Run("parses multiple entries", func(t *testing.T) {
		content := "# Log\n\n## 2026-01-15 ingest | Note One\n\nIngested test notes\n\n## 2026-01-16 delete | Deleted Page\n\nRemoved page notes/deleted.md from knowledge base\n"
		kbRoot := setupTempKB(t, content)
		entries, err := ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Operation != "ingest" {
			t.Errorf("entry[0] Operation = %q, want %q", entries[0].Operation, "ingest")
		}
		if entries[1].Operation != "delete" {
			t.Errorf("entry[1] Operation = %q, want %q", entries[1].Operation, "delete")
		}
	})

	t.Run("parses multiline description", func(t *testing.T) {
		content := "# Log\n\n## 2026-01-15 update | Big Update\n\nLine one\nLine two\nLine three\n"
		kbRoot := setupTempKB(t, content)
		entries, err := ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		expected := "Line one\nLine two\nLine three"
		if entries[0].Description != expected {
			t.Errorf("Description = %q, want %q", entries[0].Description, expected)
		}
	})

	t.Run("returns error when log.md does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatalf("mkdir kb: %v", err)
		}
		_, err := ReadLog(tmpDir)
		if err == nil {
			t.Fatal("expected error for missing log.md, got nil")
		}
	})
}

func TestAppendLog(t *testing.T) {
	t.Run("appends entry with auto-generated date", func(t *testing.T) {
		kbRoot := setupTempKB(t, "# Log\n\n")
		err := AppendLog(kbRoot, "ingest", "Ingested new notes", "My Note")
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}

		entries, err := ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		today := time.Now().Format("2006-01-02")
		if entries[0].Date != today {
			t.Errorf("Date = %q, want %q", entries[0].Date, today)
		}
		if entries[0].Operation != "ingest" {
			t.Errorf("Operation = %q, want %q", entries[0].Operation, "ingest")
		}
		if entries[0].Description != "Ingested new notes" {
			t.Errorf("Description = %q, want %q", entries[0].Description, "Ingested new notes")
		}
		if entries[0].Title != "My Note" {
			t.Errorf("Title = %q, want %q", entries[0].Title, "My Note")
		}
	})

	t.Run("appends entry with title includes title in heading", func(t *testing.T) {
		kbRoot := setupTempKB(t, "# Log\n\n")
		err := AppendLog(kbRoot, "delete", "Removed a page", "Deleted Page")
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}

		logPath := filepath.Join(kbRoot, "kb", "log.md")
		data, err := os.ReadFile(logPath) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("read log.md: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, " | Deleted Page") {
			t.Errorf("log.md should contain ' | Deleted Page', got:\n%s", content)
		}
	})

	t.Run("appends entry without title omits pipe", func(t *testing.T) {
		kbRoot := setupTempKB(t, "# Log\n\n")
		err := AppendLog(kbRoot, "lint", "Ran linter", "")
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}

		logPath := filepath.Join(kbRoot, "kb", "log.md")
		data, err := os.ReadFile(logPath) //nolint:gosec // test reading known temp file
		if err != nil {
			t.Fatalf("read log.md: %v", err)
		}
		content := string(data)
		if strings.Contains(content, " | ") {
			t.Errorf("log.md should NOT contain ' | ' when title is empty, got:\n%s", content)
		}
		if !strings.Contains(content, "## ") {
			t.Errorf("log.md should contain '## ' heading, got:\n%s", content)
		}
	})

	t.Run("appends to existing entries", func(t *testing.T) {
		content := "# Log\n\n## 2026-01-15 ingest | First\n\nFirst entry\n"
		kbRoot := setupTempKB(t, content)
		err := AppendLog(kbRoot, "update", "Updated something", "Second")
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}

		entries, err := ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Operation != "ingest" {
			t.Errorf("entry[0] Operation = %q, want %q", entries[0].Operation, "ingest")
		}
		if entries[1].Operation != "update" {
			t.Errorf("entry[1] Operation = %q, want %q", entries[1].Operation, "update")
		}
	})

	t.Run("creates log.md if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		kbDir := filepath.Join(tmpDir, "kb")
		if err := os.MkdirAll(kbDir, 0750); err != nil {
			t.Fatalf("mkdir kb: %v", err)
		}
		err := AppendLog(tmpDir, "query", "Searched for something", "Search")
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}

		logPath := filepath.Join(tmpDir, "kb", "log.md")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Fatal("AppendLog should have created log.md")
		}

		entries, err := ReadLog(tmpDir)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
	})
}

func TestFilterByType(t *testing.T) {
	entries := []LogEntry{
		{Date: "2026-01-15", Operation: "ingest", Title: "A", Description: "desc a"},
		{Date: "2026-01-15", Operation: "delete", Title: "B", Description: "desc b"},
		{Date: "2026-01-16", Operation: "ingest", Title: "C", Description: "desc c"},
		{Date: "2026-01-16", Operation: "update", Title: "D", Description: "desc d"},
	}

	t.Run("returns only matching operation type", func(t *testing.T) {
		result := FilterByType(entries, "ingest")
		if len(result) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(result))
		}
		for _, e := range result {
			if e.Operation != "ingest" {
				t.Errorf("Operation = %q, want %q", e.Operation, "ingest")
			}
		}
	})

	t.Run("returns empty slice for no matches", func(t *testing.T) {
		result := FilterByType(entries, "query")
		if len(result) != 0 {
			t.Errorf("expected 0 entries, got %d", len(result))
		}
		// Must be empty slice, not nil
		if result == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns empty slice for empty input", func(t *testing.T) {
		result := FilterByType([]LogEntry{}, "ingest")
		if len(result) != 0 {
			t.Errorf("expected 0 entries, got %d", len(result))
		}
		if result == nil {
			t.Error("expected empty slice, got nil")
		}
	})
}

func TestFilterByLast(t *testing.T) {
	entries := []LogEntry{
		{Date: "2026-01-15", Operation: "ingest", Title: "First", Description: "d1"},
		{Date: "2026-01-16", Operation: "delete", Title: "Second", Description: "d2"},
		{Date: "2026-01-17", Operation: "update", Title: "Third", Description: "d3"},
		{Date: "2026-01-18", Operation: "lint", Title: "Fourth", Description: "d4"},
	}

	t.Run("returns last N entries", func(t *testing.T) {
		result := FilterByLast(entries, 2)
		if len(result) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(result))
		}
		if result[0].Title != "Third" {
			t.Errorf("result[0].Title = %q, want %q", result[0].Title, "Third")
		}
		if result[1].Title != "Fourth" {
			t.Errorf("result[1].Title = %q, want %q", result[1].Title, "Fourth")
		}
	})

	t.Run("returns all entries when n > total", func(t *testing.T) {
		result := FilterByLast(entries, 10)
		if len(result) != 4 {
			t.Fatalf("expected 4 entries, got %d", len(result))
		}
	})

	t.Run("returns empty for empty input", func(t *testing.T) {
		result := FilterByLast([]LogEntry{}, 5)
		if len(result) != 0 {
			t.Errorf("expected 0 entries, got %d", len(result))
		}
	})

	t.Run("returns empty for n=0", func(t *testing.T) {
		result := FilterByLast(entries, 0)
		if len(result) != 0 {
			t.Errorf("expected 0 entries, got %d", len(result))
		}
	})
}

func TestRenderLog(t *testing.T) {
	t.Run("renders entries with title", func(t *testing.T) {
		entries := []LogEntry{
			{Date: "2026-01-15", Operation: "ingest", Title: "Note One", Description: "Ingested test notes"},
		}
		result := RenderLog(entries)
		if !strings.Contains(result, "# Log") {
			t.Error("rendered output should contain '# Log'")
		}
		if !strings.Contains(result, "## 2026-01-15 ingest | Note One") {
			t.Errorf("should contain heading with title, got:\n%s", result)
		}
		if !strings.Contains(result, "Ingested test notes") {
			t.Errorf("should contain description, got:\n%s", result)
		}
	})

	t.Run("renders entries without title", func(t *testing.T) {
		entries := []LogEntry{
			{Date: "2026-01-15", Operation: "delete", Title: "", Description: "Removed something"},
		}
		result := RenderLog(entries)
		if strings.Contains(result, " | ") {
			t.Errorf("should NOT contain pipe when no title, got:\n%s", result)
		}
		if !strings.Contains(result, "## 2026-01-15 delete") {
			t.Errorf("should contain heading without pipe, got:\n%s", result)
		}
	})

	t.Run("renders multiple entries", func(t *testing.T) {
		entries := []LogEntry{
			{Date: "2026-01-15", Operation: "ingest", Title: "Note One", Description: "Ingested test notes"},
			{Date: "2026-01-15", Operation: "delete", Title: "Deleted Page", Description: "Removed page notes/deleted.md from knowledge base"},
		}
		result := RenderLog(entries)
		if !strings.Contains(result, "## 2026-01-15 ingest | Note One") {
			t.Errorf("should contain first heading, got:\n%s", result)
		}
		if !strings.Contains(result, "## 2026-01-15 delete | Deleted Page") {
			t.Errorf("should contain second heading, got:\n%s", result)
		}
	})

	t.Run("renders empty entries as fresh log", func(t *testing.T) {
		result := RenderLog([]LogEntry{})
		expected := "# Log\n\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("round-trip: render then parse yields same entries", func(t *testing.T) {
		entries := []LogEntry{
			{Date: "2026-01-15", Operation: "ingest", Title: "Note One", Description: "Ingested test notes"},
			{Date: "2026-01-16", Operation: "delete", Title: "", Description: "Removed something"},
		}
		rendered := RenderLog(entries)

		kbRoot := setupTempKB(t, rendered)
		parsed, err := ReadLog(kbRoot)
		if err != nil {
			t.Fatalf("ReadLog failed: %v", err)
		}
		if len(parsed) != len(entries) {
			t.Fatalf("expected %d entries, got %d", len(entries), len(parsed))
		}
		for i, e := range entries {
			if parsed[i].Date != e.Date {
				t.Errorf("entry[%d].Date = %q, want %q", i, parsed[i].Date, e.Date)
			}
			if parsed[i].Operation != e.Operation {
				t.Errorf("entry[%d].Operation = %q, want %q", i, parsed[i].Operation, e.Operation)
			}
			if parsed[i].Title != e.Title {
				t.Errorf("entry[%d].Title = %q, want %q", i, parsed[i].Title, e.Title)
			}
			if parsed[i].Description != e.Description {
				t.Errorf("entry[%d].Description = %q, want %q", i, parsed[i].Description, e.Description)
			}
		}
	})
}
