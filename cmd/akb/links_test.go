package main

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/linkgraph"
)

func setupLinksTestKB(t *testing.T) string {
	t.Helper()
	kbRoot := t.TempDir()
	setupTestKBWithGit(t, kbRoot)

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origCwd) })

	if err := os.Chdir(kbRoot); err != nil {
		t.Fatal(err)
	}

	return kbRoot
}

func setupLinkGraphDB(t *testing.T, kbRoot string) *sql.DB {
	t.Helper()
	akbDir := filepath.Join(kbRoot, ".akb")
	dbPath := filepath.Join(akbDir, "search.db")
	d, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.CreateSchema(d); err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}
	return d
}

func insertTestPage(t *testing.T, d *sql.DB, pagePath string) {
	t.Helper()
	_, err := d.Exec("INSERT OR REPLACE INTO pages (path, title, summary) VALUES (?, '', '')", pagePath)
	if err != nil {
		t.Fatalf("insert page %q: %v", pagePath, err)
	}
}

func writeTestPage(t *testing.T, kbRoot, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(kbRoot, "kb", relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func captureOutput(f func() error) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	err = f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func TestLinksShow_WithResolvedBrokenAmbiguous(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	d := setupLinkGraphDB(t, kbRoot)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertTestPage(t, d, "kb/target.md")
	insertTestPage(t, d, "kb/notes/note.md")
	insertTestPage(t, d, "kb/archive/note.md")

	writeTestPage(t, kbRoot, "index.md", "---\ntitle: Index\n---\nSee [[target]] and [[missing]] and [[note]].")

	err := g.UpdatePageLinks(ctx, "kb/index.md", "See [[target]] and [[missing]] and [[note]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	output, err := captureOutput(func() error {
		return runLinksShow(nil, []string{"index.md"})
	})
	if err != nil {
		t.Fatalf("runLinksShow: %v", err)
	}

	if !strings.Contains(output, "Outbound:") {
		t.Error("expected Outbound section")
	}
	if !strings.Contains(output, "target -> kb/target.md") {
		t.Error("expected resolved outbound link target -> kb/target.md")
	}
	if !strings.Contains(output, "Broken:") {
		t.Error("expected Broken section")
	}
	if !strings.Contains(output, "missing") {
		t.Error("expected broken link 'missing'")
	}
	if !strings.Contains(output, "Ambiguous:") {
		t.Error("expected Ambiguous section")
	}
	if !strings.Contains(output, "note -> AMBIGUOUS:") {
		t.Error("expected ambiguous link for 'note'")
	}
}

func TestLinksShow_NoLinks(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	d := setupLinkGraphDB(t, kbRoot)

	insertTestPage(t, d, "kb/empty.md")
	writeTestPage(t, kbRoot, "empty.md", "---\ntitle: Empty\n---\nNo links here.")

	output, err := captureOutput(func() error {
		return runLinksShow(nil, []string{"empty.md"})
	})
	if err != nil {
		t.Fatalf("runLinksShow: %v", err)
	}

	if !strings.Contains(output, "Outbound:") || !strings.Contains(output, "Broken:") || !strings.Contains(output, "Ambiguous:") {
		t.Error("expected all three sections even with no links")
	}
}

func TestBacklinks_WithInboundLinks(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	d := setupLinkGraphDB(t, kbRoot)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertTestPage(t, d, "kb/target.md")

	err := g.UpdatePageLinks(ctx, "kb/page1.md", "See [[target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks page1: %v", err)
	}
	err = g.UpdatePageLinks(ctx, "kb/page2.md", "Also see [[target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks page2: %v", err)
	}

	writeTestPage(t, kbRoot, "target.md", "---\ntitle: Target\n---\nContent.")

	output, err := captureOutput(func() error {
		return runBacklinks(nil, []string{"target.md"})
	})
	if err != nil {
		t.Fatalf("runBacklinks: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	sources := map[string]bool{}
	for _, line := range lines {
		if line != "" {
			sources[line] = true
		}
	}
	if !sources["kb/page1.md"] {
		t.Error("expected backlink from kb/page1.md")
	}
	if !sources["kb/page2.md"] {
		t.Error("expected backlink from kb/page2.md")
	}
}

func TestBacklinks_NoInboundLinks(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	d := setupLinkGraphDB(t, kbRoot)

	insertTestPage(t, d, "kb/lonely.md")
	writeTestPage(t, kbRoot, "lonely.md", "---\ntitle: Lonely\n---\nNo one links here.")

	output, err := captureOutput(func() error {
		return runBacklinks(nil, []string{"lonely.md"})
	})
	if err != nil {
		t.Fatalf("runBacklinks: %v", err)
	}

	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output for no backlinks, got %q", output)
	}
}

