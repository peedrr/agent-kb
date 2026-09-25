// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

func TestTypeOrphanChecker(t *testing.T) {
	checker := NewTypeOrphanChecker()

	if checker.Name() != "type_orphan" {
		t.Errorf("expected name 'type_orphan', got %q", checker.Name())
	}

	t.Run("page with type that has no template", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "test.md",
					HasFrontmatter: true,
					Body:           []byte("Content here."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "missing-type",
						Title: "Test",
					},
				},
			},
			Templates: map[string]template.Template{},
		}

		issues, err := checker.Check(context.Background(), kb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}

		if issues[0].Type != "type_orphan" {
			t.Errorf("expected issue type 'type_orphan', got %q", issues[0].Type)
		}
		if issues[0].Path != "test.md" {
			t.Errorf("expected path 'test.md', got %q", issues[0].Path)
		}
		if issues[0].Severity != "error" {
			t.Errorf("expected severity 'error', got %q", issues[0].Severity)
		}
		if issues[0].Message != "page has type 'missing-type' but no template 'missing-type' exists" {
			t.Errorf("unexpected message: %q", issues[0].Message)
		}
	})

	t.Run("page with type that has template", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "test.md",
					HasFrontmatter: true,
					Body:           []byte("Content here."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "note",
						Title: "Test",
					},
				},
			},
			Templates: map[string]template.Template{
				"note": {},
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
					Body:           []byte("Some content without frontmatter."),
					Content:        []byte("Some content without frontmatter."),
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

	t.Run("page with empty type", func(t *testing.T) {
		kb := &KB{
			Pages: []PageData{
				{
					RelPath:        "test.md",
					HasFrontmatter: true,
					Body:           []byte("Content here."),
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:  "",
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
}
