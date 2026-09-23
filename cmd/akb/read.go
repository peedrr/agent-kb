package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

var readCmd = &cobra.Command{
	Use:   "read <path>",
	Short: "Read a page from the knowledge base",
	Long:  `Display the contents of a page from the knowledge base. The path is relative to the kb/ directory.`,
	Example: `  # Read a page
  akb read notes/my-note.md

  # Read from a subdirectory
  akb read adr/use-sqlite-search.md`,
	Args: cobra.ExactArgs(1),
	RunE: runRead,
}

func runRead(_ *cobra.Command, args []string) error {
	inputPath := args[0]

	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	_, err = config.Load(filepath.Join(kbRoot, ".akb", ".akb.yaml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	resolvedPath, err := path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		if errors.Is(err, path.ErrUseAKBRawWrite) {
			return &usageError{msg: "use `akb raw read`"}
		}
		return fmt.Errorf("resolve path: %w", err)
	}

	provider := storage.NewFilesystemProvider(kbRoot)

	data, err := provider.Read(context.Background(), resolvedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("page not found: %s. Use 'akb list' to see available pages", inputPath)
		}
		return fmt.Errorf("read page: %w", err)
	}

	fmt.Print(string(data))
	return nil
}
