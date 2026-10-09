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
summary: "Run log for the rfc-templates research: question, method, owner-ratified forks F1–F5, lane activity, and the decision trail behind the RFC authoring contract."
tags:
- research-log
- rfc-format
- rfc-templates
- process
title: Research run log — RFC format for agents (rfc-templates)
type: research
updated: "2026-10-09T16:35:15Z"
---
# LOG — rfc-templates research run

**Question:** What structure should RFC (design proposal) documents take when their
primary consumer is an AI agent in an agent-driven SDLC — filling the docs-writer
router's remaining STOP-stub (`references/rfc.md`)?

**Date:** started 2026-10-02 · **Mode:** start · **Status: COMPLETED 2026-10-02**
(audit 18/0/0; deliverables live)

## Run parameters

- SUBJECT: `rfc-templates`
- INSTRUCTIONS (owner, verbatim intent): consume prior-art runs directly
  (user override on coordinator-only rule, limited to listed files + cross-cutting
  research they identify): adr-templates & spec-templates REPORT/SYNTHESIS, the two
  resulting references (`references/adr.md`, `references/spec.md`), the docs-writer
  SKILL.md, and decision-forms skill only if owner input required. Research the same
  requirements for RFCs; do NOT repeat already-performed research — gap-fill only for
  RFC-specific requirements. Both prior doc types unified on hybrid EARS-2119;
  determine whether it should be adopted for RFC agent/LLM comprehension too.
  Reference files must be maximum LLM comprehension at minimum token cost.

## Prior art consumed directly (user override)

- `.pi/research/adr-templates/{REPORT,SYNTHESIS}.md` — ADR format: rationale zone +
  machine-payload zone, prohibition-first, frontmatter-as-retrieval-API, staleness
  machinery part of format. 14 claims audited (11 supported, 3 corrected, 0 killed).
- `.pi/research/spec-templates/{REPORT,SYNTHESIS}.md` — SPEC format: frozen
  change-scoped process unit, `verified` staleness block + anchors + drift ledger,
  EARS-2119 grammar, load-path-as-format-content (V16), two-space promotion
  (change|capability). 20 claims audited (15 supported, 5 corrected, 0 killed).
- `global-harness/skills/docs-writer/SKILL.md` — router + shared EARS-2119 grammar;
  RFC row is the last STOP-stub.
- `global-harness/skills/docs-writer/references/{adr,spec}.md` — the two shipped
  authoring contracts (sibling parity target for rfc.md).
- `.pi/research/spec-templates/lane-classic.md` (cross-cutting) — Rust RFC template
  (dual-register guide/reference + "return to examples" clause; PRs #2059/#3982),
  PEP 1 (15 parts, status model, rejection bars, "historical document not living
  specification"), Gerrit design docs (split per-alternative files, 10-day conclusion
  hold), Google design docs (Ubl), Amazon PR/FAQ. Conventions 1–12 + rejects 1–9.
- `.pi/research/adr-templates/lane-classic.md` (cross-cutting) — keep/drop/change
  table; absence-invisibility (tianpan); over-adherence (Davidson); bit-rot (Defendi).
- `.pi/research/spec-templates/probe-keywords.md` (cross-cutting) — EARS-2119 evidence:
  no keyword-choice→compliance measurement exists (verified negative); layers-not-
  competitors; negation fragility (SFF/DKB) → tether rule; field converges on RFC 2119
  family; RFC 8174 all-caps-only.

## Trivial local check (coordinator, exempt)

- akb repo = `~/code/projects/tools/pi/agent-memory/agent-kb/`; owner's in-use RFC
  exemplars: `docs/kb/rfc/RFC-002-open-validation-engine.md` (ratified 2026-09-26/27,
  supersedes RFC-001), referenced heavily by `agent-memory/DESIGN.md` (D-register +
  P-register). These are the local prior-art RFC instances.
- docs-writer SKILL.md routing table: RFC row = STOP-stub (only remaining one).

## Decisions made

- **R1 (owner, interview 2026-10-02, q1): DEFER the authority-role fork to synthesis**, with
  owner's load-bearing note verbatim: intention is that a **ratified + current (not
  superseded) RFC IS an additional authoritative source** in the hierarchy, consumed by
  the high-reasoning fix-designer for auto-resolution of scope-creep adjudication; akb
  provides provenance/lineage (tags, wikilinks) so the designer can trace which RFC an
  ADR/SPEC spawned from; once akb is operational it replaces exec-profile.md's AUTHORITY
  section. ⇒ The format MUST make ratification+currentness machine-checkable (status
  enum + gates), and synthesis must decide the authority-chain placement (status-gated
  member vs pre-authority). Lane briefs carry both branches.
