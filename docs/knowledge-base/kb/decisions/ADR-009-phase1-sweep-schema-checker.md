---
created: "2026-10-09"
deciders:
- Pete Hope
grammar: 1
id: ADR-009
is_draft: false
provenance: agent-drafted
revisit:
- the Phase-3 lint config registry is descoped or delayed — duplicate missing-field reporting persists and this interim needs re-examination
- a downstream typed gate is verified to tolerate key loss in by_check — removal of required_fields could accelerate
- the schema checker's per-page degradation masks a systematic fault class — fault-isolation depth is section-11 Phase-A scope, not this record
scope:
- internal/lint/**
- cmd/akb/**
- internal/template/**
status: accepted
summary: A built-in schema lint checker ships with Phase 1 (before cel_lint, one shared now); required_fields stays registered until the Phase-3 registry subsumes it, so by_check never loses a key.
tags:
- rfc-002
- phase-1
- lint
- json-schema
- validation
title: Phase 1 Ships the Sweep-Time Schema Checker and Keeps required_fields
type: adr
updated: "2026-10-09"
---

# ADR-009: Phase 1 Ships the Sweep-Time Schema Checker and Keeps required_fields

> In the context of RFC-002 Phase 1's validation core, facing [[ADR-003-santhosh-v6-format-assertion|ADR-003]]'s blocking verification gate that presumes a sweep-time schema checker whose placement §17 never names, and a `required_fields` lint checker §11 declares subsumed whose removal would drop a key from the frozen `lint --json` envelope, we decided that Phase 1 ships the sweep-time schema checker as a built-in and retains `required_fields` until the Phase-3 config registry subsumes it, and neglected deferring the sweep checker to Phase 3 and removing `required_fields` in Phase 1, to achieve complete P7 sweep coverage from Phase 1 with no envelope break, accepting two phases of duplicate missing-field reports from two checkers, because a safety net with a two-phase hole contradicts the ratified degrade-on-sweep contract and a lost `by_check` key breaks downstream typed gates.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

Phase 1 adds a built-in `schema` lint checker that validates every page's document view against its template's compiled JSON Schema at sweep time, running before `cel_lint` under one shared `now`, and the existing `required_fields` lint checker stays registered until the Phase-3 lint config registry provides the keep-keys subsumption mechanism, because the sweep is the safety net for every page that bypasses `akb write` and the frozen envelope may never lose a key.

## Invariants

- **I1**: WHEN `akb lint` runs, the system MUST include a built-in checker named `schema` that validates each page's document view against its page type's compiled template schema, with format assertion per ADR-003 I2/I3.
- **I2**: WHEN a sweep runs, the schema checker MUST run before `cel_lint`, and the schema pass and the CEL lint pass MUST share one `now` stamp.
- **I3**: WHEN the schema checker cannot evaluate a page, the system MUST degrade to a per-page issue and MUST continue the sweep.
- **I4**: The schema checker MUST emit `LintIssue` rows; the sweep MUST NOT emit the [[SPEC-002-structured-validation-report|SPEC-002]] write-time report or its row shape.
- **I5**: The `required_fields` lint checker MUST remain registered and reporting until the Phase-3 lint config registry lands; `akb lint --json` output MUST keep the `by_check["required_fields"]` key for as long as the checker exists, and any later removal MUST preserve the key at 0 through the registry's keep-keys mechanism.
- **I6**: The Phase-1 removal of the write-path required-field pre-gate (RFC-002 §7 folds it into the schema pass) MUST NOT extend to the sweep; the lint `required_fields` checker remains per I5.

## Negative Constraints

- **N1** (MUST NOT · scope: `cmd/akb/**`, `internal/lint/**`): The system MUST NOT remove or disable the `required_fields` lint checker before the Phase-3 config registry provides the keep-keys disable mechanism; the checker MUST stay registered so the frozen `lint --json` envelope never loses a `by_check` key.
- **N2** (MUST NOT · scope: `internal/lint/**`): The schema checker MUST NOT share the SPEC-002 report row type; it MUST emit `LintIssue` rows, because the write-time locators (schema keyword + JSON Pointer) differ from sweep locators (page path) and N5 freezes `LintIssue` independently.
- **N3** (MUST NOT · scope: `internal/lint/**`): The schema checker MUST NOT abort the sweep on a single page's evaluation fault; the system MUST emit a per-page issue naming the page and continue, per the degrade-on-sweep contract.

## Exceptions

No exceptions are permitted. A KB that finds the duplicate missing-field reports intolerable may note so for Phase-3 planning, but no Phase-1 or Phase-2 mechanism may silence either checker — the config registry is the first sanctioned disable surface.

## Verification

- **I1**: testscript — a page written to disk bypassing `akb write` (direct file creation) whose content violates its template's schema is flagged by `akb lint` as a `schema` issue · gate: `go test ./test/` · mode: **block** once implemented. This is ADR-003's own sweep verification gate; both records close with one check.
- **I2**: integration test asserting checker execution order and a single shared `now` across the schema and `cel_lint` passes · gate: `go test ./internal/lint/ ./cmd/akb/` · mode: **block** once implemented.
- **I3, N3**: testscript — a page the schema checker cannot evaluate produces a per-page issue and the sweep completes and reports the remaining pages · gate: `go test ./test/` · mode: **block** once implemented.
- **I4, N2**: review — the schema checker's implementation emits only `lint.LintIssue`; `rg -n 'ValidationReport|report' internal/lint/` finds no import of the write-time report type · mode: **advisory**.
- **I5, N1**: testscript — `akb lint --json` output includes the `by_check["required_fields"]` key · gate: `go test ./test/` · mode: **block**; `rg -n 'required_fields' cmd/akb/lint.go` shows the registration line present · mode: **block**.
- **Human-only residue**: whether the duplicate missing-field reports confuse KB owners is judgment; the `revisit` tripwires are the watched signal, and Phase-3 planning hears the complaints first.

## Context

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A20, researched and owner-ratified 2026-10-09 via decision form in the research run ([[rfc-002-phase1-scope-log|run LOG]]). The RFC left a placement gap: §17's Phase-1 list omits the sweep-time schema checker, while ADR-003's verification matrix blocks on one ("the sweep-time schema checker flags the same page"), and §11's disposition table declares `required_fields` "subsumed by the sweep-time schema checker" without naming a phase. Recon showed the built-in cost is trivial — the `LintChecker` interface is `Name()` plus `Check(ctx, *KB)`, registration is one `AddChecker` line at `cmd/akb/lint.go:157-167`, and the `cel.go:60-62` pattern already demonstrates rebuilding the page view per page inside a checker ([[rfc-002-phase1-scope-code-surface|code-surface recon]]). The freeze argument is §11's own: disabled checkers keep their `by_check` key at 0 so the envelope "never loses keys (load-bearing for downstream typed gates)" — removing `required_fields` in Phase 1, before the registry's keep-keys mechanism exists, would be the only key loss in the envelope's life, against a consumer the RFC's risk table flags as cited-secondhand and therefore unverifiable. The duplicate-report overlap is real but visible and temporary: both checkers already share identical message wording, so a missing field reads as the same sentence from two named sources.

## Decision Drivers

- P7 completeness: the sweep is the safety net for pages that bypass write (git pull, direct edits, ingestion); a two-phase hole contradicts the ratified contract.
- ADR-003 verifiability: its sweep gate is blocking and must be closable in the phase that ships the schema pass.
- Envelope freeze (N5): additive changes only; key loss is a contract break.
- P3 tension is acceptable when it is visible, documented, and scheduled for resolution — the Phase-3 registry is the resolution.
- Cost asymmetry: the checker is one struct plus one registration line; the envelope break it avoids is unpriced.

## Alternatives Considered

- **Defer the sweep schema checker to Phase 3** — rejected: ADR-003's blocking sweep gate cannot close, and pages bypassing `akb write` carry zero schema scrutiny for two phases — precisely the hole P4/P7 exists to close. Do not re-propose unless P7 is amended to accept a write-only schema layer.
- **Ship the sweep checker in Phase 1 and remove `required_fields` immediately** — rejected: P3-clean and literally what §11's "subsumed" says, but `by_check` loses a key with no keep-keys mechanism yet, against an unverifiable downstream consumer. Do not re-propose unless the registry's keep-keys mechanism ships earlier than Phase 3.

## Consequences

- Good, because ADR-003 becomes fully verifiable in Phase 1 and the P7 safety net covers bypassed pages from the day the schema pass ships.
- Good, because the `lint --json` envelope evolves additively only — the `schema` key appears, no key disappears — and downstream typed gates never break.
- Good, because the interim overlap is honest: identical message wording across both checkers, this record documenting the duplication as deliberate, and a named phase that ends it.
- Bad, because a page missing a required field is flagged twice per sweep for two phases — a visible P3 tension accepted as the price of envelope stability.
- Bad, because the schema checker is a built-in special case for two phases (users cannot disable it until the registry), a privilege §11 intends to dissolve; accepted as sequencing, not principle.
- Neutral, because the write/append/approve required-field pre-gate still retires per RFC-002 §7 — approve's refusal-wording parity then follows the schema pass's `required` message, which the schema-pass tech-spec must preserve (SPEC-002's approve-parity preserved behavior).

## References

- Research run: [[rfc-002-phase1-scope-log|LOG]] (decision record), [[rfc-002-phase1-scope-code-surface|code-surface recon]] (lint interface, registration site, PageData shape), [[rfc-002-phase1-scope-test-pinning|test-pinning recon]] (required-fields wording pins).
- [[RFC-002-open-validation-engine|RFC-002]] §7 (pre-gate folds into schema), §11 (disposition table, sweep ordering, shared `now`, envelope freeze), §17 (phasing — amended on this ratification), §18 (candidate lifecycle).
- [[ADR-003-santhosh-v6-format-assertion|ADR-003]] I2/I3 and its sweep verification gate — the presumption this record makes true.
- [[SPEC-002-structured-validation-report|SPEC-002]] — the report/LintIssue boundary (its non-goal is this record's I4).
- Code anchors at ratification (`a316f0a`): `internal/lint/engine.go:37-40` (checker interface), `cmd/akb/lint.go:157-167` (registration), `internal/lint/cel.go:60-62` (per-page rebuild pattern), `internal/lint/required_fields.go` (the retained checker).
