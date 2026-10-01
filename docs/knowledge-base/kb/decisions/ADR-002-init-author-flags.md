---
created: "2026-10-01"
grammar: 1
id: ADR-002
provenance: agent-drafted
revisit:
- scaffolding agents are found to pass the author flags reflexively, baking ephemeral identities into bases
- an akb command gains ownership of deliberate identity edits, making flag-recording redundant
scope:
- cmd/akb/**
- internal/storage/**
status: proposed
summary: A complete --author-name/--author-email pair at akb init is recorded in the new base's akb.yaml as git-author/git-email; environment and git-config identities are never recorded.
tags:
- init
- commit-identity
- akb-yaml
- attribution
title: Init Author Flags Record a Durable Commit Identity in akb.yaml
type: adr
updated: "2026-10-01T16:14:29Z"
---

# ADR-002: Init Author Flags Record a Durable Commit Identity in akb.yaml

> In the context of `akb init`'s `--author-name`/`--author-email` flags, facing an explicitly typed identity that attributes the init commit and is then discarded — leaving no CLI path to a durable base identity — we decided that a complete flag pair is recorded as `git-author`/`git-email` in the new base's `akb.yaml` and neglected the single-use status quo, to achieve a declared, travels-with-the-base commit identity settable at creation time, accepting that reflexively passed flags now persist beyond their invocation, because a flag is a deliberate declaration by the invoker, not an ambient property of the machine.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

When `akb init` receives both `--author-name` and `--author-email`, the resolved identity is written into the new base's `akb.yaml` as `git-author`/`git-email`, exactly as the akb default is recorded today, and the commit-time precedence is unchanged so the environment still outranks the file — because a typed flag is a deliberate act about the base being created, the same category of act as the deliberate `akb.yaml` edit that is the durable identity channel, while environment and git-config identities remain ambient and unrecorded.

## Invariants

- **I1**: WHEN `akb init` receives both `--author-name` and `--author-email`, the system MUST write the resolved identity to the new base's `akb.yaml` as `git-author` and `git-email`.
- **I2**: WHEN `akb init` records a flag-supplied identity, the exit-0 output MUST state that the identity was recorded in `akb.yaml` and name the keys to edit to change it.
- **I3**: WHEN `akb init` runs with `--no-git`, the system MUST leave the flags with no effect on the base's files; a base that never commits records no identity.

## Negative Constraints

- **N1** (MUST NOT · scope: `cmd/akb/init.go`, `internal/storage/identity.go`): WHEN exactly one of `--author-name`/`--author-email` is passed, the system MUST NOT write the pair — including its git-config-filled field — to `akb.yaml`; the system MUST treat a partial pair as invocation-scoped, exactly as environment and git-config identities are treated.
- **N2** (MUST NOT · scope: `cmd/akb/**`, `internal/storage/**`): The system MUST NOT write an identity resolved from `AKB_AUTHOR_*` or git config to `akb.yaml`; the system MUST record only complete flag pairs and the akb default.

## Exceptions

No exceptions are permitted. A case that appears to need one — a legitimate partial pair, a flag identity that must not persist — is a proposal to supersede this record, recorded as a new ADR.

## Verification

- **I1, N1, N2**: unit tests in `cmd/akb/init_test.go` plus the testscript suite `test/testdata/init_identity.txt`, asserting `akb.yaml` contents for each flag combination · gate: `go test ./cmd/akb/ ./test/` · mode: **block** once implemented · remediation: make the recording behaviour match the invariant; do not weaken the test to match the behaviour.
- **I2**: assertion on the exit-0 notice text in the same suites · gate: `go test ./cmd/akb/ ./test/` · mode: **block** once implemented · remediation: reword the notice to state the recording and name `git-author`/`git-email`.
- **I3**: testscript case `akb init --no-git --author-name … --author-email …` asserting the new `akb.yaml` holds no `git-author` key · gate: `go test ./test/` · mode: **block** once implemented.
- **Human-only residue**: whether the new notice wording is self-sufficient is judgment; whether scaffolding agents passing the flags reflexively becomes a real mis-attribution source is a `revisit` trigger reviewers watch for, not a mechanical check.

## Context

The pre-KB design record described every non-default identity as "resolved" and ruled that it belongs to the invocation or machine, so it is never written to the file that travels with the base. That rationale is sound for ambient sources — `AKB_AUTHOR_*` and git config describe the machine — but "resolved" misdescribes a flag: a flag is explicitly set for this act of creation. Under the status quo there is no CLI path to the durable channel at all; the only route to a recorded identity is hand-editing `akb.yaml` after init, which agents that manage bases through akb cannot always do. [[SPEC-001-init-versioning]] REQ-012 records the implemented rule: an identity named by the init flags, the environment, or git config is never written into `akb.yaml`. This ADR proposes reversing the flag clause of that rule while leaving the environment and git-config clauses intact. The both-or-nothing rule exists because init resolution fills a missing flag field from git config per-field, so a recorded partial pair would bake one machine-local field into the traveling file.

## Decision Drivers

- `akb.yaml` travels with the base across push and clone; ambient machine identity must not be baked into it.
- A typed flag is a deliberate declaration about the base being created — the deliberate edit the durable channel exists for, expressed at creation time.
- Commit-time environment precedence makes a wrongly recorded identity overridable per invocation without editing the file.
- Partial pairs silently mix declared and ambient identity into one record.

## Alternatives Considered

- **Single-use flags (status quo)** — rejected: leaves no CLI path to the durable channel; the invoker must hand-edit `akb.yaml` after init. Do not re-propose unless an akb command exists that owns deliberate identity edits.
- **Record the fully resolved pair even when a field came from git config** — rejected: bakes a machine-local name or email into the traveling file, the exact mis-attribution the no-record rule guards against. Do not re-propose unless init resolution stops filling missing flag fields from git config.
- **Record only the flag-named fields, leaving the other key empty** — rejected: leaves asymmetric half-identities in the file whose completion depends on whichever machine commits next. Do not re-propose unless partial pairs become a recurring real need.

## Consequences

- Good, because a durable, travels-with-the-base identity can be declared at creation time — for example `agent <agent@agent-kb>` for this repository's own knowledge base.
- Good, because the change is confined to `initIdentity` in `cmd/akb/init.go`; the write path already records any non-nil identity.
- Bad, because reflexively passed flags now persist beyond their invocation; the mitigation is that `AKB_AUTHOR_*` outranks the file at commit time, and the `revisit` tripwire fires if this becomes a mis-attribution source.
- Bad, because tests and documentation asserting the single-use behaviour must be reworked: `cmd/akb/init_test.go`, `test/testdata/init_identity.txt`, the flag help text, and the identity pin in the project instructions.
- Neutral, because [[SPEC-001-init-versioning]]'s REQ-012 and AC-019 describe current behavior as corrected on 2026-10-01; on ratification they are amended — or the change is absorbed by the planned `kind: capability` promotion — with a Drift Ledger row pointing here.

## References

- [[SPEC-001-init-versioning]] — the completed change record this ADR extends (REQ-011 through REQ-015).
- `raw/spec/init-versioning-spec.md` section 6 — the pre-KB design record whose "resolved local identity" wording this record corrects.
