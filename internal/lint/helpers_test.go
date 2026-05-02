package lint

import "strings"

// containsSubstr checks if s contains substr as a substring.
func containsSubstr(s, substr string) bool {
	return strings.Contains(s, substr)
}