package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Entry represents a single KB in the registry.
type Entry struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Created string `yaml:"created"`
}

// RegistryPath returns the path to the registry file (~/.config/agent-kb/registry.yaml).
func RegistryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "agent-kb", "registry.yaml"), nil
}

// Load reads the registry file and returns all entries.
// Returns an empty list if the file does not exist.
func Load(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var entries []Entry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}
	return entries, nil
}

// Save writes the entries to the registry file, creating parent directories as needed.
func Save(path string, entries []Entry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create registry directory: %w", err)
	}

	data, err := yaml.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}
	return nil
}

// AddEntry adds an entry to the registry file.
// Returns an error if an entry with the same name already exists.
func AddEntry(entry Entry) error {
	regPath, err := RegistryPath()
	if err != nil {
		return err
	}

	entries, err := Load(regPath)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.Name == entry.Name {
			return fmt.Errorf("KB %q already registered at %s", entry.Name, e.Path)
		}
	}

	entries = append(entries, entry)
	return Save(regPath, entries)
}

// FindByName looks up an entry by name in the registry.
func FindByName(name string) (*Entry, error) {
	regPath, err := RegistryPath()
	if err != nil {
		return nil, err
	}

	entries, err := Load(regPath)
	if err != nil {
		return nil, err
	}

	for i := range entries {
		if entries[i].Name == name {
			return &entries[i], nil
		}
	}
	return nil, nil
}
