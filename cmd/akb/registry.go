package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/registry"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "List registered knowledge bases",
	Long:  `List all registered KBs with the active default marked with *.`,
	Example: `  # List all registered KBs
  akb registry`,
	Args: cobra.NoArgs,
	RunE: runRegistry,
}

func runRegistry(cmd *cobra.Command, _ []string) error {
	regPath, err := registry.Path()
	if err != nil {
		return fmt.Errorf("get registry path: %w", err)
	}

	reg, err := registry.Load(regPath)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	if len(reg.Entries) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No KBs registered. Use 'akb init <name>' to create one.") //nolint:errcheck // stdout write failure non-critical
		return nil
	}

	for _, e := range reg.Entries {
		marker := " "
		if e.Name == reg.Default {
			marker = "*"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s\n", marker, e.Name, e.Path) //nolint:errcheck // stdout write failure non-critical
	}

	return nil
}
