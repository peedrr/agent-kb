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

// Schema defines the frontmatter schema for a template.
type Schema struct {
	Frontmatter map[string]FieldSchema `yaml:"frontmatter"`
}

// FieldSchema defines the type and constraints for a single frontmatter field.
type FieldSchema struct {
	Type     string   `yaml:"type"`
	Required bool     `yaml:"required"`
	Enum     []string `yaml:"enum,omitempty"`
}

// ValidationRule defines a CEL validation rule for a template.
type ValidationRule struct {
	ID          string `yaml:"id"`
	Rule        string `yaml:"rule"`
	Requirement string `yaml:"requirement,omitempty"`
	Expect      string `yaml:"expect"`
}

// LintRule defines a lint rule for a template.
type LintRule struct {
	ID       string `yaml:"id"`
	Rule     string `yaml:"rule"`
	Severity string `yaml:"severity"` // "warning" | "error"
	Expect   string `yaml:"expect"`
}

// Template defines a typed page template with schema, validations, and lint rules.
type Template struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Dir         string           `yaml:"dir"`
	Schema      Schema           `yaml:"schema"`
	Validations []ValidationRule `yaml:"validations"`
	LintRules   []LintRule       `yaml:"lint_rules"`
	filename    string
}

// DefaultFS holds the embedded default templates.
//
//go:embed embedded/*
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

func detectOldFormat(raw map[string]any, filename string) error {
	for _, key := range []string{"required", "optional", "body"} {
		if _, ok := raw[key]; ok {
			return fmt.Errorf("parse %s: Template format has changed. Please update to the new schema.", filename)
		}
	}
	return nil
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

		var raw map[string]any
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		if err := detectOldFormat(raw, entry.Name()); err != nil {
			return nil, err
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
