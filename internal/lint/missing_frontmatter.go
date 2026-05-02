package lint

import (
	"context"
)

// MissingFrontmatterChecker detects pages lacking YAML frontmatter.
//
//nolint:revive // intentionally exported for use by consumers
type MissingFrontmatterChecker struct{}

// NewMissingFrontmatterChecker creates a MissingFrontmatterChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewMissingFrontmatterChecker() *MissingFrontmatterChecker {
	return &MissingFrontmatterChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *MissingFrontmatterChecker) Name() string {
	return "missing_frontmatter"
}

// Check runs the missing frontmatter check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *MissingFrontmatterChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			issues = append(issues, LintIssue{
				Type:     "missing_frontmatter",
				RuleID:   "missing_frontmatter",
				Message:  "page lacks YAML frontmatter delimiters",
				Path:     page.RelPath,
				Severity: "error",
			})
		}
	}
	return issues, nil
}
