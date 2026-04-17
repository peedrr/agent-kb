# PROJECT KNOWLEDGE BASE

**Generated:** 2026-04-17
**Commit:** b25a580
**Branch:** master
**Status:** Prototype/MVP - subject to change

## OVERVIEW

Agent KB (akb): Go CLI for managing a git-tracked knowledge base with SQLite index, YAML frontmatter, and typed page templates.

## STRUCTURE

```
agent-kb/
├── akb             # Compiled binary (dev convenience)
├── akb.test        # Test binary
├── cmd/akb/        # CLI commands (Cobra)
├── internal/       # Core packages
│   ├── config/     # .akb.yaml handling
│   ├── db/         # SQLite schema
│   ├── frontmatter/# YAML frontmatter parsing
│   ├── index/      # Markdown index management
│   ├── linkgraph/  # Page link tracking (NoOp stub)
│   ├── log/        # kb/log.md append-only log
│   ├── path/       # KB path resolution & validation
│   ├── registry/   # ~/.config/agent-kb/registry.yaml
│   ├── search/     # Search interface (NoOp stub)
│   ├── storage/    # GitProvider + FilesystemProvider
│   └── template/   # Typed page templates (YAML)
└── test/           # Integration tests + testdata
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add command | `cmd/akb/` | New subcommand = new file |
| KB path logic | `internal/path/path.go` | KBRoot(), ResolveKBPath() |
| Page write flow | `cmd/akb/write.go` | stdin → frontmatter → git |
| Index management | `internal/index/index.go` | kb/index.md parsing/rendering |
| Template validation | `internal/template/template.go` | Typed pages (agent, adr, etc.) |
| Git integration | `internal/storage/git.go` | Auto-commit with merge conflict checks |
| Config format | `internal/config/config.go` | YAML .akb.yaml |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| RootCmd | Cobra.Command | cmd/akb/root.go:9 | Base CLI command |
| version | string | cmd/akb/main.go:8 | CLI version |
| StorageProvider | interface | internal/storage/provider.go:6 | Write/Read/Delete/Exists |
| Searcher | interface | internal/search/searcher.go:20 | IndexPage/Search/RebuildIndex |
| LinkGraphUpdater | interface | internal/linkgraph/updater.go:7 | UpdatePageLinks |
| Template | struct | internal/template/template.go:50 | Page type definition |
| IndexEntry | struct | internal/index/index.go:12 | Page in index |
| Entry | struct | internal/registry/registry.go:13 | Registry KB entry |
| Config | struct | internal/config/config.go:12 | .akb.yaml |

## CONVENTIONS (THIS PROJECT)

- **KB root**: Located by walking up for `.akb/` directory marker
- **Paths**: Always relative to KB root; `kb/` prefix stripped
- **Managed files**: `index.md`, `log.md` cannot be written directly
- **Raw access**: `raw/` prefix → separate storage; use `akb raw` commands
- **Commits**: Git commits auto-created with `akb: write/delete <path>` messages
- **Merge conflicts**: Blocked; must resolve before any write/delete
- **Git config**: Auto-sets `user.name=akb` / `user.email=akb@local` if unset

## ANTI-PATTERNS (THIS PROJECT)

- Do NOT write to `index.md` or `log.md` directly
- Do NOT use raw path for KB pages
- Do NOT upgrade `modernc.org/libc` independently (go.mod comment)

## COMMANDS

```bash
go build -o akb ./cmd/akb     # Build
go test ./...                 # Unit tests
./akb.test -test.v            # Integration tests
nix develop                   # Dev shell (Go, gopls, delve, golangci-lint)
```

## NOTES

- Prototype stage: APIs may change
- Search/linkgraph are NoOp stubs (interfaces exist, implementations don't)
- Templates embedded in binary via `//go:embed embedded/*.yaml`
