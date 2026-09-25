// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package search

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/index"
)

// SQLiteFTS5Searcher implements the Searcher interface using SQLite FTS5.
type SQLiteFTS5Searcher struct {
	db *sql.DB
}

var _ TxSearcher = (*SQLiteFTS5Searcher)(nil)

// NewSQLiteFTS5Searcher creates a new SQLiteFTS5Searcher with the given database connection.
func NewSQLiteFTS5Searcher(db *sql.DB) *SQLiteFTS5Searcher {
	return &SQLiteFTS5Searcher{db: db}
}

// IndexPage adds or updates a page in the search index.
// It writes to both the documents and pages tables to maintain consistency.
func (s *SQLiteFTS5Searcher) IndexPage(ctx context.Context, path, title, content, tags, summary, pageType string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	if err := s.IndexPageTx(ctx, tx, path, title, content, tags, summary, pageType); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// IndexPageTx writes the documents, pages_fts, and pages rows for one page
// using tx, a transaction on the searcher's database. The caller decides
// whether the write commits, so the search-index step can be grouped with other
// index steps in one transaction.
func (s *SQLiteFTS5Searcher) IndexPageTx(ctx context.Context, tx *sql.Tx, path, title, content, tags, summary, pageType string) error {
	_, _ = tx.ExecContext(ctx, "DELETE FROM pages_fts WHERE rowid = (SELECT id FROM documents WHERE path = ?)", path)

	_, err := tx.ExecContext(ctx,
		"INSERT OR REPLACE INTO documents (path, title, content, tags, summary, created, updated, type) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'), ?)",
		path, title, content, tags, summary, pageType)
	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}

	var id int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM documents WHERE path = ?", path).Scan(&id); err != nil {
		return fmt.Errorf("get document id: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO pages_fts(rowid, title, content, tags, summary) VALUES (?, ?, ?, ?, ?)", id, title, content, tags, summary); err != nil {
		return fmt.Errorf("insert fts: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO pages (path, title, summary) VALUES (?, ?, ?)", path, title, summary); err != nil {
		return fmt.Errorf("insert page: %w", err)
	}

	return nil
}

// RemovePage removes a page from the search index.
// It deletes from both the documents and pages tables to maintain consistency.
func (s *SQLiteFTS5Searcher) RemovePage(ctx context.Context, path string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	if err := s.RemovePageTx(ctx, tx, path); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// RemovePageTx deletes the documents, pages_fts, and pages rows for one page
// using tx, a transaction on the searcher's database. The caller decides
// whether the deletion commits, so the removal can be grouped with other index
// steps in one transaction.
func (s *SQLiteFTS5Searcher) RemovePageTx(ctx context.Context, tx *sql.Tx, path string) error {
	// Delete from FTS first — requires document id which is deleted next
	if _, err := tx.ExecContext(ctx, "DELETE FROM pages_fts WHERE rowid = (SELECT id FROM documents WHERE path = ?)", path); err != nil {
		return fmt.Errorf("delete fts: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM documents WHERE path = ?", path); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM pages WHERE path = ?", path); err != nil {
		return fmt.Errorf("delete page: %w", err)
	}

	return nil
}

// fts5Operators matches FTS5 boolean operators at word boundaries.
var fts5Operators = regexp.MustCompile(`\b(?:OR|AND|NOT)\b`)

// escapeFTS5Query escapes FTS5 special characters and operators from a user
// query, then wraps each remaining whitespace-separated token in double quotes.
// Quoting keeps punctuation inside a token (the hyphen in `event-driven`, for
// example) literal instead of letting FTS5 parse it as query syntax.
func escapeFTS5Query(query string) string {
	query = fts5Operators.ReplaceAllString(query, " ")

	replacer := strings.NewReplacer(
		`"`, " ",
		`'`, " ",
		"(", " ",
		")", " ",
		"*", " ",
		":", " ",
	)

	cleaned := strings.TrimSpace(replacer.Replace(query))
	if cleaned == "" {
		return ""
	}

	tokens := strings.Fields(cleaned)
	for i, token := range tokens {
		tokens[i] = `"` + token + `"`
	}
	return strings.Join(tokens, " ")
}

// Search performs a full-text search over the index using FTS5 BM25 ranking.
func (s *SQLiteFTS5Searcher) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	escaped := escapeFTS5Query(query)
	if escaped == "" {
		return nil, errors.New("search query must not be empty")
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 10
	}

	args := []any{escaped}
	whereClause := ""

	if opts.Tag != "" {
		whereClause += " AND d.tags LIKE ?"
		args = append(args, "%"+opts.Tag+"%")
	}
	if opts.Type != "" {
		whereClause += " AND d.type = ?"
		args = append(args, opts.Type)
	}
	if opts.After != "" {
		whereClause += " AND d.created >= ?"
		args = append(args, opts.After)
	}

	args = append(args, limit)

	query = fmt.Sprintf(`
		SELECT d.path, d.title, d.summary,
		       snippet(pages_fts, -1, '→', '←', '...', 32) as snippet,
		       bm25(pages_fts, 10.0, 1.0, 5.0, 3.0) as rank
		FROM pages_fts
		JOIN documents d ON pages_fts.rowid = d.id
		WHERE pages_fts MATCH ?%s
		ORDER BY rank
		LIMIT ?`, whereClause)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close() //nolint:errcheck // rows.Err() checked after iteration; close error non-critical

	results := []SearchResult{}
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Path, &r.Title, &r.Summary, &r.Snippet, &r.Rank); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate results: %w", err)
	}
	return results, nil
}

