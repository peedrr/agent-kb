package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	yaml "github.com/goccy/go-yaml"
	"github.com/google/cel-go/common/types"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/peedrr/agent-kb/internal/cel"
	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/frontmatter"
	"github.com/peedrr/agent-kb/internal/linkgraph"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/search"
	"github.com/peedrr/agent-kb/internal/storage"
	"github.com/peedrr/agent-kb/internal/template"
)

var writeAppend bool
var writeFrontmatter []string

var writeCmd = &cobra.Command{
	Use:   "write <path>",
	Short: "Write a page to the knowledge base",
	Long:  `Read content from stdin, validate frontmatter, and write to the type-derived directory in the KB.`,
	Example: `  # Write a simple note
  echo "# My Note

Content here" | akb write my-note.md

  # Write with frontmatter (type and title required)
  cat <<'EOF' | akb write adr/use-sqlite-search.md
  ---
  type: adr
  title: Use SQLite for search
  ---
  We decided to use SQLite because it provides FTS5 full-text search.
  EOF

  # Write to a type-derived directory (e.g., notes/)
  cat <<'EOF' | akb write notes/idea.md
  ---
  type: note
  title: A new idea
  ---
  This note goes into the notes/ directory.
  EOF`,
	Args: cobra.ExactArgs(1),
	RunE: runWrite,
}

func init() {
	writeCmd.Flags().BoolVar(&writeAppend, "append", false, "append stdin to existing page body")
	writeCmd.Flags().StringArrayVar(&writeFrontmatter, "frontmatter", nil, "update frontmatter field(s) as key=value")
}

