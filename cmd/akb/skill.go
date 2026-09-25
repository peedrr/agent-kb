// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/skill"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage skills",
	Long:  `Install skills from the embedded skill repository.`,
	Example: `  # Install a skill
  akb skill install --location ~/.agents/skills kb-management`,
	Run: func(cmd *cobra.Command, _ []string) {
		_ = cmd.Help() //nolint:errcheck // help display failure is non-fatal
	},
}

var skillInstallCmd = &cobra.Command{
	Use:   "install --location <path> <name>",
	Short: "Install a skill to a directory",
	Example: `  # Install a skill to the default location
  akb skill install --location ~/.agents/skills kb-management`,
	Args: cobra.ExactArgs(1),
	RunE: runSkillInstall,
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
			return fmt.Errorf("skill %q already installed at %s. Remove it first", name, filepath.Join(skillLocation, name))
		}
		if errors.Is(err, skill.ErrSkillNotFound) {
			return fmt.Errorf("unknown skill %q", name)
		}
		return fmt.Errorf("install skill: %w", err)
	}

	fmt.Printf("Skill '%s' installed to %s\n", name, filepath.Join(skillLocation, name))
	return nil
}
