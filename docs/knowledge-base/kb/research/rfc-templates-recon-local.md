---
as_of: "2026-10-01"
created: "2026-10-01"
grammar: 1
informs:
- RFC-003
provenance: agent-drafted
run: rfc-templates
scope:
- "docs/knowledge-base/**"
status: final
summary: "Local recon: the owner's akb RFC-001/002 exemplars, the RFC doc type's missing committed home, and how the harness consumes — or fails to consume — RFCs."
tags:
- recon
- rfc-format
- harness
title: "Recon local — RFC document formats: akb exemplars and harness consumption (rfc-templates)"
type: research
updated: "2026-10-01"
---
# recon-local — RFC document formats (akb exemplars + harness consumption)

READ-ONLY recon. No file edited except this output.

## 0. PATH CORRECTION (load-bearing — read first)

The task-stated path does **not** exist:
- `~/code/projects/tools/pi/agent-memory/agent-kb/docs/kb/rfc/` → `No such file or directory`.
  `~/code/projects/tools/pi/agent-memory/agent-kb/` contains exactly one file: `VALIDATION-2026-09-07.md`.

Actual exemplar location (found by filesystem search):
- **`/home/pete/code/projects/tools/llm-wiki/agent-kb/kb/rfc/`** — 2 files, plain directory.
- This dir is **not** in any git repo (`git rev-parse` from it: "not a git repository"; no `.agent-kb`/`akb.yaml` marker).
- The akb Go repo is `…/llm-wiki/agent-kb/agent-kb/` (has `cmd/akb`, `internal/…`); it contains **zero** `RFC-*` files (`find . -name 'RFC-*'` → empty; `git log --all -- '**/RFC-*'` → empty). Its `docs/` holds only `knowledge-base/` and `wikilinks.md`; `docs/kb` does not exist.
- `agent-memory/DESIGN.md:567` and `:381` cite `agent-kb` repo `docs/kb/rfc/RFC-002-open-validation-engine.md` — that cited path also does not resolve on disk. **Gap: the RFC doc type has no committed home in the akb repo.** Confidence: HIGH (direct `ls`/`find`/`git`).

---

## Q1 — Exemplar structure

Files (all of them): `RFC-001-json-schema-cel-coexistence.md` (**402 lines**), `RFC-002-open-validation-engine.md` (**806 lines**). No `0000-template`, no `RFC-000`, no third exemplar. Confidence: HIGH.

**(a) Metadata header — NO YAML frontmatter.** Both files open with H1 on line 1; metadata is a 2-column markdown table with bold labels and no field-name closure of any kind. Field names verbatim:
- RFC-001 (`lines 3–9`): `Status`, `Date`, `Authors`, `Research base`, `Supersedes / amends`, `Downstream artifacts`
- RFC-002 (`lines 3–9`): `Status`, `Date`, `Authors`, `Supersedes`, `Decision register`, `Research base`, `Downstream artifacts`

**(b) Ordered headings** (H1 has line 1 both files).
RFC-001:
```
14 ## 1. Abstract          25 ## 2. Motivation        46 ## 3. Goals / Non-goals
48 ### Goals               70 ### Non-goals          84 ## 4. Design principles
102 ## 5. The document model 104 ### 5.1 Two artifacts 116 ### 5.2 Derived views are akb-authored, output-only
127 ### 5.3 Pre-publication hazards…  140 ## 6. Template format  142 ### 6.1 Shape…
197 ### 6.2 Mockup obligations…       206 ## 7. Validation pipeline (write/append)
224 ## 8. JSON Schema coverage commitment (G1)  250 ### 8.1 Worked example…
260 ## 9. CEL coverage commitment (G2)  287 ## 10. Overlap guidance (G6) — the SKILL decision rule
305 ## 11. CLI surface       320 ## 12. Approve gate fix…   327 ## 13. Sweep design (lint)
336 ## 14. Roadmap interactions…  347 ## 15. Open questions → ADR / tech-spec candidates
368 ## 16. Phasing…          378 ## 17. Risks           389 ## Appendix A — Research base
395 ## Appendix B — Glossary
```
RFC-002 (same spine, longer):
```
15 ## 1. Abstract  40 ## 2. Motivation  61 ## 3. Relationship to RFC-001  77 ## 4. Design principles
99 ## 5. The document model (D12, D13, D7)  101 ### 5.1 Shape  131 ### 5.2 Two artifacts, one builder…  169 ### 5.3 Declared projections
189 ## 6. Template format (V3 — breaking…)     263 ## 7. Validation pipeline (write/append)
282 ## 8. The write path: input = output (D3, P6)  284 ### 8.1 What dies  299 ### 8.2 What remains, transformed  323 ### 8.3 Paths (D11)  333 ### 8.4 Defect repairs riding Phase 2 (no policy content)
357 ## 9. JSON Schema coverage (G1)  369 ## 10. CEL coverage (G2 + D8)  390 ## 11. Lint ownership (D5)
460 ## 12. index.md becomes a derived artifact (D4)  491 ## 13. The markdown parsing pivot (D6)  493 ### 13.1 Phase (a′)…  508 ### 13.2 Phase (a)…  518 ### 13.3 Vocabularies and dialect
534 ## 14. Config surface (D9)  574 ### 14.1 Search semantics repairs (D17)  591 ## 15. JSON I/O and agent integration (D12, D14, D15)
651 ## 16. What akb still prescribes (the honest, complete list)  682 ## 17. Phasing…  709 ## 18. ADR / tech-spec candidates  727 ## 19. Risks
739 ## Appendix A — Decision register (owner-ratified)  764 ### P-register — drafting positions, owner-ratified 2026-09-27
782 ## Appendix B — Carried from RFC-001 unchanged  800 ## Appendix C — Research base
```
Note the ordering signature: numbered `## N. <Section>` / `### N.M`, goal IDs (`G1`), principle IDs (`P1`), decision IDs (`D12`) baked into headings; **Appendices after the numbered sections**. Confidence: HIGH.

