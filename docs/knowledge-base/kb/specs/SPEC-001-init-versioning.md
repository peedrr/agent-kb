---
grammar: 1
type: spec
id: SPEC-001
title: Init Selects the KB Versioning Mode and Resolves Commit Identity
kind: change
status: completed
provenance: human
created: 2026-09-25
updated: 2026-10-01
scope: ["cmd/akb/**", "internal/storage/**", "internal/config/**", "internal/skill/embedded/**"]
verified:
  commit: 15dfaacdac9a35e425b49e1cdb6758af8dcab3fd
  branch: main
  at: "2026-10-01T15:55:21Z"
  method: manual-review
  state: unverified
  next_review_by: "2027-01-01"
anchors:
  - claim: REQ-005
    path: internal/storage/git.go
    symbol: CommitFiles
    kind: flow
    verify: { method: check, check: "go test ./internal/storage/ -run TestGitProviderCommitLeavesUnrelatedStagedChanges" }
    state: unverified
    checked_at: "2026-10-01T15:55:21Z"
    evidence: "pathspec machinery named in the source record (section 3); not re-resolved at migration"
  - claim: REQ-006
    path: internal/storage/mode.go
    symbol: OpenStore
    kind: flow
    verify: { method: check, check: "go test ./internal/storage/..." }
    state: unverified
    checked_at: "2026-10-01T15:55:21Z"
    evidence: "opens a base's storage in the mode its akb.yaml records, per the project code map; not re-resolved at migration"
  - claim: REQ-011
    path: internal/storage/identity.go
    symbol: ResolveInitIdentity
    kind: local
    verify: { method: check, check: "go test ./internal/storage/..." }
    state: unverified
    checked_at: "2026-10-01T15:55:21Z"
    evidence: "init-time identity precedence per the project code map; not re-resolved at migration"
  - claim: REQ-013
    path: internal/storage/identity.go
    symbol: ResolveIdentity
    kind: flow
    verify: { method: check, check: "go test ./internal/storage/..." }
    state: unverified
    checked_at: "2026-10-01T15:55:21Z"
    evidence: "commit-time identity precedence per the project code map; not re-resolved at migration"
  - claim: REQ-016
    path: internal/storage/git.go
    symbol: checkMergeConflicts
    kind: local
    verify: { method: ask, ask: "confirm the clean-merge detection (MERGE_HEAD) lives in the merge preflight" }
    state: unverified
    checked_at: "2026-10-01T15:55:21Z"
    evidence: "extended to detect clean in-progress merges per the source record (section 7); location not re-resolved at migration"
---

# SPEC-001: Init Selects the KB Versioning Mode and Resolves Commit Identity

> **For agents:** this SPEC is authority only while `status: active` AND
> `verified.state: live`. If your task conflicts with it, or any anchor fails
> to resolve at HEAD, stop and name the conflict. Do not silently work around
> it; propose an amendment or supersession instead.

Migrated from the pre-KB design record preserved at `raw/spec/init-versioning-spec.md`. The change it specifies has landed (released in v0.22.0), so this record is `completed`.

## Why

Dogfooding motivated the work: a KB that travels with the code it describes, versioned in the project's own history. That requires `akb init` to be safe and explicit inside an existing git worktree shared by multiple agents. It also requires commit identity to attribute correctly across machines without akb ever writing repository configuration.

## Goals and Non-Goals

- Goal: init inside a repository is an explicit, refused-by-default choice between embedding in the host's history and no versioning.
- Goal: embedded commits record only the KB's own paths; host worktree state is never swept in.
- Goal: commit identity resolves from environment, then `akb.yaml`, then git-native resolution — never from repository config akb writes.
- Goal: exit-2 refusals are self-sufficient.
- Non-goal: KB-as-submodule support (rejected; see Decisions).
- Non-goal: remote/push functionality of any kind.
- Non-goal: per-command `--author-name`/`--author-email` flags; the environment channel covers per-agent attribution.

## Current Behaviour

Pre-change baseline this SPEC altered (the change has landed; this records what it replaced):

- `akb init` always created a standalone git repository, nesting an unregistered repo when invoked inside one.
- Init clobbered an existing `.gitignore` via `os.WriteFile`.
- Identity reached git through `-c user.name=…`, and init wrote `user.name` into newly created repositories (`ensureGitConfig`).
- Merge preflight scanned porcelain only; a clean in-progress merge passed, and the raw git partial-commit fatal surfaced after the page was written and staged.

