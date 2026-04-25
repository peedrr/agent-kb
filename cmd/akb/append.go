// Package main provides the akb CLI commands.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/peedrr/agent-kb/internal/storage"
)

var appendCmd = &cobra.Command{
	Use:   "append <path>",
	Short: "Append content to an existing page in the knowledge base",
	Long:  `Read content from stdin and append it to the body of an existing page. Frontmatter is preserved.`,
	Example: `  # Append content to a page
  echo "
More content here" | akb append notes/my-note.md`,
	Args: cobra.ExactArgs(1),
	RunE: runAppend,
}

var isStdinTTY = func() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("check stdin: %w", err)
	}
	return (stat.Mode() & os.ModeCharDevice) != 0, nil
}

func runAppend(_ *cobra.Command, args []string) error {
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

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	dbConn, err := db.OpenKB(kbRoot)
	if err != nil {
		if isMissingDB(err) {
			return fmt.Errorf("run `akb index rebuild` to create the search index")
		}
		return fmt.Errorf("open search database: %w", err)
	}
	defer dbConn.Close() //nolint:errcheck // DB close error non-critical on command exit

	_, err = config.Load(filepath.Join(kbRoot, ".akb", ".akb.yaml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
		return fmt.Errorf("use `akb raw write`")
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")

	// Guard: block append to managed files
	base := filepath.Base(cleanPath)
	if base == "index.md" {
		return fmt.Errorf("cannot append to index.md; use 'akb index add' to update")
	}
	if base == "log.md" {
		return fmt.Errorf("cannot append to log.md; it is a managed file")
	}

	_, err = path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	exists, err := fileExists(fullPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("page not found: %s. Use `akb write` to create", inputPath)
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
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

	relPath := filepath.ToSlash(filepath.Join("kb", cleanPath))
	commitMsg := fmt.Sprintf("akb: append %s", relPath)
	if err := store.WriteWithCommitMsg(ctx, fullPath, []byte(fullContent), commitMsg); err != nil {
		return fmt.Errorf("write page: %w", err)
	}

	searcher := search.NewSQLiteFTS5Searcher(dbConn)
	tags := search.ExtractTags(fm.Fields)
	summary := search.ExtractSummary(fm.Fields)
	if err := searcher.IndexPage(ctx, relPath, fm.Title, newBody, tags, summary, fm.Type); err != nil {
		return fmt.Errorf("index page: %w", err)
	}

	updater := linkgraph.NewSQLiteLinkGraph(dbConn)
	if err := updater.UpdatePageLinks(ctx, relPath, fullContent); err != nil {
		return fmt.Errorf("update links: %w", err)
	}

	fmt.Printf("Appended to %s\n", relPath)

	return nil
}
