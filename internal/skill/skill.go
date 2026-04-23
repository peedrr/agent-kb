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

type Skill struct {
	Name  string
	Files map[string][]byte
}

var ErrSkillNotFound = errors.New("skill not found")

var ErrAlreadyInstalled = errors.New("skill already installed")

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

func InstallSkill(name, location string) error {
	_, err := GetSkill(name)
	if err != nil {
		return err
	}

	targetDir := filepath.Join(location, name)
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("%w: %s already installed at %s", ErrAlreadyInstalled, name, targetDir)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	skillPath := filepath.Join("embedded", name)
	return fs.WalkDir(embeddedFS, skillPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			relPath, err := filepath.Rel(skillPath, path)
			if err != nil {
				return err
			}
			if relPath == "." {
				return nil
			}
			targetPath := filepath.Join(targetDir, relPath)
			return os.MkdirAll(targetPath, 0755)
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
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("write file %s: %w", targetPath, err)
		}

		return nil
	})
}