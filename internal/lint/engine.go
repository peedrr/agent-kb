package lint

import (
	"context"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/markdown"
	"github.com/peedrr/agent-kb/internal/template"
)

type LintIssue struct {
	Type     string `json:"check"`
	Message  string `json:"message"`
	Path     string `json:"path"`
	Severity string `json:"severity"`
}

type LintReport struct {
	Issues       []LintIssue    `json:"issues"`
	PagesChecked int            `json:"pages_checked"`
	ByCheck      map[string]int `json:"by_check"`
}

type LintChecker interface {
	Name() string
	Check(ctx context.Context, kb *KB) ([]LintIssue, error)
}

type KB struct {
	Root      string
	LinkGraph *linkgraph.SQLiteLinkGraph
	Templates map[string]template.Template
	Manifest  []manifest.Entry
	Pages     []PageData
}

type PageData struct {
	RelPath           string
	Content           []byte
	Body              []byte
	Frontmatter       *frontmatter.ParsedFrontmatter
	ProvenanceMarkers []markdown.ProvenanceMarker
	Annotations       []markdown.Annotation
	HasFrontmatter    bool
}

type LintEngine struct {
	checkers []LintChecker
}

func NewLintEngine() *LintEngine {
	return &LintEngine{}
}

func (e *LintEngine) AddChecker(c LintChecker) {
	e.checkers = append(e.checkers, c)
}

func (e *LintEngine) Run(ctx context.Context, kb *KB) (*LintReport, error) {
	report := &LintReport{
		ByCheck: make(map[string]int),
	}
	report.PagesChecked = len(kb.Pages)

	for _, checker := range e.checkers {
		issues, err := checker.Check(ctx, kb)
		if err != nil {
			return nil, err
		}
		report.Issues = append(report.Issues, issues...)
		report.ByCheck[checker.Name()] = len(issues)
	}

	return report, nil
}
