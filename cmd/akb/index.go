package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/index"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage the knowledge base index",
	Long:  `Manage the knowledge base index: show, add, remove, or rebuild entries.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var indexShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the index",
	Long:  `Print the contents of kb/index.md to stdout.`,
	Args:  cobra.NoArgs,
	RunE:  runIndexShow,
}

var indexAddCmd = &cobra.Command{
	Use:   "add <path> <summary>",
	Short: "Add or update an entry in the index",
	Long:  `Add or update an entry in kb/index.md with the given path and summary.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runIndexAdd,
}

var indexRemoveCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Remove an entry from the index",
	Long:  `Remove an entry from kb/index.md by path.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runIndexRemove,
}

var indexRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild the index from the filesystem",
	Long:  `Regenerate kb/index.md by walking the kb/ directory and parsing frontmatter from each .md file.`,
	Args:  cobra.NoArgs,
	RunE:  runIndexRebuild,
}

func init() {
	indexCmd.AddCommand(indexShowCmd)
	indexCmd.AddCommand(indexAddCmd)
	indexCmd.AddCommand(indexRemoveCmd)
	indexCmd.AddCommand(indexRebuildCmd)
}

func runIndexShow(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(kbRoot, "kb", "index.md")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("read index: %w", err)
	}

	fmt.Print(string(data))
	return nil
}

func runIndexAdd(cmd *cobra.Command, args []string) error {
	entryPath := args[0]
	summary := args[1]

	// Reject index.md and log.md
	base := filepath.Base(entryPath)
	if base == "index.md" || base == "log.md" {
		return fmt.Errorf("cannot add %s to index", base)
	}

	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	// Resolve the full file path
	cleanPath := strings.TrimPrefix(entryPath, "kb/")
	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	// Read the file to get frontmatter for title and type
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read file %s: %w", entryPath, err)
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
	newContent, err := os.ReadFile(idxPath)
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

func runIndexRemove(cmd *cobra.Command, args []string) error {
	entryPath := args[0]

	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(kbRoot, "kb", "index.md")
	oldContent, _ := os.ReadFile(indexPath)

	// Normalize path to kb/ prefix
	cleanPath := strings.TrimPrefix(entryPath, "kb/")
	relPath := filepath.Join("kb", cleanPath)
	relPath = filepath.ToSlash(relPath)

	if err := index.RemoveEntry(kbRoot, relPath); err != nil {
		return fmt.Errorf("remove entry: %w", err)
	}

	newContent, err := os.ReadFile(indexPath)
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

func runIndexRebuild(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(kbRoot, "kb", "index.md")
	oldContent, _ := os.ReadFile(indexPath)

	if err := index.RebuildIndex(kbRoot); err != nil {
		return fmt.Errorf("rebuild index: %w", err)
	}

	newContent, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("read rebuilt index: %w", err)
	}

	if string(oldContent) == string(newContent) {
		fmt.Println("Index rebuilt")
		return nil
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	ctx := context.Background()
	if err := store.WriteWithCommitMsg(ctx, indexPath, newContent, "akb: index rebuild"); err != nil {
		return fmt.Errorf("commit index: %w", err)
	}

	fmt.Println("Index rebuilt")
	return nil
}
