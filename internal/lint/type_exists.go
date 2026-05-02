package lint

import (
	"context"
	"fmt"
)

// TypeExistsChecker detects pages using types that don't have a template.
//
//nolint:revive // intentionally exported for use by consumers
type TypeExistsChecker struct{}

// NewTypeExistsChecker creates a TypeExistsChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewTypeExistsChecker() *TypeExistsChecker {
	return &TypeExistsChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *TypeExistsChecker) Name() string {
	return "type_exists"
}

// Check runs the type exists check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *TypeExistsChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}
		fm := page.Frontmatter
		if fm.Type == "" {
			continue
		}
		if _, ok := kb.Templates[fm.Type]; !ok {
			issues = append(issues, LintIssue{
				Type:     "type_exists",
				RuleID:   "type_exists",
				Message:  fmt.Sprintf("unknown type '%s'. Create .akb/templates/%s.yaml first.", fm.Type, fm.Type),
				Path:     page.RelPath,
				Severity: "error",
			})
		}
	}
	return issues, nil
}
