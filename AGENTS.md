# PROJECT KNOWLEDGE BASE

**Generated:** 2026-05-10
**Commit:** 02e26d2
**Branch:** master
**Status:** v0.16.0 — template CRUD complete

## OVERVIEW

Agent KB (akb): Go CLI for managing a git-tracked knowledge base with SQLite FTS5 search, SQLite link graph, YAML frontmatter, typed page templates, CEL-driven validation, lint engine, and raw drift detection.

## STRUCTURE

```
agent-kb/
├── cmd/akb/        # CLI commands (Cobra, 25+ commands)
├── internal/       # Core packages
│   ├── cel/        # CEL expression engine (validation + lint)
│   ├── config/     # .akb.yaml handling
│   ├── db/         # SQLite schema + WAL mode
│   ├── frontmatter/# YAML frontmatter parsing (goldmark + goccy/go-yaml)
│   ├── index/      # kb/index.md management
│   ├── linkgraph/  # SQLite link tracking (wikilinks)
│   ├── lint/       # 9 lint checkers + engine
│   ├── log/        # kb/log.md append-only log
│   ├── manifest/   # raw/files.log SHA-256 manifest
│   ├── markdown/   # Wikilink, annotation, provenance, AST parsers
│   ├── path/       # KB path resolution & guards
│   ├── registry/   # ~/.config/agent-kb/registry.yaml
│   ├── search/     # SQLite FTS5 full-text search
│   ├── skill/      # Embedded skill management (//go:embed)
│   ├── storage/    # GitProvider (auto-commit)
│   └── template/   # Typed page templates (TemplateV2 with CEL rules)
├── test/           # Integration tests (testscript)
├── flake.nix       # Nix flake (dev shell + build package)
└── Makefile        # Build, test, lint targets
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add command | `cmd/akb/` | New subcommand = new file |
| KB path logic | `internal/path/path.go` | KBRoot(), ResolveKBPath() |
| Page write flow | `cmd/akb/write.go` | stdin → frontmatter → CEL validation → git → search → links |
| Index management | `internal/index/index.go` | kb/index.md parsing/rendering |
| Search | `internal/search/sqlite.go` | SQLite FTS5 with BM25 ranking |
| Link graph | `internal/linkgraph/sqlite.go` | 3-step wikilink resolution |
| Wikilink parser | `internal/markdown/wikilink.go` | Excludes code blocks, inline code, HTML comments |
| Annotation parser | `internal/markdown/annotation.go` | `<!-- olw-auto: ... -->` HTML comments |
| Provenance markers | `internal/markdown/provenance.go` | `^[inferred]`, `^[ambiguous]`, `^[extracted]` |
| Markdown AST | `internal/markdown/ast.go` | Goldmark AST flattener for CEL (headings, links, code blocks) |
| CEL engine | `internal/cel/engine.go` | Environment builder, rule compiler, evaluator with panic recovery |
| CEL page builder | `internal/cel/pagebuilder.go` | Builds `page`/`old_page` maps from frontmatter + AST |
| Template loader | `internal/template/template.go` | TemplateV2 with schema, validations, lint_rules |
| Template commands | `cmd/akb/template.go`, `templates_write.go`, `template_delete.go` | `template get/list`, `templates write`, `template delete` |
| Template delete | `cmd/akb/template_delete.go` | Impact analysis (page count), `--force`, path traversal prevention |
| Git integration | `internal/storage/git.go` | Auto-commit, merge conflict detection |
| DB schema | `internal/db/db.go` | documents, pages, links tables + FTS5 |
| Config format | `internal/config/config.go` | YAML .akb.yaml |
| Lint engine | `internal/lint/engine.go` | LintEngine, LintChecker interface |
| Lint checks | `internal/lint/*.go` | 9 checkers (broken_links, orphans, empty_pages, missing_frontmatter, index_consistency, citations, provenance, cel_lint, type_orphan) |
| CEL lint checker | `internal/lint/cel.go` | Evaluates template `lint_rules` with `now` injection |
| Manifest | `internal/manifest/manifest.go` | raw/files.log SHA-256 tracking |
| Skill install | `internal/skill/skill.go` | `//go:embed embedded/*` |
| Registry | `internal/registry/registry.go` | Multi-KB registry with default |
| Raw drift | `cmd/akb/raw_status.go` | `akb raw status` exits 0/1/2 |
| Nix flake | `flake.nix` | Dev shell (`nix develop`) + build package |
| Write append | `cmd/akb/write.go` | `--append` to append to existing page body |
| Write frontmatter | `cmd/akb/write.go` | `--frontmatter key=val` for partial updates |
| Batch approve | `cmd/akb/approve.go` | `--all-drafts` to approve all draft pages |
| Dimensional search | `cmd/akb/search.go` | `--tag`, `--type`, `--after` filters |
| List drafts | `cmd/akb/list.go` | `--json` includes `is_draft` field |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| RootCmd | Cobra.Command | cmd/akb/root.go:11 | Base CLI command |
| version | string | cmd/akb/main.go:8 | CLI version (injected at build via LDFLAGS) |
| Provider | interface | internal/storage/provider.go:6 | Write/Read/Delete/Exists/List |
| GitProvider | struct | internal/storage/git.go:15 | Git-tracked file operations |
| Searcher | interface | internal/search/searcher.go:27 | IndexPage/RemovePage/Search/RebuildIndex |
| SQLiteFTS5Searcher | struct | internal/search/sqlite.go:18 | FTS5 implementation |
| Updater | interface | internal/linkgraph/updater.go:7 | UpdatePageLinks/RemovePage |
| SQLiteLinkGraph | struct | internal/linkgraph/sqlite.go:24 | SQLite implementation |
| Link | struct | internal/linkgraph/sqlite.go:16 | Source, target, display, resolved |
| LintChecker | interface | internal/lint/engine.go:37 | Name()/Check() interface |
| LintEngine | struct | internal/lint/engine.go:69 | Orchestrates all checkers |
| LintIssue | struct | internal/lint/engine.go:17 | Type/RuleID/Message/Path/Severity |
| LintReport | struct | internal/lint/engine.go:28 | Issues/PagesChecked/ByCheck |
| KB | struct | internal/lint/engine.go:45 | Lint context: root, linkgraph, templates, pages |
| PageData | struct | internal/lint/engine.go:56 | Parsed page with frontmatter + markers |
| Template | struct | internal/template/template.go:43 | TemplateV2: schema, validations, lint_rules |
| Schema | struct | internal/template/template.go:15 | Frontmatter schema definition |
| ValidationRule | struct | internal/template/template.go:27 | Write-time CEL validation rule |
| LintRule | struct | internal/template/template.go:35 | Sweep-time CEL lint rule |
| NewEnv | func | internal/cel/engine.go:19 | Creates CEL env with page/old_page/now variables |
| CompileRule | func | internal/cel/engine.go:29 | Parses/compiles CEL expr, caches programs |
| Evaluate | func | internal/cel/engine.go:51 | Evaluates CEL program with panic recovery |
| ValidationError | struct | internal/cel/errors.go:5 | RuleID/Message/Line/Severity |
| BuildPage | func | internal/cel/pagebuilder.go:48 | Assembles page map from frontmatter + AST |
| BuildOldPage | func | internal/cel/pagebuilder.go:88 | Reads on-disk page, builds old_page map |
| IndexEntry | struct | internal/index/index.go:16 | Page in index |
| ParsedFrontmatter | struct | internal/frontmatter/frontmatter.go:18 | Type, Title, Fields, IsDraft |
| Wikilink | struct | internal/markdown/wikilink.go:9 | Target, Display, Heading |
| Annotation | struct | internal/markdown/annotation.go:10 | olw-auto HTML comment parser |
| ProvenanceMarker | struct | internal/markdown/provenance.go:9 | ^[type] marker parser |
| Entry | struct | internal/manifest/manifest.go:17 | Filename, SHA256, LastUpdated |
| Manager | struct | internal/manifest/manifest.go:24 | Read/Write/Add/Remove/Update entries |
| Skill | struct | internal/skill/skill.go:17 | Name, Files map |
| Entry | struct | internal/registry/registry.go:14 | Registry KB entry |
| Config | struct | internal/config/config.go:15 | .akb.yaml |
| DB | struct | internal/db/db.go:14 | Wrapper around sql.DB |

## CONVENTIONS (THIS PROJECT)

- **KB root**: Selected per invocation by the `--kb <path>` flag, falling back to the `AKB_KB` environment variable; with neither, commands fail with a usage error that lists the bases discovered nearby. No registry and no stored default (`akb use`/`akb registry` are removed; `akb discover` only reports). A relative path resolves against the working directory and `~` expands to the home directory (`internal/path.ResolveKB()`).
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
- **Template format**: TemplateV2 uses `schema.frontmatter`, `validations[]`, `lint_rules[]` (old `required[]`/`optional[]`/`body` rejected)
- **CEL variables**: `page` (map), `old_page` (nullable map), `now` (timestamp) injected at evaluation time
- **Date fields**: ISO-8601 strings (`created`, `updated`) auto-converted to `time.Time` for CEL `timestamp()`
- **Template name validation**: Names must match `^[a-zA-Z0-9_-]+$` (regex-enforced, prevents path traversal)
- **`--force` semantics**: On template commands, bypasses existence warning only; never bypasses mockup validation

## ANTI-PATTERNS (THIS PROJECT)

- Do NOT write to `index.md` or `log.md` directly (use `akb index add` or `akb log`)
- Do NOT use raw path for KB pages (use `akb write/read`)
- Do NOT upgrade `modernc.org/libc` independently (go.mod comment)
- Do NOT run concurrent DB operations (max 1 open connection)
- Do NOT add `--type` flag on `akb write` or `akb append` (type comes from frontmatter)
- Do NOT build binary in project root (use `bin/`)
- Do NOT add v2 features (MCP, goreleaser, TUI, AI exports) in v1 code
- Do NOT use `gopkg.in/yaml.v3` (replaced by `github.com/goccy/go-yaml`)
- Do NOT construct `old_page` from in-memory modified state (must read on-disk)
- Do NOT inject `old_page` for lint sweeps (lint is sweep-time, not write-time)
- Do NOT use `--force` to bypass mockup validation on `templates write` (stale mockups rejected regardless)

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
- Templates embedded in binary via `//go:embed embedded/*` (includes `.yaml` + `_pass.md` + `_fail.md`)
- Skills embedded in binary via `//go:embed embedded/*`
- Integration tests use testscript framework (`.txt` files in testdata/)
- 9 lint checks: 6 structural + 1 template-driven (cel_lint) + 2 semantic (provenance, citations)
- Provenance drift threshold: 0.20 (still checked by provenance.go)
- Lint thresholds are hardcoded constants (not configurable via `.akb.yaml` in v1)
- CEL engine: programs cached in sync.Map, cost limit 100000, panics recovered as "exceeded compute budget"
- `akb template get <name>` returns Writer View (schema + requirements only)
- `akb template get <name> --example` returns `_pass.md` content (validates against current CEL rules; warns on stderr if stale, still displays mockup)
- `akb template get <name> --full` returns complete YAML with CEL rules
- `akb templates write` validates CEL syntax and test-driven mockups; overwrite: shows diff + page count, reuses existing mockups, `--force` bypasses existence warning only
- All-errors aggregation on write: ALL failed rules reported, file NOT written if any fail
- Exit codes: 0=success, 1=validation failure, 2=internal error
- `akb template delete <name>` impact analysis: counts pages using the type before deletion; auto-runs lint after
