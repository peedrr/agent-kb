---
created: "2026-10-10"
grammar: 1
id: ADR-011
is_draft: false
provenance: agent-drafted
revisit:
- a mutating command gains a validation step whose feedback depends on git state — ordering I1 needs re-examination for that command
- git merge semantics change to include unstaged paths in a merge commit — the N2 bypass loses its safety argument
- the post-RFC-002 reorder work unit lands — the transitional inconsistency named in Consequences must be gone
scope:
- cmd/akb/**
- internal/storage/**
status: proposed
summary: All mutating commands validate first and run the git-state preflight (merge, identity) immediately before the storage mutation; --no-commit skips every git-state check, not just identity.
tags:
- rfc-002
- git
- preflight
- validation
- ux
title: Validation Precedes Git-State Checks and --no-commit Skips Them Entirely
type: adr
updated: "2026-10-10"
---

# ADR-011: Validation Precedes Git-State Checks and --no-commit Skips Them Entirely

> In the context of [[RFC-002-open-validation-engine|RFC-002]]'s validation-pipeline work and
> multi-agent use of akb, facing an inconsistent ordering in which `write`/`append` run the
> merge/identity preflight inside `store.Write` — after validation — while seven other mutating
> commands preflight before their first mutation, and a `--no-commit` flag that skips only
> identity resolution while still refusing on an in-progress merge, we decided that every
> mutating command validates first and runs the git-state preflight immediately before its
> storage mutation, and that `--no-commit` skips all git-state checks, and neglected the RFC-002
> §7 literal merge-check-first order and the working-tree-protection reading of the merge guard,
> to achieve validation feedback on the artifact the author controls before environmental
> refusals that may clear on their own, accepting that a `--no-commit` write can land in a
> merging working tree and that RFC-002 §7's pipeline text needs amendment, because the time an
> agent spends fixing a validation error is exactly the window in which another agent's merge
> finishes, and a flag that promises no commit makes git-state preconditions irrelevant.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not
> silently work around it; propose supersession instead.

## Decision

Every mutating command orders its work as: page validation and preconditions first, git-state
preflight (merge state, commit identity) immediately before the storage mutation it guards.
`--no-commit` disables the git-state preflight entirely — merge check and identity resolution —
because no commit will be attempted, so git-state preconditions cannot be preconditions of the
write. The current code does neither: `write`/`append` preflight late by accident (inside
`store.Write`), the other mutating commands preflight early on purpose, and
`Store.Preflight` runs `checkMergeState` before its `noCommit` short-circuit
(`internal/storage/mode.go:115` vs `:118`).

## Invariants

- **I1**: WHEN a mutating command runs, the system MUST evaluate page validation and
  preconditions before any git-state preflight.
- **I2**: The git-state preflight MUST run immediately before the storage mutation it guards;
  the system MUST NOT run a git-state preflight before validation.
- **I3**: WHEN `--no-commit` is passed, the system MUST skip every git-state check; the system
  MUST perform only the filesystem write under the store lock.
- **I4**: The internal preflight inside `store.Write`/`WriteWithCommitMsg` MUST remain as the
  last-line safety net for every path, including paths that preflighted earlier.
- **I5**: `versioning: none` bases MUST remain free of git invocations, as they are today.

## Negative Constraints

- **N1** (MUST NOT · scope: `cmd/akb/**`): A mutating command MUST NOT invoke
  `store.Preflight()` before its validation steps; the command MUST place the preflight
  immediately before its storage mutation.
- **N2** (MUST NOT · scope: `internal/storage/**`): `Store.Preflight` MUST NOT run
  `checkMergeState` when `Store.noCommit` is set; it MUST return before any git invocation, the
  way it already returns before `resolveIdentity`.
- **N3** (MUST NOT · scope: `cmd/akb/**`): No command MAY define its own ordering of validation
  versus git-state preflight; every command MUST follow the I1–I2 order.

## Exceptions

`akb init` is excepted: its identity flow (`ResolveInitIdentity`) is creation-time configuration
of a new base, not a mutation of an existing one. No other waivers are permitted.

## Verification

- **N1/I2**: grep audit — `rg -n 'Preflight' cmd/akb/` — every call site sits after the
  command's validation steps · gate: PR review until the reorder work unit lands, then
  test-pinned · mode: **block** (after the work unit) / **advisory** (before).
- **N2/I3**: unit test — `Store` with `noCommit=true` in a merging repository writes without
  invoking git (the current behavior is unpinned: no test combines `--no-commit` with a merge
  in progress) · mode: **block** · validate: `go test ./internal/storage/ -run NoCommit`.
- **I4**: existing provider tests (`TestGitProviderWriteRejectsCleanInProgressMerge`,
  `TestStoreCommitRejectsCleanInProgressMerge`) MUST stay green.
- **Human-only residue**: "immediately before the storage mutation" is a code-review judgment;
  reviewers check that no validation or I/O step intervenes between preflight and mutation.

## Context

akb is operated by agents, often several against one repository, and agents move fast: a merge
that blocks a write at 10:00:01 is frequently resolved by 10:00:20. Recon for SPEC-004
([[rfc-002-s7-code-recon]], [[rfc-002-s7-nocommit-recon]]) established that RFC-002 §7's stated
pipeline (`merge-conflict check → … → write`) does not match the code — `write`/`append` reach
the preflight only inside `store.Write` — and that `--no-commit` was implemented as an
identity-only bypass (`mode.go:118`), leaving the merge guard firing even when no commit is
possible. The merge guard's original purpose is protecting the auto-commit, not the working
tree: git never sweeps unstaged paths into a merge commit.

## Decision Drivers

- Feedback should concern the artifact the author controls (the page) before the environment
  they do not (repository git state).
- Flag semantics should be honest: `--no-commit` promises no commit; every git-state check
  exists to protect a commit.
- One ordering for all mutating commands; per-command divergence has now been rediscovered
  twice by recon.
- Environmental faults may clear during validation repair; validation faults never clear
  during a git-state wait.

## Alternatives Considered

- **Literal RFC-002 §7 order (merge check first, all commands)** — rejected: blocks validation
  feedback behind a fault that may clear on its own; punishes the common multi-agent case.
  Do not re-propose unless akb's usage becomes single-operator with rare concurrent merges.
- **Keep the merge check under `--no-commit` (working-tree protection)** — rejected: git never
  includes unstaged paths in a merge commit, so a `--no-commit` write is no more entangled than
  an editor save; the guard protects a commit that will never happen. Do not re-propose unless
  git merge semantics change to include unstaged paths.
- **Status quo (per-command ad hoc ordering)** — rejected: the divergence between `write`/
  `append` and the other seven mutating commands is invisible in any single diff and has
  required recon to establish twice. Do not re-propose.

## Consequences

- Good, because validation errors — the feedback the author can act on — arrive first, and a
  transient merge no longer masks them.
- Good, because `--no-commit` becomes a complete git bypass, matching operator expectation and
  making agent scratch writes independent of repository state.
- Bad, because a `--no-commit` write can now land in a merging working tree; the files stay
  unstaged and outside the merge, but an operator resolving the merge sees unfamiliar untracked
  changes. Accepted: the mitigation is git's own isolation plus the flag being opt-in.
- Bad, because nine commands' preflight call sites move and `Store.Preflight` changes contract;
  this is a dedicated post-RFC-002 work unit, not folded into SPEC-004. Accepted: until it
  lands, the codebase remains in the transitional inconsistency this ADR names.
- Neutral, because RFC-002 §7's pipeline text is amended to state the git-state preflight in
  its true position — immediately before the storage mutation — rather than first.

## References

- [[RFC-002-open-validation-engine|RFC-002]] §7 (pipeline text this ADR amends) and §18 (work-unit row).
- Research: [[rfc-002-s7-nocommit-recon]], [[rfc-002-s7-code-recon]] (D5), [[rfc-002-s7-log]].
- [[ADR-002-init-author-flags|ADR-002]] (commit-identity resolution the preflight invokes).
- [[SPEC-004-schema-pass-integration|SPEC-004]] (validation-stage owner; makes no preflight-ordering claims).
