package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/markdown"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

var approveCmd = &cobra.Command{
	Use:   "approve <path>",
	Short: "Approve a page by removing draft status and annotations",
	Long:  `Read a page, strip olw-auto annotations and provenance markers, set is_draft to false, and write it back.`,
	Example: `  # Approve a draft page
  akb approve notes/my-draft.md`,
	Args: cobra.ExactArgs(1),
	RunE: runApprove,
}

var annotationRe = regexp.MustCompile(`(?s)<!--\s*olw-auto:.*?-->`)

func runApprove(_ *cobra.Command, args []string) error {
	inputPath := args[0]

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")
	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	ctx := context.Background()
	store := storage.NewGitProvider(kbRoot, noCommit)

	content, err := store.Read(ctx, fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("page '%s' not found. Use 'akb list' to see available pages", inputPath)
		}
		return fmt.Errorf("read page: %w", err)
	}

	fm, body, err := frontmatter.Parse(content)
	if err != nil {
		return fmt.Errorf("parse frontmatter: %w", err)
	}

	if !frontmatter.IsDraft(fm.Fields) {
		fmt.Printf("Page '%s' is already approved\n", inputPath)
		return nil
	}

	bodyStr := string(body)
	bodyStr = annotationRe.ReplaceAllString(bodyStr, "")
	bodyStr = markdown.StripProvenanceMarkers(bodyStr)

	fm.Fields["is_draft"] = false

	allFields := map[string]any{
		"type":  fm.Type,
		"title": fm.Title,
	}
	for k, v := range fm.Fields {
		allFields[k] = v
	}

	yamlBytes, err := yaml.Marshal(allFields)
	if err != nil {
		return fmt.Errorf("re-serialize frontmatter: %w", err)
	}

	finalContent := []byte("---\n" + string(yamlBytes) + "---\n" + bodyStr)

	if err := store.WriteWithCommitMsg(ctx, fullPath, finalContent, fmt.Sprintf("akb: approve %s", inputPath)); err != nil {
		return fmt.Errorf("write page: %w", err)
	}

	fmt.Printf("Approved '%s'\n", inputPath)
	return nil
}
