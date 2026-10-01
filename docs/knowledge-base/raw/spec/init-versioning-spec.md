# Spec: `akb init` in existing repositories, non-git KBs, and commit identity

Status: agreed design, pre-implementation. Source: design discussion + three decision
rounds. This file is the implementation contract; when the work lands, its
behavioral pins move into AGENTS.md / the skill / KNOWN-LIMITATIONS.md as noted.

Motivation: dogfooding — a KB that travels with the code it describes (ADRs, tech specs),
versioned in the project's own history. This requires `akb init` to be safe and explicit
inside an existing git worktree shared by multiple agents.

## 1. Init behavior matrix

| Invocation | Context | Result |
|---|---|---|
| `akb init foo` | no enclosing repo | Standalone git repo (today's behavior), no flags |
| `akb init foo` | inside a repo | exit 2 → `--embed` (KB in host history) or `--no-git` (local-only) |
| `akb init .` | repo root | exit 2 → `--embed --force` or `--no-git --force` |
| `akb init .` | not a repo | Standalone repo at `.`, name from dir basename |

- `--force` keeps its single project-wide meaning: bypass the repo-root *layout* warning.
  Mode selection (`--embed` / `--no-git`) is orthogonal and never implied by `--force`.
- `--standalone` does not exist. Nested unregistered repos are not a mode (see §8).
- `akb init .` (anywhere) derives the KB name from the directory basename and reports the
  adopted name in the exit-0 output.
- Refusal wording distinguishes membership: "`<dir>` is inside git repository `<root>`"
  for subdirectories; "`<name>` is a git repository" when the target *is* the root.

## 2. Refusal-message convention (codify)

Exit-2 refusals are self-sufficient: they state what was found, what each flag does, and
the implication of each choice — actionable in a few lines, low token count. The skill
remains the deep resource, but an agent that never read it must be able to proceed
correctly from the error alone. This becomes a written convention in AGENTS.md next to
the exit-code contract.

Refusal for init-inside-repo:

```text
usage: <KB-name> is inside git repository </path/to/containing/repo> —
choose how the KB is versioned:

  --embed    the KB joins this repository's history; akb commits only its
             own paths (kb/, raw/, .agent-kb/), never other worktree changes
  --no-git   the KB is not versioned; akb excludes it from the host's
             git status via .git/info/exclude (local to this clone)
```

## 3. Embedded mode (`--embed`)

- No `git init`, no repo-config writes. The host repo owns identity; see §6.
- Scaffold the standard tree, then commit via the existing `CommitFiles()` pathspec
  machinery: the init commit records exactly `kb/`, `raw/`, `.agent-kb/`, `.gitignore` under
  the KB root. Host worktree WIP — staged or not — is never touched. (This machinery is
  already tested: `TestGitProviderCommitLeavesUnrelatedStagedChanges`,
  `TestCommitFilesRecordsDeletionInNestedKB`.)
- Success message informs: `kb: <kb-name> is embedded inside repository <root>; akb
  commits only kb/, raw/, .agent-kb/`.
- Repo-root layout additionally requires `--force` (§1). The refusal names the
  subdirectory alternative: `This will write .agent-kb/, kb/ and raw/ directly to the repo root. --force if desired, otherwise prefer a subdirectory: akb init <name> --embed.`

## 4. Non-git mode (`--no-git` + `versioning: none`)

Config: `versioning` key in `akb.yaml`, written **explicitly** at init in both modes
(`versioning: git` / `versioning: none`) so the mode is visible and correctable in the
file. `config.Load` treats a missing key as `git` (existing KBs stay valid).

Implementation shape (plumbing, not architecture — `FilesystemProvider` already fully
implements `Provider`):

- Provider factory replaces the ~8 hardcoded `NewGitProvider` call sites; selection is
  driven by the config key.
- The direct `storage.LockRepo` / `CommitFiles` / `NothingToCommit` calls (raw_write,
  raw_delete, raw_sync, template_delete, templates_write, delete, approve, index, write,
  append) become conditional on git mode, or move behind a mode-aware helper.
- Locking: non-git KBs flock `.agent-kb/akb.lock` (`LockRepo` shells out to
  `git rev-parse --git-common-dir`, so it cannot be reused as-is).
- `akb status` currently hard-requires git (`git status --porcelain`); in non-git mode it
  reports `Versioning: none` instead of the git line.
- `--no-commit` flags on other commands are no-ops in non-git mode.
- Inside a repo, init appends the KB dir to the host's **`.git/info/exclude`** (not the
  tracked `.gitignore`) and informs: `kb: excluded docs/ from host git tracking via
  .git/info/exclude (local to this clone; remove that line to undo)`. Not force-gated:
  the file is local-only and the edit is trivially reversible.
- Repo-root layout still requires `--force` (§1).

## 5. `.gitignore` handling — merge, never clobber

Init must append missing lines to an existing target `.gitignore` rather than truncating
it (today `writeGitignoreFiles` uses `os.WriteFile`, which clobbers — this is a bug
regardless of mode). Applies to both files init writes (root `.gitignore`,
`.agent-kb/.gitignore`), making init safely re-runnable. In embedded subdir mode the root
`.gitignore` written is `<target>/.gitignore` (new file, so usually moot); the host's
root `.gitignore` is never touched except in the `akb init .` layout.

