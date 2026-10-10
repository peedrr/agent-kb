---
grammar: 1
type: research
title: "RFC-002-S6 requirement extraction — RFC-002 sections, ADR-010, SPEC format contract"
status: final
provenance: agent-drafted
run: rfc-002-s6
as_of: "2026-10-09"
created: "2026-10-09"
updated: "2026-10-09"
scope:
- docs/knowledge-base/kb/**
tags:
- rfc-002
- templatev3
- requirements
summary: "Every requirement RFC-002 imposes on S6 with section anchors, the S6/S7/S8 boundary table, ADR-010 verbatim invariants, and the SPEC-003 page format contract."
informs:
- SPEC-003
---

# RFC-002-S6 Extraction — TemplateV3 Core Format + Loader (drafts as SPEC-003)

Sources:
- `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md` (938 lines; cited as RFC §N)
- `docs/knowledge-base/kb/decisions/ADR-010-templatev3-minimal-key-set.md` (ADR-010)
- `docs/knowledge-base/kb/decisions/ADR-006-schema-frontmatter-retirement.md` (ADR-006)
- `docs/knowledge-base/kb/decisions/ADR-008-phase1-current-document-shape.md` (ADR-008)
- `docs/knowledge-base/kb/decisions/ADR-009-phase1-sweep-schema-checker.md` (ADR-009)
- `docs/knowledge-base/kb/specs/SPEC-002-structured-validation-report.md` (SPEC-002, exemplar)
- `docs/knowledge-base/.agent-kb/templates/spec.yaml` via `akb template get spec --full`
- Loader anchors: `internal/template/template.go:43-51`, `:73-78`, `:106,114`; `cmd/akb/templates_write.go:461-469` (detectOldFormat copies)

---

## 1. S6 REQUIREMENTS

Each derives from RFC-002 (candidate RFC-002-S6, §18) plus the decision IDs named. Anchors in parens.

| # | Requirement | Anchor | Derives from |
|---|---|---|---|
| R1 | TemplateV3 is a **breaking format change**, loudly rejected like the v1/v2 rejections (no silent acceptance). | RFC §6 heading | D2 posture; RFC-010 context |
| R2 | The Phase-1 loader and `akb template write` MUST recognize **exactly six** top-level keys: `name`, `description`, `dir`, `schema`, `validations`, `lint_rules`. | RFC §6 "Phase-1 key set"; ADR-010 I1 | ADR-010 |
| R3 | WHEN `akb template write` sees any other top-level key it MUST refuse with **exit 1** and a message naming the offending key **and** the full supported set. | RFC §6; ADR-010 I2 | ADR-010; D9 principle |
| R4 | WHEN the **load path** sees an unrecognized top-level key in an otherwise-V3 template it MUST warn on stderr naming the key and **continue** loading the recognized keys (never hard-fail). | RFC §6 "load time only warns"; ADR-010 I3, N2 | ADR-010; RFC-002-A13 forward-compat |
| R5 | The retired `schema.frontmatter` homegrown block MUST be hard-rejected via **one shared detection helper** with **one self-sufficient message** (names file, retired construct, 4-step migration essence, CHANGELOG home); no migrate command, no alias. | RFC §6; ADR-006 I1–I7, N1/N2 | ADR-006 |
| R6 | WHEN a template mixes the retired `schema.frontmatter` block with V3-shaped keys, the retired-block refusal MUST apply **before** any V3/unknown-key judgment. | ADR-006 I5; ADR-010 I5 | ADR-006; ADR-010 |
| R7 | The `schema:` key value MUST be a **JSON Schema 2020-12** document. In Phase 1 it validates the **current** document shape — paths `/frontmatter/*`, `content.raw`, `ast.headings`, `akb.*`, `file.*` — NOT the flattened `derived.*` shape. | ADR-010 I6; ADR-008 I1/I2, N1 | ADR-008; RFC §5.2/§9 |
| R8 | Presence moves from the V2 `required`/`optional` lists to schema `required`; the **guard invariant** holds: the schema `required` set under `properties.frontmatter` is exactly the set of frontmatter keys CEL rules may read unguarded. | RFC §6 | RFC §6 |
| R9 | `type`/`title` stop being engine gates (D10): `frontmatter.Parse` MUST stop routing `type`/`title` **out of** `Fields`; they become ordinary keys visible to **both** validators. | RFC §6; D10 | D10 |
| R10 | `type` survives **only** as the template selector: a page naming an unknown template is refused at write (the one surviving door-bouncer, P7) and flagged by the `type_orphan` sweep checker. | RFC §6; RFC §11; D10 | D10 |
| R11 | `validations:` stays UNCHANGED: CEL, write-time, **fail closed**. | RFC §6 | P4; RFC §6 |
| R12 | `lint_rules:` stays UNCHANGED: CEL, sweep-time. | RFC §6 | P4; RFC §6 |
| R13 | `dir` MUST keep its **current** semantics throughout Phase 1; the rename to `path` (D11 resolution-hint) rides Phase 3. | ADR-010 I4, N3; RFC §6 "`dir` renamed/re-specified as `path`" | D11 |
| R14 | S6 MUST NOT rename `dir`→`path` in the loader, the embedded templates, or this repo's KB templates. | ADR-010 N3 | ADR-010 |
| R15 | Structural overlap rejection carries from RFC-001 §6.2 as a **warning, not an error** (permissive P2). ⚠ likely S8-owned (see Boundaries). | RFC §6 | P2 |
| R16 | The **embedded templates** MUST be converted to V3 (the in-binary `internal/template/embedded/*.yaml`). | RFC §18 S6 row; ADR-010 scope | ADR-008 I5 (two embedded) |
| R17 | The loader is non-strict today (plain unmarshal, silently ignores keys — `internal/template/template.go:106,114`); S6 closes that trap for templates by applying the D9 principle: unknown key refused at authoring, warned at load. | ADR-010 Context/Decision Drivers; RFC §6 | D9 |
| R18 | Forward compatibility: a template authored by a newer akb never bricks an older binary (symmetric with ADR-006 N2). | RFC §6; ADR-010 N2 | ADR-010 |
| R19 | Deferred V3 keys (`state_field`, `functions`, `projections`, `title_field`, `summary_field`, `strip_on_approve`, `path`) MUST NOT be recognized-then-ignored on any Phase-1 path; they are rejected at write / warned at load. | ADR-010 N1 | ADR-010 |
| R20 | Defaults for the new keys (all deferred): `functions:` empty; `projections:` none; `state_field:` absent = no lifecycle; `approve` on such a page is a usage error pointing at `lint`. (Phase-1 relevant only as "not yet recognized".) | RFC §6 defaults; P-c | P-c; A15 |

**Confidence:** HIGH — every requirement is anchored to a ratified ADR invariant or a distinct RFC §6/§18 statement. R15's ownership is the only ambiguity (see GAPS/G1).

---

## 2. BOUNDARIES — what S6 owns vs S7/S8/ADRs

| Area | Owner | Anchor |
|---|---|---|
| Template **format** (six keys, strictness, `type`/`title` un-routing, V2→V3 rejection, loader) | **S6** | RFC §18 S6 row; ADR-010 I1–I3 |
| Embedded-template conversion to V3 | **S6** | RFC §18 S6 row |
| Schema-pass **integration**: document construction, pipeline placement, compiled-schema cache, error-tree → SPEC-002 report mapping, pre-gate removal | **S7** | RFC §18 S7 row (drafts SPEC-004) |
| Sweep-time schema checker | **S7** (ADR-009) | RFC §18 S7; ADR-009 I1–I6 |
| `template write` **schema obligations**: mockup-vs-schema proofs, `<!-- FAILS: schema: ... -->` parser, overlap warning, ADR-005 static-check wiring | **S8** | RFC §18 S8 row (drafts SPEC-005); ADR-006 N7 |
| V2→V3 migration recipe/message | ADR-006 (not S6) | ADR-006 I1–I7 |
| Date bridge coercion / temporal declarations | ADR-005; surfaced by S8's static check | RFC §5.2; RFC-002-A4 |
| Validator library + format assertion policy | ADR-003 (S7 consumes) | RFC §18 A1; ADR-003 |
| Flatten (`derived.*`, `content` string, `word_count()`, `write --json`) | Deferred to flatten cohort | ADR-008 I1–I5, N1–N4 |
| `state_field` (Phase 2), `path`/`projections`/`title_field`/`summary_field` (Phase 3), `functions` (Phase 4) | Later phases | ADR-010 Context; RFC §17 |
| Phase-1 validation core as a whole (santhosh v6, pipeline reorder, merged error model, sweep checker) | Phase 1 / S6–S8 + ADR-003/009 | RFC §17 |

Key boundary facts:
- S6 is a **Phase-1** tech-spec (RFC §17 "Phase 1 — validation core"; RFC §18 S6 row Phase column = 1). Phase 1 is bounded by ADR-008 (current document shape) and ADR-010 (minimal key set).
- `FAILS:` parser + mockup-vs-schema proof explicitly moved **out** of SPEC-002 (its Non-Goals) and **into** Phase-1 `template write` work under S8 (ADR-006 N7). S6 must not claim them.
- S6 must NOT add: `word_count()`, `write --json`, `derived.*` schema paths, `path:` rename, `title_field:` (ADR-008 N1/N3/N4, ADR-010 N1/N3).
- ADR-010's scope includes `internal/template/**`, `cmd/akb/**`, `docs/knowledge-base/.agent-kb/templates/**`, `internal/skill/embedded/**` — so S6 arguably extends to the repo's **four KB templates** (adr/research/rfc/spec), not only the two embedded ones (ADR-008 I5 says "two embedded + four KB templates"). The S6 row's phrase "embedded template conversion" is narrower than ADR-010's scope.

**Confidence:** HIGH on S7/S8/D-register boundaries; MEDIUM on the KB-template-conversion ownership (G2).

---

## 3. SPEC FORMAT CONTRACT (for the SPEC-003 page itself)

### 3.1 Frontmatter (from `spec.yaml` via `akb template get spec --full`)

**Required** (V2 `required: true`; becomes schema `required` under V3, ADR-010 I6):
`grammar` (int), `type` (string = `spec`), `id` (string, `^SPEC-[0-9]{3}$`), `title`, `kind` (`change`|`capability`), `status` (`draft`|`proposed`|`active`|`completed`|`superseded`|`withdrawn`), `provenance` (`human`|`agent-drafted`), `created`, `updated`, `scope` (list, non-empty), `verified` (map).

**Optional:** `ratified_by` (list), `supersedes` (string), `superseded_by` (string), `anchors` (list).

`verified` map shape (from SPEC-002 frontmatter, lines 1–45): `{at, branch, commit, method, next_review_by, state}` with `state ∈ live|drifted|lost|unverified`. `anchors[]` each require `claim, path, symbol, kind ∈ local|flow|invariant` (template rule `anchors_role_bearing`) plus `verify`, `evidence`, `state`, `checked_at` (SPEC-002 exemplar).

Frontmatter values from SPEC-002 (exemplar): `grammar: 1`, `type: spec`, `id: SPEC-002`, `kind: change`, `status: active`, `provenance: agent-drafted`, `is_draft: false`, `ratified_by: [Pete Hope]`, `scope: [cmd/akb/**, internal/**]`, `verified.state: live`.

### 3.2 Required body sections (H2 headings, enforced by template CEL rules)

Required always: **Why** (present, not heading-enforced), **Goals and Non-Goals**, **Requirements**, **Acceptance Criteria**, **Assumptions and Open Questions**, **Affected Surface and Ordering**, **Verification Plan**, **Drift Ledger**, **Revisit Triggers**.
Conditional: **Current Behaviour** — required when `kind: capability` (`current_behaviour_for_capability`).
Optional: **Decisions and Rejected Alternatives** (if present, must contain the phrase `Do not re-propose unless`), **Contract and Invariants** / **Preserved Behaviour** / **Evidence Appendix** (used by SPEC-002, not heading-enforced).

Body content gates: verbatim challenge affordance containing `this SPEC is authority only while` (`challenge_affordance`); ≥1 `REQ-[0-9]{3}`; ≥1 `AC-[0-9]{3} (verifies REQ-[0-9]{3})`; no `[NEEDS CLARIFICATION`; body ≥50 words (`min_word_count`); every fenced code block carries a language (`code_blocks_have_language`).

### 3.3 EARS / acceptance conventions (from SPEC-002)

- Requirements numbered `REQ-NNN`, EARS-2119 grammar: `WHEN <trigger>, the system MUST <response>`; `The system MUST …`; `WHILE <state>, … MUST …`. One speech-act per requirement.
- Acceptance criteria numbered `AC-NNN (verifies REQ-NNN):`, each with a `verify:` entry — `{ method: check, check: "<command>" }` (hermetic) or `{ method: ask, ask: "<human-only>" }`.
- Anchors cite `claim: REQ-NNN` with a resolvable `path` + `symbol` (no line numbers; symbol-first).
- Drift Ledger is append-only `date · doc · from → to · reason · evidence`.
- Revisit Triggers are path globs checkable against a diff.
- `[ASSUMPTION: …]` markers for open assumptions; unresolved ambiguity blocks ratification.

> Note: `spec.yaml` on disk is currently **V2** (`schema.frontmatter`). SPEC-003 is authored under current akb, so it must satisfy this V2 form until S6 converts templates. The **content** requirements above are unchanged by the format conversion (ADR-006 I4: presence moves to schema `required`).

**Confidence:** HIGH — taken verbatim from `akb template get spec --full` and the SPEC-002 exemplar structure.

---

## 4. ADR-010 — key decisions (verbatim)

- **Decision:** "Phase-1 TemplateV3 is a strict, minimal format: the loader and `akb template write` recognize exactly six top-level keys — `name`, `description`, `dir`, `schema`, `validations`, `lint_rules` — and every other key is refused at authoring time with a message that names the offending key and the supported set … while the load path warns only so a newer template never bricks an older binary."
- **I1:** loader MUST recognize exactly `name`, `description`, `dir`, `schema`, `validations`, `lint_rules`.
- **I2:** `akb template write` on any other top-level key → refuse, **exit 1**, message naming the unrecognized key + the I1 supported set.
- **I3:** load path on an unrecognized top-level key in an otherwise-V3 template → **warn on stderr** naming the key, continue loading recognized keys.
- **I4:** `dir` MUST keep current semantics throughout Phase 1; rename to `path` (RFC §8.3, D11) rides Phase 3.
- **I5:** retired `schema.frontmatter` shape → apply ADR-006's refusal before any unknown-key judgment.
- **I6:** `schema` value MUST be a JSON Schema 2020-12 document over the current document shape, per ADR-008 I1/I2.
- **N1:** MUST NOT recognize-then-ignore deferred keys (`path`, `state_field`, `functions`, `projections`, `title_field`, `summary_field`, `strip_on_approve`).
- **N2:** load path MUST NOT hard-fail on unknown top-level keys in an otherwise-V3 template.
- **N3:** MUST NOT rename `dir`→`path` in loader, embedded templates, or this repo's KB templates during Phase 1.
- **No exceptions.** "A template that needs a deferred key is a template waiting on a later phase."
- **Rationale:** "an unenforceable declaration is a trap; traps fail loudly at authoring time or they fail silently in production." V2's `FieldSchema.Type` is the cautionary precedent (declared-but-dead).
- **Phase map (ADR-010 Context):** `schema`+`validations`+`lint_rules` Phase 1; `state_field` Phase 2 (A15); `path`, `projections` (A17), `title_field`/`summary_field` (A13) Phase 3; `functions` (A5) Phase 4.
- **Code anchors at ratification (`a316f0a`):** `internal/template/template.go:43-51` (struct keys), `:106,114` (non-strict unmarshal), `:73-78` and `cmd/akb/templates_write.go:461-469` (the two `detectOldFormat` copies that must converge per ADR-006 I3).

**Confidence:** HIGH — quoted directly.

---

## 5. D-register excerpts used (RFC Appendix A)

- **D3** — Managed fields: Template-declared + input=output. (RFC §6, §8)
- **D7** — Dates: Schema-declared bridge. (RFC §5.2, §6; ADR-005)
- **D9** — akb.yaml: Behavioral config home + unknown-key strictness (principle applied to templates in §6/ADR-010). (RFC §6, §14)
- **D10** — Type gate: Keep, **selector only**. (RFC §6, §11)
- **D11** — Layout: `path:` resolution-hint only; caller path verbatim; policy via CEL/schema. (RFC §6, §8.3 — deferred to Phase 3)
- **D12** — Envelope: Flatten: `content` = markdown string everywhere; `derived` output-only. (RFC §5 — deferred by ADR-008)
- **D13** — Metrics: `char_count` scrapped; `word_count` = registry function; facts-not-computations. (RFC §5 — deferred by ADR-008)
- **D18** — `title`: Declared display field (`title_field:`, default `title`); documented fallbacks. (RFC §6, §12 — key is Phase 3)
- **P-c** — `approve` without `state_field` = usage error pointing at `lint`. (RFC §6)

**Confidence:** HIGH — copied from Appendix A and the P-register table.

---

## 6. GAPS

- **G1 (ownership):** §6's "overlap rejection carries as a warning" and the full mockup-obligation set are not clearly assigned between S6 (format/loader), S8 (`template write` schema obligations). ADR-006 N7 puts the `FAILS:` parser in Phase-1 `template write` = S8. Resolve before drafting SPEC-003's requirements boundary.
- **G2 (scope of "embedded template conversion"):** the S6 row says "embedded template conversion" (2 in-binary templates) but ADR-010's scope and ADR-008 I5 name **six** templates (2 embedded + 4 KB: adr/research/rfc/spec). Decide whether S6 also converts `docs/knowledge-base/.agent-kb/templates/*.yaml`.
- **G3:** SPEC-003 is authored while `spec.yaml` is still V2, so the exemplar format contract is the pre-V3 shape; confirm the SPEC-003 page is written under current akb (V2 template) and does not assume S6's output.
- **G4:** §6's full V3 example shows `path`, `state_field`, `functions`, `projections`, `title_field`, `strip_on_approve` — a SPEC-003 reader could mistake them for Phase-1. ADR-010 already forestalls this; SPEC-003 must restate the six-key set explicitly.
- **G5:** exact wording of the I2 refusal message and the six-key supported set is human-reviewed residue (ADR-010 "Human-only residue"); SPEC-003 must pin the message contract, not just the exit code.
- **G6:** `dir` currently is a **placement** field (V2). ADR-010 I4 keeps its semantics; confirm whether Phase-1 `dir` retains any placement behavior or only the current meaning, before writing acceptance criteria.
