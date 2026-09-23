package cel

import (
	"testing"
)

func TestValidationError(t *testing.T) {
	e1 := ValidationError{RuleID: "require_context", Message: "Missing required H2 section 'Context'.", Line: 12, Severity: "error"}
	got1 := e1.Error()
	want1 := "[require_context] Missing required H2 section 'Context'. (near line 12)"
	if got1 != want1 {
		t.Errorf("Error() = %q, want %q", got1, want1)
	}

	e2 := ValidationError{RuleID: "no_unknown_fields", Message: "Unknown frontmatter field 'auther'.", Line: 0, Severity: "error"}
	got2 := e2.Error()
	want2 := "[no_unknown_fields] Unknown frontmatter field 'auther'."
	if got2 != want2 {
		t.Errorf("Error() = %q, want %q", got2, want2)
	}
}
