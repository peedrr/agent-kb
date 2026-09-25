// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

func TestCELLintCheckerGuardedRuleWithAbsentKey(t *testing.T) {
	tests := []struct {
		name       string
		fields     map[string]any
		wantIssues int
	}{
		{
			name:       "absent optional key passes vacuously",
			fields:     nil,
			wantIssues: 0,
		},
		{
			name:       "present stale key still fails",
			fields:     map[string]any{"updated": "2000-01-01"},
			wantIssues: 1,
		},
	}

	templates := map[string]template.Template{
		"note": {
			Name: "note",
			LintRules: []template.LintRule{
				{
					ID:       "note_stale",
					Rule:     `!has(page.frontmatter.updated) || now - timestamp(page.frontmatter.updated) < duration("2160h")`,
					Severity: "warning",
					Expect:   "Note hasn't been updated in 90 days",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := &KB{
				Templates: templates,
				Pages: []PageData{
					{
						RelPath: "notes/no-updated.md",
						Body:    []byte("Content."),
						Frontmatter: &frontmatter.ParsedFrontmatter{
							Type:   "note",
							Title:  "No Updated",
							Fields: tt.fields,
						},
						HasFrontmatter: true,
					},
				},
			}

			issues, err := NewCELLintChecker().Check(context.Background(), kb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(issues) != tt.wantIssues {
				t.Fatalf("got %d issues, want %d: %+v", len(issues), tt.wantIssues, issues)
			}
			if tt.wantIssues > 0 && issues[0].RuleID != "note_stale" {
				t.Errorf("issue rule_id = %q, want %q", issues[0].RuleID, "note_stale")
			}
		})
	}
}

func TestCELLintCheckerEvaluationErrorDoesNotAbortSweep(t *testing.T) {
	templates := map[string]template.Template{
		"note": {
			Name: "note",
			LintRules: []template.LintRule{
				{
					ID:       "stale_unguarded",
					Rule:     `now - timestamp(page.frontmatter.updated) < duration("2160h")`,
					Severity: "warning",
					Expect:   "Note hasn't been updated in 90 days",
				},
				{
					ID:       "min_words",
					Rule:     `page.content.word_count > 5`,
					Severity: "warning",
					Expect:   "Content should have more than 5 words",
				},
			},
		},
	}

	kb := &KB{
		Templates: templates,
		Pages: []PageData{
			{
				RelPath: "notes/no-updated.md",
				Body:    []byte("One."),
				Frontmatter: &frontmatter.ParsedFrontmatter{
					Type:  "note",
					Title: "No Updated",
				},
				HasFrontmatter: true,
			},
			{
				RelPath: "notes/stale.md",
				Body:    []byte("One."),
				Frontmatter: &frontmatter.ParsedFrontmatter{
					Type:   "note",
					Title:  "Stale",
					Fields: map[string]any{"updated": "2000-01-01"},
				},
				HasFrontmatter: true,
			},
		},
	}

	issues, err := NewCELLintChecker().Check(context.Background(), kb)
	if err != nil {
		t.Fatalf("sweep aborted on rule evaluation error: %v", err)
	}

	var evalErrIssue *LintIssue
	issuesByPath := map[string][]LintIssue{}
	for i := range issues {
		issuesByPath[issues[i].Path] = append(issuesByPath[issues[i].Path], issues[i])
		if issues[i].Path == "notes/no-updated.md" && strings.HasPrefix(issues[i].Message, "rule evaluation error: ") {
			evalErrIssue = &issues[i]
		}
	}

	if evalErrIssue == nil {
		t.Fatalf("no rule evaluation error issue reported for notes/no-updated.md: %+v", issues)
	}
	if evalErrIssue.Type != "cel_lint" {
		t.Errorf("issue type = %q, want %q", evalErrIssue.Type, "cel_lint")
	}
	if evalErrIssue.RuleID != "stale_unguarded" {
		t.Errorf("issue rule_id = %q, want %q", evalErrIssue.RuleID, "stale_unguarded")
	}
	if evalErrIssue.Severity != "error" {
		t.Errorf("issue severity = %q, want %q", evalErrIssue.Severity, "error")
	}

	// The same page's remaining rules still run.
	if !hasIssue(issuesByPath["notes/no-updated.md"], "min_words") {
		t.Errorf("rule after the failing one did not run for notes/no-updated.md: %+v", issuesByPath["notes/no-updated.md"])
	}
	// Subsequent pages are still swept.
	if !hasIssue(issuesByPath["notes/stale.md"], "stale_unguarded") {
		t.Errorf("subsequent page notes/stale.md was not swept: %+v", issuesByPath["notes/stale.md"])
	}
}

func hasIssue(issues []LintIssue, ruleID string) bool {
	for _, issue := range issues {
		if issue.RuleID == ruleID {
			return true
		}
	}
	return false
}

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
