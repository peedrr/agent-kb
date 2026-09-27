// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package config handles akb.yaml configuration loading and validation.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

// Values of the `versioning` key: how the base records its history.
const (
	// VersioningGit versions the base in git: every mutation is committed.
	VersioningGit = "git"
	// VersioningNone versions nothing: mutations write their files and stop.
	VersioningNone = "none"
)

// Config holds the akb.yaml configuration.
type Config struct {
	Name        string `yaml:"name"`
	Created     string `yaml:"created"`
	Description string `yaml:"description,omitempty"`
	// Versioning is the base's versioning mode. A base whose config predates
	// the key is versioned in git.
	Versioning string `yaml:"versioning,omitempty"`
	// GitAuthor and GitEmail are the commit identity the base supplies. They
	// hold the akb default when nothing else resolved an identity at init, and
	// let a base cloned to a machine without a git identity commit.
	GitAuthor string `yaml:"git-author,omitempty"`
	GitEmail  string `yaml:"git-email,omitempty"`
}

// Load reads a Config from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is caller-provided config file path
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	content := string(data)
	if strings.Contains(content, "<<<<<<<") {
		return nil, errors.New("resolve merge conflicts in akb.yaml before proceeding")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}

	// A config written before the key existed names no mode; git is the mode
	// those bases were created in, so they keep working unchanged.
	if cfg.Versioning == "" {
		cfg.Versioning = VersioningGit
	}

	if cfg.Name == "" {
		return nil, errors.New("missing required field 'name' in akb.yaml")
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
