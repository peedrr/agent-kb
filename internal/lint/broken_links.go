package lint

import (
	"context"
	"fmt"
)

type BrokenLinksChecker struct{}

func NewBrokenLinksChecker() *BrokenLinksChecker {
	return &BrokenLinksChecker{}
}

func (c *BrokenLinksChecker) Name() string {
	return "broken_links"
}

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