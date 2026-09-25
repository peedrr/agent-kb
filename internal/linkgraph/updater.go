// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package linkgraph

import (
	"context"
	"database/sql"
)

// Updater updates the link graph for the knowledge base.
// It tracks page links and allows removal of pages from the graph.
//
// Every method runs in a transaction of its own. TxUpdater offers the same
// writes on a transaction owned by the caller instead.
type Updater interface {
	// UpdatePageLinks extracts links from the given content and updates the graph.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	UpdatePageLinks(ctx context.Context, path string, content string) error

	// RemovePage removes a page from the link graph.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	RemovePage(ctx context.Context, path string) error
}

// TxUpdater updates the link graph on a transaction owned by the caller, so the
// link-graph step can commit or roll back together with other index steps. The
// transaction must belong to the same database as the implementation.
type TxUpdater interface {
	// UpdatePageLinksTx extracts links from the given content and updates the graph using tx.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	UpdatePageLinksTx(ctx context.Context, tx *sql.Tx, path string, content string) error

	// RemovePageTx removes a page from the link graph using tx.
	// The path is relative to KB root (e.g., "notes/my-note.md").
	RemovePageTx(ctx context.Context, tx *sql.Tx, path string) error
}