**(c) Status/lifecycle vocabulary** — free-text, one line each, quoted verbatim:
- RFC-001 `:5` — `| **Status** | **SUPERSEDED by RFC-002** (`RFC-002-open-validation-engine.md`, 2026-09-27) — retained as research record; its technical spine carries into RFC-002 (see RFC-002 §3 and Appendix B) |`
- RFC-001 `:9` — `| **Supersedes / amends** | `agent-memory/DESIGN.md` §4.2/§4.3/§5 (amends; JSON direction is new relative to that ratified design) |`
- RFC-002 `:5` — `| **Status** | Draft — awaiting owner ratification |`
- RFC-002 `:7` — `| **Supersedes** | RFC-001 (`RFC-001-json-schema-cel-coexistence.md`) — retained as research record |`
Vocabulary actually used: `Draft`, `awaiting owner ratification`, `SUPERSEDED by RFC-002`, `retained as research record`, `Supersedes`, `amends`; elsewhere `ratified in direction`, `owner-ratified` (RFC-002 `:739`, `:764`), `OVERTURNED` (RFC-002 `:779`). **No enum, no `status:` key, not machine-parseable.** Confidence: HIGH.

**(d) Totals:** RFC-001 = 402 lines; RFC-002 = 806 lines; dir = 2 files, 79 319 bytes.

**(e) Decision recording** — bespoke, two mechanisms:
- RFC-001 has **no register**; instead inline principles `P1–P5` (`:86–99`, e.g. `- **P1. The seam rule.** A constraint belongs to JSON Schema iff its verdict is a pure function…`) and forward-deferral at §15 (`:347` `## 15. Open questions → ADR / tech-spec candidates`, 8 numbered items).
- RFC-002 Appendix A (`:739`) is a **D-register**: ``| # | Decision | Ratified (2026-09) |`` with rows `D1`…`D19` + `D-seq`. Example row (`:743`): `| D1 | RFC disposition | New RFC-002 supersedes RFC-001 (26/27) |`.
- RFC-002 `:764` **P-register**: ``| # | Position | Outcome | Where |``, rows `P-a`…`P-i`. Example row (`:768`): `| P-a | Stripped-`relPath` convention kept (`file.path: "notes/foo.md"`) | **Ratified** | §5.1 |`; and (`:779`) `| P-i | No new per-invocation flags | **OVERTURNED** — flags allowed where they earn their keep (sandboxed agents may not read the KB config); precedence flag > akb.yaml > default | §14 |`.
- Downstream work is enumerated as an **"ADR / tech-spec candidates"** numbered list (RFC-001 `:347`, 8 items; RFC-002 `:709`, 19 items) — the RFC does NOT contain the ADRs; it names them. Confidence: HIGH.

