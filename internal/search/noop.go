// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
