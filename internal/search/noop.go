package search

import "context"

// NoOpSearcher is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
type NoOpSearcher struct{}

// IndexPage is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
func (n *NoOpSearcher) IndexPage(ctx context.Context, path, title, content, tags, summary string) error {
	return nil
}

// RemovePage is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
func (n *NoOpSearcher) RemovePage(ctx context.Context, path string) error {
	return nil
}

// Search is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
func (n *NoOpSearcher) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	return []SearchResult{}, nil
}

// RebuildIndex is a no-op stub. Replaced by real implementation in internal/search/sqlite.go.
func (n *NoOpSearcher) RebuildIndex(ctx context.Context, kbRoot string) error {
	return nil
}
