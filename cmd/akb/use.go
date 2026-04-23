package main

import (
	"fmt"

	"github.com/peedrr/agent-kb/internal/registry"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the default knowledge base",
	Long:  `Set the default KB by name from the registry.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	regPath, err := registry.RegistryPath()
	if err != nil {
		return err
	}

	reg, err := registry.Load(regPath)
	if err != nil {
		return err
	}

	var matches []registry.Entry
	for _, e := range reg.Entries {
		if e.Name == name {
			matches = append(matches, e)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("KB '%s' not found in registry. Use 'akb registry' to list available KBs.", name)
	}

	if len(matches) > 1 {
		return fmt.Errorf("Multiple KBs named '%s' found in registry. Manual fix required.", name)
	}

	if err := registry.SetDefault(name); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Default KB set to '%s' at %s\n", name, matches[0].Path)
	return nil
}
