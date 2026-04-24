package lint

import (
	"context"
	"fmt"
	"strings"
)

// CategoryDirsChecker checks that typed pages live in their expected directories.
//
//nolint:revive // intentionally exported for use by consumers
type CategoryDirsChecker struct{}

// NewCategoryDirsChecker creates a CategoryDirsChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewCategoryDirsChecker() *CategoryDirsChecker {
	return &CategoryDirsChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *CategoryDirsChecker) Name() string {
	return "category_dirs"
}

// Check runs the category directory check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *CategoryDirsChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter || page.Frontmatter.Type == "" {
			continue
		}
		tmpl, ok := kb.Templates[page.Frontmatter.Type]
		if !ok {
			continue
		}

		parts := strings.Split(page.RelPath, "/")
		var actualDir string
		if len(parts) >= 3 {
			actualDir = parts[1] // "kb/<dir>/file.md" → parts[1] is the directory
		}

		expectedDir := tmpl.Dir
		if actualDir != expectedDir {
			issues = append(issues, LintIssue{
				Type:     "category_dirs",
				Message:  fmt.Sprintf("page of type '%s' should be in kb/%s/ but is in kb/%s/", page.Frontmatter.Type, expectedDir, actualDir),
				Path:     page.RelPath,
				Severity: "warning",
			})
		}
	}
	return issues, nil
}
