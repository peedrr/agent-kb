package lint

import (
	"context"
	"fmt"
	"strings"
)

// FrontmatterSchemaChecker validates pages have required fields for their type.
//
//nolint:revive // intentionally exported for use by consumers
type FrontmatterSchemaChecker struct{}

// NewFrontmatterSchemaChecker creates a FrontmatterSchemaChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewFrontmatterSchemaChecker() *FrontmatterSchemaChecker {
	return &FrontmatterSchemaChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *FrontmatterSchemaChecker) Name() string {
	return "frontmatter_schema"
}

// Check runs the frontmatter schema check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *FrontmatterSchemaChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}
		fm := page.Frontmatter
		if fm.Type == "" {
			continue
		}
		tmpl, ok := kb.Templates[fm.Type]
		if !ok {
			continue
		}

		for fieldName, field := range tmpl.Schema.Frontmatter {
			var value any
			var exists bool
			switch fieldName {
			case "type":
				value = fm.Type
				exists = fm.Type != ""
			case "title":
				value = fm.Title
				exists = fm.Title != ""
			default:
				value, exists = fm.Fields[fieldName]
			}

			if field.Required && !exists {
				issues = append(issues, LintIssue{
					Type:     "frontmatter_schema",
					RuleID:   "frontmatter_schema",
					Message:  fmt.Sprintf("missing required field '%s' for type '%s'", fieldName, fm.Type),
					Path:     page.RelPath,
					Severity: "error",
				})
				continue
			}

			if !exists {
				continue
			}

			if len(field.Enum) > 0 {
				strVal, ok := value.(string)
if !ok {
				issues = append(issues, LintIssue{
					Type:     "frontmatter_schema",
					RuleID:   "frontmatter_schema",
					Message:  fmt.Sprintf("field '%s' must be a string for type '%s'", fieldName, fm.Type),
					Path:     page.RelPath,
					Severity: "error",
				})
				continue
			}
if !isInEnum(strVal, field.Enum) {
				issues = append(issues, LintIssue{
					Type:     "frontmatter_schema",
					RuleID:   "frontmatter_schema",
					Message:  fmt.Sprintf("field '%s' value '%s' is not valid; must be one of: %s", fieldName, strVal, strings.Join(field.Enum, ", ")),
					Path:     page.RelPath,
					Severity: "error",
				})
			}
			}
		}
	}
	return issues, nil
}

func isInEnum(val string, enum []string) bool {
	for _, e := range enum {
		if val == e {
			return true
		}
	}
	return false
}
