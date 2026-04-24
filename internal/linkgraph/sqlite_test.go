package linkgraph

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/peedrr/agent-kb/internal/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	d, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() }) //nolint:errcheck // test cleanup — failure is non-fatal
	if err := db.CreateSchema(d); err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}
	return d
}

func insertPage(t *testing.T, d *sql.DB, path string) {
	t.Helper()
	_, err := d.Exec("INSERT OR REPLACE INTO pages (path, title, summary) VALUES (?, '', '')", path)
	if err != nil {
		t.Fatalf("insert page %q: %v", path, err)
	}
}

func TestSQLiteLinkGraph_UpdatePageLinks_ExactMatch(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "my-note.md")

	err := g.UpdatePageLinks(ctx, "index.md", "See [[my-note]] for details.")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "index.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1", len(links))
	}
	l := links[0]
	if l.SourcePage != "index.md" {
		t.Errorf("SourcePage = %q, want %q", l.SourcePage, "index.md")
	}
	if l.RawTarget != "my-note" {
		t.Errorf("RawTarget = %q, want %q", l.RawTarget, "my-note")
	}
	if l.Display != "my-note" {
		t.Errorf("Display = %q, want %q", l.Display, "my-note")
	}
	if l.ResolvedTo != "my-note.md" {
		t.Errorf("ResolvedTo = %q, want %q", l.ResolvedTo, "my-note.md")
	}
}

func TestSQLiteLinkGraph_UpdatePageLinks_NamespacePrefix(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "kb/my-note.md")

	err := g.UpdatePageLinks(ctx, "index.md", "See [[my-note]] for details.")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "index.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1", len(links))
	}
	if links[0].ResolvedTo != "kb/my-note.md" {
		t.Errorf("ResolvedTo = %q, want %q", links[0].ResolvedTo, "kb/my-note.md")
	}
}

func TestSQLiteLinkGraph_UpdatePageLinks_AmbiguousMatch(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "notes/my-note.md")
	insertPage(t, d, "archive/my-note.md")

	err := g.UpdatePageLinks(ctx, "index.md", "See [[my-note]] for details.")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "index.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1", len(links))
	}
	want := "AMBIGUOUS:archive/my-note.md,notes/my-note.md"
	if links[0].ResolvedTo != want {
		t.Errorf("ResolvedTo = %q, want %q", links[0].ResolvedTo, want)
	}
}

func TestSQLiteLinkGraph_UpdatePageLinks_BrokenLink(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	err := g.UpdatePageLinks(ctx, "index.md", "See [[nonexistent]] for details.")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "index.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1", len(links))
	}
	if links[0].ResolvedTo != "" {
		t.Errorf("ResolvedTo = %q, want empty string (NULL)", links[0].ResolvedTo)
	}
	if links[0].RawTarget != "nonexistent" {
		t.Errorf("RawTarget = %q, want %q", links[0].RawTarget, "nonexistent")
	}
}

func TestSQLiteLinkGraph_UpdatePageLinks_UpsertsSourcePage(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	err := g.UpdatePageLinks(ctx, "new-page.md", "Content with [[link]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	var path string
	err = d.QueryRow("SELECT path FROM pages WHERE path = ?", "new-page.md").Scan(&path)
	if err != nil {
		t.Fatalf("page not found in pages table: %v", err)
	}
	if path != "new-page.md" {
		t.Errorf("path = %q, want %q", path, "new-page.md")
	}
}

func TestSQLiteLinkGraph_UpdatePageLinks_ReplacesExistingLinks(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "a.md")
	insertPage(t, d, "b.md")

	err := g.UpdatePageLinks(ctx, "index.md", "See [[a]] and [[b]].")
	if err != nil {
		t.Fatalf("first UpdatePageLinks: %v", err)
	}

	links, _ := g.GetOutboundLinks(ctx, "index.md")
	if len(links) != 2 {
		t.Fatalf("after first update: len(links) = %d, want 2", len(links))
	}

	err = g.UpdatePageLinks(ctx, "index.md", "Only [[a]] now.")
	if err != nil {
		t.Fatalf("second UpdatePageLinks: %v", err)
	}

	links, err = g.GetOutboundLinks(ctx, "index.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("after second update: len(links) = %d, want 1", len(links))
	}
	if links[0].RawTarget != "a" {
		t.Errorf("RawTarget = %q, want %q", links[0].RawTarget, "a")
	}
}

func TestSQLiteLinkGraph_UpdatePageLinks_SkipsHeadingOnlyLinks(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	err := g.UpdatePageLinks(ctx, "index.md", "See [[#Introduction]] for details.")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "index.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("len(links) = %d, want 0 (heading-only links skipped)", len(links))
	}
}

