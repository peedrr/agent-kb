package markdown

import "bytes"

// offsetToLine converts a byte offset in the source to a 1-based line number.
func offsetToLine(source []byte, offset int) int {
	if offset < 0 || offset > len(source) {
		return 1
	}
	return bytes.Count(source[:offset], []byte("\n")) + 1
}
