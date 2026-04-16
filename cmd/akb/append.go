package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/spf13/cobra"
)

var appendCmd = &cobra.Command{
	Use:   "append <path>",
	Short: "Append content to an existing page in the knowledge base",
	Long:  `Read content from stdin and append it to the body of an existing page. Frontmatter is preserved.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAppend,
}

var isStdinTTY = func() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	return (stat.Mode() & os.ModeCharDevice) != 0, nil
}

func runAppend(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	isTTY, err := isStdinTTY()
	if err != nil {
		return fmt.Errorf("check stdin: %w", err)
	}
	if isTTY {
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

	_, err = config.Load(filepath.Join(kbRoot, ".akb", ".akb.yaml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
		return fmt.Errorf("use `akb raw write`")
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")

	_, err = path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		return err
	}

	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	exists, err := fileExists(fullPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("page not found: %s. Use `akb write` to create.", inputPath)
	}

	store := storage.NewGitProvider(kbRoot, true)
	ctx := context.Background()

	existingContent, err := store.Read(ctx, fullPath)
	if err != nil {
		return fmt.Errorf("read page: %w", err)
	}

	fm, body, err := frontmatter.Parse(existingContent)
	if err != nil {
		return fmt.Errorf("parse frontmatter: %w", err)
	}

	closingLF := []byte("\n---\n")
	closingCRLF := []byte("\r\n---\r\n")

	var frontmatterEnd int
	closingIdx := bytes.Index(existingContent, closingLF)
	if closingIdx != -1 {
		frontmatterEnd = closingIdx + len(closingLF)
	} else {
		closingIdx = bytes.Index(existingContent, closingCRLF)
		if closingIdx != -1 {
			frontmatterEnd = closingIdx + len(closingCRLF)
		} else {
			return fmt.Errorf("invalid frontmatter format: closing delimiter not found")
		}
	}

	frontmatterPortion := existingContent[:frontmatterEnd]
	newBody := string(body) + "\n" + string(stdinContent)
	fullContent := string(frontmatterPortion) + newBody

	if err := store.Write(ctx, fullPath, []byte(fullContent)); err != nil {
		return fmt.Errorf("write page: %w", err)
	}

	relPath := filepath.Join("kb", cleanPath)
	relPath = filepath.ToSlash(relPath)

	if !noCommit {
		if err := gitCommitAppend(kbRoot, relPath); err != nil {
			return err
		}
	}

	searcher := &search.NoOpSearcher{}
	tags := ""
	if t, ok := fm.Fields["tags"]; ok {
		tags = fmt.Sprintf("%v", t)
	}
	summary := ""
	if s, ok := fm.Fields["summary"]; ok {
		summary = fmt.Sprintf("%v", s)
	}
	if err := searcher.IndexPage(ctx, relPath, fm.Title, string(body), tags, summary); err != nil {
		return fmt.Errorf("index page: %w", err)
	}

	updater := &linkgraph.NoOpLinkGraphUpdater{}
	if err := updater.UpdatePageLinks(ctx, relPath, string(fullContent)); err != nil {
		return fmt.Errorf("update links: %w", err)
	}

	fmt.Printf("Appended to %s\n", relPath)

	return nil
}

func gitCommitAppend(kbRoot, relPath string) error {
	gitConfig := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = kbRoot
		_, err := cmd.Output()
		return err
	}
	if err := gitConfig("config", "--local", "user.name"); err != nil {
		if err := gitConfig("config", "user.name", "akb"); err != nil {
			return fmt.Errorf("set git user.name: %w", err)
		}
	}
	if err := gitConfig("config", "--local", "user.email"); err != nil {
		if err := gitConfig("config", "user.email", "akb@local"); err != nil {
			return fmt.Errorf("set git user.email: %w", err)
		}
	}

	commitMsg := fmt.Sprintf("akb: append %s", relPath)
	cmd := exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = kbRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