func TestOrphans_WithOrphanPages(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	d := setupLinkGraphDB(t, kbRoot)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertTestPage(t, d, "kb/linked.md")
	insertTestPage(t, d, "kb/orphan.md")

	err := g.UpdatePageLinks(ctx, "kb/index.md", "See [[linked]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	output, err := captureOutput(func() error {
		return runOrphans(nil, nil)
	})
	if err != nil {
		t.Fatalf("runOrphans: %v", err)
	}

	if !strings.Contains(output, "kb/orphan.md") {
		t.Error("expected kb/orphan.md in orphans output")
	}
	if strings.Contains(output, "kb/linked.md") {
		t.Error("kb/linked.md should not be an orphan")
	}
}

func TestOrphans_EmptyKB(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	setupLinkGraphDB(t, kbRoot)

	output, err := captureOutput(func() error {
		return runOrphans(nil, nil)
	})
	if err != nil {
		t.Fatalf("runOrphans: %v", err)
	}

	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output for empty KB, got %q", output)
	}
}

func TestLinksShow_NonexistentPage(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	setupLinkGraphDB(t, kbRoot)

	err := runLinksShow(nil, []string{"nonexistent.md"})
	if err == nil {
		t.Fatal("expected error for nonexistent page")
	}
	if !strings.Contains(err.Error(), "page not found") {
		t.Errorf("expected 'page not found' error, got %q", err.Error())
	}
}

func TestBacklinks_NonexistentPage(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	setupLinkGraphDB(t, kbRoot)

	err := runBacklinks(nil, []string{"nonexistent.md"})
	if err == nil {
		t.Fatal("expected error for nonexistent page")
	}
	if !strings.Contains(err.Error(), "page not found") {
		t.Errorf("expected 'page not found' error, got %q", err.Error())
	}
}

func TestLinksShow_MissingDB(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	os.Remove(filepath.Join(kbRoot, ".akb", "search.db"))

	err := runLinksShow(nil, []string{"some.md"})
	if err == nil {
		t.Fatal("expected error for missing DB")
	}
	if !strings.Contains(err.Error(), "akb index rebuild") {
		t.Errorf("expected 'akb index rebuild' error, got %q", err.Error())
	}
}

func TestBacklinks_MissingDB(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	os.Remove(filepath.Join(kbRoot, ".akb", "search.db"))

	err := runBacklinks(nil, []string{"some.md"})
	if err == nil {
		t.Fatal("expected error for missing DB")
	}
	if !strings.Contains(err.Error(), "akb index rebuild") {
		t.Errorf("expected 'akb index rebuild' error, got %q", err.Error())
	}
}

func TestOrphans_MissingDB(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	os.Remove(filepath.Join(kbRoot, ".akb", "search.db"))

	err := runOrphans(nil, nil)
	if err == nil {
		t.Fatal("expected error for missing DB")
	}
	if !strings.Contains(err.Error(), "akb index rebuild") {
		t.Errorf("expected 'akb index rebuild' error, got %q", err.Error())
	}
}

func TestLinksShow_KbPrefixStripped(t *testing.T) {
	kbRoot := setupLinksTestKB(t)
	d := setupLinkGraphDB(t, kbRoot)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertTestPage(t, d, "kb/target.md")
	writeTestPage(t, kbRoot, "source.md", "---\ntitle: Source\n---\nSee [[target]].")

	err := g.UpdatePageLinks(ctx, "kb/source.md", "See [[target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	output, err := captureOutput(func() error {
		return runLinksShow(nil, []string{"kb/source.md"})
	})
	if err != nil {
		t.Fatalf("runLinksShow with kb/ prefix: %v", err)
	}

	if !strings.Contains(output, "target -> kb/target.md") {
		t.Error("expected resolved link with kb/ prefix stripped from input")
	}
}
