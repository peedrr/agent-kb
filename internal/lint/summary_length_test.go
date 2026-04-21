package lint

import (
	"context"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
)

func TestSummaryLengthChecker(t *testing.T) {
	tests := []struct {
		name       string
		pages      []PageData
		wantIssues int
		checkMsg   string
		checkPath  string
	}{
		{
			name: "summary too short returns warning",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"summary": "Hi"},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkMsg:   "too short",
			checkPath:  "notes/test.md",
		},
		{
			name: "summary too long returns warning",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"summary": strings.Repeat("a", 250)},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkMsg:   "too long",
			checkPath:  "notes/test.md",
		},
		{
			name: "summary within bounds returns no issue",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"summary": "A proper summary that is neither too short nor too long"},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "summary at minimum length returns no issue",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"summary": strings.Repeat("a", SummaryMinLength)},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "summary at maximum length returns no issue",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"summary": strings.Repeat("a", SummaryMaxLength)},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "page without summary is skipped",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "page without frontmatter is skipped",
			pages: []PageData{
				{
					RelPath:        "notes/plain.md",
					HasFrontmatter: false,
				},
			},
			wantIssues: 0,
		},
		{
			name: "non-string summary is skipped",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"summary": 123},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := &KB{
				Pages: tt.pages,
			}

			checker := NewSummaryLengthChecker()
			issues, err := checker.Check(context.Background(), kb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(issues) != tt.wantIssues {
				t.Errorf("got %d issues, want %d", len(issues), tt.wantIssues)
				for _, i := range issues {
					t.Logf("issue: %+v", i)
				}
			}

			if tt.wantIssues > 0 && len(issues) > 0 {
				if issues[0].Path != tt.checkPath {
					t.Errorf("issue path = %q, want %q", issues[0].Path, tt.checkPath)
				}
				if tt.checkMsg != "" && !containsSubstr(issues[0].Message, tt.checkMsg) {
					t.Errorf("issue message = %q, want to contain %q", issues[0].Message, tt.checkMsg)
				}
				if issues[0].Severity != "warning" {
					t.Errorf("issue severity = %q, want %q", issues[0].Severity, "warning")
				}
			}
		})
	}
}
