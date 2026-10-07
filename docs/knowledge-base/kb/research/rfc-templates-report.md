---
as_of: "2026-10-02"
created: "2026-10-02"
grammar: 1
informs:
- RFC-003
is_draft: false
provenance: agent-drafted
run: rfc-templates
scope:
- docs/knowledge-base/**
status: final
summary: "Findings report: the RFC authoring contract for agent-consumed design proposals — closed-world frontmatter, human-only ratification, derived staleness postures, prose argument."
tags:
- rfc-format
- agent-consumption
- design-proposals
- research-report
title: RFC format for agent-consumed design proposals — research report (rfc-templates)
type: research
updated: "2026-10-02"
---
# RESEARCH REPORT — RFC format for agent-consumed design proposals

**Run:** `rfc-templates` · **Date:** 2026-10-02 · **Mode:** start (completed)
**Method:** coordinator + 2 recon lanes (one local, one web) + 4 deep lanes (web,
primary sources) + lighter independent audit (18 claims). Full evidence:
`recon-*.md`, `lane-*.md`, `audit-report.md`, working log `LOG.md`, design analysis
`SYNTHESIS.md` in this directory. Prior-art runs consumed directly per owner
override: `../adr-templates/`, `../spec-templates/` (REPORT/SYNTHESIS, lane-classic
×2, probe-keywords, both shipped references, docs-writer SKILL.md).

**Deliverables produced:**
- `global-harness/skills/docs-writer/references/rfc.md` — the RFC authoring
  contract, live in all projects.
- `global-harness/skills/docs-writer/SKILL.md` — RFC routing row activated
  (the last STOP-stub; the router is now complete: ADR, SPEC, RFC).

---

## 1. Question

What structure should RFC (design proposal) documents take when their primary
consumer is an AI agent in an agent-driven SDLC — as the deliberation-stage
process unit that, when ratified, spawns the ADRs/SPECs that bind — optimised
format-first, with ratification machine-checkable and stall detection mechanical?

## 2. Answer in one paragraph

A proposal lifecycle, not a decision record: closed-world frontmatter as the machine
contract (7-status enum `draft|review|accepted|implemented|rejected|superseded|
abandoned`, required `steward`, `review_by` deliberation deadline, human-only
ratification with `decided_by` ≠ `author`), staleness as **derived CEL postures that
cannot rot** (`is_stale` ⇒ advisory; the clock never writes a status; `abandoned` is
a steward-only write), a body whose register varies by section type (machine contract
for metadata, lists for enumerations, **narrative prose for the argument** because
bullets-pretending-to-be-reasoning is the documented slop failure, **EARS-2119 scoped
to constraint statements**), lightweight deliberation/resolution machinery (verdict
enum `yes|yes_if(cond)|not_yet`, two-axis dispositions where a timeout is never
recorded as agreement), mandatory Do-Nothing alternative, implementation-detail
exclusion with a positive shape allowance, and a load path that makes a ratified RFC
status-gated authority for **decision provenance only** — never for current
behaviour, which stays with the spawned ADR/SPEC (`adr > spec > plan` unchanged).

## 3. Load-bearing findings (anchors in lane files)

### Local recon (recon-local)
1. **The owner's two exemplar RFCs (akb RFC-001/002) are a pre-docs-writer third
   format**: prose metadata table, free-text status, zero RFC-2119/EARS usage,
   D/P decision registers, and an "ADR/tech-spec candidates" handoff section — the
   RFC names downstream ADRs/SPECs rather than containing them. Nothing in the
   harness consumes an RFC mechanically (authority enum `["adr","spec","plan","none"]`,
   `spec-builder.md:13`).

### Deliberation & ratification (lane-deliberation)
2. **No proposal-scoped machine-readable disposition-of-comments format exists**
   (falsification verdict: PARTIAL — five agent-era disposition records exist for
   findings/decisions, none for proposal comments; the gap is the combination).
3. **CSSWG's two-axis model is the design steal**: disposition ≠ commenter
   acceptance, and `Commenter Timed Out (Assumed Satisfied)` must never read as
   agreement. IESG: DISCUSS blocking, cleared only by its holder.
4. **Ratified-for-decision is near-universal; ratified-for-behaviour is
   contradicted** (PEP: "historical document rather than a living specification";
   Oxide's consensus state is the editable one). Approver ≠ author is checkable
   (KEP, verbatim).
5. **AIDR is shipped prior art** for freeze-at-ratification + human-only arbiter;
   cited and differentiated (proposal lifecycle vs decision record).

### Lifecycle & staleness (lane-lifecycle)
6. **Hard clock-driven expiry failed at IETF** ("illusory… wasted effort";
   abolition drafts; procon-2026bis — still a draft — softens to a marking).
   Hand-maintained status rots (KEP 617's kep.yaml has read `status: provisional`
   since 2018-09-08; that its substance shipped is corroborated inference) — so
   staleness is derived, never stored.
7. **Peer-reviewed expiry-pressure split**: targeted reminders to an accountable
   actor work (Nudge −60% resolution time; NudgeBot −6.8%, guardrails unchanged);
   blanket auto-close harms (−24% merges, −14% contributors over a year).
8. **Our freeze-on-ratify is stricter than every institutional precedent**
   (recorded honestly; nearest support: AIDR, Rust's prose rule).

### Agent-era format moves (lane-agent-era)
9. **Convergent and adopted**: structured header (only move with production
   evidence), rubric pre-flight (four independent implementations), evidence
   traceability, implementation-detail exclusion with positive shape allowance
   (pre-AI, 10 years), mandatory Do-Nothing + straw-man check.
10. **Confirmed negative: no measured study of proposal-format comprehension by
    agents exists.** The format sells falsifiability/governability/stall-detection,
    never quality.
11. **Prose-vs-bullets resolved per-section-type** (lemieux's explicit rule); the
    prose register protects the human ratifier, not the agent.

### Verification (lane-verify)
12. **kepval validates only the typed control file** (strict schema, 6 required
    keys, conditional PRR gating) — prose is never machine-checked; supports the
    frontmatter-as-machine-contract posture. **IETF publishes an I-D-stage
    required-content list** — a fixed section skeleton has precedent. Explicit
    computed dates, never "six months".

## 4. Audit outcome

Lighter audit (18 load-bearing claims, evidence-auditor, every primary fetched
fresh): **18 supported, 0 supported-with-correction, 0 killed** (`audit-report.md`).
One primary unfetchable (NudgeBot preprint PDF timed out) — verified against the
official ESEC/FSE 2022 program page instead. Residual strength-level caveats folded
in: KEP 617's "substance shipped" is corroborated inference (finding 6); marzouk /
Calçado / lemieux / create-rfc are practitioner or single-operator sources used as
practice evidence, not measurement; the procon-2026bis expiry softening is draft
text (-11), not a published RFC.

## 5. Ratified decisions (owner)

- **Requirements (interview 1):** R1 authority-role deferred to synthesis (with the
  owner's intent: ratified + current RFC = authoritative source for fix-designer
  auto-resolution; akb provides lineage) · R2 freeze at ratification + deliberation
  expiry · R3 4 lanes + lighter audit (akb validates: JSON Schema static + CEL
  time-based) · R4 prose reference + reports only (akb `rfc.yaml` deferred until
  RFC-002 V3 lands).
  - **AMENDED 2026-10-02 (owner override):** R4's deferral was overridden — RFC-002
    and its spawns needed a typed home now, so a minimal interim V2 `rfc` template
    shipped (the V2 `spec` template already occupied the same position); the
    no-migration clause was overridden with it (RFC-001/002 are being adopted into the
    KB). Full record: `LOG.md` R4 amendment.
- **Format forks (interview 2, all as recommended):** F1 EARS-2119 adopted, scoped
  to constraint sections (the instruction's named question) · F2 per-section-type
  prose rule (deliberate divergence from sibling bullets+fragments) · F3 lifecycle
  as synthesized (7 statuses, derived postures, clock never writes status) ·
  F4 lightweight DoC sections (full per-comment schema parked) · F5 ratified RFC =
  status-gated authority for decision provenance + untranscribed constraints, never
  for behaviour.
- Design verdicts V1–V20 + traceability: `SYNTHESIS.md` §2/§4.

## 6. Harness-change implications (ranked by dependency)

1. **docs-writer SKILL.md RFC row** — DONE this run (router complete).
2. **akb `rfc.yaml` adopter template** (JSON Schema + CEL + pass/fail mockups) —
   deferred until RFC-002 V3 lands (R4). Every gate in the reference is
   pure-function-of-document (P1-eligible): closed-world frontmatter, status enum,
   `decided_by` ≠ `author`, `review_by` requiredness, derived postures (time CEL),
   supersede symmetry, EARS-2119 lint.
   **AMENDED 2026-10-02 (owner override):** an interim V2 template shipped instead;
   the V3 adopter template remains the target, and the interim header lists the
   deferred gates (see `LOG.md` R4 amendment).
3. **Adjudication consumption of ratified RFCs** — today quotes flow through the
   ADR/SPEC channels post-transcription, with the RFC cited as lineage (no enum
   change). When akb replaces `exec-profile.md` AUTHORITY (owner R1 note), the
   RFC joins the lineage graph via `spawns`/`spawned_by`.
4. **Staleness sweep** — the §4 derived postures are the sweep's RFC-side checks;
   no harness runner exists yet (same gap as the SPEC run's external checker,
   parked there).
5. **Convention detection** — akb's `kb/rfc/` exemplars are the detection target;
   they predate the format and stay as research record (no migration, R4).
   **AMENDED 2026-10-02 (owner override):** the no-migration clause was overridden —
   RFC-001 (as `superseded`) and RFC-002 (as `draft`) are being adopted into the KB
   under the interim `rfc` template (see `LOG.md` R4 amendment).

## 7. Parked / deferred items (ranked)

1. **Full per-comment machine-readable DoC record** (schema drafted in
   `lane-deliberation.md` §DoC-schema-candidates) — mandate when multi-reviewer
   review arrives; today's single-ratifier pipeline uses the lightweight sections.
2. **akb `rfc.yaml` template authoring** (implication #2, post-RFC-002 V3).
   **Partially delivered 2026-10-02 (owner override):** the interim V2 template
   shipped; the V3 adopter-template work remains open (see `LOG.md` R4 amendment).
3. **External staleness-checker implementation** (shared with spec-templates
   parked item #3 — one checker should serve both formats' postures).
4. **Calçado's exact rubric** (auth-gated Google Doc) — obtain by manual/
   authenticated route; only his five check categories are citable today.
5. **In-house format-comprehension experiment** (lane-agent-era next-step #2):
   vary RFC shape at constant content, measure agent comprehension on a fixed
   question set — nothing in the literature does this.
6. **AIDR/AgDR interop watch** — if either gains adoption, declare conformance or
   a delta table (lane-deliberation next-step #1).
7. **IETF immutability/errata gap** (lane-deliberation missing evidence) — cheap
   follow-up fetch (RFC 6410, errata policy) if the freeze rule is ever challenged.

## 8. Limitations

- No measured study of proposal-format comprehension by agents exists (confirmed
  negative); every format-level claim is design inference from transferred +
  adjacent evidence. The contract ships v1-with-revision-triggers and carries its
  self-claim boundary (falsifiability/governability/stall-detection only).
- Freeze-on-ratify is stricter than every institutional precedent (lane-lifecycle
  contradiction #1); AIDR support is days-old, single-author.
- The DoC scope-down (lightweight sections vs full record) is a judgment call for a
  single-ratifier harness.
- Marzouk's extractor claims are asserted, not measured; his "39%" statistic is a
  mis-cite of arXiv:2505.06120 (multi-turn degradation) and was NOT propagated.
- The untranscribed-constraints-quotable rule (F5) is a judgment mapping onto
  evidence-backed components.
- arXiv:2608.21747 (format barely matters on frontier models) bounds the value of
  register choices for agent readers; the prose register is for the human ratifier.
