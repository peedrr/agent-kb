// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
)

func TestMissingFrontmatterChecker(t *testing.T) {
	checker := NewMissingFrontmatterChecker()

	if checker.Name() != "missing_frontmatter" {
		t.Errorf("expected name 'missing_frontmatter', got %q", checker.Name())
	}

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

		if len(issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(issues))
		}

		if issues[0].Type != "missing_frontmatter" {
			t.Errorf("expected issue type 'missing_frontmatter', got %q", issues[0].Type)
		}
		if issues[0].Path != "test.md" {
			t.Errorf("expected path 'test.md', got %q", issues[0].Path)
		}
		if issues[0].Severity != "error" {
			t.Errorf("expected severity 'error', got %q", issues[0].Severity)
		}
	})

	t.Run("page with frontmatter", func(t *testing.T) {
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
