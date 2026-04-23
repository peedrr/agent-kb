package main

import (
	"fmt"

	"github.com/peedrr/agent-kb/internal/registry"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "List registered knowledge bases",
	Long:  `List all registered KBs with the active default marked with *.`,
	Args:  cobra.NoArgs,
	RunE:  runRegistry,
}

func runRegistry(cmd *cobra.Command, args []string) error {
	regPath, err := registry.RegistryPath()
	if err != nil {
		return err
	}

	reg, err := registry.Load(regPath)
	if err != nil {
		return err
	}

	if len(reg.Entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No KBs registered. Use 'akb init <name>' to create one.")
		return nil
	}

	for _, e := range reg.Entries {
		marker := " "
		if e.Name == reg.Default {
			marker = "*"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s\n", marker, e.Name, e.Path)
	}

	return nil
}