// RebuildIndex rebuilds the entire search index from the filesystem.
// It walks the kb/ directory, parses frontmatter from each .md file,
// clears all tables, and re-inserts all pages.
// The clear and re-insert steps run in one transaction, so a walk failure
// leaves the previously indexed content untouched.
func (s *SQLiteFTS5Searcher) RebuildIndex(ctx context.Context, kbRoot string) error {
	kbDir := filepath.Join(kbRoot, "kb")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	if _, err := tx.ExecContext(ctx, "DELETE FROM pages"); err != nil {
		return fmt.Errorf("delete pages: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM documents"); err != nil {
		return fmt.Errorf("delete documents: %w", err)
	}

	// Ensure type column exists (migration for existing databases)
	var typeColCount int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM pragma_table_info('documents') WHERE name = 'type'").Scan(&typeColCount); err == nil && typeColCount == 0 {
		_, _ = tx.ExecContext(ctx, "ALTER TABLE documents ADD COLUMN type TEXT NOT NULL DEFAULT ''")
	}

	// Drop and recreate FTS5 table to handle schema migration (adding summary column)
	if _, err := tx.ExecContext(ctx, "DROP TABLE IF EXISTS pages_fts"); err != nil {
		return fmt.Errorf("drop fts5 table: %w", err)
	}

	// Recreate FTS5 table with current schema
	if _, err := tx.ExecContext(ctx, `CREATE VIRTUAL TABLE IF NOT EXISTS pages_fts USING fts5(title, content, tags, summary, content=documents, content_rowid=id)`); err != nil {
		return fmt.Errorf("create fts5 table: %w", err)
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

		content, err := os.ReadFile(path) //nolint:gosec // path from filepath.WalkDir within KB root
		if err != nil {
			return nil
		}

		fm, body, err := frontmatter.Parse(content)
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(kbRoot, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		tags := ExtractTags(fm.Fields)
		summary := ExtractSummary(fm.Fields)

		if err := s.IndexPageTx(ctx, tx, relPath, fm.Title, string(body), tags, summary, fm.Type); err != nil {
			return fmt.Errorf("index page %s: %w", relPath, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk kb directory: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	if err := index.RebuildIndex(kbRoot); err != nil {
		return fmt.Errorf("rebuild index.md: %w", err)
	}

	return nil
}

// ExtractTags extracts the tags field from frontmatter fields.
func ExtractTags(fields map[string]any) string {
	if t, ok := fields["tags"]; ok {
		switch v := t.(type) {
		case string:
			return v
		case []any:
			parts := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					parts = append(parts, s)
				}
			}
			return strings.Join(parts, " ")
		}
	}
	return ""
}

// ExtractSummary extracts the summary field from frontmatter fields.
func ExtractSummary(fields map[string]any) string {
	if s, ok := fields["summary"]; ok {
		if str, ok := s.(string); ok {
			return str
		}
	}
	return ""
}
