---
created: "2026-10-06"
deciders:
- Pete Hope
grammar: 1
id: ADR-004
is_draft: false
provenance: agent-drafted
revisit:
- cel-go publishes a release newer than v0.32.0 — re-check timestamp() acceptance, the cost model, and the pinned extension versions before any bump
- the issue
- akb begins ingesting templates from untrusted sources — set cel.RegexProgramSizeLimit
- a rule needs a capability behind OptionalTypes, ext.Regex, TwoVarComprehensions, Encoders, or Native — draft that capability's own ADR
- cel-go deprecates cel.CostLimit or changes interpreter.EvalCancelledError
scope:
- internal/**
- cmd/akb/**
- go.mod
- go.sum
status: accepted
summary: cel-go upgrades to cel.dev/cel-go v0.32.0 in one require+import rewrite; five ext libraries enabled in NewEnv pinned at their highest v0.32.0 versions; regex plan-size knob stays unbounded.
tags:
- cel
- dependency-upgrade
- extensions
- validation
- rfc-002
title: CEL Runs on cel.dev/cel-go v0.32 with a Pinned Extension Set
type: adr
updated: "2026-10-07T19:00:53Z"
---

# ADR-004: CEL Runs on cel.dev/cel-go v0.32 with a Pinned Extension Set

> In the context of RFC-002 Phase 1's validation core, facing a cel-go pin four releases behind, exposed to two published advisories, and a breaking module-path migration at v0.32.0, we decided to upgrade to `cel.dev/cel-go` v0.32.0 in a single require-plus-import rewrite and to enable `ext.Strings`, `ext.Lists`, `ext.Sets`, `ext.Math`, and `ext.Bindings` in `NewEnv` pinned at the highest versions shipping in v0.32.0, and neglected staying on v0.28, unversioned library-default extension enablement, the remaining extension families, the v0.31 `RegexProgramSizeLimit` knob, and v0.32's `ParseTimestamp`/`NativeToValue` helpers, to achieve current upstream support whose rule semantics no dependency bump can silently alter, accepting a seven-file import rewrite and a deliberate extension-version review at every future cel-go bump, because a fail-closed validation gate may change its verdicts only by a reviewed act.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

Upgrade cel-go from `github.com/google/cel-go` v0.28.0 to `cel.dev/cel-go` v0.32.0 — one change rewriting the `go.mod` require and every import — and enable exactly five extension libraries in the single shared environment constructor, each pinned at the highest version shipping in v0.32.0 (Strings v5, Lists v2, Math v2; Sets and Bindings are unversioned), because the pinned v0.28.0 is four releases behind and exposed to two advisories fixed upstream, and a validation engine's rule semantics must change only when someone reviews the change.

## Invariants

- **I1**: The system MUST name `cel.dev/cel-go` as the cel-go module path in `go.mod` and in every Go import statement.
- **I2**: The system MUST pin cel-go at a release in the v0.32.x line.
- **I3**: WHEN a cel-go upgrade beyond the pinned line is proposed, the system MUST re-verify `timestamp()` string acceptance and runtime cost behavior against the ADR-003 I4 profile and the behavioral cost tests before the bump merges.
- **I4**: The system MUST enable exactly the extension libraries `ext.Strings`, `ext.Lists`, `ext.Sets`, `ext.Math`, and `ext.Bindings` in `NewEnv` (`internal/cel/engine.go`) for every CEL environment.
- **I5**: WHEN enabling an extension library that publishes numbered versions, the system MUST pass the explicit version option naming the highest version shipping in the pinned cel-go release (`StringsVersion(5)`, `ListsVersion(2)`, `MathVersion(2)` at v0.32.0).
- **I6**: The system MUST keep exactly one CEL environment shape for all templates — the five libraries are enabled globally, never per template.

## Negative Constraints

- **N1** (MUST NOT · scope: `go.mod`, `go.sum`, `internal/**`, `cmd/akb/**`, `test/**`): The system MUST NOT reference `github.com/google/cel-go` in any require line or import statement; the system MUST reference `cel.dev/cel-go` everywhere.
- **N2** (MUST NOT · scope: `internal/cel/**`): WHEN an extension library publishes numbered versions, the system MUST NOT enable it via the no-option default — the default installs every future version on each dependency bump; the system MUST pass the explicit pinned version of I5.
- **N3** (MUST NOT · scope: `internal/cel/**`, `cmd/akb/**`): The system MUST NOT enable extension families beyond the five of I4 — including `ext.Regex`, `cel.OptionalTypes`, `ext.TwoVarComprehensions`, `ext.Encoders`, and `ext.NativeTypes`; a new capability MUST arrive through its own ADR per RFC-002 §10's deferral list.
- **N4** (MUST NOT · scope: `internal/cel/**`): The A3 upgrade MUST NOT change the program-cache key and MUST NOT introduce per-template environment variation; the `(env identity, expression)` cache key and function selection are RFC-002-A5's scope, and this upgrade keeps one environment shape.
- **N5** (MUST NOT · scope: `internal/cel/**`): The system MUST NOT set `cel.RegexProgramSizeLimit`; the system MUST rely on the 100000 runtime cost bound, RE2's linear-time guarantee, and cel-go v0.30's compile-time expression caps (100k code points, 100k nodes, depth 250).
- **N6** (MUST NOT · scope: `internal/**`): The system MUST NOT define date-time acceptance by cel-go's string-to-timestamp gate or adopt v0.32's `types.ParseTimestamp` helper; the ADR-003 I4 schema checkers remain the sole acceptance definition, and `convertDateField`'s Go-stdlib pre-conversion stays the only bridge.

## Exceptions

No exceptions are permitted. A case that appears to need one — including a rule that wants a sixth extension family — is a proposal for that capability's own ADR, not a waiver of this record.

## Verification

- **I1, N1**: `rg -n 'github.com/google/cel-go' --type go go.mod` returns zero hits · gate: `go build ./...` at the upgrade commit · mode: **block** · remediation: rewrite the import to `cel.dev/cel-go`.
- **I2**: `go list -m cel.dev/cel-go` prints a `v0.32.x` version · gate: upgrade commit · mode: **block**.
- **I3**: ADR-003's date-time test matrix and the `internal/cel` cost tests pass on the new version · gate: `go test ./...` at every cel-go bump · mode: **block**.
- **I4**: unit tests compile and evaluate one representative expression per library — `split`/`join` (Strings), `sets.contains` (Sets), `math.greatest` (Math), `distinct` (Lists), a `cel.bind` macro (Bindings) · gate: `go test ./internal/cel/` · mode: **block**.
- **I5, N2**: `rg -n 'ext\.(Strings|Lists|Math)\(\)' internal/cel/` returns zero bare calls; the exact option symbols are verified to exist at the pinned release — if a symbol is absent, the implementation records the fallback (the no-option default equals the pinned version today) in the upgrade PR · mode: **block** at the upgrade commit, **advisory** thereafter.
- **I6, N4**: human-only — the single-environment shape lives in one constructor; reviewers check `NewEnv` at PR time.
- **N3**: `rg -n 'ext\.(Regex|TwoVarComprehensions|Encoders|NativeTypes)|OptionalTypes' internal/ cmd/` returns zero hits · mode: **block**.
- **N5**: `rg -n 'RegexProgramSizeLimit' --type go` returns zero hits · mode: **block**.
- **N6**: covered by ADR-003's verification matrix; additionally a unit test asserts `timestamp("2025-01-17T01:00:00,001Z")` (comma fraction) errors under v0.32 · gate: `go test ./internal/cel/` · mode: **block**.
- **Human-only residue**: whether a future extension-version bump's new functions are worth opting into, and whether upstream cost-model drift (the issue #1476 class) matters to akb's rules, are judgment calls exercised at each cel-go upgrade; the `revisit` tripwires are the watched signals.

## Context

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A3 (RFC §10), researched 2026-10-06 in the research session ([[rfc-002-a3-log]]); both decision forks were owner-ratified that day via interview. Ratified 2026-10-06 by Pete Hope on review of the research record, and per the RFC-002 §18 lifecycle the RFC's A3 row moves to ratified as ADR-004 and its body references are rewritten on that ratification. akb pins cel-go v0.28.0 — four releases behind v0.32.0 (2026-08-19) and exposed to CVE-2026-83530 and GHSA-gcjh-h69q-9w9g, both fixed by v0.30. v0.32.0 moved the module to `cel.dev/cel-go` (PR #1413); the old path is a read-only alias shim whose last release is v0.31.0, so the require line and all imports must change together — a version-only bump fails with a module-identity error. akb's cel-go surface is small (seven files, four subpackages): `cel.CostLimit` is doc-identical at v0.28 and v0.32 (no `ActualCostLimit` exists anywhere in the v0.32 API), the panic-recovery contract is unchanged, and the one compile-break surface across the four releases (v0.29's `InterpretableV2`) touches only custom `Interpretable` implementors, which akb is not. The v0.30 `timestamp()` strictness gate (PR #1338) does not disturb [[ADR-003-santhosh-v6-format-assertion|ADR-003]]: I4's profile is a verified subset of both versions' acceptance sets, and akb pre-converts declared temporal fields to `time.Time` before CEL sees them, so only literal `timestamp("…")` rules ever hit the gate. Extension functions participate in cost accounting where estimators are registered — Strings registers them from its version 5, which the no-option default already selects today; I5 makes that explicit and stable. Uncosted functions count as one. Upstream's v0.32 cost-number drift (issue #1476) touches `optional.*` macro estimates only; akb enables no optional types and its cost test is behavioral (a cubic rule over 2048 headings must abort inside the limit), not numeric.

## Decision Drivers

- Fail-closed write-gate determinism: rule semantics change only by reviewed acts — this drove pinned extension versions (owner-ratified).
- Security posture: the v0.28 pin is exposed to two advisories fixed upstream by v0.30.
- The ADR-003 interlock: date-time acceptance must be identical before and after the upgrade, with no page re-validation.
- Minimal blast radius: every cel-go API akb calls is stable across the four releases; the migration is mechanical.
- Scope discipline: the program-cache key and per-template function selection belong to RFC-002-A5; the remaining extension families are deferred in RFC-002 §10.
- Threat-model honesty: regex plan-size hardening guards against untrusted rule authors, which akb does not have — pages are data and are never compiled (owner-ratified declination).

## Alternatives Considered

- **Stay on v0.28** — rejected: exposed to two advisories fixed upstream, four releases behind, and Phase 1 needs the extension libraries. Do not re-propose unless the v0.32 line demonstrates a regression against akb's API surface.
- **Unversioned (library-default) extension enablement** — rejected by the owner 2026-10-06: every dependency bump would auto-opt into future extension versions, changing write-gate semantics without review. Do not re-propose unless pinning demonstrably blocks a needed function that upstream ships only under a new version.
- **Pin Strings only** — rejected: partial determinism; behavior-versioned changes in Math or Lists would still drift. Do not re-propose unless pinning all five proves a real maintenance burden in practice.
- **Set `cel.RegexProgramSizeLimit`** — rejected by the owner 2026-10-06: the knob guards a threat model akb does not have (rule authors are KB owners; page content is never compiled), and its bounds duplicate the runtime cost limit and v0.30's parser caps. Do not re-propose unless akb ingests templates from untrusted sources.
- **Adopt v0.32's `types.ParseTimestamp` or `NativeToValue` JSON support** — rejected: `convertDateField` is two stdlib calls and the page builder constructs plain maps; the helpers add coupling for no gain. Do not re-propose unless the document builder (RFC-002-A6) adopts CEL-native types.
- **Enable the remaining extension families now (Regex, OptionalTypes, TwoVarComprehensions, Encoders, Native)** — rejected: RFC-002 §10's deferral list; each materially changes the rule-authoring surface. Do not re-propose except as a per-capability ADR with a concrete rule that needs it.
- **Wait for a post-#1476 tag (v0.32.1 or later)** — rejected: the drift touches `optional.*` estimates only; akb enables no optional types and its cost test is behavioral. Do not re-propose unless the upgrade's own test run shows real cost fallout.

## Consequences

- Good, because the dependency is current, both advisories are cleared, and template authors gain string, list, set, math, and binding functions whose costs are registered with the interpreter's accounting.
- Good, because ADR-003 survives untouched — the I4 subset argument was verified against v0.32 source, so no existing page needs re-validation.
- Good, because extension semantics are pinned: a future `go get -u` cannot silently change what an existing rule means.
- Bad, because the upgrade is a flag-day rewrite of seven files plus `go.mod`/`go.sum`; any open branch touching cel imports will conflict.
- Bad, because every future cel-go bump now carries a deliberate extension-version review step — the price of N2's determinism.
- Bad, because tests pinning cel-go-owned error strings (`no such key: …`, the `overload` substring) may need message updates on the bump; the upgrade PR must distinguish message churn from behavior change.
- Neutral, because `cel.dev/expr` stays at v0.25.1 and `antlr4-go/antlr/v4` remains an indirect dependency at v0.32.0.
- Neutral, because `github.com/google/cel-go` persists as a read-only alias shim, so any transitive dependency still on the old path links safely.

## References

- Research run: [[rfc-002-a3-log|LOG.md]] (decision log) and [[rfc-002-a3-web-cel-go-releases|web-cel-go-releases.md]] (source-verified upstream findings); code-recon and prior-research extraction reports in the session's subagent artifacts.
- cel-go releases v0.29.0–v0.32.0 and PR #1413 (module switch): `https://github.com/cel-expr/cel-go/releases/tag/v0.32.0` and siblings; PR #1338 (`timestamp()` gate); `common/types/timestamp.go`, `common/types/string.go`, `cel/options.go`, `cel/program.go`, `ext/strings.go`, `ext/README.md` at tag v0.32.0.
- cel-go issues #1476 (v0.32 cost drift) and #1329 (repository move); advisories CVE-2026-83530 and GHSA-gcjh-h69q-9w9g.
- [[ADR-003-santhosh-v6-format-assertion|ADR-003]] — the date-time profile this upgrade must not disturb; its cel-go-upgrade revisit trigger is discharged by this record.
- [[RFC-002-open-validation-engine|RFC-002]] §10 (upgrade, extension set, and the deferred extension families), §17 Phase 1, §18 (candidate lifecycle); [[RFC-001-json-schema-cel-coexistence|RFC-001]] §9 (release analysis — historical research record only).
