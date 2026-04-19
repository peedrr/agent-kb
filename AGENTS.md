# PROJECT KNOWLEDGE BASE

**Generated:** 2026-04-19
**Commit:** 3d471c8
**Branch:** master
**Status:** WIP/prototype - APIs subject to change

## OVERVIEW

Agent KB (akb): Go CLI for managing a git-tracked knowledge base with SQLite FTS5 search, SQLite link graph, YAML frontmatter, and typed page templates.

## STRUCTURE

```
agent-kb/
├── cmd/akb/        # CLI commands (Cobra)
├── internal/       # Core packages
│   ├── config/     # .akb.yaml handling
│   ├── db/         # SQLite schema + WAL mode
│   ├── frontmatter/# YAML frontmatter parsing
│   ├── index/      # kb/index.md management
│   ├── linkgraph/  # SQLite link tracking (wikilinks)
│   ├── log/        # kb/log.md append-only log
│   ├── markdown/   # Wikilink parser ([[...]])
│   ├── path/       # KB path resolution & guards
│   ├── registry/   # ~/.config/agent-kb/registry.yaml
│   ├── search/     # SQLite FTS5 full-text search
│   ├── storage/    # GitProvider (auto-commit)
│   └── template/   # Typed page templates
└── test/          # Integration tests (testscript)
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add command | `cmd/akb/` | New subcommand = new file |
| KB path logic | `internal/path/path.go` | KBRoot(), ResolveKBPath() |
| Page write flow | `cmd/akb/write.go` | stdin → frontmatter → git → search → links |
| Index management | `internal/index/index.go` | kb/index.md parsing/rendering |
| Search | `internal/search/sqlite.go` | SQLite FTS5 with BM25 ranking |
| Link graph | `internal/linkgraph/sqlite.go` | 3-step wikilink resolution |
| Wikilink parser | `internal/markdown/wikilink.go` | Excludes code blocks, inline code, HTML comments |
| Git integration | `internal/storage/git.go` | Auto-commit, merge conflict detection |
| DB schema | `internal/db/db.go` | documents, pages, links tables + FTS5 |
| Template validation | `internal/template/template.go` | Typed pages (agent, adr, note) |
| Config format | `internal/config/config.go` | YAML .akb.yaml |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| RootCmd | Cobra.Command | cmd/akb/root.go:9 | Base CLI command |
| version | string | cmd/akb/main.go:8 | CLI version |
| StorageProvider | interface | internal/storage/provider.go:6 | Write/Read/Delete/Exists |
| GitProvider | struct | internal/storage/git.go:13 | Git-tracked file operations |
| Searcher | interface | internal/search/searcher.go:20 | IndexPage/Search/RebuildIndex |
| SQLiteFTS5Searcher | struct | internal/search/sqlite.go:18 | FTS5 implementation |
| LinkGraphUpdater | interface | internal/linkgraph/updater.go:7 | UpdatePageLinks |
| SQLiteLinkGraph | struct | internal/linkgraph/sqlite.go:24 | SQLite implementation |
| Link | struct | internal/linkgraph/sqlite.go:16 | Source, target, display, resolved |
| Template | struct | internal/template/template.go:50 | Page type definition |
| IndexEntry | struct | internal/index/index.go:12 | Page in index |
| ParsedFrontmatter | struct | internal/frontmatter/frontmatter.go:11 | Type, Title, Fields |
| Wikilink | struct | internal/markdown/wikilink.go:8 | Target, Display, Heading |
| Entry | struct | internal/registry/registry.go:13 | Registry KB entry |
| Config | struct | internal/config/config.go:12 | .akb.yaml |
| DB | struct | internal/db/db.go:12 | Wrapper around sql.DB |

## CONVENTIONS (THIS PROJECT)

- **KB root**: Located by walking up for `.akb/` directory marker
- **Paths**: Always relative to KB root; `kb/` prefix stripped
- **Managed files**: `index.md`, `log.md` cannot be written directly
- **Raw access**: `raw/` prefix → separate storage; use `akb raw` commands
- **Commits**: Git commits auto-created with `akb: write/delete <path>` messages
- **Merge conflicts**: Blocked; must resolve before any write/delete
- **Git config**: Auto-sets `user.name=akb` / `user.email=akb@local` if unset
- **DB path**: `.akb/search.db` with WAL mode, single connection

## ANTI-PATTERNS (THIS PROJECT)

- Do NOT write to `index.md` or `log.md` directly (use `akb index add` or `akb log`)
- Do NOT use raw path for KB pages (use `akb write/read`)
- Do NOT upgrade `modernc.org/libc` independently (go.mod comment)
- Do NOT run concurrent DB operations (max 1 open connection)

## COMMANDS

```bash
go build -o akb ./cmd/akb     # Build
go test ./...                 # Unit tests
./akb.test -test.v            # Integration tests
nix develop                   # Dev shell (Go, gopls, delve, golangci-lint)
```

## NOTES

- WIP: APIs may change between commits
- Wikilink resolution: exact → namespace prefix → basename match
- FTS5 query escaping strips FTS5 operators (OR, AND, NOT) and special chars
- Templates embedded in binary via `//go:embed embedded/*.yaml`
- Integration tests use testscript framework (`.txt` files in testdata/)
