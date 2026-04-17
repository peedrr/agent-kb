package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/spf13/cobra"
)

var linksCmd = &cobra.Command{
	Use:   "links",
	Short: "Query the link graph",
	Long:  `Query the link graph: show outbound/broken/ambiguous links, backlinks, or orphan pages.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var linksShowCmd = &cobra.Command{
	Use:   "show <path>",
	Short: "Show outbound, broken, and ambiguous links for a page",
	Long:  `Show outbound, broken, and ambiguous links for a page in the knowledge base.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLinksShow,
}

var backlinksCmd = &cobra.Command{
	Use:   "backlinks <path>",
	Short: "Show inbound links to a page",
	Long:  `Show pages that link to the given page in the knowledge base.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBacklinks,
}

var orphansCmd = &cobra.Command{
	Use:   "orphans",
	Short: "Show pages with zero inbound links",
	Long:  `Show pages in the knowledge base that have no inbound links from other pages.`,
	Args:  cobra.NoArgs,
	RunE:  runOrphans,
}

func init() {
	linksCmd.AddCommand(linksShowCmd)
	linksCmd.AddCommand(backlinksCmd)
	linksCmd.AddCommand(orphansCmd)
}

func openLinkGraphDB(kbRoot string) (*sql.DB, *linkgraph.SQLiteLinkGraph, error) {
	d, err := db.OpenKB(kbRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("run `akb index rebuild` to create the search index")
	}
	return d, linkgraph.NewSQLiteLinkGraph(d), nil
}

func resolvePagePath(kbRoot, inputPath string) (string, error) {
	cleanPath := strings.TrimPrefix(inputPath, "kb/")
	fullPath := filepath.Join(kbRoot, "kb", cleanPath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("page not found: %s", cleanPath)
	}

	return cleanPath, nil
}

func runLinksShow(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	d, g, err := openLinkGraphDB(kbRoot)
	if err != nil {
		return err
	}
	defer d.Close()

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

func runBacklinks(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	d, g, err := openLinkGraphDB(kbRoot)
	if err != nil {
		return err
	}
	defer d.Close()

	relPath, err := resolvePagePath(kbRoot, args[0])
	if err != nil {
		return err
	}

	ctx := context.Background()

	inbound, err := g.GetInboundLinks(ctx, relPath)
	if err != nil {
		return fmt.Errorf("query inbound links: %w", err)
	}

	for _, l := range inbound {
		fmt.Println(l.SourcePage)
	}

	return nil
}

func runOrphans(cmd *cobra.Command, args []string) error {
	kbRoot, err := path.KBRoot()
	if err != nil {
		return err
	}

	d, g, err := openLinkGraphDB(kbRoot)
	if err != nil {
		return err
	}
	defer d.Close()

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
