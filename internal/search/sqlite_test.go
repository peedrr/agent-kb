package search

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/linkgraph"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	conn, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := db.CreateSchema(conn); err != nil {
		t.Fatalf("CreateSchema failed: %v", err)
	}
	return conn
}

func setupTestKB(t *testing.T, files map[string]string) string {
	t.Helper()
	kbRoot := t.TempDir()
	kbDir := filepath.Join(kbRoot, "kb")
	if err := os.MkdirAll(kbDir, 0750); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	for relPath, content := range files {
		fullPath := filepath.Join(kbRoot, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}
	return kbRoot
}

func TestSQLiteFTS5Searcher_IndexPage_InsertsIntoBothTables(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	err := s.IndexPage(ctx, "notes/test.md", "Test Title", "test content here", "tag1 tag2", "a summary", "")
	if err != nil {
		t.Fatalf("IndexPage failed: %v", err)
	}

	var docCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents WHERE path = ?", "notes/test.md").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 1 {
		t.Errorf("expected 1 document, got %d", docCount)
	}

	var pageCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM pages WHERE path = ?", "notes/test.md").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("expected 1 page, got %d", pageCount)
	}

	var title, summary string
	if err := conn.QueryRow("SELECT title, summary FROM pages WHERE path = ?", "notes/test.md").Scan(&title, &summary); err != nil {
		t.Fatalf("query page: %v", err)
	}
	if title != "Test Title" {
		t.Errorf("expected title %q, got %q", "Test Title", title)
	}
	if summary != "a summary" {
		t.Errorf("expected summary %q, got %q", "a summary", summary)
	}
}

func TestSQLiteFTS5Searcher_IndexPage_UpdatesExistingEntry(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	if err := s.IndexPage(ctx, "notes/test.md", "Original Title", "original content", "tag1", "original summary", ""); err != nil {
		t.Fatalf("first IndexPage failed: %v", err)
	}
	if err := s.IndexPage(ctx, "notes/test.md", "Updated Title", "updated content", "tag2", "updated summary", ""); err != nil {
		t.Fatalf("second IndexPage failed: %v", err)
	}

	var docCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents WHERE path = ?", "notes/test.md").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 1 {
		t.Errorf("expected 1 document after update, got %d", docCount)
	}

	var pageCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM pages WHERE path = ?", "notes/test.md").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("expected 1 page after update, got %d", pageCount)
	}

	var title, content, summary string
	if err := conn.QueryRow("SELECT title, content, summary FROM documents WHERE path = ?", "notes/test.md").Scan(&title, &content, &summary); err != nil {
		t.Fatalf("query document: %v", err)
	}
	if title != "Updated Title" {
		t.Errorf("expected title %q, got %q", "Updated Title", title)
	}
	if content != "updated content" {
		t.Errorf("expected content %q, got %q", "updated content", content)
	}
	if summary != "updated summary" {
		t.Errorf("expected summary %q, got %q", "updated summary", summary)
	}

	var pageTitle, pageSummary string
	if err := conn.QueryRow("SELECT title, summary FROM pages WHERE path = ?", "notes/test.md").Scan(&pageTitle, &pageSummary); err != nil {
		t.Fatalf("query page: %v", err)
	}
	if pageTitle != "Updated Title" {
		t.Errorf("expected page title %q, got %q", "Updated Title", pageTitle)
	}
	if pageSummary != "updated summary" {
		t.Errorf("expected page summary %q, got %q", "updated summary", pageSummary)
	}
}

func TestSQLiteFTS5Searcher_RemovePage_DeletesFromBothTables(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	if err := s.IndexPage(ctx, "notes/test.md", "Test Title", "test content", "tag1", "summary", ""); err != nil {
		t.Fatalf("IndexPage failed: %v", err)
	}

	if err := s.RemovePage(ctx, "notes/test.md"); err != nil {
		t.Fatalf("RemovePage failed: %v", err)
	}

	var docCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents WHERE path = ?", "notes/test.md").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 0 {
		t.Errorf("expected 0 documents after removal, got %d", docCount)
	}

	var pageCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM pages WHERE path = ?", "notes/test.md").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 0 {
		t.Errorf("expected 0 pages after removal, got %d", pageCount)
	}
}

