// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package markdown

import "testing"

func TestOffsetToLine(t *testing.T) {
	source := []byte("a\nb\nc")
	tests := []struct {
		offset int
		want   int
	}{
		{0, 1},
		{2, 2},
		{4, 3},
	}
	for _, tt := range tests {
		got := offsetToLine(source, tt.offset)
		if got != tt.want {
			t.Errorf("offsetToLine(%q, %d) = %d, want %d", source, tt.offset, got, tt.want)
		}
	}
}
