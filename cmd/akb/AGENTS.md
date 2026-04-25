# cmd/akb/

**Parent:** `./AGENTS.md`

## OVERVIEW

CLI commands using Cobra framework. Each subcommand is a separate file. 25+ commands total.

## FILES

| Command | File | Description |
|---------|------|-------------|
| init | `init.go` | Initialize new KB |
| write | `write.go` | Write page (stdin → frontmatter → git → search → links) |
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
| lint | `lint.go` | Run all lint checks |
| registry | `registry.go` | List registered KBs |
| use | `use.go` | Set default KB in registry |
| approve | `approve.go` | Strip annotations, set `is_draft: false`, commit |
| stale | `stale.go` | Cross-KB freshness report |
| skill | `skill.go` | `skill install` — extract embedded skill |
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
- KB root resolved via `path.KBRoot()` at start of each command
- DB opened via `db.OpenKB(kbRoot)` for search/linkgraph/lint operations
- `approve` strips provenance markers via `markdown.StripProvenanceMarkers()`; `--all-drafts` for batch approval
- `stale` iterates ALL KBs in registry; `--json` and `--all` flags supported
- `raw delete` scans KB pages for frontmatter `sources` referencing the deleted file
- `write` supports `--append` (append to body) and `--frontmatter key=val` (partial updates)
- `search` supports `--tag`, `--type`, `--after` dimensional filters
- `list` `--json` includes `is_draft` field

## KEY DEPENDENCIES

```go
// Typical command structure
kbRoot, err := path.KBRoot()           // Find KB
dbConn, err := db.OpenKB(kbRoot)       // Open search DB
store := storage.NewGitProvider(...)    // Git-backed storage
searcher := search.NewSQLiteFTS5Searcher(dbConn)
updater := linkgraph.NewSQLiteLinkGraph(dbConn)
engine := lint.NewLintEngine()          // Add checkers, then Run()
```
