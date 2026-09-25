// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"fmt"
	"time"

	"github.com/google/cel-go/common/types"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"

	"github.com/peedrr/agent-kb/internal/cel"
)

// CELLintChecker evaluates CEL lint_rules from templates against pages.
//
//nolint:revive // intentionally exported for use by consumers
type CELLintChecker struct{}

// NewCELLintChecker creates a new CELLintChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewCELLintChecker() *CELLintChecker {
	return &CELLintChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *CELLintChecker) Name() string {
	return "cel_lint"
}

// Check evaluates CEL lint rules for each page against its template.
//
//nolint:revive // intentionally exported for use by consumers
func (c *CELLintChecker) Check(ctx context.Context, kb *KB) ([]LintIssue, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return nil, fmt.Errorf("create CEL environment: %w", err)
	}

	var issues []LintIssue
	now := time.Now()

	for _, page := range kb.Pages {
		if !page.HasFrontmatter || page.Frontmatter.Type == "" {
			continue
		}

		tmpl, ok := kb.Templates[page.Frontmatter.Type]
		if !ok {
			continue
		}

		md := goldmark.New()
		doc := md.Parser().Parse(text.NewReader(page.Body))
		pageMap := cel.BuildPage(page.RelPath, page.Frontmatter, page.Body, doc, page.Body)

		for _, rule := range tmpl.LintRules {
			prg, err := cel.CompileRule(env, rule.Rule)
			if err != nil {
				return nil, fmt.Errorf("compile lint rule %s: %w", rule.ID, err)
			}

			result, err := cel.Evaluate(ctx, prg, map[string]any{
				"page": pageMap,
				"now":  now,
			})
			if err != nil {
				// A rule that cannot be evaluated is reported against the page it
				// failed on; the sweep continues with the remaining rules and pages
				// so one broken rule does not hide every other finding.
				issues = append(issues, LintIssue{
					Type:     "cel_lint",
					RuleID:   rule.ID,
					Message:  fmt.Sprintf("rule evaluation error: %v", err),
					Path:     page.RelPath,
					Severity: "error",
				})
				continue
			}

			if result != types.True {
				issues = append(issues, LintIssue{
					Type:     "cel_lint",
					RuleID:   rule.ID,
					Message:  rule.Expect,
					Path:     page.RelPath,
					Severity: rule.Severity,
				})
			}
		}
	}

	return issues, nil
}
