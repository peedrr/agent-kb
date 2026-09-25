// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/linkgraph"
)

func TestBrokenLinksChecker_New(t *testing.T) {
	c := NewBrokenLinksChecker()
	if c == nil {
		t.Fatal("NewBrokenLinksChecker returned nil")
	}
}

func TestBrokenLinksChecker_Name(t *testing.T) {
	c := NewBrokenLinksChecker()
	if c.Name() != "broken_links" {
		t.Errorf("Name() = %q, want %q", c.Name(), "broken_links")
	}
}

func TestBrokenLinksChecker_DetectsBrokenLinks(t *testing.T) {
	d := setupDB(t)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	_, err := d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "index.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}

	err = g.UpdatePageLinks(ctx, "index.md", "See [[missing-page]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	c := NewBrokenLinksChecker()
	kb := &KB{
		Root:      t.TempDir(),
		LinkGraph: g,
	}

	issues, err := c.Check(ctx, kb)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("len(issues) = %d, want 1", len(issues))
	}
	if issues[0].Severity != "error" {
		t.Errorf("Severity = %q, want %q", issues[0].Severity, "error")
	}
	if issues[0].Path != "index.md" {
		t.Errorf("Path = %q, want %q", issues[0].Path, "index.md")
	}
}

func TestBrokenLinksChecker_DetectsAmbiguousLinks(t *testing.T) {
	d := setupDB(t)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	_, err := d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "notes/note.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}
	_, err = d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "archive/note.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}

	err = g.UpdatePageLinks(ctx, "index.md", "See [[note]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	c := NewBrokenLinksChecker()
	kb := &KB{
		Root:      t.TempDir(),
		LinkGraph: g,
	}

	issues, err := c.Check(ctx, kb)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("len(issues) = %d, want 1", len(issues))
	}
	if issues[0].Severity != "warning" {
		t.Errorf("Severity = %q, want %q", issues[0].Severity, "warning")
	}
}

func TestBrokenLinksChecker_BothBrokenAndAmbiguous(t *testing.T) {
	d := setupDB(t)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	_, err := d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "index.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}

	err = g.UpdatePageLinks(ctx, "index.md", "See [[missing]] and [[ambiguous-target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	_, err = d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "notes/ambiguous-target.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}
	_, err = d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "archive/ambiguous-target.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}

	err = g.RemovePage(ctx, "index.md")
	if err != nil {
		t.Fatalf("RemovePage: %v", err)
	}
	err = g.UpdatePageLinks(ctx, "index.md", "See [[missing]] and [[ambiguous-target]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	c := NewBrokenLinksChecker()
	kb := &KB{
		Root:      t.TempDir(),
		LinkGraph: g,
	}

	issues, err := c.Check(ctx, kb)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("len(issues) = %d, want 2", len(issues))
	}
}

func TestBrokenLinksChecker_NoIssues(t *testing.T) {
	d := setupDB(t)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	_, err := d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "page-a.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}
	_, err = d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "page-b.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}

	err = g.UpdatePageLinks(ctx, "page-a.md", "See [[page-b]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	c := NewBrokenLinksChecker()
	kb := &KB{
		Root:      t.TempDir(),
		LinkGraph: g,
	}

	issues, err := c.Check(ctx, kb)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("len(issues) = %d, want 0", len(issues))
	}
}

func setupDB(t *testing.T) *sql.DB {
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
