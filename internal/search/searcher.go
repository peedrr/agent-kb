package search

import "context"

// SearchResult represents a search result from the index.
//
//nolint:revive // intentionally exported for use by consumers
type SearchResult struct {
	Path    string
	Title   string
	Summary string
	Snippet string
	Rank    float64
}

// SearchOptions configures a search query.
//
//nolint:revive // intentionally exported for use by consumers
type SearchOptions struct {
	Limit int // 0 means default of 10
}

// Searcher provides full-text search over the knowledge base.
type Searcher interface {
	// IndexPage adds or updates a page in the search index.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	IndexPage(ctx context.Context, path, title, content, tags, summary string) error

	// RemovePage removes a page from the search index.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	RemovePage(ctx context.Context, path string) error

	// Search performs a full-text search over the index.
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)

	// RebuildIndex rebuilds the entire search index from scratch.
	RebuildIndex(ctx context.Context, kbRoot string) error
}
