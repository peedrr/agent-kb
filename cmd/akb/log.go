package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	kblog "github.com/peedrr/agent-kb/internal/log"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

var validOperations = []string{"ingest", "delete", "update", "lint", "query"}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Manage the knowledge base log",
	Long:  `View and append entries in the KB log (kb/log.md).`,
	Example: `  # Show all log entries
  akb log show

  # Append a new entry
  akb log append ingest "Imported pages from old wiki"`,
	Run: func(cmd *cobra.Command, _ []string) {
		_ = cmd.Help() //nolint:errcheck // help display failure is non-fatal
	},
}

var logShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show log entries",
	Long:  `Display log entries from kb/log.md. Supports filtering by type and limiting output.`,
	Example: `  # Show all log entries
  akb log show

  # Show last 10 entries
  akb log show --last 10

  # Filter by operation type
  akb log show --type ingest`,
	Args: cobra.NoArgs,
	RunE: runLogShow,
}

var logAppendCmd = &cobra.Command{
	Use:   "append <operation> <description>",
	Short: "Append a log entry",
	Long:  `Add a new entry to the KB log. Operation must be one of: ingest, delete, update, lint, query.`,
	Example: `  # Log an ingestion operation
  akb log append ingest "Imported 50 pages from old wiki"

  # Log with optional title
  akb log append update "Updated configuration structure" --title "Config refactor"`,
	Args: cobra.ExactArgs(2),
	RunE: runLogAppend,
}

var logShowLast int
var logShowType string
var logAppendTitle string

func init() {
	logShowCmd.Flags().IntVar(&logShowLast, "last", 0, "show last N entries (0 = all)")
	logShowCmd.Flags().StringVar(&logShowType, "type", "", "filter by operation type")

	logAppendCmd.Flags().StringVar(&logAppendTitle, "title", "", "optional title for the log entry")

	logCmd.AddCommand(logShowCmd)
	logCmd.AddCommand(logAppendCmd)
}

func runLogShow(cmd *cobra.Command, _ []string) error {
	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	entries, err := kblog.ReadLog(kbRoot)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "log.md") {
			return nil
		}
		return fmt.Errorf("read log: %w", err)
	}

	if logShowType != "" {
		entries = kblog.FilterByType(entries, logShowType)
	}

	if logShowLast > 0 {
		entries = kblog.FilterByLast(entries, logShowLast)
	}

	if len(entries) == 0 {
		return nil
	}

	rendered := kblog.RenderLog(entries)
	_, _ = fmt.Fprint(cmd.OutOrStdout(), rendered) //nolint:errcheck // stdout write failure non-critical
	return nil
}

func runLogAppend(_ *cobra.Command, args []string) error {
	operation := args[0]
	description := args[1]

	if !isValidOperation(operation) {
		return fmt.Errorf("invalid operation %q; must be one of: %s", operation, strings.Join(validOperations, ", "))
	}

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	if err := kblog.AppendLog(kbRoot, operation, description, logAppendTitle); err != nil {
		return fmt.Errorf("append log: %w", err)
	}

	logPath := filepath.Join(kbRoot, "kb", "log.md")
	data, err := os.ReadFile(logPath) //nolint:gosec // path validated by ResolveKBPath
	if err != nil {
		return fmt.Errorf("read log after append: %w", err)
	}

	store := storage.NewGitProvider(kbRoot, noCommit)
	ctx := context.Background()
	commitMsg := "akb: log append " + operation
	if err := store.WriteWithCommitMsg(ctx, logPath, data, commitMsg); err != nil {
		return fmt.Errorf("commit log: %w", err)
	}

	return nil
}

func isValidOperation(op string) bool {
	for _, valid := range validOperations {
		if op == valid {
			return true
		}
	}
	return false
}
