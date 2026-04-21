package lint

import (
	"context"
	"testing"
	"time"

	"github.com/peedrr/agent-kb/internal/frontmatter"
)

func TestFreshnessChecker(t *testing.T) {
	checker := NewFreshnessChecker()

	if checker.Name() != "freshness" {
		t.Errorf("expected name 'freshness', got %q", checker.Name())
	}

	// fixedNow is 2025-04-15 for deterministic tests
	fixedNow := time.Date(2025, 4, 15, 12, 0, 0, 0, time.UTC)

	t.Run("page updated 90 days ago with confidence high", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}
		ninetyDaysAgo := fixedNow.AddDate(0, 0, -90).Format("2006-01-02")

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "stale.md",
					HasFrontmatter: true,
					Body:           []byte("content"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Stale",
						Fields: map[string]any{
							"updated":    ninetyDaysAgo,
							"confidence": "high",
						},
					},
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}
		if issues[0].Type != "freshness" {
			t.Errorf("expected type 'freshness', got %q", issues[0].Type)
		}
		if issues[0].Severity != "warning" {
			t.Errorf("expected severity 'warning', got %q", issues[0].Severity)
		}
		if issues[0].Path != "stale.md" {
			t.Errorf("expected path 'stale.md', got %q", issues[0].Path)
		}
	})

	t.Run("page updated recently", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}
		recentDate := fixedNow.AddDate(0, 0, -5).Format("2006-01-02")

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "fresh.md",
					HasFrontmatter: true,
					Body:           []byte("content"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Fresh",
						Fields: map[string]any{
							"updated": recentDate,
						},
					},
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})

	t.Run("fallback to created field", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}
		ninetyDaysAgo := fixedNow.AddDate(0, 0, -90).Format("2006-01-02")

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "no-updated.md",
					HasFrontmatter: true,
					Body:           []byte("content"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "NoUpdated",
						Fields: map[string]any{
							"created":    ninetyDaysAgo,
							"confidence": "high",
						},
					},
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue (fallback to created), got %d", len(issues))
		}
	})

	t.Run("no updated or created field", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "no-dates.md",
					HasFrontmatter: true,
					Body:           []byte("content"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "NoDates",
						Fields: map[string]any{},
					},
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})

	t.Run("future date skipped", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}
		futureDate := fixedNow.AddDate(0, 0, 10).Format("2006-01-02")

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "future.md",
					HasFrontmatter: true,
					Body:           []byte("content"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Future",
						Fields: map[string]any{
							"updated": futureDate,
						},
					},
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues for future date, got %d", len(issues))
		}
	})

	t.Run("unparseable date skipped", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "bad-date.md",
					HasFrontmatter: true,
					Body:           []byte("content"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "BadDate",
						Fields: map[string]any{
							"updated": "not-a-date",
						},
					},
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues for unparseable date, got %d", len(issues))
		}
	})

	t.Run("confidence low makes page staler faster", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}
		// 30 days ago: with high confidence score = 100 * 2^(-1) * 1.0 = 50 (exactly threshold, not < 50)
		// with low confidence score = 100 * 2^(-1) * 0.4 = 20 < 50 → stale
		thirtyDaysAgo := fixedNow.AddDate(0, 0, -30).Format("2006-01-02")

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "low-conf.md",
					HasFrontmatter: true,
					Body:           []byte("content"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "LowConf",
						Fields: map[string]any{
							"updated":    thirtyDaysAgo,
							"confidence": "low",
						},
					},
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue (low confidence = staler), got %d", len(issues))
		}
	})

	t.Run("RFC3339 timestamp parsed", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}
		ninetyDaysAgo := fixedNow.AddDate(0, 0, -90).Format(time.RFC3339)

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "rfc3339.md",
					HasFrontmatter: true,
					Body:           []byte("content"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "RFC3339",
						Fields: map[string]any{
							"updated":    ninetyDaysAgo,
							"confidence": "high",
						},
					},
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue for RFC3339 date, got %d", len(issues))
		}
	})

	t.Run("page without frontmatter skipped", func(t *testing.T) {
		c := &FreshnessChecker{nowFunc: func() time.Time { return fixedNow }}

		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "no-fm.md",
					HasFrontmatter: false,
					Body:           []byte("content"),
				},
			},
		}

		issues, err := c.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})
}
