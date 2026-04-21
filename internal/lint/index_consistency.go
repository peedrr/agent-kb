package lint

import (
	"context"
	"fmt"

	"github.com/peedrr/agent-kb/internal/index"
)

type IndexConsistencyChecker struct{}

func NewIndexConsistencyChecker() *IndexConsistencyChecker {
	return &IndexConsistencyChecker{}
}

func (c *IndexConsistencyChecker) Name() string {
	return "index_consistency"
}

func (c *IndexConsistencyChecker) Check(_ context.Context, kb *KB) ([]LintIssue, error) {
	var issues []LintIssue

	entries, err := index.ReadIndex(kb.Root)
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}

	indexPaths := make(map[string]bool)
	for _, e := range entries {
		indexPaths[e.Path] = true
	}

	pagePaths := make(map[string]bool)
	for _, p := range kb.Pages {
		pagePaths[p.RelPath] = true
	}

	for path := range indexPaths {
		if !pagePaths[path] {
			issues = append(issues, LintIssue{
				Type:     "index_consistency",
				Message: "page listed in index.md but file does not exist",
				Path:     path,
				Severity: "error",
			})
		}
	}

	for path := range pagePaths {
		if !indexPaths[path] {
			issues = append(issues, LintIssue{
				Type:     "index_consistency",
				Message: "page exists but is missing from index.md",
				Path:     path,
				Severity: "warning",
			})
		}
	}

	return issues, nil
}