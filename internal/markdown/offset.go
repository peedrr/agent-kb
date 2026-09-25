// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package markdown

import "bytes"

// offsetToLine converts a byte offset in the source to a 1-based line number.
func offsetToLine(source []byte, offset int) int {
	if offset < 0 || offset > len(source) {
		return 1
	}
	return bytes.Count(source[:offset], []byte("\n")) + 1
}
