package linkgraph

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/search"
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

func TestSQLiteLinkGraph_UpdatePageLinks_DuplicateTargets(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	insertPage(t, d, "foo.md")

	err := g.UpdatePageLinks(ctx, "index.md", "See [[foo]] and again [[foo|Foo]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	var count int
	if err := d.QueryRow("SELECT COUNT(*) FROM links WHERE source_page = ?", "index.md").Scan(&count); err != nil {
		t.Fatalf("count links: %v", err)
	}
	if count != 1 {
		t.Fatalf("links count = %d, want 1", count)
	}

	links, err := g.GetOutboundLinks(ctx, "index.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1", len(links))
	}
	if links[0].RawTarget != "foo" {
		t.Errorf("RawTarget = %q, want %q", links[0].RawTarget, "foo")
	}
	if links[0].Display != "foo" {
		t.Errorf("Display = %q, want %q (first occurrence kept)", links[0].Display, "foo")
	}
	if links[0].ResolvedTo != "foo.md" {
		t.Errorf("ResolvedTo = %q, want %q", links[0].ResolvedTo, "foo.md")
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

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

func TestSQLiteLinkGraph_GetOrphans_SelfLinkIsNotInbound(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	if err := g.UpdatePageLinks(ctx, "notes/self-note.md", "A page that links to [[self-note]] only."); err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	// The self-link resolves to its own page while it is the only page with that
	// basename.
	links, err := g.GetOutboundLinks(ctx, "notes/self-note.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1", len(links))
	}
	if links[0].ResolvedTo != "notes/self-note.md" {
		t.Fatalf("ResolvedTo = %q, want %q", links[0].ResolvedTo, "notes/self-note.md")
	}

	orphans, err := g.GetOrphans(ctx)
	if err != nil {
		t.Fatalf("GetOrphans: %v", err)
	}
	if !containsPath(orphans, "notes/self-note.md") {
		t.Errorf("orphans = %v, want notes/self-note.md", orphans)
	}
}

func TestSQLiteLinkGraph_GetOrphans_AmbiguousBasenameCandidates(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	// alpha is written first, so its [[shared-name]] link resolves to itself. The
	// later beta page makes that basename ambiguous, but alpha's link keeps the
	// resolution it got at write time.
	if err := g.UpdatePageLinks(ctx, "notes/alpha/shared-name.md", "Alpha content with [[shared-name]]."); err != nil {
		t.Fatalf("UpdatePageLinks alpha: %v", err)
	}
	if err := g.UpdatePageLinks(ctx, "notes/beta/shared-name.md", "Beta content with [[shared-name]]."); err != nil {
		t.Fatalf("UpdatePageLinks beta: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "notes/alpha/shared-name.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1", len(links))
	}
	if links[0].ResolvedTo != "notes/alpha/shared-name.md" {
		t.Fatalf("ResolvedTo = %q, want the stale write-time resolution %q", links[0].ResolvedTo, "notes/alpha/shared-name.md")
	}

	orphans, err := g.GetOrphans(ctx)
	if err != nil {
		t.Fatalf("GetOrphans: %v", err)
	}
	for _, want := range []string{"notes/alpha/shared-name.md", "notes/beta/shared-name.md"} {
		if !containsPath(orphans, want) {
			t.Errorf("orphans = %v, want %s", orphans, want)
		}
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

func writeKBFile(t *testing.T, kbRoot, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(kbRoot, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
		t.Fatalf("MkdirAll %q: %v", filepath.Dir(fullPath), err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile %q: %v", relPath, err)
	}
}

func TestSQLiteLinkGraph_RebuildLinks_RollsBackPriorLinksOnWalkFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("unreadable directories stay listable as root, so the walk cannot be failed")
	}
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	ctx := context.Background()

	kbRoot := t.TempDir()
	// aaa-target.md sorts before mmm-source.md, so the link resolves during a rebuild
	// that starts from an empty pages table.
	writeKBFile(t, kbRoot, "kb/notes/aaa-target.md", "Target body.\n")
	writeKBFile(t, kbRoot, "kb/notes/mmm-source.md", "See [[aaa-target]] for details.\n")

	if err := g.RebuildLinks(ctx, kbRoot); err != nil {
		t.Fatalf("initial RebuildLinks: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "kb/notes/mmm-source.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d before failed rebuild, want 1", len(links))
	}
	if links[0].ResolvedTo != "kb/notes/aaa-target.md" {
		t.Fatalf("ResolvedTo = %q before failed rebuild, want %q", links[0].ResolvedTo, "kb/notes/aaa-target.md")
	}

	// nnn-gamma.md sorts before the unreadable directory, so the failing walk resolves it
	// before aborting.
	writeKBFile(t, kbRoot, "kb/notes/nnn-gamma.md", "See [[aaa-target]] again.\n")
	lockedDir := filepath.Join(kbRoot, "kb", "zzz-locked")
	if err := os.Mkdir(lockedDir, 0750); err != nil {
		t.Fatalf("Mkdir locked dir failed: %v", err)
	}
	if err := os.Chmod(lockedDir, 0000); err != nil {
		t.Fatalf("Chmod locked dir failed: %v", err)
	}
	// The directory is empty so that it can be removed without restoring its
	// permissions first.
	t.Cleanup(func() { _ = os.Remove(lockedDir) }) //nolint:errcheck // removal only needs write access to the parent directory

	err = g.RebuildLinks(ctx, kbRoot)
	if err == nil {
		t.Fatal("RebuildLinks succeeded despite unreadable directory, want error")
	}
	if !strings.Contains(err.Error(), "walk kb directory") {
		t.Errorf("RebuildLinks error = %v, want walk failure", err)
	}

	// Links resolved before the failed walk are still intact.
	links, err = g.GetOutboundLinks(ctx, "kb/notes/mmm-source.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks after failed rebuild: %v", err)
	}
	if len(links) != 1 {
		t.Errorf("len(links) = %d after failed rebuild, want 1", len(links))
	} else if links[0].ResolvedTo != "kb/notes/aaa-target.md" {
		t.Errorf("ResolvedTo = %q after failed rebuild, want %q", links[0].ResolvedTo, "kb/notes/aaa-target.md")
	}

	var linkCount int
	if err := d.QueryRow("SELECT COUNT(*) FROM links").Scan(&linkCount); err != nil {
		t.Fatalf("query links: %v", err)
	}
	if linkCount != 1 {
		t.Errorf("expected 1 link row after failed rebuild, got %d", linkCount)
	}

	// Rows written by the aborted walk are gone.
	gammaLinks, err := g.GetOutboundLinks(ctx, "kb/notes/nnn-gamma.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks for gamma: %v", err)
	}
	if len(gammaLinks) != 0 {
		t.Errorf("expected 0 links for gamma after failed rebuild, got %d", len(gammaLinks))
	}

	var strayPages int
	if err := d.QueryRow("SELECT COUNT(*) FROM pages WHERE path = ?", "kb/notes/nnn-gamma.md").Scan(&strayPages); err != nil {
		t.Fatalf("query stray pages: %v", err)
	}
	if strayPages != 0 {
		t.Errorf("expected 0 pages for gamma after failed rebuild, got %d", strayPages)
	}

	var pageCount int
	if err := d.QueryRow("SELECT COUNT(*) FROM pages").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 2 {
		t.Errorf("expected 2 page rows (aaa-target, mmm-source) after failed rebuild, got %d", pageCount)
	}
}

func TestSQLiteLinkGraph_ImplementsInterface(_ *testing.T) {
	var _ Updater = (*SQLiteLinkGraph)(nil)
}

func TestSQLiteLinkGraph_UpdatePageLinksTx_CommitsWithSearchStep(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	s := search.NewSQLiteFTS5Searcher(d)
	ctx := context.Background()

	insertPage(t, d, "kb/notes/aaa-target.md")

	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	// First step: the link graph records the page's links on the shared transaction.
	if err := g.UpdatePageLinksTx(ctx, tx, "kb/notes/mmm-source.md", "See [[aaa-target]] for details."); err != nil {
		t.Fatalf("UpdatePageLinksTx: %v", err)
	}
	// Second step: the search index records the same page on the same transaction.
	if err := s.IndexPageTx(ctx, tx, "kb/notes/mmm-source.md", "Source Title", "source body content", "tag1", "source summary", "note"); err != nil {
		t.Fatalf("IndexPageTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	links, err := g.GetOutboundLinks(ctx, "kb/notes/mmm-source.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d after commit, want 1", len(links))
	}
	if links[0].ResolvedTo != "kb/notes/aaa-target.md" {
		t.Errorf("ResolvedTo = %q after commit, want %q", links[0].ResolvedTo, "kb/notes/aaa-target.md")
	}

	results, err := s.Search(ctx, "Source", search.SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d after commit, want 1", len(results))
	}
	if results[0].Path != "kb/notes/mmm-source.md" {
		t.Errorf("search result path = %q after commit, want %q", results[0].Path, "kb/notes/mmm-source.md")
	}
}

func TestSQLiteLinkGraph_UpdatePageLinksTx_RollsBackWhenSearchStepFails(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	s := search.NewSQLiteFTS5Searcher(d)
	ctx := context.Background()

	// Pre-command state: a resolved target, a source page linking to it, and one
	// indexed page.
	insertPage(t, d, "kb/notes/aaa-target.md")
	insertPage(t, d, "kb/notes/bbb-target.md")
	if err := g.UpdatePageLinks(ctx, "kb/notes/mmm-source.md", "See [[aaa-target]] for details."); err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}
	if err := s.IndexPage(ctx, "kb/notes/zzz-report.md", "Report Title", "report body content", "tag1", "report summary", "note"); err != nil {
		t.Fatalf("IndexPage: %v", err)
	}

	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	// First step: the link graph re-resolves the source page on the shared transaction.
	if err := g.UpdatePageLinksTx(ctx, tx, "kb/notes/mmm-source.md", "See [[bbb-target]] now."); err != nil {
		t.Fatalf("UpdatePageLinksTx: %v", err)
	}
	var pendingTarget string
	if err := tx.QueryRowContext(ctx, "SELECT raw_target FROM links WHERE source_page = ?", "kb/notes/mmm-source.md").Scan(&pendingTarget); err != nil {
		t.Fatalf("query links inside transaction: %v", err)
	}
	if pendingTarget != "bbb-target" {
		t.Errorf("pending raw target = %q, want %q", pendingTarget, "bbb-target")
	}

	// Second step fails: the pages table is dropped inside the transaction, so the
	// search-index step errors on its final statement.
	if _, err := tx.ExecContext(ctx, "DROP TABLE pages"); err != nil {
		t.Fatalf("drop pages table: %v", err)
	}
	err = s.IndexPageTx(ctx, tx, "kb/notes/nnn-new.md", "New Title", "new body content", "tag1", "new summary", "note")
	if err == nil {
		t.Fatal("expected the search-index step to fail without the pages table, got nil error")
	}
	if !strings.Contains(err.Error(), "no such table: pages") {
		t.Errorf("search-index step error = %v, want missing pages table", err)
	}

	// The caller rolls the shared transaction back.
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	// The link graph is back to the pre-command state.
	links, err := g.GetOutboundLinks(ctx, "kb/notes/mmm-source.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks after rollback: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d after rollback, want 1", len(links))
	}
	if links[0].RawTarget != "aaa-target" {
		t.Errorf("RawTarget = %q after rollback, want %q", links[0].RawTarget, "aaa-target")
	}
	if links[0].ResolvedTo != "kb/notes/aaa-target.md" {
		t.Errorf("ResolvedTo = %q after rollback, want %q", links[0].ResolvedTo, "kb/notes/aaa-target.md")
	}

	var linkCount int
	if err := d.QueryRow("SELECT COUNT(*) FROM links").Scan(&linkCount); err != nil {
		t.Fatalf("query links after rollback: %v", err)
	}
	if linkCount != 1 {
		t.Errorf("links count = %d after rollback, want 1", linkCount)
	}

	// The search index is back to the pre-command state; the dropped table came back
	// with the rollback and without the aborted step's rows.
	var docCount int
	if err := d.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount); err != nil {
		t.Fatalf("query documents after rollback: %v", err)
	}
	if docCount != 1 {
		t.Errorf("documents count = %d after rollback, want 1", docCount)
	}
	var strayDocuments int
	if err := d.QueryRow("SELECT COUNT(*) FROM documents WHERE path = ?", "kb/notes/nnn-new.md").Scan(&strayDocuments); err != nil {
		t.Fatalf("query documents for the aborted page: %v", err)
	}
	if strayDocuments != 0 {
		t.Errorf("expected nnn-new.md absent from documents after rollback, got %d rows", strayDocuments)
	}

	results, err := s.Search(ctx, "Report", search.SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len(results) = %d after rollback, want 1", len(results))
	}
}

