package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/markdown"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/peedrr/agent-kb/internal/storage"
)

var approveAllDrafts bool

var approveCmd = &cobra.Command{
	Use:   "approve <path>",
	Short: "Approve a page by removing draft status and annotations",
	Long:  `Read a page, strip olw-auto annotations and provenance markers, set is_draft to false, and write it back.`,
	Example: `  # Approve a draft page
  akb approve notes/my-draft.md`,
	Args: func(_ *cobra.Command, args []string) error {
		if approveAllDrafts {
			if len(args) > 0 {
				return errors.New("path argument not allowed with --all-drafts")
			}
			return nil
		}
		if len(args) != 1 {
			return errors.New("requires exactly 1 arg(s), only received 0")
		}
		return nil
	},
	RunE: runApprove,
}

func init() {
	approveCmd.Flags().BoolVar(&approveAllDrafts, "all-drafts", false, "approve all draft pages")
}

func runApprove(_ *cobra.Command, args []string) error {
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	ctx := context.Background()

	dbConn, err := db.OpenKB(kbRoot)
	if err != nil {
		if isMissingDB(err) {
			return fmt.Errorf("run `akb index rebuild` to create the search index")
		}
		return fmt.Errorf("open search database: %w", err)
	}
	defer dbConn.Close() //nolint:errcheck // DB close error non-critical on command exit

	if approveAllDrafts {
		return approveAllDraftPages(ctx, dbConn, kbRoot)
	}

	inputPath := args[0]
	fullPath, err := path.ResolveKBPath(kbRoot, inputPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	approved, err := approvePage(ctx, dbConn, kbRoot, fullPath, inputPath)
	if err != nil {
		return err
	}
	if !approved {
		fmt.Printf("Page '%s' is already approved\n", inputPath)
		return nil
	}
	fmt.Printf("Approved '%s'\n", inputPath)
	return nil
}

func approvePage(ctx context.Context, dbConn *sql.DB, kbRoot, fullPath, inputPath string) (bool, error) {
	// Hold the repository lock from the read of the page through its commit and
	// the search and link-graph updates, so the approved body is the body that
	// gets committed.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return false, fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	store := storage.NewGitProvider(kbRoot, noCommit)

	content, err := store.Read(ctx, fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
			return false, fmt.Errorf("page '%s' not found. Use 'akb list' to see available pages", inputPath)
		}
		return false, fmt.Errorf("read page: %w", err)
	}

	fm, body, err := frontmatter.Parse(content)
	if err != nil {
		return false, fmt.Errorf("parse frontmatter: %w", err)
	}

	if !frontmatter.IsDraft(fm.Fields) {
		return false, nil
	}

	bodyStr := markdown.StripAnnotations(string(body))
	bodyStr = markdown.StripProvenanceMarkers(bodyStr)

	fm.Fields["is_draft"] = false

	allFields := map[string]any{
		"type":  fm.Type,
		"title": fm.Title,
	}
	for k, v := range fm.Fields {
		allFields[k] = v
	}

	yamlBytes, err := yaml.Marshal(allFields)
	if err != nil {
		return false, fmt.Errorf("re-serialize frontmatter: %w", err)
	}

	finalContent := []byte("---\n" + string(yamlBytes) + "---\n" + bodyStr)

	if err := store.WriteWithCommitMsg(ctx, fullPath, finalContent, fmt.Sprintf("akb: approve %s", inputPath)); err != nil {
		return false, fmt.Errorf("write page: %w", err)
	}

	relPath, err := filepath.Rel(kbRoot, fullPath)
	if err != nil {
		return false, fmt.Errorf("compute relative path: %w", err)
	}
	relPath = filepath.ToSlash(relPath)

	tags := search.ExtractTags(fm.Fields)
	summary := search.ExtractSummary(fm.Fields)

	// The search index and the link graph describe the same approved page, so
	// both steps share one transaction: a failure in either leaves both at
	// their pre-approval state instead of one step behind the other.
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin index transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // deferred rollback is no-op after successful commit

	searcher := search.NewSQLiteFTS5Searcher(dbConn)
	if err := searcher.IndexPageTx(ctx, tx, relPath, fm.Title, bodyStr, tags, summary, fm.Type); err != nil {
		return false, fmt.Errorf("index page: %w", err)
	}

	updater := linkgraph.NewSQLiteLinkGraph(dbConn)
	if err := updater.UpdatePageLinksTx(ctx, tx, relPath, string(finalContent)); err != nil {
		return false, fmt.Errorf("update links: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit index transaction: %w", err)
	}

	return true, nil
}

// draftCandidate is a page the batch approve validated and is ready to approve:
// the page's full path and the base-relative path that names it in the approval
// commit and in the approval report.
type draftCandidate struct {
	fullPath string
	relPath  string
}

// approveAllDraftPages approves every draft page of the base in two phases. The
// first phase collects every draft candidate and validates it, and reports the
// first failure before anything is approved. Only a candidate set that passed
// validation reaches the second phase, which approves each candidate through
// approvePage. A rejected candidate therefore cannot leave the base
// half-approved with approval commits already landed.
func approveAllDraftPages(ctx context.Context, dbConn *sql.DB, kbRoot string) error {
	candidates, err := collectDraftCandidates(ctx, kbRoot)
	if err != nil {
		return err
	}

	var approvedCount int
	for _, candidate := range candidates {
		approved, err := approvePage(ctx, dbConn, kbRoot, candidate.fullPath, candidate.relPath)
		if err != nil {
			return err
		}
		if approved {
			approvedCount++
		}
	}

	fmt.Printf("Approved %d drafts\n", approvedCount)
	return nil
}

// collectDraftCandidates walks the pages of the base and returns the drafts
// among them, in walk order. Every page it reads must pass the checks the
// approval of that page relies on: the path stays inside the base once the
// filesystem follows symlinks, the file is readable, and its frontmatter
// parses. The first page that fails any check fails the whole collection, so
// the caller approves nothing and no page is left half-approved. A page that is
// already approved is not a candidate and is skipped.
func collectDraftCandidates(ctx context.Context, kbRoot string) ([]draftCandidate, error) {
	kbDir := filepath.Join(kbRoot, "kb")
	store := storage.NewGitProvider(kbRoot, noCommit)

	var candidates []draftCandidate

	err := filepath.WalkDir(kbDir, func(fullPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(fullPath, ".md") {
			return nil
		}

		relPath, err := filepath.Rel(kbDir, fullPath)
		if err != nil {
			return fmt.Errorf("relativize %s: %w", fullPath, err)
		}
		if relPath == "index.md" || relPath == "log.md" {
			return nil
		}

		if err := path.AssertContained(kbRoot, fullPath); err != nil {
			return fmt.Errorf("check page path: %w", err)
		}

		content, err := store.Read(ctx, fullPath)
		if err != nil {
			return fmt.Errorf("read page %s: %w", relPath, err)
		}

		fm, _, err := frontmatter.Parse(content)
		if err != nil {
			return fmt.Errorf("parse frontmatter of %s: %w", relPath, err)
		}
		if !frontmatter.IsDraft(fm.Fields) {
			return nil
		}

		candidates = append(candidates, draftCandidate{fullPath: fullPath, relPath: relPath})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk kb directory: %w", err)
	}

	return candidates, nil
}
