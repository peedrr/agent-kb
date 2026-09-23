# cmd/akb/

**Parent:** `./AGENTS.md`

## OVERVIEW

CLI commands using Cobra framework. Each subcommand is a separate file. 25+ commands total.

## FILES

| Command | File | Description |
|---------|------|-------------|
| init | `init.go` | Initialize new KB |
| write | `write.go` | Write page (stdin → frontmatter → CEL validation → git → search → links) |
| read | `read.go` | Read page content |
| append | `append.go` | Append content to existing page |
| delete | `delete.go` | Delete page from KB, SQLite, index.md, log.md |
| list | `list.go` | List pages in kb/ |
| search | `search.go` | FTS5 BM25-ranked search |
| index | `index.go` | Manage kb/index.md (show/add/remove/rebuild) |
| log | `log.go` | Manage kb/log.md (show/append) |
| links | `links.go` | Show outbound/inbound/broken/ambiguous links |
| backlinks | `links.go` | Inbound links only |
| orphans | `links.go` | Pages with zero inbound links |
| status | `status.go` | KB name, path, page count, git status |
| discover | `discover.go` | List nearby knowledge bases (read-only, `--json`) |
| lint | `lint.go` | Run all lint checks; `RunLint(ctx)` helper extracted |
| approve | `approve.go` | Strip annotations and provenance markers, set `is_draft: false`, commit, then reindex the search index and link graph in one transaction |
| skill | `skill.go` | `skill install` — extract embedded skill |
| template | `template.go` | `template get <name>` (Writer/Mockup/Maintainer views, `--example` validates mockup), `template list` |
| templates write | `templates_write.go` | Overwrite protection (diff + page count), mockup reuse, `--force`, stale mockup rejection |
| template delete | `template_delete.go` | Delete template with impact analysis (page count), `--force`, auto-lint after |
| raw | `raw.go` | Parent command for raw namespace |
| raw write | `raw_write.go` | Write raw file with SHA-256 manifest |
| raw read | `raw_read.go` | Read raw file content |
| raw list | `raw_list.go` | List raw files |
| raw status | `raw_status.go` | Drift detection (exit 0/1/2) |
| raw sync | `raw_sync.go` | Reconcile manifest with filesystem |
| raw delete | `raw_delete.go` | Delete raw file, scan dependent pages |

## CONVENTIONS

- `noCommit` flag: `RootCmd.PersistentFlags().BoolVar(&noCommit, "no-commit", false, ...)`
- Commands validate stdin with `os.Stdin.Stat()` checking `ModeCharDevice`
- KB root resolved via `path.ResolveKB(kbFlag)` at the start of each command; a missing or non-KB selection is a usage error (exit 2)
- DB opened via `db.OpenKB(kbRoot)` for search/linkgraph/lint operations
- `approve` strips provenance markers via `markdown.StripProvenanceMarkers()`; `--all-drafts` for batch approval
- `raw delete` scans KB pages for frontmatter `sources` referencing the deleted file
- `write` supports `--append` (append to body) and `--frontmatter key=val` (partial updates)
- `write` runs CEL validations before file write; exits 1 on validation failure, 2 on internal error
- `search` supports `--tag`, `--type`, `--after` dimensional filters
- `list` `--json` includes `is_draft` field
- Template name validation: names must match `^[a-zA-Z0-9_-]+$` (path traversal prevention)
- `--force` on template write bypasses existence warning only, never mockup validation
- Existing mockups reused on overwrite if `--pass`/`--fail` flags omitted
- `template get <name>` returns Writer View (schema + requirements only)
- `template get <name> --example` returns `_pass.md` content
- `template get <name> --full` returns complete YAML with CEL rules
- `templates write` validates CEL syntax, pass mockup passes all validations, fail mockup fails at least one

## KEY DEPENDENCIES

```go
// Typical command structure
kbRoot, err := path.ResolveKB(kbFlag)   // --kb or AKB_KB selection
dbConn, err := db.OpenKB(kbRoot)        // Open search DB
store := storage.NewGitProvider(...)    // Git-backed storage
searcher := search.NewSQLiteFTS5Searcher(dbConn)
updater := linkgraph.NewSQLiteLinkGraph(dbConn)
engine := lint.NewLintEngine()          // Add checkers, then Run()
```
