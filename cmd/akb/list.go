package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/path"
)

// pageInfo holds page metadata for list output.
type pageInfo struct {
	Path    string `json:"path"`
	IsDraft bool   `json:"is_draft"`
}

var listJSON bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pages in the knowledge base",
	Long:  `Display all markdown pages in the kb/ directory, sorted alphabetically.`,
	Example: `  # List all pages
  akb list`,
	Args: cobra.NoArgs,
	RunE: runList,
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
}

func runList(_ *cobra.Command, _ []string) error {
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	pages, err := listPages(kbRoot)
	if err != nil {
		return err
	}

	if listJSON {
		data, err := json.Marshal(pages)
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	for _, page := range pages {
		if page.IsDraft {
			fmt.Printf("[DRAFT] %s\n", page.Path)
		} else {
			fmt.Println(page.Path)
		}
	}

	return nil
}

func listPages(kbRoot string) ([]pageInfo, error) {
	kbDir := filepath.Join(kbRoot, "kb")
	if _, err := os.Stat(kbDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("check kb directory: %w", err)
	}

	var pages []pageInfo
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

		content, err := os.ReadFile(path) //nolint:gosec // path is produced by walking the kb directory
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		var isDraft bool
		fm, _, err := frontmatter.Parse(content)
		if err != nil {
			isDraft = true
		} else {
			isDraft = frontmatter.IsDraft(fm.Fields)
		}

		pages = append(pages, pageInfo{Path: relPath, IsDraft: isDraft})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk kb directory: %w", err)
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Path < pages[j].Path
	})
	return pages, nil
}
