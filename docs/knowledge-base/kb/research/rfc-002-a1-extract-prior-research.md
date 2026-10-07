---
as_of: "2026-10-06"
created: "2026-10-06"
grammar: 1
informs:
- ADR-003
is_draft: false
provenance: agent-drafted
run: rfc-002-a1-validator
scope:
- internal/cel/**
- internal/template/**
status: final
summary: Audit of the RFC-001-era research base on JSON Schema validator libraries and date/format handling, flagging verified, unverified, and stale claims.
tags:
- prior-research
- json-schema
- validation
- rfc-002
title: Prior-Research Extraction on Validator Libraries and Date/Format Handling (RFC-002-A1)
type: research
updated: "2026-10-06"
---
# Extract: prior research on JSON Schema validator libraries & date/format handling

Scope: RFC-001-era research base (`.pi/subagents/proposals/json-cel/research/`) plus the
RFCs it fed (`docs/knowledge-base/kb/rfcs/`). Every claim carries a `file:line` anchor.
`[UNVERIFIED]` = prior research flagged as assumed/not-proven; `[STALE]` = likely superseded.

> **Correction to the task brief:** `schema-survey.md` is **not** a Go-library survey. It is a
> survey of pi-subagents TypeBox/`outputSchema` consumption + a JSON-Schema→CEL gap analysis.
> The Go validator-library recommendation lives in `LOG.md`, `SYNTHESIS-openness.md`, RFC-001 §8,
> and RFC-002 §9. Read accordingly.

---

## A. Libraries named

### A.1 `github.com/santhosh-tekuri/jsonschema/v6` — RECOMMENDED (Go)
- `LOG.md:177` — "Validator library = coverage strategy (santhosh v6 recommended → ADR); Go validator makes [both layers RE2 → the ECMA-262/RE2 dialect concern dissolves in practice]."
- `SYNTHESIS-openness.md:92` — carry-over list: "JSON Schema 2020-12 + santhosh-tekuri recommendation (ADR 1)."
- `docs/.../RFC-001-json-schema-cel-coexistence.md:243-246` — "Recommended library: **`github.com/santhosh-tekuri/jsonschema/v6`** (full 2020-12, optional format assertions; Go-native). Final choice → ADR (candidates: santhosh v6, invopop/jsonschema for generation, qri-io/jsonschema)."
- `docs/.../RFC-002-open-validation-engine.md:404-406` — same recommendation; "final choice → RFC-002-A1"; asserts "Full keyword coverage… `contains`/`minContains`, `if/then/else`, `$ref`/`$defs` including cycles, `dependentRequired`/`dependentSchemas`, `unevaluated*`."
- `RFC-002:733` — Phase 1: "santhosh v6 integration". `RFC-002:775` — ADR row "RFC-002-A1 | Validator library (santhosh v6 recommended) + `format` assertion policy".
- **API details on format assertion/override: NONE in the research base.** The exact capability the
  task asks about (register/override a `date-time` checker) was never established — it is RFC-002-A1
  OQ1 (`.pi/research/rfc-002-a1-validator/LOG.md:50-52`).

### A.2 `invopop/jsonschema` — candidate, GENERATION only
- `RFC-001:246` — listed as candidate "for generation" (i.e. emit schema from Go structs), not validation. No version, no assessment.

### A.3 `qri-io/jsonschema` — candidate, no assessment
- `RFC-001:246` — named as a candidate only. No version, coverage, or verdict recorded anywhere.

### A.4 TypeBox (`typebox@1.1.38`, JS, NOT a Go option) — the incumbent validator to reach parity with
- `schema-survey.md:59-61` — pi-subagents validator is **TypeBox `Compile`**, not Ajv; dependency `typebox@1.1.38` (`pi-subagents/package.json:29`; `node_modules/typebox/`). "The old recon prose referencing `@sinclair/typebox` is stale." [STALE marker on any `@sinclair/typebox` mention]
- `schema-survey.md:64` — validation call: `validateStructuredOutputValue` → `compile(schema); validator.Check(value)`.
- `schema-survey.md:87` — TypeBox `Compile` supports JSON Schema **drafts 3→2020-12** (`typebox/readme.md:194`, coverage `readme.md:227+`); only partials: `dynamicRef`, `unevaluatedItems`/`unevaluatedProperties`.
- `schema-survey.md:147-160` in-use keyword inventory: `type`, `properties`, `required`, `items` (schema form only), `minItems`, `enum` (25×), `const` (5×), `minLength` (20×, always 1), `additionalProperties:false` (1×), `$defs`+`$ref` (15×, local only), `allOf:[{if,then}]` + `if`/`then` (7× each). No `else`, `anyOf`, `oneOf`, `not`, `pattern`, numeric bounds, tuple `items` in any agent schema.
- `schema-survey.md:141` — **transport rule**: "every schema node MUST carry explicit type" — opencode-go grammar compiler 400s on typeless nodes (incident e53cac92; `task-sizer.md:13-21`). Real production constraint.
- `schema-survey.md:96-110` — wrapper rewrite: `createStructuredOutputToolParameters` wraps the schema; local `$ref`s rewritten one level down (`structured-output.js:94`). Rejection errors bounded to 4096 bytes.
- `schema-survey.md:243-244` / `LOG.md:239-242` — typed gates call the **same** `validateStructuredOutputValue`; typed gate cannot coexist with `outputSchema` (`TYPED_VERIFY_OUTPUT_SCHEMA_CONFLICT`, `acceptance.js:174`); stdout bounded to 12000 chars.
- Relevance: **not a Go dependency**; it is the behavior/feature set the akb Go validator must match. No Go port exists.

---

## B. Coverage findings (JSON Schema 2020-12)

- `RFC-002:404-411` — akb commits to **full 2020-12 keyword coverage over the flattened document** by delegating to the library ("akb hand-rolls nothing"): `type, enum, const, required, properties, additionalProperties, items, prefixItems, contains/minContains, min/maxItems, min/maxLength, pattern, numeric bounds, multipleOf, $ref/$defs (incl. local reuse & cycles), allOf/anyOf/oneOf/not, if/then/else, dependentRequired, dependentSchemas, propertyNames, unevaluatedProperties/Items, format`.
- `RFC-001:239-251` — same list; "coverage = the validator library's 2020-12 conformance".
- **Unresolved coverage questions for A1** (`LOG.md:50-55`): confirm `contains`/`minContains`, `$ref` cycles, `unevaluated*` actually supported by santhosh v6 (the RFC asserts it; no source test was run).
- **Prior research determined no CEL→schema translation is needed for coverage** — ecosystem consensus is coexistence, not compilation: `cel-bridge-research.md:7-8`, `cel-bridge-research-b.md:8-12` (K8s KEP-2876, protovalidate keep declarative vocabulary primary; CEL reserved for cross-field).

## C. Format-assertion findings (topic)

- `RFC-001:247` — chosen library has "**optional format assertions**"; `RFC-001:249` family — "`format` (annotation-default; assertion opt-in per template — details → ADR)."
- `RFC-002:408-409` — "`format` policy (RFC-002-A1): assertion enabled **at least for declared temporal fields** (the §5.2 precondition guarantee depends on it); **default for other formats decided per RFC-002-A1**." ⇒ OQ2 for the web research.
- JSON Schema spec fact recorded: `format` is annotation-only unless enabled — `cel-bridge-research-b.md:135` ("annotation-only unless enabled, and *'MUST be disabled by default'* — turning `format` into a hard rule changes semantics"), repeated `:239-240`.
- No prior research enumerated which `format` values the library asserts or how to disable/override them.
  `cel-bridge-research-b.md:134` notes only that CEL has no stdlib email/uri/uuid/hostname/ipv4
  validators (protovalidate added custom ones) — relevant if assertion is delegated to CEL (it is not; schema layer owns it).

## D. RFC3339 / date parsing / coercion findings (topic)

- **Go-strict RFC3339 profile (binding decision):** `RFC-002:189-200` — a temporal field is one the
  schema declares `format: date-time` or `format: date`. akb's `date-time` assertion is **defined as
  the CEL acceptance set**, not full RFC 3339: RFC 3339 permits lowercase `t`/`z` and leap seconds;
  Go `time.RFC3339` (what cel-go `timestamp()` parses with) rejects both — **flagged "verified" in
  RFC-002:195, but A1 explicitly says "re-verify"** (`LOG.md:16-25`). Profile: uppercase `T`, numeric
  offset or uppercase `Z`, no leap seconds. "the validator's format checker is registered/overridden accordingly."
- **Coercion rule:** `RFC-002:201-207` — `format: date` fields coerce date-only → **midnight UTC,
  in memory only, never on disk**; `format: time` validates as a string-only format, **never coerced**
  (no CEL time type). Undeclared fields are never coerced. `old_page` gets the identical bridge.
- **Coercion is evaluation-only:** `LOG.md:236-243` (session 4 finding 2) — `BuildPage` builds a fresh
  `fmMap` (`internal/cel/pagebuilder.go:29-63`); `convertDateField` returns new values; `fm.Fields`
  untouched; disk always raw YAML. `RFC-002:183` restates defensively.
- **cel-go `timestamp()` strictness:** `cel-bridge-research.md:76` / `:133` — "`format: date-time`:
  no 'validate parse cleanly' primitive in CEL" (verdict **caveats**); only works if the value is a
  CEL timestamp (akb converts date-parseable frontmatter to `time.Time`). `RFC-002:195` names
  `time.RFC3339` as cel-go's parser.
- **`timestamp()` verified against cel-go v0.28 source** (`LOG.md:335-340`): `timestamp()` = strict
  RFC3339 (`time.Parse(time.RFC3339)`); `duration()` = **Go-style `time.ParseDuration` ("1h30m"),
  NOT ISO 8601**. JSON Schema `format: duration` (ISO 8601) ≠ CEL `duration()` vocabulary — real seam.
- **cel-go version feature note** `[STALE/unverified]`: `LOG.md:163-170` — go.mod pins **v0.28.0**;
  owner corrected the claim that it was current; latest claimed **v0.32.0**; v0.30 "`timestamp()`
  rejects non-RFC3339" listed as a feature. This is a training-data-derived claim, corrected same day,
  and the underlying release facts are **not independently sourced** in the research base.
  Confirmed current pin: `go.mod:12` = `github.com/google/cel-go v0.28.0`.
- **Round 6 date-guarantee repair:** `LOG.md:605-612` — RFC 3339 permits lowercase t/z + leap
  seconds; Go `time.RFC3339` rejects both; fix = akb date-time assertion DEFINED as Go-strict/CEL
  acceptance set, pinned ADR 1. `format: time` string-only. Also: §5.2 gotcha — `old_page` needs the
  identical bridge under the *current* template's declarations.
- **Date regime decision (D7):** `LOG.md:340-346`, `SYNTHESIS-openness.md:47` (D7 row) — schema-declared
  bridge replaces name-agnostic silent coercion; declared fields only; traversal at any depth
  (top-level-only limitation dies, `RFC-002:189-192`).
- **RE2 dialect:** `RFC-001:250-251`, `RFC-002:406-407` — a Go validator compiles `pattern` with Go
  RE2 (== CEL `matches` dialect), so the ECMA-262-vs-RE2 divergence from `cel-bridge-research.md:80`
  and `cel-bridge-research-b.md:120` "becomes a non-issue in practice: **akb is RE2 everywhere**."

## E. Performance findings

- **No benchmark of any Go JSON Schema library exists in the research base.** Performance appears only
  as (a) TypeBox parity timing (none measured) and (b) CEL cost, not schema-validator cost.
- CEL cost model (adjacent, if A1 needs it): `cel-bridge-research.md:74-77` (regex ≈ len×len; list
  literal base 40) and `cel-bridge-research-b.md:140-151` (K8s `StaticEstimatedCostLimit=10_000_000`
  per expression vs akb's 100000 cap = 100× tighter; comprehension cost formula).
- ⇒ **GAP:** A1/OQ-old "performance" requirement is unanswered by prior research.

## F. Decision history (G1 / library)

- `LOG.md:105` (session 2) — "Dialect must be pinned (recommend 2020-12). **No JSON Schema library in go.mod yet.**"
- `LOG.md:177` (session 3, RFC-001) — library chosen as the *coverage strategy*: santhosh v6 recommended.
- `LOG.md:167` — v0.29–v0.32 cel-go feature list folded into RFC §9/G2/ADR-3.
- `SYNTHESIS-openness.md:88-93` (carry-over) — G1/G2 coverage commitments + santhosh-tekuri recommendation survive into RFC-002.
- `seam-core-pipeline.md:336` and `seam-periphery.md:196` — both flag "no JSON Schema library in `go.mod`; adding one is a dependency decision for the parent design, not a peripheral one."
- `akb-recon.md:172` — confirmed no JSON Schema handling anywhere in the repo (only a prose mention in `.sisyphus/plans/akb-p5-lint.md:1103`).
- Current state: `go.mod:12` pins `cel-go v0.28.0`; **no JSON Schema library present**. `VERSION` = 0.23.0.

---

## G. What prior research did NOT answer (gaps the new web research must close)

1. **OQ1 — santhosh v6 format API:** can the built-in `date-time` checker be registered/overridden with a Go-strict RFC3339 assertion? No evidence exists; RFC-002:199 *asserts* it ("registered/overridden accordingly") without a source.
2. **OQ2 — default format policy** for non-temporal formats (assert-all / annotate-only / curated subset). RFC defers to A1 (`RFC-002:409`).
3. **OQ3 — alternatives since RFC-001 era.** Prior candidates (invopop, qri-io) never assessed; no survey of newer Go validators.
4. **Actual 2020-12 conformance verification** of santhosh v6 (the `contains/minContains`, `$ref` cycles, `unevaluated*` claims are RFC assertions, not tested).
5. **Performance**: zero benchmarks of any Go library.
6. **Maintenance status / module path & license** of santhosh v6 — required by A1's brief (`LOG.md:16-25`), absent from prior base.
7. **RFC 3339 Go-strict claim is marked "verified" in the RFC but explicitly flagged for re-verification** by A1 (`LOG.md:24-25`); no primary source is recorded.
8. **cel-go version currency** (v0.28 pin vs claimed v0.32) is a training-data claim corrected by the owner, not source-verified.

## Confidence summary

| Claim | Confidence | Basis |
|---|---|---|
| santhosh v6 recommended | high | 4 independent anchors (LOG/SYNTH/2 RFCs) |
| TypeBox is pi-subagents validator (not Ajv) | high | source-read anchors in schema-survey |
| Go-strict RFC3339 profile = design intent | high | RFC-002 §5.2 explicit |
| "RFC 3339 rejects lowercase t/z + leap sec; Go rejects too" | medium | RFC says "verified", A1 says re-verify; no primary source |
| cel-go latest = v0.32 / `timestamp()` rejects non-RFC3339 at v0.30 | low | training-data-derived, owner-corrected, unsourced |
| No Go validator benchmark exists | high | absence across all 6 files searched |
| santhosh format-override capability | unknown | never researched |
