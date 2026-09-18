package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/index"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/peedrr/agent-kb/internal/template"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage the knowledge base index",
	Long:  `Manage the knowledge base index: show, add, remove, or rebuild entries.`,
	Example: `  # Show all index entries
  akb index show

  # Add a page to the index
  akb index add notes/my-note.md "A summary of the note"

  # Rebuild the entire index
  akb index rebuild`,
	Run: func(cmd *cobra.Command, _ []string) {
		_ = cmd.Help() //nolint:errcheck // help display failure is non-fatal
	},
}

var indexShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the index",
	Long:  `Print the contents of kb/index.md to stdout.`,
	Example: `  # Show the index
  akb index show`,
	Args: cobra.NoArgs,
	RunE: runIndexShow,
}

var indexAddCmd = &cobra.Command{
	Use:   "add <path> <summary>",
	Short: "Add or update an entry in the index",
	Long:  `Add or update an entry in kb/index.md with the given path and summary.`,
	Example: `  # Add a page to the index
  akb index add notes/my-note.md "A simple note about stuff"

  # Add with kb/ prefix (optional)
  akb index add kb/adr/use-sqlite-search.md "ADR for SQLite search"`,
	Args: cobra.ExactArgs(2),
	RunE: runIndexAdd,
}

var indexRemoveCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Remove an entry from the index",
	Long:  `Remove an entry from kb/index.md by path.`,
	Example: `  # Remove an entry from the index
  akb index remove notes/my-note.md`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexRemove,
}

var indexRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild the index from the filesystem",
	Long:  `Regenerate kb/index.md by walking the kb/ directory and parsing frontmatter from each .md file.`,
	Example: `  # Rebuild the index
  akb index rebuild`,
	Args: cobra.NoArgs,
	RunE: runIndexRebuild,
}

func init() {
	indexCmd.AddCommand(indexShowCmd)
	indexCmd.AddCommand(indexAddCmd)
	indexCmd.AddCommand(indexRemoveCmd)
	indexCmd.AddCommand(indexRebuildCmd)
}

func runIndexShow(_ *cobra.Command, _ []string) error {
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	indexPath := filepath.Join(kbRoot, "kb", "index.md")
	data, err := os.ReadFile(indexPath) //nolint:gosec // path validated by ResolveKBPath
	if err != nil {
		return fmt.Errorf("read index: %w", err)
	}

	fmt.Print(string(data))
	return nil
}

