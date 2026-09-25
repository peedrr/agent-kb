// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"github.com/spf13/cobra"
)

var rawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Manage raw files",
	Long:  `Manage raw files in the knowledge base with SHA-256 manifest tracking.`,
	Example: `  # List raw files
  akb raw list

  # Write a raw file
  echo '{"key": "value"}' | akb raw write config.json

  # Check for drift
  akb raw status`,
}

func init() {
	rawCmd.AddCommand(rawWriteCmd)
	rawCmd.AddCommand(rawReadCmd)
	rawCmd.AddCommand(rawListCmd)
	rawCmd.AddCommand(rawStatusCmd)
	rawCmd.AddCommand(rawSyncCmd)
	rawCmd.AddCommand(rawDeleteCmd)
}
