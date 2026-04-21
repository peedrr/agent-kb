package lint

import (
	"context"
	"fmt"
)

type TypeExistsChecker struct{}

func NewTypeExistsChecker() *TypeExistsChecker {
	return &TypeExistsChecker{}
}

func (c *TypeExistsChecker) Name() string {
	return "type_exists"
}

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
				Message:  fmt.Sprintf("unknown type '%s'. Create .akb/templates/%s.yaml first.", fm.Type, fm.Type),
				Path:     page.RelPath,
				Severity: "error",
			})
		}
	}
	return issues, nil
}
