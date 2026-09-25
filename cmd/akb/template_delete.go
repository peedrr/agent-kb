// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

var tdForce bool

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a template",
	Long:  `Delete a template and its associated mockup files. Requires --force when pages use the template type.`,
	Example: `  # Delete a template (refused if pages use it)
  akb template delete adr

  # Force delete even if pages use the template type
  akb template delete adr --force

  # Delete without creating a git commit
  akb template delete adr --force --no-commit`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateDelete,
}

func init() {
	templateDeleteCmd.Flags().BoolVar(&tdForce, "force", false, "force deletion even if pages use this template type")
	templateCmd.AddCommand(templateDeleteCmd)
}

var templateNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// templateCommitPaths returns the KB-relative paths of the files a template
// operation on name records: the template and its pass and fail mockups.
func templateCommitPaths(name string) []string {
	return []string{
		filepath.Join(".akb", "templates", name+".yaml"),
		filepath.Join(".akb", "templates", name+"_pass.md"),
		filepath.Join(".akb", "templates", name+"_fail.md"),
	}
}

func runTemplateDelete(_ *cobra.Command, args []string) error {
	name := args[0]

	if !templateNameRe.MatchString(name) {
		return fmt.Errorf("invalid template name %q: must contain only letters, numbers, hyphens, and underscores", name)
	}

	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	tmplPath := filepath.Join(kbRoot, ".akb", "templates", name+".yaml")
	if err := path.AssertContained(kbRoot, tmplPath); err != nil {
		return fmt.Errorf("resolve template path: %w", err)
	}
	if _, err := os.Stat(tmplPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template %q not found", name)
		}
		return fmt.Errorf("stat template: %w", err)
	}

	sqlDB, err := db.OpenKB(kbRoot)
	if err != nil {
		if isMissingDB(err) {
			return fmt.Errorf("run `akb init` to initialize the knowledge base")
		}
		return fmt.Errorf("open search database: %w", err)
	}
	defer sqlDB.Close() //nolint:errcheck // DB close error non-critical on command exit

	var count int
	row := sqlDB.QueryRow("SELECT COUNT(*) FROM documents WHERE type = ?", name)
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("count pages: %w", err)
	}

	if !tdForce {
		fmt.Printf("Template %q is used by %d pages. Run `akb template delete %s --force` to delete.\n", name, count, name)
		return &usageError{msg: "deletion refused without --force"}
	}

	// Hold the repository lock across the removal of the template files and the
	// commit that records them; the lint sweep that follows only reads the KB.
	removeErr := func() error {
		repoLock, lockErr := storage.LockRepo(kbRoot)
		if lockErr != nil {
			return fmt.Errorf("lock repository: %w", lockErr)
		}
		defer repoLock.Release()

		for _, relPath := range templateCommitPaths(name) {
			removalPath := filepath.Join(kbRoot, relPath)
			if err := path.AssertContained(kbRoot, removalPath); err != nil {
				return fmt.Errorf("resolve template path: %w", err)
			}
			if err := os.Remove(removalPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("delete file %s: %w", filepath.Base(relPath), err)
			}
		}

		if noCommit {
			return nil
		}
		commitMsg := fmt.Sprintf("akb: template delete %s", name)
		if err := storage.CommitFiles(kbRoot, commitMsg, templateCommitPaths(name)...); err != nil {
			return fmt.Errorf("commit template delete: %w", err)
		}
		return nil
	}()
	if removeErr != nil {
		return removeErr
	}

	ctx := context.Background()
	report, err := RunLint(ctx)
	if err != nil {
		fmt.Printf("Template %q deleted. %d pages now have orphan type.\n", name, count)
		return fmt.Errorf("lint: %w", err)
	}

	if err := printLintText(report); err != nil {
		fmt.Printf("Template %q deleted. %d pages now have orphan type.\n", name, count)
		return err
	}

	fmt.Printf("Template %q deleted. %d pages now have orphan type.\n", name, count)

	var errorCount int
	for _, issue := range report.Issues {
		if issue.Severity == "error" {
			errorCount++
		}
	}
	if errorCount > 0 {
		return errLintIssues
	}

	return nil
}
