package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <path>",
	Short: "Read a page from the knowledge base",
	Long:  `Display the contents of a page from the knowledge base. The path is relative to the kb/ directory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRead,
}

func runRead(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	_, err = config.Load(filepath.Join(kbRoot, ".akb", ".akb.yaml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	_, err = path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		if errors.Is(err, path.ErrUseAKBRawWrite) {
			return fmt.Errorf("use `akb raw read`")
		}
		return err
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")
	resolvedPath := filepath.Join(kbRoot, "kb", cleanPath)

	provider := storage.NewFilesystemProvider(kbRoot)

	data, err := provider.Read(context.Background(), resolvedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("page not found: %s", inputPath)
		}
		return err
	}

	fmt.Print(string(data))
	return nil
}
