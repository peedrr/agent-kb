package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListSkills(t *testing.T) {
	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
	if len(skills) != 1 || skills[0] != "kb-management" {
		t.Errorf("ListSkills() = %v, want [kb-management]", skills)
	}
}

func TestGetSkill(t *testing.T) {
	skill, err := GetSkill("kb-management")
	if err != nil {
		t.Fatalf("GetSkill(kb-management) error = %v", err)
	}
	if skill.Name != "kb-management" {
		t.Errorf("skill.Name = %q, want %q", skill.Name, "kb-management")
	}
	if len(skill.Files) != 8 {
		t.Errorf("len(skill.Files) = %d, want 8", len(skill.Files))
	}
	if _, ok := skill.Files["SKILL.md"]; !ok {
		t.Error("skill.Files missing SKILL.md")
	}
	if _, ok := skill.Files["references/INIT.md"]; !ok {
		t.Error("skill.Files missing references/INIT.md")
	}
}

func TestGetSkillNotFound(t *testing.T) {
	_, err := GetSkill("nonexistent")
	if err == nil {
		t.Error("GetSkill(nonexistent) expected error, got nil")
	}
}

func TestInstallSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "akb-skill-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp error = %v", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	installDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(installDir, 0750); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	err = InstallSkill("kb-management", installDir)
	if err != nil {
		t.Fatalf("InstallSkill error = %v", err)
	}

	skillDir := filepath.Join(installDir, "kb-management")
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		t.Fatalf("ReadDir %s error = %v", skillDir, err)
	}
	if len(entries) < 1 {
		t.Error("skill directory is empty")
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Error("SKILL.md not installed")
	}

	refDir := filepath.Join(skillDir, "references")
	if _, err := os.Stat(refDir); os.IsNotExist(err) {
		t.Error("references directory not installed")
	}
}

func TestInstallSkillDuplicate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "akb-skill-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp error = %v", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	installDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(installDir, 0750); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	if err := InstallSkill("kb-management", installDir); err != nil {
		t.Fatalf("first InstallSkill error = %v", err)
	}

	err = InstallSkill("kb-management", installDir)
	if err == nil {
		t.Error("second InstallSkill expected error, got nil")
	}
}

func TestInstallSkillNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "akb-skill-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp error = %v", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal

	err = InstallSkill("nonexistent", tmpDir)
	if err == nil {
		t.Error("InstallSkill(nonexistent) expected error, got nil")
	}
}
