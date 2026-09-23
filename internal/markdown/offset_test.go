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
