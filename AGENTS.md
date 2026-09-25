# PROJECT KNOWLEDGE BASE

**Status:** v0.20.0 — BREAKING: the KB state directory is renamed `.akb/` → `.agent-kb/` and its config is now `akb.yaml` (no leading dot), with the layout centralized behind `internal/path` helpers; no compatibility shim (behavior recorded in `CHANGELOG.md`; the version lives in `VERSION`, is derived into `flake.nix` and the `justfile`, and the human-facing copies — this line, the CHANGELOG section, the git tag — are checked by `scripts/check-version-lockstep.sh`)

## OVERVIEW

Agent KB (akb): Go CLI for managing a git-tracked knowledge base with SQLite FTS5 search, SQLite link graph, YAML frontmatter, typed page templates, CEL-driven validation, lint engine, and raw drift detection.

## STRUCTURE

```
agent-kb/
├── cmd/akb/        # CLI commands (Cobra, 25+ commands)
├── internal/       # Core packages
│   ├── cel/        # CEL expression engine (validation + lint)
│   ├── config/     # akb.yaml handling
│   ├── db/         # SQLite schema + WAL mode
│   ├── frontmatter/# YAML frontmatter parsing (goldmark + goccy/go-yaml)
│   ├── index/      # kb/index.md management
│   ├── linkgraph/  # SQLite link tracking (wikilinks)
│   ├── lint/       # 10 lint checkers + engine
│   ├── log/        # kb/log.md append-only log
│   ├── manifest/   # raw/files.log SHA-256 manifest
│   ├── markdown/   # Wikilink, annotation, provenance parsers
│   ├── path/       # KB path resolution & guards
│   ├── search/     # SQLite FTS5 full-text search
│   ├── skill/      # Embedded skill management (//go:embed)
│   ├── storage/    # GitProvider (auto-commit)
│   └── template/   # Typed page templates (TemplateV2 with CEL rules)
├── test/           # Integration tests (testscript)
├── flake.nix       # Nix flake (dev shell + build package)
├── justfile        # Task runner (build/test/lint/release); version injected from VERSION
└── VERSION         # Single source of truth for the release version
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add command | `cmd/akb/` | New subcommand = new file |
| KB path logic | `internal/path/path.go` | `ResolveKB()` selects the invocation's base; `KBRoot()` walks up to the nearest `.agent-kb/` root, the check that marks a directory as a knowledge base |
| Page write flow | `cmd/akb/write.go` | stdin → frontmatter → CEL validation → git → search → links |
| Index management | `internal/index/index.go` | kb/index.md parsing/rendering |
| Search | `internal/search/sqlite.go` | SQLite FTS5 with BM25 ranking |
| Link graph | `internal/linkgraph/sqlite.go` | 3-step wikilink resolution |
| Wikilink parser | `internal/markdown/wikilink.go` | Excludes code blocks, inline code, HTML comments |
| Annotation parser | `internal/markdown/annotation.go` | `<!-- olw-auto: ... -->` HTML comments |
| Provenance markers | `internal/markdown/provenance.go` | `^[inferred]`, `^[ambiguous]`, `^[extracted]` |
| CEL engine | `internal/cel/engine.go` | Environment builder, rule compiler, evaluator with panic recovery |
| CEL page builder | `internal/cel/pagebuilder.go` | Builds `page`/`old_page` maps; walks the Goldmark AST into `page.ast` (headings, links, code blocks) |
| Template loader | `internal/template/template.go` | TemplateV2 with schema, validations, lint_rules |
| Template commands | `cmd/akb/template.go`, `templates_write.go`, `template_delete.go` | `template get/list/write/delete` |
| Template delete | `cmd/akb/template_delete.go` | Impact analysis (page count), `--force`, path traversal prevention |
| Git integration | `internal/storage/git.go` | Auto-commit, merge conflict detection |
| DB schema | `internal/db/db.go` | documents, pages, links tables + FTS5 |
| Config format | `internal/config/config.go` | YAML akb.yaml |
| Lint engine | `internal/lint/engine.go` | LintEngine, LintChecker interface |
| Lint checks | `internal/lint/*.go` | 10 checkers (broken_links, orphans, empty_pages, missing_frontmatter, required_fields, index_consistency, citations, provenance, cel_lint, type_orphan) |
| CEL lint checker | `internal/lint/cel.go` | Evaluates template `lint_rules` with `now` injection |
| Manifest | `internal/manifest/manifest.go` | raw/files.log SHA-256 tracking |
| Skill install | `internal/skill/skill.go` | `//go:embed embedded/*` |
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
| RootCmd | Cobra.Command | cmd/akb/root.go:43 | Base CLI command |
| version | string | cmd/akb/main.go:11 | CLI version (injected at build via LDFLAGS) |
| Provider | interface | internal/storage/provider.go:6 | Write/Read/Delete/Exists/List |
| GitProvider | struct | internal/storage/git.go:41 | Git-tracked file operations |
| Searcher | interface | internal/search/searcher.go:33 | IndexPage/RemovePage/Search/RebuildIndex |
| SQLiteFTS5Searcher | struct | internal/search/sqlite.go:18 | FTS5 implementation |
| Updater | interface | internal/linkgraph/updater.go:13 | UpdatePageLinks/RemovePage |
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
| NewEnv | func | internal/cel/engine.go:25 | Creates CEL env with page/old_page/now variables |
| CompileRule | func | internal/cel/engine.go:41 | Parses/compiles CEL expr with the cost limit, caches programs |
| Evaluate | func | internal/cel/engine.go:64 | Evaluates CEL program with panic recovery |
| ValidationError | struct | internal/cel/errors.go:7 | RuleID/Message/Line/Severity |
| BuildPage | func | internal/cel/pagebuilder.go:45 | Assembles page map from frontmatter + AST |
| BuildOldPage | func | internal/cel/pagebuilder.go:85 | Reads on-disk page, builds old_page map |
| IndexEntry | struct | internal/index/index.go:16 | Page in index |
| ParsedFrontmatter | struct | internal/frontmatter/frontmatter.go:18 | Type, Title, Fields; draft state lives in Fields and is read via `frontmatter.IsDraft()` |
| Wikilink | struct | internal/markdown/wikilink.go:9 | Target, Display, Heading |
| Annotation | struct | internal/markdown/annotation.go:10 | olw-auto HTML comment parser |
| ProvenanceMarker | struct | internal/markdown/provenance.go:9 | ^[type] marker parser |
| Entry | struct | internal/manifest/manifest.go:17 | Filename, SHA256, LastUpdated |
| Manager | struct | internal/manifest/manifest.go:24 | Read/Write/Add/Remove/Update entries |
| Skill | struct | internal/skill/skill.go:17 | Name, Files map |
| Config | struct | internal/config/config.go:14 | akb.yaml |
| StateDirName | const | internal/path/path.go:29 | KB state dir name (`.agent-kb`) |
| ConfigFileName | const | internal/path/path.go:34 | KB config file name (`akb.yaml`) |
| StateDir/ConfigPath/TemplatesDir/SearchDBPath | funcs | internal/path/path.go | Layout constructors — the only definition site for on-disk state paths |

## CONVENTIONS (THIS PROJECT)

- **KB root**: Selected per invocation by the `--kb <path>` flag, falling back to the `AKB_KB` environment variable; with neither, commands fail with a usage error that lists the bases discovered nearby. The resolved path must hold the `.agent-kb/akb.yaml` config file that marks a base — a regular file, not a directory; without it the command fails with a usage error (exit 2) naming the missing marker and pointing at `akb discover`. No registry and no stored default (`akb use`/`akb registry` are removed; `akb discover` only reports). A relative path resolves against the working directory and `~` expands to the home directory (`internal/path.ResolveKB()`).
- **Paths**: Always relative to KB root; `kb/` prefix stripped
- **Managed files**: `index.md`, `log.md` cannot be written directly
- **Raw access**: `raw/` prefix → separate storage; use `akb raw` commands
- **Commits**: Git commits auto-created with `akb: write/delete <path>` messages
- **Merge conflicts**: Blocked; must resolve before any write/delete
- **Git identity**: A commit is authored by the repository's configured `user.name`; a repository without one gets the per-invocation fallback `-c user.name=akb -c user.email=akb@local`. Repository config is never written, except by `akb init` (`ensureGitConfig`)
- **DB path**: `.agent-kb/search.db` with WAL mode, single connection
- **is_draft**: Auto-managed frontmatter field; new pages are implicit drafts; `akb approve` sets `is_draft: false`
- **Type enforcement**: All pages MUST declare `type` in frontmatter matching a template in `.agent-kb/templates/`. `akb write` and `akb append` also refuse a page that leaves out a field the template marks `required: true` (exit 1, every missing field listed) before the CEL rules run; `akb template write` requires its PASS mockup to carry every required field
- **Build output**: Always `-o bin/akb` (never project root)
- **Template format**: TemplateV2 uses `schema.frontmatter`, `validations[]`, `lint_rules[]` (old `required[]`/`optional[]`/`body` rejected)
- **CEL variables**: `page` (map), `old_page` (nullable map), `now` (timestamp) injected at evaluation time
- **CEL rule evaluation**: write-time CEL eval errors fail closed (exit 1, write blocked, author-directed message); lint-time eval errors degrade to per-page issues and the sweep continues. The divergence is deliberate.
- **Date fields**: any frontmatter string value that parses as RFC3339 or date-only `2006-01-02` is converted to `time.Time` for CEL `timestamp()` and duration math
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
- Do NOT use `--force` to bypass mockup validation on `template write` (stale mockups rejected regardless)

## COMMANDS

```bash
just build                      # Build with version injected (NEVER output to project root)
just test                       # Unit tests
go test ./test/ -test.v         # Integration tests (testscript)
just lint                       # Lint (golangci-lint)
just release                    # Preconditions + lockstep + tag v$(cat VERSION)
nix develop                     # Dev shell (Go, gopls, delve, golangci-lint, just)

# Plain `go build -o bin/akb ./cmd/akb/` works but produces version "dev" —
# the version reaches the binary only through -ldflags, which `just build`
# and `nix build` inject from VERSION.
```

## NOTES

- WIP: APIs may change between commits
- Wikilink resolution: exact → namespace prefix → basename match; `[[display]](dest)` is the explicit-destination form (bracket = label, paren = destination, which wins) and its `dest` is normalized — surrounding whitespace and angle brackets, a leading `./`, and a trailing `.md` are stripped — before resolving
- FTS5 query escaping strips FTS5 operators (OR, AND, NOT) and special chars
- Templates embedded in binary via `//go:embed embedded/*` (includes `.yaml` + `_pass.md` + `_fail.md`)
- Skills embedded in binary via `//go:embed embedded/*`
- Integration tests use testscript framework (`.txt` files in testdata/)
- 10 lint checks: 7 structural + 1 template-driven (cel_lint) + 2 semantic (provenance, citations)
- Provenance drift threshold: 0.20 (still checked by provenance.go)
- Lint thresholds are hardcoded constants (not configurable via `akb.yaml` in v1)
- CEL engine: programs cached in sync.Map, cost limit 100000, panics recovered as "exceeded compute budget"
- `akb template get <name>` returns Writer View (schema + requirements only)
- `akb template get <name> --example` returns `_pass.md` content (validates against current CEL rules; warns on stderr if stale, still displays mockup)
- `akb template get <name> --full` returns complete YAML with CEL rules
- `akb template write` validates CEL syntax and test-driven mockups; overwrite: shows diff + page count, reuses existing mockups, `--force` bypasses existence warning only
- All-errors aggregation on write: ALL failed rules reported, file NOT written if any fail
- Exit codes: 0=success; 1=the command ran but produced a result to act on (failed page validation, raw drift); 2=the command could not do its work — bad invocations print a `usage:` prefix, akb faults an `internal:` prefix. Command failures that are neither — a git error, for example — exit 1 with an `Error: ` prefix (`classifyExit` in `cmd/akb/main.go`)
- `akb template delete <name>` impact analysis: counts pages using the type before deletion; auto-runs lint after