func runWrite(_ *cobra.Command, args []string) error {
	inputPath := args[0]

	// Resolve KB root
	kbRoot, err := path.ResolveKB(kbFlag)
	if err != nil {
		return fmt.Errorf("resolve knowledge base: %w", err)
	}

	dbConn, err := db.OpenKB(kbRoot)
	if err != nil {
		if isMissingDB(err) {
			return fmt.Errorf("run `akb index rebuild` to create the search index")
		}
		return fmt.Errorf("open search database: %w", err)
	}
	defer dbConn.Close() //nolint:errcheck // DB close error non-critical on command exit

	_, err = config.Load(filepath.Join(kbRoot, ".akb", ".akb.yaml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Load templates
	templates, err := template.LoadTemplates(filepath.Join(kbRoot, ".akb", "templates"))
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	// Create storage provider (needed for old_page and write)
	store := storage.NewGitProvider(kbRoot, noCommit)

	// Create CEL environment
	celEnv, err := cel.NewEnv()
	if err != nil {
		return &internalError{err: fmt.Errorf("CEL engine error: %w", err)}
	}

	var fm *frontmatter.ParsedFrontmatter
	var body []byte
	var writeContent []byte
	var fullPath string
	var relPath string
	var oldPage map[string]any

	if len(writeFrontmatter) > 0 {
		if writeAppend {
			return &usageError{msg: "--frontmatter and --append cannot be used together"}
		}

		// Validate .md extension
		if !strings.HasSuffix(inputPath, ".md") {
			return &usageError{msg: "page filename must end with .md"}
		}

		// Reject raw/ prefix
		if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
			return &usageError{msg: "use `akb raw write`"}
		}

		// Strip kb/ prefix
		cleanPath := strings.TrimPrefix(inputPath, "kb/")

		// Guard: block write to managed files
		base := filepath.Base(cleanPath)
		if base == "index.md" {
			return &usageError{msg: "cannot write index.md; use 'akb index add' to update"}
		}
		if base == "log.md" {
			return &usageError{msg: "cannot write log.md; it is a managed file"}
		}

		// Reject .. and absolute paths via ResolveKBPath
		_, err = path.ResolveKBPath(kbRoot, inputPath)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}

		// Hold the repository lock from the read of the existing page through the
		// commit and the search and link-graph updates that follow it: the update
		// is assembled from the page as it is at commit time, so concurrent
		// updates of the same page cannot overwrite each other.
		pageLock, err := storage.LockRepo(kbRoot)
		if err != nil {
			return fmt.Errorf("lock repository: %w", err)
		}
		defer pageLock.Release()

		// Try to find the existing file
		candidates := []string{filepath.Join(kbRoot, "kb", cleanPath)}
		for _, tmpl := range templates {
			if tmpl.Dir != "" {
				candidates = append(candidates, filepath.Join(kbRoot, "kb", tmpl.Dir, cleanPath))
			}
		}

		var existingContent []byte
		var found bool
		for _, candidate := range candidates {
			existingContent, err = os.ReadFile(candidate)
			if err == nil {
				fullPath = candidate
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("page does not exist; use 'akb write' without --frontmatter to create")
		}

		fm, body, err = frontmatter.Parse(existingContent)
		if err != nil {
			return fmt.Errorf("parse existing frontmatter: %w", err)
		}

		if err := frontmatter.ValidateType(fm, templates); err != nil {
			return fmt.Errorf("validate type: %w", err)
		}
		if err := frontmatter.ValidateTitle(fm); err != nil {
			return fmt.Errorf("validate title: %w", err)
		}

		// Compute relative path for output and indexing
		relPath = filepath.ToSlash(strings.TrimPrefix(fullPath, kbRoot+string(filepath.Separator)))

		// Build old_page from pre-modification state
		oldPage, err = cel.BuildOldPage(relPath, store)
		if err != nil {
			return &internalError{err: fmt.Errorf("CEL engine error: %w", err)}
		}

		for _, arg := range writeFrontmatter {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid frontmatter argument %q: expected key=value", arg)
			}
			key := parts[0]
			value := parts[1]
			if key == "type" {
				fm.Type = value
			} else if key == "title" {
				fm.Title = value
			} else {
				fm.Fields[key] = value
			}
		}

		// Validate type after potential update
		if err := frontmatter.ValidateType(fm, templates); err != nil {
			return fmt.Errorf("validate type: %w", err)
		}

		_, ok := templates[fm.Type]
		if !ok {
			return fmt.Errorf("unknown type %q", fm.Type)
		}

		allFields := map[string]any{
			"type":  fm.Type,
			"title": fm.Title,
		}
		for k, v := range fm.Fields {
			allFields[k] = v
		}
		yamlBytes, err := yaml.Marshal(allFields)
		if err != nil {
			return fmt.Errorf("re-serialize frontmatter: %w", err)
		}
		writeContent = []byte("---\n" + string(yamlBytes) + "---\n" + string(body))
	} else {
		// Read from stdin — error if stdin is a TTY
		stat, err := os.Stdin.Stat()
		if err != nil {
			return fmt.Errorf("check stdin: %w", err)
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return &usageError{msg: "input required: pipe content to stdin"}
		}

		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}

		if writeAppend {
			// Validate .md extension
			if !strings.HasSuffix(inputPath, ".md") {
				return &usageError{msg: "page filename must end with .md"}
			}

			// Reject raw/ prefix
			if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
				return &usageError{msg: "use `akb raw write`"}
			}

			// Strip kb/ prefix
			cleanPath := strings.TrimPrefix(inputPath, "kb/")

			// Guard: block write to managed files
			base := filepath.Base(cleanPath)
			if base == "index.md" {
				return &usageError{msg: "cannot write index.md; use 'akb index add' to update"}
			}
			if base == "log.md" {
				return &usageError{msg: "cannot write log.md; it is a managed file"}
			}

			// Reject .. and absolute paths via ResolveKBPath
			_, err = path.ResolveKBPath(kbRoot, inputPath)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			// Hold the repository lock from the read of the existing page through
			// the commit and the search and link-graph updates that follow it, so
			// concurrent appends cannot overwrite each other's content.
			pageLock, err := storage.LockRepo(kbRoot)
			if err != nil {
				return fmt.Errorf("lock repository: %w", err)
			}
			defer pageLock.Release()

			fullPath = filepath.Join(kbRoot, "kb", cleanPath)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				return fmt.Errorf("page does not exist; use 'akb write' without --append to create")
			}

			existingContent, err := os.ReadFile(fullPath)
			if err != nil {
				return fmt.Errorf("read existing page: %w", err)
			}

			fm, body, err = frontmatter.Parse(existingContent)
			if err != nil {
				return fmt.Errorf("parse existing frontmatter: %w", err)
			}

			// Validate type
			if err := frontmatter.ValidateType(fm, templates); err != nil {
				return fmt.Errorf("validate type: %w", err)
			}

			// Validate title
			if err := frontmatter.ValidateTitle(fm); err != nil {
				return fmt.Errorf("validate title: %w", err)
			}

			// Compute relative path
			relPath = filepath.ToSlash(filepath.Join("kb", cleanPath))

			// Build old_page from pre-modification state
			oldPage, err = cel.BuildOldPage(relPath, store)
			if err != nil {
				return &internalError{err: fmt.Errorf("CEL engine error: %w", err)}
			}

			newBody := string(body) + "\n" + string(stdinContent)

			allFields := map[string]any{
				"type":  fm.Type,
				"title": fm.Title,
			}
			for k, v := range fm.Fields {
				allFields[k] = v
			}
			yamlBytes, err := yaml.Marshal(allFields)
			if err != nil {
				return fmt.Errorf("re-serialize frontmatter: %w", err)
			}
			writeContent = []byte("---\n" + string(yamlBytes) + "---\n" + newBody)
			body = []byte(newBody)
		} else {
			// Parse frontmatter from stdin
			fm, body, err = frontmatter.Parse(stdinContent)
			if err != nil {
				return fmt.Errorf("parse frontmatter: %w", err)
			}

			// Validate type
			if err := frontmatter.ValidateType(fm, templates); err != nil {
				return fmt.Errorf("validate type: %w", err)
			}

			// Validate title
			if err := frontmatter.ValidateTitle(fm); err != nil {
				return fmt.Errorf("validate title: %w", err)
			}

			writeContent = stdinContent
			needsReserialize := false

			if _, ok := fm.Fields["created"]; !ok {
				fm.Fields["created"] = time.Now().UTC().Format(time.RFC3339)
				needsReserialize = true
			}

			if val, ok := fm.Fields["is_draft"]; ok {
				if val == true || val == "true" {
					delete(fm.Fields, "is_draft")
					needsReserialize = true
				}
			}

			if needsReserialize {
				allFields := map[string]any{
					"type":  fm.Type,
					"title": fm.Title,
				}
				for k, v := range fm.Fields {
					allFields[k] = v
				}
				yamlBytes, err := yaml.Marshal(allFields)
				if err != nil {
					return fmt.Errorf("re-serialize frontmatter: %w", err)
				}
				writeContent = []byte("---\n" + string(yamlBytes) + "---\n" + string(body))
			}

			// Resolve type-derived directory
			tmpl, ok := templates[fm.Type]
			if !ok {
				// Should not happen after ValidateType, but be safe
				return fmt.Errorf("unknown type %q", fm.Type)
			}
			dirFromType := tmpl.Dir

			// Validate .md extension
			if !strings.HasSuffix(inputPath, ".md") {
				return &usageError{msg: "page filename must end with .md"}
			}

			// Reject raw/ prefix
			if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
				return &usageError{msg: "use `akb raw write`"}
			}

			// Strip kb/ prefix
			cleanPath := strings.TrimPrefix(inputPath, "kb/")

			// Guard: block write to managed files
			base := filepath.Base(cleanPath)
			if base == "index.md" {
				return &usageError{msg: "cannot write index.md; use 'akb index add' to update"}
			}
			if base == "log.md" {
				return &usageError{msg: "cannot write log.md; it is a managed file"}
			}

			// Strip type-dir prefix if it matches the type's Dir
			if dirFromType != "" {
				typeDirPrefix := dirFromType + "/"
				cleanPath = strings.TrimPrefix(cleanPath, typeDirPrefix)
			}

			// Reject .. and absolute paths via ResolveKBPath
			_, err = path.ResolveKBPath(kbRoot, inputPath)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			// Construct final path
			if dirFromType != "" {
				fullPath = filepath.Join(kbRoot, "kb", dirFromType, cleanPath)
			} else {
				fullPath = filepath.Join(kbRoot, "kb", cleanPath)
			}

			// Compute relative path for output and indexing
			if dirFromType != "" {
				relPath = filepath.Join("kb", dirFromType, cleanPath)
			} else {
				relPath = filepath.Join("kb", cleanPath)
			}
			relPath = filepath.ToSlash(relPath)

			// old_page is nil for new stdin writes
			oldPage = nil
		}
	}

	// CEL validation
	tmpl, ok := templates[fm.Type]
	if !ok {
		return fmt.Errorf("unknown type %q", fm.Type)
	}

	// Build page map from post-modification state
	md := goldmark.New()
	var astDoc ast.Node
	astDoc = md.Parser().Parse(text.NewReader(body))
	page := cel.BuildPage(relPath, fm, body, astDoc, body)

	// Run CEL validations
	var validationErrors []cel.ValidationError
	for _, rule := range tmpl.Validations {
		prg, err := cel.CompileRule(celEnv, rule.Rule)
		if err != nil {
			return &internalError{err: fmt.Errorf("CEL engine error: compile rule %s: %w", rule.ID, err)}
		}
		var oldPageAny any
		if oldPage != nil {
			oldPageAny = oldPage
		}
		result, err := cel.Evaluate(context.Background(), prg, map[string]any{
			"page":     page,
			"old_page": oldPageAny,
			"now":      time.Now(),
		}, cel.MaxCostLimit)
		if err != nil {
			return &internalError{err: fmt.Errorf("CEL engine error: evaluate rule %s: %w", rule.ID, err)}
		}
		if result != types.True {
			validationErrors = append(validationErrors, cel.ValidationError{
				RuleID:   rule.ID,
				Message:  rule.Expect,
				Line:     0,
				Severity: "error",
			})
		}
	}

	if len(validationErrors) > 0 {
		for _, ve := range validationErrors {
			fmt.Fprintln(os.Stderr, ve.Error())
		}
		return validationFailure{}
	}

	// Extract tags and summary from frontmatter fields
	tags := search.ExtractTags(fm.Fields)
	summary := search.ExtractSummary(fm.Fields)

	// Hold the repository lock from the page write through its commit and the
	// search and link-graph updates that follow it.
	repoLock, err := storage.LockRepo(kbRoot)
	if err != nil {
		return fmt.Errorf("lock repository: %w", err)
	}
	defer repoLock.Release()

	// Write content
	ctx := context.Background()
	if err := store.Write(ctx, fullPath, writeContent); err != nil {
		return fmt.Errorf("write page: %w", err)
	}

	searcher := search.NewSQLiteFTS5Searcher(dbConn)
	if err := searcher.IndexPage(ctx, relPath, fm.Title, string(body), tags, summary, fm.Type); err != nil {
		return fmt.Errorf("index page: %w", err)
	}

	updater := linkgraph.NewSQLiteLinkGraph(dbConn)
	if err := updater.UpdatePageLinks(ctx, relPath, string(writeContent)); err != nil {
		return fmt.Errorf("update links: %w", err)
	}

	// Output
	if len(writeFrontmatter) > 0 {
		fmt.Printf("Updated frontmatter for %s\n", relPath)
	} else if writeAppend {
		fmt.Printf("Appended to %s\n", relPath)
	} else {
		fmt.Printf("Written to %s\n", relPath)
		fmt.Printf("Don't forget to update the index! `akb index add %s <summary>`\n", relPath)
	}

	return nil
}
