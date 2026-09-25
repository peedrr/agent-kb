// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// The helper process spawned by TestOpenKBWaitsForWriterInAnotherProcess is
// this test binary re-executed with these environment variables set. It opens
// the search database, holds a write transaction, and reports readiness on
// stdout while the lock is still held.
const (
	lockHelperModeEnv = "AKB_DB_TEST_LOCK_HELPER"
	lockHelperRootEnv = "AKB_DB_TEST_LOCK_ROOT"
	lockHelperReady   = "locked"
	lockHelperHold    = 1200 * time.Millisecond
)

// seedSearchDB creates the search database with its full schema in kbRoot.
func seedSearchDB(t *testing.T, kbRoot string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(kbRoot, ".akb"), 0750); err != nil {
		t.Fatalf("create .akb dir: %v", err)
	}

	db, err := InitDB(filepath.Join(kbRoot, ".akb", "search.db"))
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

	if err := CreateSchema(db); err != nil {
		t.Fatalf("CreateSchema failed: %v", err)
	}
}

// TestDBLockHolderHelper is a helper process, not a test: it does nothing
// unless the calling test set the helper environment variables.
func TestDBLockHolderHelper(t *testing.T) {
	if os.Getenv(lockHelperModeEnv) == "" {
		t.Skip("helper process only")
	}

	db, err := OpenKB(os.Getenv(lockHelperRootEnv))
	if err != nil {
		t.Fatalf("OpenKB failed: %v", err)
	}
	defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	if _, err := tx.Exec("INSERT INTO documents (path, title) VALUES (?, ?)", "helper.md", "helper"); err != nil {
		t.Fatalf("insert helper document: %v", err)
	}

	if _, err := fmt.Fprintln(os.Stdout, lockHelperReady); err != nil {
		t.Fatalf("report lock: %v", err)
	}
	time.Sleep(lockHelperHold)

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit transaction: %v", err)
	}
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

func TestSQLiteDSNBusyTimeout(t *testing.T) {
	t.Run("InitDB", func(t *testing.T) {
		db, err := InitDB(filepath.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		assertBusyTimeout(t, db)
	})

	t.Run("OpenKB", func(t *testing.T) {
		kbRoot := t.TempDir()
		seedSearchDB(t, kbRoot)

		db, err := OpenKB(kbRoot)
		if err != nil {
			t.Fatalf("OpenKB failed: %v", err)
		}
		defer db.Close() //nolint:errcheck // test cleanup — failure is non-fatal

		assertBusyTimeout(t, db)
	})
}

func assertBusyTimeout(t *testing.T, db *sql.DB) {
	t.Helper()

	var timeout int
	if err := db.QueryRow("PRAGMA busy_timeout").Scan(&timeout); err != nil {
		t.Fatalf("read busy_timeout: %v", err)
	}
	if timeout != 5000 {
		t.Errorf("expected busy_timeout=5000, got %d", timeout)
	}
}

// TestOpenKBWaitsForWriterInAnotherProcess checks that a second process writing
// while another process holds the write lock waits for the lock to be released
// instead of failing immediately with SQLITE_BUSY.
func TestOpenKBWaitsForWriterInAnotherProcess(t *testing.T) {
	kbRoot := t.TempDir()
	seedSearchDB(t, kbRoot)

	parent, err := OpenKB(kbRoot)
	if err != nil {
		t.Fatalf("OpenKB failed: %v", err)
	}
	defer parent.Close() //nolint:errcheck // test cleanup — failure is non-fatal

	cmd := exec.Command(os.Args[0], "-test.run=^TestDBLockHolderHelper$") //nolint:gosec // helper process re-executing its own test binary
	cmd.Env = append(os.Environ(),
		lockHelperModeEnv+"=hold",
		lockHelperRootEnv+"="+kbRoot,
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("open helper stdout: %v", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper process: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill() //nolint:errcheck // no-op once the helper has exited
	})

	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	ready := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(stdout)
		line, readErr := reader.ReadString('\n')
		if readErr != nil && line == "" {
			ready <- ""
			return
		}
		ready <- strings.TrimSpace(line)
		_, _ = io.Copy(io.Discard, reader)
	}()

	select {
	case line := <-ready:
		if line != lockHelperReady {
			t.Fatalf("helper process did not report holding the write lock (got %q, exit %v, stderr: %s)",
				line, <-waitDone, stderr.String())
		}
	case <-time.After(60 * time.Second):
		t.Fatal("timed out waiting for the helper process to take the write lock")
	}

	start := time.Now()
	tx, err := parent.Begin()
	if err != nil {
		t.Fatalf("begin transaction while another process holds the write lock: %v", err)
	}
	if _, err := tx.Exec("INSERT INTO documents (path, title) VALUES (?, ?)", "parent.md", "parent"); err != nil {
		t.Fatalf("insert document while another process holds the write lock: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit transaction: %v", err)
	}
	elapsed := time.Since(start)

	if err := <-waitDone; err != nil {
		t.Fatalf("helper process failed: %v (stderr: %s)", err, stderr.String())
	}
	if elapsed < lockHelperHold/4 {
		t.Errorf("write succeeded after %v, expected it to wait for the helper process to release the write lock", elapsed)
	}

	var paths int
	if err := parent.QueryRow(
		"SELECT COUNT(*) FROM documents WHERE path IN ('parent.md', 'helper.md')",
	).Scan(&paths); err != nil {
		t.Fatalf("count documents: %v", err)
	}
	if paths != 2 {
		t.Errorf("expected both processes' documents to be committed, got %d", paths)
	}
}
