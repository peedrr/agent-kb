package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/markdown"
)

func TestConfidenceChecker(t *testing.T) {
	tests := []struct {
		name       string
		pages      []PageData
		wantIssues int
		checkMsg   string
		checkPath  string
		severity   string
	}{
		{
			name: "invalid confidence returns error",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"confidence": "super-high"},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkMsg:   "not valid",
			checkPath:  "notes/test.md",
			severity:   "error",
		},
		{
			name: "valid confidence high returns no issue",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"confidence": "high"},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "valid confidence medium returns no issue",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"confidence": "medium"},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "low confidence without provenance markers returns warning",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"confidence": "low"},
					},
					HasFrontmatter:    true,
					ProvenanceMarkers: []markdown.ProvenanceMarker{},
				},
			},
			wantIssues: 1,
			checkMsg:   "no provenance markers",
			checkPath:  "notes/test.md",
			severity:   "warning",
		},
		{
			name: "low confidence with provenance markers returns no issue",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"confidence": "low"},
					},
					HasFrontmatter: true,
					ProvenanceMarkers: []markdown.ProvenanceMarker{
						{Type: "inferred", Position: 10},
					},
				},
			},
			wantIssues: 0,
		},
		{
			name: "non-string confidence returns error",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"confidence": 123},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkMsg:   "not valid",
			checkPath:  "notes/test.md",
			severity:   "error",
		},
		{
			name: "page without confidence is skipped",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := &KB{
				Pages: tt.pages,
			}

			checker := NewConfidenceChecker()
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
				if tt.severity != "" && issues[0].Severity != tt.severity {
					t.Errorf("issue severity = %q, want %q", issues[0].Severity, tt.severity)
				}
			}
		})
	}
}
