package lint

import (
	"context"
	"fmt"
)

// OrphansChecker validates Orphans issues.
type OrphansChecker struct{}

// NewOrphansChecker creates a new OrphansChecker.
func NewOrphansChecker() *OrphansChecker {
	return &OrphansChecker{}
}

// Name returns the checker name.
func (c *OrphansChecker) Name() string {
	return "orphans"
}

// Check runs the checker and returns issues.
func (c *OrphansChecker) Check(ctx context.Context, kb *KB) ([]LintIssue, error) {
	paths, err := kb.LinkGraph.GetOrphans(ctx)
	if err != nil {
		return nil, fmt.Errorf("query orphans: %w", err)
	}

	var issues []LintIssue
	for _, p := range paths {
		issues = append(issues, LintIssue{
			Type:     "orphans",
			Message:  "page has no inbound links",
			Path:     p,
			Severity: "warning",
		})
	}

	return issues, nil
}