func TestSQLiteFTS5Searcher_RemovePage_NonexistentPath(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	err := s.RemovePage(ctx, "nonexistent/path.md")
	if err != nil {
		t.Errorf("expected no error for removing nonexistent path, got: %v", err)
	}
}

func TestSQLiteFTS5Searcher_Search_ReturnsBM25RankedResults(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	if err := s.IndexPage(ctx, "notes/go.md", "Go Programming", "Go is a statically typed language", "programming", "about go", ""); err != nil {
		t.Fatalf("IndexPage go.md failed: %v", err)
	}
	if err := s.IndexPage(ctx, "notes/rust.md", "Rust Programming", "Rust is a systems language", "programming", "about rust", ""); err != nil {
		t.Fatalf("IndexPage rust.md failed: %v", err)
	}
	if err := s.IndexPage(ctx, "notes/python.md", "Python Programming", "Python is a dynamic language", "programming", "about python", ""); err != nil {
		t.Fatalf("IndexPage python.md failed: %v", err)
	}

	results, err := s.Search(ctx, "Go", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result, got 0")
	}

	found := false
	for _, r := range results {
		if r.Path == "notes/go.md" {
			found = true
			if r.Title != "Go Programming" {
				t.Errorf("expected title %q, got %q", "Go Programming", r.Title)
			}
			if r.Snippet == "" {
				t.Error("expected non-empty snippet")
			}
			break
		}
	}
	if !found {
		t.Error("expected to find notes/go.md in results")
	}
}

func TestSQLiteFTS5Searcher_Search_TitleRankedHigherThanContent(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	if err := s.IndexPage(ctx, "notes/alpha.md", "UniqueKeyword Guide", "some general content", "tag1", "summary1", ""); err != nil {
		t.Fatalf("IndexPage alpha failed: %v", err)
	}
	if err := s.IndexPage(ctx, "notes/beta.md", "General Title", "UniqueKeyword appears in content here", "tag2", "summary2", ""); err != nil {
		t.Fatalf("IndexPage beta failed: %v", err)
	}

	results, err := s.Search(ctx, "UniqueKeyword", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}

	if results[0].Path != "notes/alpha.md" {
		t.Errorf("expected title match (alpha.md) to rank first, got %s (rank=%.4f)", results[0].Path, results[0].Rank)
	}
}

func TestSQLiteFTS5Searcher_Search_EmptyQuery(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	_, err := s.Search(ctx, "", SearchOptions{Limit: 10})
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
	if err.Error() != "search query must not be empty" {
		t.Errorf("expected error %q, got %q", "search query must not be empty", err.Error())
	}
}

func TestSQLiteFTS5Searcher_Search_QueryWithOnlySpecialChars(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	_, err := s.Search(ctx, "***", SearchOptions{Limit: 10})
	if err == nil {
		t.Fatal("expected error for query with only special chars, got nil")
	}
	if err.Error() != "search query must not be empty" {
		t.Errorf("expected error %q, got %q", "search query must not be empty", err.Error())
	}
}

func TestSQLiteFTS5Searcher_Search_EscapesFTS5SpecialCharacters(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	if err := s.IndexPage(ctx, "notes/test.md", "Test Page", "some content about testing", "tag1", "summary", ""); err != nil {
		t.Fatalf("IndexPage failed: %v", err)
	}

	results, err := s.Search(ctx, `Test "Page"`, SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search with quotes failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for query with escaped quotes")
	}

	results, err = s.Search(ctx, "Test AND Page", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search with AND failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for query with AND operator")
	}

	results, err = s.Search(ctx, "Test OR Page", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search with OR failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for query with OR operator")
	}

	results, err = s.Search(ctx, "Test*Page", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search with asterisk failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for query with asterisk")
	}
}