## Requirements

REQ-001: WHEN `akb init <name>` runs outside any git repository, the system MUST create a standalone git repository for the new base without requiring a mode flag.
REQ-002: WHEN `akb init <name>` runs inside a git repository and names no versioning mode, the system MUST refuse with exit 2.
REQ-003: WHEN a command refuses with exit 2, the refusal MUST state what was found, what each choice does, and the implication of each choice.
REQ-004: WHERE the init target is a repository root, the system MUST require `--force` before scaffolding.
REQ-005: WHEN `akb init` runs with `--embed`, the init commit MUST record exactly the KB's own paths (`kb/`, `raw/`, `.agent-kb/`, `.gitignore` under the KB root); the commit MUST NOT include any other host worktree change, staged or not.
REQ-006: WHEN `akb init` creates a base, the system MUST write the selected mode to `akb.yaml` as an explicit `versioning: git` or `versioning: none` key.
REQ-007: WHEN an `akb.yaml` carries no `versioning` key, the system MUST treat the base as `git`-versioned.
REQ-008: WHILE a base records `versioning: none`, a mutation MUST NOT invoke git; the mutation MUST take the base's own lock at `.agent-kb/akb.lock`.
REQ-009: WHEN `akb init --no-git` runs inside a repository, the system MUST exclude the base from the host's git status through `.git/info/exclude`; the system MUST NOT edit the host's tracked `.gitignore`.
REQ-010: WHEN init writes a `.gitignore` that already exists, the system MUST append only the missing lines; the system MUST NOT truncate existing content.
REQ-011: WHEN `akb init` records a commit identity, the system MUST resolve it in precedence order: init flags, then `AKB_AUTHOR_*` environment, then git config, then the akb default.
REQ-012: WHEN the environment or git config yields an identity at init, the system MUST NOT write that identity to `akb.yaml`; the system MUST record the akb default only when no source names an identity.
REQ-013: WHILE a committing command runs, the system MUST resolve commit identity in precedence order: `AKB_AUTHOR_*` environment, then `git-author`/`git-email` in `akb.yaml`, then git-native resolution.
REQ-014: WHEN a committing command takes its identity from the environment or `akb.yaml`, the system MUST export both `GIT_AUTHOR_*` and `GIT_COMMITTER_*` to the git subprocess.
REQ-015: IF a committing command cannot resolve a commit identity, THEN the system MUST refuse with exit 2 before any file mutation, naming the `AKB_AUTHOR_*` environment variables and `akb.yaml` `git-author`/`git-email` as remedies.
REQ-016: WHEN a merge is in progress in the host repository, conflicted or clean, the system MUST block its commit before writing anything.
REQ-017: WHEN `akb init .` runs, the system MUST derive the base name from the directory basename and report the adopted name in the exit-0 output.
REQ-018: WHERE `--no-commit` is passed to `akb init`, the system MUST leave the scaffolded base staged but uncommitted.

## Acceptance Criteria

AC-001 (verifies REQ-001): WHEN `akb init foo` runs in a directory with no enclosing repository, the command MUST exit 0 and `foo/.git` MUST exist.
  verify: { method: check, check: "go test ./test/..." }
AC-002 (verifies REQ-002): WHEN `akb init foo` runs inside a repository with no mode flag, the command MUST exit 2 and its output MUST name both `--embed` and `--no-git`.
  verify: { method: check, check: "go test ./test/..." }
AC-003 (verifies REQ-003): WHEN the init-inside-repo refusal prints, an agent that has read nothing else MUST be able to choose correctly between `--embed` and `--no-git` from the message alone.
  verify: { method: ask, ask: "run `akb init foo` inside a repository and judge the refusal against REQ-003" }
AC-004 (verifies REQ-004): WHEN `akb init . --embed` runs at a repository root without `--force`, the command MUST exit 2; the same invocation with `--force` MUST exit 0.
  verify: { method: check, check: "go test ./test/..." }
AC-005 (verifies REQ-005): WHEN `akb init --embed` commits with an unrelated host file already staged, the init commit MUST contain only KB paths and the unrelated file MUST remain staged.
  verify: { method: check, check: "go test ./internal/storage/ -run TestGitProviderCommitLeavesUnrelatedStagedChanges" }