## 6. Commit identity

**Init-time resolution.** Resolve in order: `--author-name`/`--author-email` (init flags)
→ `AKB_AUTHOR_NAME`/`AKB_AUTHOR_EMAIL` env → `git config user.name`/`user.email` (merged)
→ default `agent-kb <agent@agent-kb>`.

- If git config yields an identity: **write nothing** to `akb.yaml`. The file travels
  with the KB; a resolved local identity baked into it would be mis-attributed on every
  other machine after push/clone. Exit-0 message: `commit identity: from git config
  (ada <ada@example.com>) — not recorded; each machine's git identity applies`.
- If nothing is found: write the default into `akb.yaml` (`git-author`/`git-email`).
  Exit-0 message: `commit identity: agent-kb <agent@agent-kb> (default — no git identity
  found; recorded in akb.yaml, edit git-author/git-email to change)`.

**Commit-time precedence** (top wins):

1. `AKB_AUTHOR_NAME` / `AKB_AUTHOR_EMAIL` env — the per-agent attribution channel
2. `git-author` / `git-email` in `akb.yaml` (only ever the default or a deliberate edit)
3. Git-native resolution: `GIT_AUTHOR_*` env → repo/global config — akb passes nothing
   and git does its normal thing

When akb supplies the identity (1 or 2), it exports both `GIT_AUTHOR_*` and
`GIT_COMMITTER_*` to the git subprocess, replacing today's `-c user.name=…` mechanism.
Git-native `GIT_AUTHOR_*` set by the user remains honored as a documented escape hatch.

**Preflight.** Committing commands resolve identity *before* any file mutation.
Unresolvable (the push→clone-to-machine-without-git-config case) → exit 2, naming both
remedies: `AKB_AUTHOR_NAME`/`AKB_AUTHOR_EMAIL`, or `git-author`/`git-email` in
`akb.yaml`. The raw git `Committer identity unknown` fatal must never surface.

**Repo config is never written.** `ensureGitConfig` is deleted; `akb init` no longer
writes `user.name` into newly created repos.

## 7. Merge preflight

Git refuses *all* partial commits during any in-progress merge — conflicted or clean
(verified: `fatal: cannot do a partial commit during a merge`). akb commits are always
partial (`--only -- <paths>`), so this coupling is a git constraint, not a relaxable
akb choice. What changes:

- Extend `checkMergeConflicts` to also detect a clean in-progress merge
  (`git rev-parse -q --verify MERGE_HEAD`), closing the gap where today's porcelain scan
  passes and the agent hits the raw git fatal *after* the page was written and staged.
- The message locates the merge relative to the KB and says what not to do:

```text
merge in progress in repository /home/pete/code/agent-kb (outside the knowledge
base, in: src/app.go): akb commits are blocked until it is completed or aborted.
This merge belongs to the host project — do not resolve it from the KB; retry
later or surface to the user.
```

## 8. KNOWN-LIMITATIONS entry (new)

*KB as a git submodule of the host project is unsupported.* A real submodule requires the
KB's own hosted remote (git refuses local-path submodule clones by default since
CVE-2022-39253 — verified git 2.55.0: `fatal: transport 'file' not allowed` without
`protocol.file.allow=always` on every machine), remote/push machinery akb does not have,
and a gitlink-bump commit in the host per KB write. Fresh host clones get an empty
directory until `git submodule update --init` — the opposite of "KB travels with the
code". Manual escape hatch: standalone KB + human-run `git submodule add`. Nested
unregistered repos are not offered by `akb init`; use `--embed` or `--no-git`.

## 9. Also in scope

- `--no-commit` on `akb init`: scaffold + stage, leave the commit to the caller
  (consistency with every other mutating command; lets an orchestrator bundle
  init + templates + overview into one commit).
- Name derivation from basename for `akb init .`, reported at exit 0.
- `.agent-kb/.gitignore` and root `.gitignore` both merge (§5).

## 10. Documentation updates (ship with the change)

- `internal/skill/embedded/kb-management/references/INIT.md`: decision step before the
  command (standalone / embedded / non-git), the embedded procedure with expected stderr,
  the refusal meanings and when `--force`/`--no-git` are the right judgment, and the
  merge-in-progress guidance (do not resolve host merges; retry or surface).
- `SKILL.md`: quick-reference row gains `--embed`, `--no-git`, `--force`, `--no-commit`
  for init; selection section notes the KB-in-repo layout.
- Project `AGENTS.md`: the refusal-message convention (§2), `versioning` config key,
  identity precedence (§6), `--force` meaning unchanged.
- `KNOWN-LIMITATIONS.md`: §8 entry in the established format.

## Out of scope (explicitly)

- `--register-submodule` or any host-repo mutation beyond `.git/info/exclude`.
- Per-command `--author-name`/`--author-email` flags (env vars cover the need; agents
  would have to know which commands commit).
- `akb status` identity line (offered, not requested).
- Remote/push functionality of any kind.
