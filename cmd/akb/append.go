// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package main provides the akb CLI commands.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	yaml "github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"

	"github.com/peedrr/agent-kb/internal/cel"
	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/peedrr/agent-kb/internal/template"
)

var appendCmd = &cobra.Command{
	Use:   "append <path>",
	Short: "Append content to an existing page in the knowledge base",
	Long:  `Read content from stdin and append it to the body of an existing page. Frontmatter is preserved.`,
	Example: `  # Append content to a page
  echo "
More content here" | akb append notes/my-note.md

  # Append a dated section
  echo "More content here" | akb append --dated notes/journal.md`,
	Args: cobra.ExactArgs(1),
	RunE: runAppend,
}

var appendDated bool

func init() {
	appendCmd.Flags().BoolVar(&appendDated, "dated", false, "prefix the appended content with a `## YYYY-MM-DD` heading (local date)")
}

// datedSection wraps content under a `## YYYY-MM-DD` heading. The heading uses
// the local calendar date, matching the heading convention of kb/log.md.
func datedSection(content string) string {
	return "## " + time.Now().Format("2006-01-02") + "\n\n" + content
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
		return &usageError{msg: "input required: pipe content to stdin"}
	}

	stdinContent, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	kbRoot, err := path.ResolveKB(kbFlag)
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
		return &usageError{msg: "use `akb raw write`"}
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")

	// Guard: block append to managed files
	base := filepath.Base(cleanPath)
	if base == "index.md" {
		return &usageError{msg: "cannot append to index.md; use 'akb index add' to update"}
	}
	if base == "log.md" {
		return &usageError{msg: "cannot append to log.md; it is a managed file"}
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	ctx := context.Background()

	// Hold the repository lock across the read-modify-write of the page body and
	// the search and link-graph updates that follow it, so concurrent appends
	// cannot overwrite each other's content.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	// A page is addressed by the path it is stored at or by its bare filename,
	// which resolves under the directory of its type. The type directories are
	// read only when the named path holds no page.
	fullPath, relPath, existingContent, err := resolveExistingPage(kbRoot, inputPath, typeDirsFromDisk(kbRoot))
	if err != nil {
		if errors.Is(err, errPageNotFound) {
			return fmt.Errorf("page not found: %s. Use `akb write` to create", inputPath)
		}
		return err
	}

	fm, body, err := frontmatter.Parse(existingContent)
	if err != nil {
		return fmt.Errorf("parse frontmatter: %w", err)
	}

	// Templates are read the way `akb write` reads them, so an append runs the
	// same write-time validation pipeline as a write.
	templatesDir := filepath.Join(kbRoot, ".akb", "templates")
	if err := path.AssertContained(kbRoot, templatesDir); err != nil {
		return fmt.Errorf("resolve templates directory: %w", err)
	}
	// The guard's error names the offending template file, so it is reported as
	// it stands.
	if err := assertTemplateFilesContained(kbRoot, templatesDir); err != nil {
		return err
	}
	templates, err := template.LoadTemplates(templatesDir)
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	if err := frontmatter.ValidateType(fm, templates); err != nil {
		return fmt.Errorf("validate type: %w", err)
	}
	if err := frontmatter.ValidateTitle(fm); err != nil {
		return fmt.Errorf("validate title: %w", err)
	}

	// old_page comes from the on-disk page, so it keeps the pre-append state.
	oldPage, err := cel.BuildOldPage(relPath, store)
	if err != nil {
		return &internalError{err: fmt.Errorf("CEL engine error: %w", err)}
	}

	// The page content changes, so stamp the update time.
	fm.Fields["updated"] = time.Now().UTC().Format(time.RFC3339)

	appended := string(stdinContent)
	if appendDated {
		appended = datedSection(appended)
	}

	newBody := string(body) + "\n" + appended

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
	fullContent := "---\n" + string(yamlBytes) + "---\n" + newBody

	// Validate the page as the append leaves it: the rules see the appended body
	// and the bumped update time, while old_page holds the pre-append state.
	tmpl, ok := templates[fm.Type]
	if !ok {
		return fmt.Errorf("unknown type %q", fm.Type)
	}
	newBodyBytes := []byte(newBody)
	md := goldmark.New()
	astDoc := md.Parser().Parse(text.NewReader(newBodyBytes))
	page := cel.BuildPage(relPath, fm, newBodyBytes, astDoc, newBodyBytes)

	celEnv, err := cel.NewEnv()
	if err != nil {
		return &internalError{err: fmt.Errorf("CEL engine error: %w", err)}
	}

	// A field the schema marks required must be present before the rules run:
	// the rules guard on key presence, so an absent key would make them
	// vacuously true.
	if missing := checkRequiredFields(tmpl, fm); len(missing) > 0 {
		fmt.Fprintln(os.Stderr, requiredFieldsMessage(tmpl, missing))
		return validationFailure{}
	}

	if err := runTemplateValidations(celEnv, tmpl, page, oldPage); err != nil {
		return err
	}

	commitMsg := fmt.Sprintf("akb: append %s", relPath)
	if err := store.WriteWithCommitMsg(ctx, fullPath, []byte(fullContent), commitMsg); err != nil {
		return fmt.Errorf("write page: %w", err)
	}

	tags := search.ExtractTags(fm.Fields)
	summary := search.ExtractSummary(fm.Fields)

	// The search index and the link graph describe the same committed page, so
	// both steps share one transaction: a failure in either leaves both at
	// their pre-append state instead of one step behind the other.
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin index transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	searcher := search.NewSQLiteFTS5Searcher(dbConn)
	if err := searcher.IndexPageTx(ctx, tx, relPath, fm.Title, newBody, tags, summary, fm.Type); err != nil {
		return fmt.Errorf("index page: %w", err)
	}

	updater := linkgraph.NewSQLiteLinkGraph(dbConn)
	if err := updater.UpdatePageLinksTx(ctx, tx, relPath, fullContent); err != nil {
		return fmt.Errorf("update links: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit index transaction: %w", err)
	}

	fmt.Printf("Appended to %s\n", relPath)

	return nil
}
