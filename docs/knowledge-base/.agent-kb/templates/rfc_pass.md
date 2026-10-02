---
grammar: 1
type: rfc
id: RFC-000
title: RFC Pages Carry Machine-Checkable Lifecycle Guardrails in This Knowledge Base
status: draft
provenance: agent-drafted
author: pi/session
steward: pete
created: 2026-10-02
updated: 2026-10-02
scope: ["docs/knowledge-base/**"]
tags: [rfc, governance, templates]
summary: Proposal pages in this base carry a machine-checked lifecycle, identity and staleness contract; the section, evidence and EARS gates arrive with the V3 engine.
spawns: []
---

# RFC-000: RFC Pages Carry Machine-Checkable Lifecycle Guardrails in This Knowledge Base

> **Reviewer brief** — Decision sought: adopt this minimal interim RFC template. · Recommended outcome: yes. · Change if accepted: RFC pages become type-checked in this base. · Affected surface: `docs/knowledge-base/**`. · Stakes: low, interim until the V3 engine.

> **For agents:** this is a PROPOSAL. It is authority only while `status: accepted | implemented`, not superseded, and not `is_stale` — and only for what was decided, never for current behaviour. You MUST NOT implement from an unratified RFC. If your task conflicts with a ratified RFC, stop and name the conflict; propose supersession.

This page is the pass mockup of the `rfc` template: the exemplar `akb template get rfc --example` returns, and the proof that the template's CEL rules and their exemplar move together.

## Problem

Proposal pages in this base had no typed home and no machine-checked lifecycle. A parked proposal could not be told from a ratified one, nothing surfaced a proposal whose revisit date had passed, and the discussion that produced a deferral lived only in chat. The docs-writer RFC contract states the lifecycle; nothing in the tool enforced it.

## Goals and Non-Goals

- Goal: every RFC page passes this template's CEL rules before it is stored.
- Goal: staleness is a derived posture, reported by the sweep and never stored.
- Non-goal: enforcing the ten-section skeleton or the EARS-2119 register — deferred to the V3 engine.

## Proposal

Adopt the minimal interim template: frontmatter is the machine contract, the status enum and transition matrix are checked at write time, identity fields are required and the ratifier must differ from the author, and derived staleness postures are reported by `akb lint`.

## Proposed Constraints

- **C1** (→ SPEC): The write gate MUST refuse a proposal whose `decided_by` equals its `author`; the page MUST stay unratified until a different human ratifies it.

## Alternatives

- **Do nothing** — keep proposals as untyped markdown outside the base. Cost: nothing is checked, nothing is linkable, and deferrals stay in chat.
- **Wait for the V3 engine** — the deferred path. Honest, but it leaves the document family incomplete and the existing proposals homeless.

## Risks and Drawbacks

The interim format will be re-authored once V3 lands; section and evidence gates are absent, so the body contract rests on review alone.

## Assumptions and Open Questions

[ASSUMPTION: the two-actor convention — author is the drafting actor, steward and decided_by are the human — holds until the actor RFC lands.]

## Deliberation

- Reviewers: pete (owner). Kind of review wanted: scope and shape of the interim template.
- Verdicts recorded as they arrive, one per reviewer.

## Resolution

Not filled; this proposal is unratified.

## Evidence and Prior Art

docs-writer `references/rfc.md`; the rfc-templates research REPORT and SYNTHESIS; RFC-002 for the V3 engine.
