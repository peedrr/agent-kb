package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/path"
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

func runTemplateDelete(_ *cobra.Command, args []string) error {
	name := args[0]

	if !templateNameRe.MatchString(name) {
		return fmt.Errorf("invalid template name %q: must contain only letters, numbers, hyphens, and underscores", name)
	}

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	tmplPath := filepath.Join(kbRoot, ".akb", "templates", name+".yaml")
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
		return fmt.Errorf("deletion refused without --force")
	}

	files := []string{
		filepath.Join(kbRoot, ".akb", "templates", name+".yaml"),
		filepath.Join(kbRoot, ".akb", "templates", name+"_pass.md"),
		filepath.Join(kbRoot, ".akb", "templates", name+"_fail.md"),
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("delete file %s: %w", filepath.Base(f), err)
		}
	}

	if !noCommit {
		gitAdd := exec.Command("git", "add", "-A", ".akb/templates/") //nolint:gosec // launching trusted git binary with controlled args
		gitAdd.Dir = kbRoot
		if out, err := gitAdd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add: %s: %w", strings.TrimSpace(string(out)), err)
		}

		if err := ensureGitConfig(kbRoot); err != nil {
			return fmt.Errorf("git config: %w", err)
		}

		commitMsg := fmt.Sprintf("akb: template delete %s", name)
		gitCommit := exec.Command("git", "commit", "-m", commitMsg) //nolint:gosec // launching trusted git binary with controlled args
		gitCommit.Dir = kbRoot
		if out, err := gitCommit.CombinedOutput(); err != nil {
			return fmt.Errorf("git commit: %s: %w", strings.TrimSpace(string(out)), err)
		}
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
