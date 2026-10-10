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
- docs/knowledge-base/**
status: final
summary: Spec page template contract, SPEC-003 anatomy to clone, SPEC-002 report-row contract quoted, SPEC-003 to SPEC-004 obligations, and research filing conventions.
tags:
- rfc-002
- spec-004
- conventions
title: RFC-002-S7 spec-conventions — spec page contract, house anatomy, SPEC-002 report contract
type: research
updated: "2026-10-10"
---
# Code Context — SPEC-004 authoring conventions (RFC-002-S7)

Read-only extraction. All paths repo-root-relative. Anchors `file:line`. Prior convention
extract: `docs/knowledge-base/kb/research/rfc-002-s1-adr-conventions.md` (cited as S1@line);
this file extracts only what is new/different for SPEC-004.

## Files Retrieved
1. `docs/knowledge-base/.agent-kb/templates/spec.yaml` (whole; rules quoted below) — the loader-enforced spec-page contract akb write will enforce.
2. `docs/knowledge-base/kb/specs/SPEC-003-templatev3-format-loader.md` (1-83 frontmatter; 85-399 body; headers 85-387) — house anatomy + the SPEC-003→SPEC-004 obligations.
3. `docs/knowledge-base/kb/specs/SPEC-002-structured-validation-report.md` (1-55 frontmatter; 120-169 Requirements; 153-168 report JSON; 170-210 ACs) — the report contract to map onto.
4. `docs/knowledge-base/kb/research/rfc-002-s1-adr-conventions.md` — prior run's convention extract (baseline; do not re-derive).
5. `docs/knowledge-base/.agent-kb/templates/research.yaml` (whole) — required frontmatter contract for filing a research page.
6. `docs/knowledge-base/kb/research/rfc-002-s1-log.md` (1-24) — a real filed run log; the page SPEC-002 cites via `[[rfc-002-s1-log]]`.

## 1. Spec page contract (`spec.yaml`) — what `akb write` enforces

`schema.frontmatter` required keys (rule `required: true`): `grammar` (int), `type`, `id`,
`title`, `kind`, `status`, `provenance`, `created`, `updated`, `scope` (list), `verified`
(map). Optional: `ratified_by` (list), `supersedes`, `superseded_by`, `anchors` (list).
Guard doctrine (spec.yaml:5-11): required keys read UNGUARDED; optional keys ALWAYS `has()`-guarded;
nested map keys ALWAYS guarded; `old_page.*` ALWAYS guarded; lint_rules guard everything except
`type`/`title`.

CEL validations (id: rule → meaning), quoted from spec.yaml:
- `grammar_is_one`: `page.frontmatter.grammar == 1`
- `id_format`: `page.frontmatter.id.matches("^SPEC-[0-9]{3}$")`
- `title_nonempty`: `page.frontmatter.title != ""`
- `title_style`: `page.frontmatter.title.matches("^[A-Z]") && !page.frontmatter.title.endsWith(".") && size(page.frontmatter.title) <= 80`
- `valid_kind`: `page.frontmatter.kind in ["change", "capability"]`
- `valid_status`: `page.frontmatter.status in ["draft", "proposed", "active", "completed", "superseded", "withdrawn"]`
- `valid_provenance`: `page.frontmatter.provenance in ["human", "agent-drafted"]`
- `scope_nonempty`: `size(page.frontmatter.scope) > 0`
- `active_requires_ratified_by`: `page.frontmatter.status != "active" || (has(page.frontmatter.ratified_by) && size(page.frontmatter.ratified_by) > 0)`
- `active_requires_live_verified`: `... || page.frontmatter.verified.state == "live"`
- `verified_state_enum`: `verified.state in ["live","drifted","lost","unverified"]`
- `supersedes_format`/`superseded_requires_superseded_by`: `^SPEC-[0-9]{3}$`
- `valid_state_transition`: draft→proposed; proposed→active|withdrawn|superseded; active→completed|withdrawn|superseded
- `created_immutable`, `id_immutable`, `kind_frozen_at_ratification`, `updated_not_before_created`, `temporal_created`

Body-structure / content rules (blocking unless noted):
- `challenge_affordance`: `page.content.raw.contains("this SPEC is authority only while")` — must carry the verbatim `> **For agents:**` sentence.
- Required H2s (each `page.ast.headings.exists(h, h.level == 2 && h.text == "<X>")`): `Goals and Non-Goals`, `Requirements`, `Acceptance Criteria`, `Assumptions and Open Questions`, `Affected Surface and Ordering`, `Verification Plan`, `Drift Ledger`, `Revisit Triggers`; `Current Behaviour` only when `kind == "capability"`.
- `requirements_have_ids`: `page.content.raw.matches("REQ-[0-9]{3}")`
- `acceptance_cites_requirements`: `page.content.raw.matches("AC-[0-9]{3} \\(verifies REQ-[0-9]{3}\\)")` — exact form `AC-NNN (verifies REQ-NNN)`.
- `no_unresolved_clarification`: body contains no `[NEEDS CLARIFICATION`.
- `rejected_alternatives_carry_clause`: if a `## Decisions and Rejected Alternatives` H2 exists, raw must contain `Do not re-propose unless`.
- `anchors_role_bearing`: every `anchors[]` entry carries `claim`, `path`, `symbol` (line numbers forbidden), `kind ∈ {local,flow,invariant}`.
- `min_word_count`: `page.content.word_count >= 50`.
- `code_blocks_have_language`: every fenced block declares a language.
- lint_rules (warnings): `spec_stale` (<180d), `proposed_too_long` (90d), `active_signal_not_live` (error), `lost_anchor_blocks` (error), `drifted_anchor_reports`, `no_external_citations` (warn if body contains `.pi/` or `.sisyphus/`).

Note: spec.yaml does NOT literally enforce EARS wording; EARS is a house convention (REQ-NNN
`WHEN … MUST …` statements), enforced only by the `requirements_have_ids` / AC-citation rules.
`kind`, `status`, `provenance`, `scope`, `verified` are mandatory frontmatter; a WRITE of
SPEC-004 will be refused at exit 1 (collect all failures) if any required key/rule fails.

## 2. House anatomy (SPEC-003)

Frontmatter block is SPEC-003:1-83 (verbatim below, condensed keys preserved):
```yaml
anchors: [ {checked_at, claim: REQ-NNN, evidence, kind, path, state, symbol, verify:{check,method}} ]
created: "2026-10-09"
grammar: 1
id: SPEC-003
is_draft: false
kind: change
provenance: agent-drafted
ratified_by: [Pete Hope]
scope: [internal/template/**, internal/frontmatter/**, internal/cel/**, internal/lint/**, ...]
status: active
title: Templates Adopt a Strict Six-Key V3 Format With a JSON Schema Block
type: spec
updated: "2026-10-10T17:15:23Z"
verified: {at, branch, commit, method: manual-review, next_review_by, state: live}
```
(`is_draft` is auto-managed; all other keys follow spec.yaml.)

Section outline (H2 headers only), SPEC-003:85-387:
`# SPEC-NNN: <title>`; `> **For agents:**` blockquote; `Origin:` paragraph; `## Why`;
`## Goals and Non-Goals`; `## Current Behaviour` (change SPECs still carry it when recon-
relevant); `## Requirements` (137); `## Acceptance Criteria` (199); `## Contract and
Invariants` (280); `## Preserved Behaviour` (293); `## Decisions and Rejected Alternatives`
(305); `## Assumptions and Open Questions` (324); `## Affected Surface and Ordering` (335,
closing with bare `depends_on:`/`blocks:` lines); `## Verification Plan` (361); `## Drift
Ledger` (372); `## Revisit Triggers` (380); `## Evidence Appendix` (387).

Conventions:
- REQ-NNN statements: `REQ-003: WHEN <trigger>, the <subject> MUST <observable response>`; aggregate rules stated as plain `The … MUST …`.
- AC-NNN lines: `AC-NNN (verifies REQ-NNN): WHEN … MUST …` then indented `verify: { method: check|ask, check|ask: "…" }` (SPEC-003:199-278).
- EARS-2119 style: WHEN/WHEN-trigger + MUST; all-caps normative keywords.
- Rejected-alternatives pattern (SPEC-003:312-313): "Rejected: compile-at-load. Do not re-propose unless SPEC-004's design proves pass-through unworkable." (spec.yaml `rejected_alternatives_carry_clause` requires the phrase `Do not re-propose unless`.)
- Linkage: bare YAML block at end of `## Affected Surface and Ordering` — SPEC-003:358 `depends_on: [ADR-006, ADR-008, ADR-010]`, :359 `blocks: [SPEC-004, SPEC-005]`. SPEC-002 uses list form with prose (SPEC-002:283-289). SPEC-004 should declare `depends_on` naming SPEC-003 (its loader pass-through) and `blocks` as applicable.
- Status lifecycle used: SPEC-003 `active` with `ratified_by: [Pete Hope]` + `verified.state: live`; drafting status is `proposed`; SPEC-002 `active` (SPEC-002:20-25). SPEC-004 should start `proposed`.
- Origin paragraph slot: immediately after the blockquote, before `## Why` (SPEC-003:86-97; SPEC-002:60-70) — there is NO `## Context` H2 in the spec template (S1§5).

## 3. SPEC-002 report contract (mapping target) — quoted

Report JSON shape (SPEC-002:153-168):
```json
{
  "valid": false,
  "path": "notes/foo.md",
  "errors": [
    {"source": "cel", "rule_id": "title_nonempty", "message": "frontmatter title must be non-empty", "severity": "error"},
    {"source": "cel", "rule_id": "min_word_count", "message": "body must hold at least 50 words", "severity": "error"}
  ],
  "summary": {"schema": 0, "cel": 2, "total": 2}
}
```
Prose (SPEC-002:169-172): "`path` is the KB-relative page path (the stripped-`relPath`
convention). Schema rows carry `keyword` and `pointer` instead of `rule_id`. A write-time
report is single-source by REQ-002 (schema failure skips CEL); the `summary` keeps both count
fields so the envelope never changes shape between failure modes."

REQ-002 (SPEC-002:126-128): "WHEN the schema pass fails for a page, the system MUST NOT
evaluate that page's CEL validations; the system MUST return a report whose rows come from the
schema pass alone." → short-circuit; a schema-pass report contains no CEL row.

REQ-003 (SPEC-002:129-131): "WHEN the schema pass fails, the system MUST report each violation
with the failing schema keyword, the JSON Pointer of the offending instance location, and a
single-line message." → schema row fields = `keyword` + `pointer` + `message`.

REQ-005 (SPEC-002:137-139): "WHEN `akb write --json` fails validation, the system MUST write
the report to stdout as one JSON object and MUST exit with code 1."

REQ-006 (SPEC-002:140-141): "The system MUST order report rows in evaluation order: schema rows
before CEL rows, and CEL rows in the template's rule declaration order." → ordering invariant
for the error-tree mapping.

REQ-008 (SPEC-002:145-146): "The report MUST NOT attach a `line` field to schema-violation
rows; a CEL row MUST omit `line` unless the CEL layer supplied a nonzero line." → schema rows
carry NO `line`.

Row shape summary (relevant to SPEC-004 mapping): schema row = `{source:"schema", keyword,
pointer, message, severity:"error"}` (no `rule_id`, no `line`); CEL row = `{source:"cel",
rule_id, message, severity:"error"}` (line only if nonzero). Envelope = `{valid, path, errors[],
summary{schema,cel,total}}`; empty collections `[]` never `null` (SPEC-002:227-228).
AC-003 (SPEC-002:181-183) pins an example schema row: missing `frontmatter.title` → keyword
`required`, pointer `/frontmatter`. AC-002 (SPEC-002:176-180) pins short-circuit:
schema-failing page → schema rows only, no CEL row.

## 4. SPEC-003 → SPEC-004 interface (obligations to honor)

- REQ-007 (SPEC-003:160-161): "The Phase-1 loader MUST NOT compile or evaluate the `schema`
  block; it MUST preserve the parsed payload uncompiled for the schema pass (SPEC-004)."
- AC-007 (SPEC-003:220-222): "WHEN the Phase-1 diff is inspected, `internal/template/` imports
  no JSON Schema compiler; `go.mod` gains no validator dependency. verify: `go list -deps
  ./internal/template | rg jsonschema; expect 0 hits`." → SPEC-004 owns the compiler/cache.
- SPEC-003:107: "the schema *pass* that consumes the block is SPEC-004".
- SPEC-003:119 (non-goal): "Compiling, caching, or evaluating the `schema` block (SPEC-004 / ADR-009)."
- SPEC-003:288: "any closure [of unknown keywords inside `schema`] rides SPEC-004's compiler integration."
- SPEC-003:311-313: "keeping all santhosh v6 coupling inside SPEC-004's compiled-schema cache
  design. Rejected: compile-at-load. Do not re-propose unless SPEC-004's design proves
  pass-through unworkable."
- SPEC-003:314-316: format-only conversion; "Do not re-propose unless the schema pass
  (SPEC-004) has landed."
- SPEC-003:359: `blocks: [SPEC-004, SPEC-005]`.
- SPEC-003:385 (Revisit Trigger): "SPEC-004 ratification — re-check REQ-005/REQ-007 boundary text."

## 5. Filing conventions (research pages)

`research.yaml` required frontmatter (rule `required: true` + validations):
`grammar`(==1), `type`, `title` (≤120 chars), `status ∈ {draft,final,superseded}`, `provenance
∈ {human,agent-drafted}`, `run` (non-empty, e.g. `rfc-002-s7`; the grouping key), `as_of`
(≤ now), `created`, `updated`, `scope` (non-empty), `tags` (≥1, each `^[a-z0-9-]+$`),
`summary` (non-empty, ≤200 chars). Optional guarded: `informs` (each `^(ADR|SPEC|RFC)-[0-9]{3}$`),
`supersedes`, `superseded_by` (`*.md`). Lifecycle draft→final|superseded, final→superseded;
`as_of` frozen once final; `min_word_count ≥ 50`; `code_blocks_have_language` is warning-only.
lint_rules: `research_draft_stale` (90d), `final_without_informs` (a final page SHOULD set
`informs` to the decision records it fed).

Path naming: `docs/knowledge-base/kb/research/<run-slug>-<topic>.md` (tracked). Examples:
`rfc-002-s1-adr-conventions.md`, `rfc-002-s1-log.md`, `rfc-002-s6-gapfill-libs.md`.
IMPORTANT: SPEC-003/spec.yaml `no_external_citations` lint rule warns on any `.pi/` or
`.sisyphus/` citation in a spec page — SPEC-004 must cite research by its TRACKED path
(`kb/research/<file>.md`) / wikilink, never `.pi/research/...`.

Wikilink citation style (SPEC-002/003 use target-only or `target|label`):
- `[[rfc-002-s1-log]]` (SPEC-002:64), `[[rfc-002-s6-log]]` (SPEC-003:97,375),
  `[[rfc-002-s6-gapfill-libs]]`, `[[rfc-002-s6-code-recon]]` (SPEC-003:398-404).
- `[[RFC-002-open-validation-engine|RFC-002]]` (SPEC-002:60, SPEC-003:88).
- `[[ADR-003-santhosh-v6-format-assertion|ADR-003]]` (SPEC-002:224, etc.).
- SPEC-002/003 `Evidence Appendix` lists governing decisions as wikilinks (SPEC-002:337-343;
  SPEC-003:398-409) and ends with a `Discard log:` bullet list.

## Architecture
SPEC page = frontmatter contract (spec.yaml) + body contract (required H2s, REQ/AC id shapes,
challenge affordance, rejected-alternatives clause). SPEC-003 is the loader half; SPEC-004 is
the schema-pass half that consumes the uncompiled `schema` payload and emits SPEC-002's
report rows. Research pages (`research.yaml`, tracked under `kb/research/`) are the evidence
layer cited by wikilink; the run's own log is `<run>-log.md` and the S7 conventions page is
`rfc-002-s7-spec-conventions.md`-style.

## Start Here
Open `docs/knowledge-base/kb/specs/SPEC-003-templatev3-format-loader.md` and clone its
frontmatter/body shape (1-83 frontmatter; 85-387 body) for SPEC-004; enforce
`docs/knowledge-base/.agent-kb/templates/spec.yaml` before writing; take the report-row
contract from `SPEC-002-structured-validation-report.md:126-168`; cite evidence as tracked
`kb/research/` wikilinks.

## Confidence + Gaps
Confidence: high — spec.yaml, SPEC-002, SPEC-003, research.yaml all read directly; quotes
anchored. Gaps: (a) `verified.commit`/`ratified_by`/`next_review_by` values for SPEC-004 are
runtime/owner inputs, not derivable here; (b) RFC-002 §7 §15 §18 body text not re-read in
this pass — confirm the S4 row and report section when authoring; (c) the exact SPEC-004
frontmatter `scope`/`status` must be chosen by the coordinator.
