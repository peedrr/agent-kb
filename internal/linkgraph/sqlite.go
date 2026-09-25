// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package linkgraph

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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

// SQLiteLinkGraph implements Updater using SQLite.
type SQLiteLinkGraph struct {
	db *sql.DB
}

var _ TxUpdater = (*SQLiteLinkGraph)(nil)

// NewSQLiteLinkGraph creates a new SQLiteLinkGraph backed by the given database.
func NewSQLiteLinkGraph(db *sql.DB) *SQLiteLinkGraph {
	return &SQLiteLinkGraph{db: db}
}

// dedupeWikilinksByTarget keeps the first occurrence of each wikilink target.
// The links table is keyed by (source_page, raw_target), so repeated targets in
// one page body would otherwise violate the primary key.
func dedupeWikilinksByTarget(wikilinks []markdown.Wikilink) []markdown.Wikilink {
	seen := make(map[string]bool, len(wikilinks))
	deduped := make([]markdown.Wikilink, 0, len(wikilinks))
	for _, wl := range wikilinks {
		if seen[wl.Target] {
			continue
		}
		seen[wl.Target] = true
		deduped = append(deduped, wl)
	}
	return deduped
}

// UpdatePageLinks implements Updater.UpdatePageLinks.
func (g *SQLiteLinkGraph) UpdatePageLinks(ctx context.Context, path string, content string) error {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	if err := g.UpdatePageLinksTx(ctx, tx, path, content); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// UpdatePageLinksTx replaces the outgoing links of one page using tx, a
// transaction on the link graph's database. The caller decides whether the
// write commits, so the link-graph step can be grouped with other index steps in
// one transaction.
func (g *SQLiteLinkGraph) UpdatePageLinksTx(ctx context.Context, tx *sql.Tx, path string, content string) error {
	wikilinks := dedupeWikilinksByTarget(markdown.ParseWikilinks(content))

	if _, err := tx.ExecContext(ctx,
		"INSERT OR IGNORE INTO pages (path, title, summary) VALUES (?, '', '')",
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

	return nil
}

// RemovePage implements Updater.RemovePage.
func (g *SQLiteLinkGraph) RemovePage(ctx context.Context, path string) error {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	if err := g.RemovePageTx(ctx, tx, path); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// RemovePageTx deletes the page row and every link that starts at or resolves to
// path using tx, a transaction on the link graph's database. The caller decides
// whether the deletion commits, so the removal can be grouped with other index
// steps in one transaction.
func (g *SQLiteLinkGraph) RemovePageTx(ctx context.Context, tx *sql.Tx, path string) error {
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

	return nil
}

// RebuildLinks rebuilds the entire link graph from the filesystem.
// It deletes all existing links and re-resolves wikilinks for every page.
// The delete and re-resolution steps run in one transaction, so a walk failure
// leaves the previously resolved links untouched.
func (g *SQLiteLinkGraph) RebuildLinks(ctx context.Context, kbRoot string) error {
	kbDir := filepath.Join(kbRoot, "kb")

	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	if _, err := tx.ExecContext(ctx, "DELETE FROM links"); err != nil {
		return fmt.Errorf("delete links: %w", err)
	}

	err = filepath.WalkDir(kbDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		base := filepath.Base(path)
		if base == "index.md" || base == "log.md" {
			return nil
		}

		content, err := os.ReadFile(path) //nolint:gosec // path validated by filepath.WalkDir within KB root
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(kbRoot, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		if err := g.UpdatePageLinksTx(ctx, tx, relPath, string(content)); err != nil {
			return fmt.Errorf("update links for %s: %w", relPath, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk kb directory: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
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
	defer rows.Close() //nolint:errcheck // rows.Err() checked after iteration; close error non-critical
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
	defer rows.Close() //nolint:errcheck // rows.Err() checked after iteration; close error non-critical
	return scanLinks(rows)
}

// GetOrphans returns pages that have no inbound links from other pages. A page
// linking to itself does not count as an inbound link from another page, even
// when that self-link resolved uniquely at write time.
//
// Resolution runs at write time, so resolved_to can go stale: a link that
// resolved uniquely becomes ambiguous once a same-basename page is created
// later. The stale target keeps counting as having an inbound link from the
// linking page until that page is written again or the graph is re-resolved with
// `akb index rebuild`.
func (g *SQLiteLinkGraph) GetOrphans(ctx context.Context) ([]string, error) {
	rows, err := g.db.QueryContext(ctx,
		`SELECT path FROM pages WHERE path NOT IN (
			SELECT resolved_to FROM links
			WHERE resolved_to IS NOT NULL AND resolved_to NOT LIKE 'AMBIGUOUS%'
			  AND source_page != resolved_to
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("query orphans: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows.Err() checked after iteration; close error non-critical

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan path: %w", err)
		}
		paths = append(paths, p)
	}
	if err := rows.Err(); err != nil {
		return paths, fmt.Errorf("iterate rows: %w", err)
	}
	return paths, nil
}

// GetBrokenLinks returns links whose target could not be resolved.
func (g *SQLiteLinkGraph) GetBrokenLinks(ctx context.Context) ([]Link, error) {
	rows, err := g.db.QueryContext(ctx,
		"SELECT source_page, raw_target, display, resolved_to FROM links WHERE resolved_to IS NULL",
	)
	if err != nil {
		return nil, fmt.Errorf("query broken links: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows.Err() checked after iteration; close error non-critical
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
	defer rows.Close() //nolint:errcheck // rows.Err() checked after iteration; close error non-critical
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
	defer rows.Close() //nolint:errcheck // rows.Err() checked after iteration; close error non-critical

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
	if err := rows.Err(); err != nil {
		return links, fmt.Errorf("iterate rows: %w", err)
	}
	return links, nil
}