AC-006 (verifies REQ-006): WHEN init completes in either mode, the new `akb.yaml` MUST contain a `versioning:` line whose value matches the selected mode.
  verify: { method: check, check: "go test ./test/..." }
AC-007 (verifies REQ-007): WHEN a base's `akb.yaml` holds no `versioning` key, commands on that base MUST behave exactly as under `versioning: git`.
  verify: { method: check, check: "go test ./internal/config/... ./internal/storage/..." }
AC-008 (verifies REQ-008): WHILE a base records `versioning: none`, a page write MUST leave the host repository's `git status` untouched and MUST create no commit.
  verify: { method: check, check: "go test ./test/..." }
AC-009 (verifies REQ-009): WHEN `akb init docs --no-git` runs inside a repository, `.git/info/exclude` MUST gain an entry covering the base and the tracked `.gitignore` MUST be byte-identical afterwards.
  verify: { method: check, check: "go test ./test/..." }
AC-010 (verifies REQ-010): WHEN init targets a directory whose `.gitignore` already holds custom lines, every original line MUST still be present after init.
  verify: { method: check, check: "go test ./test/..." }
AC-011 (verifies REQ-011): WHEN `--author-name`, `AKB_AUTHOR_NAME`, and git config all name different identities, the init commit MUST carry the flag's identity.
  verify: { method: check, check: "go test ./internal/storage/..." }
AC-012 (verifies REQ-012): WHEN git config yields an identity at init, the new `akb.yaml` MUST NOT contain `git-author`; WHEN nothing yields one, it MUST contain the akb default.
  verify: { method: check, check: "go test ./test/..." }
AC-013 (verifies REQ-013): WHEN `AKB_AUTHOR_*` and `akb.yaml` `git-author` both exist, a mutation's commit MUST carry the environment identity.
  verify: { method: check, check: "go test ./internal/storage/..." }
AC-014 (verifies REQ-014): WHEN akb supplies the identity, the resulting commit's author and committer MUST both equal the supplied identity.
  verify: { method: check, check: "go test ./internal/storage/..." }
AC-015 (verifies REQ-015): WHEN no identity source exists, a committing command MUST exit 2 with no page written; the message MUST name `AKB_AUTHOR_NAME` and `git-author`.
  verify: { method: check, check: "go test ./test/..." }
AC-016 (verifies REQ-016): WHEN a clean merge is in progress in the host, `akb write` MUST refuse before staging and its message MUST contain the host repository path.
  verify: { method: check, check: "go test ./test/..." }
AC-017 (verifies REQ-017): WHEN `akb init .` runs in a directory named `foo`, the exit-0 output MUST report `foo` as the adopted base name.
  verify: { method: check, check: "go test ./test/..." }
AC-018 (verifies REQ-018): WHEN `akb init foo --no-commit` runs, the scaffold MUST exist on disk and the repository MUST hold no akb init commit.
  verify: { method: check, check: "go test ./test/..." }

## Contract and Invariants

- The system MUST NOT write repository configuration for commit identity; identity MUST travel through `akb.yaml` or the environment (`ensureGitConfig` is removed).
- The system MUST NOT let `--force` select a versioning mode; `--embed` or `--no-git` MUST select the mode.
- IF an `akb.yaml` carries a `versioning` value other than `git` or `none`, THEN the system MUST reject it rather than read it as `git`.
- WHILE a base records `versioning: none`, `--no-commit` flags on other commands MUST be accepted as no-ops.
- WHEN `akb status` runs against a `versioning: none` base, the command MUST report the versioning mode instead of the git line.
- The init refusal MUST distinguish membership: "inside git repository `<root>`" for subdirectories, "is a git repository" when the target is the root.
- The merge-in-progress refusal MUST name the host repository, locate the merge relative to the KB, and state that the merge belongs to the host project — never resolve it from the KB.

## Preserved Behaviour

- Standalone init outside any repository keeps its pre-change behaviour and requires no flags.
- Existing bases without a `versioning` key remain valid.
- Unrelated host worktree changes, staged or not, MUST survive every akb commit untouched.
- User-set `GIT_AUTHOR_*` in the environment MUST remain honored as the documented escape hatch when akb supplies no identity.
- `--force` MUST keep its single project-wide meaning: bypassing the repository-root layout warning.

