package lint

import (
	"context"
	"fmt"
)

// SummaryLengthChecker detects summaries that are too short or too long.
//
//nolint:revive // intentionally exported for use by consumers
type SummaryLengthChecker struct{}

// NewSummaryLengthChecker creates a SummaryLengthChecker.
//
//nolint:revive // intentionally exported for use by consumers
func NewSummaryLengthChecker() *SummaryLengthChecker {
	return &SummaryLengthChecker{}
}

// Name returns the checker name.
//
//nolint:revive // intentionally exported for use by consumers
func (c *SummaryLengthChecker) Name() string {
	return "summary_length"
}

// Check runs the summary length check.
//
//nolint:revive // intentionally exported for use by consumers
func (c *SummaryLengthChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}
		summaryVal, ok := page.Frontmatter.Fields["summary"]
		if !ok || summaryVal == nil {
			continue
		}

		summary, ok := summaryVal.(string)
		if !ok {
			continue
		}

		if len(summary) < SummaryMinLength {
			issues = append(issues, LintIssue{
				Type:     "summary_length",
				Message:  fmt.Sprintf("summary is too short (%d chars, minimum is %d)", len(summary), SummaryMinLength),
				Path:     page.RelPath,
				Severity: "warning",
			})
		}
		if len(summary) > SummaryMaxLength {
			issues = append(issues, LintIssue{
				Type:     "summary_length",
				Message:  fmt.Sprintf("summary is too long (%d chars, maximum is %d)", len(summary), SummaryMaxLength),
				Path:     page.RelPath,
				Severity: "warning",
			})
		}
	}
	return issues, nil
}
