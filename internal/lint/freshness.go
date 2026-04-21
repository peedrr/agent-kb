package lint

import (
	"context"
	"fmt"
	"math"
	"time"
)

type FreshnessChecker struct {
	nowFunc func() time.Time
}

func NewFreshnessChecker() *FreshnessChecker {
	return &FreshnessChecker{nowFunc: time.Now}
}

func (c *FreshnessChecker) Name() string {
	return "freshness"
}

func (c *FreshnessChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue
	now := c.nowFunc()
	for _, page := range kb.Pages {
		if !page.HasFrontmatter {
			continue
		}

		tsStr, ok := getStringField(page.Frontmatter.Fields, "updated")
		if !ok {
			tsStr, ok = getStringField(page.Frontmatter.Fields, "created")
			if !ok {
				continue
			}
		}

		parsedTime, err := parseTimestamp(tsStr)
		if err != nil {
			continue
		}

		if parsedTime.After(now) {
			continue
		}

		daysSince := now.Sub(parsedTime).Hours() / 24

		confidenceWeight := 0.7
		if confStr, ok := getStringField(page.Frontmatter.Fields, "confidence"); ok {
			switch confStr {
			case "high":
				confidenceWeight = 1.0
			case "medium":
				confidenceWeight = 0.7
			case "low":
				confidenceWeight = 0.4
			}
		}

		score := 100.0 * math.Pow(2, -daysSince/float64(FreshnessHalfLifeDays)) * confidenceWeight

		if score < FreshnessScoreThreshold {
			issues = append(issues, LintIssue{
				Type:     "freshness",
				Message:  fmt.Sprintf("page is stale (freshness score: %.1f, threshold: %.1f). Last updated %.0f days ago.", score, FreshnessScoreThreshold, daysSince),
				Path:     page.RelPath,
				Severity: "warning",
			})
		}
	}
	return issues, nil
}

func getStringField(fields map[string]any, key string) (string, bool) {
	val, ok := fields[key]
	if !ok || val == nil {
		return "", false
	}
	str, ok := val.(string)
	if !ok {
		return "", false
	}
	return str, true
}

func parseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unparseable date: %s", s)
}
