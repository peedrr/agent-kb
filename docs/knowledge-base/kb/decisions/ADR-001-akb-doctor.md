---
grammar: 1
type: adr
id: ADR-001
title: KB Diagnosis and Repair Live in a Dedicated akb Doctor Command
summary: A new akb doctor command owns diagnosis and repair of nonconformant KBs; remediation text across akb names an akb command, never raw git or hand-edited config.
tags: [doctor, repair, conformance, agent-ux, error-handling]
status: proposed
created: 2026-10-01
updated: 2026-10-01
provenance: agent-drafted
scope: ["cmd/akb/**", "internal/storage/**"]
revisit: ["akb doctor --fix is asked to convert a base's versioning mode", "an embedded base whose host repository is gone needs an automated repair path"]
---

# ADR-001: KB Diagnosis and Repair Live in a Dedicated akb Doctor Command

> In the context of knowledge bases that carry the `.agent-kb` marker yet fail the requirements of the versioning mode their `akb.yaml` records, facing agents whose sandboxes may permit akb commands but not arbitrary git invocations, we decided that diagnosis and repair of nonconformant bases live in a dedicated `akb doctor` command — read-only by default, repairing only under `--fix` — and neglected extending `akb init` with a `--repair` flag or emitting raw-git remediation text, to achieve in-tool repair discoverable from `akb --help`, accepting a new command surface and its test burden, because `init` connotes create-from-nothing and agents are loss-averse toward running it against a populated base.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

Diagnosis and repair of a nonconformant knowledge base live in a new `akb doctor` command: read-only diagnosis by default, repair only under an explicit `--fix`, and every akb error that can be remediated names the akb command that owns the remedy, because akb is the authorised hands through which agents manage a base and must never send them to raw git or hand-edited config.

## Invariants

- **I1**: WHILE `akb doctor` runs without `--fix`, the command MUST NOT modify base or repository state; the command MUST limit itself to reporting findings and remediation guidance.
- **I2**: WHEN every conformance check passes, `akb doctor` MUST exit 0.
- **I3**: WHEN any conformance check finds the base nonconformant, `akb doctor` MUST exit 1.
- **I4**: WHEN the invocation itself is at fault — an unresolvable base, an unreadable config — `akb doctor` MUST exit 2.
- **I5**: The `akb doctor --fix` repair MUST restore the requirements of the versioning mode the base's `akb.yaml` records.
- **I6**: WHEN `--fix` repairs a git-mode base that has no repository, the repair MUST create the repository and commit exactly the base's own paths (`kb/`, `raw/`, `.agent-kb/`, `.gitignore`), preserving every existing page.
- **I7**: WHERE an akb error is remediable by the user, the error text MUST name the akb command that owns the remedy.
- **I8**: `akb doctor` MUST distinguish a standalone git-mode base with no repository from an embedded base whose host repository is gone, and MUST report the two as different findings.

## Negative Constraints

- **N1** (MUST NOT · scope: `cmd/akb/**`, `internal/storage/**`): Error and remediation text MUST NOT instruct running `git` directly against a base; the text MUST name the akb command that owns the repair.
- **N2** (MUST NOT · scope: `cmd/akb/**`): `akb doctor --fix` MUST NOT change a base's recorded `versioning` mode; the command MUST report mode conversion as an explicit choice outside doctor's repair scope.
- **N3** (MUST NOT · scope: `cmd/akb/**`): `akb doctor --fix` MUST NOT delete, rename, or rewrite an existing page; repairs MUST be additive — creating missing state and committing current content.
- **N4** (MUST NOT · scope: `cmd/akb/init.go`): `akb init` MUST NOT grow repair behaviour for an existing base; its refusal for a nonconformant base MUST name `akb doctor` as the diagnosis path.
- **N5** (MUST NOT · scope: `cmd/akb/**`): `akb doctor --fix` MUST NOT re-home an embedded base whose host repository is missing by initialising a standalone repository over it; the command MUST report the finding and stop.

## Exceptions

