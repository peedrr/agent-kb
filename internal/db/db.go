package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB for convenience.
type DB struct {
	*sql.DB
}

// Open opens a SQLite database at the given path and returns a wrapped DB.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// Ping verifies the database connection is alive.
func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

// InitDB opens or creates a SQLite database at dbPath and verifies connectivity.
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

// CreateSchema executes all DDL statements to create the database schema.
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
