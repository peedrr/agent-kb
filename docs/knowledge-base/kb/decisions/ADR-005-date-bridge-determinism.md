---
created: "2026-10-06"
grammar: 1
id: ADR-005
provenance: agent-drafted
revisit:
- cel-go changes duration() or timestamp() string acceptance (re-check at every cel-go upgrade alongside ADR-004's revisit)
- santhosh-tekuri/jsonschema changes its built-in duration checker or RegisterFormat override semantics
- a template needs ISO 8601 month/year duration arithmetic in CEL — route to RFC-002-A5's registry decision
- the permissive applicator traversal coerces a value a template author did not intend (restrict traversal or widen the static check)
- the CEL env gains typed frontmatter declarations — re-evaluate the syntactic-only static check
scope:
- internal/**
- cmd/akb/**
- internal/skill/embedded/**
status: proposed
summary: The date bridge coerces only schema-declared temporal fields, identically for page/old_page; template write warns on unbacked temporal calls; the duration seam is documented, bridging deferred to A5.
tags:
- cel
- json-schema
- date-bridge
- determinism
- template-authoring
- rfc-002
title: The Date Bridge Coerces Only Schema-Declared Fields and Is Statically Checked
type: adr
updated: "2026-10-06T21:32:47Z"
---

# ADR-005: The Date Bridge Coerces Only Schema-Declared Fields and Is Statically Checked

> In the context of RFC-002's open validation engine replacing akb's silent, name-agnostic date coercion, facing the need for a bridge whose coercion set, acceptance profile, and failure behavior are specified exactly, we decided that a static traversal of the template's schema builds a temporal-coercion plan over frontmatter — applied identically to page and old_page, gated by the ADR-003 strict profiles, falling back to the raw string on parse failure — and that `akb template write` syntactically warns on `timestamp()`/`duration()` calls the schema cannot back, and neglected annotation-driven match-based coercion, unconditional-path-only traversal, shipping `iso_duration()` here, and error-severity static checks, to achieve a bridge whose output is a pure function of schema plus bytes, accepting warnings an author can ignore and coercion of parseable strings under non-matching applicator branches, because a temporal field is what the schema declares, and the fail-closed schema gate makes coercion failure unreachable where a declaration applies.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

The CEL page map's temporal coercion becomes a deterministic bridge: akb statically traverses the template's `schema:` block to build a coercion plan of frontmatter paths declared `format: date-time` or `format: date`, coerces exactly those string values that parse under the ADR-003 strict profiles, applies the identical plan to `page` and `old_page`, and leaves every other value untouched, in memory only. `akb template write` gains a syntactic static check that warns — never fails — when a rule calls `timestamp()` on a frontmatter path the schema does not declare temporal, or `duration()` on a path the schema declares `format: duration`. The ISO 8601 versus Go duration vocabulary mismatch is documented in the kb-management SKILL and its bridge function is deferred to RFC-002-A5, because declaration, not inference, must define what CEL sees as a timestamp.

## Invariants

- **I1**: WHEN the system builds a CEL `page` or `old_page` map, the system MUST coerce a frontmatter value only when the template's `schema:` block declares `format: date-time` or `format: date` at a schema path that statically reaches the value's frontmatter path.
- **I2**: The system MUST derive the coercion plan by static traversal of the schema's `frontmatter` subtree, descending the subschema keywords `properties`, `patternProperties`, `additionalProperties`, `items`, `prefixItems`, `contains`, `allOf`, `anyOf`, `oneOf`, `if`, `then`, and `else`.
- **I3**: WHEN the traversal encounters a local `$ref`, the system MUST resolve it into the referenced subschema; the system MUST terminate reference cycles through a visited set keyed by schema path.
- **I4**: WHEN coercing a value at a `format: date-time` path, the system MUST accept exactly the strict profile of [[ADR-003-santhosh-v6-format-assertion|ADR-003]] I4, using the same checker function the schema layer registers.
- **I5**: WHEN coercing a value at a `format: date` path, the system MUST accept exactly the RFC 3339 `full-date` profile of ADR-003 I5, using the same checker function the schema layer registers.
- **I6**: WHEN a value at a declared temporal path fails the strict-profile parse, the system MUST leave the raw string in the page map unchanged.
- **I7**: WHEN coercing a `format: date` value, the system MUST produce midnight UTC as the time-of-day component.
- **I8**: The system MUST apply one identical coercion plan when building the `page` map and the `old_page` map for a write or append, with `old_page` built from a fresh on-disk read under the current template's declarations.
- **I9**: WHEN `akb template write` compiles a validation rule or a lint rule, the system MUST statically inspect the compiled CEL AST for `timestamp` and `duration` calls whose argument is a `page.frontmatter.*` or `old_page.frontmatter.*` attribute path.
- **I10**: WHEN the static inspection finds a `timestamp()` call on a frontmatter path where the schema declares no `format: date-time` or `format: date` assertion, the system MUST emit a stderr warning naming the rule ID, the attribute path, and the remediation.
- **I11**: WHEN the static inspection finds a `duration()` call on a frontmatter path where the schema declares `format: duration`, the system MUST emit a stderr warning naming the rule ID, the attribute path, the vocabulary mismatch (the schema asserts ISO 8601 per RFC 3339 Appendix A while CEL `duration()` parses Go `time.ParseDuration` syntax), and the remediations: assert Go-style syntax with a `pattern` in place of `format: duration`, or remove the `duration()` call from the rule.
- **I12**: WHEN the argument of a `timestamp()` or `duration()` call is not a statically resolvable frontmatter attribute path, the system MUST NOT emit a static-check warning for that call.
- **I13**: The kb-management SKILL MUST document the duration vocabulary seam: JSON Schema `format: duration` asserts the RFC 3339 Appendix A `duration` ABNF, CEL `duration()` accepts Go `time.ParseDuration` syntax only, and no akb bridge converts between the two.

## Negative Constraints

- **N1** (MUST NOT · scope: `internal/cel/**`): The system MUST NOT register a temporal declaration found inside a `not` subschema into the coercion plan; the system MUST traverse only the keywords enumerated in I2, whose subschemas assert against instance values.
- **N2** (MUST NOT · scope: `internal/cel/**`): The system MUST NOT coerce a frontmatter value whose path carries no temporal declaration; an undeclared value MUST reach CEL as its raw YAML-parsed value.
- **N3** (MUST NOT · scope: `internal/**`, `cmd/akb/**`): The system MUST NOT persist a coerced value to disk or expose one in the document view (`read --json`, schema-validation input); coerced values MUST exist only in the in-memory `page` and `old_page` maps (RFC-002 P6).
- **N4** (MUST NOT · scope: `internal/cel/**`): The system MUST NOT define the bridge's acceptance set by raw `time.Parse(time.RFC3339, …)` or by any profile other than the ADR-003-registered checkers; the bridge MUST reuse those checker functions so assertion acceptance and bridge acceptance are identical by construction.
- **N5** (MUST NOT · scope: `internal/**`, `cmd/akb/**`, `internal/skill/**`): The system MUST NOT ship an `iso_duration()` function or any ISO-to-Go duration conversion under this record; duration bridging MUST be decided in RFC-002-A5's closed-registry record.
- **N6** (MUST NOT · scope: `cmd/akb/**`): The static temporal check MUST NOT fail `akb template write` or block a template from being written; findings MUST be emitted as stderr warnings only, per the RFC-002 §5.2 "→ warning" scope and the ADR-003 I7 unknown-format precedent.
- **N7** (MUST NOT · scope: `internal/cel/**`): The static check MUST NOT base its verdicts on CEL type information, because frontmatter values type as `dyn`; the check MUST be syntactic over the compiled AST's call and select expressions.
- **N8** (MUST NOT · scope: `internal/cel/**`): The implementation of this record MUST NOT change the CEL program-cache key or the single-environment shape pinned by [[ADR-004-cel-go-upgrade-pinned-extensions|ADR-004]] N4; the AST-inspection helper MUST be additive to `internal/cel`.

## Exceptions

No exceptions are permitted. A case that appears to need one — including a template that wants match-driven coercion under applicators — is a proposal to supersede this record.

## Verification

- **I1, I6, N2**: unit-test matrix over plan application — declared nested path coerced, declared array items each coerced, undeclared sibling untouched, parse-failure leaves the raw string, `format: time` value untouched · gate: `go test ./internal/cel/` · mode: **block** once implemented · remediation: make plan application match the matrix; never widen coercion to match a failing test.
- **I2, I3, N1**: unit tests enumerating traversal coverage — every I2 keyword, a recursive `$ref` terminating through the cycle guard, a `not`-subtree declaration absent from the plan · gate: `go test ./internal/cel/` · mode: **block** once implemented.
- **I4, I5, N4**: unit tests asserting the bridge and the schema format checkers are the same functions, plus the ADR-003 acceptance matrix executed through the bridge · gate: `go test ./internal/cel/` · mode: **block** once implemented.
- **I7**: unit test asserting a coerced `format: date` value equals midnight UTC · gate: `go test ./internal/cel/` · mode: **block** once implemented.
- **I8**: unit test asserting `page` and `old_page` receive identical coercions under one template · gate: `go test ./internal/cel/` · mode: **block** once implemented.
- **N3**: testscript case: after a write with declared temporal fields, the on-disk bytes and `akb read --json` show the raw strings · gate: `go test ./test/` · mode: **block** once implemented.
- **I9–I12, N6**: testscript matrix — `timestamp(page.frontmatter.x)` with `x` undeclared → exit 0 plus a stderr warning naming rule and path; `x` declared → silent; `old_page.` variant → warning; `duration(page.frontmatter.d)` with `format: duration` → seam warning naming rule, path, and both remediations; non-path argument → silent · gate: `go test ./test/` · mode: **block** once implemented.
- **I13**: human check at the Phase-2 doctrine update that the SKILL carries the seam paragraph · mode: **advisory**.
- **N5**: `rg -n 'iso_duration' internal/ cmd/` returns zero hits · mode: **block** until RFC-002-A5 ratifies a registry.
- **N8**: review confirms `internal/cel/engine.go` cache-key and environment construction are unchanged in the implementing PR · mode: **advisory**.
- **Human-only residue**: whether permissive traversal semantics under applicators surprise template authors is judgment — watched through the `revisit` tripwires, not a mechanical check.

## Context

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-A4 (RFC §5.2 names the scope: bridge determinism spec, `template write` static temporal check, duration seam), researched 2026-10-06 in `.pi/research/rfc-002-a4/`. Today's bridge is name-agnostic, top-level-only string parsing (`internal/cel/pagebuilder.go:29-44`) whose acceptance is wider than the ADR-003 strict profile and whose nested values are never coerced — the asymmetry this record kills; every `BuildPage` call site already holds the template, so schema-driven traversal is mechanical, not architectural. Upstream facts verified against pinned sources during the research run: cel-go v0.32's `duration()` is Go `time.ParseDuration` only, with no ISO 8601 duration anywhere in core or extension libraries; santhosh-tekuri/jsonschema v6 ships a built-in `duration` checker implementing exactly the RFC 3339 Appendix A profile, so under ADR-003's assert-all policy `format: duration` already asserts with zero new code; the two vocabularies share zero string syntax, so any bridge is a translation, not a parse. The AST the static check needs is discarded by `CompileRule` today; cel-go's `common/ast` navigable API (`MatchDescendants`, `FunctionMatcher`, `SelectExpr` chains) was probe-verified at the pinned v0.28 and confirmed unchanged at v0.32. Permissive applicator traversal is the declaration-faithful reading of RFC-002 §5.2 ("a temporal field is one the template's schema declares … at whatever depth it is declared"); a declared path whose value sits under a non-matching branch simply fails the parse and stays a string (I6), which the sweep's schema checker reports per P4.

## Decision Drivers

- Determinism: the page map is a pure function of schema plus bytes — no clock, no locale, no Go-version-dependent acceptance.
- The acceptance trinity: schema assertion, bridge coercion, and CEL `timestamp()` acceptance are one profile (ADR-003 I4/I5), so no value can pass one layer and fail the next.
- P4 layering: the schema gate fails closed at write, making coercion failure unreachable where a declaration applies; the sweep degrades per page for everything else.
- P6 (input = output): coercion exists in memory only; the document view and disk bytes carry what the caller wrote.
- Boundary discipline: closed-registry contents, selection syntax, and per-function cost belong to RFC-002-A5, not to this record.
- Template-authoring posture (P2): guidance at authoring time is a warning, never a gate — the ADR-003 I7 unknown-format precedent.

## Alternatives Considered

- **Annotation-driven (match-based) coercion** — rejected: couples page-map construction to the validator's annotation output and contradicts the RFC's declaration-based definition of a temporal field. Do not re-propose unless santhosh v6 exposes a stable annotation API and permissive traversal proves surprising in practice.
- **Unconditional-path-only traversal** — rejected: a temporal field declared only inside `anyOf`/`oneOf` branches would silently never coerce — the D9 species of trap. Do not re-propose unless applicator traversal demonstrates a concrete correctness hazard.
- **Raw `time.Parse(time.RFC3339, …)` as the bridge acceptance** — rejected: wider than the ADR-003 profile and Go-version-dependent (comma fractions, single-digit hours, out-of-range offsets). Do not re-propose unless ADR-003 is superseded.
- **Ship `iso_duration()` in this record** — rejected: registry contents, selection syntax, and cost treatment are RFC-002-A5's scope, and ISO months and years have no fixed length, a calendar-semantics question A5 must own. Do not re-propose except through RFC-002-A5.
- **Static check as an authoring error** — rejected: RFC-002 §5.2 fixes the severity at warning, matching the ADR-003 I7 unknown-format precedent; template writers own their templates. Do not re-propose unless the warning proves systematically ignored.
- **Type-based static check** — rejected: frontmatter values type as `dyn` in the CEL environment, so only syntactic attribute-path analysis is sound. Do not re-propose unless the environment gains typed frontmatter declarations.

## Consequences

- Good, because the page map becomes a pure function of schema plus bytes, and assertion acceptance, bridge acceptance, and CEL `timestamp()` acceptance are identical by construction.
- Good, because temporal fields declared at any depth — nested objects and array items — are coerced uniformly, killing the probe-verified top-level/nested asymmetry.
- Good, because the two guaranteed-failure patterns (a `timestamp()` on an undeclared path, a `duration()` on an ISO-asserted path) are caught at authoring time by one AST walk, and the duration seam is documented before a template hits it at runtime.
- Bad, because permissive traversal may coerce a parseable string at a path a non-matching applicator branch declared — benign by I6's parse gate, but semantically generous, and recorded here so it reads as a tradeoff, not a defect.
- Bad, because the static check is syntactic: dynamically constructed `timestamp()` arguments receive no warning (I12's documented limitation).
- Bad, because `BuildPage` and `BuildOldPage` gain a template (or plan) parameter, rippling to five call sites — mechanical, since every caller already holds the template.
- Neutral, because `format: duration` assertion works with zero new code under ADR-003's assert-all policy, and `iso_duration()` can arrive through RFC-002-A5 without re-opening this record.
- Neutral, because an RFC-002-A5 decision ratifying `iso_duration()` adds a third remediation to I11's warning rather than retiring it: `duration()` is a CEL builtin the closed registry cannot redefine, so the vocabulary mismatch — and the warning's trigger — persists in every registry outcome.

## References

- Research run record: `.pi/research/rfc-002-a4/LOG.md` (decision log); child reports in the session's subagent artifacts (code recon, prior-research extraction, upstream facts).
- [[RFC-002-open-validation-engine|RFC-002]] §5.2 (date bridge scope), §6 (template format), §18 (candidate lifecycle).
- [[ADR-003-santhosh-v6-format-assertion|ADR-003]] — I4/I5 strict profiles, I6 `format: time` exclusion, I7 unknown-format warning precedent, assert-all policy.
- [[ADR-004-cel-go-upgrade-pinned-extensions|ADR-004]] — N4 (cache key, single environment), N6 (`convertDateField` pre-conversion is the only bridge).
- cel-go v0.32.0: `common/types/string.go` (`duration()` = `time.ParseDuration`; `timestamp()` strict gate), `common/ast` navigable API; v0.28.0 module-cache probe of the same surface.
- JSON Schema 2020-12 validation spec §7.3.1 (`format: duration` = RFC 3339 Appendix A `duration` ABNF); RFC 3339 Appendix A.
- santhosh-tekuri/jsonschema v6.0.3 `format.go` (built-in `duration` checker), `compiler.go` (`RegisterFormat` override, `AssertFormat`).
