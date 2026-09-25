// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cel

import "fmt"

// ValidationError describes a CEL rule failure reported at write time or by a
// lint sweep.
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
