package main

import (
	"github.com/spf13/cobra"
)

var rawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Manage raw files",
	Long:  `Manage raw files in the knowledge base with SHA-256 manifest tracking.`,
}

// Stub subcommands — implemented in later phases
var rawStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show manifest status (not yet implemented)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var rawSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize manifest with filesystem (not yet implemented)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var rawDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a raw file (not yet implemented)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rawCmd.AddCommand(rawWriteCmd)
	rawCmd.AddCommand(rawReadCmd)
	rawCmd.AddCommand(rawListCmd)
	rawCmd.AddCommand(rawStatusCmd)
	rawCmd.AddCommand(rawSyncCmd)
	rawCmd.AddCommand(rawDeleteCmd)
}
