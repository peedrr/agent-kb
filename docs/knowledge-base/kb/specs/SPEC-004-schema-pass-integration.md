---
anchors:
- checked_at: "2026-10-10T18:36:00+01:00"
  claim: REQ-004
  evidence: "rfc-002-s7 code recon §1a: tail at write.go:507-547."
  kind: flow
  path: cmd/akb/write.go
  state: live
  symbol: runWrite
  verify:
    ask: Confirm the converged tail still orders selector gate, build, validations, store.Write.
    method: ask
- checked_at: "2026-10-10T18:36:00+01:00"
  claim: REQ-004
  evidence: "rfc-002-s7 code recon §1b: append.go:203-230."
  kind: flow
  path: cmd/akb/append.go
  state: live
  symbol: runAppend
  verify:
    ask: Confirm append mirrors the write tail ordering.
    method: ask
- checked_at: "2026-10-10T18:36:00+01:00"
  claim: REQ-017
  evidence: "rfc-002-s7 code recon §3: call sites write.go:523, append.go:220, approve.go:227, templates_write.go:152."
  kind: local
  path: cmd/akb/required_fields.go
  state: live
  symbol: checkRequiredFields
  verify:
    ask: Confirm the pre-gate helper and its four call sites.
    method: ask
- checked_at: "2026-10-10T18:36:00+01:00"
  claim: REQ-002
  evidence: "rfc-002-s7 code recon D6: coercion at pagebuilder.go:63; document view must not alias it."
  kind: invariant
  path: internal/cel/pagebuilder.go
  state: live
  symbol: BuildPage
  verify:
    ask: Confirm frontmatter values pass through convertDateField today.
    method: ask
- checked_at: "2026-10-10T18:36:00+01:00"
  claim: REQ-020
  evidence: "rfc-002-s7 code recon §4: engine.go:94-107; registration cmd/akb/lint.go:157-167."
  kind: local
  path: internal/lint/engine.go
  state: live
  symbol: LintEngine.AddChecker
  verify:
    ask: Confirm checkers run in registration order and any error aborts the sweep.
    method: ask
