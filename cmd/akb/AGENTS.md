# cmd/akb/

**Parent:** `./AGENTS.md`

## OVERVIEW

CLI commands using Cobra framework. Each subcommand is a separate file.

## FILES

| Command | File | Description |
|---------|------|-------------|
| init | `init.go` | Initialize new KB |
| write | `write.go` | Write page (stdin → frontmatter → git → search → links) |
| read | `read.go` | Read page content |
| delete | `delete.go` | Delete page |
| list | `list.go` | List pages |
| search | `search.go` | Full-text search via FTS5 |
| index | `index.go` | Manage kb/index.md |
| log | `log.go` | Append to kb/log.md |
| links | `links.go` | Show outbound/inbound links |
| backlinks | `links.go` | Backlinks command |
| orphans | `links.go` | Pages with no inbound links |
| append | `append.go` | Append content to page |
| status | `status.go` | Git status of KB |

## CONVENTIONS

- `noCommit` flag: `RootCmd.PersistentFlags().BoolVar(&noCommit, "no-commit", false, ...)`
- Commands validate stdin with `os.Stdin.Stat()` checking `ModeCharDevice`
- KB root resolved via `path.KBRoot()` at start of each command
- DB opened via `db.OpenKB(kbRoot)` for search/linkgraph operations

## KEY DEPENDENCIES

```go
// Typical command structure
kbRoot, err := path.KBRoot()           // Find KB
dbConn, err := db.OpenKB(kbRoot)       // Open search DB
store := storage.NewGitProvider(...)    // Git-backed storage
searcher := search.NewSQLiteFTS5Searcher(dbConn)
updater := linkgraph.NewSQLiteLinkGraph(dbConn)
```
