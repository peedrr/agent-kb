package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

func TestCELLintChecker(t *testing.T) {
	tests := []struct {
		name       string
		templates  map[string]template.Template
		pages      []PageData
		wantIssues int
		checkRule  string
		checkMsg   string
		checkSev   string
	}{
		{
			name: "page matching all lint rules returns no issues",
			templates: map[string]template.Template{
				"note": {
					Name: "note",
					LintRules: []template.LintRule{
						{
							ID:       "has_title",
							Rule:     `page.frontmatter.title != ""`,
							Severity: "error",
							Expect:   "Page must have a title",
						},
					},
				},
			},
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Body:    []byte("Some content here."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Test Note",
						Fields: map[string]any{
							"tags": []string{"test"},
						},
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "page failing lint rule returns issue with correct RuleID Severity Message",
			templates: map[string]template.Template{
				"note": {
					Name: "note",
					LintRules: []template.LintRule{
						{
							ID:       "min_words",
							Rule:     `page.content.word_count > 10`,
							Severity: "warning",
							Expect:   "Content should have more than 10 words",
						},
					},
				},
			},
			pages: []PageData{
				{
					RelPath: "notes/short.md",
					Body:    []byte("Short text."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Short Note",
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 1,
			checkRule:  "min_words",
			checkMsg:   "Content should have more than 10 words",
			checkSev:   "warning",
		},
		{
			name: "now variable works in temporal rules",
			templates: map[string]template.Template{
				"note": {
					Name: "note",
					LintRules: []template.LintRule{
						{
							ID:       "future_check",
							Rule:     `now > timestamp("2000-01-01T00:00:00Z")`,
							Severity: "error",
							Expect:   "System clock appears incorrect",
						},
					},
				},
			},
			pages: []PageData{
				{
					RelPath: "notes/temporal.md",
					Body:    []byte("Content."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Temporal Note",
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "page without frontmatter is skipped",
			templates: map[string]template.Template{
				"note": {
					Name: "note",
					LintRules: []template.LintRule{
						{
							ID:       "has_title",
							Rule:     `page.frontmatter.title != ""`,
							Severity: "error",
							Expect:   "Page must have a title",
						},
					},
				},
			},
			pages: []PageData{
				{
					RelPath:        "notes/plain.md",
					Body:           []byte("No frontmatter."),
					HasFrontmatter: false,
				},
			},
			wantIssues: 0,
		},
		{
			name: "page without type is skipped",
			templates: map[string]template.Template{
				"note": {
					Name: "note",
					LintRules: []template.LintRule{
						{
							ID:       "has_title",
							Rule:     `page.frontmatter.title != ""`,
							Severity: "error",
							Expect:   "Page must have a title",
						},
					},
				},
			},
			pages: []PageData{
				{
					RelPath: "notes/no-type.md",
					Body:    []byte("Content."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Title: "No Type",
					},
					HasFrontmatter: true,
				},
			},
			wantIssues: 0,
		},
		{
			name: "page with unknown type is skipped",
			templates: map[string]template.Template{
				"note": {
					Name: "note",
					LintRules: []template.LintRule{
						{
							ID:       "has_title",
							Rule:     `page.frontmatter.title != ""`,
							Severity: "error",
							Expect:   "Page must have a title",
						},
					},
				},
			},
			pages: []PageData{
				{
					RelPath: "notes/unknown.md",
					Body:    []byte("Content."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "unknown",
						Title: "Unknown Type",
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
				Templates: tt.templates,
				Pages:     tt.pages,
			}

			checker := NewCELLintChecker()
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
				if issues[0].RuleID != tt.checkRule {
					t.Errorf("issue rule_id = %q, want %q", issues[0].RuleID, tt.checkRule)
				}
				if issues[0].Message != tt.checkMsg {
					t.Errorf("issue message = %q, want %q", issues[0].Message, tt.checkMsg)
				}
				if issues[0].Severity != tt.checkSev {
					t.Errorf("issue severity = %q, want %q", issues[0].Severity, tt.checkSev)
				}
				if issues[0].Type != "cel_lint" {
					t.Errorf("issue type = %q, want %q", issues[0].Type, "cel_lint")
				}
			}
		})
	}
}
