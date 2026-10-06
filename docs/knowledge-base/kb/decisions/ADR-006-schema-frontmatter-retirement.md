---
created: "2026-10-06"
deciders:
- Pete Hope
grammar: 1
id: ADR-006
provenance: agent-drafted
revisit:
- external KBs running TemplateV2 templates appear in the wild before Phase 1 ships (revisit the no-command decision)
- the rejection message or CHANGELOG recipe proves insufficient in practice (author confusion or support burden)
- a future template-format break re-evaluates alias posture under post-1.0 stability expectations
- the frontmatter-key detection predicate false-positives on a real V3 schema
scope:
- internal/template/**
- cmd/akb/**
- internal/skill/embedded/**
- CHANGELOG.md
status: accepted
summary: TemplateV3 hard-rejects the retired schema.frontmatter block with one self-sufficient message; migration is a CHANGELOG recipe driven by ADR-005 warnings — no command, no alias, no page rewriting.
tags:
- template-format
- migration
- json-schema
- validation
- rfc-002
title: TemplateV3 Retires schema.frontmatter — Migration Is a Recipe, Not a Command
type: adr
updated: "2026-10-06T22:49:48Z"
---

# ADR-006: TemplateV3 Retires schema.frontmatter — Migration Is a Recipe, Not a Command

> In the context of RFC-002's TemplateV3 replacing the homegrown `schema.frontmatter` block with a JSON Schema 2020-12 document, facing the choice of migration machinery for a pre-1.0 tool whose known TemplateV2 corpus is five templates, we decided that akb hard-rejects the retired block through one shared helper with one self-sufficient message and ships the migration recipe in the releasing CHANGELOG — no migrate command, no alias — and neglected an automated `akb template migrate` scaffold, a report-only checker, page-sampling format inference, a one-release alias, and SKILL-resident migration doctrine, to achieve a retirement whose only lasting surfaces are the rejection message and a historical record, accepting hand-migration by every TemplateV2 author, because temporal-target detection is already shipped by ADR-005's template-write warnings, date declarations are corpus-dependent judgments no tool can make unattended, and one-time doctrine must not tax every future agent session's context.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

TemplateV3 refuses any template whose `schema` block carries the retired homegrown `schema.frontmatter` shape — at load and at `akb template write`, through one shared detection helper and one self-sufficient message — and ships the complete migration recipe in the CHANGELOG of the release that introduces the refusal, with [[ADR-005-date-bridge-determinism|ADR-005]]'s template-write temporal warnings as the author's detection loop, because the known TemplateV2 corpus is five templates, the corpus-dependent date-declaration judgment cannot be automated safely, and one-time migration doctrine must not become a permanent command surface or a permanent resident of agent context.

## Invariants

- **I1**: WHEN the template loader reads a template file whose top-level `schema` mapping contains a `frontmatter` key, the system MUST refuse the template and exit with code 1.
- **I2**: WHEN `akb template write` reads a template file whose top-level `schema` mapping contains a `frontmatter` key, the system MUST refuse the template and exit with code 1.
- **I3**: The system MUST produce the I1 and I2 refusals from a single shared detection helper emitting a single message text.
- **I4**: The refusal message MUST name the template file, the retired construct (`schema.frontmatter`), the four-step migration essence — presence moves to schema `required`; `type`/`enum` translate to JSON Schema keywords; temporal fields gain author-verified `format` declarations; the author re-runs `akb template write` until no temporal warnings remain — and the CHANGELOG as the recipe's home.
- **I5**: WHEN a template file mixes the retired block with V3-shaped keys, the system MUST apply the I1/I2 refusal before any V3 interpretation.
- **I6**: WHEN the release that introduces the V3 refusal ships, its CHANGELOG section MUST carry the full migration recipe: the TemplateV2→V3 mapping table, the rule that an enum lives in exactly one validation layer, the temporal-declaration doctrine with its corpus-check mandate, and the mockup obligations of RFC-002 §6.
- **I7**: WHEN an author migrates a template, temporal-target detection MUST be the ADR-005 static check at `akb template write`.

## Negative Constraints

- **N1** (MUST NOT · scope: `cmd/akb/**`): The system MUST NOT ship `akb template migrate` or any command that rewrites template files into TemplateV3; the system MUST refuse retired templates per I1/I2 and direct the author to the CHANGELOG recipe.
- **N2** (MUST NOT · scope: `internal/template/**`, `cmd/akb/**`): The system MUST NOT accept, alias, or silently upgrade a retired `schema.frontmatter` block on any code path, including a lenient load path; the system MUST refuse per I1/I2 — unknown *V3* keys follow RFC-002 §6's warn-at-load rule, but the retired block refuses everywhere.
- **N3** (MUST NOT · scope: `internal/**`, `cmd/akb/**`): The system MUST NOT emit `format: date`, `format: date-time`, or any other `format` declaration on a template author's behalf; the author MUST write every temporal declaration after the corpus check the recipe mandates.
- **N4** (MUST NOT · scope: `internal/**`, `cmd/akb/**`): The system MUST NOT add migration-specific template-analysis machinery (target scanners, page samplers, format inferrers); the ADR-005 static check MUST be the only temporal-target detector.
- **N5** (MUST NOT · scope: `internal/**`, `cmd/akb/**`): The system MUST NOT modify page files as part of template migration; page-level non-conformance with a migrated template MUST surface only through the sweep's per-page degradation (RFC-002 P4, §19).
- **N6** (MUST NOT · scope: `internal/skill/embedded/**`): The kb-management SKILL and its references MUST NOT carry TemplateV2→V3 migration doctrine; migration guidance MUST live in the CHANGELOG recipe and the I4 message.
- **N7** (MUST NOT · scope: `cmd/akb/**`): The implementation of this record MUST NOT add machine parsing of `<!-- FAILS: ... -->` mockup comments; the schema-pointer FAILS convention is Phase-1 `template write` machinery under RFC-002 §6.

## Exceptions

No exceptions are permitted. A future need for automated migration — demonstrated external TemplateV2 adoption — is a proposal to supersede this record, not a waiver.

## Verification

- **I1, I2, I5**: unit tests for the detection predicate — a V2-shaped template refused, a V3 schema carrying `properties` accepted, a mixed-shape file refused · gate: `go test ./internal/template/` · mode: **block** once implemented.
- **I3, I4**: testscript matrix — an old-shape template hits `akb lint`, `akb write`, and `akb template write`; each exits 1 with the identical message containing the file name, `schema.frontmatter`, the four steps, and the CHANGELOG pointer · gate: `go test ./test/` · mode: **block** once implemented · remediation: fix the message at the single helper; never fork the text per call site (the divergent v1/v2 copies, `template.go:77` vs `templates_write.go:464`, are the anti-pattern).
- **I6**: human release check that the shipping CHANGELOG section carries the full recipe · mode: **advisory** (release checklist).
- **I7, N4**: ADR-005's own verification (its I9–I12 testscript matrix) covers the detector; review confirms no migration-specific analysis code lands · mode: **advisory**.
- **N1**: `rg -ni 'migrate' cmd/akb internal/template` returns zero hits · mode: **block**.
- **N3**: review confirms no code path writes `format` declarations into template files · mode: **advisory**.
- **N5**: testscript — a page on disk is byte-identical after its template's old-shape refusal · gate: `go test ./test/` · mode: **block** once implemented.
- **N6**: `rg -ni 'migrat' internal/skill/embedded/` returns zero hits after the Phase-1 doctrine update · mode: **block**.
- **N7**: `rg -n 'FAILS' --include='*.go' .` returns zero hits until the Phase-1 `template write` work lands a parser deliberately · mode: **block** until then.
- **Human-only residue**: whether the CHANGELOG recipe is clear to a human author is judgment — watched through the `revisit` tripwires, not a mechanical check.

## Context

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A2 (RFC §6 charters the hard rejection; §18 governs this record's lifecycle), researched 2026-10-06 in `.pi/research/rfc-002-a2/`; the three material forks were owner-ratified in that session: no automated migrate command, ADR-005-warning-driven detection with an author corpus check, and CHANGELOG-as-recipe-home with SKILL silence. The v1→v2 precedent (`.sisyphus/drafts/cel-validation-plan.md:232-240`) was a deliberate hard break with no tool under the premise "akb is not in the wild"; that premise is weaker today — five TemplateV2 templates exist, three in active use in this repo's own KB — so this record buys message quality and a complete recipe rather than an alias or a converter. Two ratified records bound the migration's semantics: ADR-003's assert-all policy means any `format` declaration a migration produces opts existing pages into strict sweep validation, and the in-tree corpus mixes shapes (`created` is date-only, `updated` is date-time), so blind `format: date-time` emission would break sweeps — which is why declarations are author-authored after a corpus check (N3). ADR-005's static check already detects every `timestamp()`/`duration()` call on an undeclared path at `akb template write`, making separate detection machinery redundant (I7, N4). The detection predicate (a `schema` mapping containing a `frontmatter` key) can false-positive on a V3 schema carrying a meaningless unknown `frontmatter` keyword; rejecting it is accepted as simpler than disambiguation, and `template write`'s strict-key rule (RFC-002 §6) rejects such schemas anyway.

## Decision Drivers

- The known TemplateV2 corpus is five templates — two embedded showcases akb rewrites itself, three in this repo's KB — so automation cannot amortize its construction cost.
- Date declarations are corpus-dependent judgments; ADR-003's assert-all policy turns a wrong declaration into sweep failures over existing pages.
- ADR-005 already ships the temporal-target detector; migration-specific machinery would duplicate it.
- akb is agent-first: the refusal must be actionable without external lookup (the self-sufficient-refusal convention), and one-time doctrine must not tax every future session's context window.
- akb is pre-1.0 with an explicit breakage-acceptable posture; a clean break needs message quality, not compatibility scaffolding.

## Alternatives Considered

- **Scaffold command (`akb template migrate <name>`)** — rejected: a permanent command surface plus AST-analysis and YAML-emission machinery for a five-template corpus. Do not re-propose unless external TemplateV2 KBs demonstrably exist before Phase 1 ships.
- **Report-only checker (`migrate --check`)** — rejected: its value, the edit list, is delivered by the static mapping table plus ADR-005 warnings without a new command. Do not re-propose unless hand-migration support burden proves the checklist insufficient.
- **Page-sampling format inference** — rejected: the heaviest machinery, and it bakes legacy corpus shapes into permanent declarations (a date-time field with date-only legacy pages locks to `date`). Do not re-propose unless the scaffold command is revived; it is meaningful only there.
- **Fully manual migration without the warning-driven loop** — rejected: a missed temporal declaration degrades to vacuous temporal rules (RFC-002 §19), the D9 silent-trap species. Do not re-propose unless ADR-005's static check is retired.
- **One-release alias / dual-format read** — rejected: doubles the loader's format surface for one release and contradicts the clean-break precedent. Do not re-propose unless akb is post-1.0 with a deployed external base at a future format break.
- **SKILL-resident migration guide** — rejected: one-time doctrine taxed against every future agent session's context window. Do not re-propose unless migration support becomes recurring, which negates the premise.

## Consequences

- Good, because no new command surface or migration machinery is built or maintained, and temporal detection reuses ratified ADR-005 machinery.
- Good, because the refusal stays self-sufficient for the agent-first audience, the two divergent `detectOldFormat` message copies are replaced by one helper, and no agent session pays context cost for one-time migration doctrine.
- Good, because the corpus-check mandate (I6, N3) prevents assert-all from turning a migrated template into a sweep-failure storm over existing pages.
- Bad, because every TemplateV2 author hand-migrates; accepted for a five-template pre-1.0 corpus and watched by the first `revisit` tripwire.
- Bad, because the CHANGELOG recipe can drift from the code it describes; mitigated because the I4 message carries the four-step essence independently of the recipe.
- Bad, because the old file itself never receives ADR-005 warnings — it refuses at load, so the warnings fire on the author's in-progress V3 draft; accepted, because that is when they are actionable.
- Neutral, because `<!-- FAILS: -->` machine parsing, `type`/`title` un-routing, and the embedded adr/note rewrite are Phase-1 implementation scope outside this record, and this repo's own KB templates (adr, rfc, spec) migrate via the recipe as its dogfood.
- Neutral, because on ratification RFC-002 was amended per its §18 lifecycle: row A2 moved to ratified as ADR-006 and its body references were rewritten.

## References

- Research run record: `.pi/research/rfc-002-a2/` — `LOG.md` (decision log + owner forks), `recon-code.md` (TemplateV2 loader, in-tree inventory, precedents), `extract-prior-research.md` (audit base + deferrals).
- [[RFC-002-open-validation-engine|RFC-002]] §6 (retirement charter, guard invariant, mockup obligations), §17 (Phase-1 placement), §18 (candidate lifecycle), §19 (migration risk posture).
- [[ADR-003-santhosh-v6-format-assertion|ADR-003]] — assert-all policy (I2/I3), strict temporal profiles (I4/I5), unknown-format warning precedent (I7).
- [[ADR-005-date-bridge-determinism|ADR-005]] — the static temporal check (I9–I12) that serves as the migration detector; warning-only posture (N6).
- Precedent anchors: `internal/template/template.go:73-82` and `cmd/akb/templates_write.go:461-467` (the divergent v1 rejection copies); `test/testdata/cel_old_format.txt`; `.sisyphus/drafts/cel-validation-plan.md:232-240` (the v1→v2 hard-break decision); `.pi/subagents/proposals/json-cel/research/seam-core-pipeline.md:259,263-270` (retirement rationale + guard invariant).
