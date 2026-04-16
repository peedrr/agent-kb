package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name    string `yaml:"name"`
	Created string `yaml:"created"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	if strings.Contains(content, "<<<<<<<") {
		return nil, errors.New("resolve merge conflicts in .akb.yaml before proceeding")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Name == "" {
		return nil, errors.New("missing required field 'name' in .akb.yaml")
	}

	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

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
