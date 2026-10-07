---
grammar: 1
type: research
title: RFC-002-A2 schema.frontmatter Retirement Research Run Log
summary: "Run log for RFC-002-A2: owner-ratified forks on schema.frontmatter retirement and migration, constraint anchors, and the trail behind ADR-006."
status: final
provenance: agent-drafted
created: "2026-10-06"
updated: "2026-10-06"
as_of: "2026-10-06"
run: rfc-002-a2
informs:
- ADR-006
scope:
- "internal/template/**"
- "cmd/akb/**"
tags:
- research-log
- templates
- migration
- rfc-002
---
# Research run: RFC-002-A2 — `schema.frontmatter` retirement + migration

- **Question:** Resolve RFC-002 candidate A2 — how `schema.frontmatter` (the homegrown
  TemplateV2 block) is retired under TemplateV3, and whether/how akb ships a migration
  path (`akb template migrate`?). Deliverable: a real ADR (next free global number)
  per docs-writer skill, written via `akb` (AKB_KB set), `Origin: RFC-002 candidate RFC-002-A2`.
- **Date:** 2026-10-06 (started)
- **Run parameters:** MODE=start; SUBJECT=rfc-002-a2; owner instructions: consume RFC-002 in
  full; A1→ADR-003 and A4→ADR-005 contain pending decisions that constrain this record;
  subagents per protocol; auto-resolvable → draft ADR; owner-input forks → owner-escalation.

## Agreed requirements (owner-ratified 2026-10-06)

- **F1**: NO automated `akb template migrate` command. Hard-reject + documented manual recipe.
  (Owner interview; recommended option ratified.)
