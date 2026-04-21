package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

func TestCategoryDirsChecker(t *testing.T) {
	templates, err := template.LoadTemplatesFromFS(template.DefaultTemplates)
	if err != nil {
		t.Fatalf("failed to load templates: %v", err)
	}

	tests := []struct {
		name       string
		pages      []PageData
		wantIssues int
		checkMsg   string
		checkPath  string
	}{
		{
			name: "note in wrong directory returns warning",
			pages: []PageData{
				{
					RelPath: "kb/random/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Test",
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkMsg:   "should be in kb/notes/",
			checkPath:  "kb/random/test.md",
		},
		{
			name: "note in correct directory returns no issues",
			pages: []PageData{
				{
					RelPath: "kb/notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Test Note",
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "adr in wrong directory returns warning",
			pages: []PageData{
				{
					RelPath: "kb/random/decision.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "adr",
						Title: "Test ADR",
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkMsg:   "should be in kb/decisions/",
			checkPath:  "kb/random/decision.md",
		},
		{
			name: "adr in correct directory returns no issues",
			pages: []PageData{
				{
					RelPath: "kb/decisions/adr-001.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "adr",
						Title: "Test ADR",
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "page without type is skipped",
			pages: []PageData{
				{
					RelPath: "kb/random/no-type.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Title: "No Type",
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
					RelPath:        "kb/random/plain.md",
					HasFrontmatter: false,
				},
			},
			wantIssues: 0,
		},
		{
			name: "page with unknown type is skipped",
			pages: []PageData{
				{
					RelPath: "kb/random/unknown.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "unknown_type",
						Title: "Unknown",
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
				Templates: templates,
				Pages:     tt.pages,
			}

			checker := NewCategoryDirsChecker()
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

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