func TestSQLiteFTS5Searcher_Search_DefaultLimit(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	for i := range 15 {
		path := filepath.ToSlash(filepath.Join("notes", "page.md"))
		if i > 0 {
			path = filepath.ToSlash(filepath.Join("notes", fmt.Sprintf("page%d.md", i)))
		}
		if err := s.IndexPage(ctx, path, "Test Page", "test content number "+string(rune('0'+i)), "tag", "summary", ""); err != nil {
			t.Fatalf("IndexPage %d failed: %v", i, err)
		}
	}

	results, err := s.Search(ctx, "test", SearchOptions{Limit: 0})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) > 10 {
		t.Errorf("expected at most 10 results with default limit, got %d", len(results))
	}
}

func TestSQLiteFTS5Searcher_Search_CustomLimit(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		path := filepath.ToSlash(filepath.Join("notes", "page"+string(rune('0'+i))+".md"))
		if err := s.IndexPage(ctx, path, "UniqueKeyword Page", "content "+string(rune('0'+i)), "tag", "summary", ""); err != nil {
			t.Fatalf("IndexPage %d failed: %v", i, err)
		}
	}

	results, err := s.Search(ctx, "UniqueKeyword", SearchOptions{Limit: 3})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) > 3 {
		t.Errorf("expected at most 3 results with custom limit, got %d", len(results))
	}
}

func TestSQLiteFTS5Searcher_RebuildIndex_RepoulatesBothTables(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	kbRoot := setupTestKB(t, map[string]string{
		"kb/notes/test.md":    "---\ntype: note\ntitle: Test Note\ntags: testing\nsummary: A test note\n---\nThis is the body of the test note.\n",
		"kb/agents/agent1.md": "---\ntype: agent\ntitle: Agent One\nsummary: First agent\n---\nAgent one body content.\n",
	})

	if err := s.IndexPage(ctx, "kb/old.md", "Old Page", "old content", "old", "old summary", ""); err != nil {
		t.Fatalf("IndexPage old page failed: %v", err)
	}

	if err := s.RebuildIndex(ctx, kbRoot); err != nil {
		t.Fatalf("RebuildIndex failed: %v", err)
	}

	var docCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 2 {
		t.Errorf("expected 2 documents after rebuild, got %d", docCount)
	}

	var pageCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM pages").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 2 {
		t.Errorf("expected 2 pages after rebuild, got %d", pageCount)
	}

	var oldCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents WHERE path = ?", "kb/old.md").Scan(&oldCount); err != nil {
		t.Fatalf("query old document: %v", err)
	}
	if oldCount != 0 {
		t.Errorf("expected old document to be removed after rebuild, got %d", oldCount)
	}

	results, err := s.Search(ctx, "Test Note", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search after rebuild failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected search results after rebuild")
	}
}

func TestSQLiteFTS5Searcher_RebuildIndex_HandlesEmptyKB(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	kbRoot := setupTestKB(t, nil)

	if err := s.RebuildIndex(ctx, kbRoot); err != nil {
		t.Fatalf("RebuildIndex on empty KB failed: %v", err)
	}

	var docCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 0 {
		t.Errorf("expected 0 documents after rebuild of empty KB, got %d", docCount)
	}

	var pageCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM pages").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 0 {
		t.Errorf("expected 0 pages after rebuild of empty KB, got %d", pageCount)
	}
}

func TestSQLiteFTS5Searcher_RebuildIndex_SkipsIndexAndLog(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	kbRoot := setupTestKB(t, map[string]string{
		"kb/index.md":      "---\ntype: index\ntitle: Index\n---\nIndex content.\n",
		"kb/log.md":        "---\ntype: log\ntitle: Log\n---\nLog content.\n",
		"kb/notes/real.md": "---\ntype: note\ntitle: Real Note\nsummary: A real note\n---\nReal note content.\n",
	})

	if err := s.RebuildIndex(ctx, kbRoot); err != nil {
		t.Fatalf("RebuildIndex failed: %v", err)
	}

	var docCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 1 {
		t.Errorf("expected 1 document (skipping index.md and log.md), got %d", docCount)
	}
}