## Decisions and Rejected Alternatives

The versioning mode is recorded explicitly in `akb.yaml` in both modes, so the mode is visible and correctable in the file. An identity resolved from the environment or git config is never written into `akb.yaml`: the file travels with the KB, and a baked-in local identity would mis-attribute on every other machine after push and clone. akb exports `GIT_AUTHOR_*`/`GIT_COMMITTER_*` to the git subprocess, replacing the `-c user.name=…` mechanism.

- **`--standalone` / nested unregistered repos as a mode** — rejected: nested unregistered repos are not a mode akb offers. Do not re-propose unless akb gains a registry or submodule support.
- **KB as a git submodule of the host** — rejected: requires a hosted remote (git refuses local-path submodule clones by default since CVE-2022-39253), push machinery akb does not have, and a gitlink-bump commit per KB write; fresh clones get an empty directory — the opposite of "travels with the code". Do not re-propose unless akb gains remote/push functionality.
- **Per-command `--author-name`/`--author-email` flags** — rejected: agents would have to know which commands commit; the `AKB_AUTHOR_*` environment channel covers the need. Do not re-propose unless a per-invocation attribution need the environment cannot express surfaces.
- **Writing the exclusion to the host's tracked `.gitignore`** — rejected: `.git/info/exclude` is local to the clone and trivially reversible. Do not re-propose unless the exclusion must travel to other clones.

## Assumptions and Open Questions

[ASSUMPTION: anchor symbols are named from the source design record and project documentation; their states are `unverified` because re-resolving them against the code was out of scope for this migration.]

No open questions.

## Affected Surface and Ordering

- `cmd/akb/` — init and every committing command (write, append, delete, approve, index, raw, template)
- `internal/storage/` — mode selection, provider factory, identity resolution, merge preflight
- `internal/config/config.go` — the `versioning` key
- `internal/skill/embedded/kb-management/` — INIT.md decision step; SKILL.md quick reference
- Project `AGENTS.md`, `KNOWN-LIMITATIONS.md` — the refusal convention and the submodule limitation entry

depends_on: []

## Verification Plan

- Unit: `go test ./internal/storage/... ./internal/config/...`
- Integration (testscript): `go test ./test/...`
- Human-only residue: refusal-message wording quality; the decision to promote this record's durable content into a `kind: capability` SPEC for init and versioning (this record stays `completed`).

## Drift Ledger

- 2026-10-01 · page creation · none → completed · migrated from the pre-KB design record after the change landed · evidence: `raw/spec/init-versioning-spec.md`, the v0.22.0 release description in project `AGENTS.md`, and the `akb 0.22.0 (ff8c434)` command surface

## Revisit Triggers

- `internal/storage/mode.go`, `internal/storage/identity.go`, `internal/storage/git.go` — any change to mode selection, identity precedence, or merge preflight
- `cmd/akb/init.go` — any change to the init behavior matrix
- Promotion of this record's durable content into a `kind: capability` SPEC (sets `superseded_by` here)
- 2027-01-01 — `verified.next_review_by`

## Evidence Appendix

- Source record: `raw/spec/init-versioning-spec.md` (file created 2026-09-25, pre-KB; self-described as "agreed design, pre-implementation" at authoring).
- Landed behaviour confirmed at migration only via project documentation and the `akb 0.22.0 (ff8c434)` command surface; anchors deliberately left `unverified` because codebase recon was out of scope for the migration.
- Source-cited external verifications: git 2.55.0 refuses `file` transport for local-path submodule clones without `protocol.file.allow=always`; git refuses partial commits during any in-progress merge (`fatal: cannot do a partial commit during a merge`).
- Source-named tests evidencing the embedded-commit pin: `TestGitProviderCommitLeavesUnrelatedStagedChanges`, `TestCommitFilesRecordsDeletionInNestedKB`.
- Discard log: marking anchors `live` from the release description alone — rejected: release notes are not symbol resolution, so `unverified` is the honest state. `provenance: agent-drafted` — rejected: the substance is a faithful restructure of a human-authored, human-agreed design; only the restructuring is the agent's contribution, disclosed here.
