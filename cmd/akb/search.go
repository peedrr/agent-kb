package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/spf13/cobra"
)

var searchJSON bool

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search the knowledge base",
	Long:  `Perform a full-text search over the knowledge base using BM25 ranking.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "output results as JSON")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	kbRoot, err := path.ResolveKB()
	if err != nil {
		return err
	}

	sqlDB, err := db.OpenKB(kbRoot)
	if err != nil {
		if isMissingDB(err) {
			return fmt.Errorf("run `akb index rebuild` to create the search index")
		}
		return err
	}
	defer sqlDB.Close()

	searcher := search.NewSQLiteFTS5Searcher(sqlDB)

	ctx := context.Background()
	results, err := searcher.Search(ctx, query, search.SearchOptions{})
	if err != nil {
		return err
	}

	if searchJSON {
		return printSearchJSON(results)
	}

	return printSearchText(results)
}

func printSearchText(results []search.SearchResult) error {
	if len(results) == 0 {
		return nil
	}
	for _, r := range results {
		snippet := r.Snippet
		if len(snippet) > 80 {
			snippet = snippet[:77] + "..."
		}
		fmt.Printf("%s — %s (%s)\n", r.Path, r.Title, snippet)
	}
	return nil
}

type searchJSONResult struct {
	Path    string  `json:"path"`
	Title   string  `json:"title"`
	Summary string  `json:"summary"`
	Snippet string  `json:"snippet"`
	Rank    float64 `json:"rank"`
}

func printSearchJSON(results []search.SearchResult) error {
	out := make([]searchJSONResult, len(results))
	for i, r := range results {
		out[i] = searchJSONResult{
			Path:    r.Path,
			Title:   r.Title,
			Summary: r.Summary,
			Snippet: r.Snippet,
			Rank:    r.Rank,
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func isMissingDB(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "invalid schema") || strings.Contains(errStr, "no such file")
}
