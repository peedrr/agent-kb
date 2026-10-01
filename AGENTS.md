# PROJECT KNOWLEDGE BASE

**Status:** v0.23.0 — a complete `akb init --author-name`/`--author-email` pair is a declared, durable commit identity: recorded in the new base's `akb.yaml` as `git-author`/`git-email`, it attributes the init commit and every later commit the base makes until the keys are edited, while a partial pair, the `AKB_AUTHOR_*` environment, and git config stay invocation-scoped and are never recorded (decision record: ADR-002 in `docs/knowledge-base`) (previous release, v0.22.0: `akb init` makes the KB's versioning a decision — `--embed`, `--no-git`, or a standalone repository — recorded as `versioning: git|none` in `akb.yaml`). The version lives in `VERSION`, is derived into `flake.nix` and the `justfile`, and the human-facing copies — this line, the CHANGELOG section, the git tag — are checked by `scripts/check-version-lockstep.sh`

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
│   ├── storage/    # Store/OpenStore (versioning modes) + GitProvider (auto-commit)
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
| Git integration | `internal/storage/mode.go`, `internal/storage/git.go` | `OpenStore`/`Store`: the versioning mode selects the provider; auto-commit, merge conflict detection |
| Commit identity | `internal/storage/identity.go` | `ResolveIdentity`: env → `akb.yaml` → git config for a mutation; `ResolveInitIdentity`: flags → env → git config → akb default for a new base |
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
| version | string | cmd/akb/main.go:16 | CLI version (injected at build via LDFLAGS) |
| commit | string | cmd/akb/main.go:20 | Source revision (injected at build via LDFLAGS; empty for plain `go build`) |
| Provider | interface | internal/storage/provider.go:10 | Write/Read/Delete/Exists/List |
| GitProvider | struct | internal/storage/git.go:40 | Git-tracked file operations |
| Store | struct | internal/storage/mode.go:64 | Storage of one base: provider, lock, preflight, commit |
| OpenStore | func | internal/storage/mode.go:82 | Opens a base's storage in the versioning mode its akb.yaml records |
| Mode | type | internal/storage/mode.go:18 | Versioning mode of a base: `git` or `none` |
| ResolveIdentity | func | internal/storage/identity.go:73 | Commit identity of a mutation: env → `akb.yaml` → git config |
| ResolveInitIdentity | func | internal/storage/identity.go:119 | Identity a new base records: flags → env → git config → akb default |
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
- **Versioning**: `.agent-kb/akb.yaml` carries `versioning: git|none` — the mode the base records its history in. A config without the key means `git`, the mode bases predating the key were created in; any other value is rejected instead of read as git. In `git` mode a mutation is committed in the repository that hosts the base; in `none` mode the files are written with no git invocation at all, and the base locks its own `.agent-kb/akb.lock`. `akb init` picks the mode: a repository of the base's own (`akb init <name>`), the enclosing repository's history (`--embed`), or no versioning (`--no-git`, excluded from the host's git status through `.git/info/exclude`). Inside a repository the mode is required: an invocation that names none refuses with exit 2.
- **Commits**: Git commits auto-created with `akb: write/delete <path>` messages
- **Merge conflicts**: Blocked; must resolve before any write/delete
- **Git identity**: The identity of a commit resolves in precedence order: `AKB_AUTHOR_NAME`/`AKB_AUTHOR_EMAIL` in the environment, then `git-author`/`git-email` in the base's `akb.yaml`, then git's own resolution (`GIT_AUTHOR_*` in the environment, then repository and global configuration). An identity akb supplies is exported as `GIT_AUTHOR_*`/`GIT_COMMITTER_*`; when it supplies none, git resolves the identity itself. An identity no source names is a configuration fault: a committing command refuses with exit 2 before writing anything. `akb init` records an identity in a new base's `akb.yaml` (`git-author`/`git-email`) in two cases: a complete `--author-name`/`--author-email` pair, which the invoker declares for the base itself (ADR-002), and the akb default (`agent-kb <agent@agent-kb>`) when no source names one. An ambient identity — the environment's, git config's, or a partial flag pair whose missing field git config filled — is never written into the file, where it would mis-attribute commits made from other machines. Repository config is never written.
- **DB path**: `.agent-kb/search.db` with WAL mode, single connection
- **is_draft**: Auto-managed frontmatter field; new pages are implicit drafts; `akb approve` sets `is_draft: false`
- **Type enforcement**: All pages MUST declare `type` in frontmatter matching a template in `.agent-kb/templates/`. `akb write` and `akb append` also refuse a page that leaves out a field the template marks `required: true` (exit 1, every missing field listed) before the CEL rules run; `akb template write` requires its PASS mockup to carry every required field
- **Build output**: Always `-o bin/akb` (never project root)
- **Template format**: TemplateV2 uses `schema.frontmatter`, `validations[]`, `lint_rules[]` (old `required[]`/`optional[]`/`body` rejected)
- **CEL variables**: `page` (map), `old_page` (nullable map), `now` (timestamp) injected at evaluation time
- **CEL rule evaluation**: write-time CEL eval errors fail closed (exit 1, write blocked, author-directed message); lint-time eval errors degrade to per-page issues and the sweep continues. The divergence is deliberate.
- **Date fields**: any frontmatter string value that parses as RFC3339 or date-only `2006-01-02` is converted to `time.Time` for CEL `timestamp()` and duration math
- **Template name validation**: Names must match `^[a-zA-Z0-9_-]+$` (regex-enforced, prevents path traversal)
- **`--force` semantics**: A warning bypass only, never a validation or mode bypass. On template commands it bypasses the existence warning but never mockup validation; on `akb init` it bypasses the repository-root layout warning but never selects a versioning mode.

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
just release                    # Preconditions + lockstep + tag v$(cat VERSION) + move release branch
nix develop                     # Dev shell (Go, gopls, delve, golangci-lint, just)

# Plain `go build -o bin/akb ./cmd/akb/` works but reports version "dev" —
# version and commit reach the binary only through -ldflags, injected by
# `just build` (git), `nix build` (flake sourceInfo), and the release
# workflow (GITHUB_SHA).
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
- Exit-2 refusals are self-sufficient: a refusal states what was found, what each choice does, and the implication of each choice, so the decision can be made from the message alone. `akb init` inside a repository names the repository and spells out `--embed` and `--no-git` with what each does to the host; a blocked commit names the host repository, the merge in progress and the path it changed, and states that the merge belongs to the host project
- `akb template delete <name>` impact analysis: counts pages using the type before deletion; auto-runs lint after
