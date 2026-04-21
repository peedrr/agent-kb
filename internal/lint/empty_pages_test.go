package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
)

func TestEmptyPagesChecker(t *testing.T) {
	checker := NewEmptyPagesChecker()

	if checker.Name() != "empty_pages" {
		t.Errorf("expected name 'empty_pages', got %q", checker.Name())
	}

	t.Run("page with frontmatter but empty body", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "test.md",
					HasFrontmatter: true,
					Body:           []byte("   \n\t \n"),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Test",
					},
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

		if issues[0].Type != "empty_pages" {
			t.Errorf("expected issue type 'empty_pages', got %q", issues[0].Type)
		}
		if issues[0].Path != "test.md" {
			t.Errorf("expected path 'test.md', got %q", issues[0].Path)
		}
		if issues[0].Severity != "warning" {
			t.Errorf("expected severity 'warning', got %q", issues[0].Severity)
		}
	})

	t.Run("page with frontmatter and content", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "test.md",
					HasFrontmatter: true,
					Body:           []byte("This is some content."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Test",
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
					RelPath:        "test.md",
					HasFrontmatter: false,
					Body:           []byte(""),
					Content:        []byte(""),
				},
			},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(issues) != 0 {
			t.Errorf("expected no issues (skipped), got %d", len(issues))
		}
	})
}