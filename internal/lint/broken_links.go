// Package lint provides lint checkers for knowledge base validation.
package lint

import (
	"context"
	"fmt"
)

// BrokenLinksChecker detects pages linking to non-existent targets.
//
//nolint:revive // intentionally exported for use by consumers
type BrokenLinksChecker struct{}

// NewBrokenLinksChecker creates a BrokenLinksChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewBrokenLinksChecker() *BrokenLinksChecker {
	return &BrokenLinksChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *BrokenLinksChecker) Name() string {
	return "broken_links"
}

// Check runs the broken links check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *BrokenLinksChecker) Check(ctx context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue

	broken, err := kb.LinkGraph.GetBrokenLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("query broken links: %w", err)
	}
	for _, l := range broken {
		issues = append(issues, LintIssue{
			Type:     "broken_links",
			Message:  fmt.Sprintf("broken link to [[%s]]", l.RawTarget),
			Path:     l.SourcePage,
			Severity: "error",
		})
	}

	ambiguous, err := kb.LinkGraph.GetAmbiguousLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("query ambiguous links: %w", err)
	}
	for _, l := range ambiguous {
		issues = append(issues, LintIssue{
			Type:     "broken_links",
			Message:  fmt.Sprintf("ambiguous link [[%s]] resolves to multiple pages: %s", l.RawTarget, l.ResolvedTo),
			Path:     l.SourcePage,
			Severity: "warning",
		})
	}

	return issues, nil
}
