package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/path"
)

var rawListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all raw files",
	Long:  `Display all files in the raw/ directory, sorted alphabetically, excluding files.log.`,
	Example: `  # List all raw files
  akb raw list`,
	Args: cobra.NoArgs,
	RunE: runRawList,
}

func runRawList(_ *cobra.Command, _ []string) error {
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	rawDir := filepath.Join(kbRoot, "raw")
	if _, err := os.Stat(rawDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat raw directory: %w", err)
	}

	var files []string
	err = filepath.WalkDir(rawDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rawDir, p)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}
		relPath = filepath.ToSlash(relPath)

		if relPath == "files.log" {
			return nil
		}

		files = append(files, relPath)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk raw directory: %w", err)
	}

	sort.Strings(files)

	for _, f := range files {
		fmt.Println(f)
	}

	return nil
}
