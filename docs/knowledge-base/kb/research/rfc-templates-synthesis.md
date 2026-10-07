---
as_of: "2026-10-02"
created: "2026-10-02"
grammar: 1
informs:
- RFC-003
provenance: agent-drafted
run: rfc-templates
scope:
- "docs/knowledge-base/**"
status: final
summary: "Coordinator's design analysis of the RFC as deliberation-stage process unit: its machine contract, ratification gate, lifecycle postures, and the ratified forks F1–F5."
tags:
- synthesis
- rfc-format
- design-proposals
- agent-consumption
title: "RFC format for agents — synthesis and implication analysis (rfc-templates)"
type: research
updated: "2026-10-02"
---
# SYNTHESIS — RFC format for agents (coordinator's implication analysis)

Status: all 4 lanes + recon consumed; owner forks F1–F5 ratified; audit COMPLETE
(`audit-report.md`: 18 supported, 0 corrected, 0 killed; strength-level caveats
folded into REPORT). Deliverables shipped.
This document is the coordinator's own analysis — child reports are evidence, not design.
Sources: `recon-*.md`, `lane-*.md` in this directory; transferred evidence from
`../adr-templates/` and `../spec-templates/` keeps its prior audit status.

---

## 1. The problem restated, precisely

The owner needs an RFC document that serves as the **deliberation-stage process unit**
in an agent-driven SDLC — the one document type whose whole point is *not yet decided*.
ADR = binding decision record; SPEC = frozen unit-of-work contract; RFC = the proposal
that, when ratified, *spawns* ADRs/SPECs. The mechanical roles, distilled from the
local consumer contracts (recon-local Q4) and the owner's ratified intent (R1):

1. **Deliberation carrier** — an agent drafts it; the owner (and optionally advisors)
   reviews it; it must make the argument *falsifiable* (claims traceable, alternatives
   honest, assumptions marked). Its worst defect is fluent emptiness: "five convincing
   pages before anyone has had a single serious thought" (lane-agent-era F39).
