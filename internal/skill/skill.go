// Package skill manages embedded skills for the knowledge base.
package skill

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed embedded/*
var embeddedFS embed.FS

// Skill represents an embedded skill with its files.
type Skill struct {
	Name  string
	Files map[string][]byte
}

// ErrSkillNotFound is returned when a skill does not exist.
var ErrSkillNotFound = errors.New("skill not found")

// ErrAlreadyInstalled is returned when a skill is already installed.
var ErrAlreadyInstalled = errors.New("skill already installed")

// ListSkills returns the names of all embedded skills.
func ListSkills() ([]string, error) {
	entries, err := fs.ReadDir(embeddedFS, "embedded")
	if err != nil {
		return nil, fmt.Errorf("read embedded directory: %w", err)
	}

	var skills []string
	for _, entry := range entries {
		if entry.IsDir() {
			skills = append(skills, entry.Name())
		}
	}
	return skills, nil
}

// GetSkill returns a Skill by name.
func GetSkill(name string) (Skill, error) {
	skillPath := filepath.Join("embedded", name)

	_, err := fs.ReadDir(embeddedFS, skillPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Skill{}, fmt.Errorf("%w: %s", ErrSkillNotFound, name)
		}
		return Skill{}, fmt.Errorf("read skill directory %s: %w", name, err)
	}

	skill := Skill{
		Name:  name,
		Files: make(map[string][]byte),
	}

	err = fs.WalkDir(embeddedFS, skillPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		data, err := embeddedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		relPath, err := filepath.Rel(skillPath, path)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}
		skill.Files[relPath] = data
		return nil
	})

	if err != nil {
		return Skill{}, fmt.Errorf("walk skill directory: %w", err)
	}

	return skill, nil
}

// InstallSkill extracts an embedded skill to the given location.
func InstallSkill(name, location string) error {
	_, err := GetSkill(name)
	if err != nil {
		return err
	}

	targetDir := filepath.Join(location, name)
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("%w: %s already installed at %s", ErrAlreadyInstalled, name, targetDir)
	}

	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	skillPath := filepath.Join("embedded", name)
	if err := fs.WalkDir(embeddedFS, skillPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			relPath, err := filepath.Rel(skillPath, path)
			if err != nil {
				return fmt.Errorf("compute relative path: %w", err)
			}
			if relPath == "." {
				return nil
			}
			targetPath := filepath.Join(targetDir, relPath)
			return os.MkdirAll(targetPath, 0750)
		}

		data, err := embeddedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded file %s: %w", path, err)
		}

		relPath, err := filepath.Rel(skillPath, path)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}
		targetPath := filepath.Join(targetDir, relPath)
		if err := os.WriteFile(targetPath, data, 0600); err != nil {
			return fmt.Errorf("write file %s: %w", targetPath, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walk skill directory: %w", err)
	}
	return nil
}
