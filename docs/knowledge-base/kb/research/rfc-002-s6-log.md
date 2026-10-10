---
grammar: 1
type: research
title: "RFC-002-S6 research run log — SPEC-003 TemplateV3 format and loader"
status: final
provenance: agent-drafted
run: rfc-002-s6
as_of: "2026-10-09"
created: "2026-10-09"
updated: "2026-10-09"
scope:
- docs/knowledge-base/kb/**
- internal/**
tags:
- rfc-002
- templatev3
- run-log
summary: "Run log for RFC-002-S6: recon fanout, owner-ratified scope forks, and filing record for SPEC-003 (TemplateV3 core format + loader)."
informs:
- SPEC-003
---

# Research run: rfc-002-s6 — LOG

## Question
Understand RFC-002-S6 (Phase-1 tech-spec: "TemplateV3 core format + loader — minimal key set
(ADR-010), strict keys, `type`/`title` un-routing (D10), embedded template conversion"; drafts as
SPEC-003) and write the SPEC via the docs-writer skill. Escalate decision forks to the owner.
Prior research exists in the KB — gap-fill only, do not repeat.

## Run parameters
- Date: 2026-10-09
- MODE: start
- SUBJECT: rfc-002-s6
- KB: /home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb/docs/knowledge-base (AKB_KB set; `research` template present)
- Deliverable: SPEC (KB `specs/` page, drafting as SPEC-003) — destination per RFC-002 §18 row; confirm at filing wave.

## Agreed requirements
Owner-ratified 2026-10-09 via interview (all recommended options accepted):
- **F1 corpus**: S6 converts ALL SIX templates — 2 embedded (adr, note + 4 mockups) AND 4 KB dogfood (adr/rfc/spec/research + 8 mockups) — in the loader's implementation cohort. SPEC-003's own page converts in that wave, not at filing.
- **F2 schema in loader**: opaque passthrough + shape check (present, is a mapping). Compilation/cache/malformed-schema errors = S7. S6 ships zero santhosh coupling.
- **F3 conversion depth**: format-only faithful translation in Phase 1 (required→schema `required`, enum→schema `enum`, dead `type:` fields drop, CEL verbatim). §17 Phase-4 'embedded example restructure' stays Phase 4; SPEC-003 records the split, amends no RFC text.

Requirement source of truth: rfc-extract.md R1–R20 (all ratified ADR invariants) + these three forks.

## Decisions made
- Run slug `rfc-002-s6`; run dir `.pi/research/rfc-002-s6/`.

## Key findings
From recon (anchors in child reports; every item ratified unless noted):
- S6 = SPEC-003: TemplateV3 core format + loader. Six top-level keys only: `name`, `description`, `dir`, `schema`, `validations`, `lint_rules` (ADR-010 I1). Refuse others at `template write` (exit 1, names key+set, I2); warn-only at load (I3/N2).
- Retired `schema.frontmatter` hard-rejected via ONE shared helper, ONE message, applied BEFORE unknown-key judgment (ADR-006 I1–I5; two divergent detectOldFormat copies today: internal/template/template.go:73-80 vs cmd/akb/templates_write.go:461-469).
- `schema:` = JSON Schema 2020-12 over the CURRENT document shape (ADR-010 I6 + ADR-008 I1/I2). No `derived.*`, no flatten.
- `dir` keeps current semantics Phase 1; `path` rename is Phase 3 (ADR-010 I4/N3).
- D10 un-routing: `frontmatter.Parse` stops pulling `type`/`title` out of Fields (frontmatter.go:61-72); blast radius: pagebuilder.go:60-66 re-injection, required_fields.go:38-49, lint/required_fields.go, templates_write.go:395-447, write.go re-marshal :252-256/:352-356/:493-497, append.go, index.go:209, search title column.
- `type` survives only as template selector (unknown type refused at write; type_orphan at sweep).
- Loader today: goccy/go-yaml, non-strict double unmarshal (template.go:106,114); no version discriminator exists; loadExampleTemplates bypasses LoadTemplates (cmd/akb/template.go:104-127).
- SPEC format contract known (spec.yaml V2 gates + SPEC-002 exemplar): frontmatter reqs, H2 skeleton, REQ-NNN EARS, AC-NNN (verifies REQ-NNN), challenge affordance, anchors {claim,path,symbol,kind}, Drift Ledger, Revisit Triggers. SPEC-003 authored under current V2 spec.yaml.
- S1 conventions: status proposed, provenance agent-drafted, verified.state unverified; Origin paragraph linking RFC-002-S6; tags: - rfc-002; index/log via akb commands; draft in run dir, file with --no-commit, owner runs one plain-git commit.

## Decisions (coordinator, logged)
- D-1: Overlap warning + mockup-vs-schema obligations = S8, not S6 (S8 row names them; ADR-006 N7). S6 owns format/loader/strict keys only.
- D-2: Phase-1 `title` display = literal `title` frontmatter key, hardcoded default per D18; `title_field:` machinery is Phase 3 (ADR-010 N1). No fallback machinery in S6.
- D-3: I2 refusal/I3 warning message text is SPEC-003's to pin (self-sufficient: names file, offending key, full six-key set); ADR-010 left wording as human-reviewed residue.
- D-5: SPEC-003 frontmatter follows the SPEC-002 exemplar (no `tags` key); the Origin paragraph carries the RFC-002-S6 provenance. The `tags: [rfc-002]` convention from the S1 conventions doc is dropped as unratified — the exemplar wins.
- D-6: No `version:` discriminator key in TemplateV3 — format told apart structurally (retired shape refused first, six-key posture otherwise); a seventh key would violate ADR-010's minimal set.
- D-7: Unknown-key detection = one shared goccy AST map-diff helper (per-key line:column, deterministic, aggregates). `yaml.Strict()` alone rejected: single nondeterministic key, no aggregation, inline-map panic at v1.19.2 (upstream #810/#849). Feasibility: gapfill-libs.md verdict.
- D-8: `name` AND `schema` are mandatory at load/write (schema-less V3 template refused like missing name). Design call — ADR-010 fixes the key set, not mandatory-ness; recorded in SPEC REQ-006.
- D-9: SPEC-003 draft is 371 lines vs docs-writer's ≤250 default. Kept: SPEC-002 (the live exemplar) is ~345 lines; the KB `spec` template is the enforced contract and has no line gate. Noted for owner.
- D-4: Un-routing downstream consumers (search columns, re-marshal, list --json, index/log title reads) move in S6's implementation as one mechanical cohort — the change doesn't compile/behave half-done. SPEC-003 lists them under Affected Surface, not as separate phases.

## Open questions
(none blocking)

## Synthesis & draft
- 2026-10-09: SPEC-003 drafted at .pi/research/rfc-002-s6/SPEC-003-templatev3-format-loader.md (371 lines; 18 REQ, 19 AC). All spec.yaml mechanical gates verified locally; validated GREEN against the live `spec` template in a throwaway KB copy (exit 0). Awaiting owner review before the KB filing wave.
- Filing wave plan (on owner grant): file 4 research reports byte-faithful (research/rfc-002-s6-{prior-research,code-recon,rfc-extract,gapfill-libs}.md, status final, as_of 2026-10-09, run rfc-002-s6, informs [SPEC-003], --no-commit), then LOG as research/rfc-002-s6-log.md (hub, wikilinks), then SPEC-003 (no --no-commit... per S1 convention owner runs the one plain-git commit; all writes --no-commit), then akb index add + akb log, then akb lint.

## Parked/deferred
(none yet)

## Filed pages (KB, 2026-10-09 wave)
- [[rfc-002-s6-prior-research]] — what was already ratified; the open gaps
- [[rfc-002-s6-code-recon]] — TemplateV2 loader/key inventory/un-routing blast radius
- [[rfc-002-s6-rfc-extract]] — RFC-002 S6 requirements, boundaries, SPEC format contract
- [[rfc-002-s6-gapfill-libs]] — goccy v1.19.2 / santhosh v6.0.3 feasibility verdicts
- [[SPEC-003-templatev3-format-loader|SPEC-003]] — the artifact (status: proposed; human flips to active)

## Child-run index
| Child | Agent | Scope | Output | Status |
|---|---|---|---|---|
| prior-research | scout | KB prior research + ADR-008/009/010 extraction; decided vs open | .pi/research/rfc-002-s6/prior-research.md | launched |
| code-recon | scout | Current TemplateV2 format/loader/type-title routing/embedded templates/tests | .pi/research/rfc-002-s6/code-recon.md | launched |
| rfc-extract | scout | RFC-002 §5–7/§17/§18/Appendix A + ADR-010 + SPEC-002 exemplar + spec template | .pi/research/rfc-002-s6/rfc-extract.md | DONE, triaged |
| gapfill-libs | researcher | goccy/go-yaml strict/unknown-key API + santhosh v6 compile surface feasibility | .pi/research/rfc-002-s6/gapfill-libs.md | DONE, triaged |
