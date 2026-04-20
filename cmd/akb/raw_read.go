package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/spf13/cobra"
)

var rawReadCmd = &cobra.Command{
	Use:   "read <path>",
	Short: "Read a raw file from the knowledge base",
	Long:  `Output the contents of a raw file from the knowledge base. The path is relative to the raw/ directory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRawRead,
}

func runRawRead(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	fullPath, err := path.ResolveRawPath(kbRoot, inputPath)
	if err != nil {
		if errors.Is(err, path.ErrUseAKBWrite) {
			return fmt.Errorf("use `akb read`")
		}
		return err
	}

	provider := storage.NewFilesystemProvider(kbRoot)

	data, err := provider.Read(context.Background(), fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("raw file not found: %s", inputPath)
		}
		return err
	}

	fmt.Print(string(data))
	return nil
}
