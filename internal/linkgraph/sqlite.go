package linkgraph

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/peedrr/agent-kb/internal/markdown"
)

// Link represents a link between pages in the knowledge base.
type Link struct {
	SourcePage string
	RawTarget  string
	Display    string
	ResolvedTo string
}

// SQLiteLinkGraph implements LinkGraphUpdater using SQLite.
type SQLiteLinkGraph struct {
	db *sql.DB
}

// NewSQLiteLinkGraph creates a new SQLiteLinkGraph backed by the given database.
func NewSQLiteLinkGraph(db *sql.DB) *SQLiteLinkGraph {
	return &SQLiteLinkGraph{db: db}
}

// UpdatePageLinks implements LinkGraphUpdater.UpdatePageLinks.
func (g *SQLiteLinkGraph) UpdatePageLinks(ctx context.Context, path string, content string) error {
	wikilinks := markdown.ParseWikilinks(content)

	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"INSERT OR REPLACE INTO pages (path, title, summary) VALUES (?, '', '')",
		path,
	); err != nil {
		return fmt.Errorf("upsert page: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM links WHERE source_page = ?",
		path,
	); err != nil {
		return fmt.Errorf("delete existing links: %w", err)
	}

	for _, wl := range wikilinks {
		if wl.Target == "" {
			continue
		}

		resolvedTo, err := resolveTarget(ctx, tx, wl.Target)
		if err != nil {
			return fmt.Errorf("resolve target %q: %w", wl.Target, err)
		}

		var resolvedToVal interface{}
		if resolvedTo != nil {
			resolvedToVal = *resolvedTo
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO links (source_page, raw_target, display, resolved_to) VALUES (?, ?, ?, ?)",
			path, wl.Target, wl.Display, resolvedToVal,
		); err != nil {
			return fmt.Errorf("insert link: %w", err)
		}
	}

	return tx.Commit()
}

// RemovePage implements LinkGraphUpdater.RemovePage.
func (g *SQLiteLinkGraph) RemovePage(ctx context.Context, path string) error {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM pages WHERE path = ?",
		path,
	); err != nil {
		return fmt.Errorf("delete page: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM links WHERE source_page = ? OR resolved_to = ?",
		path, path,
	); err != nil {
		return fmt.Errorf("delete links: %w", err)
	}

	return tx.Commit()
}

// GetOutboundLinks returns all links originating from the given page.
func (g *SQLiteLinkGraph) GetOutboundLinks(ctx context.Context, path string) ([]Link, error) {
	rows, err := g.db.QueryContext(ctx,
		"SELECT source_page, raw_target, display, resolved_to FROM links WHERE source_page = ?",
		path,
	)
	if err != nil {
		return nil, fmt.Errorf("query outbound links: %w", err)
	}
	defer rows.Close()
	return scanLinks(rows)
}

// GetInboundLinks returns all links pointing to the given page.
func (g *SQLiteLinkGraph) GetInboundLinks(ctx context.Context, path string) ([]Link, error) {
	rows, err := g.db.QueryContext(ctx,
		"SELECT source_page, raw_target, display, resolved_to FROM links WHERE resolved_to = ?",
		path,
	)
	if err != nil {
		return nil, fmt.Errorf("query inbound links: %w", err)
	}
	defer rows.Close()
	return scanLinks(rows)
}

// GetOrphans returns pages that have no inbound links from other pages.
func (g *SQLiteLinkGraph) GetOrphans(ctx context.Context) ([]string, error) {
	rows, err := g.db.QueryContext(ctx,
		`SELECT path FROM pages WHERE path NOT IN (
			SELECT resolved_to FROM links
			WHERE resolved_to IS NOT NULL AND resolved_to NOT LIKE 'AMBIGUOUS%'
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("query orphans: %w", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan path: %w", err)
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

// GetBrokenLinks returns links whose target could not be resolved.
func (g *SQLiteLinkGraph) GetBrokenLinks(ctx context.Context) ([]Link, error) {
	rows, err := g.db.QueryContext(ctx,
		"SELECT source_page, raw_target, display, resolved_to FROM links WHERE resolved_to IS NULL",
	)
	if err != nil {
		return nil, fmt.Errorf("query broken links: %w", err)
	}
	defer rows.Close()
	return scanLinks(rows)
}

// GetAmbiguousLinks returns links whose target resolved to multiple pages.
func (g *SQLiteLinkGraph) GetAmbiguousLinks(ctx context.Context) ([]Link, error) {
	rows, err := g.db.QueryContext(ctx,
		"SELECT source_page, raw_target, display, resolved_to FROM links WHERE resolved_to LIKE 'AMBIGUOUS%'",
	)
	if err != nil {
		return nil, fmt.Errorf("query ambiguous links: %w", err)
	}
	defer rows.Close()
	return scanLinks(rows)
}

// resolveTarget implements the 3-step link resolution algorithm:
//  1. Exact match: raw_target + ".md" in pages table
//  2. Namespace prefix: "kb/" + raw_target + ".md" in pages table
//  3. Short name (basename): pages whose filename (without .md) equals raw_target
//     - 0 matches → nil (broken)
//     - 1 match → resolved path
//     - 2+ matches → "AMBIGUOUS:path1,path2" (sorted for determinism)
func resolveTarget(ctx context.Context, tx *sql.Tx, rawTarget string) (*string, error) {
	// Step 1: Exact match.
	exactPath := rawTarget + ".md"
	var found string
	err := tx.QueryRowContext(ctx,
		"SELECT path FROM pages WHERE path = ?",
		exactPath,
	).Scan(&found)
	if err == nil {
		return &found, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("exact match query: %w", err)
	}

	// Step 2: Namespace prefix.
	nsPath := "kb/" + rawTarget + ".md"
	err = tx.QueryRowContext(ctx,
		"SELECT path FROM pages WHERE path = ?",
		nsPath,
	).Scan(&found)
	if err == nil {
		return &found, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("namespace match query: %w", err)
	}

	// Step 3: Short name (basename) match.
	rows, err := tx.QueryContext(ctx, "SELECT path FROM pages")
	if err != nil {
		return nil, fmt.Errorf("basename match query: %w", err)
	}
	defer rows.Close()

	var matches []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan path: %w", err)
		}
		base := strings.TrimSuffix(filepath.Base(p), ".md")
		if base == rawTarget {
			matches = append(matches, p)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		return &matches[0], nil
	default:
		sort.Strings(matches)
		ambiguous := "AMBIGUOUS:" + strings.Join(matches, ",")
		return &ambiguous, nil
	}
}

func scanLinks(rows *sql.Rows) ([]Link, error) {
	var links []Link
	for rows.Next() {
		var l Link
		var resolvedTo sql.NullString
		if err := rows.Scan(&l.SourcePage, &l.RawTarget, &l.Display, &resolvedTo); err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		if resolvedTo.Valid {
			l.ResolvedTo = resolvedTo.String
		}
		links = append(links, l)
	}
	return links, rows.Err()
}
