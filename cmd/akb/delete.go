// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
	logmod "github.com/peedrr/agent-kb/internal/log"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/peedrr/agent-kb/internal/storage"
)

var (
	deleteOrphans bool
	deleteForce   bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a page from the knowledge base",
	Long:  `Remove a page from the KB and commit the change to git.`,
	Example: `  # Delete a page
  akb delete notes/my-note.md

  # Delete with kb/ prefix (optional)
  akb delete kb/adr/use-sqlite-search.md`,
	Args: func(cmd *cobra.Command, args []string) error {
		if deleteOrphans {
			if len(args) > 0 {
				return fmt.Errorf("accepts no args when --orphans is set, received %d", len(args))
			}
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: runDeleteCmd,
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteOrphans, "orphans", false, "delete all orphan pages")
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "force deletion without confirmation")
}

func runDeleteCmd(_ *cobra.Command, args []string) error {
	ctx := context.Background()

	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	dbConn, err := db.OpenKB(kbRoot)
	if err != nil {
		if isMissingDB(err) {
			return fmt.Errorf("run `akb index rebuild` to create the search index")
		}
		return fmt.Errorf("open search database: %w", err)
	}
	defer dbConn.Close() //nolint:errcheck // DB close error non-critical on command exit

	if deleteOrphans {
		g := linkgraph.NewSQLiteLinkGraph(dbConn)
		orphans, err := g.GetOrphans(ctx)
		if err != nil {
			return fmt.Errorf("query orphans: %w", err)
		}

		var filtered []string
		for _, p := range orphans {
			base := filepath.Base(p)
			if base == "index.md" || base == "log.md" {
				continue
			}
			filtered = append(filtered, p)
		}

		if len(filtered) == 0 {
			fmt.Println("No orphan pages found")
			return nil
		}

		if !deleteForce {
			fmt.Println("Would delete:")
			for _, p := range filtered {
				fmt.Printf("  %s\n", p)
			}
			fmt.Println("use --force to confirm deletion")
			return nil
		}

		deleted := 0
		for _, relPath := range filtered {
			fullPath, resolveErr := path.ResolveKBPath(kbRoot, relPath)
			if resolveErr != nil {
				fmt.Fprintf(os.Stderr, "warning: invalid orphan path %s: %v — skipping\n", relPath, resolveErr)
				continue
			}
			cleanPath := strings.TrimPrefix(relPath, "kb/")

			exists, err := fileExists(fullPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: check file existence for %s: %v\n", relPath, err)
				continue
			}
			if !exists {
				fmt.Fprintf(os.Stderr, "warning: page not found: %s — skipping\n", relPath)
				continue
			}

			title := ""
			if content, err := os.ReadFile(fullPath); err == nil { //nolint:gosec // path validated by ResolveKBPath
				if fm, _, err := frontmatter.Parse(content); err == nil {
					title = fm.Title
				}
			}

			if err := deletePage(ctx, kbRoot, dbConn, relPath, fullPath, cleanPath, title); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to delete %s: %v\n", relPath, err)
				continue
			}
			deleted++
		}

		fmt.Printf("Deleted %d orphan pages\n", deleted)
		return nil
	}

	inputPath := args[0]

	if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
		return &usageError{msg: "use `akb raw delete`"}
	}

	fullPath, err := path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")

	// Guard: block delete on managed files
	base := filepath.Base(cleanPath)
	if base == "index.md" {
		return &usageError{msg: "cannot delete index.md; use 'akb index rebuild' to reset"}
	}
	if base == "log.md" {
		return &usageError{msg: "cannot delete log.md; it is a managed file"}
	}

	exists, err := fileExists(fullPath)
	if err != nil {
		return fmt.Errorf("check file existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("page not found: %s. Use 'akb list' to see available pages", inputPath)
	}

	relPath := filepath.Join("kb", cleanPath)
	relPath = filepath.ToSlash(relPath)

	// Try to get title from frontmatter for log entry
	title := ""
	if content, err := os.ReadFile(fullPath); err == nil { //nolint:gosec // path validated by ResolveKBPath
		if fm, _, err := frontmatter.Parse(content); err == nil {
			title = fm.Title
		}
	}

	if err := deletePage(ctx, kbRoot, dbConn, relPath, fullPath, cleanPath, title); err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", relPath)

	return nil
}

func deletePage(ctx context.Context, kbRoot string, dbConn *sql.DB, relPath, fullPath, cleanPath, title string) error {
	// Hold the repository lock across the whole deletion: the index and log
	// updates, the commit, and the search and link-graph removals.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	// The search index and the link graph describe the same deleted page, so
	// both removals share one transaction: a failure in either leaves both
	// records in place instead of one index ahead of the other.
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin index transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	searcher := search.NewSQLiteFTS5Searcher(dbConn)
	if err := searcher.RemovePageTx(ctx, tx, relPath); err != nil {
		return fmt.Errorf("remove from search index: %w", err)
	}

	updater := linkgraph.NewSQLiteLinkGraph(dbConn)
	if err := updater.RemovePageTx(ctx, tx, relPath); err != nil {
		return fmt.Errorf("remove from link graph: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit index transaction: %w", err)
	}

	if err := index.RemoveEntry(kbRoot, relPath); err != nil {
		switch {
		case os.IsNotExist(err):
			fmt.Fprintln(os.Stderr, "warning: index.md not found at kb/index.md — run 'akb index rebuild' to regenerate")
		case strings.Contains(err.Error(), "not found in"):
			fmt.Fprintln(os.Stderr, "warning: page not found in index — run 'akb index rebuild' to resync")
		default:
			fmt.Fprintf(os.Stderr, "warning: failed to update index.md: %v — run 'akb index rebuild' to resync\n", err)
		}
	}

	if err := logmod.AppendLog(kbRoot, "delete", "Removed page "+cleanPath, title); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to log deletion: %v — check that kb/log.md exists\n", err)
	}

	if err := storage.StageFiles(kbRoot, "kb/index.md", "kb/log.md"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to stage index.md/log.md for commit: %v\n", err)
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	if err := store.Delete(ctx, fullPath); err != nil {
		return fmt.Errorf("delete page: %w — the page was already removed from the search index and link graph but its file removal or its commit did not complete; if the page file still exists, run `akb index rebuild` to rebuild the search index and link graph; if the file is gone, run `git status` and commit the staged deletion manually", err)
	}

	return nil
}

var fileExists = func(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("check file existence: %w", err)
}
