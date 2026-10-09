---
created: "2026-10-09"
deciders:
- Pete Hope
grammar: 1
id: ADR-008
is_draft: false
provenance: agent-drafted
revisit:
- RFC-002-A5 is ratified before Phase 1 completes — re-evaluate pulling the flatten forward with its registry
- an external consumer demonstrably needs the structured write-failure report before Phase 2
- the flatten's cohort is scheduled (Phase 2 or the A6/A17 parser phase) — the I5 re-path obligation activates
scope:
- internal/cel/**
- internal/template/**
- cmd/akb/**
- docs/knowledge-base/.agent-kb/templates/**
- internal/skill/embedded/**
status: accepted
summary: Phase 1 validates today's document shape (content.raw, ast.*); the section-5 flatten, derived.* renames, word_count registry, and write --json all ship with the flatten's later cohort.
tags:
- rfc-002
- phase-1
- document-model
- validation
- json-schema
title: The RFC-002 Flattened Document Model Is Not Phase-1 Scope
type: adr
updated: "2026-10-09"
---

# ADR-008: The RFC-002 Flattened Document Model Is Not Phase-1 Scope

> In the context of RFC-002 Phase 1 adding a JSON Schema layer over akb's document view, facing §5.3's claim that the flattened document shape and its CEL renames ride the Phase-1 V3 template revision while the `word_count()` registry the flatten requires is unratified Phase-4 work, we decided that Phase 1 validates the current document shape and defers the flatten, its renames, the `word_count` registry, and `write --json` to the flatten's later cohort, and neglected flattening in Phase 1 with an A5 fast-track, an interim materialized `word_count`, CEL-native word-count rewrites, or an interim `--json` envelope, to achieve a validation core with near-zero migration churn and no premature ratifications, accepting schema paths that say `content.raw` until the flatten ships and five in-tree schema blocks that get re-pathed at flatten time, because no Phase-1 requirement needs the flatten and every alternative either violates a ratified principle or strands shipped rules.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

Phase 1 of [[RFC-002-open-validation-engine|RFC-002]] ships its validation core — the santhosh schema pass, the V3 `schema:` block, the pipeline reorder, the merged error model, the cel-go upgrade, the `template write` obligations, and the `schema.frontmatter` hard rejection — over the **current** document shape, and everything that presupposes the §5 flattened shape ships later with the flatten's cohort, because the flatten exists for Phase-2 JSON I/O and its Phase-1 cost is a ratification shortcut plus ~27 test rewrites for zero Phase-1 gain.

## Invariants

- **I1**: WHEN the Phase-1 schema pass validates a page, the system MUST validate the document in its current shape: `frontmatter` holding raw YAML values, `content` as `{raw, word_count, char_count}`, `ast` as `{headings, links, code_blocks}`, `akb` as `{provenance_markers, annotations}`, and `file` as `{path, name, dir}`.
- **I2**: WHEN a TemplateV3 `schema:` block is authored in Phase 1, its property paths and JSON Pointers MUST reference the current shape (for example `/frontmatter/title`, `content.raw`, `ast.headings`).
- **I3**: The system MUST keep the materialized `page.content.word_count` and `page.content.char_count` fields until the flatten ships.
- **I4**: Phase-1 deliverables MUST NOT include a `--json` flag on `akb write` or `akb read`; [[SPEC-002-structured-validation-report|SPEC-002]]'s stdout-emission requirements (REQ-005, AC-005–009) land with Phase 2's `write --json`, exactly as that SPEC's `depends_on` block declares.
- **I5**: WHEN the flatten ships in a later phase, the system MUST re-path every shipped `schema:` block — the two embedded templates and this repo's four KB templates — in the same change, and the release CHANGELOG MUST carry the re-path recipe.

## Negative Constraints

- **N1** (MUST NOT · scope: `internal/template/**`, `docs/knowledge-base/.agent-kb/templates/**`, `internal/skill/embedded/**`): Phase-1 template work MUST NOT author `schema:` blocks against the §5.1 flattened shape (`derived.*` paths, `content` as a plain string); templates MUST author schema blocks against the current shape per I2.
- **N2** (MUST NOT · scope: `internal/cel/**`): Phase-1 work MUST NOT rename or remove the page-map keys `content.raw`, `content.word_count`, `content.char_count`, `ast`, or `akb`; the system MUST leave the `BuildPage`/`BuildOldPage` output shape unchanged until the flatten's cohort performs the §5.3 renames as one breaking change.
- **N3** (MUST NOT · scope: `cmd/akb/**`): The system MUST NOT ship an interim `write --json` whose envelope reflects the current shape; the flag MUST arrive with the Phase-2 JSON I/O work in the RFC-final `{frontmatter, content}` envelope, because an N5-grade frozen contract may not be born with a planned breaking change.
- **N4** (MUST NOT · scope: `internal/cel/**`, `internal/template/embedded/**`, `docs/knowledge-base/.agent-kb/templates/**`): The system MUST NOT add a `word_count()` CEL function or rewrite the five `min_word_count` template rules as part of Phase 1; word counting stays materialized until RFC-002-A5's closed registry is ratified and ships with the flatten.

## Exceptions

No exceptions are permitted. A requirement that appears to need the flattened shape in Phase 1 is a proposal to supersede this record — most likely by pulling the whole flatten cohort, not a piece of it, into the phase.

## Verification

- **I1, I2**: testscript — a V3 template whose `schema:` constrains `/frontmatter/title` (via `required`) and `content.raw` (via `minLength`) validates pages at `akb write` · gate: `go test ./test/` · mode: **block** once implemented.
- **I2, N1**: `rg -n 'derived\.' internal/template/embedded docs/knowledge-base/.agent-kb/templates` returns zero hits until the flatten · mode: **block**.
- **I3, N2**: `go test ./internal/cel/` — the pagebuilder suite passes unchanged through Phase 1 · mode: **block**.
- **I4, N3**: `rg -n '"json"' cmd/akb/write.go cmd/akb/read.go` returns zero hits · mode: **block**.
- **N4**: `rg -n 'word_count\(' internal/ docs/knowledge-base/.agent-kb/templates` returns zero hits until RFC-002-A5 lands · mode: **block**.
- **I5**: human release check at the flatten's phase — the re-path and CHANGELOG recipe ship in the same change · mode: **advisory**.
- **Human-only residue**: whether the flatten's eventual cohort is Phase 2 or the A6/A17 parser phase is a scheduling judgment for those phases' specs, not mechanically checkable.

## Context

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A19, researched and owner-ratified 2026-10-09 via decision form in the research run ([[rfc-002-phase1-scope-log|run LOG]]). The RFC contradicted itself: §5.3 states the CEL renames "ride the already-breaking V3 revision" (Phase 1), while §17 places the closed function registry in Phase 4 and candidate A5 — which owns `word_count()` — is unratified; flattening in Phase 1 therefore strands the five `page.content.word_count >= 50` rules (one embedded, four in this repo's KB) with no sanctioned replacement. Recon measured the flatten's Phase-1 cost at ~27 Go test-func rewrites, 4 testscript assertions, ~30 rule renames across 8 production call sites, versus approximately zero churn for validating the current shape ([[rfc-002-phase1-scope-test-pinning|test-pinning recon]]), and confirmed the current shape is directly validatable: goccy v1.19.2 leaves date-like scalars as strings and produces only JSON-safe Go types, so no structural normalization pass is needed ([[rfc-002-phase1-scope-code-surface|code-surface recon]]). The `write --json` question is coupled, not separate: SPEC-002 REQ-005's stdout report needs the flag, and shipping it with a current-shape interim envelope would freeze a contract Phase 2 must break — so the flag defers wholesale, which is precisely the split SPEC-002's ratified `depends_on` already declares. On ratification, RFC-002 is amended per its §18 lifecycle: row A19 moves to ratified as ADR-008, §5.3's "riding the already-breaking V3 revision" is corrected, and §17 names this record as the Phase-1 scope bound.

## Decision Drivers

- Phase 1's purpose is the validation core; a shape migration buys it nothing its requirements demand.
- P5/D13 coherence: scrapping materialized `word_count` is only meaningful alongside the registry that replaces it, and A5 owns that registry.
- Envelope freeze discipline: no N5-grade surface ships with a planned break in it.
- RFC-002 is still `status: draft`: amending §5.3 is cheap; ratifying A5 under time pressure is not.
- SPEC-002's `depends_on` already phases its own acceptance criteria; this record honors that split instead of reopening a ratified SPEC.

## Alternatives Considered

- **Flatten in Phase 1 and pull `word_count()` forward** — rejected: requires ratifying an A5 slice (registry contents, selection syntax, cost treatment) before Phase 1 can start, then pays the full churn table for no Phase-1 requirement. Do not re-propose unless A5 is ratified ahead of Phase 1 for independent reasons.
- **Flatten in Phase 1, keep materialized `word_count` as a declared interim** — rejected: violates P5 (facts, not computations) and D13 in the same release that enacts them; doctrine would teach one rule while the engine ships another. Do not re-propose unless P5 itself is amended.
- **Flatten in Phase 1, rewrite the five word-count rules in plain CEL** — rejected: `size(page.content.split(' '))` is not `strings.Fields` semantics (leading, trailing, and repeated whitespace), so shipped rules silently change meaning and are replaced by the registry anyway. Do not re-propose unless the registry is cancelled outright.
- **Minimal failure-only `write --json` in Phase 1 with the RFC-final envelope** — rejected: pulls the JSON-to-YAML assembly landmine (`count: 3.0`) into Phase 1 and leaves the string-`content` envelope diverging from the `content.raw` validation document until the flatten. Do not re-propose unless the flatten itself moves into Phase 1.
- **Minimal `write --json` with an interim current-shape envelope** — rejected: an N5-grade machine contract born with a planned breaking change. Do not re-propose.

## Consequences

- Good, because Phase 1 carries no shape-migration churn — ~27 test funcs, 4 testscript assertions, and ~30 template rules stay untouched — and A5 keeps its Phase-4 deliberation space.
- Good, because SPEC-002 needs no amendment: its Phase-1-closeable acceptance criteria (AC-001, AC-004, AC-006) land with the merged error model, and the rest land exactly where its `depends_on` put them.
- Good, because the schema pass and the CEL page map share one shape throughout Phase 1, so no document/page divergence can arise ([[ADR-005-date-bridge-determinism|ADR-005]] N3's raw-view guarantee included).
- Bad, because schema blocks authored in Phase 1 reference `content.raw` and `ast.*` and must be re-pathed when the flatten ships; mitigated by I5's same-change recipe and a corpus of six in-tree templates.
- Bad, because agents cannot consume the machine-readable write-failure report until Phase 2's `write --json`; accepted, since no agent harness currently drives `akb write` programmatically.
- Neutral, because `char_count`'s zero-consumer scrap (D13) also defers: a known-dead field ships for one more phase rather than alone.

## References

- Research run: [[rfc-002-phase1-scope-log|LOG]] (decision record), [[rfc-002-phase1-scope-code-surface|code-surface recon]] (page-map shape, rule consumers, loader), [[rfc-002-phase1-scope-test-pinning|test-pinning recon]] (churn table, frontmatter types, BuildPage call sites).
- [[RFC-002-open-validation-engine|RFC-002]] §5 (document model), §5.3 (renames — amended on this ratification), §8.2/§15 (`write --json` surface), §17 (phasing), §18 (candidate lifecycle).
- [[SPEC-002-structured-validation-report|SPEC-002]] — REQ-005/AC-005–009 and the `depends_on` split this record honors.
- [[ADR-005-date-bridge-determinism|ADR-005]] N3 — coerced values never enter the document view; the current-shape document stays raw.
- Code anchors at ratification (`a316f0a`): `internal/cel/pagebuilder.go:51,91` (the two chokepoints), `cmd/akb/write.go:69-73` (no `--json` today).
