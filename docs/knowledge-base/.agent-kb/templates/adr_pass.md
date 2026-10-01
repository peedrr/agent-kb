---
grammar: 1
type: adr
id: ADR-000
title: Decision Records Carry Machine-Checkable Guardrails in This Knowledge Base
summary: ADRs in this base follow the docs-writer contract — EARS-2119 constraint statements, a lifecycle state machine, and per-rule verification — enforced by this template at write time.
tags: [adr, governance, templates]
status: proposed
created: 2026-09-30
updated: 2026-09-30
provenance: agent-drafted
scope: ["docs/knowledge-base/**"]
revisit: ["the docs-writer ADR contract changes", "a second grammar version is introduced"]
---

# ADR-000: Decision Records Carry Machine-Checkable Guardrails in This Knowledge Base

> In the context of a knowledge base whose decisions future agents must not silently reverse, facing decision records that drift into unverifiable prose, we decided that every ADR follows the docs-writer contract and is gated by this template's CEL rules at write time and neglected free-form decision notes, to achieve verbatim-quotable, mechanically checkable guardrails, accepting a heavier authoring bar for each record, because a constraint an agent cannot check is a constraint an agent will eventually violate.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

Every architecture decision record in this base MUST follow the docs-writer ADR contract: frontmatter as the retrieval API, EARS-2119 constraint statements, and a verification entry per rule, because a decision that cannot be mechanically checked cannot be mechanically defended.

## Invariants

- **I1**: Every ADR MUST carry the nine contract sections as H2 headings, from Decision through Consequences.

## Negative Constraints

- **N1** (MUST NOT · scope: `docs/knowledge-base/**`): An ADR MUST NOT reach `accepted` without a non-empty human `deciders` list; the record MUST stay `proposed` until a human ratifies it.

## Exceptions

No exceptions are permitted. A case that appears to need one is a proposal to supersede this record.

## Verification

- **I1**: CEL rules `require_decision` through `require_consequences` in this template · gate: `akb write` · mode: **block**.
- **N1**: CEL rule `accepted_requires_deciders` in this template · gate: `akb write` · mode: **block**.
- **Human-only residue**: whether a title or summary is genuinely decision-carrying is judgment; reviewers check at ratification time.

## Context

This mockup is the template's own exemplar: it exists so `akb template write` can prove the rules and the exemplar move together.

## Decision Drivers

- Constraints must be checkable by the tool that gates the page.
- Rejected alternatives must stay rejected without re-litigation.

## Alternatives Considered

- **Free-form decision notes** — rejected: unverifiable prose drifts and gets silently reversed. Do not re-propose unless the CEL gating is removed from this template.

## Consequences

- Good, because every rule carries its own compliance check and remediation.
- Bad, because authoring a record takes longer than writing a note.