- **F2 (refined after owner challenge)**: temporal-target detection = ADR-005's ratified
  `template write` static warnings during the author's edit loop; the static recipe carries
  doctrine only (mapping table + 'every timestamp()/duration() target needs a format
  declaration' + MANDATORY corpus check for date vs date-time). NO new detection machinery.
  Owner clarification: the recipe must NOT live in the SKILL (one-time doctrine, standing
  context cost) — ratified via follow-up question.
- **F3**: Recipe home = CHANGELOG of the shipping release (full recipe) + compact
  self-sufficient rejection message (mapping essence + CHANGELOG pointer); SKILL documents
  V3 only, never migration. (Owner ratified; recommended option.)
- Scope: templates + mockups only; existing pages never rewritten — sweep per RFC-002 §19.
- Rejection posture: clean break, no alias; unify the two divergent detectOldFormat copies
  into one helper/message.

## Constraints inherited from ratified records (anchors)

- RFC-002 §6: TemplateV3 format; `schema.frontmatter` retired with RFC-001 hard-rejection
  migration path (precedent: `detectOldFormat`); guard invariant: schema `required` set under
  `properties.frontmatter` = exactly the set of keys CEL rules may read unguarded; unknown
  template keys hard-rejected at `template write`, warn-only at load; mockup obligations
  incl. `<!-- FAILS: schema: /pointer -->` convention extension.
- RFC-002 §19 migration risk row: V3 hard rejection w/ migration message; existing pages'
  `created`/`updated` stay on disk as data; guarded temporal rules degrade to vacuous until
  templates declare.
- ADR-003 (A1): santhosh v6, AssertFormat global (I2/I3); strict I4/I5 date-time/date
  checkers; unknown format names warn at `template write` (I7); no per-template opt-outs (N4).
  → migration that emits `format: date-time` declarations opts old fields into assertion
  (strict profile) — old pages with lax forms would fail sweep; must surface.
- ADR-005 (A4): bridge coerces only schema-declared temporal fields (I1); permissive
  applicator traversal (I2/N1); `template write` syntactic static check warns on
  `timestamp()`/`duration()` calls the schema can't back (I9–I12, warning only per N6);
  duration seam documented, `iso_duration()` deferred to A5 (N5/I11/I13).
  → migration must detect CEL rules calling `timestamp()`/`duration()` on frontmatter paths
  and decide declarations (or warn) — same static analysis ADR-005 specifies.

## Decisions made

(None yet.)

## Key findings

From `recon-code.md` (scout 1, full anchors there):

- TemplateV2 `Schema.Frontmatter map[string]FieldSchema` — only `Required` is ever read;
  `Type`/`Enum` are dead weight (template.go:14-55). Loader validates name/dup/old-format only;
  non-strict unmarshal confirmed at template.go:105-116 (unknown keys silently ignored, both
  load-time and `template write`-time — templates_write.go:100).
- `detectOldFormat` precedent: TWO divergent copies (template.go:73-82 capital/period vs
  templates_write.go:461-467 lower/semicolon); rejects top-level `required`/`optional`/`body`;
  plain error → exit 1; v1→v2 was hard break, NO converter, forward-pointing message only
  (.sisyphus plan notes confirm deliberate no-migration-tool decision).
- Required-field gate: cmd/akb/required_fields.go:16-35, called from write/append/approve/
  templates_write; exit 1, all missing listed sorted. Sweep twin: internal/lint/required_fields.go:83.
- type/title routing (frontmatter.Parse:61-72) is the widest blast radius: special-cases in
  required_fields.go:38-49, templates_write.go:399-447, injection in cel.BuildPage:60-66.
- In-tree TemplateV2 inventory: exactly 5 YAMLs (embedded adr+note; KB adr/rfc/spec).
  ALL 5 call `timestamp()` on `created`/`updated` declared as plain `type: string` → all break
  or warn under a schema-declared bridge. Sharpest case: KB spec.yaml `verified.at` /
  `verified.next_review_by` are nested keys inside a `map` field — caught only incidentally by
  today's heuristic bridge (pagebuilder.go:38-46); JSON Schema CAN declare nested formats.
- `FAILS:` convention is NOT machine-parsed anywhere (0 .go hits) — documentary only;
  RFC-002's `<!-- FAILS: schema: /pointer -->` extension requires a NEW parser.
- No migration machinery exists (grep `migrate` = 0). Hook point: new cmd/akb/template_migrate.go
  self-registering on templateCmd (template.go:26-56), reusing the temp+rename+commit swap
  (templates_write.go:298-351).
- `template get --full` serializes tmpl.Schema verbatim (cmd/akb/template.go:174) — V3 schema
  block must flow through it.
- testscript precedent: test/testdata/cel_old_format.txt asserts both message spellings; no Go
  unit test for detectOldFormat (testscript only).

## Key findings (prior-research extraction)

From `extract-prior-research.md` (scout 2, anchors there):

- Retirement rationale settled: `schema.frontmatter` is decorative — Type/Enum never enforced,
  only `.Required` live at 3 sites (seam-core-pipeline.md:259; FIX-SHAPES-2026-09-07.md:73).
  `schema:` key is REUSED with new meaning (Template.Schema `yaml:"schema"`), not added →
  hard-reject is the only honest path (RFC-001:198-200).
- Guard invariant: required set under properties.frontmatter = exactly keys CEL may read
  unguarded; three consumers follow (write gate, lint mirror, mockup gate)
  (seam-core-pipeline.md:263-270).
- v1→v2 was a DELIBERATE hard break, no tool, no dual format — but premise was "not in the
  wild, not production" (.sisyphus/drafts/cel-validation-plan.md:232-240). Premise weaker now:
  this repo's own KB runs 3 TemplateV2 templates in active use.
- `akb template migrate` exists ONLY as a bracketed question (RFC-001:368; RFC-002:792). No
  source argues either side. The fork is genuinely open.
- Phasing conflict resolved by RFC-002 itself: RFC-001 put retirement in Phase 3; RFC-002 §17
  puts `schema.frontmatter` hard rejection in Phase 1. RFC-002 supersedes → Phase 1.
- A3/A4 defer NOTHING to A2. A1 hands forward the `schema:` collision + strict-write/
  lenient-load landing. ADR-003/ADR-005 constraints as recorded below.
- Enum duplication seam: adr.yaml hand-mirrors dead `.Enum` into a CEL `in [...]` rule —
  migration must NOT duplicate enum into both layers (seam-periphery.md:191).
- CEL field-selector caveat: JSON Schema `properties` allows keys that aren't CEL-identifier-
  safe; templates must keep keys CEL-safe (seam-core-pipeline.md:271-273).
- No prior source proposes page-content migration — only sweep-degrade policy. Existing-page
  disposition is policy-only.
- NEW corpus fact (from reading KB pages): existing pages mix date shapes — `created:` is
  date-only ("2026-09-27") while `updated:` is RFC3339 date-time. Under ADR-003 assert-all,
  declaring `created: {format: date-time}` makes every existing page fail the sweep. The
  corpus-correct declarations are `format: date` for created, `format: date-time` for updated
  — a judgment an automated tool cannot safely make without sampling pages.

## Deliverable — RATIFIED 2026-10-06

- **ADR-006** `kb/decisions/ADR-006-schema-frontmatter-retirement.md` — status `accepted`,
  `deciders: [Pete Hope]`, approved (`is_draft: false`). Ratification edits: the
  ratify-dependent consequence line (old line 106) rewritten to record the completed §18
  amendment. Commits: f5d8101 (write), 496d0fb (approve).
- **RFC-002 amended per §18** (commit eab6708): A2 row → `ratified` resolving as
  [[ADR-006-schema-frontmatter-retirement|ADR-006]]; §6 retirement bullet now names ADR-006;
  `updated` bumped to 2026-10-06T22:47:27Z. Candidate ID retained in the spawns frontmatter
  list and the row's ID column, matching the A1/A4 precedent. RUN COMPLETE.

## Decisions made (coordinator, low-stakes, logged)

- D-a: Phasing = Phase 1 (RFC-002 §17 supersedes RFC-001 §16 Phase-3 placement).
- D-b: `<!-- FAILS: schema: /pointer -->` MACHINE-PARSING is Phase-1 template-write machinery,
  not A2 scope (today's convention is documentary-only; parser is new code either way).
  A2 only requires the migration path to preserve/translate the documentary comments.
- D-c: type/title ordinary-keys mechanics (frontmatter.Parse un-routing, BuildPage injection
  removal) belong to the V3 implementation, not A2; A2's migration mapping must only specify
  how type/title appear in the emitted schema (const type, minLength title).

## Open questions (owner forks)

- F1: Ship `akb template migrate` (automated, scaffold+report, or report-only) vs.
  hard-reject + documented manual recipe? (The candidate's titular question; v1→v2 premise
  "not in the wild" no longer fully holds.)
- F2: Date-field handling in migration: tool/recipe auto-detects timestamp()/duration()
  targets via ADR-005 static analysis and emits format declarations (author verifies), vs.
  fully author-declared with the recipe + ADR-005 static warnings as the net? Corpus mixed
  date shapes make blind `date-time` emission actively harmful.
- F3: Scope confirmation: migration covers templates (+mockups) only; existing pages are
  dispositioned by sweep + (optionally) a report, never rewritten. §19-aligned.
- F4: Rejection posture: clean break + unified message (fix the two divergent detectOldFormat
  copies) vs. one-release alias. Lean clean break per v1→v2 precedent + pre-1.0 status.

## Parked/deferred items

- `<!-- FAILS: schema: /pointer -->` parser — Phase-1 template-write work (D-b).
- frontmatter.Parse un-routing / BuildPage injection removal — V3 implementation (D-c).
- Embedded adr/note rewrite as V3 showcases — Phase-1 implementation; A2 decides mechanics
  only. This repo's KB templates (adr/rfc/spec) = dogfood corpus for whatever path F1 picks.
- CEL-identifier-safe key linting — flag for Phase-1 template write validation (from
  seam-core-pipeline.md:271-273); not A2's decision.
- Enum duplication doctrine (don't state enum in both layers) — belongs to G6 SKILL guidance,
  Phase 4; migration mapping just emits enum into schema and drops the dead FieldSchema.Enum.

## Parked/deferred items

(None yet.)

## Child-run index

- scout fad1cf92 (code-recon) — COMPLETE → `.pi/research/rfc-002-a2/recon-code.md` (high
  confidence, all sections anchored).
- scout b5f52928 (prior-research-extract) — COMPLETE →
  `.pi/research/rfc-002-a2/extract-prior-research.md` (342 lines, per-source anchors + gaps list).
