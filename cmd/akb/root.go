package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var noCommit bool

var RootCmd = &cobra.Command{
	Use:     "akb",
	Short:   "Agent Knowledge Base CLI",
	Long:    `A CLI tool for managing the Agent Knowledge Base.`,
	Version: version,
	Run: func(cmd *cobra.Command, _ []string) {
		_ = cmd.Help() //nolint:errcheck // help display failure is non-fatal
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
	RootCmd.AddCommand(skillCmd)
}

func Execute() error {
	RootCmd.SetVersionTemplate("akb {{.Version}}\n")
	if err := RootCmd.Execute(); err != nil {
		return fmt.Errorf("execute command: %w", err)
	}
	return nil
}
