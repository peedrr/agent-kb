package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/lint"
	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/markdown"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/template"
	"github.com/spf13/cobra"
)

var lintJSON bool

var errLintIssues = errors.New("lint issues found")

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate knowledge base pages",
	Long:  `Run lint checks on KB pages for validation errors, schema violations, and quality issues.`,
	Args:  cobra.NoArgs,
	RunE:  runLint,
}

func init() {
	lintCmd.Flags().BoolVar(&lintJSON, "json", false, "output results as JSON")
}

func runLint(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return err
	}

	sqlDB, err := db.OpenKB(kbRoot)
	if err != nil {
		if isMissingDB(err) {
			return fmt.Errorf("run `akb init` to initialize the knowledge base")
		}
		return err
	}
	defer sqlDB.Close()

	templates, err := template.LoadTemplates(filepath.Join(kbRoot, ".akb", "templates"))
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	entries, err := manifest.NewManager(kbRoot).ReadManifest()
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}

	lg := linkgraph.NewSQLiteLinkGraph(sqlDB)

	kbDir := filepath.Join(kbRoot, "kb")
	var pages []lint.PageData

	err = filepath.WalkDir(kbDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		base := filepath.Base(path)
		if base == "index.md" || base == "log.md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(kbRoot, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		fm, body, err := frontmatter.Parse(content)
		hasFrontmatter := err == nil && fm != nil
		if !hasFrontmatter {
			body = content
		}

		var markers []markdown.ProvenanceMarker
		var annotations []markdown.Annotation
		if hasFrontmatter {
			markers = markdown.ParseProvenanceMarkers(string(body))
			annotations = markdown.ParseAnnotations(string(content))
		}

		pages = append(pages, lint.PageData{
			RelPath:           relPath,
			Content:           content,
			Body:              body,
			Frontmatter:       fm,
			ProvenanceMarkers: markers,
			Annotations:       annotations,
			HasFrontmatter:    hasFrontmatter,
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("walk kb directory: %w", err)
	}

	kb := &lint.KB{
		Root:      kbRoot,
		LinkGraph: lg,
		Templates: templates,
		Manifest:  entries,
		Pages:     pages,
	}

	engine := lint.NewLintEngine()
	engine.AddChecker(lint.NewBrokenLinksChecker())
	engine.AddChecker(lint.NewOrphansChecker())
	engine.AddChecker(lint.NewEmptyPagesChecker())
	engine.AddChecker(lint.NewMissingFrontmatterChecker())
	engine.AddChecker(lint.NewIndexConsistencyChecker())
	engine.AddChecker(lint.NewTypeExistsChecker())
	engine.AddChecker(lint.NewFrontmatterSchemaChecker())
	engine.AddChecker(lint.NewCategoryDirsChecker())
	engine.AddChecker(lint.NewCitationsChecker())
	engine.AddChecker(lint.NewConfidenceChecker())
	engine.AddChecker(lint.NewSummaryLengthChecker())
	engine.AddChecker(lint.NewProvenanceChecker())
	engine.AddChecker(lint.NewFreshnessChecker())

	ctx := context.Background()
	report, err := engine.Run(ctx, kb)
	if err != nil {
		return fmt.Errorf("lint: %w", err)
	}

	if lintJSON {
		err = printLintJSON(report)
	} else {
		err = printLintText(report)
	}
	if err != nil {
		if errors.Is(err, errLintIssues) {
			os.Exit(1)
		}
		return err
	}
	return nil
}

func printLintText(report *lint.LintReport) error {
	if len(report.Issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	byPath := make(map[string][]lint.LintIssue)
	for _, issue := range report.Issues {
		byPath[issue.Path] = append(byPath[issue.Path], issue)
	}

	for path, issues := range byPath {
		fmt.Printf("=== %s ===\n", path)
		for _, issue := range issues {
			fmt.Printf("  [%s] %s: %s\n", issue.Type, issue.Severity, issue.Message)
		}
		fmt.Println()
	}

	fmt.Printf("%d issue(s) found across %d page(s).\n", len(report.Issues), report.PagesChecked)

	return errLintIssues
}

func printLintJSON(report *lint.LintReport) error {
	type jsonOutput struct {
		Issues       []lint.LintIssue `json:"issues"`
		Summary      struct {
			Total        int            `json:"total"`
			PagesChecked int            `json:"pages_checked"`
			ByCheck      map[string]int `json:"by_check"`
		} `json:"summary"`
	}

	var out jsonOutput
	out.Issues = report.Issues
	out.Summary.Total = len(report.Issues)
	out.Summary.PagesChecked = report.PagesChecked
	out.Summary.ByCheck = report.ByCheck

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return err
	}

	if len(report.Issues) > 0 {
		return errLintIssues
	}
	return nil
}