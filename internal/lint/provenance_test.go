package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/markdown"
)

func TestProvenanceChecker(t *testing.T) {
	checker := NewProvenanceChecker()

	if checker.Name() != "provenance" {
		t.Errorf("expected name 'provenance', got %q", checker.Name())
	}

	t.Run("drift exceeds threshold", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "drift.md",
					HasFrontmatter: true,
					Body:           []byte("line one\nline two\nline three\nline four\nline five\n"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Drift",
						Fields: map[string]any{
							"provenance": map[string]any{
								"inferred": 0.3,
							},
						},
					},
					ProvenanceMarkers: []markdown.ProvenanceMarker{},
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}
		if issues[0].Type != "provenance" {
			t.Errorf("expected type 'provenance', got %q", issues[0].Type)
		}
		if issues[0].Severity != "warning" {
			t.Errorf("expected severity 'warning', got %q", issues[0].Severity)
		}
		if issues[0].Path != "drift.md" {
			t.Errorf("expected path 'drift.md', got %q", issues[0].Path)
		}
	})

	t.Run("provenance matches inline markers", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "match.md",
					HasFrontmatter: true,
					Body:           []byte("line one\nline two\nline three\nline four\nline five\n"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Match",
						Fields: map[string]any{
							"provenance": map[string]any{
								"inferred": 0.4,
							},
						},
					},
					ProvenanceMarkers: []markdown.ProvenanceMarker{
						{Type: "inferred", Position: 0},
						{Type: "inferred", Position: 10},
					},
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})

	t.Run("page without provenance field", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "no-prov.md",
					HasFrontmatter: true,
					Body:           []byte("content\n"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "NoProv",
						Fields: map[string]any{},
					},
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})

	t.Run("provenance is not a map", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "bad-prov.md",
					HasFrontmatter: true,
					Body:           []byte("content\n"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "BadProv",
						Fields: map[string]any{
							"provenance": "not-a-map",
						},
					},
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})

	t.Run("non-numeric provenance value", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "non-numeric.md",
					HasFrontmatter: true,
					Body:           []byte("content\n"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "NonNumeric",
						Fields: map[string]any{
							"provenance": map[string]any{
								"inferred": "not-a-number",
							},
						},
					},
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})

	t.Run("zero body lines", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "empty-body.md",
					HasFrontmatter: true,
					Body:           []byte("   \n\n\t\n"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "EmptyBody",
						Fields: map[string]any{
							"provenance": map[string]any{
								"inferred": 0.5,
							},
						},
					},
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})

	t.Run("page without frontmatter", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "no-fm.md",
					HasFrontmatter: false,
					Body:           []byte("content\n"),
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d", len(issues))
		}
	})

	t.Run("integer provenance value", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "int-prov.md",
					HasFrontmatter: true,
					Body:           []byte("line one\nline two\nline three\nline four\nline five\n"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "IntProv",
						Fields: map[string]any{
							"provenance": map[string]any{
								"inferred": 1,
							},
						},
					},
					ProvenanceMarkers: []markdown.ProvenanceMarker{},
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue (drift=1.0 > 0.20), got %d", len(issues))
		}
	})
}
