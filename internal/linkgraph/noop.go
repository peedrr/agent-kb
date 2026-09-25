// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package linkgraph provides SQLite-backed wikilink tracking between pages.
package linkgraph

import "context"

// NoOpLinkGraphUpdater is a no-op stub. Replaced by real implementation in internal/linkgraph/sqlite.go.
type NoOpLinkGraphUpdater struct{}

// UpdatePageLinks is a no-op stub. Replaced by real implementation in internal/linkgraph/sqlite.go.
func (n *NoOpLinkGraphUpdater) UpdatePageLinks(_ context.Context, _ string, _ string) error {
	return nil
}

// RemovePage is a no-op stub. Replaced by real implementation in internal/linkgraph/sqlite.go.
func (n *NoOpLinkGraphUpdater) RemovePage(_ context.Context, _ string) error {
	return nil
}