**(f) RFC 2119 / EARS:** **absent.**
- `grep -E '\b(MUST|SHALL|SHOULD|MAY|REQUIRED|RECOMMENDED|OPTIONAL|MUST NOT|SHOULD NOT)\b'` → RFC-001: **0 hits**; RFC-002: 4 hits, **all inside YAML template comments**, not constraint statements: `:193 path: notes  # OPTIONAL resolution hint…`, `:196 state_field: is_draft  # OPTIONAL…`, `:221 functions: [word_count]  # OPTIONAL subset…`, `:222 projections: […]  # OPTIONAL declared derived views…`.
- EARS shapes: `grep -E '\b(WHEN|WHILE|WHERE|IF|THEN)\b'` → **0 hits in both files**.
- Normative force is carried by bold prose (e.g. RFC-002 `:19` `**The write path is byte-faithful: input = output.**`), the seam rule `P1`, and lowercase `must`/`should`/~13 occurrences each. Confidence: HIGH.

---

## Q2 — Lifecycle vocabulary

Observed state machine (from the two Status lines + LOG.md tail): `Draft` → owner ratifies **in direction** → `D1: NEW RFC-002 supersedes RFC-001` → prior RFC `SUPERSEDED … retained as research record`; a still-active RFC can additionally `amend` another design doc (`Supersedes / amends`). Drafting positions taken *under* the RFC can be `Ratified` or `OVERTURNED` in the P-register.

Transition log (narrative, not schema): `…/llm-wiki/agent-kb/agent-kb/.pi/subagents/proposals/json-cel/research/LOG.md`
- `:188` — `RFC-001 in Draft, awaiting owner ratification. Next: owner review → ADRs per RFC §15.`
- `:320` — `### Owner decisions (interview, 2026-09-27) — RFC-002 RATIFIED IN DIRECTION`
- `:322` — `- **D1: NEW RFC-002, supersedes RFC-001** (RFC-001 stays as research record).`
- `:441` — `**…RFC-002-open-validation-engine.md written in full.**`
- `:445` — `- **RFC-001 marked SUPERSEDED** (header only; content untouched as research record).`
- `:449` — `RFC-002 in Draft, awaiting owner section-by-section ratification. Next: owner review → ADRs per RFC-002 §18.`

So the RFC lifecycle is **two-state-to-three-state, human-flipped, prose-recorded**: Draft → (ratified-in-direction) → Superseded. There is no `proposed`/`accepted`/`active` vocabulary and no `provenance`/`ratified_by` field. Confidence: HIGH for vocabulary; HIGH for transitions.

---

## Q3 — Template / authoring convention

- **No RFC template file exists anywhere.**
  - `find ~/code/projects/tools -iname '*0000*'` → no RFC-template match (hits are unrelated zed/opencode migrations).
  - akb's dogfood KB templates are `agent-kb/agent-kb/docs/knowledge-base/.agent-kb/templates/`: `adr.yaml`, `adr_pass.md`, `adr_fail.md`, `spec.yaml`, `spec_pass.md`, `spec_fail.md` — **no `rfc.yaml`**.
  - No `rfc` string in akb Go source / skills / templates; only in `.pi/subagents/proposals/json-cel/research/*.md` (LOG.md, cel-bridge-research-b.md, openness-config-surface.md, SYNTHESIS-openness.md).
- **The harness declares the RFC reference explicitly unauthored** — `global-harness/skills/docs-writer/SKILL.md`:
  - `:3` (frontmatter description): `… Will also route RFC authoring once that reference lands — ask before improvising it.`
  - `:20` (routing table): `| RFC, design proposal | *(not yet available — STOP and tell the user the RFC reference has not been authored yet)* | `RFC-NNN-<slug>.md` |`
  - `references/` contains only `adr.md` (11 859 B) and `spec.md` (17 539 B).
- **The only de-facto convention is the two exemplars + the research LOG narrative**: research reports → `SYNTHESIS` → "Extended **RFC-style proposal**" (LOG.md:155) → owner greenlight → RFC drafted section-by-section → owner ratifies → D/P register appended in-place (Appendix A) → downstream ADR/tech-spec candidates listed.
- Drafts were physically staged at `.pi/subagents/proposals/json-cel/RFC-002-open-validation-engine.md` (LOG.md:441) then moved to the `kb/rfc/` dir; that staging dir now holds only `research/`.
Confidence: HIGH.

---

## Q4 — Harness consumption points (who reads an RFC, when, what gate)