created: "2026-10-10"
grammar: 1
id: SPEC-004
is_draft: false
kind: change
provenance: agent-drafted
scope:
- cmd/akb/**
- internal/schema/**
- internal/lint/**
- internal/cel/**
- internal/template/**
- test/**
status: proposed
title: The Write Pipeline Validates Pages Against Template JSON Schemas Before CEL
type: spec
updated: "2026-10-10"
verified:
  at: "2026-10-10T18:40:00+01:00"
  branch: main
  commit: c8bac2efcaf01d0186ce1730d1f3d17f0131ac7d
  method: manual-review
  next_review_by: "2026-11-09"
  state: live
---

# SPEC-004: The Write Pipeline Validates Pages Against Template JSON Schemas Before CEL

> **For agents:** this SPEC is authority only while `status: active` AND
> `verified.state: live`. If your task conflicts with it, or any anchor fails
> to resolve at HEAD, stop and name the conflict. Do not silently work around
> it; propose an amendment or supersession instead.

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-S7 (:834), drafted from the
[[rfc-002-s7-log|rfc-002-s7 research run]]. Consumes the uncompiled `schema:` payload that
[[SPEC-003-templatev3-format-loader|SPEC-003]] REQ-007 hands through, and emits the report rows
[[SPEC-002-structured-validation-report|SPEC-002]] defines.

## Why

RFC-002 makes JSON Schema the structural validator and CEL the semantic one, but nothing today
compiles or evaluates a template's `schema:` block. S7 integrates the schema pass: it builds the
validation document, places the pass in the write/append pipeline, compiles-and-caches schemas
per ADR-003, maps santhosh error trees into SPEC-002 report rows, retires the required-fields
pre-gate, and adds the ADR-009 sweep-time checker.

## Goals and Non-Goals

Goals:
- Write/append/approve validate every page against its type's compiled JSON Schema before CEL.
- One compiled-schema cache per process, shared by the write path and the sweep checker.
- Schema failures surface as SPEC-002 report rows; sweep failures surface as LintIssues.
- The required-fields pre-gate leaves the write path; the schema's own `required` replaces it.

Non-goals:
- Compiling or caching at template load (SPEC-003 REQ-007 pass-through; compile-at-load rejected).
- Mockup-vs-schema proofs, the `FAILS:` parser, overlap warnings, ADR-005 static-check wiring (SPEC-005/S8).
- Document flattening, `derived.*`, `word_count()`, `write --json` stdout emission (ADR-008; SPEC-002 phases it).
- Date-bridge coercion of CEL values (ADR-005, Phase 2).
- Unknown-format warnings at `template write` (SPEC-005 owns the surface; this SPEC only keeps validation lenient).
- Git-state preflight ordering (merge/identity checks) — parked to ADR-011 and its post-RFC-002 work unit; this SPEC makes no preflight claims.

## Current Behaviour

At `verified.commit`, validation is CEL-only: `runWrite`'s converged tail runs a presence-only
required-fields pre-gate (`checkRequiredFields`, write.go:521-526) between `cel.BuildPage` and
`runTemplateValidations`; `runAppend` mirrors it (append.go:218-222); `approve` calls it inside
`requiredFieldsRefusal` (approve.go:221-234). `BuildPage` date-coerces frontmatter values
(pagebuilder.go:63). The lint sweep has no schema checker; `cel_lint` builds its page map per
page (internal/lint/cel.go:60-62) with a per-`Check` clock (cel.go:44). `go.mod` names no JSON
Schema dependency.

## Requirements

REQ-001: WHEN the write or append pipeline validates a page, the system MUST construct a
validation document with exactly the top-level keys `frontmatter`, `content`, `ast`, `akb`,
`file`, in the current Phase-1 shape ([[ADR-008-phase1-current-document-shape|ADR-008]] I1).

REQ-002: The validation document's `frontmatter` values MUST be the raw YAML-parsed values; the
system MUST NOT apply `convertDateField` coercion to the validation document; the system MUST
leave coercion to the CEL page map alone.

REQ-003: WHEN the system compiles a template schema, the system MUST lower the loader's parsed
`schema:` value to JSON-shaped Go values before handing it to the compiler; the system MUST NOT
pass goccy-native maps (e.g. `map[interface{}]interface{}`) to `AddResource`.

REQ-004: The write/append validation order MUST be: selector gate, document construction, JSON
Schema pass, CEL validations, storage write.

REQ-005: WHEN the schema pass fails for a page, the system MUST NOT evaluate that page's CEL
validations; the system MUST return the schema-failure report alone.

REQ-006: The schema pass MUST NOT mutate page bytes; a failing pass MUST leave the on-disk page
byte-identical and a passing pass MUST write the pipeline's input bytes unchanged.

REQ-007: The system MUST compile each template's schema at most once per process and MUST cache
the compiled schema keyed by template identity ([[ADR-003-santhosh-v6-format-assertion|ADR-003]] I8).

REQ-008: The system MUST register akb's strict `date-time` and `date` format checkers before any
schema compilation, and the compiler MUST have `AssertFormat()` enabled (ADR-003 I2/I4/I5).

REQ-009: The write path and the sweep checker MUST share one compiled-schema cache; the system
MUST serialize all compiler and cache access behind a mutex.

REQ-010: WHEN a template schema declares a format name akb has not registered, the system MUST
treat that format as unasserted; the system MUST NOT fail compilation or validation over it.

REQ-011: WHEN the schema pass reports a failure, the system MUST map each leaf
`*ValidationError` (empty `Causes`) of the error tree to exactly one report row.

REQ-012: A schema report row MUST carry the failing keyword (last segment of
`ErrorKind.KeywordPath()`), the RFC 6901 JSON Pointer of the leaf's `InstanceLocation` escaped
`~`→`~0` then `/`→`~1`, and a single-line message.

REQ-013: A schema report row MUST NOT carry a `line` field; the row MUST carry
`source: "schema"` in place of a CEL `rule_id` (SPEC-002 REQ-008).

REQ-014: WHEN a schema failure is a `required` violation under `properties.frontmatter`, the
system MUST emit the retired pre-gate's sentence verbatim for that violation:
`missing required frontmatter field(s): <names> (declared required by template "<name>")`.

REQ-015: The system MUST NOT run the required-fields pre-gate in the write, append, or approve
paths; required-presence MUST be enforced by the schema's `properties.frontmatter.required`.

REQ-016: The system MUST retain `checkRequiredFields` and `requiredFieldsMessage` for the
`template write` mockup path until SPEC-005, and the `required_fields` lint checker MUST remain
registered (ADR-009 I5).

REQ-017: The system MUST provide a built-in `schema` lint checker that validates each page's
validation document against its type's compiled schema with format assertion enabled
([[ADR-009-phase1-sweep-schema-checker|ADR-009]] I1).

REQ-018: The `schema` checker MUST be registered before `cel_lint`, and both checkers MUST
evaluate one sweep with a single shared `now` stamp (ADR-009 I2).

REQ-019: WHEN schema validation of a page faults during a sweep, the `schema` checker MUST
degrade the fault to a per-page `LintIssue` and continue the sweep (ADR-009 I3/N3).

REQ-020: The `schema` checker MUST emit `LintIssue` rows of type `schema`; it MUST NOT emit the
SPEC-002 write-time report envelope during a sweep (ADR-009 I4/N2).

REQ-021: WHEN a template's schema fails to compile in any validation path, the system MUST
refuse with an internal fault naming the template and the compiler's typed error; the system
MUST NOT report it as a page-validation failure. In the sweep, the `schema` checker MUST
degrade the compile fault to one per-page `LintIssue` per affected page and continue.

## Acceptance Criteria

AC-001 (verifies REQ-001, REQ-002): WHEN a page with a date-like frontmatter string is
schema-validated, the document MUST present that value as the raw string and the schema's
`format: date-time` assertion MUST evaluate against it.
  verify: { method: check, check: "go test ./internal/schema/ -run TestBuildDocument" }

AC-002 (verifies REQ-004, REQ-005): WHEN a write violates both the schema and a CEL rule, the
report MUST contain schema rows only and the page MUST NOT be written.
  verify: { method: check, check: "go test ./cmd/akb/ -run TestWriteSchemaShortCircuitsCEL" }

AC-003 (verifies REQ-012): WHEN a required field is missing, the failing row MUST carry keyword
`required` and pointer `/frontmatter`; a property named `a/b` MUST render in the pointer as
`a~1b`.
  verify: { method: check, check: "go test ./internal/schema/ -run TestErrorTreeMapping" }

AC-004 (verifies REQ-007, REQ-008, REQ-009): WHEN two pages of one type are validated in one
process, the template's schema MUST be compiled once, and `format: date-time` MUST assert akb's
strict RFC 3339 profile (lowercase `t`/`z` rejected).
  verify: { method: check, check: "go test ./internal/schema/ -run TestCompileCache" }

AC-005 (verifies REQ-014, REQ-015): WHEN `akb write`, `akb append`, and `akb approve` each meet a
page missing a template-required field, all three MUST refuse with the pre-gate's sentence and
exit 1, and the page MUST be byte-identical afterwards.
  verify: { method: check, check: "go test ./cmd/akb/ -run 'TestWriteMissingRequired|TestAppendMissingRequired|TestApproveRefusesDraft'" }

AC-006 (verifies REQ-017, REQ-018, REQ-019, REQ-020): WHEN the sweep meets a page violating its
type's schema, `akb lint` MUST report a `schema`-typed issue for that page, MUST still report
`cel_lint` results computed with the same `now`, and MUST NOT abort the sweep.
  verify: { method: check, check: "go test ./internal/lint/ -run TestSchemaChecker" }

AC-007 (verifies REQ-021): WHEN a page's type carries a malformed schema, `akb write` MUST exit 2
with an `internal:`-prefixed message naming the template, and MUST NOT write the page.
  verify: { method: check, check: "go test ./cmd/akb/ -run TestWriteMalformedTemplateSchema" }

AC-008 (verifies REQ-003): WHEN a template schema compiled from the loader's parsed value
contains nested maps and anchors, compilation MUST succeed and assert as written.
  verify: { method: check, check: "go test ./internal/schema/ -run TestSchemaLowering" }

## Contract and Invariants

- The validation document builder MUST NOT read `page["frontmatter"]`; it MUST build from
  `fm.Fields` plus the routed `type`/`title`, because the CEL map's values are date-coerced.
- Pointer escaping MUST apply `~`→`~0` before `/`→`~1`; reversing the order double-escapes.
- A report-row message MUST be single-line; multi-line library renderings MUST be collapsed.
- All `Compiler.Compile` calls MUST hold the cache mutex; the santhosh `Compiler` mutates shared
  maps during compilation (compiler.go:211, loader.go:138, v6.0.3) and documents no
  goroutine-safety.
- `RegisterFormat` calls MUST precede the first `Compile`; a cache hit returns the previously
  compiled `*Schema` without re-reading formats.
- `ErrorKind.LocalizedString` MUST NOT be called with a nil printer; it dereferences the
  printer (probe-verified SIGSEGV). Messages SHOULD come from the library's own output
  rendering (e.g. `BasicOutput()` units), which needs no new dependency.

## Preserved Behaviour

- `ValidateType`/`ValidateTitle` still run before the schema pass in every branch; their
  refusal wording and exit codes MUST NOT change.
- Validation (schema and CEL) still precedes `store.Write`; the search/linkgraph transaction
  still runs after the commit funnel; a refused write still touches nothing.
- The `required_fields` lint checker's behaviour, wording, and `by_check` key MUST NOT change.
- The `template write` PASS-mockup required-field check MUST NOT change (SPEC-005 owns it).
- Exit-code classes MUST NOT change: validation refusal exit 1, usage exit 2 with `usage:`,
  akb fault exit 2 with `internal:`, other command failure exit 1 with `Error:`.
- Git-state preflight placement and `--no-commit` semantics MUST NOT change (ADR-011's work
  unit owns them).

## Decisions and Rejected Alternatives

- The report row type, compiled-schema cache, and document builder live in a new
  `internal/schema` package; `internal/template` is forbidden the compiler dependency by
  SPEC-003 AC-007 and `internal/cel` stays CEL-only.
- The cache is one mutex-guarded instance shared by the write path and the sweep checker,
  per ADR-003 I8's "compile once keyed by template identity".
- Git-state preflight ordering is parked to ADR-011 and its post-RFC-002 work unit; this SPEC
  neither reorders nor amends preflight behavior.
- Rejected: validating the CEL page map instead of a dedicated document. The page map is
  date-coerced; ADR-005 N3 and ADR-008 I1 require raw values. Do not re-propose unless ADR-005
  or ADR-008 is amended.
- Rejected: per-call-site schema compilation (no cache). santhosh `Compile` is expensive and
  ADR-003 I8 mandates caching. Do not re-propose unless compilation is proven cheap at the
  largest in-tree schema.
- Rejected: deriving report rows solely from `BasicOutput()` units. Leaf-walking the error tree
  is the enumeration contract; `BasicOutput` flattening cannot be re-shredded. Do not
  re-propose unless santhosh exports a stable leaf-enumeration API.

## Assumptions and Open Questions

- [ASSUMPTION: santhosh-tekuri/jsonschema is pinned at v6.0.3 — the version the rfc-002-s7
  libs-verify report verified from module source and runtime probes. Any other v6.x re-runs
  those probes.]
- [ASSUMPTION: concurrent `Validate` on one compiled `*Schema` is safe — no `*Schema` writes
  exist in validator.go and a 50-goroutine probe ran clean, but no `-race` run was possible
  (no CGO locally). The mutex requirement above covers `Compile`, not `Validate`.]
- OQ-1 (RESOLVED 2026-10-10): an earlier draft required moving the merge/identity preflight
  ahead of validation. Owner decision: parked out of SPEC-004 — ADR-011 records the
  validate-first ordering and the full `--no-commit` git-state bypass as a post-RFC-002 work
  unit across all mutating commands.
- OQ-2 (RESOLVED 2026-10-10): owner ratified REQ-021 as written — malformed template schema is
  an internal fault (exit 2) on the write path and a degraded per-page issue in the sweep.
  Template-write-time schema exercise (mockup-vs-schema proofs) is SPEC-005/S8 scope; until it
  lands, use-time surfacing is the only surface, and remains the backstop thereafter.

## Affected Surface and Ordering

- `cmd/akb/write.go` (tail: build, schema pass, pre-gate removal)
- `cmd/akb/append.go` (same), `cmd/akb/approve.go` (requiredFieldsRefusal retires)
- `cmd/akb/required_fields.go` (helpers retained for mockups), `cmd/akb/lint.go` (registration)
- `internal/schema/` (new: document builder, compile cache, error-tree mapping, report rows)
- `internal/lint/schema.go` (new checker), `internal/lint/cel.go` (shared-`now` threading)
- `internal/cel/pagebuilder.go` (untouched; the document builder must not reuse its coercion)
- `test/testdata/`: `lint_tests_adr_write.txt` (rewrite to schema-required message),
  `exit_codes.txt` (short-circuit changes the `-count=6` pin), `superseded_adr.txt`,
  `lint_required_fields.txt` (issue counts), `lint_tests.txt` (eleventh checker)

depends_on: [SPEC-002, SPEC-003, ADR-003, ADR-008, ADR-009, ADR-010]
blocks: [SPEC-005]

## Verification Plan

- `go test ./internal/schema/... ./internal/lint/... ./cmd/akb/...` — unit coverage above.
- `go test ./test/ -test.v` — testscript suite with the updated pins named in Affected Surface.
- `just lint` — golangci-lint clean.
- Human-only residue: a `-race` run of the concurrency tests on a CGO-enabled toolchain
  (unavailable locally at authoring time).

## Drift Ledger

- 2026-10-10 · drafted at c8bac2e from the rfc-002-s7 research run · state: proposed.
- 2026-10-10 · preflight-ordering requirement removed after owner review · parked to ADR-011.
- 2026-10-10 · malformed-schema surfacing ratified by owner (exit 2 internal + sweep degrade) · REQ-021 finalized.

## Revisit Triggers

- SPEC-003 amendment of REQ-007 (loader pass-through) — re-check this SPEC's compile-site design.
- `go.mod` — any change to the santhosh-tekuri/jsonschema version (re-run the libs-verify probes).
- RFC-002 §7 pipeline text or ADR-008 document-shape cohort — re-check REQ-001/REQ-004.
- ADR-011's reorder work unit landing — re-check the Preserved Behaviour preflight line.

## Evidence Appendix

At `verified.commit` c8bac2efcaf01d0186ce1730d1f3d17f0131ac7d:

- Pipeline tail: `cmd/akb/write.go:507-547`; append mirror `cmd/akb/append.go:203-230`.
- Pre-gate anatomy: `cmd/akb/required_fields.go:16-35`; call sites write.go:523,
  append.go:220, approve.go:227, templates_write.go:152.
- Date coercion: `internal/cel/pagebuilder.go:29,63` (`convertDateField`).
- Preflight inside write (unchanged by this SPEC; ADR-011 owns): `internal/storage/mode.go:155-159`.
- Sweep model: `internal/lint/cel.go:38-100`; registration `cmd/akb/lint.go:157-167`;
  abort-on-error `internal/lint/engine.go:100-103`.
- santhosh v6.0.3: `ValidationError` shape validator.go:960-972; `kind` package (39 kinds);
  private pointer renderer output.go:103-110 + util.go:55-59; `typeOf` schema.go:129-156;
  `AssertFormat` doc compiler.go:47-52; format laxity format.go:440,484,503.
- Test pins: `cmd/akb/write_test.go:2165`, `append_test.go:770`, `approve_test.go:673,725`;
  `test/testdata/lint_tests_adr_write.txt:18`, `exit_codes.txt:41`.

Research (filed pages): [[rfc-002-s7-log]], [[rfc-002-s7-rfc-extract]],
[[rfc-002-s7-prior-research]], [[rfc-002-s7-code-recon]], [[rfc-002-s7-libs-verify]],
[[rfc-002-s7-spec-conventions]], [[rfc-002-s7-nocommit-recon]].

Governing decisions: [[ADR-003-santhosh-v6-format-assertion|ADR-003]],
[[ADR-008-phase1-current-document-shape|ADR-008]],
[[ADR-009-phase1-sweep-schema-checker|ADR-009]], ADR-010,
[[ADR-011-validation-before-git-preflight|ADR-011]],
[[SPEC-002-structured-validation-report|SPEC-002]],
[[SPEC-003-templatev3-format-loader|SPEC-003]],
[[RFC-002-open-validation-engine|RFC-002]].

Discard log:
- "Round-trip the instance document through JSON" — rejected; santhosh `typeOf` accepts goccy's
  native `int64`/`uint64`/`float64`/`string`/`bool`/`[]any`/`map[string]any` (probe-verified),
  so only the *schema* value needs lowering.
- "Reproduce the ~60-kind classification surface in akb" — rejected; the kind count is 39 and
  rows need only `KeywordPath()`, not per-kind branches.
- "ADR-003's 'documented not thread-safe' claim" — corrected by libs-verify: the module
  contains zero concurrency text; the mutex requirement stands on source evidence instead.
- "Early merge/identity preflight in write/append (draft REQ-007 v1)" — withdrawn after owner
  review 2026-10-10; the uniform validate-first ordering and full `--no-commit` bypass belong
  to ADR-011 and its post-RFC-002 work unit, not to this SPEC.
