// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package linkgraph

import (
	"context"
	"testing"
)

func TestNoOpLinkGraphUpdater(t *testing.T) {
	updater := &NoOpLinkGraphUpdater{}
	ctx := context.Background()

	t.Run("UpdatePageLinks returns nil", func(t *testing.T) {
		err := updater.UpdatePageLinks(ctx, "notes/test.md", "# Hello\n\n[Link](other.md)")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("RemovePage returns nil", func(t *testing.T) {
		err := updater.RemovePage(ctx, "notes/test.md")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}
