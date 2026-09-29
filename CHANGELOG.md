# Changelog

Behavior changes and fixes worth acting on. Versions before 0.18.0 predate this file.

## [0.22.0] — 2026-09-29

### Changed

- **BREAKING: `akb init` inside a git repository refuses until the KB's versioning mode is
  chosen.** The invocation exits 2 naming the repository it found and what each flag does:
  `--embed` puts the KB into that repository's history — akb commits only `kb/`, `raw/` and
  `.agent-kb/`, never the host's other worktree changes — and `--no-git` leaves the KB
  unversioned and out of the host's `git status` through `.git/info/exclude`, a file local
  to the clone. Previously the same invocation created a repository nested inside the
  host's, and `akb init` wrote `user.name` into it. A KB that is not inside a repository is
  unchanged: `akb init <name>` with no flag still creates a repository of its own.

- **The versioning mode is recorded in `.agent-kb/akb.yaml` as `versioning: git` or
  `versioning: none`, written explicitly at init in both modes.** A config that predates
  the key means `git` — the mode those bases were created in — so an existing KB keeps
  working unchanged; any other value is rejected rather than read as git, so a misspelled
  mode never commits under the wrong assumption. A `none` base invokes no git at all:
  mutations write their files and stop, the base takes its lock on `.agent-kb/akb.lock`
  instead of on the host repository, and `akb status` reports `Versioning: none` in place of
  the git line. `--no-commit` is a no-op there, as it is for any base that never commits.

- **Commit identity resolves from the environment, then the base, then git, and is exported
  to the commit instead of passed as `-c user.name=…`.** `AKB_AUTHOR_NAME`/`AKB_AUTHOR_EMAIL`
  win, then `git-author`/`git-email` in the base's `akb.yaml`, then git's own resolution
  (`GIT_AUTHOR_*` in the environment, then the merged repository and global configuration).
  An identity akb supplies is exported as `GIT_AUTHOR_*` and `GIT_COMMITTER_*`; when it
  supplies none, git resolves the identity itself, which leaves a user-set `GIT_AUTHOR_*`
  working as a documented escape hatch. No source naming an identity is a usage refusal
  (exit 2) that fires before anything is written, replacing the raw
  `Committer identity unknown` fatal an agent could only hit after its page was staged.

- **A mutation blocked by an in-progress merge is refused up front instead of failing at
  the commit.** Detecting a clean in-progress merge (`MERGE_HEAD`), not only unmerged paths,
  closes the gap where a `git status --porcelain` scan passed and
  `fatal: cannot do a partial commit during a merge` surfaced after the page had been
  written and staged. The refusal names the host repository and the path it changed, and
  states that the merge belongs to the host project rather than to the KB.

- **Every mutating command goes through one mode-aware storage helper.** `storage.OpenStore`
  selects the provider from the recorded mode and owns the lock, the identity and merge
  preflight, and the commit, replacing the hardcoded `NewGitProvider` call sites and the
  direct `LockRepo`/`CommitFiles`/`NothingToCommit` calls across the command packages. The
  base lock is now held across the `kb/log.md` appends too.

### Added

- **`akb init` gained `--embed`, `--no-git`, `--force`, `--author-name`/`--author-email`,
  and `akb init .`.** `--force` bypasses the refusal to write `.agent-kb/`, `kb/` and
  `raw/` directly into a repository root; it selects no versioning mode, so a repository-root
  invocation still names one (`akb init . --embed --force`). `--author-name`/
  `--author-email` supply the init commit's identity for that commit only and are never
  written to `akb.yaml`: the akb default (`agent-kb <agent@agent-kb>`) is recorded there
  only when neither the environment nor git config names an identity, because an identity
  baked into a file that travels with the base would be mis-attributed after a clone.
  `akb init .` initializes the directory it runs in and adopts that directory's basename as
  the KB name, reported on the exit-0 line. `akb init` honors `--no-commit`: it scaffolds and
  stages the base and leaves the commit to the caller, so an orchestrator can bundle init
  with templates and an overview into one commit.

- **`KNOWN-LIMITATIONS.md` records that a KB cannot be a git submodule of the host
  project**, with the reason (a base's repository is local to the machine that initialized
  it and akb has no remote or push machinery, while git refuses the local-path clone form
  since CVE-2022-39253: `fatal: transport 'file' not allowed`) and the supported answers
  (`--embed`, `--no-git`, or a human-run `git submodule add` over a hosted repository).

### Fixed

- **`akb init` no longer clobbers an existing `.gitignore`.** Both files it writes append
  the lines they need and keep every rule already there, so re-running init over a
  directory that carries an ignore file is safe instead of destructive.

- **Repository-root `.git/info/exclude` entries are anchored with a leading slash**
  (`/.agent-kb/`, `/kb/`, `/raw/`). Unanchored, they also matched a same-named directory
  anywhere below the root, hiding unrelated host files from `git status`.

- **A `--no-git` base at a repository root no longer writes a root `.gitignore`** for
  exclusions the host's exclude file already carries.

- **An invocation with nothing to record no longer fails on a commit identity it would
  never use.** `akb raw sync`, `akb index remove` and `akb template write` settle the
  no-change case under the base lock before the preflight; a sync that finds nothing also
  stops rewriting the manifest.

- **Post-write failure guidance is mode-accurate**: an unversioned base is no longer told to
  resolve a merge or finish a commit by hand.

## [0.21.0] — 2026-09-25

### Added

