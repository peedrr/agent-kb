// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/path"
)

var discoverJSON bool

var discoverCmd = &cobra.Command{
	Use:   "discover [dir]",
	Short: "List knowledge bases near a directory",
	Long: `Scan a directory's neighborhood for knowledge bases and list them nearest first.

The scan covers the directory itself, the directories below it to two levels, and
the same for every ancestor up to the home directory, skipping .git,
node_modules, vendor, and hidden directories. Discovery only reports what it
finds: it never selects a knowledge base, so address one of the listed paths
with --kb or AKB_KB.`,
	Example: `  # Discover knowledge bases near the working directory
  akb discover

  # Discover knowledge bases near another directory
  akb discover ../other-project

  # Machine-readable output
  akb discover --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDiscover,
}

func init() {
	discoverCmd.Flags().BoolVar(&discoverJSON, "json", false, "output the discovered knowledge bases as JSON")
}

func runDiscover(_ *cobra.Command, args []string) error {
	root, err := discoverRoot(args)
	if err != nil {
		return err
	}

	discovered := path.Discover(root)

	if discoverJSON {
		data, err := json.Marshal(discovered)
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(discovered) == 0 {
		fmt.Printf("no knowledge bases found near %s\n", root)
		return nil
	}

	fmt.Println(path.FormatDiscovered(discovered))
	return nil
}

// discoverRoot resolves the directory a discovery scan starts from: the
// command's argument when given, otherwise the working directory.
func discoverRoot(args []string) (string, error) {
	dir := ""
	if len(args) == 1 {
		dir = args[0]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
		dir = cwd
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve scan directory %q: %w", dir, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", &usageError{msg: fmt.Sprintf("scan directory %s: %v", abs, err)}
	}
	if !info.IsDir() {
		return "", &usageError{msg: fmt.Sprintf("scan directory %s: not a directory", abs)}
	}
	return abs, nil
}