**Case-insensitive `rfc` grep over `pi-meta-config/{global-harness,.pi}` (excl. `.pi/research/`, `.pi/recon/`, artifacts) → 5 hits, all in docs-writer:**
| file:line | context | is it about RFC documents? |
|---|---|---|
| `global-harness/skills/docs-writer/SKILL.md:3` | "Will also route RFC authoring once that reference lands" | YES (says: not authored) |
| `global-harness/skills/docs-writer/SKILL.md:20` | routing row: RFC → "not yet available" | YES (says: STOP) |
| `global-harness/skills/docs-writer/references/adr.md:7` | `[RFC2119] [RFC8174]` BCP-14 citation | NO — standard reference |
| `global-harness/skills/docs-writer/references/spec.md:6` | `[RFC2119]` cit. | NO |
| `global-harness/skills/docs-writer/references/spec.md:7` | `[RFC8174]` cit. | NO |

**No agent, prompt, command, or profile references an RFC document.** The authority chain today:
- Producer: execution-profile `§AUTHORITY` section, pasted verbatim as `AUTHORITY DOCS` — `global-harness/profiles/rust.md:47`, `global-harness/profiles/go.md:44`; consumed in `execute-plan/SKILL.md:379, 612, 687–688, 827, 900–901`.
- Consumers: `global-harness/agents/spec-builder.md:35–61` and `global-harness/agents/fix-designer.md:36–70` — precedence `ADR > SPEC; the plan is the floor; ADDITIONAL INSTRUCTIONS outrank the plan`.
- **Gate (the closest thing to an RFC acceptance gate that exists):** `spec-builder.md:56–58` (identical `fix-designer.md:65–67`): `a SPEC is binding only while `status: active` AND `verified.state: live`; an ADR only while `status: accepted`. A document failing its gate is advisory — never quote it as binding…`. Failing frontmatter is recorded as a `discoveredRisks` item with `authority.quote` = the failing line verbatim.
- **The authority enum is `["adr","spec","plan","none"]`** (`spec-builder.md:13` outputSchema, `authority.doc`) — **no `rfc` member**: an RFC can be neither named nor quoted as authority by any consumer.
- `pi-meta-config/.pi/exec-profile.md` has **no §AUTHORITY** section at all (only `GATES`/`SYMBOL-TOOLS`/`TERMINOLOGY`/`CONVENTIONS`), so in this repo `AUTHORITY DOCS` resolves to the literal `none` (`execute-plan/SKILL.md:687–688`).
- In the **akb project** (not the harness), the RFC is consumed as prose authority by the memory-layer design docs: `agent-memory/DESIGN.md:381` `**RFC-002 (…RFC-002-open-validation-engine.md) is ratified as akb's direction**`, `:567` listing RFC-002 + init-versioning spec as inputs; `STATE.md:82` D14 records the `supersedes RFC-001` adoption. The RFC's §15/§18 candidate lists are the hand-off node into ADR/SPEC authoring.

**Inference (marked):** RFCs would slot into the existing machinery as a *pre-authority* doc — read by the human/owner and by `docs-writer` when authoring the ADR/SPEC the RFC names; they are **not** read by any automated gate today and are not admissible authority. No code change to authority enums was found that anticipates them. Confidence: HIGH for the negative; MEDIUM for the "natural seam" inference.

---

## Q5 — Resemblance to the docs-writer ADR/SPEC contracts

docs-writer contracts (`global-harness/skills/docs-writer/references/adr.md`, `spec.md`, shared grammar in `SKILL.md`):
- **YAML frontmatter, closed-world** — `adr.md:59–62` "Every ADR MUST begin with a YAML frontmatter block — **closed-world: unknown keys are errors**". Required: `grammar` (=1), `type`, `id`, `title`, `status`, `created`, `updated`, `provenance`, `scope`, `tags`/`summary` (ADR), `kind`/`verified`/`anchors` (SPEC); `supersedes`/`superseded_by`/`revisit` (`adr.md:78–82`).
- **Status enums** — ADR `proposed|accepted|deprecated|superseded|rejected`; SPEC `draft|proposed|active|completed|superseded|withdrawn` (`spec.md:69`); `only `active` authorizes adjudication quotes` (`spec.md:85`).
- **EARS-2119 grammar** — defined once in `SKILL.md:52–72` (6 clause shapes; tether rule; atomicity; ALL-CAPS only), references point at it.
- **Stable IDs + anchors** — `I1…`, `N1…` rules; SPEC `anchors` claim inventory with `verify.check` commands and `verified.{at,branch,commit,method}` staleness block (`spec.md:74–81`); mandatory "challenge affordance" blockquote.

