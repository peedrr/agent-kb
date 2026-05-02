// Package registry manages the ~/.config/agent-kb/registry.yaml file.
package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
)

// Entry represents a single KB in the registry.
type Entry struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Created string `yaml:"created"`
}

// Registry represents the full registry file with a default KB.
type Registry struct {
	Default string  `yaml:"default"`
	Entries []Entry `yaml:"entries"`
}

// Path returns the path to the registry file (~/.config/agent-kb/registry.yaml).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "agent-kb", "registry.yaml"), nil
}

// Load reads the registry file and returns a Registry struct.
// Returns an empty Registry if the file does not exist.
// Supports migration from old flat format ([]Entry) to new Registry format.
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is caller-provided registry file path
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Registry{Default: "", Entries: nil}, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}

	return parseRegistry(data)
}

func parseRegistry(data []byte) (*Registry, error) {
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err == nil && !looksLikeOldFormat(data, reg) {
		return &reg, nil
	}

	var entries []Entry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}

	return &Registry{Default: "", Entries: entries}, nil
}

func looksLikeOldFormat(data []byte, reg Registry) bool {
	if len(data) == 0 {
		return false
	}
	if data[0] == '-' {
		return true
	}
	return reg.Default == "" && reg.Entries == nil
}

// Save writes the registry to the registry file, creating parent directories as needed.
func Save(path string, reg *Registry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create registry directory: %w", err)
	}

	data, err := yaml.Marshal(reg)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}
	return nil
}

// AddEntry adds an entry to the registry file.
// Returns an error if an entry with the same name already exists.
func AddEntry(entry Entry) error {
	regPath, err := Path()
	if err != nil {
		return err
	}

	reg, err := Load(regPath)
	if err != nil {
		return err
	}

	for _, e := range reg.Entries {
		if e.Name == entry.Name {
			return fmt.Errorf("KB %q already registered at %s", entry.Name, e.Path)
		}
	}

	reg.Entries = append(reg.Entries, entry)
	return Save(regPath, reg)
}

// FindByName looks up an entry by name in the registry.
func FindByName(name string) (*Entry, error) {
	regPath, err := Path()
	if err != nil {
		return nil, err
	}

	reg, err := Load(regPath)
	if err != nil {
		return nil, err
	}

	for i := range reg.Entries {
		if reg.Entries[i].Name == name {
			return &reg.Entries[i], nil
		}
	}
	return nil, nil
}

// SetDefault sets the default KB by name.
// Returns an error if the name is not found in the registry.
func SetDefault(name string) error {
	regPath, err := Path()
	if err != nil {
		return err
	}

	reg, err := Load(regPath)
	if err != nil {
		return err
	}

	found := false
	for _, e := range reg.Entries {
		if e.Name == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("KB %q not found in registry", name)
	}

	reg.Default = name
	return Save(regPath, reg)
}

// GetDefault returns the default KB entry.
// Returns an error if no default is set.
func GetDefault() (Entry, error) {
	regPath, err := Path()
	if err != nil {
		return Entry{}, err
	}

	reg, err := Load(regPath)
	if err != nil {
		return Entry{}, err
	}

	if reg.Default == "" {
		return Entry{}, errors.New("no default KB set")
	}

	for _, e := range reg.Entries {
		if e.Name == reg.Default {
			return e, nil
		}
	}

	return Entry{}, fmt.Errorf("default KB %q not found in registry", reg.Default)
}
