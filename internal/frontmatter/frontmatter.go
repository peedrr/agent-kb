// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package frontmatter parses and validates YAML frontmatter from markdown pages.
package frontmatter

import (
	"bytes"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"

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
	yamlFmt := frontmatter.Format{
		Name:      "YAML",
		Delim:     '-',
		Unmarshal: yaml.Unmarshal,
	}

	md := goldmark.New(
		goldmark.WithExtensions(&frontmatter.Extender{
			Formats: []frontmatter.Format{yamlFmt},
		}),
	)

	ctx := parser.NewContext()
	_ = md.Parser().Parse(text.NewReader(content), parser.WithContext(ctx))

	data := frontmatter.Get(ctx)
	if data == nil {
		return nil, nil, fmt.Errorf("no frontmatter found in content")
	}

	var raw map[string]any
	if err := data.Decode(&raw); err != nil {
		return nil, nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	if len(raw) == 0 {
		return nil, nil, fmt.Errorf("empty frontmatter")
	}

	body, err := extractBody(content)
	if err != nil {
		return nil, nil, err
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

func extractBody(content []byte) ([]byte, error) {
	idx := bytes.Index(content, []byte("\n"))
	if idx == -1 {
		return nil, fmt.Errorf("no frontmatter found in content")
	}

	firstLine := bytes.TrimRight(content[:idx], "\r")
	if !bytes.HasPrefix(firstLine, []byte("---")) {
		return nil, fmt.Errorf("no frontmatter found in content")
	}

	rest := content[idx+1:]
	for {
		nextIdx := bytes.Index(rest, []byte("\n"))
		var line []byte
		if nextIdx == -1 {
			line = rest
			rest = nil
		} else {
			line = rest[:nextIdx]
			rest = rest[nextIdx+1:]
		}

		trimmed := bytes.TrimRight(line, "\r")
		if bytes.Equal(trimmed, []byte("---")) {
			if rest == nil {
				return []byte{}, nil
			}
			return rest, nil
		}

		if nextIdx == -1 {
			break
		}
	}

	return nil, fmt.Errorf("no frontmatter found in content")
}

// ValidateType checks that the frontmatter type matches a known template.
func ValidateType(fm *ParsedFrontmatter, templates map[string]template.Template) error {
	if fm.Type == "" {
		return fmt.Errorf("missing required field 'type' in frontmatter")
	}
	if _, ok := templates[fm.Type]; !ok {
		return fmt.Errorf("unknown type '%s'. Author .akb/templates/%s.yaml (see the kb-management skill's TEMPLATE.md) or copy a showcase with `akb template list --examples`", fm.Type, fm.Type)
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
