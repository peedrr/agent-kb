package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/peedrr/agent-kb/internal/path"
	"github.com/spf13/cobra"
)

var rawListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all raw files",
	Long:  `Display all files in the raw/ directory, sorted alphabetically, excluding files.log.`,
	Args:  cobra.NoArgs,
	RunE:  runRawList,
}

func runRawList(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return err
	}

	rawDir := filepath.Join(kbRoot, "raw")
	if _, err := os.Stat(rawDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
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
			return err
		}
		relPath = filepath.ToSlash(relPath)

		if relPath == "files.log" {
			return nil
		}

		files = append(files, relPath)
		return nil
	})
	if err != nil {
		return err
	}

	sort.Strings(files)

	for _, f := range files {
		fmt.Println(f)
	}

	return nil
}
