package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show KB status information",
	Long:  `Display the current KB's name, path, page count, and git status.`,
	Args:  cobra.NoArgs,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(filepath.Join(kbRoot, ".akb", ".akb.yaml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	absPath, err := filepath.Abs(kbRoot)
	if err != nil {
		return fmt.Errorf("resolve absolute path: %w", err)
	}

	pageCount, err := countPages(kbRoot)
	if err != nil {
		return fmt.Errorf("count pages: %w", err)
	}

	gitStatus, err := getGitStatus(kbRoot)
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}

	fmt.Printf("Name: %s\n", cfg.Name)
	fmt.Printf("Path: %s\n", absPath)
	fmt.Printf("Pages: %d\n", pageCount)
	fmt.Printf("Git: %s\n", gitStatus)

	return nil
}

func countPages(kbRoot string) (int, error) {
	kbDir := filepath.Join(kbRoot, "kb")
	if _, err := os.Stat(kbDir); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	count := 0
	err := filepath.Walk(kbDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, err := filepath.Rel(kbDir, path)
			if err != nil {
				return err
			}
			base := filepath.Base(relPath)
			parent := filepath.Dir(relPath)
			if base == "index.md" && parent == "." {
				return nil
			}
			if base == "log.md" && parent == "." {
				return nil
			}
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func getGitStatus(kbRoot string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = kbRoot
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git: %w", err)
	}

	if strings.TrimSpace(string(out)) == "" {
		return "clean", nil
	}
	return "dirty", nil
}
