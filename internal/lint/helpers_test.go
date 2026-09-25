// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

import "strings"

// containsSubstr checks if s contains substr as a substring.
func containsSubstr(s, substr string) bool {
	return strings.Contains(s, substr)
}