- **R2 (owner, q2): freeze at ratification + deliberation expiry.** While in discussion:
  machine-checkable deliberation metadata (review-by/expiry, comment disposition); at
  ratification the RFC freezes; binding content moves into spawned ADR/SPEC.
- **R3 (owner, q3): 4 deep lanes + lighter audit.** Owner note: akb does validation —
  CEL today; RFC-002 expands to JSON Schema (static doc) + CEL (maths, e.g. time-based
  validation à la Kubernetes). ⇒ time-based deliberation expiry must be CEL-expressible.
- **R4 (owner, q4): deliverables = `references/rfc.md` + REPORT.md + SYNTHESIS.md only.**
  akb `rfc.yaml` deferred until RFC-002 V3 lands; no RFC-001/002 migration analysis.
  - **AMENDED 2026-10-02 (owner override).** R4 was overridden in a follow-on session,
    same day. Rationale: RFC-002's intended spawns (ADRs/SPECs) and RFC-002 itself
    needed a typed home in the KB *immediately*, and exploration/implementation could
    not wait for the V3 engine. Result: a minimal interim V2 `rfc` template was
    authored, validated (pass/fail mockups) and committed, joining the already-existing
    V2 `spec` template; its header documents exactly what is deferred to V3. The
    no-migration clause was overridden with it: RFC-001 (as `superseded`) and RFC-002
    (as `draft`) are being adopted into the KB under the interim template. Supporting
    precedent: `adr-templates/recon-local.md` finding 6 explicitly permits authoring
    now if rename-tolerant, and the base's `spec` template already overrode the
    equivalent parking in `spec-templates/REPORT.md` §7 item 2. The original R4 text
    above stands; this is an override record, not a rewrite.
- **R5 (coordinator):** recon anomaly reported to owner: RFC exemplars live in
  `~/code/projects/tools/llm-wiki/agent-kb/kb/rfc/`, NOT in git; DESIGN.md-cited path
  unresolvable. Owner informed; no action requested.

## Key findings

### recon-local (landed; output at session artifact path, to be copied into run dir)

- **Exemplar path correction:** RFCs live at `~/code/projects/tools/llm-wiki/agent-kb/kb/rfc/`
  (RFC-001 402 lines, RFC-002 806 lines), **not in git**, no template file. The
  DESIGN.md-cited path `agent-kb/docs/kb/rfc/` does not exist on disk.
- **Exemplar format:** NO YAML frontmatter — 2-column markdown metadata table
  (`Status`, `Date`, `Authors`, `Supersedes`, `Research base`, `Downstream
  artifacts`, RFC-002 adds `Decision register`). Numbered sections + appendices.
  Status is free text: `Draft — awaiting owner ratification`, `SUPERSEDED by
  RFC-002 … retained as research record`. No enum, not machine-parseable.
- **No RFC 2119 keywords, no EARS shapes** in either exemplar (0 hits in constraint
  positions); normative force carried by bold prose + seam rules (P1–P5).
- **Decision recording is bespoke:** RFC-002 Appendix A = D-register (D1–D19,
  owner-ratified) + P-register (P-a…P-i, incl. OVERTURNED). Downstream handoff =
  "Open questions → ADR / tech-spec candidates" numbered sections (RFC-001 §15: 8
  items; RFC-002 §18: 19 items) — the RFC names ADRs/SPECs, does not contain them.
- **Lifecycle observed:** Draft → ratified-in-direction (owner, via interview,
  recorded in research LOG) → Superseded (header-only edit; content retained as
  research record). Human-flipped, prose-recorded.
