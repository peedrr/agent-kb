package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/peedrr/agent-kb/internal/path"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pages in the knowledge base",
	Long:  `Display all markdown pages in the kb/ directory, sorted alphabetically.`,
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
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
		return nil, err
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
		return nil, err
	}

	sort.Strings(pages)
	return pages, nil
}