No exceptions are permitted. A case that appears to need one — a repair doctor refuses, a remediation that genuinely has no akb command — is a proposal to supersede this record, recorded as a new ADR.

## Verification

- **I1–I6, I8, N2, N3, N5**: testscript integration tests under `test/testdata/` once the command exists · gate: `go test ./test/` · mode: **block** · remediation: make the doctor behaviour match the invariant, do not weaken the test to match the behaviour.
- **I7, N1, N4**: reviewer grep over new error strings — `rg -n 'git (init|add|commit|status)' cmd/akb/ internal/storage/` inspected for messages addressed to the user rather than at the subprocess · gate: PR review · mode: **advisory** · remediation: reword the message to name the owning akb command and what it will do.
- **Human-only residue**: whether a remediation message is *meaningful* — states what was found, what the named command does, and the implication — is judgment; reviewers classify against the exit-2 self-sufficiency convention in the project instructions. Whether a base's nonconformance is intentional is likewise judgment doctor cannot make; it reports, the user decides.

## Context

This decision follows an audit prompted by `akb status` failing on a git-less base with the opaque `git: exit status 128` — stderr was discarded at `cmd/akb/status.go`'s `getGitStatus`, and the same swallow pattern exists at `internal/storage/git.go`'s `isStaged` and `isInHEAD` and at `cmd/akb/init.go`'s `ensureGitConfig`. The audit also found that `akb init` refuses any target whose `.agent-kb/` exists, conformant or not, so a repairable base and a healthy base get the identical blunt refusal. The user constraint that settled the design: agents may be sandboxed to akb commands only, so remediation that requires raw git or a hand-edited `akb.yaml` is remediation they cannot perform. The stderr-swallow fixes themselves are ordinary bugfixes, orthogonal to this record; their messages become doctor pointers once the command exists.

## Decision Drivers

- Agent-first operation: every remediation must be performable through akb itself.
- Discoverability: a broken-base diagnosis must be findable in `akb --help`.
- Loss aversion: agents hesitate — correctly — to run a create-from-nothing verb against a populated base.
- One canonical pointer: every error site should name the same diagnosis command.
- Exit-code consistency: 0 healthy, 1 findings to act on, 2 the command could not do its work.

## Alternatives Considered

- **`akb init --repair`** — rejected: the verb connotes create-from-nothing and reads as potentially destructive against a populated base, and a repair surface hidden behind a create verb is undiscoverable. Do not re-propose unless init's create semantics are split into distinct verbs.
- **Remediation text only, no repair command** — rejected: the only honest text would instruct raw git or hand-editing `akb.yaml`, both outside akb and both possibly forbidden by an agent's sandbox. Do not re-propose unless another akb command exists that owns the repair.
- **Graceful `akb status` on a git-less base (report inline, exit 0)** — rejected: a structurally broken base would look successful to anything keying on exit codes, and it diverges from the write path, which must still fail. Do not re-propose unless status gains a machine-readable health field distinct from its exit code.

## Consequences

- Good, because every git-environment failure across akb gains one canonical remediation pointer — `akb doctor --kb <path>` — instead of per-site advice.
- Good, because the read-only default matches how agents approach a sick system: look first, touch only under an explicit flag.
- Good, because conformance knowledge gets a home that can grow beyond the git check — search.db validity, seed files, templates directory.
- Bad, because a new command and its test surface are added, and `akb init`'s refusal message must still be reworked to point at doctor — two commands change, not one.
- Bad, because an embedded base whose host repository is gone is reported but not repaired in this design; an agent hitting that finding still needs a human.
- Neutral, because converting a base's versioning mode stays outside doctor's scope; a base that should become `versioning: none` needs a deliberate, separate decision rather than a repair flag.

## References

- Swallowed-stderr sites: `cmd/akb/status.go` (`getGitStatus`), `internal/storage/git.go` (`isStaged`, `isInHEAD`), `cmd/akb/init.go` (`ensureGitConfig`).
- Project conventions this record leans on: exit-code contract (0 success, 1 actionable result, 2 could-not-work) and exit-2 self-sufficient refusals, both in the project instructions.