- **`akb init` accepts `--description` to record what a knowledge base holds.**
  The description is stored as `description` in the base's `akb.yaml` and shown
  by `akb discover` (text and `--json`) next to the base's name and path, so a
  listing of discovered bases says what each contains. The flag is optional; a
  base without one discovers exactly as before, and a description can be added
  later by editing `akb.yaml` directly.

## [0.20.0] — 2026-09-25

### Changed

- **BREAKING: the knowledge-base state directory is renamed `.akb/` → `.agent-kb/`,
  and its config file loses the redundant leading dot: `.akb.yaml` → `akb.yaml`.**
  The directory now carries the project's name; the tool-facing files inside it
  keep the tool's. `akb init` creates the new layout, and every command —
  discovery, path resolution, read/write, lint, templates, search — recognizes
  only it. There is no compatibility shim: a base in the old layout is reported
  as not a knowledge base, with the expected marker path named. Migrate an
  existing base by renaming the directory and config file (`mv .akb .agent-kb &&
  mv .agent-kb/.akb.yaml .agent-kb/akb.yaml`) and updating the `.akb/search.db*`
  line in the base's `.gitignore`. Tool-named surfaces are unchanged: the
  `AKB_KB` environment variable, the `*.akb.bak` backup extension, and the
  `page.akb` CEL map key.
- **The on-disk layout now has one definition site.** `internal/path` exports
  `StateDirName`, `ConfigFileName`, and the `StateDir`/`ConfigPath`/
  `TemplatesDir`/`SearchDBPath` helpers, replacing the `".akb"` string literals
  scattered across the command and internal packages.

## [0.19.1] — 2026-09-25

### Fixed

- **`akb init` stages and commits the new base through the index-lock retry.** A concurrent
  process holding the git index lock no longer aborts the initial staging of base files;
  the retry now covers both stages.
- **The compute-budget error is identified by sentinel, not by text.** `internal/cel`
  exports `ErrComputeBudget`, and a rule's evaluation error embeds the key names the rule
  reads — so text matching handed the budget remedy to rules whose field names collide with
  the phrase. Write-path dispatch now matches with `errors.Is`.
- **`akb template get --example` warns about every broken mockup rule, not just the first.**
  Revalidation previously stopped at the first unevaluable rule, hiding later failures and
  dropping the variant's staleness headline. Every failed rule is collected, then named.
- **The template-load failure is the error reported, not a stale page-read error.** When
  `akb index add` falls back to templates and the template set fails to load, the wrapped
  load error is surfaced — it names the cause where the old message named only the page;
  the per-file guard at the append, index-add, lint, and type-directory call sites keeps
  its own error, its classification, and its exit code.

## [0.19.0] — 2026-09-24

### Added

- **`akb lint` gained a tenth check, `required_fields`.** The sweep now reports every on-disk
  page that leaves a schema-required frontmatter field unset, as an error naming the missing
  fields and the template that declares them. Write-time validation only covers pages that
  passed `akb write`, so a page pulled from git, planted by hand, or written before its
  template declared a field stayed invisible until now. A page without frontmatter and a page
  whose `type` has no template declare no schema and stay with `missing_frontmatter` and
  `type_orphan`. The check reads presence only — a field set to an empty value counts as set
  — matching `akb write`, and type and enum constraints stay with the template's CEL rules.

### Changed

- **`akb approve` refuses a draft that leaves a schema-required frontmatter field unset.**
  Approval runs the write path's required-field check before it rewrites the page, so a page
  `akb write` would reject can no longer be published by approving it: the run reports the
  missing fields and the declaring template and exits 1, leaving the page untouched and no
  approval commit behind. `akb approve --all-drafts` refuses the incomplete
  drafts and approves the rest, then fails with the count of the refused drafts, so a refusal
  is never reported as part of a successful batch. A page whose `type` has no template is
  approved as before.

## [0.18.0] — 2026-09-24

### Changed

- **CEL reports wikilink targets in the link graph's spelling.** `page.ast.links[].target`
  for an explicit-destination wikilink (`[[display]](dest)`) is now the normalized
  destination — the spelling the link graph records — instead of goldmark's destination as
  written. `[[Label]](notes/x.md)` reports target `notes/x` in both readers, so a template
  rule comparing `target` with a page path sees one convention. `text` is unchanged: it
  still preserves the raw source token. The `superseded_requires_link` rule in the embedded
  `adr` template now matches a `.md`-bearing body link against a `.md`-less `supersedes`
  value, which it previously could not. Behavior spec: `docs/wikilinks.md`.

- **Explicit-destination link-graph output changed for titled and angle-bracketed
  destinations.** A destination carrying a CommonMark link title
  (`[[a]](notes/x.md "title")`) or angle brackets (`[[a]](<notes/x.md>)`) is now recorded
  under the normalized target (`notes/x`) rather than the bracket label or the raw
  destination. The same pass tightened validity to match goldmark: a malformed explicit
  destination — an unclosed angle-bracket destination (`[[a]](<b)`) or an escaped closing
  paren (`[[a]](\)`) — now falls back to a plain wikilink in both the link graph and CEL
  instead of producing a bogus destination.

- **Pre-existing pages that use `[[a]](dest)` shift their recorded target on the next write
  or index rebuild.** The recorded target becomes the normalized destination (`notes/x`
  instead of the bracket label `a` or the raw `notes/x.md`). Consistent with the new
  convention, but it is a data change: backlinks, orphan status, and broken-link lint for
  such pages reflect the new target after the page is written again or
  `akb index rebuild` runs. This shift was never noted in 0.17.0.
