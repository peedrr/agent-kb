package lint

import (
	"context"
	"fmt"
)

// ConfidenceChecker validates Confidence issues.
type ConfidenceChecker struct{}

// NewConfidenceChecker creates a new ConfidenceChecker.
func NewConfidenceChecker() *ConfidenceChecker {
	return &ConfidenceChecker{}
}

// Name returns the checker name.
func (c *ConfidenceChecker) Name() string {
	return "confidence"
}

var validConfidenceLevels = map[string]bool{
	"high":   true,
	"medium": true,
	"low":    true,
}

// Check runs the checker and returns issues.
func (c *ConfidenceChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}
		confidenceVal, ok := page.Frontmatter.Fields["confidence"]
		if !ok || confidenceVal == nil {
			continue
		}

		strVal, ok := confidenceVal.(string)
		if !ok {
			issues = append(issues, LintIssue{
				Type:     "confidence",
				RuleID:   "confidence",
				Message:  fmt.Sprintf("confidence '%v' is not valid; must be one of: high, medium, low", confidenceVal),
				Path:     page.RelPath,
				Severity: "error",
			})
			continue
		}

		if !validConfidenceLevels[strVal] {
			issues = append(issues, LintIssue{
				Type:     "confidence",
				RuleID:   "confidence",
				Message:  fmt.Sprintf("confidence '%s' is not valid; must be one of: high, medium, low", strVal),
				Path:     page.RelPath,
				Severity: "error",
			})
			continue
		}

		if strVal == "low" && len(page.ProvenanceMarkers) == 0 {
			issues = append(issues, LintIssue{
				Type:     "confidence",
				RuleID:   "confidence",
				Message:  "low confidence but no provenance markers; add ^[inferred] or ^[ambiguous] markers to support the confidence rating",
				Path:     page.RelPath,
				Severity: "warning",
			})
		}
	}
	return issues, nil
}
