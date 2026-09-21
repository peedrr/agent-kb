package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

var rawReadCmd = &cobra.Command{
	Use:   "read <path>",
	Short: "Read a raw file from the knowledge base",
	Long:  `Output the contents of a raw file from the knowledge base. The path is relative to the raw/ directory.`,
	Example: `  # Read a raw file
  akb raw read data/config.json`,
	Args: cobra.ExactArgs(1),
	RunE: runRawRead,
}

func runRawRead(_ *cobra.Command, args []string) error {
	inputPath := args[0]

	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	fullPath, err := path.ResolveRawPath(kbRoot, inputPath)
	if err != nil {
		if errors.Is(err, path.ErrUseAKBWrite) {
			return &usageError{msg: "use `akb read`"}
		}
		return fmt.Errorf("resolve raw path: %w", err)
	}

	provider := storage.NewFilesystemProvider(kbRoot)

	data, err := provider.Read(context.Background(), fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("raw file not found: %s", inputPath)
		}
		return fmt.Errorf("read file: %w", err)
	}

	fmt.Print(string(data))
	return nil
}