func runIndexAdd(_ *cobra.Command, args []string) error {
	entryPath := args[0]
	summary := args[1]

	// Reject index.md and log.md
	base := filepath.Base(entryPath)
	if base == "index.md" || base == "log.md" {
		return &usageError{msg: fmt.Sprintf("cannot add %s to index", base)}
	}

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	_, err = path.ResolveKBPath(kbRoot, entryPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Hold the repository lock across the read-modify-write of kb/index.md and
	// the commit that records it, so a command running concurrently against the
	// same index cannot overwrite this entry.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	// Resolve the full file path
	cleanPath := strings.TrimPrefix(entryPath, "kb/")
	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	// Read the file to get frontmatter for title and type
	content, err := os.ReadFile(fullPath) //nolint:gosec // path validated by ResolveKBPath
	if err != nil {
		// Direct path not found — try template-aware resolution
		templates, tmplErr := template.LoadTemplates(filepath.Join(kbRoot, ".akb", "templates"))
		if tmplErr != nil {
			return fmt.Errorf("read file %s: %w", entryPath, err)
		}

		for _, tmpl := range templates {
			if tmpl.Dir == "" {
				continue
			}
			candidatePath := filepath.Join(kbRoot, "kb", tmpl.Dir, cleanPath)
			candidateContent, readErr := os.ReadFile(candidatePath) //nolint:gosec
			if readErr != nil {
				continue
			}
			fm, _, parseErr := frontmatter.Parse(candidateContent)
			if parseErr != nil {
				continue
			}
			if fm.Type == tmpl.Name {
				content = candidateContent
				break
			}
		}

		if content == nil {
			return fmt.Errorf("read file %s: %w", entryPath, err)
		}
	}

	fm, _, err := frontmatter.Parse(content)
	if err != nil {
		return fmt.Errorf("parse frontmatter: %w", err)
	}

	// Use kb/ prefixed path for the entry
	relPath := filepath.Join("kb", cleanPath)
	relPath = filepath.ToSlash(relPath)

	entry := index.IndexEntry{
		Path:    relPath,
		Title:   fm.Title,
		Summary: summary,
		Type:    fm.Type,
	}

	if err := index.AddEntry(kbRoot, entry); err != nil {
		return fmt.Errorf("add entry: %w", err)
	}

	// Read the updated index content for git commit
	idxPath := filepath.Join(kbRoot, "kb", "index.md")
	newContent, err := os.ReadFile(idxPath) //nolint:gosec // path validated by ResolveKBPath
	if err != nil {
		return fmt.Errorf("read updated index: %w", err)
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	ctx := context.Background()
	commitMsg := fmt.Sprintf("akb: index add %s", relPath)
	if err := store.WriteWithCommitMsg(ctx, idxPath, newContent, commitMsg); err != nil {
		return fmt.Errorf("commit index: %w", err)
	}

	fmt.Printf("Added %s to index\n", relPath)
	return nil
}

func runIndexRemove(_ *cobra.Command, args []string) error {
	entryPath := args[0]

	// Reject index.md and log.md
	base := filepath.Base(entryPath)
	if base == "index.md" || base == "log.md" {
		return &usageError{msg: fmt.Sprintf("cannot remove %s from index", base)}
	}

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	_, err = path.ResolveKBPath(kbRoot, entryPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Hold the repository lock across the read-modify-write of kb/index.md and
	// the commit that records it, so a command running concurrently against the
	// same index cannot overwrite this removal.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	indexPath := filepath.Join(kbRoot, "kb", "index.md")
	oldContent, _ := os.ReadFile(indexPath) //nolint:gosec // path validated by ResolveKBPath

	// Normalize path to kb/ prefix
	cleanPath := strings.TrimPrefix(entryPath, "kb/")
	relPath := filepath.Join("kb", cleanPath)
	relPath = filepath.ToSlash(relPath)

	if err := index.RemoveEntry(kbRoot, relPath); err != nil {
		return fmt.Errorf("remove entry: %w", err)
	}

	newContent, err := os.ReadFile(indexPath) //nolint:gosec // path validated by ResolveKBPath
	if err != nil {
		return fmt.Errorf("read updated index: %w", err)
	}

	if string(oldContent) == string(newContent) {
		fmt.Printf("Removed %s from index\n", relPath)
		return nil
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	ctx := context.Background()
	commitMsg := fmt.Sprintf("akb: index remove %s", relPath)
	if err := store.WriteWithCommitMsg(ctx, indexPath, newContent, commitMsg); err != nil {
		return fmt.Errorf("commit index: %w", err)
	}

	fmt.Printf("Removed %s from index\n", relPath)
	return nil
}

func runIndexRebuild(_ *cobra.Command, _ []string) error {
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	indexPath := filepath.Join(kbRoot, "kb", "index.md")

	// Hold the repository lock across the rebuild of kb/index.md and the commit
	// that records it, so a command running concurrently against the same index
	// cannot overwrite the rebuilt entries.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	oldContent, _ := os.ReadFile(indexPath) //nolint:gosec // path validated by ResolveKBPath

	// Open or create search database
	sqlDB, err := openOrCreateSearchDB(kbRoot)
	if err != nil {
		return fmt.Errorf("open search database: %w", err)
	}
	defer sqlDB.Close() //nolint:errcheck // DB close error non-critical on command exit

	// Rebuild search index (also rebuilds index.md via index.RebuildIndex)
	ctx := context.Background()
	searcher := search.NewSQLiteFTS5Searcher(sqlDB)
	if err := searcher.RebuildIndex(ctx, kbRoot); err != nil {
		return fmt.Errorf("rebuild search index: %w", err)
	}

	linkGraph := linkgraph.NewSQLiteLinkGraph(sqlDB)
	if err := linkGraph.RebuildLinks(ctx, kbRoot); err != nil {
		return fmt.Errorf("rebuild link graph: %w", err)
	}

	newContent, err := os.ReadFile(indexPath) //nolint:gosec // path validated by ResolveKBPath
	if err != nil {
		return fmt.Errorf("read rebuilt index: %w", err)
	}

	if string(oldContent) == string(newContent) {
		fmt.Println("Index rebuilt")
		return nil
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	if err := store.WriteWithCommitMsg(ctx, indexPath, newContent, "akb: index rebuild"); err != nil {
		return fmt.Errorf("commit index: %w", err)
	}

	fmt.Println("Index rebuilt")
	return nil
}

// openOrCreateSearchDB opens the search database, creating it if it doesn't exist.
func openOrCreateSearchDB(kbRoot string) (*sql.DB, error) {
	dbPath := filepath.Join(kbRoot, ".akb", "search.db")

	// Try to open existing database
	sqlDB, err := db.OpenKB(kbRoot)
	if err == nil {
		return sqlDB, nil
	}

	// If error is not about missing DB, return the error
	if !isSearchDBMissing(kbRoot, err) {
		return nil, fmt.Errorf("open search database: %w", err)
	}

	// Create the .akb directory if it doesn't exist
	akbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(akbDir, 0750); err != nil {
		return nil, fmt.Errorf("create .akb directory: %w", err)
	}

	// Initialize and create schema for new database
	sqlDB, err = db.InitDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	if err := db.CreateSchema(sqlDB); err != nil {
		_ = sqlDB.Close() //nolint:errcheck // close error secondary to schema error
		return nil, fmt.Errorf("create schema: %w", err)
	}

	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = sqlDB.Close() //nolint:errcheck // close error secondary to WAL error
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	return sqlDB, nil
}

// isSearchDBMissing checks if the error indicates a missing or invalid database.
func isSearchDBMissing(kbRoot string, err error) bool {
	if err == nil {
		return false
	}
	dbPath := filepath.Join(kbRoot, ".akb", "search.db")
	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "invalid schema")
}