func TestSQLiteLinkGraph_RemovePage(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "target.md")

	err := g.UpdatePageLinks(ctx, "source.md", "See [[target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	err = g.RemovePage(ctx, "target.md")
	if err != nil {
		t.Fatalf("RemovePage: %v", err)
	}

	var count int
	err = d.QueryRow("SELECT COUNT(*) FROM pages WHERE path = ?", "target.md").Scan(&count)
	if err != nil {
		t.Fatalf("count pages: %v", err)
	}
	if count != 0 {
		t.Errorf("pages count = %d, want 0", count)
	}

	links, err := g.GetInboundLinks(ctx, "target.md")
	if err != nil {
		t.Fatalf("GetInboundLinks: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("inbound links count = %d, want 0", len(links))
	}
}

func TestSQLiteLinkGraph_RemovePage_DeletesOutboundLinks(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "target.md")

	err := g.UpdatePageLinks(ctx, "source.md", "See [[target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	err = g.RemovePage(ctx, "source.md")
	if err != nil {
		t.Fatalf("RemovePage: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "source.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("outbound links count = %d, want 0", len(links))
	}
}

func TestSQLiteLinkGraph_GetOutboundLinks(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "a.md")
	insertPage(t, d, "b.md")

	err := g.UpdatePageLinks(ctx, "index.md", "See [[a|Page A]] and [[b]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "index.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("len(links) = %d, want 2", len(links))
	}

	foundA, foundB := false, false
	for _, l := range links {
		if l.RawTarget == "a" {
			foundA = true
			if l.Display != "Page A" {
				t.Errorf("display for 'a' = %q, want %q", l.Display, "Page A")
			}
			if l.ResolvedTo != "a.md" {
				t.Errorf("resolved_to for 'a' = %q, want %q", l.ResolvedTo, "a.md")
			}
		}
		if l.RawTarget == "b" {
			foundB = true
			if l.ResolvedTo != "b.md" {
				t.Errorf("resolved_to for 'b' = %q, want %q", l.ResolvedTo, "b.md")
			}
		}
	}
	if !foundA {
		t.Error("link to 'a' not found")
	}
	if !foundB {
		t.Error("link to 'b' not found")
	}
}

func TestSQLiteLinkGraph_GetInboundLinks(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "target.md")

	err := g.UpdatePageLinks(ctx, "page1.md", "See [[target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks page1: %v", err)
	}
	err = g.UpdatePageLinks(ctx, "page2.md", "Also see [[target|The Target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks page2: %v", err)
	}

	links, err := g.GetInboundLinks(ctx, "target.md")
	if err != nil {
		t.Fatalf("GetInboundLinks: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("len(links) = %d, want 2", len(links))
	}

	sources := map[string]bool{}
	for _, l := range links {
		sources[l.SourcePage] = true
		if l.ResolvedTo != "target.md" {
			t.Errorf("ResolvedTo = %q, want %q", l.ResolvedTo, "target.md")
		}
	}
	if !sources["page1.md"] {
		t.Error("inbound from page1.md not found")
	}
	if !sources["page2.md"] {
		t.Error("inbound from page2.md not found")
	}
}

func TestSQLiteLinkGraph_GetOrphans(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "linked.md")
	insertPage(t, d, "orphan.md")

	err := g.UpdatePageLinks(ctx, "index.md", "See [[linked]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	orphans, err := g.GetOrphans(ctx)
	if err != nil {
		t.Fatalf("GetOrphans: %v", err)
	}

	orphansMap := map[string]bool{}
	for _, p := range orphans {
		orphansMap[p] = true
	}
	if !orphansMap["orphan.md"] {
		t.Error("orphan.md should be an orphan")
	}
	if orphansMap["linked.md"] {
		t.Error("linked.md should not be an orphan")
	}
}

func TestSQLiteLinkGraph_GetBrokenLinks(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	err := g.UpdatePageLinks(ctx, "index.md", "See [[missing1]] and [[missing2]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	broken, err := g.GetBrokenLinks(ctx)
	if err != nil {
		t.Fatalf("GetBrokenLinks: %v", err)
	}
	if len(broken) != 2 {
		t.Fatalf("len(broken) = %d, want 2", len(broken))
	}

	targets := map[string]bool{}
	for _, l := range broken {
		targets[l.RawTarget] = true
		if l.ResolvedTo != "" {
			t.Errorf("ResolvedTo = %q, want empty string", l.ResolvedTo)
		}
	}
	if !targets["missing1"] {
		t.Error("missing1 not in broken links")
	}
	if !targets["missing2"] {
		t.Error("missing2 not in broken links")
	}
}

func TestSQLiteLinkGraph_GetAmbiguousLinks(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "notes/note.md")
	insertPage(t, d, "archive/note.md")

	err := g.UpdatePageLinks(ctx, "index.md", "See [[note]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	ambiguous, err := g.GetAmbiguousLinks(ctx)
	if err != nil {
		t.Fatalf("GetAmbiguousLinks: %v", err)
	}
	if len(ambiguous) != 1 {
		t.Fatalf("len(ambiguous) = %d, want 1", len(ambiguous))
	}
	want := "AMBIGUOUS:archive/note.md,notes/note.md"
	if ambiguous[0].ResolvedTo != want {
		t.Errorf("ResolvedTo = %q, want %q", ambiguous[0].ResolvedTo, want)
	}
}

func TestSQLiteLinkGraph_ImplementsInterface(_ *testing.T) {
	var _ Updater = (*SQLiteLinkGraph)(nil)
}
