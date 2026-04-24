package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/registry"
)

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the default knowledge base",
	Long:  `Set the default KB by name from the registry.`,
	Example: `  # Switch to a different KB
  akb use my-other-kb`,
	Args: cobra.ExactArgs(1),
	RunE: runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	regPath, err := registry.Path()
	if err != nil {
		return fmt.Errorf("get registry path: %w", err)
	}

	reg, err := registry.Load(regPath)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	var matches []registry.Entry
	for _, e := range reg.Entries {
		if e.Name == name {
			matches = append(matches, e)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("kb '%s' not found in registry; use 'akb registry' to list available kbs", name)
	}

	if len(matches) > 1 {
		return fmt.Errorf("multiple kbs named '%s' found in registry; manual fix required", name)
	}

	if err := registry.SetDefault(name); err != nil {
		return fmt.Errorf("set default registry: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Default KB set to '%s' at %s\n", name, matches[0].Path) //nolint:errcheck // stdout write failure non-critical
	return nil
}
