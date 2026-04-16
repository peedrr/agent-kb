package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/peedrr/agent-kb/internal/template"
	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:   "write <path>",
	Short: "Write a page to the knowledge base",
	Long:  `Read content from stdin, validate frontmatter, and write to the type-derived directory in the KB.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWrite,
}

func runWrite(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

	// Read from stdin — error if stdin is a TTY
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

	// Resolve KB root
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	// Load config
	_, err = config.Load(filepath.Join(kbRoot, ".akb", ".akb.yaml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Load templates
	templates, err := template.LoadTemplates(filepath.Join(kbRoot, ".akb", "templates"))
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	// Parse frontmatter
	fm, body, err := frontmatter.Parse(stdinContent)
	if err != nil {
		return err
	}

	// Validate type
	if err := frontmatter.ValidateType(fm, templates); err != nil {
		return err
	}

	// Validate title
	if err := frontmatter.ValidateTitle(fm); err != nil {
		return err
	}

	// Resolve type-derived directory
	tmpl, ok := templates[fm.Type]
	if !ok {
		// Should not happen after ValidateType, but be safe
		return fmt.Errorf("unknown type %q", fm.Type)
	}
	dirFromType := tmpl.Dir

	// Validate .md extension
	if !strings.HasSuffix(inputPath, ".md") {
		return fmt.Errorf("page filename must end with .md")
	}

	// Reject raw/ prefix
	if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
		return fmt.Errorf("use `akb raw write`")
	}

	// Strip kb/ prefix
	cleanPath := strings.TrimPrefix(inputPath, "kb/")

	// Strip type-dir prefix if it matches the type's Dir
	if dirFromType != "" {
		typeDirPrefix := dirFromType + "/"
		if strings.HasPrefix(cleanPath, typeDirPrefix) {
			cleanPath = strings.TrimPrefix(cleanPath, typeDirPrefix)
		}
	}

	// Reject .. and absolute paths via ResolveKBPath
	_, err = path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		return err
	}

	// Construct final path
	var fullPath string
	if dirFromType != "" {
		fullPath = filepath.Join(kbRoot, "kb", dirFromType, cleanPath)
	} else {
		fullPath = filepath.Join(kbRoot, "kb", cleanPath)
	}

	// Compute relative path for output and indexing
	var relPath string
	if dirFromType != "" {
		relPath = filepath.Join("kb", dirFromType, cleanPath)
	} else {
		relPath = filepath.Join("kb", cleanPath)
	}
	relPath = filepath.ToSlash(relPath)

	// Extract tags and summary from frontmatter fields
	tags := ""
	if t, ok := fm.Fields["tags"]; ok {
		tags = fmt.Sprintf("%v", t)
	}
	summary := ""
	if s, ok := fm.Fields["summary"]; ok {
		summary = fmt.Sprintf("%v", s)
	}

	// Create storage provider
	store := storage.NewGitProvider(kbRoot, noCommit)

	// Write content
	ctx := context.Background()
	if err := store.Write(ctx, fullPath, stdinContent); err != nil {
		return fmt.Errorf("write page: %w", err)
	}

	// Call NoOp searcher
	searcher := &search.NoOpSearcher{}
	if err := searcher.IndexPage(ctx, relPath, fm.Title, string(body), tags, summary); err != nil {
		return fmt.Errorf("index page: %w", err)
	}

	// Call NoOp link graph updater
	updater := &linkgraph.NoOpLinkGraphUpdater{}
	if err := updater.UpdatePageLinks(ctx, relPath, string(stdinContent)); err != nil {
		return fmt.Errorf("update links: %w", err)
	}

	// Output
	fmt.Printf("Written to %s\n", relPath)
	fmt.Printf("Don't forget to update the index! `akb index add %s <summary>`\n", relPath)

	return nil
}
