// Package config handles .akb.yaml configuration loading and validation.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	yaml "github.com/goccy/go-yaml"
)

// Config holds the .akb.yaml configuration.
type Config struct {
	Name    string `yaml:"name"`
	Created string `yaml:"created"`
}

// Load reads a Config from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is caller-provided config file path
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	content := string(data)
	if strings.Contains(content, "<<<<<<<") {
		return nil, errors.New("resolve merge conflicts in .akb.yaml before proceeding")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}

	if cfg.Name == "" {
		return nil, errors.New("missing required field 'name' in .akb.yaml")
	}

	return &cfg, nil
}

// Save writes a Config to the given path.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// Validate checks that a Config has required fields.
func Validate(cfg *Config) error {
	if cfg.Name == "" {
		return errors.New("name is required")
	}

	if cfg.Created == "" {
		return errors.New("created is required")
	}

	_, err := time.Parse(time.RFC3339, cfg.Created)
	if err != nil {
		_, err = time.Parse("2006-01-02", cfg.Created)
		if err != nil {
			return errors.New("created must be valid ISO-8601")
		}
	}

	return nil
}
