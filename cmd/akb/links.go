package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/path"
)

var backlinksJSON bool

var linksCmd = &cobra.Command{
	Use:   "links <path>",
	Short: "Show outbound, broken, and ambiguous links for a page",
	Long:  `Show outbound, broken, and ambiguous links for a page in the knowledge base.`,
	Example: `  # Show links for a page
  akb links notes/my-note.md`,
	Args: cobra.ExactArgs(1),
	RunE: runLinksShow,
}

var backlinksCmd = &cobra.Command{
	Use:   "backlinks <path>",
	Short: "Show inbound links to a page",
	Long:  `Show pages that link to the given page in the knowledge base.`,
	Example: `  # Show backlinks to a page
  akb backlinks notes/my-note.md`,
	Args: cobra.ExactArgs(1),
	RunE: runBacklinks,
}

var orphansCmd = &cobra.Command{
	Use:   "orphans",
	Short: "Show pages with zero inbound links",
	Long:  `Show pages in the knowledge base that have no inbound links from other pages.`,
	Example: `  # Show orphan pages
  akb orphans`,
	Args: cobra.NoArgs,
	RunE: runOrphans,
}

func init() {
	backlinksCmd.Flags().BoolVar(&backlinksJSON, "json", false, "output results as JSON")
}

func openLinkGraphDB(kbRoot string) (*sql.DB, *linkgraph.SQLiteLinkGraph, error) {
	d, err := db.OpenKB(kbRoot)
	if err != nil {
		if isMissingDB(err) {
			return nil, nil, fmt.Errorf("run `akb index rebuild` to create the search index")
		}
		return nil, nil, fmt.Errorf("open search database: %w", err)
	}
	return d, linkgraph.NewSQLiteLinkGraph(d), nil
}

func resolvePagePath(kbRoot, inputPath string) (string, error) {
	fullPath, err := path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	cleanPath := strings.TrimPrefix(inputPath, "kb/")

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("page not found: %s. Use 'akb list' to see available pages", cleanPath)
	}

	relPath := filepath.Join("kb", cleanPath)
	return filepath.ToSlash(relPath), nil
}

func runLinksShow(_ *cobra.Command, args []string) error {
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	d, g, err := openLinkGraphDB(kbRoot)
	if err != nil {
		return err
	}
	defer d.Close() //nolint:errcheck // DB close error non-critical on command exit

	relPath, err := resolvePagePath(kbRoot, args[0])
	if err != nil {
		return err
	}

	ctx := context.Background()

	outbound, err := g.GetOutboundLinks(ctx, relPath)
	if err != nil {
		return fmt.Errorf("query outbound links: %w", err)
	}

	allBroken, err := g.GetBrokenLinks(ctx)
	if err != nil {
		return fmt.Errorf("query broken links: %w", err)
	}

	allAmbiguous, err := g.GetAmbiguousLinks(ctx)
	if err != nil {
		return fmt.Errorf("query ambiguous links: %w", err)
	}

	var pageBroken []linkgraph.Link
	for _, l := range allBroken {
		if l.SourcePage == relPath {
			pageBroken = append(pageBroken, l)
		}
	}

	var pageAmbiguous []linkgraph.Link
	for _, l := range allAmbiguous {
		if l.SourcePage == relPath {
			pageAmbiguous = append(pageAmbiguous, l)
		}
	}

	fmt.Println("Outbound:")
	for _, l := range outbound {
		if l.ResolvedTo != "" && !strings.HasPrefix(l.ResolvedTo, "AMBIGUOUS") {
			fmt.Printf("  %s -> %s\n", l.RawTarget, l.ResolvedTo)
		}
	}

	fmt.Println("Broken:")
	for _, l := range pageBroken {
		fmt.Printf("  %s\n", l.RawTarget)
	}

	fmt.Println("Ambiguous:")
	for _, l := range pageAmbiguous {
		fmt.Printf("  %s -> %s\n", l.RawTarget, l.ResolvedTo)
	}

	return nil
}

// backlinkJSON is the per-link representation emitted by `akb backlinks --json`.
type backlinkJSON struct {
	SourcePage string `json:"source_page"`
	RawTarget  string `json:"raw_target"`
	Display    string `json:"display"`
	ResolvedTo string `json:"resolved_to"`
}

// backlinksResponse is the top-level shape of `akb backlinks --json`: always an
// object holding a "backlinks" array, empty when no page links back.
type backlinksResponse struct {
	Backlinks []backlinkJSON `json:"backlinks"`
}

func runBacklinks(_ *cobra.Command, args []string) error {
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	d, g, err := openLinkGraphDB(kbRoot)
	if err != nil {
		return err
	}
	defer d.Close() //nolint:errcheck // DB close error non-critical on command exit

	relPath, err := resolvePagePath(kbRoot, args[0])
	if err != nil {
		return err
	}

	ctx := context.Background()

	inbound, err := g.GetInboundLinks(ctx, relPath)
	if err != nil {
		return fmt.Errorf("query inbound links: %w", err)
	}

	if backlinksJSON {
		out := backlinksResponse{Backlinks: make([]backlinkJSON, 0, len(inbound))}
		for _, l := range inbound {
			out.Backlinks = append(out.Backlinks, backlinkJSON{
				SourcePage: l.SourcePage,
				RawTarget:  l.RawTarget,
				Display:    l.Display,
				ResolvedTo: l.ResolvedTo,
			})
		}
		data, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(inbound) == 0 {
		fmt.Println("No backlinks found")
		return nil
	}

	for _, l := range inbound {
		fmt.Println(l.SourcePage)
	}

	return nil
}

func runOrphans(_ *cobra.Command, _ []string) error {
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	d, g, err := openLinkGraphDB(kbRoot)
	if err != nil {
		return err
	}
	defer d.Close() //nolint:errcheck // DB close error non-critical on command exit

	ctx := context.Background()

	orphans, err := g.GetOrphans(ctx)
	if err != nil {
		return fmt.Errorf("query orphans: %w", err)
	}

	for _, p := range orphans {
		fmt.Println(p)
	}

	return nil
}
