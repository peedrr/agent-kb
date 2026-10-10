---
as_of: "2026-10-10"
created: "2026-10-10"
grammar: 1
informs:
- SPEC-004
is_draft: false
provenance: agent-drafted
run: rfc-002-s7
scope:
- docs/knowledge-base/kb/**
status: final
summary: Verbatim S7 row + S1–S6 one-liners, SPEC inventory (confirming SPEC-004 absent), S7 deliverables/open questions with anchors, and prior S-run characterization.
tags:
- rfc-002
- spec-004
- schema-pass
- requirements
title: RFC-002-S7 extraction — S7 scope, SPEC-004 context, prior-run characterization
type: research
updated: "2026-10-10"
---

# RFC-002-S7 Extraction — Schema-Pass Integration (drafts as SPEC-004)

Sources (read this run):
- `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md` (945 lines; cited as RFC §N / `:line`)
- `docs/knowledge-base/kb/decisions/ADR-003,006,008,009,010`
- `docs/knowledge-base/kb/specs/SPEC-002-structured-validation-report.md`, `SPEC-003-templatev3-format-loader.md`
- `docs/knowledge-base/kb/research/rfc-002-phase1-scope-log.md`, `rfc-002-s6-rfc-extract.md`

---

## 1. S7 — verbatim definition

`RFC-002-open-validation-engine.md:834` (the §18 candidate table, the S7 row):

> | RFC-002-S7 | SPEC | 1 | Schema-pass integration: document construction, pipeline placement, compiled-schema cache, error-tree → SPEC-002 report mapping, pre-gate removal, sweep checker (ADR-009) — drafts as SPEC-004 | proposed | — |

Status: **proposed** (not ratified — unlike S1 and S6, which have `ratified` rows resolving to SPEC-002/SPEC-003). `Resolves-as` is blank, so no SPEC-004 page has been created. Phase column = **1**.

Supporting/adjacent text that further defines S7 (not the row itself):
- RFC §18 preamble `:793-806` — candidate lifecycle: S7 is a *candidate*, not implementation-final; when drafted it takes the next free global `SPEC-NNN` and records `Origin: RFC-002 candidate RFC-002-S7`.
- RFC §18 renumbering note `:839-856` — rows S6–S8 and the Phase column were added 2026-10-09; S6–S8 "enumerate the Phase-1 tech-specs scoped by ADR-008/009/010 (drafting as SPEC-003/004/005)".
- RFC §17 Phase 1 bullet `:763-770` — Phase 1 = "validation core"; tech-specs RFC-002-S6/S7/S8; bounded by ADR-008 (current document shape) and ADR-010 (minimal V3 key set).

---

## 2. S1–S6 — one-liners with anchors

All rows are in the RFC §18 table:

- **S1** (`:815`) — SPEC 1+2 — "Structured validation report (widened `runTemplateValidations`)" → **ratified** as `SPEC-002`. (Report internals land Phase 1; stdout emission lands Phase 2.)
- **S2** (`:818`) — SPEC 3 — "(a′) AST-derived exclusions — implementation tech-spec" — proposed.
- **S3** (`:820`) — SPEC 4 — "`kb.*` execution model (fact surface, freshness, cost semantics)" — proposed.
- **S4** (`:831`) — SPEC 3 — "Search: `search.index_fields` DDL + rebuild, `--tag` matching semantics, `documents.updated` disposition" — proposed.
- **S5** (`:832`) — SPEC 2 — "Input-schema derivation mechanics (normalization, reference keyword-walk)" — proposed.
- **S6** (`:833`) — SPEC 1 — "TemplateV3 core format + loader: minimal key set (ADR-010), strict keys, `type`/`title` un-routing (D10), embedded template conversion" → **ratified** as `SPEC-003`.
- **S7** (`:834`) — SPEC 1 — see above — proposed, drafts SPEC-004.
- **S8** (`:835`) — SPEC 1 — "`template write` schema obligations: mockup-vs-schema proofs, `FAILS:` parser (ADR-006 N7), overlap warning, ADR-005 static-check wiring — drafts as SPEC-005" — proposed.

S-phase → ADR mapping (from RFC §18 rows `:810-838`): A1→ADR-003, A2→ADR-006, A3→ADR-004, A4→ADR-005, A19→ADR-008, A20→ADR-009, A21→ADR-010. S1→SPEC-002, S6→SPEC-003, S7→SPEC-004 (pending), S8→SPEC-005 (pending). No S-phase maps to an ADR directly; the Phase-1 scope ADRs (008/009/010) bound S6/S7/S8.

---

## 3. SPEC inventory

Spec pages live in `docs/knowledge-base/kb/specs/`. `type: spec` template is `docs/knowledge-base/.agent-kb/templates/spec.yaml`.

| Path | ID | Title | RFC-002 phase / origin |
|---|---|---|---|
| `docs/knowledge-base/kb/specs/SPEC-001-init-versioning.md` | SPEC-001 | Init Selects the KB Versioning Mode and Resolves Commit Identity | Precondition (D-seq) — lands **before** Phase 1; status `completed`; not an S-candidate (`depends_on: []`) |
| `docs/knowledge-base/kb/specs/SPEC-002-structured-validation-report.md` | SPEC-002 | Write-Time Validation Failures Merge Into One Structured Report | **S1** — Phase 1 internals + Phase 2 stdout; status `active`; `depends_on: [ADR-003, write --json]` |
| `docs/knowledge-base/kb/specs/SPEC-003-templatev3-format-loader.md` | SPEC-003 | Templates Adopt a Strict Six-Key V3 Format With a JSON Schema Block | **S6** — Phase 1; status `active`; `depends_on: [ADR-006, ADR-008, ADR-010]`, **`blocks: [SPEC-004, SPEC-005]`** |

**SPEC-004 does not yet exist.** Confirmed by directory listing (only SPEC-001/002/003) and a repo-wide search: the string `SPEC-004` appears only prospectively/proleptically (RFC §18 `:834`; SPEC-003 `:107,119,161,288,312,313,316,359,385`; research logs). No `SPEC-004*.md` file anywhere.

Key forward-reference from SPEC-003 that constrains SPEC-004's starting point:
- SPEC-003 REQ-007 `:160-161` — the Phase-1 loader MUST NOT compile/evaluate `schema:`; it hands the parsed payload **uncompiled** to the schema pass (SPEC-004). AC-007 `:228` gates it: `internal/template/` imports no JSON Schema compiler, `go.mod` gains no validator dependency.
- SPEC-003 `:312-313` — "Rejected: compile-at-load. Do not re-propose unless SPEC-004's design proves pass-through unworkable."
- SPEC-003 `:359` — `blocks: [SPEC-004, SPEC-005]`; `:385` revisit trigger — "SPEC-004 ratification — re-check REQ-005/REQ-007 boundary text."

---

## 4. What S7 must deliver (per the RFC + bound ADRs)

The §18 row `:834` names six work items. Their RFC/ADR elaborations:

1. **Document construction.** RFC §7 `:336-339`: "build document (frontmatter + content parse, goldmark parse, declared projections materialized)" is an explicit pipeline step because the schema may constrain the materialized document (`contains` over `derived.ast.headings`). Placement/ordering is a genuine design point ("document-model audit, open question 4"). **Bound by ADR-008 I1**: Phase-1 document shape is the **current** one — `frontmatter` (raw YAML), `content` as `{raw, word_count, char_count}`, `ast` as `{headings, links, code_blocks}`, `akb` as `{provenance_markers, annotations}`, `file` as `{path,name,dir}`. **Not** the §5.1 flattened shape.
2. **Pipeline placement.** RFC §7 `:331-338` gives the write/append pipeline: `merge-conflict check → selector gate (type names a template; template loaded with its projections/functions declarations) → build document → JSON Schema → CEL validations → write`. RFC §7 `:340-346`: the required-presence gate is the schema's own `required` (**no separate pre-gate**), and **no mutation step exists** — pass writes bytes, fail writes nothing.
3. **Compiled-schema cache.** ADR-003 I8 `:43`: "compile each template schema once and cache the compiled schema keyed by template identity; every `RegisterFormat` call MUST precede compilation." ADR-003 I2 `:40`: `Compiler.AssertFormat()` on every compiled schema, **write time and sweep time**. ADR-003 Consequences `:104`: `Compiler.Compile` is not thread-safe and concurrent `Validate` on one compiled schema is undocumented — carried constraint. ADR-003 N2/N3/I4/I5 require akb's own strict-profile `date-time`/`date` checkers registered **before** compilation; I6: `format: time` asserted as string, never coerced.
4. **Error-tree → SPEC-002 report mapping.** SPEC-002 REQ-002/003/008 (`SPEC-002:126-148`): schema failure **short-circuits CEL**; each schema row carries the failing **`keyword`** and the **JSON Pointer** of the instance location plus a single-line message; schema rows MUST NOT carry `line`. SPEC-002 REQ-006: schema rows before CEL rows. SPEC-002 REQ-005 report shape (`:151-168`) + `summary: {schema, cel, total}`. SPEC-002 assumes the santhosh v6 error tree is `ValidationError{SchemaURL, InstanceLocation []string, ErrorKind, Causes}` with a public `kind` package — re-verify against the go.mod-pinned version (SPEC-002 Assumptions `:269-272`).
5. **Pre-gate removal.** RFC §7 `:343-345`; RFC §16 `:731-737` folds the required-field pre-gate into the schema pass; SPEC-002 Current Behaviour `:88-90` names `cmd/akb/required_fields.go:18` as the gate to remove; ADR-009 I6 `:56` — the Phase-1 pre-gate removal MUST NOT extend to the sweep (`required_fields` lint checker stays per ADR-009 I5).
6. **Sweep checker (ADR-009).** ADR-009 Decision+Invariants: add a built-in `schema` lint checker validating each page's document view against its type's compiled schema, with ADR-003 I2/I3 format assertion (I1); runs **before `cel_lint`** and **shares one `now`** (I2); on eval fault **degrades to a per-page issue, sweep continues** (I3/N3); emits `LintIssue` rows, **not** the SPEC-002 write-time report row shape (I4/N2); `required_fields` stays registered so `by_check` never loses a key (I5/N1). RFC §11 `:519-521` restates the ordering rule.

Guard invariant (S7-relevant): RFC §6 `:287-290` — "the schema `required` set under `properties.frontmatter` is exactly the set of keys CEL rules may read unguarded." RFC §7 `:346-349` — error merge (schema by keyword + JSON Pointer; CEL by rule ID; one exit-1 report) and the `runTemplateValidations` widening carry unchanged to SPEC-002.

---

## 5. Explicit open questions / decision points parked for S7

The RFC has **no dedicated "open questions for S7" list** — they are scattered. Collected:

- **Report-type package home** — SPEC-002 Open Questions `:274-280`: "the Go package hosting the report type (`internal/cel` is CEL-only; a schema+CEL report wants a new home) — an implementation choice." SPEC-002 v6 is CEL-free, so S7 must decide where the combined report + compiled-cache live.
- **Pipeline ordering vs materialization** — RFC §7 `:336-339` cites "document-model audit, open question 4"; the build step must precede schema validation.
- **Scope of the sweep checker's shared state** — ADR-009 I2/I4 and RFC §11 `:519-521` (ordering + one `now` + `LintIssue` shape) are fixed, but the checker's fault-isolation *depth* is explicitly deferred: ADR-009 revisit + RFC §11 `:511-515` say per-checker fault-isolation depth is "section-11 Phase-A scope, not this record."
- **`approve` refusal-wording parity coverage gap** — `rfc-002-phase1-scope-log.md:195`: "approve refusal-wording parity (`approve.go:221`) has no testscript guard, only `approve_test.go:651,694` — candidate hardening note for SPEC-004." (S7 touches the write/approve refusal wording via pre-gate removal.)
- **Whether santhosh v6 pass-through from the loader stays workable** — SPEC-003 `:312-313` parks the compile-at-load-vs-pass-through decision as contingent on SPEC-004's design (pass-through is the current ratified posture).
- **`format: time` and undeclared-format posture** — resolved by ADR-003 (I6/I7), but S7 must implement: unknown format name → `template write` warns, load path lenient.
- **Not S7's to answer (boundaries):** `FAILS:` parser + mockup-vs-schema proofs + overlap warning → **S8/SPEC-005** (RFC §18 `:835`, ADR-006 N7); flatten/`derived.*`/`word_count()`/`write --json` → later cohort (ADR-008 I1,I4,N1-N4); date-bridge coercion → ADR-005/Phase 2.

---

## 6. Dependencies / ordering constraints for SPEC-004

- **Depends on SPEC-003 (S6):** SPEC-003 `blocks: [SPEC-004, SPEC-005]`; the loader hands `schema` through **uncompiled** (SPEC-003 REQ-007). S7 begins where S6 ends.
- **Consumes ADR-003** (validator library, format assertion, compiled-cache, strict temporal checkers) — S7 is ADR-003's implementation seam ("ADR-003's domain" per SPEC-002 Non-Goals `:91-92`).
- **Consumes ADR-009** (sweep checker invariants I1–I6, N1–N3).
- **Bounded by ADR-008** (current document shape; no flatten, no `write --json`, no `word_count()`).
- **Consumes ADR-006** (schema.frontmatter already retired by S6; ADR-006 N7 puts the `FAILS:` parser in S8) and **ADR-010** (minimal V3 key set: `name, description, dir, schema, validations, lint_rules`).
- **Produces into SPEC-002** (report contract) and is gated by SPEC-002's `depends_on` + open question on report-type home.

---

## 7. Prior S-run characterization (one paragraph)

The prior Phase-1 S-runs (files live under `docs/knowledge-base/kb/research/`, not `rfcs/`) followed a consistent pattern: a run **LOG** (`rfc-002-s1-log.md`, `rfc-002-s6-log.md`), a **prior-research extraction** (ratified ADR invariants + SPEC conventions), an **RFC requirement extraction** (every requirement with section anchors + an explicit S6/S7/S8 ownership boundary table), a **code recon** (read-only `file:line` map of the current loader/validation/error machinery and call sites), and an optional **gapfill/libs** report (pinned-version feasibility, e.g. `rfc-002-s6-gapfill-libs.md` verified goccy v1.19.2 strict-key behavior and santhosh v6.0.3 compile surface). S1 (→SPEC-002) covered the write-time validation error machinery (`runTemplateValidations`, `cel.ValidationError`, `classifyExit`, `required_fields`) and ADR/spec conventions; S6 (→SPEC-003) covered the TemplateV2→V3 format/loader, key-to-consumer inventory, `type`/`title` un-routing blast radius, and test pinning. The S6 run explicitly produced the **S6/S7/S8 boundary table** naming S7 as owner of "Schema-pass integration… and the sweep-time schema checker (ADR-009)" (`rfc-002-s6-rfc-extract.md:73`), which is the cleanest prior scoping statement of S7. The Phase-1 scope run (`rfc-002-phase1-scope-log.md`) then ratified the three forks that bound S7 — validate-current-shape (ADR-008), Phase-1 sweep checker + retained `required_fields` (ADR-009), minimal V3 key set (ADR-010) — and logged `approve` refusal-wording parity as a SPEC-004 hardening note.

---

## 8. Confidence and gaps

**Confidence: HIGH** on (a) the verbatim S7 row and its six work items, (b) S1–S6 anchors, (c) SPEC-004's non-existence, (d) the ADR-003/008/009 invariant set that bounds S7.

**Gaps / uncertainties:**
- The RFC gives **no consolidated S7 open-questions list**; the items in §5 are inferred from scattered RFC/SPEC/ADR/run-log text and may under- or over-count. SPEC-004 authoring should treat §5 as a checklist to confirm, not an authoritative list.
- No S7 research artifacts existed before this run (`.pi/research/rfc-002-s7/` held only `LOG.md`); no prior `rfc-002-s7-*` KB research page exists, so nothing has been lost to reuse.
- The santhosh v6 **error-tree shape** is a SPEC-002 assumption, not re-verified here against the pinned module (SPEC-002 flags re-verification under ADR-003) — a live unknown for S7's error-tree mapping design.
- The **report-type package home** and the **sweep checker's compiler/cache sharing** with the write path are explicitly implementation choices, unresolved.
- Line numbers cited are valid at the current working tree only (RFC `:945` lines; SPEC-003 anchored at commit `5a2ea67`).
