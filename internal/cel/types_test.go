package cel

import "testing"

func TestTypes(_ *testing.T) {
	_ = Heading{Level: 1, Text: "Test", Line: 1}
	_ = Link{Target: "x", Text: "y", IsWikilink: true, Line: 2}
	_ = CodeBlock{Language: "go", Line: 3}
}