- **Harness consumption today: NONE machine-readable.** Authority enum is
  `["adr","spec","plan","none"]` (spec-builder.md:13, fix-designer.md) — no `rfc`
  member; an RFC cannot be quoted as authority by any consumer. Gate pattern:
  ADR binding iff `status: accepted`; SPEC binding iff `active` + `verified.state:
  live` (spec-builder.md:56–58). RFCs are *pre-authority*: read by owner +
  docs-writer when authoring the ADR/SPEC the RFC names; akb DESIGN.md:381 consumes
  RFC-002 as prose authority ("ratified as akb's direction").
- **docs-writer family already live in akb** but in a different tree:
  `agent-kb/agent-kb/docs/knowledge-base/kb/decisions/ADR-001/002`,
  `specs/SPEC-001` carry the exact contract frontmatter + skeletons; akb templates
  `adr.yaml`/`spec.yaml` exist; **no `rfc.yaml`**. RFC exemplars predate the
  format family (09-26/27 vs 10-01) — they are a pre-docs-writer third format.
- Core open design question (recon): reuse ADR/SPEC closed-world frontmatter +
  EARS-2119, or stay prose-heavy? No existing artifact answers it.

## Ratifications (owner interview #2, 2026-10-02 — format forks)

- **F1: EARS-2119 ADOPTED, scoped to constraint sections** (`## Proposed Constraints` +
  any constraint statement elsewhere); proposed-vs-binding force carried by status gate,
  not grammar absence. (The instruction's named question — determined, ratified.)
- **F2: per-section-type prose rule** — frontmatter/tables = machine contract,
  enumerations = lists, Problem/Proposal/Alternatives = narrative prose with budget,
  constraints = EARS-2119. Deliberate divergence from sibling bullets+fragments.
- **F3: lifecycle as synthesized** — 7 statuses (draft|review|accepted|implemented|
  rejected|superseded|abandoned), steward required, review_by required in
  [review,accepted], CEL-derived postures, clock never writes status, abandoned =
  steward-only, human-only ratification with decided_by ≠ author.
- **F4: DoC = lightweight sections** (`## Deliberation` + `## Resolution`, verdict enum
  yes|yes_if(cond)|not_yet, two-axis dispositions, timeout ≠ agreement); full per-comment
  YAML schema parked (SYNTHESIS §7).
- **F5: R1 resolved as synthesized** — ratified RFC = status-gated authority for decision
  provenance + untranscribed constraints; NEVER for current behaviour; is_stale ⇒
  advisory; precedence adr > spec > plan unchanged, RFC behind the chain as provenance.

## Open questions

- What mechanical role does an RFC serve in this harness? (pre-decision deliberation
  artifact whose ratification spawns SPEC/ADR? harness/process proposal like RFC-002?
  both?)
- Does an accepted RFC remain quotable authority, freeze as history, or promote?
- EARS-2119 for RFC constraint statements? (Prior art: both siblings adopted it; RFCs
  are *proposed* constraints, not binding ones — register question is RFC-specific.)
- Scope: does the run also cover akb's `rfc` adopter-template shape, or prose-first
  only (as spec run did: prose now, akb template post-RFC-002-landing)?

## Parked/deferred items

(carried from spec-templates REPORT §7, RFC-relevant): per-plan authority relevance
resolver; akb template authoring post-RFC-002 V3; `kb.*` cross-page invariants.

## Deliverables (drafted 2026-10-02, pre-audit)

- `global-harness/skills/docs-writer/references/rfc.md` — RFC authoring contract
  (V1–V20 encoded; forks F1–F5 as ratified). LIVE.
- `global-harness/skills/docs-writer/SKILL.md` — RFC routing row activated; router
  now complete (ADR/SPEC/RFC). LIVE.
- `.pi/research/rfc-templates/REPORT.md` — final report (audit section pending).
- `.pi/research/rfc-templates/SYNTHESIS.md` — coordinator analysis.

## Child-run index

- audit (evidence-auditor, async, run 30c253c7) → `audit-report.md` — LANDED:
  **18 supported, 0 corrected, 0 killed.** NudgeBot verified via ESEC/FSE program
  page (preprint PDF timed out). Caveats folded into REPORT §4: KEP-617-substance =
  corroborated inference; practitioner sources = practice evidence not measurement;
  procon-2026bis = draft text. (Coordinator had passed `output: "string"` by mistake;
  retrieved from artifact path and renamed.)

RUN COMPLETE. Remaining follow-ups live in REPORT §7 (parked, ranked). Behavioural
verification of the skill change requires a NEW Pi session (config activates on
session start).

- recon-local (scout, async, run 7811c326) → `recon-local.md` — akb RFC-001/002
  exemplar structure, lifecycle vocabulary, harness RFC touchpoints, consumption points.
- recon-landscape (researcher, async, run 92100e49) → `recon-landscape.md` — LIGHT web
  gap-fill: IETF proposal machinery, Oxide RFD/KEP/HashiCorp templates, deliberation
  machinery (FCP/comment disposition), agent-era proposal-format writing,
  proposal-vs-record distinction. LANDED; key: design space splits on lifecycle not
  sections; no agent-readable DoC format found (light-pass, to be falsified); agent-era
  convergent moves = structured header / rubric pre-flight / evidence traceability /
  implementation-detail exclusion.
- lane-deliberation (researcher, async, run 5f19686b) → `lane-deliberation.md` — LANDED.
  FALSIFICATION VERDICT: PARTIAL — no institutional machine-readable DoC (W3C DisCo = 4
  plaintext keys; CSSWG issuegen = regex-parsed lines; IETF defines NONE, substitutes an
  appeal clock), and none proposal-scoped — BUT five agent-era machine-readable
  disposition records exist for findings/decisions (fro.bot synthesis ledger —
  Zod-validated, totality-enforced, introduced after measured 7-artifacts-7-shapes
  failure; tessl review-retrospective; design-court candidate-vs-judged split; kla
  approval-event schema with terminal `expired`; **AIDR — closest prior art: ratified ⇒
  append-only + supersede-never-rewrite + human-only arbiter field**, days-old,
  single-author). Gap = the combination (proposal-scoped × comment-granular ×
  authority-bearing × commenter-state-bearing × validator-enforced). KEY MODEL: CSSWG
  two axes — disposition (Closed) SEPARATE from commenter acceptance (Verified), incl.
  `Commenter Timed Out (Assumed Satisfied)` — a timeout must never read as agreement.
  RATIFICATION SEMANTICS verdict: ratified-for-decision YES (KEP implementable / PEP
  Accepted / Rust merged / IETF approval); ratified-for-BEHAVIOUR CONTRADICTED (PEP:
  "historical document rather than a living specification", authority moves elsewhere;
  Oxide: editable-consensus paradox — `published` is the editable consensus state,
  `committed` only post-implementation). Approver ≠ author is checkable (KEP verbatim).
  orfc expresses disposition by DELETING the comment — the exact property not to
  inherit. Verdict enum convergence: Squarespace yes/yes_if/not_yet ≈ rfc-critic
  Approve/Approve-with-changes/Needs-redesign. Recommended minimal DoC record drafted
  in-lane (schema_version, revision_digest, deliberation window, per-comment disposition
  axes, unresolved_blocking gate).
- lane-lifecycle (researcher, async, run 18bf74c3) → `lane-lifecycle.md` — LANDED.
  Key: hard clock expiry FAILED at IETF ("illusory… wasted effort", no-op refreshes,
  2026bis softens to marking); KEP 617 `provisional` since 2018 = stored status rots,
  derived posture cannot; peer-reviewed expiry-pressure split: targeted reminders work
  (Nudge −60% resolution; NudgeBot −6.8%), blanket auto-close harms (−24% merges,
  −14% contributors). Recommended state machine (adapted in synthesis): 6 statuses,
  steward required, review_by required in [review, accepted], derived is_stale/
  is_dormant/is_at_risk via CEL, clock NEVER writes status, abandoned = steward-only,
  accepted requires steward+review_by (anti-graveyard). Contradiction recorded: our
  freeze-on-ratify is STRICTER than every precedent (all keep escape hatches; nearest
  = FEP replaces/replacedBy + one designated live artifact). "review_by", never
  "expires_at" (word-choice ambiguity source). philcalcado "expiry" = author-owned
  REVISIT, not countdown. FEP: notice-then-withdraw doesn't work (operator report).
- lane-agent-era (researcher, async, run 91019606) → `lane-agent-era.md` — LANDED.
  Convergent-moves verdicts: ADOPT structured header (Marzouk production extractor =
  only operational evidence), rubric pre-flight (4 independent impls; lemieux
  mechanics: category+severity+confidence≥80+quote-location+never-suggest-fixes),
  evidence traceability (claim↔source machine-checkable), implementation-detail
  exclusion WITH positive allowance (shape yes, artifacts no), anti-plausible-prose
  detection, mandatory do-nothing + straw-man check; ADAPT length budget (declared
  field + section shares, diagrams/contracts exempt), provenance (typed enum +
  accountability + no-agent-as-sole-arbiter). CONFIRMED NEGATIVE: no measured study
  of proposal-format comprehension by agents (arXiv:2608.21747 is nearest — format
  barely matters on frontier models, hugely on weak). SHARPEST CONTRADICTION:
  prose-vs-bullets cuts against sibling ADR/SPEC bullets+fragments — resolution
  supported by evidence = per-section-type (table/frontmatter=machine contract,
  prose=claims/rationale/alternatives, lists=enumerations; lemieux explicit). Other
  open contradictions: Open-Questions section banned vs required; one decided design
  vs options matrix; time estimates banned vs required. Calçado's exact rubric =
  unfetchable Google Doc (only 5 check categories citable). Marzouk's "39% accuracy"
  stat is a mis-cite of arXiv:2505.06120 (multi-turn degradation) — do not propagate.
- lane-verify (researcher, async, run 86228f16) → `lane-verify.md` — LANDED. Verdicts:
  (1) CORRECTED: IETF DOES publish an I-D-stage required-content list
  (authors.ietf.org: IPR/Abstract/Intro/Security/IANA/References/Authors + header
  fields) — fixed section skeleton has IETF precedent; (2) CORRECTED: kepval = strict
  schema + 6 required keys + conditional PRR approver gating ONLY — prose body and
  questionnaire answers are convention/advisory (machine checking covers the typed
  control file, never prose); (3) CONFIRMED with wording correction: RFC 7322 §4.3
  abstract self-containment + §4.8.6.4 I-Ds informative-only + Work in Progress;
  (4) CONFIRMED: 185 days is the operative rule, "six months" frozen boilerplate —
  use explicit computed dates, never "six months" phrasing.

NEXT: requirements AGREED (interview 2026-10-02) → full fanout launched (4 deep lanes,
async) → synthesis → decision fork(s) per decision-forms (EARS-2119 recommendation +
authority-role resolution) → lighter audit → deliverables.

## Pages

Reports filed from this run (ADR-007):

- [[rfc-templates-report]] — findings report: the RFC authoring contract (closed-world frontmatter, human-only ratification, derived staleness postures).
- [[rfc-templates-synthesis]] — coordinator's design analysis of the RFC as deliberation-stage process unit; ratified forks F1–F5.
- [[rfc-templates-recon-landscape]] — web recon: published RFC/proposal templates mapped by revision lifecycle.
- [[rfc-templates-recon-local]] — local recon: the owner's akb RFC-001/002 exemplars and the RFC doc type's missing committed home.
- [[rfc-templates-lane-agent-era]] — deep lane: the 2026 agent-era RFC literature.
- [[rfc-templates-lane-deliberation]] — deep lane: deliberation, comment disposition, and the human-only ratification gate.
- [[rfc-templates-lane-lifecycle]] — deep lane: proposal lifecycle, IETF's 185-day expiry, derived staleness postures.
- [[rfc-templates-lane-verify]] — verification lane: four IETF and KEP/kepval items re-checked against primaries.
- [[rfc-templates-audit-report]] — independent audit of this run: 18 claims verified, 18 supported, 0 corrected, 0 killed.