func TestSQLiteLinkGraph_RemovePageTx_RollsBackWhenSearchStepFails(t *testing.T) {
	d := setupTestDB(t)
	g := NewSQLiteLinkGraph(d)
	s := search.NewSQLiteFTS5Searcher(d)
	ctx := context.Background()

	// Pre-command state: the source page is in both indexes with one link.
	insertPage(t, d, "kb/notes/aaa-target.md")
	if err := g.UpdatePageLinks(ctx, "kb/notes/mmm-source.md", "See [[aaa-target]] for details."); err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}
	if err := s.IndexPage(ctx, "kb/notes/mmm-source.md", "Source Title", "source body content", "tag1", "source summary", "note"); err != nil {
		t.Fatalf("IndexPage: %v", err)
	}

	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	// First step: the link graph drops the page and its links on the shared transaction.
	if err := g.RemovePageTx(ctx, tx, "kb/notes/mmm-source.md"); err != nil {
		t.Fatalf("RemovePageTx: %v", err)
	}
	var pendingLinks int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM links WHERE source_page = ?", "kb/notes/mmm-source.md").Scan(&pendingLinks); err != nil {
		t.Fatalf("query links inside transaction: %v", err)
	}
	if pendingLinks != 0 {
		t.Errorf("links inside transaction = %d, want 0", pendingLinks)
	}

	// Second step fails: the pages table is dropped inside the transaction, so the
	// search-index removal errors on its final statement.
	if _, err := tx.ExecContext(ctx, "DROP TABLE pages"); err != nil {
		t.Fatalf("drop pages table: %v", err)
	}
	err = s.RemovePageTx(ctx, tx, "kb/notes/mmm-source.md")
	if err == nil {
		t.Fatal("expected the search-index removal to fail without the pages table, got nil error")
	}
	if !strings.Contains(err.Error(), "no such table: pages") {
		t.Errorf("search-index removal error = %v, want missing pages table", err)
	}

	// The caller rolls the shared transaction back.
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	// Both index structures are back to the pre-command state.
	links, err := g.GetOutboundLinks(ctx, "kb/notes/mmm-source.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks after rollback: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("len(links) = %d after rollback, want 1", len(links))
	}
	if links[0].RawTarget != "aaa-target" {
		t.Errorf("RawTarget = %q after rollback, want %q", links[0].RawTarget, "aaa-target")
	}

	var pageCount int
	if err := d.QueryRow("SELECT COUNT(*) FROM pages WHERE path = ?", "kb/notes/mmm-source.md").Scan(&pageCount); err != nil {
		t.Fatalf("query pages after rollback: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("pages count = %d after rollback, want 1", pageCount)
	}

	var title string
	if err := d.QueryRow("SELECT title FROM documents WHERE path = ?", "kb/notes/mmm-source.md").Scan(&title); err != nil {
		t.Fatalf("query documents after rollback: %v", err)
	}
	if title != "Source Title" {
		t.Errorf("document title = %q after rollback, want %q", title, "Source Title")
	}

	results, err := s.Search(ctx, "Source", search.SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d after rollback, want 1", len(results))
	}
	if results[0].Path != "kb/notes/mmm-source.md" {
		t.Errorf("search result path = %q after rollback, want %q", results[0].Path, "kb/notes/mmm-source.md")
	}
}
