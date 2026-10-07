// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package index

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupKB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	kbDir := filepath.Join(dir, "kb")
	if err := os.MkdirAll(kbDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kbDir, "index.md"), []byte("# Index\n\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeIndexContent(t *testing.T, kbRoot, content string) {
	t.Helper()
	path := filepath.Join(kbRoot, "kb", "index.md")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func readIndexContent(t *testing.T, kbRoot string) string {
	t.Helper()
	path := filepath.Join(kbRoot, "kb", "index.md")
	data, err := os.ReadFile(path) //nolint:gosec // test reading known temp file
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestReadIndex_Fresh(t *testing.T) {
	kbRoot := setupKB(t)
	entries, err := ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestReadIndex_WithEntries(t *testing.T) {
	kbRoot := setupKB(t)
	writeIndexContent(t, kbRoot, "# Index\n\n## Notes\n\n- [My Note](kb/notes/my-note.md) \u2014 A test note\n\n")

	entries, err := ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Title != "My Note" {
		t.Errorf("Title = %q, want %q", entries[0].Title, "My Note")
	}
	if entries[0].Path != "kb/notes/my-note.md" {
		t.Errorf("Path = %q, want %q", entries[0].Path, "kb/notes/my-note.md")
	}
	if entries[0].Summary != "A test note" {
		t.Errorf("Summary = %q, want %q", entries[0].Summary, "A test note")
	}
	if entries[0].Type != "note" {
		t.Errorf("Type = %q, want %q", entries[0].Type, "note")
	}
}

func TestReadIndex_MultipleTypes(t *testing.T) {
	kbRoot := setupKB(t)
	content := "# Index\n\n## Notes\n\n- [Note 1](kb/notes/note-1.md) \u2014 First note\n\n## Adrs\n\n- [ADR 1](kb/decisions/adr-1.md) \u2014 First ADR\n\n"
	writeIndexContent(t, kbRoot, content)

	entries, err := ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Type != "note" {
		t.Errorf("entries[0].Type = %q, want %q", entries[0].Type, "note")
	}
	if entries[1].Type != "adr" {
		t.Errorf("entries[1].Type = %q, want %q", entries[1].Type, "adr")
	}
}

func TestAddEntry_CreatesHeadingAndEntry(t *testing.T) {
	kbRoot := setupKB(t)

	entry := IndexEntry{
		Path:    "kb/notes/my-note.md",
		Title:   "My Note",
		Summary: "A test note",
		Type:    "note",
	}

	if err := AddEntry(kbRoot, entry); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	content := readIndexContent(t, kbRoot)
	expected := "# Index\n\n## Notes\n\n- [My Note](kb/notes/my-note.md) \u2014 A test note\n\n"
	if content != expected {
		t.Errorf("index content:\ngot:\n%s\nwant:\n%s", content, expected)
	}
}

func TestAddEntry_UpdatesExistingPath(t *testing.T) {
	kbRoot := setupKB(t)

	entry1 := IndexEntry{
		Path:    "kb/notes/my-note.md",
		Title:   "My Note",
		Summary: "Original summary",
		Type:    "note",
	}
	if err := AddEntry(kbRoot, entry1); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	entry2 := IndexEntry{
		Path:    "kb/notes/my-note.md",
		Title:   "Updated Note",
		Summary: "Updated summary",
		Type:    "note",
	}
	if err := AddEntry(kbRoot, entry2); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	entries, err := ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Title != "Updated Note" {
		t.Errorf("Title = %q, want %q", entries[0].Title, "Updated Note")
	}
	if entries[0].Summary != "Updated summary" {
		t.Errorf("Summary = %q, want %q", entries[0].Summary, "Updated summary")
	}
}

func TestAddEntry_AppendsToExistingHeading(t *testing.T) {
	kbRoot := setupKB(t)

	entry1 := IndexEntry{
		Path:    "kb/notes/note-1.md",
		Title:   "Note 1",
		Summary: "First note",
		Type:    "note",
	}
	if err := AddEntry(kbRoot, entry1); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	entry2 := IndexEntry{
		Path:    "kb/notes/note-2.md",
		Title:   "Note 2",
		Summary: "Second note",
		Type:    "note",
	}
	if err := AddEntry(kbRoot, entry2); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	entries, err := ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Type != "note" || entries[1].Type != "note" {
		t.Errorf("expected both entries to have type 'note'")
	}
}

func TestRemoveEntry_RemovesEntryAndCleansEmptyHeadings(t *testing.T) {
	kbRoot := setupKB(t)

	noteEntry := IndexEntry{
		Path:    "kb/notes/my-note.md",
		Title:   "My Note",
		Summary: "A test note",
		Type:    "note",
	}
	adrEntry := IndexEntry{
		Path:    "kb/decisions/my-adr.md",
		Title:   "My ADR",
		Summary: "A test ADR",
		Type:    "adr",
	}
	if err := AddEntry(kbRoot, noteEntry); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}
	if err := AddEntry(kbRoot, adrEntry); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	if err := RemoveEntry(kbRoot, "kb/notes/my-note.md"); err != nil {
		t.Fatalf("RemoveEntry failed: %v", err)
	}

	content := readIndexContent(t, kbRoot)
	if strings.Contains(content, "## Notes") {
		t.Error("expected Notes heading to be removed")
	}
	if !strings.Contains(content, "## Adrs") {
		t.Error("expected Adrs heading to remain")
	}

	entries, err := ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "kb/decisions/my-adr.md" {
		t.Errorf("expected remaining entry to be the ADR, got %q", entries[0].Path)
	}
}

func TestRemoveEntry_NonexistentPath(t *testing.T) {
	kbRoot := setupKB(t)

	entry := IndexEntry{
		Path:    "kb/notes/my-note.md",
		Title:   "My Note",
		Summary: "A test note",
		Type:    "note",
	}
	if err := AddEntry(kbRoot, entry); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}

	if err := RemoveEntry(kbRoot, "kb/notes/nonexistent.md"); err != nil {
		t.Fatalf("RemoveEntry should succeed silently for nonexistent path, got: %v", err)
	}

	entries, err := ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after removing nonexistent path, got %d", len(entries))
	}
}

