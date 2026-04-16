package template

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Field struct {
	Name string
	Enum []string
}

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

type Template struct {
	Name         string  `yaml:"name"`
	Dir          string  `yaml:"dir"`
	Required     []Field `yaml:"required"`
	Optional     []Field `yaml:"optional"`
	BodyTemplate string  `yaml:"body"`
	filename     string
}

//go:embed embedded/*.yaml
var DefaultFS embed.FS

var DefaultTemplates fs.FS

func init() {
	var err error
	DefaultTemplates, err = fs.Sub(DefaultFS, "embedded")
	if err != nil {
		panic(fmt.Sprintf("failed to create sub FS: %v", err))
	}
}

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
		data, err := os.ReadFile(path)
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

func CopyDefaults(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
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

		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", targetPath, err)
		}
	}

	return nil
}
