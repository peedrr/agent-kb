---
grammar: 1
type: spec
id: SPEC-000
title: Spec Records in This Knowledge Base Carry Machine-Checkable Guardrails
kind: change
status: proposed
provenance: agent-drafted
created: 2026-10-01
updated: 2026-10-01
scope: ["docs/knowledge-base/**"]
verified:
  commit: 4b18229034b577024d9ede7199cee638920e7ba6
  branch: main
  at: "2026-10-01T00:00:00Z"
  method: manual-review
  state: unverified
  next_review_by: "2027-01-01"
anchors:
  - claim: REQ-001
    path: internal/cel/engine.go
    symbol: CompileRule
    kind: local
    verify: { method: check, check: "go test ./internal/cel/..." }
    state: live
    checked_at: "2026-10-01T00:00:00Z"
    evidence: "CEL rule compilation gates every typed page write in this repository"
---

# SPEC-000: Spec Records in This Knowledge Base Carry Machine-Checkable Guardrails

> **For agents:** this SPEC is authority only while `status: active` AND
> `verified.state: live`. If your task conflicts with it, or any anchor fails
> to resolve at HEAD, stop and name the conflict. Do not silently work around
> it; propose an amendment or supersession instead.

This page is the pass mockup of the `spec` template: the exemplar `akb template get spec --example` returns, and the proof that the template's CEL rules and their exemplar move together.

## Why

Decision records in this base are already gated by the `adr` template; specifications had no equivalent, so a planning agent had nothing stable and verbatim-quotable to decompose from. This template closes that gap at write time rather than at review time.

## Goals and Non-Goals

- Goal: every spec page passes this template's CEL rules before it is stored.
- Goal: the staleness signal is recorded in the page and re-checkable by anyone.
- Non-goal: judging whether an acceptance criterion genuinely verifies its requirement — that is human ratification work.

## Requirements

REQ-001: WHEN a spec page is written to this knowledge base, the write MUST pass every validation rule of the `spec` template before the page is stored.

REQ-002: The `spec` template MUST NOT let an agent-drafted page reach `status: active` without a named human ratifier; the page MUST stay `proposed` until `ratified_by` names one.

REQ-003: WHILE a spec page carries `kind: capability`, the page MUST include an H2 `Current Behaviour` section.

## Acceptance Criteria

AC-001 (verifies REQ-001): WHEN `akb template write spec` runs with this mockup, the command MUST exit 0.
  verify: { method: check, check: "akb template write spec --template spec.yaml --pass spec_pass.md --fail spec_fail.md" }

AC-002 (verifies REQ-002): WHEN the template's fail mockup — an active, agent-drafted page naming no ratifier — is evaluated, the rule `active_requires_ratified_by` MUST be the only rule that fails.
  verify: { method: check, check: "akb template write spec --template spec.yaml --pass spec_pass.md --fail spec_fail.md" }

AC-003 (verifies REQ-003): WHEN a `kind: capability` page without a `## Current Behaviour` heading is written, the write MUST exit 1 naming `current_behaviour_for_capability`.
  verify: { method: ask, ask: "attempt an akb write of a capability spec omitting the section" }

## Decisions and Rejected Alternatives

The template reuses the ADR template's guard doctrine verbatim rather than inventing a parallel one.

- **Free-form spec notes under kb/** — rejected: unverifiable prose cannot be decomposed or adjudicated. Do not re-propose unless the `spec` template is removed from this base.

## Assumptions and Open Questions

[ASSUMPTION: anchor `state` transitions are recorded by hand at review time until a staleness checker lands in this base.]

No open questions; any ambiguity that surfaces blocks ratification.

## Affected Surface and Ordering

- `docs/knowledge-base/.agent-kb/templates/spec.yaml` — the template itself
- `docs/knowledge-base/kb/specs/**` — where spec pages land

depends_on: []

## Verification Plan

- Template acceptance: the `akb template write spec` invocation that admitted this mockup, re-run green.
- Human-only residue: whether a title is genuinely change-carrying; whether each criterion adds an observable boundary; closed-world typo review (akb checks field presence, not unknown keys); Preserved Behaviour content for change specs touching existing code.

## Drift Ledger

- 2026-10-01 · page creation · none → proposed · recorded at authoring · the template write that admitted this mockup

## Revisit Triggers

- `docs/knowledge-base/.agent-kb/templates/spec.yaml` — any edit re-runs mockup validation
- a second grammar version is introduced
