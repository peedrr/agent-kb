package lint

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

// RequiredFieldsChecker detects pages that leave a schema-required frontmatter
// field unset. Write-time validation covers only the pages that passed
// `akb write`, so a page pulled from git, planted by hand, or written before
// its template declared a field stays invisible without a sweep.
//
//nolint:revive // intentionally exported for use by consumers
type RequiredFieldsChecker struct{}

// NewRequiredFieldsChecker creates a RequiredFieldsChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewRequiredFieldsChecker() *RequiredFieldsChecker {
	return &RequiredFieldsChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *RequiredFieldsChecker) Name() string {
	return "required_fields"
}

// Check runs the required fields check. Pages without frontmatter and pages
// whose type has no template are skipped: neither has a schema to check
// against, and missing_frontmatter and type_orphan report them.
//
//nolint:revive // intentionally exported for use by consumers
func (c *RequiredFieldsChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue

	for _, page := range kb.Pages {
		if !page.HasFrontmatter || page.Frontmatter.Type == "" {
			continue
		}

		tmpl, ok := kb.Templates[page.Frontmatter.Type]
		if !ok {
			continue
		}

		missing := missingRequiredFields(tmpl, page.Frontmatter)
		if len(missing) == 0 {
			continue
		}

		issues = append(issues, LintIssue{
			Type:   "required_fields",
			RuleID: "required_fields",
			// The wording matches the write path's required-field refusal, so a
			// page reads the same here and in a rejected write.
			Message: fmt.Sprintf("missing required frontmatter field(s): %s (declared required by template %q)",
				strings.Join(missing, ", "), tmpl.Name),
			Path:     page.RelPath,
			Severity: "error",
		})
	}

	return issues, nil
}

// missingRequiredFields returns the fields the template marks required that the
// page leaves unset, sorted by name. It mirrors the write path's check
// (cmd/akb/required_fields.go): presence only, so a value is never inspected and
// type and enum constraints stay with the template's CEL rules.
func missingRequiredFields(tmpl template.Template, fm *frontmatter.ParsedFrontmatter) []string {
	var missing []string
	for key, field := range tmpl.Schema.Frontmatter {
		if !field.Required || frontmatterFieldPresent(fm, key) {
			continue
		}
		missing = append(missing, key)
	}
	sort.Strings(missing)
	return missing
}

// frontmatterFieldPresent reports whether the parsed frontmatter sets key. It
// mirrors the write path's frontmatterKeyPresent: Parse routes type and title
// into dedicated fields instead of Fields, so both are checked before the Fields
// lookup.
func frontmatterFieldPresent(fm *frontmatter.ParsedFrontmatter, key string) bool {
	switch key {
	case "type":
		return fm.Type != ""
	case "title":
		return fm.Title != ""
	}
	_, ok := fm.Fields[key]
	return ok
}