func TestEscapeFTS5Query(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple word", "hello", "hello"},
		{"multiple words", "hello world", "hello world"},
		{"strip quotes", `"hello world"`, "hello world"},
		{"strip single quotes", `'hello'`, "hello"},
		{"strip AND operator", "hello AND world", "hello   world"},
		{"strip OR operator", "hello OR world", "hello   world"},
		{"strip NOT operator", "hello NOT world", "hello   world"},
		{"strip parentheses", "(hello)", "hello"},
		{"strip asterisk", "hello*", "hello"},
		{"strip colon", "title:hello", "title hello"},
		{"strip all special", `"hello" AND (world OR test)*`, "hello     world   test"},
		{"empty after strip", "***", ""},
		{"only operators", "AND OR NOT", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeFTS5Query(tt.input)
			if result != tt.expected {
				t.Errorf("escapeFTS5Query(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSQLiteFTS5Searcher_RebuildIndex_RollsBackPriorContentOnWalkFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("unreadable directories stay listable as root, so the walk cannot be failed")
	}
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	kbRoot := setupTestKB(t, map[string]string{
		"kb/notes/alpha.md": "---\ntype: note\ntitle: Alpha Title\nsummary: Alpha summary\n---\nalpha body content\n",
	})
	if err := s.RebuildIndex(ctx, kbRoot); err != nil {
		t.Fatalf("initial RebuildIndex failed: %v", err)
	}

	// beta.md sorts before the unreadable directory, so the failing walk indexes it
	// before aborting.
	if err := os.WriteFile(filepath.Join(kbRoot, "kb", "notes", "beta.md"), []byte("---\ntype: note\ntitle: Beta Title\nsummary: Beta summary\n---\nbeta body content\n"), 0600); err != nil {
		t.Fatalf("WriteFile beta failed: %v", err)
	}
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

	err := s.RebuildIndex(ctx, kbRoot)
	if err == nil {
		t.Fatal("RebuildIndex succeeded despite unreadable directory, want error")
	}
	if !strings.Contains(err.Error(), "walk kb directory") {
		t.Errorf("RebuildIndex error = %v, want walk failure", err)
	}

	// The index built before the failed walk is still intact.
	var docCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 1 {
		t.Errorf("expected 1 document after failed rebuild, got %d", docCount)
	}
	var path string
	if err := conn.QueryRow("SELECT path FROM documents").Scan(&path); err != nil {
		t.Fatalf("query document path: %v", err)
	}
	if path != "kb/notes/alpha.md" {
		t.Errorf("expected document path %q, got %q", "kb/notes/alpha.md", path)
	}

	var pageCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM pages").Scan(&pageCount); err != nil {
		t.Fatalf("query pages: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("expected 1 page after failed rebuild, got %d", pageCount)
	}

	// Rows inserted by the aborted walk are gone.
	for _, q := range []string{"documents", "pages"} {
		var betaCount int
		if err := conn.QueryRow("SELECT COUNT(*) FROM "+q+" WHERE path = ?", "kb/notes/beta.md").Scan(&betaCount); err != nil {
			t.Fatalf("query %s for beta: %v", q, err)
		}
		if betaCount != 0 {
			t.Errorf("expected beta.md to be absent from %s after failed rebuild, got %d rows", q, betaCount)
		}
	}

	// The FTS table was dropped and recreated inside the rolled back transaction,
	// so it must still serve the previously indexed content.
	results, err := s.Search(ctx, "Alpha", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search after failed rebuild: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 search result after failed rebuild, got %d", len(results))
	}
	if results[0].Path != "kb/notes/alpha.md" {
		t.Errorf("expected search result path %q, got %q", "kb/notes/alpha.md", results[0].Path)
	}

	betaResults, err := s.Search(ctx, "Beta", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search for beta after failed rebuild: %v", err)
	}
	if len(betaResults) != 0 {
		t.Errorf("expected 0 search results for beta after failed rebuild, got %d", len(betaResults))
	}
}

func TestSQLiteFTS5Searcher_Search_NoResults(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	ctx := context.Background()

	if err := s.IndexPage(ctx, "notes/test.md", "Test Title", "test content", "tag1", "summary", ""); err != nil {
		t.Fatalf("IndexPage failed: %v", err)
	}

	results, err := s.Search(ctx, "nonexistent", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent query, got %d", len(results))
	}
}

func TestSQLiteFTS5Searcher_IndexPageTx_CommitsWithLinkGraphStep(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	g := linkgraph.NewSQLiteLinkGraph(conn)
	ctx := context.Background()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	// First step: the search index records the page on the shared transaction.
	if err := s.IndexPageTx(ctx, tx, "kb/notes/alpha.md", "Alpha Title", "alpha body content", "tag1", "alpha summary", "note"); err != nil {
		t.Fatalf("IndexPageTx failed: %v", err)
	}
	// Second step: the link graph records the page's links on the same transaction.
	if err := g.UpdatePageLinksTx(ctx, tx, "kb/notes/alpha.md", "See [[beta]] for details."); err != nil {
		t.Fatalf("UpdatePageLinksTx failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	var docCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM documents WHERE path = ?", "kb/notes/alpha.md").Scan(&docCount); err != nil {
		t.Fatalf("query documents: %v", err)
	}
	if docCount != 1 {
		t.Errorf("expected 1 document after commit, got %d", docCount)
	}

	results, err := s.Search(ctx, "Alpha", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 search result after commit, got %d", len(results))
	}
	if results[0].Path != "kb/notes/alpha.md" {
		t.Errorf("search result path = %q, want %q", results[0].Path, "kb/notes/alpha.md")
	}

	links, err := g.GetOutboundLinks(ctx, "kb/notes/alpha.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 outbound link after commit, got %d", len(links))
	}
	if links[0].RawTarget != "beta" {
		t.Errorf("link raw target = %q, want %q", links[0].RawTarget, "beta")
	}
}

func TestSQLiteFTS5Searcher_IndexPageTx_RollsBackWhenLinkGraphStepFails(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	g := linkgraph.NewSQLiteLinkGraph(conn)
	ctx := context.Background()

	// Pre-command state: the page is indexed with its original content.
	if err := s.IndexPage(ctx, "kb/notes/alpha.md", "Alpha Title", "alpha original body", "tag1", "alpha summary", "note"); err != nil {
		t.Fatalf("IndexPage failed: %v", err)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	// First step: the rewritten page is visible on the shared transaction.
	if err := s.IndexPageTx(ctx, tx, "kb/notes/alpha.md", "Alpha Rewritten", "alpha rewritten body", "tag2", "alpha summary", "note"); err != nil {
		t.Fatalf("IndexPageTx failed: %v", err)
	}
	var pendingTitle string
	if err := tx.QueryRowContext(ctx, "SELECT title FROM documents WHERE path = ?", "kb/notes/alpha.md").Scan(&pendingTitle); err != nil {
		t.Fatalf("query document inside transaction: %v", err)
	}
	if pendingTitle != "Alpha Rewritten" {
		t.Errorf("pending title = %q, want %q", pendingTitle, "Alpha Rewritten")
	}

	// Second step fails: its table is dropped inside the transaction, so the
	// link-graph step errors on a real SQL statement.
	if _, err := tx.ExecContext(ctx, "DROP TABLE links"); err != nil {
		t.Fatalf("drop links table: %v", err)
	}
	err = g.UpdatePageLinksTx(ctx, tx, "kb/notes/alpha.md", "See [[alpha]] for details.")
	if err == nil {
		t.Fatal("expected the link-graph step to fail without its table, got nil error")
	}
	if !strings.Contains(err.Error(), "no such table: links") {
		t.Errorf("link-graph step error = %v, want missing links table", err)
	}

	// The caller rolls the shared transaction back.
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// The search index is back to the pre-command state.
	var title, content string
	if err := conn.QueryRow("SELECT title, content FROM documents WHERE path = ?", "kb/notes/alpha.md").Scan(&title, &content); err != nil {
		t.Fatalf("query document after rollback: %v", err)
	}
	if title != "Alpha Title" || content != "alpha original body" {
		t.Errorf("document after rollback = (%q, %q), want pre-command (%q, %q)", title, content, "Alpha Title", "alpha original body")
	}

	// The dropped table came back with the rollback, without the aborted step's rows.
	var linkCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM links").Scan(&linkCount); err != nil {
		t.Fatalf("query links after rollback: %v", err)
	}
	if linkCount != 0 {
		t.Errorf("expected 0 links after rollback, got %d", linkCount)
	}

	// The full-text content matches the rolled back row, not the pending rewrite.
	results, err := s.Search(ctx, "original", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 search result for the pre-command body, got %d", len(results))
	}
	rewritten, err := s.Search(ctx, "rewritten", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(rewritten) != 0 {
		t.Errorf("expected 0 search results for the rolled back body, got %d", len(rewritten))
	}
}

func TestSQLiteFTS5Searcher_RemovePageTx_RollsBackWhenLinkGraphStepFails(t *testing.T) {
	conn := setupTestDB(t)
	defer conn.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	s := NewSQLiteFTS5Searcher(conn)
	g := linkgraph.NewSQLiteLinkGraph(conn)
	ctx := context.Background()

	// Pre-command state: the page is in both indexes with one outgoing link.
	if err := s.IndexPage(ctx, "kb/notes/alpha.md", "Alpha Title", "alpha body content", "tag1", "alpha summary", "note"); err != nil {
		t.Fatalf("IndexPage failed: %v", err)
	}
	if err := g.UpdatePageLinks(ctx, "kb/notes/alpha.md", "See [[beta]] for details."); err != nil {
		t.Fatalf("UpdatePageLinks failed: %v", err)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	// First step: the page is gone from the search index on the shared transaction.
	if err := s.RemovePageTx(ctx, tx, "kb/notes/alpha.md"); err != nil {
		t.Fatalf("RemovePageTx failed: %v", err)
	}
	var pendingCount int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents WHERE path = ?", "kb/notes/alpha.md").Scan(&pendingCount); err != nil {
		t.Fatalf("query document inside transaction: %v", err)
	}
	if pendingCount != 0 {
		t.Errorf("expected the page to be gone inside the transaction, got %d documents", pendingCount)
	}

	// Second step fails: its table is dropped inside the transaction, so the
	// link-graph removal errors on a real SQL statement.
	if _, err := tx.ExecContext(ctx, "DROP TABLE links"); err != nil {
		t.Fatalf("drop links table: %v", err)
	}
	err = g.RemovePageTx(ctx, tx, "kb/notes/alpha.md")
	if err == nil {
		t.Fatal("expected the link-graph removal to fail without its table, got nil error")
	}
	if !strings.Contains(err.Error(), "no such table: links") {
		t.Errorf("link-graph removal error = %v, want missing links table", err)
	}

	// The caller rolls the shared transaction back.
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Both index structures are back to the pre-command state.
	results, err := s.Search(ctx, "Alpha", SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 search result after rollback, got %d", len(results))
	}
	if results[0].Path != "kb/notes/alpha.md" {
		t.Errorf("search result path = %q, want %q", results[0].Path, "kb/notes/alpha.md")
	}

	links, err := g.GetOutboundLinks(ctx, "kb/notes/alpha.md")
	if err != nil {
		t.Fatalf("GetOutboundLinks failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 outbound link after rollback, got %d", len(links))
	}
	if links[0].RawTarget != "beta" {
		t.Errorf("link raw target = %q, want %q", links[0].RawTarget, "beta")
	}

	var pageCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM pages WHERE path = ?", "kb/notes/alpha.md").Scan(&pageCount); err != nil {
		t.Fatalf("query pages after rollback: %v", err)
	}
	if pageCount != 1 {
		t.Errorf("expected 1 page row after rollback, got %d", pageCount)
	}
}