**The RFC exemplars share almost none of this** — they are a **pre-docs-writer, third format**:

| axis | docs-writer ADR/SPEC | RFC exemplars |
|---|---|---|
| frontmatter | YAML closed-world, validated | **none** — prose/bold markdown table |
| status | enum + human flip + gate | free text (`Draft — awaiting owner ratification`) |
| provenance/ratified_by | required | absent (flip recorded only in LOG.md / Appendix A) |
| 2119/EARS | mandatory constraint grammar | **absent** (0 MUST/EARS hits) |
| rule IDs | `I1`/`N1`, stable | `P1–P5`, `D1–D19`, `P-a…P-i` (design/decision, not constraints) |
| machine anchors | `anchors[]` + `verified{}` | none |
| decision recording | YAML + prose, one decision | Appendix A D/P registers, batch |
| headings | exact H2 skeleton | numbered sections + appendices |

Dating supports this: RFCs are 2026-09-26/27; the docs-writer-format ADR/SPEC files are 2026-10-01.

**Is the docs-writer family in use in akb? YES — but in a different tree than the RFCs:**
- `…/llm-wiki/agent-kb/agent-kb/docs/knowledge-base/kb/decisions/ADR-001-akb-doctor.md`, `ADR-002-init-author-flags.md`, `specs/SPEC-001-init-versioning.md` carry exactly the contract frontmatter (`grammar: 1`, `type: adr|spec`, `id`, `title`, `status`, `provenance`, `scope`, `tags`/`summary`/`revisit`, `verified:`/`anchors:`) and the exact H2 skeletons (`## Decision / ## Invariants / ## Negative Constraints / ## Exceptions / ## Verification / ## Context / ## Decision Drivers / ## Alternatives Considered / ## Consequences / ## References`; SPEC: `## Why / ## Goals and Non-Goals / ## Current Behaviour / ## Requirements / ## Acceptance Criteria / …`). ADR-001 invents EARS rules with IDs: `- **I1**: WHILE `akb doctor` runs without `--fix`, the command MUST NOT modify base or repository state; …`.
- Backed by machine templates `…/docs/knowledge-base/.agent-kb/templates/adr.yaml` (`status: enum [proposed, accepted, deprecated, superseded, rejected]`) and `spec.yaml` (`status: enum [draft, proposed, active, completed, superseded, withdrawn]`, `verified`, `anchors`) plus `akb.yaml` `versioning: git`.
- **But the RFC's own dir siblings are NOT that family:** `…/llm-wiki/agent-kb/kb/adr/` is **empty**; `kb/spec/init-versioning-spec.md` is pre-format ad-hoc (`# Spec: …` / `Status: agreed design, pre-implementation` / numbered `## 1.`…`## 10.`, no frontmatter, 2026-09-25). So "docs/kb/adr" does not exist and the shared `kb/` tree predates the owner format.

Confidence: HIGH.

---

## START HERE (for the coordinator)
1. `…/llm-wiki/agent-kb/kb/rfc/RFC-002-open-validation-engine.md` lines 1–14 + 739–806 — the fullest exemplar header, decision register, and P-register in one place.
2. `global-harness/skills/docs-writer/SKILL.md` lines 1–72 — the router + shared EARS-2119 grammar the future RFC reference must conform to (or deliberately diverge from).
3. `global-harness/agents/spec-builder.md` lines 33–62 — the authority-gating pattern (`ADR > SPEC > plan`, status/verified gate, enum `["adr","spec","plan","none"]`) an RFC would have to extend.

## GAPS / OPEN QUESTIONS
- Exact exemplar path differs from the brief; the RFCs live in a **non-git scratch workspace** — confirm with the owner whether that is the intended canonical home or an artifact of a move.
- RFC status vocabulary is free text; no `rfc.yaml`, no frontmatter, no anchors → an RFC cannot be machine-gated as-is. Whether the future RFC reference should reuse ADR/SPEC's closed-world frontmatter + EARS-2119 grammar, or stay prose-heavy, is the core design question and is **not answered by any existing artifact**.
- Whether RFCs become a first-class `authority.doc` enum member (pre-ratification authority above ADR?) or remain a pre-authority proposal has no precedent in the harness — DESIGN.md's "ratified as direction" prose is the only usage.
- `docs/kb/spec/init-versioning-spec.md` (ad-hoc) vs `docs/knowledge-base/kb/specs/SPEC-001-…` (contract format) coexist in akb — migration/convention-detection behavior for RFC placement is undetermined.
