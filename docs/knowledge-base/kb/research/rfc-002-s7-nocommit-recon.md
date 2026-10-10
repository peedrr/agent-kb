---
as_of: "2026-10-10"
created: "2026-10-10"
grammar: 1
informs:
- ADR-011
is_draft: false
provenance: agent-drafted
run: rfc-002-s7
scope:
- cmd/akb/**
- internal/storage/**
status: final
summary: --no-commit skips only identity resolution; the merge check always runs in git mode (mode.go:115 before :118, rechecked at git.go:56); ModeNone bases are fully git-free.
tags:
- rfc-002
- git
- preflight
- no-commit
title: RFC-002-S7 nocommit-recon — --no-commit vs merge and identity preflight behavior
type: research
updated: "2026-10-10"
---
# RFC-002-S7 — `--no-commit` vs preflight (merge + identity) recon

Read-only recon. No akb commands run. All anchors from source in `agent-kb`.

## Bottom line

`--no-commit` disables ONLY commit-identity resolution. The merge-in-progress
check ALWAYS runs in git mode, with or without `--no-commit`. So
`akb write --no-commit` CAN fail on a host merge (exit 2), but CANNOT fail on
an unresolvable commit identity.

---

## Q1 — Which commands accept `--no-commit`; how the flag flows

**Definition:** `RootCmd.PersistentFlags().BoolVar(&noCommit, "no-commit", false, "skip git commit")` — `cmd/akb/root.go:66`. It is a persistent global flag, so **every** command parses it; it carries meaning only for commands that open storage. VERIFIED.

**Flow:** each committing command calls `storage.OpenStore(kbRoot, noCommit)`:
- `cmd/akb/write.go:133`, `cmd/akb/append.go:118`, `cmd/akb/approve.go:120` and `:294`
- `cmd/akb/delete.go:208` (`deletePage`), `cmd/akb/index.go:130,254,335`, `cmd/akb/log.go:133`
- `cmd/akb/raw_write.go:77`, `cmd/akb/raw_sync.go:42`, `cmd/akb/raw_delete.go:71`
- `cmd/akb/templates_write.go:256`, `cmd/akb/template_delete.go:99`, `cmd/akb/init.go:270`

`OpenStore` (`internal/storage/mode.go:82-97`) stores it as `Store.noCommit`
(`mode.go:67,94`) and hands it to `newProvider` (`mode.go:46-51`) →
`NewGitProvider(kbRoot, noCommit)` → `GitProvider.noCommit` (`git.go:40-47`).
The `Provider` interface itself carries no versioning semantics; `noCommit` is
carried as a struct field on the git provider, not passed per call. VERIFIED.

## Q2 — `akb write --no-commit`: does Preflight run?

**`Store.Preflight()` — `internal/storage/mode.go:111-122`** (exact order):
```text
if s.mode == ModeNone { return nil }        // 112-114  (mode-none short-circuit)
if err := checkMergeState(s.kbRoot); ...     // 115-117  (ALWAYS in git mode)
if s.noCommit { return nil }                 // 118-120  (skips identity ONLY)
return s.resolveIdentity()                   // 121      (skipped when noCommit)
```text
So the merge check runs unconditionally for a git base and precedes the
`noCommit` short-circuit; the identity resolution is the only piece
`--no-commit` removes. VERIFIED.

**Trace for `akb write`:** `cmd/akb/write.go:539` `store.Lock()` (flock only — no
merge check) → `write.go:547` `store.Write(...)` → `mode.go:156` `Preflight()` →
`mode.go:115` `checkMergeState`. No explicit `Preflight` call in `write.go`;
it is reached through `Store.Write`. `write.go:548` wraps it as `write page: %w`.
VERIFIED.

`append` is identical: `append.go:127` `Lock()` → `append.go:230`
`WriteWithCommitMsg` → `mode.go:163` `Preflight`. Same for `approve.go:175`.
VERIFIED.

**Provider-side second check:** in git mode, even if Store.Preflight were
skipped, `GitProvider.WriteWithCommitMsg` (`git.go:51-83`) runs
`checkMergeConflicts()` (`git.go:56`, → `git.go:198`, → `checkMergeState`) under
its own repo lock, then `commitIdentity()` (`git.go:59`). `commitIdentity`
returns `(nil,nil)` when `g.noCommit` (`git.go:186-191`), so identity is skipped
there too. `GitProvider.Delete` mirrors this (`git.go:104-133`). VERIFIED —
merge is checked twice; identity is never resolved with `--no-commit`.

## Q3 — Can `akb write --no-commit` fail?

**(a) Merge in progress — YES.** With a merge in progress, `mode.go:115`
`checkMergeState` (`git.go:215-254`) runs `git status --porcelain` + `MERGE_HEAD`
and returns `ErrMergeInProgress` (`git.go:202-205`). The command wraps it
(`write.go:548`), and `classifyExit` (`cmd/akb/main.go:80-96`) maps the sentinel
via `errors.Is(err, storage.ErrMergeInProgress)` to **exit 2** with a `usage: `
prefix (`main.go:88`). Message form:
`usage: write page: merge in progress in repository <repo> (outside the knowledge base, in: <changed>): akb commits are blocked ... retry later or surface to the user.`
VERIFIED.

**(b) Unresolvable identity — NO.** `mode.go:118` returns before
`s.resolveIdentity()` (121), and `git.go:188` returns `(nil,nil)`. VERIFIED.

## Q4 — `versioning: none` bases

Independent mechanisms, not a shared path:
- `Preflight` short-circuits `ModeNone` at `mode.go:112` **before** merge and
  identity — a none base runs neither check.
- `newProvider` selects `FilesystemProvider` for `ModeNone` (`mode.go:47-49`),
  which performs only `os.WriteFile`/`os.Remove` — no git at all
  (`filesystem.go:24-50`).
- `Store.Lock` uses `lockFile(path.LockPath(...))` for `ModeNone`
  (`mode.go:143-147`), resolving no git state.
- `Store.Commit` returns immediately for `ModeNone` (`mode.go:222`).

Thus `--no-commit` is a **no-op** in a none base (`mode.go:220` comment;
`test/testdata/init_no_git.txt:58-62`). They are distinct: `ModeNone` skips
*everything* git; `--no-commit` in git mode skips only the commit/identity, not
the merge preflight. VERIFIED.

## Q5 — Tests pinning `--no-commit` vs merge/identity

- **Identity:** `internal/storage/identity_test.go:326-346`
  `TestGitProviderNoCommitNeedsNoIdentity` — pins that a provider with
  `noCommit=true` writes its file with no resolvable identity. VERIFIED.
  `identity_test.go:348-368` `TestStorePreflightReportsTheIdentity` pins the
  *refusal* but with `OpenStore(repo,false)`. VERIFIED.
- **Merge:** `internal/storage/mode_test.go` `TestGitProviderWriteRejectsClean-
  InProgressMerge` (~443) and `TestStoreCommitRejectsCleanInProgressMerge`
  (~470) both build the provider/store with `noCommit=false`. VERIFIED.
- **testscript:** `test/testdata/init_merge_preflight.txt:31` uses
  `git merge --no-commit` (git's own flag, not akb's) and the refusing command
  (`akb raw write`) does NOT pass akb `--no-commit`. The `--no-commit` scenarios
  (`no_commit.txt`, `git_edge_cases.txt:81`, `core_io_verification.txt:134`,
  `init_no_git.txt:60`, `raw_*`, `template_write_noop.txt`) exercise
  no-commit-with-clean-repo only.

**GAP (UNVERIFIED by test):** no test combines akb `--no-commit` with a
merge-in-progress repository. The behavior (`--no-commit` still refuses on
merge) follows from source order at `mode.go:115` vs `:118` and is not pinned by
a test. Recommend a test if the owner decision turns on this.

---

## Residual risks / gaps

- No test pins `--no-commit` + merge-in-progress; conclusion rests on source
  ordering (strong, but untested).
- `akb init --no-commit` follows a different path: it skips `initIdentity`
  (`cmd/akb/init.go:139` `!noCommit`) and stages instead of committing
  (`init.go:276-281` via `Store.StageFiles`). Not traced to merge semantics here.
- I did not run `akb` (per constraints); all findings are static-read.

## Confidence

High on Q1–Q4 (direct source + existing tests). Medium on the merge+no-commit
interaction only insofar as it is untested (Q5 gap), but source order is explicit.

## Supervisor coordination

No blocker; no decision needed from me. Report is complete.
