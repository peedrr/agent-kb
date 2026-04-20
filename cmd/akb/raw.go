package main

import (
	"github.com/spf13/cobra"
)

var rawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Manage raw files",
	Long:  `Manage raw files in the knowledge base with SHA-256 manifest tracking.`,
}

func init() {
	rawCmd.AddCommand(rawWriteCmd)
	rawCmd.AddCommand(rawReadCmd)
	rawCmd.AddCommand(rawListCmd)
	rawCmd.AddCommand(rawStatusCmd)
	rawCmd.AddCommand(rawSyncCmd)
	rawCmd.AddCommand(rawDeleteCmd)
}
