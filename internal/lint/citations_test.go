// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"testing"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/manifest"
)

func TestCitationsChecker(t *testing.T) {
	tests := []struct {
		name       string
		pages      []PageData
		manifest   []manifest.Entry
		wantIssues int
		checkMsg   string
		checkPath  string
	}{
		{
			name: "source not in manifest returns error",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"sources": "missing.csv"},
					},
					HasFrontmatter: true,
				},
			},
			manifest:   []manifest.Entry{},
			wantIssues: 1,
			checkMsg:   "not found in raw manifest",
			checkPath:  "notes/test.md",
		},
		{
			name: "source in manifest returns no issue",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"sources": "data.csv"},
					},
					HasFrontmatter: true,
				},
			},
			manifest: []manifest.Entry{
				{Filename: "data.csv"},
			},
			wantIssues: 0,
		},
		{
			name: "sources as array with one missing returns error",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"sources": []any{"exists.csv", "missing.csv"}},
					},
					HasFrontmatter: true,
				},
			},
			manifest: []manifest.Entry{
				{Filename: "exists.csv"},
			},
			wantIssues: 1,
			checkMsg:   "missing.csv",
			checkPath:  "notes/test.md",
		},
		{
			name: "sources as array all present returns no issue",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"sources": []any{"file1.csv", "file2.csv"}},
					},
					HasFrontmatter: true,
				},
			},
			manifest: []manifest.Entry{
				{Filename: "file1.csv"},
				{Filename: "file2.csv"},
			},
			wantIssues: 0,
		},
		{
			name: "page without sources is skipped",
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
			manifest:   []manifest.Entry{},
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
			manifest:   []manifest.Entry{},
			wantIssues: 0,
		},
		{
			name: "sources with non-string type is skipped",
			pages: []PageData{
				{
					RelPath: "notes/test.md",
					Frontmatter: &frontmatter.ParsedFrontmatter{
						Type:   "note",
						Title:  "Test",
						Fields: map[string]any{"sources": 123},
					},
					HasFrontmatter: true,
				},
			},
			manifest:   []manifest.Entry{},
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := &KB{
				Pages:    tt.pages,
				Manifest: tt.manifest,
			}

			checker := NewCitationsChecker()
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
				if issues[0].Severity != "error" {
					t.Errorf("issue severity = %q, want %q", issues[0].Severity, "error")
				}
			}
		})
	}
}
