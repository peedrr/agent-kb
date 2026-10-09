---
author: pi/session
created: "2026-10-02"
grammar: 1
id: RFC-003
provenance: agent-drafted
scope:
- docs/knowledge-base/**
- cmd/akb/**
- internal/**
spawns: []
status: draft
steward: pete
summary: Expose a resolved actor identity to akb's rule engine so templates can assert who drafted, stewards, and ratifies — mechanism in akb, roster policy in the KB.
tags:
- identity
- authorization
- validation
- governance
title: Give akb a Declared Actor Identity So Rules Can Assert Who Acted
type: rfc
updated: "2026-10-09T16:35:25Z"
---

# RFC-003: Give akb a Declared Actor Identity So Rules Can Assert Who Acted

> **Reviewer brief** — Decision sought: give akb a document-independent `actor` value that CEL rules can read. · Recommended outcome: yes, with the roster declared in KB config and templates. · Change if accepted: ratification and freeze rules can assert *who* acted, not only what the frontmatter claims. · Affected surface: the validation pipeline and `akb.yaml`. · Stakes: medium — a new engine capability, sequenced after RFC-002 V3.

> **For agents:** this is a PROPOSAL. It is authority only while `status: accepted | implemented`, not superseded, and not `is_stale` — and only for what was decided, never for current behaviour. You MUST NOT implement from an unratified RFC. If your task conflicts with a ratified RFC, stop and name the conflict; propose supersession.

## Problem

Design discussions between the owner and his agents happen in chat, and they evaporate when the session closes — the options considered, the risks, the likely direction, the preconditions that would unblock a decision, and often the decision to defer. The knowledge base could already record a decision (an ADR) and a unit of work (a SPEC); as of the interim `rfc` template it can now carry an undecided proposal, and this page is the first. But a proposal's authority rests on claims about *who* drafted, stewarded, and ratified it, and akb cannot check any of those claims. RFC-002's document model (section 5.1) materializes `file`, `frontmatter`, `content`, and parse facts — no actor. Its config surface (section 14) adds lint, validation, git and search keys, but `git-author`/`git-email` is a commit identity for auto-commits and is never visible to a rule. Its identity guarantee (section 16) stops at "a commit has an identity". So `decided_by != author` is a string comparison the drafting agent controls, and docs-writer's rule that an agent must not ratify is unenforceable. The 2026-10-02 session that produced the interim template — and deferred this question — is the motivating example: without a record, the discussion would live only in the owner's memory.

## Goals and Non-Goals

- Goal: akb exposes a resolved actor identity to the rule engine as a document-independent value, beside `old_page` and `now`.
- Goal: templates and KB config can assert actor-relative rules — who may ratify, who may abandon, who may change a frozen field.
- Goal: the actor is recorded in commit history so attribution survives outside the engine.
- Non-goal: akb shipping an authorization policy or a roster of people. That is KB and template policy (RFC-002 principle P2).
- Non-goal: multi-tenant or federated identity.
- Non-goal: replacing git's own commit identity.

## Proposal

akb resolves an `actor` value for each invocation from a declared chain and materializes it for CEL alongside `old_page` and `now`. The chain's exact shape belongs to the spawned ADR; the intended order is an explicit `AKB_ACTOR` environment value, then a KB-config mapping from the invoking machine or user to a declared actor, then the git commit identity when one resolves, and otherwise an explicit `unknown` actor rather than a silent default. The value carries at least an identifier and a declared kind (`human` or `agent`), so a rule can say `actor.kind == "human"`. Templates then express policy with rules they already know how to write: `decided_by == actor.id` requires the ratifier to be the acting human, `actor.kind == "human"` forbids agent ratification, and `actor.id == steward` guards the transition to `abandoned`. The KB declares the roster — identifiers, kinds, and any role names — in `akb.yaml`; the engine ships the capability and example templates only. The actor is also written as a commit trailer so `git log` answers "who acted" without akb. This is sequenced after RFC-002 V3 deliberately: V3 already replaces the document model and adds document-independent bindings, which is the natural place for one more.

## Proposed Constraints

- **C1** (→ ADR): The engine MUST resolve an actor identity for every invocation that reaches the write gate.
- **C2** (→ SPEC): WHEN a template rule references the actor value, the engine MUST evaluate that rule at write time and fail closed on a false verdict, exactly as for any other rule.
- **C3** (→ SPEC): akb MUST NOT ship a built-in authorization policy; akb MUST ship the actor value and example templates, and a KB MUST declare its own roster.

## Alternatives

### Do Nothing

Keep actor claims self-declared. Nothing new is built, and the current convention — author is the drafting actor, steward and ratifier are the human — continues to work for a solo owner. The cost is that every identity claim is only as good as the writer's honesty: an agent can write any `decided_by`, so the ratification gate, the steward-only `abandoned` write, and the freeze rules are advisory. This is the status quo, and it already fails the requirement that an agent must not ratify.

### A Template-Declared Enum Roster

Restrict `author`, `steward`, and `decided_by` to an enum declared in the template. It is cheap, and it makes string equality meaningful because typos and formatting variants fail at write time. But it hardcodes people into a template, breaks with the first new collaborator, and still provides no attestation: the writer types the value either way.

### Cryptographic Attestation

Require signed commits or page signatures so the actor is proven rather than declared. This is the strongest option, but it proves possession of a key, not a role; it imposes signing infrastructure on every KB; and it does not by itself say which actors may ratify. It is a plausible successor to this proposal, not a substitute for the actor surface.

### Actor Identity From Git Only

Read the existing commit identity in rules without adding a variable. It adds no surface and no config, but the value is unavailable in `versioning: none` bases and before a commit exists, and it cannot express roles or kinds, so it cannot distinguish a human from an agent.

## Risks and Drawbacks

A declared actor is spoofable through the environment variable that supplies it, and users may over-read a declared identity as an authenticated one; the reference must state that boundary explicitly. Bases running `versioning: none` have no git identity to fall back on, so the resolution chain must define an explicit `unknown` and rules must handle it. The feature depends on RFC-002 V3, so it cannot land before the engine it extends. Finally, there is a standing temptation to let convenience authorize a policy in the engine; C3 exists to hold that line.

## Assumptions and Open Questions

[ASSUMPTION: RFC-002 V3 lands and is the right host for a document-independent actor binding, because it already introduces `now` and `kb.*` bindings.]

[ASSUMPTION: two actor kinds — human and agent — suffice for this project; additional kinds are a later extension, not a design constraint.]

- [NEEDS CLARIFICATION: the resolution order, and whether the roster lives in `akb.yaml` or per-template — owner: pete — decide by: when the V3 config surface (Phase 3) is specified]
- [NEEDS CLARIFICATION: how an actor is recorded in `versioning: none` bases, where no commit identity exists — owner: pete — decide by: before V3 Phase 4]
- [NEEDS CLARIFICATION: whether any rule may rely on a declared rather than attested actor, and what the reference must warn about — owner: pete — decide by: at ratification]

## Deliberation

- Reviewers: pete (owner). Kind of review wanted: direction and scope, and whether this belongs in akb at all rather than in template convention.
- Origin: the 2026-10-02 session on the interim `rfc` template. Drafted by pi/session under the operating human pete.
- Discussed and deferred there: the three identity representations (structured map, canonical handle, enum); the two-actor convention (an agent drafts, the human stewards and ratifies); the self-declared-versus-attested fork; and whether authorized actors belong in `akb.yaml`.
- Disposition of that discussion: captured here, as the first page of the new `rfc` type, at `status: draft`, so it can be revisited instead of lost.

## Resolution

Not filled. Ratification is human-only and requires `decided_by` to differ from `author`; neither is set while this page is `draft`.

## Evidence and Prior Art

- docs-writer `references/rfc.md` — the proposal lifecycle this page follows, including human-only ratification and the status-gated authority boundary.
- The research run that produced the RFC contract: [[rfc-templates-report|REPORT.md]] and [[rfc-templates-synthesis|SYNTHESIS.md]] (run log: [[rfc-templates-log|LOG.md]]) — performed in the pi-meta-config repository's scratch and backfilled into this KB under [[ADR-007-research-reports-first-class|ADR-007]]; approver-distinct-from-author (KEP), human arbitration (AIDR), and the two-axis disposition model (CSSWG) are the relevant prior art; its eighteen claims were independently audited in [[rfc-templates-audit-report|the audit report]] (18 supported, 0 corrected, 0 killed).
- `kb/rfc/RFC-002-open-validation-engine.md` sections 5.1, 14, 15 and 16 — the document model with no actor surface, the config surface, the harness flow, and the commit-identity chain that stops at git.
- The 2026-10-02 session that produced the interim `rfc` template — this RFC's origin and motivating example.
