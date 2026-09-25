// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import (
	"context"
	"fmt"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/markdown"
	"github.com/peedrr/agent-kb/internal/template"
)

// LintIssue represents a single lint issue.
//
//nolint:revive // intentionally exported for use by consumers
type LintIssue struct {
	Type     string `json:"check"`
	RuleID   string `json:"rule_id,omitempty"`
	Message  string `json:"message"`
	Path     string `json:"path"`
	Severity string `json:"severity"`
}

// LintReport holds lint results for a KB run.
//
//nolint:revive // intentionally exported for use by consumers
type LintReport struct {
	Issues       []LintIssue    `json:"issues"`
	PagesChecked int            `json:"pages_checked"`
	ByCheck      map[string]int `json:"by_check"`
}

// LintChecker is the interface implemented by all lint checkers.
//
//nolint:revive // intentionally exported for use by consumers
type LintChecker interface {
	Name() string
	Check(ctx context.Context, kb *KB) ([]LintIssue, error)
}

// KB is the lint context holding KB root, linkgraph, templates, manifest, and parsed pages.
//
//nolint:revive // intentionally exported for use by checkers
type KB struct {
	Root      string
	LinkGraph *linkgraph.SQLiteLinkGraph
	Templates map[string]template.Template
	Manifest  []manifest.Entry
	Pages     []PageData
}

// PageData holds parsed page data for lint checking.
//
//nolint:revive // intentionally exported for use by checkers
type PageData struct {
	RelPath           string
	Content           []byte
	Body              []byte
	Frontmatter       *frontmatter.ParsedFrontmatter
	ProvenanceMarkers []markdown.ProvenanceMarker
	Annotations       []markdown.Annotation
	HasFrontmatter    bool
}

// LintEngine orchestrates running lint checkers over a KB.
//
//nolint:revive // intentionally exported for use by consumers
type LintEngine struct {
	checkers []LintChecker
}

// NewLintEngine creates a new LintEngine.
//
//nolint:revive // intentionally exported for use by consumers
func NewLintEngine() *LintEngine {
	return &LintEngine{}
}

// AddChecker registers a lint checker with the engine.
//
//nolint:revive // intentionally exported for use by consumers
func (e *LintEngine) AddChecker(c LintChecker) {
	e.checkers = append(e.checkers, c)
}

// Run executes all registered lint checkers.
//
//nolint:revive // intentionally exported for use by consumers
func (e *LintEngine) Run(ctx context.Context, kb *KB) (*LintReport, error) {
	report := &LintReport{
		ByCheck: make(map[string]int),
	}
	report.PagesChecked = len(kb.Pages)

	for _, checker := range e.checkers {
		issues, err := checker.Check(ctx, kb)
		if err != nil {
			return nil, fmt.Errorf("run checker %s: %w", checker.Name(), err)
		}
		report.Issues = append(report.Issues, issues...)
		report.ByCheck[checker.Name()] = len(issues)
	}

	return report, nil
}
