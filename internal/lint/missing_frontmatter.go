package lint

import (
	"context"
)

type MissingFrontmatterChecker struct{}

func NewMissingFrontmatterChecker() *MissingFrontmatterChecker {
	return &MissingFrontmatterChecker{}
}

func (c *MissingFrontmatterChecker) Name() string {
	return "missing_frontmatter"
}

func (c *MissingFrontmatterChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			issues = append(issues, LintIssue{
				Type:     "missing_frontmatter",
				Message: "page lacks YAML frontmatter delimiters",
				Path:    page.RelPath,
				Severity: "error",
			})
		}
	}
	return issues, nil
}