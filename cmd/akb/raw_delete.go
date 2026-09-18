package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

var rawDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a raw file from the knowledge base",
	Long:  `Remove a raw file from the knowledge base and update the manifest. Warns about KB pages that reference the file in their sources field.`,
	Example: `  # Delete a raw file
  akb raw delete data/config.json`,
	Args: cobra.ExactArgs(1),
	RunE: runRawDelete,
}

func runRawDelete(_ *cobra.Command, args []string) error {
	inputPath := args[0]

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	fullPath, err := path.ResolveRawPath(kbRoot, inputPath)
	if err != nil {
		return fmt.Errorf("resolve raw path: %w", err)
	}

	// Get relative path from raw/ directory
	relPath := strings.TrimPrefix(inputPath, "raw/")
	if relPath == inputPath {
		relPath = inputPath
	}
	if inputPath == "raw" {
		relPath = "."
	}
	relPath = filepath.ToSlash(relPath)

	// Block files.log as managed file
	if filepath.Base(relPath) == "files.log" {
		return fmt.Errorf("cannot delete files.log; it is a managed file")
	}

	// Verify file exists before deletion
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("raw file not found: %s", inputPath)
		}
		return fmt.Errorf("stat file: %w", err)
	}

	// Sources scan: find KB pages referencing this file (before deletion)
	referencing := scanSources(kbRoot, relPath)

	// Hold the repository lock across the file and manifest removals and the
	// commit that records them.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	// Delete file from disk
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	// Remove entry from manifest
	mgr := manifest.NewManager(kbRoot)
	if err := mgr.RemoveEntry(relPath); err != nil {
		return fmt.Errorf("remove manifest entry: %w", err)
	}

	// Git commit
	if !noCommit {
		commitMsg := fmt.Sprintf("akb: raw delete %s", relPath)
		if err := storage.CommitFiles(kbRoot, commitMsg, filepath.Join("raw", relPath), filepath.Join("raw", "files.log")); err != nil {
			return fmt.Errorf("commit raw delete: %w", err)
		}
	}

	fmt.Printf("Deleted raw/%s\n", relPath)

	if len(referencing) > 0 {
		fmt.Printf("Pages referencing this file: %s\n", strings.Join(referencing, ", "))
	}

	return nil
}

// scanSources walks all KB .md files and returns paths of pages
// whose frontmatter sources field references the deleted file.
func scanSources(kbRoot, deletedFile string) []string {
	var referencing []string
	kbDir := filepath.Join(kbRoot, "kb")
	_ = filepath.WalkDir(kbDir, func(p string, d fs.DirEntry, err error) error { //nolint:errcheck // best-effort scan; errors handled per-file
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		content, err := os.ReadFile(p) //nolint:gosec // path validated by filepath.WalkDir within KB root
		if err != nil {
			return nil
		}
		fm, _, err := frontmatter.Parse(content)
		if err != nil {
			return nil // skip pages without frontmatter
		}
		sourcesRaw, ok := fm.Fields["sources"]
		if !ok {
			return nil
		}
		// sources can be string or []any
		switch v := sourcesRaw.(type) {
		case string:
			if v == deletedFile {
				relPath, _ := filepath.Rel(kbDir, p)
				referencing = append(referencing, filepath.ToSlash(relPath))
			}
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok && s == deletedFile {
					relPath, _ := filepath.Rel(kbDir, p)
					referencing = append(referencing, filepath.ToSlash(relPath))
				}
			}
		}
		return nil
	})
	return referencing
}
