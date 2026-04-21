package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

func TestTypeExistsChecker(t *testing.T) {
	templates, err := template.LoadTemplatesFromFS(template.DefaultTemplates)
	if err != nil {
		t.Fatalf("failed to load templates: %v", err)
	}

	tests := []struct {
		name        string
		pages       []PageData
		wantIssues  int
		checkType   string
		checkPath   string
		checkMsg    string
	}{
		{
			name: "page with unknown type returns error",
			pages: []PageData{
				{
					RelPath: "decisions/unknown.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "unknown_type",
						Title: "Test",
					},
					HasFrontmatter: true,
				},
			},
			wantIssues:  1,
			checkType:  "type_exists",
			checkPath:  "decisions/unknown.md",
			checkMsg:   "unknown type 'unknown_type'",
		},
		{
			name: "page with known type returns no issues",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
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
			name: "page without type is skipped",
			pages: []PageData{
				{
					RelPath: "notes/no-type.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Title: "Test Note",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := &KB{
				Templates: templates,
				Pages:     tt.pages,
			}

			checker := NewTypeExistsChecker()
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
				if issues[0].Type != tt.checkType {
					t.Errorf("issue type = %q, want %q", issues[0].Type, tt.checkType)
				}
				if issues[0].Path != tt.checkPath {
					t.Errorf("issue path = %q, want %q", issues[0].Path, tt.checkPath)
				}
				if tt.checkMsg != "" && !contains(issues[0].Message, tt.checkMsg) {
					t.Errorf("issue message = %q, want to contain %q", issues[0].Message, tt.checkMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
