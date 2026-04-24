package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

func TestFrontmatterSchemaChecker(t *testing.T) {
	templates, err := template.LoadTemplatesFromFS(template.DefaultTemplates)
	if err != nil {
		t.Fatalf("failed to load templates: %v", err)
	}

	tests := []struct {
		name       string
		pages      []PageData
		wantIssues int
		checkType  string
		checkPath  string
		checkMsg   string
	}{
		{
			name: "ADR page missing status field returns schema violation",
			pages: []PageData{
				{
					RelPath: "decisions/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "adr",
						Title: "Test ADR",
						Fields: map[string]any{
							"summary":  "Test summary",
							"tags":     []any{"test"},
							"deciders": []any{"team"},
							"created":  "2024-01-01",
							"updated":  "2024-01-01",
						},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkType:  "frontmatter_schema",
			checkPath:  "decisions/test.md",
			checkMsg:   "missing required field 'status'",
		},
		{
			name: "ADR page with invalid status enum returns enum violation",
			pages: []PageData{
				{
					RelPath: "decisions/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "adr",
						Title: "Test ADR",
						Fields: map[string]any{
							"summary":  "Test summary",
							"tags":     []any{"test"},
							"deciders": []any{"team"},
							"created":  "2024-01-01",
							"updated":  "2024-01-01",
							"status":   "invalid_status",
						},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkType:  "frontmatter_schema",
			checkPath:  "decisions/test.md",
			checkMsg:   "not valid; must be one of",
		},
		{
			name: "ADR page with valid status returns no issues",
			pages: []PageData{
				{
					RelPath: "decisions/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "adr",
						Title: "Test ADR",
						Fields: map[string]any{
							"summary":  "Test summary",
							"tags":     []any{"test"},
							"deciders": []any{"team"},
							"created":  "2024-01-01",
							"updated":  "2024-01-01",
							"status":   "accepted",
						},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "note page with all required fields returns no issues",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Test Note",
						Fields: map[string]any{
							"summary": "Test summary",
							"tags":    []any{"test"},
						},
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
			name: "page with unknown type is skipped",
			pages: []PageData{
				{
					RelPath: "unknown/page.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "unknown_type",
						Title: "Test",
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

			checker := NewFrontmatterSchemaChecker()
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
