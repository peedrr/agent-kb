// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"fmt"
)

// CitationsChecker validates Citations issues.
type CitationsChecker struct{}

// NewCitationsChecker creates a new CitationsChecker.
func NewCitationsChecker() *CitationsChecker {
	return &CitationsChecker{}
}

// Name returns the checker name.
func (c *CitationsChecker) Name() string {
	return "citations"
}

// Check runs the checker and returns issues.
func (c *CitationsChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	manifestFiles := make(map[string]bool)
	for _, e := range kb.Manifest {
		manifestFiles[e.Filename] = true
	}

	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}
		sources, ok := page.Frontmatter.Fields["sources"]
		if !ok || sources == nil {
			continue
		}

		var filenames []string
		switch s := sources.(type) {
		case string:
			filenames = []string{s}
		case []any:
			for _, item := range s {
				if str, ok := item.(string); ok {
					filenames = append(filenames, str)
				}
			}
		default:
			continue
		}

		for _, fn := range filenames {
			if !manifestFiles[fn] {
				issues = append(issues, LintIssue{
					Type:     "citations",
					RuleID:   "citations",
					Message:  fmt.Sprintf("source '%s' not found in raw manifest; run 'akb raw sync' to update", fn),
					Path:     page.RelPath,
					Severity: "error",
				})
			}
		}
	}
	return issues, nil
}
