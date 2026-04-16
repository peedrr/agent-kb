package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/index"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	logmod "github.com/peedrr/agent-kb/internal/log"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a page from the knowledge base",
	Long:  `Remove a page from the KB and commit the change to git.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDeleteCmd,
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
		return fmt.Errorf("use `akb raw delete`")
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")

	// Guard: block delete on managed files
	if cleanPath == "index.md" {
		return fmt.Errorf("cannot delete index.md; use 'akb index rebuild' to reset")
	}
	if cleanPath == "log.md" {
		return fmt.Errorf("cannot delete log.md; it is a managed file")
	}

	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	exists, err := fileExists(fullPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("page not found: %s", inputPath)
	}

	relPath := filepath.Join("kb", cleanPath)
	relPath = filepath.ToSlash(relPath)

	// Try to get title from frontmatter for log entry
	title := ""
	if content, err := os.ReadFile(fullPath); err == nil {
		if fm, _, err := frontmatter.Parse(content); err == nil {
			title = fm.Title
		}
	}

	// Remove from index.md (succeeds silently if not in index)
	_ = index.RemoveEntry(kbRoot, relPath)

	// Log the deletion
	_ = logmod.AppendLog(kbRoot, "delete", "Removed page "+cleanPath, title)

	// Stage index.md and log.md so they're included in the delete commit
	gitAdd := exec.Command("git", "add", "kb/index.md", "kb/log.md")
	gitAdd.Dir = kbRoot
	gitAdd.Run() // best effort

	store := storage.NewGitProvider(kbRoot, noCommit)
	ctx := context.Background()
	if err := store.Delete(ctx, fullPath); err != nil {
		return err
	}

	updater := &linkgraph.NoOpLinkGraphUpdater{}
	if err := updater.RemovePage(ctx, relPath); err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", relPath)

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
	return false, err
}