func TestRebuildIndex(t *testing.T) {
	dir := t.TempDir()
	kbDir := filepath.Join(dir, "kb")
	notesDir := filepath.Join(kbDir, "notes")
	decisionsDir := filepath.Join(kbDir, "decisions")

	if err := os.MkdirAll(notesDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(decisionsDir, 0750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(kbDir, "index.md"), []byte("# Index\n\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kbDir, "log.md"), []byte("# Log\n\n"), 0600); err != nil {
		t.Fatal(err)
	}

	noteContent := "---\ntype: note\ntitle: My Note\nsummary: A test note\n---\nNote body here.\n"
	if err := os.WriteFile(filepath.Join(notesDir, "my-note.md"), []byte(noteContent), 0600); err != nil {
		t.Fatal(err)
	}

	adrContent := "---\ntype: adr\ntitle: My ADR\nsummary: A test ADR\n---\nADR body here.\n"
	if err := os.WriteFile(filepath.Join(decisionsDir, "my-adr.md"), []byte(adrContent), 0600); err != nil {
		t.Fatal(err)
	}

	if err := RebuildIndex(dir); err != nil {
		t.Fatalf("RebuildIndex failed: %v", err)
	}

	entries, err := ReadIndex(dir)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	foundNote := false
	foundADR := false
	for _, e := range entries {
		if e.Type == "note" && e.Title == "My Note" {
			foundNote = true
			if e.Summary != "A test note" {
				t.Errorf("note summary = %q, want %q", e.Summary, "A test note")
			}
		}
		if e.Type == "adr" && e.Title == "My ADR" {
			foundADR = true
			if e.Summary != "A test ADR" {
				t.Errorf("ADR summary = %q, want %q", e.Summary, "A test ADR")
			}
		}
	}
	if !foundNote {
		t.Error("expected to find note entry")
	}
	if !foundADR {
		t.Error("expected to find ADR entry")
	}
}

func TestRenderIndex(t *testing.T) {
	entries := []IndexEntry{
		{Path: "kb/notes/note-1.md", Title: "Note 1", Summary: "First note", Type: "note"},
		{Path: "kb/notes/note-2.md", Title: "Note 2", Summary: "Second note", Type: "note"},
		{Path: "kb/decisions/adr-1.md", Title: "ADR 1", Summary: "First ADR", Type: "adr"},
	}

	result := RenderIndex(entries)
	expected := "# Index\n\n## Notes\n\n- [Note 1](kb/notes/note-1.md) \u2014 First note\n- [Note 2](kb/notes/note-2.md) \u2014 Second note\n\n## Adrs\n\n- [ADR 1](kb/decisions/adr-1.md) \u2014 First ADR\n\n"

	if result != expected {
		t.Errorf("RenderIndex()\ngot:\n%s\nwant:\n%s", result, expected)
	}
}

func TestRenderIndex_Empty(t *testing.T) {
	result := RenderIndex([]IndexEntry{})
	expected := "# Index\n\n"
	if result != expected {
		t.Errorf("RenderIndex() = %q, want %q", result, expected)
	}
}

func TestRenderIndex_EntryWithoutSummary(t *testing.T) {
	entries := []IndexEntry{
		{Path: "kb/notes/note-1.md", Title: "Note 1", Summary: "", Type: "note"},
	}

	result := RenderIndex(entries)
	expected := "# Index\n\n## Notes\n\n- [Note 1](kb/notes/note-1.md)\n\n"

	if result != expected {
		t.Errorf("RenderIndex()\ngot:\n%s\nwant:\n%s", result, expected)
	}
}

// parseEntryLine regression tests for the corruption discovered by KB dogfooding
// (ADR-007 backfill): titles containing "(" or " — " misparsed because the path
// and summary were extracted with first-occurrence strings.Cut over the whole
// line. The entry contract is the renderer's: - [Title](Path)[ — Summary], so
// the "](" boundary is the LAST one and the summary separator only counts after
// the path. These tests pin that contract; the function they exercise is
// scheduled for deletion in RFC-002 Phase 3 and the tests die with it.
func TestParseEntryLine_HostileTitles(t *testing.T) {
	cases := []struct {
		name        string
		line        string
		wantTitle   string
		wantPath    string
		wantSummary string
	}{
		{
			name:        "title with parenthesized suffix",
			line:        "- [Go JSON Schema 2020-12 Validators and RFC 3339 Parsing Facts (RFC-002-A1)](kb/research/rfc-002-a1-research-web.md) — Sourced ecosystem survey",
			wantTitle:   "Go JSON Schema 2020-12 Validators and RFC 3339 Parsing Facts (RFC-002-A1)",
			wantPath:    "kb/research/rfc-002-a1-research-web.md",
			wantSummary: "Sourced ecosystem survey",
		},
		{
			name:        "title with em dash and parens",
			line:        "- [Research run log — JSON Schema/CEL bridge (json-cel)](kb/research/json-cel-log.md) — The run log and hub page",
			wantTitle:   "Research run log — JSON Schema/CEL bridge (json-cel)",
			wantPath:    "kb/research/json-cel-log.md",
			wantSummary: "The run log and hub page",
		},
		{
			name:        "title with em dash and no summary",
			line:        "- [Openness audit — write path](kb/research/openness-write-path.md)",
			wantTitle:   "Openness audit — write path",
			wantPath:    "kb/research/openness-write-path.md",
			wantSummary: "",
		},
		{
			name:        "summary containing em dash is preserved whole",
			line:        "- [Note](kb/notes/n.md) — first part — second part",
			wantTitle:   "Note",
			wantPath:    "kb/notes/n.md",
			wantSummary: "first part — second part",
		},
		{
			name:        "multiple paren groups in title",
			line:        "- [TemplateV2 to V3 Migration Recon (RFC-002-A2) (interim)](kb/research/rfc-002-a2-recon-code.md) — recon",
			wantTitle:   "TemplateV2 to V3 Migration Recon (RFC-002-A2) (interim)",
			wantPath:    "kb/research/rfc-002-a2-recon-code.md",
			wantSummary: "recon",
		},
		{
			name:        "title containing bracket pair",
			line:        "- [Draft [WIP] notes](kb/notes/d.md) — s",
			wantTitle:   "Draft [WIP] notes",
			wantPath:    "kb/notes/d.md",
			wantSummary: "s",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := parseEntryLine(tc.line, "research")
			if err != nil {
				t.Fatalf("parseEntryLine(%q) returned error: %v", tc.line, err)
			}
			if entry.Title != tc.wantTitle {
				t.Errorf("Title = %q, want %q", entry.Title, tc.wantTitle)
			}
			if entry.Path != tc.wantPath {
				t.Errorf("Path = %q, want %q", entry.Path, tc.wantPath)
			}
			if entry.Summary != tc.wantSummary {
				t.Errorf("Summary = %q, want %q", entry.Summary, tc.wantSummary)
			}
		})
	}
}

func TestRenderParseRoundTrip_HostileTitles(t *testing.T) {
	entries := []IndexEntry{
		{Path: "kb/research/rfc-002-a1-research-web.md", Title: "Go JSON Schema 2020-12 Validators and RFC 3339 Parsing Facts (RFC-002-A1)", Summary: "Sourced ecosystem survey", Type: "research"},
		{Path: "kb/research/json-cel-log.md", Title: "Research run log — JSON Schema/CEL bridge (json-cel)", Summary: "The run log and hub page", Type: "research"},
		{Path: "kb/research/openness-write-path.md", Title: "Openness audit — write path", Summary: "", Type: "research"},
	}

	parsed, err := parseIndex(RenderIndex(entries))
	if err != nil {
		t.Fatalf("parseIndex(RenderIndex(entries)) failed: %v", err)
	}
	if len(parsed) != len(entries) {
		t.Fatalf("round trip lost entries: got %d, want %d", len(parsed), len(entries))
	}
	for i := range entries {
		if parsed[i] != entries[i] {
			t.Errorf("entry %d = %+v, want %+v", i, parsed[i], entries[i])
		}
	}
}

// Regression for the compounding corruption: with the hostile entry already in
// the index, a subsequent AddEntry read-modify-write must leave it intact.
func TestAddEntry_DoesNotCorruptHostileTitles(t *testing.T) {
	kbRoot := setupKB(t)
	hostile := IndexEntry{
		Path:    "kb/research/rfc-002-a1-research-web.md",
		Title:   "Go JSON Schema 2020-12 Validators (RFC-002-A1)",
		Summary: "Sourced ecosystem survey",
		Type:    "research",
	}
	if err := AddEntry(kbRoot, hostile); err != nil {
		t.Fatalf("AddEntry(hostile) failed: %v", err)
	}
	other := IndexEntry{Path: "kb/research/plain.md", Title: "Plain", Summary: "plain summary", Type: "research"}
	if err := AddEntry(kbRoot, other); err != nil {
		t.Fatalf("AddEntry(other) failed: %v", err)
	}

	entries, err := ReadIndex(kbRoot)
	if err != nil {
		t.Fatalf("ReadIndex failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0] != hostile {
		t.Errorf("first entry corrupted by second AddEntry: got %+v, want %+v", entries[0], hostile)
	}
}
