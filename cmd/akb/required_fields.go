package main

import (
	"sort"

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
