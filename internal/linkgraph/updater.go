package linkgraph

import "context"

// Updater updates the link graph for the knowledge base.
// It tracks page links and allows removal of pages from the graph.
type Updater interface {
	// UpdatePageLinks extracts links from the given content and updates the graph.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	UpdatePageLinks(ctx context.Context, path string, content string) error

	// RemovePage removes a page from the link graph.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	RemovePage(ctx context.Context, path string) error
}
