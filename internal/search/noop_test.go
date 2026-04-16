package search

import (
	"context"
	"testing"
)

func TestNoOpSearcher(t *testing.T) {
	searcher := &NoOpSearcher{}
	ctx := context.Background()

	t.Run("IndexPage returns nil", func(t *testing.T) {
		err := searcher.IndexPage(ctx, "notes/test.md", "Test Title", "content here", "tag1,tag2", "summary")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("RemovePage returns nil", func(t *testing.T) {
		err := searcher.RemovePage(ctx, "notes/test.md")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("Search returns empty non-nil slice", func(t *testing.T) {
		result, err := searcher.Search(ctx, "test query", SearchOptions{Limit: 10})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if result == nil {
			t.Error("expected non-nil slice, got nil")
		}
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %d items", len(result))
		}
	})

	t.Run("RebuildIndex returns nil", func(t *testing.T) {
		err := searcher.RebuildIndex(ctx, "/some/kb/root")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}
