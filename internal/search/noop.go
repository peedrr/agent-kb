// Package search provides full-text search over the knowledge base.
package search

import "context"

// NoOpSearcher is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
type NoOpSearcher struct{}

// IndexPage is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
func (n *NoOpSearcher) IndexPage(_ context.Context, _, _, _, _, _, _ string) error {
	return nil
}

// RemovePage is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
func (n *NoOpSearcher) RemovePage(_ context.Context, _ string) error {
	return nil
}

// Search is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
func (n *NoOpSearcher) Search(_ context.Context, _ string, _ SearchOptions) ([]SearchResult, error) {
	return []SearchResult{}, nil
}

// RebuildIndex is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
func (n *NoOpSearcher) RebuildIndex(_ context.Context, _ string) error {
	return nil
}
