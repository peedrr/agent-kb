package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/path"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pages in the knowledge base",
	Long:  `Display all markdown pages in the kb/ directory, sorted alphabetically.`,
	Example: `  # List all pages
  akb list`,
	Args: cobra.NoArgs,
	RunE: runList,
}

func runList(_ *cobra.Command, _ []string) error {
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	pages, err := listPages(kbRoot)
	if err != nil {
		return err
	}

	for _, page := range pages {
		fmt.Println(page)
	}

	return nil
}

func listPages(kbRoot string) ([]string, error) {
	kbDir := filepath.Join(kbRoot, "kb")
	if _, err := os.Stat(kbDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("check kb directory: %w", err)
	}

	var pages []string
	err := filepath.Walk(kbDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath := strings.TrimPrefix(path, kbDir+string(filepath.Separator))
		if relPath == "index.md" || relPath == "log.md" {
			return nil
		}

		pages = append(pages, relPath)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk kb directory: %w", err)
	}

	sort.Strings(pages)
	return pages, nil
}
