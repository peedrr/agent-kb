package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func VerifySchema(db *sql.DB) error {
	expected := map[string]bool{
		"documents":          false,
		"pages_fts":          false,
		"pages":              false,
		"links":              false,
		"idx_links_source":   false,
		"idx_links_resolved": false,
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type IN ('table', 'index')")
	if err != nil {
		return fmt.Errorf("query sqlite_master: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan name: %w", err)
		}
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate rows: %w", err)
	}

	var missing []string
	for name, found := range expected {
		if !found {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing schema objects: %v", missing)
	}
	return nil
}

func CreateSchema(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	ddl := []string{
		`CREATE TABLE IF NOT EXISTS documents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL DEFAULT '',
			content TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			path TEXT NOT NULL UNIQUE,
			summary TEXT NOT NULL DEFAULT '',
			created TEXT NOT NULL DEFAULT '',
			updated TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS pages_fts USING fts5(
			title, content, tags,
			content=documents,
			content_rowid=id
		)`,
		`CREATE TABLE IF NOT EXISTS pages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL DEFAULT '',
			summary TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS links (
			source_page TEXT NOT NULL,
			raw_target TEXT NOT NULL,
			display TEXT NOT NULL DEFAULT '',
			resolved_to TEXT,
			PRIMARY KEY (source_page, raw_target)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_links_source ON links(source_page)`,
		`CREATE INDEX IF NOT EXISTS idx_links_resolved ON links(resolved_to)`,
	}

	for _, stmt := range ddl {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec DDL: %w", err)
		}
	}
	return nil
}

func OpenKB(kbRoot string) (*sql.DB, error) {
	dbPath := filepath.Join(kbRoot, ".akb", "search.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := VerifySchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("invalid schema: %w (run 'akb init' to fix)", err)
	}

	return db, nil
}