2. **Ratification gate** — a human flips it to binding; the flip must be machine-
   checkable (status enum + approver ≠ author + human-only arbitration), because the
   adjudication machinery downstream keys on it (KEP verbatim: "Approvers should be a
   distinct set from authors"; AIDR: "Arbitration is human").
3. **Status-gated authority for decision provenance** — owner intent (R1): a ratified +
   current RFC is an authoritative source for the fix-designer's auto-resolution, with
   akb tracing ADR/SPEC → RFC lineage. Evidence constrains the shape: ratified-for-
   *decision* is near-universal (KEP/PEP/Rust/IETF); ratified-for-*current behaviour*
   is contradicted (PEP: "historical document rather than a living specification";
   Oxide's editable-consensus paradox). ⇒ The RFC is quotable authority for **what was
   decided, by whom, when, on which revision, and why** — never for how the system
   currently behaves (that is the SPEC/code surface).
4. **Stall detection** — the measured failure mode of proposal pipelines is stalling,
   not bad writing (marzouk: ~60% of RFCs stale >6 weeks; lane-lifecycle F37). Hard
   expiry failed at IETF ("illusory… wasted effort"); stored status rots (KEP 617).
   ⇒ Staleness must be a **derived posture** (CEL, recomputed every run, cannot rot),
   and the closing act must remain an accountable write.
5. **Load-path definition** (transferred, spec V16): a format that defines only
   sections gets read at 10.6%. The RFC must state who reads it, when, and which gate
   cannot be passed without it.

## 2. Design verdicts (each with basis; [JUDGMENT] where no evidence exists)

| # | Verdict | Basis |
|---|---|---|
| V1 | **Closed-world YAML frontmatter is the machine contract** — status enum, steward, review_by, scope, tags, summary, supersession, spawns. Prose is never machine-checked | kepval verifies the typed control file only, never prose (lane-verify #2); marzouk's production extractor exists *because* prose hallucinates (lane-agent-era F15); sibling parity (adr.md §3, spec.md §3) |
| V2 | **Status enum: `draft \| review \| accepted \| implemented \| rejected \| superseded \| abandoned`**, human-only ratification flip, approver ≠ author checkable, `accepted` requires steward + review_by (anti-graveyard guard) | lane-lifecycle recommendation adapted; KEP implementable/implemented split (decided ≠ built); PEP Accepted ≠ Final; zby/commonplace counter (5/5 `proposed` ADRs were stale markers) answered by the guard, not by deleting the state; `rejected` retained (Lubow Documented-No: prevents re-proposal without new information; universal except AIDR) |
| V3 | **Freeze at ratification; change = supersession** (owner R2). Record honestly: stricter than every institutional precedent; nearest support is AIDR's append-only-at-arbitrated + Rust's "new RFCs, not edits" prose rule | lane-lifecycle contradiction #1; AIDR (lane-deliberation F19); Rust RFC README |
| V4 | **Deliberation expiry = `review_by` + derived postures; the clock never writes a status.** CEL-derived: `is_at_risk` (14d warning), `is_stale` (past review_by, unsuspended), `is_dormant` (draft silence), `is_terminal`, `needs_steward`. `abandoned` is a steward-only write with a written reason | IETF hard-kill failure (lane-lifecycle F4/F6); KEP 617 rot (F9 — derived cannot rot); Nudge/NudgeBot vs Stale-bot evidence (F15: reminders to an accountable actor work; blanket auto-close harms); FEP operator report (notice-then-withdraw doesn't work); owner's CEL note (R3) |
| V5 | **Timeout is never agreement.** Any silence-based closure is a distinct terminal value (`assumed_satisfied_on_timeout` ≠ `satisfied`) | CSSWG label set (lane-deliberation F5) — the single best-supported deliberation element |
| V6 | **Deliberation/Resolution sections, not a full per-comment DoC schema.** `## Deliberation` (reviewers, review_by, kind of review wanted) during review; `## Resolution` at ratification (decision, human decider, verdict-per-reviewer on the `yes \| yes_if(cond) \| not_yet` enum, blocking-concern dispositions on CSSWG's two axes, resolution link). Full YAML DoC record parked as future akb work | Squarespace three-valued verdict ≈ rfc-critic's Approve/Approve-with-changes/Needs-redesign (independent convergence); CSSWG two-axis model (disposition ≠ commenter acceptance); IESG blocking/non-blocking + clearer-identity; single-ratifier harness today — the full record is over-engineering for one reviewer [JUDGMENT] |
| V7 | **EARS-2119 adopted for RFC constraint statements, scoped to designated sections** (§6 below — the instruction's named question) | probe-keywords (transferred); sibling parity; analysis in §6 |
| V8 | **Per-section-type prose rule** (diverges from sibling bullets+fragments): frontmatter/tables = machine contract; enumerations = lists; **claims/rationale/alternatives = narrative prose with a budget**. Constraint statements = EARS-2119 regardless | lane-agent-era contradiction #1 resolution (lemieux's explicit rule, F27); slop-signal evidence (nesbitt heuristics; Calçado's bullets-into-plausible-prose); humans review RFCs (the owner ratifies) — the slop signal has a reader here; ADR/SPEC are records consumed by agents mid-task, RFCs are arguments consumed at decision time |
| V9 | **Implementation-detail exclusion WITH a positive allowance**: shape yes (interface sketch, pseudocode, toy example), artifacts no (filenames, method signatures, DDL, migration sequencing, config examples). The RFC names its handoff target (`spawns`) so the prohibition has a destination | Calçado 2018+2026 (pre-AI, 10 years); lemieux "became a spec" list; spec-kit/SuperSpec/prd-to-plan countermeasures; create-rfc "RFC is for decisions, not implementation" |
| V10 | **Mandatory Do-Nothing alternative + straw-man check**; alternatives are *live contenders* during deliberation and get "do not re-propose unless …" only at Resolution (when actually rejected) | Calçado ("the only part of NABC I'm really married to"); create-rfc; adr.md §4.8 parity for the rejected form; live-vs-rejected distinction is the proposal-vs-record boundary (recon-landscape §5) |
| V11 | **Length budget as a declared field + section-share guidance; machine-dense blocks (diagrams, API contracts) exempt** | Calçado ("religious about having a budget", not the number); lemieux (500–1000 lines, Proposal 40–60%, dense-block exemption = the token lever); Lubow over-indexing cost test |
| V12 | **Evidence traceability on load-bearing claims**: claim↔source, machine-resolvable; the reviewer agent can check support, not just existence. Rubric pre-flight section in the reference (Calçado's five check categories + lemieux mechanics: category, severity, confidence ≥80, quote-the-location, never-suggest-fixes) | Calçado F6/F8; lemieux reviewer prompt; four independent implementations of rubric pre-flight — highest-confidence convergent move (lane-agent-era) |
| V13 | **Typed provenance enum + accountable ownership + no agent as sole arbiter**: `provenance: human \| agent-drafted` (sibling parity) + human ratifier required; the RFC MUST NOT be ratified by an agent | Calçado/Lubow accountability; 51%/74% OSS-policy study (disclosure contested → typed field, not prose obligation); Blender/Fedora "no AI as sole arbiter"; AIDR human-arbiter rule; sibling provenance fields |
| V14 | **Reviewer-brief BLUF block at the top** (the "ask" in ≤6 lines: decision sought, recommended outcome, change-if-accepted, affected surface, stakes) + decision-carrying title + ≤200-char summary | eugenelim new-rfc Reviewer-brief grid; RFC 7322 abstract self-containment (lane-verify #3); Vercel index evidence (transferred: front-matter lines are what gets routed on) |
| V15 | **Assumptions/open questions with markers; unresolved `[NEEDS CLARIFICATION]` gates ratification**; genuine open questions capped (~3) with owner + decide-by | spec.md §4.9 parity; eugenelim cap; OpenShell "preserve uncertainty"; lemieux's ban rejected (context-poor agent drafters need the channel) [JUDGMENT on the cap] |
| V16 | **Load path stated in the contract** (§5 below): drafting agent → owner ratification → adjudication consumption (status-gated) → staleness sweep (derived postures) → spawn transcription | spec V16 transfer; lane-lifecycle (derived posture is sweep-computable); owner R1 |
| V17 | **Bidirectional supersession + `spawns` lineage field** (`spawns: [RFC→ADR/SPEC candidates]` at proposal; links as they land; reverse `spawned_by` on the children when authored) | PEP Replaces/Superseded-By; KEP replaces/superseded-by; adr.md/spec.md same-commit rule; owner's akb-lineage intent (R1); exemplar "ADR/tech-spec candidates" sections (recon-local Q1e) |
| V18 | **Self-containment hard requirement**; Research base/Evidence section is additive, never required reading | adr run V3 (audited); exemplars carry "Research base" (transferred owner convention) |
| V19 | **Minimal ceremony within the skeleton; one proposal per RFC; narrow significance bar** (a ratified RFC must be expensive-to-reverse or cross-cutting enough to justify spawning ADRs/SPECs) | McMillan structural null (transferred); PEP 1 single-key-proposal bar; sibling significance bars |
| V20 | **No measured comprehension benefit is claimable**: the format sells falsifiability, governability, stall-detection — never quality | lane-agent-era F38 (confirmed negative: no proposal-format comprehension study); spec run's self-claim boundary parity |

## 3. Resolving R1 (the deferred authority fork)

Owner intent: ratified + current RFC = additional authoritative source for
fix-designer's auto-resolution; akb traces lineage. Evidence: ratified-for-decision
yes / ratified-for-behaviour no (§1.3). **Resolution:**

- A ratified RFC (`status: accepted | implemented`, not superseded/abandoned, derived
  posture not `is_stale`) IS quotable authority for: the decision, its rationale, its
  rejected alternatives (with their do-not-re-propose conditions), and any EARS-2119
  constraint statements **not yet transcribed** into a spawned ADR/SPEC.
- It is NOT authority for current system behaviour; once a constraint is transcribed
  into an ADR/SPEC, the ADR/SPEC owns it (precedence `adr > spec > plan` unchanged;
  the RFC stands *behind* the chain as provenance root, never above it).
- The gate is mechanical and matches the existing pattern (`spec-builder.md:56–58`):
  failing status/staleness ⇒ advisory only, recorded as discoveredRisks.
- Untranscribed-constraint authority has an expiry flavour: an `accepted` RFC whose
  `review_by` passes without spawn transcription goes `is_stale` ⇒ advisory. This
  makes the handoff enforceable rather than ceremonial.

This is a [JUDGMENT] mapping onto evidence-backed components; it gives the owner's
stated intent with the PEP/Oxide boundary honoured.

## 4. The proposed format (contract sketch — full text is the reference deliverable)

### 4.1 Frontmatter (closed-world; CEL/JSON-Schema checkable per R3)

```yaml
---
grammar: 1
type: rfc
id: RFC-NNN
title: "Move all financial-state jobs onto the transactional outbox"   # proposal-carrying
status: draft              # draft | review | accepted | implemented | rejected | superseded | abandoned
provenance: agent-drafted  # human | agent-drafted
author: "pi/research-agent"        # accountable drafter(s)
steward: "pete"            # REQUIRED, exactly one accountable party; only the steward may write abandoned
created: 2026-10-02
updated: 2026-10-02
review_by: 2026-10-16      # REQUIRED iff status in [review, accepted]; deliberation deadline / revisit date
decided_by: null           # human ratifier, REQUIRED iff accepted|implemented|rejected; MUST NOT equal author; MUST NOT be an agent
decided_at: null
resolution: null           # link/ref to the decision record (interview log, thread, commit)
scope: ["src/billing/**"]  # routing key (parity)
tags: [billing, jobs]
summary: "Propose moving financial-state jobs to the outbox; bans direct queue calls"  # ≤200 chars
supersedes: null
superseded_by: null        # REQUIRED iff superseded — same commit, both sides
spawns: []                 # candidate ADR/SPEC IDs at proposal; links as they land
disposition_requested_at: null  # formal ratification request filed → suspends staleness
---
```

Derived postures (CEL, never stored): `is_at_risk` (now > review_by − 14d) ·
`is_stale` (status ∈ [review, accepted] ∧ now > review_by ∧ ¬suspended) ·
`is_dormant` (draft ∧ silence) · `is_terminal` · `needs_steward`.
`is_stale` ⇒ advisory only; MUST NOT be quoted as binding until re-affirmed
(steward bumps review_by with a note).

### 4.2 Body (skeleton strict; register per section type — V8)

```markdown
# RFC-NNN: <title>

> **Reviewer brief** — Decision sought: … · Recommended outcome: … ·
> Change if accepted: … · Affected surface: … · Stakes: …

> **For agents:** this is a PROPOSAL. It is authority only while
> `status: accepted|implemented`, not superseded, and not stale — and only for
> what was decided, never for current behaviour. You MUST NOT implement from an
> unratified RFC. If your task conflicts with a ratified RFC, stop and name the
> conflict; propose supersession.

## Problem                     (narrative prose; what, why now, how bad)
## Goals and Non-Goals         (lists; non-goal = checkable exclusion)
## Proposal                    (narrative; shape-level: interface sketch/pseudocode/
                                toy example allowed; NO filenames, signatures, DDL,
                                migration sequencing, config dumps)
## Proposed Constraints        (EARS-2119 statements, stable IDs C1…; proposed force;
                                each names its spawn target: ADR | SPEC)
## Alternatives                (narrative, live contenders; Do Nothing mandatory;
                                honest costs; straw-man check applies)
## Risks and Drawbacks         (why NOT do this; top failure modes)
## Assumptions and Open Questions   ([ASSUMPTION]/[NEEDS CLARIFICATION]; ≤3 genuine
                                open questions, each with owner + decide-by;
                                unresolved NEEDS CLARIFICATION gates ratification)
## Deliberation                (reviewers, review_by, kind of review wanted;
                                during review: verdict records yes|yes_if(cond)|not_yet)
## Resolution                  (filled ONLY at ratification: decision, decided_by,
                                verdict-per-reviewer, blocking concerns dispositioned
                                (two axes: disposition ≠ commenter acceptance;
                                timeout ≠ agreement), rejected alternatives each get
                                "do not re-propose unless …", confirmed spawns)
## Evidence and Prior Art      (claim↔source traceability; research base; additive)
```

Length: declared budget field in the reference (~250 lines default; Proposal carries
40–60%; dense blocks exempt) [JUDGMENT defaults, no measured threshold — parity with
sibling caps].

## 5. Load path and harness implications

1. **docs-writer routing row** — RFC STOP-stub → `references/rfc.md` (the deliverable).
2. **Drafting**: agent drafts `status: draft` (or `review` when ready for owner);
   rubric pre-flight runs (Calçado's 5 categories); ends by requesting human review.
3. **Ratification**: human-only; `decided_by` ≠ author; Resolution section filled;
   spawns confirmed. Unresolved `[NEEDS CLARIFICATION]` blocks.
4. **Adjudication consumption**: spec-builder/fix-designer MAY quote a ratified,
   current RFC for decision provenance per §3; the authority enum gains no member
   today — quotes flow through the existing ADR/SPEC channels once transcribed; the
   RFC itself is cited as lineage. (akb replaces exec-profile AUTHORITY later — R1.)
5. **Staleness sweep**: derived postures computed per run; `is_stale` accepted RFCs
   surfaced to the steward; `is_dormant` drafts listed.
6. **Spawn transcription**: ratification follow-through; akb tracks `spawns`/
   `spawned_by` lineage; an accepted RFC with no transcription progress by review_by
   goes stale (§3).
7. **akb `rfc.yaml` template** — deferred until RFC-002 V3 lands (R4); every gate
   above is pure-function-of-document (P1-eligible).
   **AMENDED 2026-10-02 (owner override):** a minimal interim V2 template shipped
   instead — RFC-002 and its spawns needed a typed home now (see `LOG.md` R4
   amendment).

## 6. EARS-2119 determination (the instruction's named question)

**Adopt, scoped.** Reasoning:

- **For**: sibling parity — one shared grammar (SKILL.md layer) cannot drift between
  three contracts; constraint statements become machine-lintable (keyword form, EARS
  shape, tether — all P1-eligible CEL/regex per the spec run's akb feasibility check);
  a ratified RFC is quotable authority (§3), and quotes need stable, checkable shape;
  the tether rule addresses measured negation fragility; field convergence on the
  RFC 2119 family (probe-keywords Q3).
- **Against**: an RFC's obligations are *proposed*, and most RFC content is argument,
  not constraint; mandating EARS shapes everywhere would push authors into premature
  spec-writing (V9 tension) and violate IESG 2025's "keywords are optional, not
  everywhere".
- **Resolution**: EARS-2119 is mandatory **only** in `## Proposed Constraints` (and
  any constraint statements elsewhere), exactly the spec.md scoping ("keywords belong
  only in constraint statements"). Proposed-vs-binding force is carried by the status
  gate and the preamble, not by grammar absence — an agent that respects the status
  gate cannot mistake proposed constraints for current ones, and an agent that doesn't
  respect the gate wouldn't be stopped by weaker grammar either. The exemplars' zero-
  keyword style is rejected: it made normative force unparseable (recon-local Q1f).

## 7. Known limitations (for the final report)

- **No measured study of proposal-format comprehension by agents exists** (confirmed
  negative, lane-agent-era F38). Every format choice rests on reliability, reviewability,
  and slop-resistance arguments + transferred evidence. Ship as v1-with-revision-triggers.
- **Freeze-on-ratify is stricter than every institutional precedent** (lane-lifecycle
  contradiction #1); nearest support is AIDR (days-old, single-author) and Rust's prose
  rule. The escape-hatch precedent (one designated live artifact) is answered by the
  Deliberation section being append-only post-ratification rather than the body being
  editable.
- **Calçado's exact rubric is unfetchable** (auth-gated Google Doc); only his five
  check categories are citable.
- **The DoC is scoped down by judgment** (V6): the full per-comment machine record is
  designed (lane-deliberation's recommended schema) but not mandated; revisit when
  multi-reviewer review arrives.
- **arXiv:2608.21747** (format barely matters on frontier models) cuts both ways: the
  machine contract is cheap insurance, the prose register is for the human ratifier.
- **AIDR overlap**: our freeze/supersession semantics duplicate shipped prior art; the
  reference should cite AIDR and differentiate (proposal lifecycle vs decision record).

## 8. Decision forks for the owner (interview)

- F1: EARS-2119 adoption (recommend: adopt, scoped — §6). Instruction-named question.
- F2: Per-section-type prose rule (V8) — diverges from sibling bullets+fragments.
- F3: Status enum + derived-posture lifecycle (V2/V4) — includes `rejected`, the
  anti-graveyard guard, and "clock never writes a status".
- F4: DoC scope (V6) — lightweight sections now vs full per-comment schema.
- F5: R1 resolution (§3) — ratified RFC as status-gated authority for decision
  provenance + untranscribed constraints; never for behaviour.
