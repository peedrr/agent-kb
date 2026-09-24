package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

// requiredFieldsADRTemplate is the schema the checker subtests validate pages
// against: three required fields beyond the title and type every page carries,
// plus one optional field that must never be reported.
func requiredFieldsADRTemplate() template.Template {
	return template.Template{
		Name: "adr",
		Schema: template.Schema{
			Frontmatter: map[string]template.FieldSchema{
				"title":    {Type: "string", Required: true},
				"type":     {Type: "string", Required: true},
				"status":   {Type: "string", Required: true},
				"deciders": {Type: "string", Required: true},
				"tags":     {Type: "list", Required: true},
				"sources":  {Type: "list", Required: false},
			},
		},
	}
}

// requiredFieldsKB builds the lint context of one page of type adr.
func requiredFieldsKB(page PageData) *KB {
	return &KB{
		Pages:     []PageData{page},
		Templates: map[string]template.Template{"adr": requiredFieldsADRTemplate()},
	}
}

func requiredFieldsPage(fields map[string]any) PageData {
	return PageData{
		RelPath:        "kb/decisions/test.md",
		HasFrontmatter: true,
		Frontmatter: &frontmatter.ParsedFrontmatter{
			Type:   "adr",
			Title:  "Test ADR",
			Fields: fields,
		},
	}
}

func TestRequiredFieldsCheckerName(t *testing.T) {
	if got := NewRequiredFieldsChecker().Name(); got != "required_fields" {
		t.Errorf("Name() = %q, want %q", got, "required_fields")
	}
}

func TestRequiredFieldsChecker(t *testing.T) {
	checker := NewRequiredFieldsChecker()

	t.Run("page missing several required fields lists them sorted", func(t *testing.T) {
		kb := requiredFieldsKB(requiredFieldsPage(map[string]any{}))

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
		}

		issue := issues[0]
		if issue.Type != "required_fields" {
			t.Errorf("issue type = %q, want %q", issue.Type, "required_fields")
		}
		if issue.RuleID != "required_fields" {
			t.Errorf("issue rule id = %q, want %q", issue.RuleID, "required_fields")
		}
		if issue.Path != "kb/decisions/test.md" {
			t.Errorf("issue path = %q, want %q", issue.Path, "kb/decisions/test.md")
		}
		if issue.Severity != "error" {
			t.Errorf("issue severity = %q, want %q", issue.Severity, "error")
		}
		want := `missing required frontmatter field(s): deciders, status, tags (declared required by template "adr")`
		if issue.Message != want {
			t.Errorf("issue message = %q, want %q", issue.Message, want)
		}
	})

	t.Run("page missing one required field", func(t *testing.T) {
		kb := requiredFieldsKB(requiredFieldsPage(map[string]any{
			"status":   "accepted",
			"deciders": "team",
		}))

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
		}
		want := `missing required frontmatter field(s): tags (declared required by template "adr")`
		if issues[0].Message != want {
			t.Errorf("issue message = %q, want %q", issues[0].Message, want)
		}
	})

	t.Run("page carrying every required field", func(t *testing.T) {
		kb := requiredFieldsKB(requiredFieldsPage(map[string]any{
			"status":   "accepted",
			"deciders": "team",
			"tags":     []any{"kb"},
		}))

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues, got %d: %v", len(issues), issues)
		}
	})

	t.Run("a set field counts as present even when empty", func(t *testing.T) {
		kb := requiredFieldsKB(requiredFieldsPage(map[string]any{
			"status":   "",
			"deciders": "team",
			"tags":     []any{},
		}))

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected presence only, got %d issues: %v", len(issues), issues)
		}
	})

	t.Run("title is required by the schema and carried by the page", func(t *testing.T) {
		kb := requiredFieldsKB(requiredFieldsPage(map[string]any{
			"status":   "accepted",
			"deciders": "team",
			"tags":     []any{"kb"},
		}))
		kb.Pages[0].Frontmatter.Title = ""

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
		}
		want := `missing required frontmatter field(s): title (declared required by template "adr")`
		if issues[0].Message != want {
			t.Errorf("issue message = %q, want %q", issues[0].Message, want)
		}
	})

	t.Run("optional fields are never reported", func(t *testing.T) {
		kb := requiredFieldsKB(requiredFieldsPage(map[string]any{
			"status":   "accepted",
			"deciders": "team",
			"tags":     []any{"kb"},
		}))

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues for a missing optional field, got %d: %v", len(issues), issues)
		}
	})

	t.Run("page whose type has no template", func(t *testing.T) {
		kb := requiredFieldsKB(PageData{
			RelPath:        "kb/notes/orphan.md",
			HasFrontmatter: true,
			Frontmatter:    &frontmatter.ParsedFrontmatter{Type: "retired-type", Title: "Orphan"},
		})

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues for a type without a template, got %d: %v", len(issues), issues)
		}
	})

	t.Run("page without frontmatter", func(t *testing.T) {
		kb := requiredFieldsKB(PageData{
			RelPath:        "kb/bare.md",
			HasFrontmatter: false,
			Body:           []byte("Body without frontmatter."),
		})

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues for a page without frontmatter, got %d: %v", len(issues), issues)
		}
	})

	t.Run("page with an empty type", func(t *testing.T) {
		kb := requiredFieldsKB(PageData{
			RelPath:        "kb/typeless.md",
			HasFrontmatter: true,
			Frontmatter:    &frontmatter.ParsedFrontmatter{Title: "Typeless"},
		})

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected no issues for a page with an empty type, got %d: %v", len(issues), issues)
		}
	})
}
