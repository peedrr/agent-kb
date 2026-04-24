# PROJECT KNOWLEDGE BASE

**Generated:** 2026-04-24
**Commit:** 4ee9316
**Branch:** master
**Status:** WIP/prototype - APIs subject to change

## OVERVIEW

Agent KB (akb): Go CLI for managing a git-tracked knowledge base with SQLite FTS5 search, SQLite link graph, YAML frontmatter, typed page templates, lint engine, and raw drift detection.

## STRUCTURE

```
agent-kb/
├── cmd/akb/        # CLI commands (Cobra, 25+ commands)
├── internal/       # Core packages
│   ├── config/     # .akb.yaml handling
│   ├── db/         # SQLite schema + WAL mode
│   ├── frontmatter/# YAML frontmatter parsing
│   ├── index/      # kb/index.md management
│   ├── linkgraph/  # SQLite link tracking (wikilinks)
│   ├── lint/       # 15 lint checkers + engine
│   ├── log/        # kb/log.md append-only log
│   ├── manifest/   # raw/files.log SHA-256 manifest
│   ├── markdown/   # Wikilink, annotation, provenance parsers
│   ├── path/       # KB path resolution & guards
│   ├── registry/   # ~/.config/agent-kb/registry.yaml
│   ├── search/     # SQLite FTS5 full-text search
│   ├── skill/      # Embedded skill management (//go:embed)
│   ├── storage/    # GitProvider (auto-commit)
│   └── template/   # Typed page templates
└── test/           # Integration tests (testscript)
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
| Annotation parser | `internal/markdown/annotation.go` | `<!-- olw-auto: ... -->` HTML comments |
| Provenance markers | `internal/markdown/provenance.go` | `^[inferred]`, `^[ambiguous]`, `^[extracted]` |
| Git integration | `internal/storage/git.go` | Auto-commit, merge conflict detection |
| DB schema | `internal/db/db.go` | documents, pages, links tables + FTS5 |
| Template validation | `internal/template/template.go` | Typed pages (note, adr, custom) |
| Config format | `internal/config/config.go` | YAML .akb.yaml |
| Lint engine | `internal/lint/engine.go` | LintEngine, LintChecker interface |
| Lint checks | `internal/lint/*.go` | 15 checkers (broken_links, orphans, freshness, etc.) |
| Manifest | `internal/manifest/manifest.go` | raw/files.log SHA-256 tracking |
| Skill install | `internal/skill/skill.go` | `//go:embed embedded/*` |
| Registry | `internal/registry/registry.go` | Multi-KB registry with default |
| Raw drift | `cmd/akb/raw_status.go` | `akb raw status` exits 0/1/2 |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| RootCmd | Cobra.Command | cmd/akb/root.go:9 | Base CLI command |
| version | string | cmd/akb/main.go:8 | CLI version (0.8.0) |
| StorageProvider | interface | internal/storage/provider.go:6 | Write/Read/Delete/Exists/List |
| GitProvider | struct | internal/storage/git.go:13 | Git-tracked file operations |
| Searcher | interface | internal/search/searcher.go:20 | IndexPage/Search/RebuildIndex |
| SQLiteFTS5Searcher | struct | internal/search/sqlite.go:18 | FTS5 implementation |
| LinkGraphUpdater | interface | internal/linkgraph/updater.go:7 | UpdatePageLinks/RemovePage |
| SQLiteLinkGraph | struct | internal/linkgraph/sqlite.go:24 | SQLite implementation |
| Link | struct | internal/linkgraph/sqlite.go:16 | Source, target, display, resolved |
| LintChecker | interface | internal/lint/engine.go:26 | Name()/Check() interface |
| LintEngine | struct | internal/lint/engine.go:49 | Orchestrates all checkers |
| LintIssue | struct | internal/lint/engine.go:13 | Type/Message/Path/Severity |
| LintReport | struct | internal/lint/engine.go:20 | Issues/PagesChecked/ByCheck |
| KB | struct | internal/lint/engine.go:31 | Lint context: root, linkgraph, templates, pages |
| PageData | struct | internal/lint/engine.go:39 | Parsed page with frontmatter + markers |
| Template | struct | internal/template/template.go:50 | Page type definition |
| IndexEntry | struct | internal/index/index.go:12 | Page in index |
| ParsedFrontmatter | struct | internal/frontmatter/frontmatter.go:11 | Type, Title, Fields, IsDraft |
| Wikilink | struct | internal/markdown/wikilink.go:8 | Target, Display, Heading |
| Annotation | struct | internal/markdown/annotation.go:8 | olw-auto HTML comment parser |
| ProvenanceMarker | struct | internal/markdown/provenance.go:9 | ^[type] marker parser |
| ManifestEntry | struct | internal/manifest/manifest.go:17 | Filename, SHA256, LastUpdated |
| ManifestManager | struct | internal/manifest/manifest.go:24 | Read/Write/Add/Remove/Update entries |
| Skill | struct | internal/skill/skill.go:15 | Name, Files map |
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
- **is_draft**: Auto-managed frontmatter field; new pages are implicit drafts; `akb approve` sets `is_draft: false`
- **Type enforcement**: All pages MUST declare `type` in frontmatter matching a template in `.akb/templates/`
- **Build output**: Always `-o bin/akb` (never project root)

## ANTI-PATTERNS (THIS PROJECT)

- Do NOT write to `index.md` or `log.md` directly (use `akb index add` or `akb log`)
- Do NOT use raw path for KB pages (use `akb write/read`)
- Do NOT upgrade `modernc.org/libc` independently (go.mod comment)
- Do NOT run concurrent DB operations (max 1 open connection)
- Do NOT add `--type` flag on `akb write` or `akb append` (type comes from frontmatter)
- Do NOT build binary in project root (use `bin/`)
- Do NOT add v2 features (MCP, Nix Flake, goreleaser, TUI, AI exports) in v1 code

## COMMANDS

```bash
go build -o bin/akb ./cmd/akb/  # Build (NEVER in project root)
go test ./...                   # Unit tests
go test ./test/ -test.v         # Integration tests (testscript)
golangci-lint run ./...         # Lint
nix develop                     # Dev shell (Go, gopls, delve, golangci-lint)
```

## NOTES

- WIP: APIs may change between commits
- Wikilink resolution: exact → namespace prefix → basename match
- FTS5 query escaping strips FTS5 operators (OR, AND, NOT) and special chars
- Templates embedded in binary via `//go:embed embedded/*.yaml`
- Skills embedded in binary via `//go:embed embedded/*`
- Integration tests use testscript framework (`.txt` files in testdata/)
- 15 lint checks: 5 structural + 6 template-driven + 4 semantic (provenance, freshness, confidence, summary_length)
- Provenance drift threshold: 0.20; freshness half-life: 30 days; freshness score threshold: 50.0
- Lint thresholds are hardcoded constants (not configurable via `.akb.yaml` in v1)
