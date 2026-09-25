// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/template"
)

// checkRequiredFields returns the fields the template marks required that the
// page leaves unset, sorted by name. Only presence is checked: a value is never
// inspected, so type and enum constraints stay with the template's CEL rules.
func checkRequiredFields(tmpl template.Template, fm *frontmatter.ParsedFrontmatter) []string {
	var missing []string
	for key, field := range tmpl.Schema.Frontmatter {
		if !field.Required || frontmatterKeyPresent(fm, key) {
			continue
		}
		missing = append(missing, key)
	}
	sort.Strings(missing)
	return missing
}

// requiredFieldsMessage reports the required fields checkRequiredFields found
// unset, naming the template that declares them.
func requiredFieldsMessage(tmpl template.Template, missing []string) string {
	return fmt.Sprintf("missing required frontmatter field(s): %s (declared required by template %q)",
		strings.Join(missing, ", "), tmpl.Name)
}
