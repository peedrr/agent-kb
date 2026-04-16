package main

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "akb",
	Short:   "Agent Knowledge Base CLI",
	Long:    `A CLI tool for managing the Agent Knowledge Base.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(statusCmd)
}

func Execute() error {
	RootCmd.SetVersionTemplate("akb {{.Version}}\n")
	return RootCmd.Execute()
}
