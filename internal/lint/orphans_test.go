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

func TestOrphansChecker_New(t *testing.T) {
	c := NewOrphansChecker()
	if c == nil {
		t.Fatal("NewOrphansChecker returned nil")
	}
}

func TestOrphansChecker_Name(t *testing.T) {
	c := NewOrphansChecker()
	if c.Name() != "orphans" {
		t.Errorf("Name() = %q, want %q", c.Name(), "orphans")
	}
}

func TestOrphansChecker_DetectsOrphan(t *testing.T) {
	d := setupDBOrphans(t)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	_, err := d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "orphan.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}

	c := NewOrphansChecker()
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
	if issues[0].Path != "orphan.md" {
		t.Errorf("Path = %q, want %q", issues[0].Path, "orphan.md")
	}
}

func TestOrphansChecker_TwoOrphans(t *testing.T) {
	d := setupDBOrphans(t)
	g := linkgraph.NewSQLiteLinkGraph(d)
	ctx := context.Background()

	_, err := d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "orphan-a.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}
	_, err = d.ExecContext(ctx, "INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')", "orphan-b.md")
	if err != nil {
		t.Fatalf("insert page: %v", err)
	}

	c := NewOrphansChecker()
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
	paths := map[string]bool{}
	for _, i := range issues {
		paths[i.Path] = true
	}
	if !paths["orphan-a.md"] || !paths["orphan-b.md"] {
		t.Errorf("expected both orphans, got %v", paths)
	}
}

func TestOrphansChecker_ReciprocalLinks(t *testing.T) {
	d := setupDBOrphans(t)
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
	err = g.UpdatePageLinks(ctx, "page-b.md", "See [[page-a]].")
	if err != nil {
		t.Fatalf("UpdatePageLinks: %v", err)
	}

	c := NewOrphansChecker()
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

func setupDBOrphans(t *testing.T) *sql.DB {
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
