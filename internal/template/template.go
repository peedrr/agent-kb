// Package template loads and validates typed page templates.
package template

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Field defines a single template field with optional enum values.
type Field struct {
	Name string
	Enum []string
}

// UnmarshalYAML parses a Field from YAML, supporting simple string or enum map.
func (f *Field) UnmarshalYAML(unmarshal func(any) error) error {
	var raw any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case string:
		f.Name = v
		return nil
	case map[string]any:
		for key, val := range v {
			f.Name = key
			if m, ok := val.(map[string]any); ok {
				if enumRaw, ok := m["enum"]; ok {
					if enumSlice, ok := enumRaw.([]any); ok {
						for _, item := range enumSlice {
							if s, ok := item.(string); ok {
								f.Enum = append(f.Enum, s)
							}
						}
					}
				}
			}
			return nil
		}
		return fmt.Errorf("empty map for field")
	default:
		return fmt.Errorf("field must be a string or map, got %T", raw)
	}
}

// Template defines a typed page template with fields and required keys.
type Template struct {
	Name         string  `yaml:"name"`
	Dir          string  `yaml:"dir"`
	Required     []Field `yaml:"required"`
	Optional     []Field `yaml:"optional"`
	BodyTemplate string  `yaml:"body"`
	filename     string
}

// DefaultFS holds the embedded default templates.
//
//go:embed embedded/*.yaml
var DefaultFS embed.FS

// DefaultTemplates is the FS used to load embedded default templates.
var DefaultTemplates fs.FS

func init() {
	var err error
	DefaultTemplates, err = fs.Sub(DefaultFS, "embedded")
	if err != nil {
		panic(fmt.Sprintf("failed to create sub FS: %v", err))
	}
}

// LoadTemplates reads template YAML files from a directory.
func LoadTemplates(dir string) (map[string]Template, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]Template{}, nil
		}
		return nil, fmt.Errorf("read template directory: %w", err)
	}

	templates := make(map[string]Template)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path) //nolint:gosec // path constructed from validated dir and fs.DirEntry
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		var tmpl Template
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		if tmpl.Name == "" {
			return nil, fmt.Errorf("missing name field in %s", entry.Name())
		}

		if existing, ok := templates[tmpl.Name]; ok {
			return nil, fmt.Errorf("duplicate template name %q in %s and %s", tmpl.Name, entry.Name(), existing.filename)
		}

		tmpl.filename = entry.Name()
		templates[tmpl.Name] = tmpl
	}

	return templates, nil
}

// LoadTemplatesFromFS reads template YAML files from an fs.FS.
func LoadTemplatesFromFS(fsys fs.FS) (map[string]Template, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read template directory: %w", err)
	}

	templates := make(map[string]Template)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		data, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		var tmpl Template
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		if tmpl.Name == "" {
			return nil, fmt.Errorf("missing name field in %s", entry.Name())
		}

		if existing, ok := templates[tmpl.Name]; ok {
			return nil, fmt.Errorf("duplicate template name %q in %s and %s", tmpl.Name, entry.Name(), existing.filename)
		}

		tmpl.filename = entry.Name()
		templates[tmpl.Name] = tmpl
	}

	return templates, nil
}

// AllowedFields returns all declared field names (required + optional) for this template,
// plus the built-in 'is_draft' field.
func (t Template) AllowedFields() []string {
	seen := make(map[string]struct{})
	var fields []string

	for _, f := range t.Required {
		if _, ok := seen[f.Name]; !ok {
			seen[f.Name] = struct{}{}
			fields = append(fields, f.Name)
		}
	}
	for _, f := range t.Optional {
		if _, ok := seen[f.Name]; !ok {
			seen[f.Name] = struct{}{}
			fields = append(fields, f.Name)
		}
	}

	// is_draft is always allowed as a built-in system field
	if _, ok := seen["is_draft"]; !ok {
		fields = append(fields, "is_draft")
	}

	return fields
}

// CopyDefaults extracts embedded default templates to a directory.
func CopyDefaults(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	entries, err := fs.ReadDir(DefaultTemplates, ".")
	if err != nil {
		return fmt.Errorf("read embedded templates: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		targetPath := filepath.Join(targetDir, entry.Name())
		if _, err := os.Stat(targetPath); err == nil {
			continue
		}

		data, err := fs.ReadFile(DefaultTemplates, entry.Name())
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", entry.Name(), err)
		}

		if err := os.WriteFile(targetPath, data, 0600); err != nil {
			return fmt.Errorf("write %s: %w", targetPath, err)
		}
	}

	return nil
}
