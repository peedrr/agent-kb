// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package search

import (
	"context"
	"database/sql"
)

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
	Limit int
	Tag   string
	Type  string
	After string
}

// Searcher provides full-text search over the knowledge base.
//
// Every method runs in a transaction of its own. TxSearcher offers the two
// index writes on a transaction owned by the caller instead.
type Searcher interface {
	// IndexPage adds or updates a page in the search index.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	IndexPage(ctx context.Context, path, title, content, tags, summary, pageType string) error

	// RemovePage removes a page from the search index.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	RemovePage(ctx context.Context, path string) error

	// Search performs a full-text search over the index.
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)

	// RebuildIndex rebuilds the entire search index from scratch.
	RebuildIndex(ctx context.Context, kbRoot string) error
}

// TxSearcher provides the search-index writes on a transaction owned by the
// caller, so the search-index step can commit or roll back together with other
// index steps. The transaction must belong to the same database as the
// implementation.
type TxSearcher interface {
	// IndexPageTx adds or updates a page in the search index using tx.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	IndexPageTx(ctx context.Context, tx *sql.Tx, path, title, content, tags, summary, pageType string) error

	// RemovePageTx removes a page from the search index using tx.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	RemovePageTx(ctx context.Context, tx *sql.Tx, path string) error
}
