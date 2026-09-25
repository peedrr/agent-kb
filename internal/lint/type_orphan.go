// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"fmt"
)

// TypeOrphanChecker detects pages whose frontmatter type has no matching template.
//
//nolint:revive // intentionally exported for use by consumers
type TypeOrphanChecker struct{}

// NewTypeOrphanChecker creates a TypeOrphanChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewTypeOrphanChecker() *TypeOrphanChecker {
	return &TypeOrphanChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *TypeOrphanChecker) Name() string {
	return "type_orphan"
}

// Check runs the type orphan check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *TypeOrphanChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue

	for _, page := range kb.Pages {
		if !page.HasFrontmatter || page.Frontmatter.Type == "" {
			continue
		}

		if _, ok := kb.Templates[page.Frontmatter.Type]; !ok {
			issues = append(issues, LintIssue{
				Type:     "type_orphan",
				RuleID:   "type_orphan",
				Message:  fmt.Sprintf("page has type '%s' but no template '%s' exists", page.Frontmatter.Type, page.Frontmatter.Type),
				Path:     page.RelPath,
				Severity: "error",
			})
		}
	}

	return issues, nil
}
