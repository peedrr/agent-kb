package lint

import (
	"context"
	"strings"
)

type EmptyPagesChecker struct{}

func NewEmptyPagesChecker() *EmptyPagesChecker {
	return &EmptyPagesChecker{}
}

func (c *EmptyPagesChecker) Name() string {
	return "empty_pages"
}

func (c *EmptyPagesChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}
		if len(strings.TrimSpace(string(page.Body))) == 0 {
			issues = append(issues, LintIssue{
				Type:     "empty_pages",
				Message: "page has no content after frontmatter",
				Path:    page.RelPath,
				Severity: "warning",
			})
		}
	}
	return issues, nil
}