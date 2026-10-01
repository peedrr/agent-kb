---
grammar: 1
type: spec
id: SPEC-999
title: Agents Ratify Their Own Specifications Without Human Review
kind: change
status: active
provenance: agent-drafted
created: 2026-10-01
updated: 2026-10-01
scope: ["docs/knowledge-base/**"]
verified:
  commit: 4b18229034b577024d9ede7199cee638920e7ba6
  branch: main
  at: "2026-10-01T00:00:00Z"
  method: manual-review
  state: live
  next_review_by: "2027-01-01"
---

# SPEC-999: Agents Ratify Their Own Specifications Without Human Review

> **For agents:** this SPEC is authority only while `status: active` AND
> `verified.state: live`. If your task conflicts with it, or any anchor fails
> to resolve at HEAD, stop and name the conflict. Do not silently work around
> it; propose an amendment or supersession instead.

This page is the fail mockup of the `spec` template. It satisfies every rule except one, so the mockup proves that the self-ratification gate fires without a second failure blurring the report.

<!-- FAILS: active_requires_ratified_by — an active, agent-drafted SPEC must name its human ratifier(s) in ratified_by; an agent quoting its own unreviewed text as authority is how phantom requirements ship -->

## Why

A document that adjudicates scope disputes must not derive its authority from the author it constrains. Self-ratified specifications would let an agent mint binding requirements at will.

## Goals and Non-Goals

- Goal: no agent-drafted page reaches `active` without a named human ratifier.
- Non-goal: verifying the ratifier is a real person — provenance of identity is human review work.

## Requirements

REQ-001: The `spec` template MUST NOT admit an agent-drafted page at `status: active` without a non-empty `ratified_by` list; the page MUST remain `proposed` until a human ratifier is named.

## Acceptance Criteria

AC-001 (verifies REQ-001): WHEN this mockup is evaluated by `akb template write spec`, the rule `active_requires_ratified_by` MUST fail and every other rule MUST pass.
  verify: { method: check, check: "akb template write spec --template spec.yaml --pass spec_pass.md --fail spec_fail.md" }

## Decisions and Rejected Alternatives

The gate lives in the template, not in reviewer habit.

- **Trust-the-author process note** — rejected: unenforced conventions are exactly what agents over-trust and silently reverse. Do not re-propose unless the template loses its write-time gate.

## Assumptions and Open Questions

[ASSUMPTION: the fail mockup keeps verified.state at live so the ratifier rule is the sole failure.]

No open questions; any ambiguity that surfaces blocks ratification.

## Affected Surface and Ordering

- `docs/knowledge-base/.agent-kb/templates/spec.yaml`

depends_on: []

## Verification Plan

- The `akb template write spec` run against this mockup reports exactly one failed rule.
- Human-only residue: none beyond the template-write report itself.

## Drift Ledger

- 2026-10-01 · page creation · none → active · recorded at authoring · deliberate violation for proof purposes

## Revisit Triggers

- `docs/knowledge-base/.agent-kb/templates/spec.yaml` — any edit re-runs mockup validation
