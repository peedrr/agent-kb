package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var noCommit bool

// usageClassificationInstalled reports whether the cobra error conversions
// were installed, so Execute installs them once per process.
var usageClassificationInstalled bool

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
	RootCmd.AddCommand(skillCmd)
	RootCmd.AddCommand(templateCmd)
}

// Execute runs the CLI and returns the failure of the command it ran, wrapped
// with the step that failed. Cobra reports invocation mistakes — unknown
// flags, wrong argument counts, unknown commands — as plain errors, so Execute
// converts them into the typed usage error first; main reports that with a
// usage: prefix and the fault exit code.
func Execute() error {
	installUsageErrorClassification()
	RootCmd.SetVersionTemplate("akb {{.Version}}\n")
	if err := RootCmd.Execute(); err != nil {
		return commandFailure{err: err}
	}
	return nil
}

// installUsageErrorClassification converts the errors cobra raises on its own —
// flag parse failures, positional argument validation, and unknown commands —
// into the typed usage error.
func installUsageErrorClassification() {
	if usageClassificationInstalled {
		return
	}
	usageClassificationInstalled = true

	RootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &usageError{msg: err.Error()}
	})
	wrapArgsErrors(RootCmd)
	// Cobra validates the arguments of a root command that has no validator of
	// its own and reports the first argument as an unknown command. Installing
	// an equivalent check keeps that report identical while typing the mistake.
	RootCmd.Args = rootArgs
}

// wrapArgsErrors types the positional-argument errors of cmd and its
// subcommands as usage errors.
func wrapArgsErrors(cmd *cobra.Command) {
	if cmd.Args != nil {
		validate := cmd.Args
		cmd.Args = func(c *cobra.Command, args []string) error {
			if err := validate(c, args); err != nil {
				return &usageError{msg: err.Error()}
			}
			return nil
		}
	}
	for _, child := range cmd.Commands() {
		wrapArgsErrors(child)
	}
}

// rootArgs reports a positional argument of the root command that names no
// subcommand.
func rootArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	return &usageError{msg: unknownCommandMessage(cmd, args[0])}
}

// unknownCommandMessage builds the report cobra makes for an unknown command,
// including the suggestions it offers for a mistyped subcommand.
func unknownCommandMessage(cmd *cobra.Command, arg string) string {
	msg := fmt.Sprintf("unknown command %q for %q", arg, cmd.CommandPath())
	if cmd.DisableSuggestions {
		return msg
	}
	if cmd.SuggestionsMinimumDistance <= 0 {
		cmd.SuggestionsMinimumDistance = 2
	}
	suggestions := cmd.SuggestionsFor(arg)
	if len(suggestions) == 0 {
		return msg
	}

	msg += "\n\nDid you mean this?\n"
	for _, suggestion := range suggestions {
		msg += fmt.Sprintf("\t%v\n", suggestion)
	}
	return msg
}
