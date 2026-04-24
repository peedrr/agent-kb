package lint

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/linkgraph"
)

type mockChecker struct {
	name   string
	issues []LintIssue
	err    error
}

func (m *mockChecker) Name() string {
	return m.name
}

func (m *mockChecker) Check(_ context.Context, _ *KB) ([]LintIssue, error) {
	return m.issues, m.err
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	d, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() }) //nolint:errcheck // test cleanup — failure is non-fatal
	if err := db.CreateSchema(d); err != nil {
		t.Fatalf("CreateSchema: %v", err)
	}
	return d
}

func TestNewLintEngine(t *testing.T) {
	e := NewLintEngine()
	if e == nil {
		t.Fatal("NewLintEngine returned nil")
	}
	if len(e.checkers) != 0 {
		t.Fatalf("new engine has %d checkers, want 0", len(e.checkers))
	}
}

func TestAddChecker(t *testing.T) {
	e := NewLintEngine()
	c1 := &mockChecker{name: "check_a"}
	c2 := &mockChecker{name: "check_b"}
	e.AddChecker(c1)
	e.AddChecker(c2)
	if len(e.checkers) != 2 {
		t.Fatalf("engine has %d checkers, want 2", len(e.checkers))
	}
	if e.checkers[0].Name() != "check_a" {
		t.Errorf("first checker = %q, want %q", e.checkers[0].Name(), "check_a")
	}
	if e.checkers[1].Name() != "check_b" {
		t.Errorf("second checker = %q, want %q", e.checkers[1].Name(), "check_b")
	}
}

func TestRun_AggregatesIssues(t *testing.T) {
	e := NewLintEngine()
	e.AddChecker(&mockChecker{
		name: "broken_links",
		issues: []LintIssue{
			{Type: "broken_links", Message: "link to missing", Path: "kb/a.md", Severity: "error"},
		},
	})
	e.AddChecker(&mockChecker{
		name: "orphans",
		issues: []LintIssue{
			{Type: "orphans", Message: "no inbound links", Path: "kb/b.md", Severity: "warning"},
			{Type: "orphans", Message: "no inbound links", Path: "kb/c.md", Severity: "warning"},
		},
	})

	kb := &KB{
		Pages: []PageData{{RelPath: "kb/a.md"}, {RelPath: "kb/b.md"}, {RelPath: "kb/c.md"}},
	}

	report, err := e.Run(context.Background(), kb)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Issues) != 3 {
		t.Fatalf("len(Issues) = %d, want 3", len(report.Issues))
	}
	if report.PagesChecked != 3 {
		t.Errorf("PagesChecked = %d, want 3", report.PagesChecked)
	}
}

func TestRun_ByCheckMap(t *testing.T) {
	e := NewLintEngine()
	e.AddChecker(&mockChecker{
		name: "broken_links",
		issues: []LintIssue{
			{Type: "broken_links", Message: "m1", Path: "a.md", Severity: "error"},
			{Type: "broken_links", Message: "m2", Path: "b.md", Severity: "error"},
		},
	})
	e.AddChecker(&mockChecker{
		name:   "orphans",
		issues: []LintIssue{},
	})

	kb := &KB{Pages: []PageData{{RelPath: "a.md"}}}

	report, err := e.Run(context.Background(), kb)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.ByCheck["broken_links"] != 2 {
		t.Errorf("ByCheck[broken_links] = %d, want 2", report.ByCheck["broken_links"])
	}
	if report.ByCheck["orphans"] != 0 {
		t.Errorf("ByCheck[orphans] = %d, want 0", report.ByCheck["orphans"])
	}
}

func TestRun_PagesChecked(t *testing.T) {
	e := NewLintEngine()
	pages := make([]PageData, 5)
	kb := &KB{Pages: pages}

	report, err := e.Run(context.Background(), kb)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.PagesChecked != 5 {
		t.Errorf("PagesChecked = %d, want 5", report.PagesChecked)
	}
}

func TestRun_EmptyEngine(t *testing.T) {
	e := NewLintEngine()
	kb := &KB{Pages: []PageData{{RelPath: "a.md"}}}

	report, err := e.Run(context.Background(), kb)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Issues) != 0 {
		t.Errorf("len(Issues) = %d, want 0", len(report.Issues))
	}
	if report.PagesChecked != 1 {
		t.Errorf("PagesChecked = %d, want 1", report.PagesChecked)
	}
}

func TestRun_CheckerError(t *testing.T) {
	e := NewLintEngine()
	e.AddChecker(&mockChecker{
		name: "failing",
		err:  sql.ErrConnDone,
	})

	kb := &KB{}
	_, err := e.Run(context.Background(), kb)
	if err == nil {
		t.Fatal("expected error from failing checker")
	}
}

func TestRun_WithLinkGraph(t *testing.T) {
	d := setupTestDB(t)
	g := linkgraph.NewSQLiteLinkGraph(d)

	e := NewLintEngine()
	e.AddChecker(&mockChecker{
		name:   "broken_links",
		issues: []LintIssue{},
	})

	kb := &KB{
		Root:      t.TempDir(),
		LinkGraph: g,
		Pages:     []PageData{{RelPath: "test.md"}},
	}

	report, err := e.Run(context.Background(), kb)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.PagesChecked != 1 {
		t.Errorf("PagesChecked = %d, want 1", report.PagesChecked)
	}
}

func TestThresholdConstants(t *testing.T) {
	if ProvenanceDriftThreshold != 0.20 {
		t.Errorf("ProvenanceDriftThreshold = %f, want 0.20", ProvenanceDriftThreshold)
	}
	if FreshnessHalfLifeDays != 30 {
		t.Errorf("FreshnessHalfLifeDays = %d, want 30", FreshnessHalfLifeDays)
	}
	if FreshnessScoreThreshold != 50.0 {
		t.Errorf("FreshnessScoreThreshold = %f, want 50.0", FreshnessScoreThreshold)
	}
	if SummaryMinLength != 10 {
		t.Errorf("SummaryMinLength = %d, want 10", SummaryMinLength)
	}
	if SummaryMaxLength != 200 {
		t.Errorf("SummaryMaxLength = %d, want 200", SummaryMaxLength)
	}
}
