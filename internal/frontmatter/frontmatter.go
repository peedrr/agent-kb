// Package frontmatter parses and validates YAML frontmatter from markdown pages.
package frontmatter

import (
	"bytes"
	"fmt"

	"github.com/adrg/frontmatter"

	"github.com/peedrr/agent-kb/internal/template"
)

// ParsedFrontmatter holds the extracted frontmatter fields from a page.
type ParsedFrontmatter struct {
	Type   string
	Title  string
	Fields map[string]any
}

// Parse extracts frontmatter from markdown content.
func Parse(content []byte) (*ParsedFrontmatter, []byte, error) {
	var raw map[string]any
	body, err := frontmatter.MustParse(bytes.NewReader(content), &raw)
	if err != nil {
		if err == frontmatter.ErrNotFound {
			return nil, nil, fmt.Errorf("no frontmatter found in content")
		}
		return nil, nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	if len(raw) == 0 {
		return nil, nil, fmt.Errorf("empty frontmatter")
	}

	fm := &ParsedFrontmatter{
		Fields: make(map[string]any),
	}

	for key, val := range raw {
		switch key {
		case "type":
			s, ok := val.(string)
			if !ok {
				return nil, nil, fmt.Errorf("field 'type' must be a string, got %T", val)
			}
			fm.Type = s
		case "title":
			s, ok := val.(string)
			if !ok {
				return nil, nil, fmt.Errorf("field 'title' must be a string, got %T", val)
			}
			fm.Title = s
		default:
			fm.Fields[key] = val
		}
	}

	return fm, body, nil
}

// ValidateType checks that the frontmatter type matches a known template.
func ValidateType(fm *ParsedFrontmatter, templates map[string]template.Template) error {
	if fm.Type == "" {
		return fmt.Errorf("missing required field 'type' in frontmatter")
	}
	if _, ok := templates[fm.Type]; !ok {
		return fmt.Errorf("unknown type '%s'. Create .akb/templates/%s.yaml first", fm.Type, fm.Type)
	}
	return nil
}

// ValidateTitle checks that the frontmatter has a non-empty title.
func ValidateTitle(fm *ParsedFrontmatter) error {
	if fm.Title == "" {
		return fmt.Errorf("missing required field 'title' in frontmatter")
	}
	return nil
}

// IsDraft returns true when a page is in draft state.
// Draft is the implicit default: if is_draft is absent or explicitly true,
// the page is a draft. Only an explicit false means approved.
func IsDraft(fields map[string]any) bool {
	if fields == nil {
		return true
	}
	v, ok := fields["is_draft"]
	if !ok {
		return true
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != "false"
	default:
		return true
	}
}
