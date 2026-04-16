package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/linkgraph"
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

	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	exists, err := fileExists(fullPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("page not found: %s", inputPath)
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	ctx := context.Background()
	if err := store.Delete(ctx, fullPath); err != nil {
		return err
	}

	updater := &linkgraph.NoOpLinkGraphUpdater{}
	relPath := filepath.Join("kb", cleanPath)
	if err := updater.RemovePage(ctx, relPath); err != nil {
		return err
	}

	relPathOutput := filepath.ToSlash(relPath)

	fmt.Printf("Deleted %s\n", relPathOutput)

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
