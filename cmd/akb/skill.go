package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/peedrr/agent-kb/internal/skill"
	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage skills",
	Long:  `Install skills from the embedded skill repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var skillInstallCmd = &cobra.Command{
	Use:   "install --location <path> <name>",
	Short: "Install a skill to a directory",
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillInstall,
}

var skillLocation string

func init() {
	skillInstallCmd.Flags().StringVar(&skillLocation, "location", "", "target directory for skill installation")
	skillCmd.AddCommand(skillInstallCmd)
}

func runSkillInstall(cmd *cobra.Command, args []string) error {
	name := args[0]

	if !cmd.Flags().Changed("location") {
		return fmt.Errorf("--location is required. Example: akb skill install --location ~/.agents/skills kb-management")
	}

	if err := skill.InstallSkill(name, skillLocation); err != nil {
		if errors.Is(err, skill.ErrAlreadyInstalled) {
			return fmt.Errorf("Skill '%s' already installed at %s. Remove it first.", name, filepath.Join(skillLocation, name))
		}
		if errors.Is(err, skill.ErrSkillNotFound) {
			return fmt.Errorf("Unknown skill '%s'.", name)
		}
		return err
	}

	fmt.Printf("Skill '%s' installed to %s\n", name, filepath.Join(skillLocation, name))
	return nil
}