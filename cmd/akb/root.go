package main

import (
	"github.com/spf13/cobra"
)

var noCommit bool

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
	RootCmd.PersistentFlags().BoolVar(&noCommit, "no-commit", false, "skip git commit")
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(writeCmd)
	RootCmd.AddCommand(readCmd)
	RootCmd.AddCommand(deleteCmd)
	RootCmd.AddCommand(appendCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(indexCmd)
	RootCmd.AddCommand(logCmd)
	RootCmd.AddCommand(searchCmd)
	RootCmd.AddCommand(linksCmd)
	RootCmd.AddCommand(backlinksCmd)
	RootCmd.AddCommand(orphansCmd)
	RootCmd.AddCommand(rawCmd)
	RootCmd.AddCommand(lintCmd)
	RootCmd.AddCommand(registryCmd)
	RootCmd.AddCommand(useCmd)
	RootCmd.AddCommand(approveCmd)
	RootCmd.AddCommand(staleCmd)
}

func Execute() error {
	RootCmd.SetVersionTemplate("akb {{.Version}}\n")
	return RootCmd.Execute()
}
