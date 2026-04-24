package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB(t *testing.T) {
	t.Run("creates database file", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Fatal("database file was not created")
		}
	})

	t.Run("opens existing database", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db1, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("first InitDB failed: %v", err)
		}
		_ = db1.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		db2, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("second InitDB failed: %v", err)
		}
		defer db2.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	})

	t.Run("rejects invalid path", func(t *testing.T) {
		_, err := InitDB("/nonexistent/dir/that/does/not/exist/test.db")
		if err == nil {
			t.Fatal("expected error for invalid path")
		}
	})
}

func TestCreateSchema(t *testing.T) {
	t.Run("creates all tables and indexes", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		if err := CreateSchema(db); err != nil {
			t.Fatalf("CreateSchema failed: %v", err)
		}

		tables := []string{"documents", "pages_fts", "pages", "links"}
		for _, table := range tables {
			var name string
			err := db.QueryRow(
				"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
				table,
			).Scan(&name)
			if err != nil {
				t.Errorf("table %q not found: %v", table, err)
			}
		}

		indexes := []string{"idx_links_source", "idx_links_resolved"}
		for _, idx := range indexes {
			var name string
			err := db.QueryRow(
				"SELECT name FROM sqlite_master WHERE type='index' AND name=?",
				idx,
			).Scan(&name)
			if err != nil {
				t.Errorf("index %q not found: %v", idx, err)
			}
		}
	})

	t.Run("idempotent on repeated calls", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		if err := CreateSchema(db); err != nil {
			t.Fatalf("first CreateSchema failed: %v", err)
		}
		if err := CreateSchema(db); err != nil {
			t.Fatalf("second CreateSchema failed: %v", err)
		}
	})

	t.Run("rejects nil db", func(t *testing.T) {
		err := CreateSchema(nil)
		if err == nil {
			t.Fatal("expected error for nil db")
		}
	})
}

func TestVerifySchema(t *testing.T) {
	t.Run("passes when schema is complete", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		if err := CreateSchema(db); err != nil {
			t.Fatalf("CreateSchema failed: %v", err)
		}

		if err := VerifySchema(db); err != nil {
			t.Fatalf("VerifySchema failed: %v", err)
		}
	})

	t.Run("fails when tables are missing", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		err = VerifySchema(db)
		if err == nil {
			t.Fatal("expected error for empty database")
		}
	})

	t.Run("fails when indexes are missing", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

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
		}
		for _, stmt := range ddl {
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("exec DDL: %v", err)
			}
		}

		err = VerifySchema(db)
		if err == nil {
			t.Fatal("expected error for missing indexes")
		}
	})
}

func TestOpen(t *testing.T) {
	t.Run("opens database and pings", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := Open(dbPath)
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		if err := db.Ping(context.Background()); err != nil {
			t.Fatalf("Ping failed: %v", err)
		}
	})
}

func TestOpenKB(t *testing.T) {
	t.Run("opens database with valid schema", func(t *testing.T) {
		dir := t.TempDir()
		akbDir := filepath.Join(dir, ".akb")
		if err := os.MkdirAll(akbDir, 0750); err != nil {
			t.Fatalf("create .akb dir: %v", err)
		}
		dbPath := filepath.Join(akbDir, "search.db")

		db, err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		if err := CreateSchema(db); err != nil {
			t.Fatalf("CreateSchema failed: %v", err)
		}
		_ = db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		opened, err := OpenKB(dir)
		if err != nil {
			t.Fatalf("OpenKB failed: %v", err)
		}
		defer opened.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		if opened.Stats().MaxOpenConnections != 1 {
			t.Errorf("expected MaxOpenConnections=1, got %d", opened.Stats().MaxOpenConnections)
		}
	})
}
