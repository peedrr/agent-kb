// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cel

import "testing"

func TestTypes(_ *testing.T) {
	_ = Heading{Level: 1, Text: "Test", Line: 1}
	_ = Link{Target: "x", Text: "y", IsWikilink: true, Line: 2}
	_ = CodeBlock{Language: "go", Line: 3}
}
