package cel

import "fmt"

type ValidationError struct {
	RuleID   string
	Message  string
	Line     int
	Severity string // "error" for validations, "warning" for lint_rules
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("[%s] %s (near line %d)", e.RuleID, e.Message, e.Line)
	}
	return fmt.Sprintf("[%s] %s", e.RuleID, e.Message)
}