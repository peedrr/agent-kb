package lint

import (
	"context"
	"strings"
)

// EmptyPagesChecker detects pages with no body content after frontmatter.
//
//nolint:revive // intentionally exported for use by consumers
type EmptyPagesChecker struct{}

// NewEmptyPagesChecker creates an EmptyPagesChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewEmptyPagesChecker() *EmptyPagesChecker {
	return &EmptyPagesChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *EmptyPagesChecker) Name() string {
	return "empty_pages"
}

// Check runs the empty pages check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *EmptyPagesChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}
		if len(strings.TrimSpace(string(page.Body))) == 0 {
			issues = append(issues, LintIssue{
				Type:     "empty_pages",
				Message:  "page has no content after frontmatter",
				Path:     page.RelPath,
				Severity: "warning",
			})
		}
	}
	return issues, nil
}
