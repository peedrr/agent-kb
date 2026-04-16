package linkgraph

import "context"

// NoOpLinkGraphUpdater is a no-op stub. Replaced by real implementation in internal/linkgraph/sqlite.go (P3).
type NoOpLinkGraphUpdater struct{}

// UpdatePageLinks is a no-op stub. Replaced by real implementation in internal/linkgraph/sqlite.go (P3).
func (n *NoOpLinkGraphUpdater) UpdatePageLinks(ctx context.Context, path string, content string) error {
	return nil
}

// RemovePage is a no-op stub. Replaced by real implementation in internal/linkgraph/sqlite.go (P3).
func (n *NoOpLinkGraphUpdater) RemovePage(ctx context.Context, path string) error {
	return nil
}
