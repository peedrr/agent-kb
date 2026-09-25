// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"fmt"

	"github.com/peedrr/agent-kb/internal/index"
)

// IndexConsistencyChecker detects pages in index.md that don't exist and vice versa.
//
//nolint:revive // intentionally exported for use by consumers
type IndexConsistencyChecker struct{}

// NewIndexConsistencyChecker creates an IndexConsistencyChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewIndexConsistencyChecker() *IndexConsistencyChecker {
	return &IndexConsistencyChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *IndexConsistencyChecker) Name() string {
	return "index_consistency"
}

// Check runs the index consistency check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *IndexConsistencyChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue

	entries, err := index.ReadIndex(kb.Root)
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}

	indexPaths := make(map[string]bool)
	for _, e := range entries {
		indexPaths[e.Path] = true
	}

	pagePaths := make(map[string]bool)
	for _, p := range kb.Pages {
		pagePaths[p.RelPath] = true
	}

	for path := range indexPaths {
		if !pagePaths[path] {
			issues = append(issues, LintIssue{
				Type:     "index_consistency",
				RuleID:   "index_consistency",
				Message:  "page listed in index.md but file does not exist",
				Path:     path,
				Severity: "error",
			})
		}
	}

	for path := range pagePaths {
		if !indexPaths[path] {
			issues = append(issues, LintIssue{
				Type:     "index_consistency",
				RuleID:   "index_consistency",
				Message:  "page exists but is missing from index.md",
				Path:     path,
				Severity: "warning",
			})
		}
	}

	return issues, nil
}
