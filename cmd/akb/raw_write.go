package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/spf13/cobra"
)

var rawWriteCmd = &cobra.Command{
	Use:   "write <path>",
	Short: "Write a raw file to the knowledge base",
	Long:  `Read content from stdin and write it as a raw file in the knowledge base. Updates the SHA-256 manifest and commits both changes together.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRawWrite,
}

func runRawWrite(cmd *cobra.Command, args []string) error {
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

	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	fullPath, err := path.ResolveRawPath(kbRoot, inputPath)
	if err != nil {
		return err
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
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("create parent directories: %w", err)
	}

	if err := os.WriteFile(fullPath, stdinContent, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	sha256hash, err := manifest.ComputeSHA256(fullPath)
	if err != nil {
		return fmt.Errorf("compute SHA-256: %w", err)
	}

	mgr := manifest.NewManager(kbRoot)
	if err := mgr.AddEntry(relPath, sha256hash); err != nil {
		return fmt.Errorf("update manifest: %w", err)
	}

	if !noCommit {
		gitAdd := exec.Command("git", "add", filepath.Join("raw", relPath), filepath.Join("raw", "files.log"))
		gitAdd.Dir = kbRoot
		if out, err := gitAdd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add: %s: %w", strings.TrimSpace(string(out)), err)
		}

		if err := ensureGitConfig(kbRoot); err != nil {
			return fmt.Errorf("git config: %w", err)
		}

		commitMsg := fmt.Sprintf("akb: raw write %s", relPath)
		gitCommit := exec.Command("git", "commit", "-m", commitMsg)
		gitCommit.Dir = kbRoot
		if out, err := gitCommit.CombinedOutput(); err != nil {
			return fmt.Errorf("git commit: %s: %w", strings.TrimSpace(string(out)), err)
		}
	}

	fmt.Printf("Written to raw/%s\n", relPath)
	return nil
}
