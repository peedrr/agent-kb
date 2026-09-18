package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

var rawWriteCmd = &cobra.Command{
	Use:   "write <path>",
	Short: "Write a raw file to the knowledge base",
	Long:  `Read content from stdin and write it as a raw file in the knowledge base. Updates the SHA-256 manifest and commits both changes together.`,
	Example: `  # Write a raw file
  echo "file content here" | akb raw write data/config.json

  # Write a binary file (base64 encoded)
  base64 -d <<< "SGVsbG8gV29ybGQ=" | akb raw write hello.txt`,
	Args: cobra.ExactArgs(1),
	RunE: runRawWrite,
}

func runRawWrite(_ *cobra.Command, args []string) error {
	inputPath := args[0]

	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("check stdin: %w", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("input required: pipe content to stdin")
	}

	stdinContent, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	fullPath, err := path.ResolveRawPath(kbRoot, inputPath)
	if err != nil {
		return fmt.Errorf("resolve raw path: %w", err)
	}

	relPath := strings.TrimPrefix(inputPath, "raw/")
	if relPath == inputPath {
		relPath = inputPath
	}
	if inputPath == "raw" {
		relPath = "."
	}
	relPath = filepath.ToSlash(relPath)

	if filepath.Base(relPath) == "files.log" {
		return fmt.Errorf("cannot write files.log directly; it is a managed file")
	}

	if err := manifest.ValidateFilename(relPath); err != nil {
		return fmt.Errorf("validate filename: %w", err)
	}

	// Hold the repository lock across the file and manifest writes and the
	// commit that records them.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
		return fmt.Errorf("create parent directories: %w", err)
	}

	if err := os.WriteFile(fullPath, stdinContent, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	sha256hash, err := manifest.ComputeSHA256(fullPath)
	if err != nil {
		return fmt.Errorf("compute SHA-256: %w", err)
	}

	mgr := manifest.NewManager(kbRoot)
	_, found, err := mgr.FindByFilename(relPath)
	if err != nil {
		return fmt.Errorf("look up manifest entry: %w", err)
	}
	if found {
		if err := mgr.UpdateEntry(relPath, sha256hash); err != nil {
			return fmt.Errorf("update manifest: %w", err)
		}
	} else if err := mgr.AddEntry(relPath, sha256hash); err != nil {
		return fmt.Errorf("update manifest: %w", err)
	}

	if !noCommit {
		commitMsg := fmt.Sprintf("akb: raw write %s", relPath)
		if err := storage.CommitFiles(kbRoot, commitMsg, filepath.Join("raw", relPath), filepath.Join("raw", "files.log")); err != nil {
			return fmt.Errorf("commit raw write: %w", err)
		}
	}

	fmt.Printf("Written to raw/%s\n", relPath)
	return nil
}
