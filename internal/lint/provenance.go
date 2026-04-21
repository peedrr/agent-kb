package lint

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/peedrr/agent-kb/internal/markdown"
)

type ProvenanceChecker struct{}

func NewProvenanceChecker() *ProvenanceChecker {
	return &ProvenanceChecker{}
}

func (c *ProvenanceChecker) Name() string {
	return "provenance"
}

func (c *ProvenanceChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}
		provenanceVal, ok := page.Frontmatter.Fields["provenance"]
		if !ok || provenanceVal == nil {
			continue
		}

		provMap, ok := provenanceVal.(map[string]any)
		if !ok {
			continue
		}

		totalLines := countNonEmptyLines(string(page.Body))
		if totalLines == 0 {
			continue
		}

		markerCounts := markdown.CountMarkersByType(page.ProvenanceMarkers)

		for markerType, declaredVal := range provMap {
			frontmatterConfidence, ok := toFloat64(declaredVal)
			if !ok {
				continue
			}

			inlineCount := markerCounts[markerType]
			inlineRatio := float64(inlineCount) / float64(totalLines)
			drift := math.Abs(frontmatterConfidence - inlineRatio)

			if drift > ProvenanceDriftThreshold {
				issues = append(issues, LintIssue{
					Type:     "provenance",
					Message:  fmt.Sprintf("provenance drift for '%s': declared %.2f but actual ratio is %.2f (drift: %.2f)", markerType, frontmatterConfidence, inlineRatio, drift),
					Path:     page.RelPath,
					Severity: "warning",
				})
			}
		}
	}
	return issues, nil
}

func countNonEmptyLines(s string) int {
	count := 0
	for line := range strings.SplitSeq(s, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
