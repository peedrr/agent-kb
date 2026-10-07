---
anchors:
- checked_at: "2026-10-07T17:55:40Z"
  claim: REQ-001
  evidence: "resolved at write.go:603 during read-only recon 2026-10-07: collects all CEL failures into []cel.ValidationError, prints stderr lines, returns data-free validationFailure{} — the widening point"
  kind: local
  path: cmd/akb/write.go
  state: unverified
  symbol: runTemplateValidations
  verify:
    check: go test ./cmd/akb/...
    method: check
- checked_at: "2026-10-07T17:55:40Z"
  claim: REQ-004
  evidence: "resolved at errors.go:11 during read-only recon 2026-10-07: row source type {RuleID, Message, Line, Severity}; no JSON tags today"
  kind: local
  path: internal/cel/errors.go
  state: unverified
  symbol: ValidationError
  verify:
    check: go test ./internal/cel/...
    method: check
- checked_at: "2026-10-07T17:55:40Z"
  claim: REQ-009
  evidence: "resolved at main.go:78 during read-only recon 2026-10-07: maps validationFailure to exit 1 and reports nothing further — the print-then-return contract the report must preserve"
  kind: flow
  path: cmd/akb/main.go
  state: unverified
  symbol: classifyExit
  verify:
    check: go test ./cmd/akb/...
    method: check
created: "2026-10-07"
grammar: 1
id: SPEC-002
kind: change
provenance: agent-drafted
scope:
- cmd/akb/**
- internal/**
status: proposed
title: Write-Time Validation Failures Merge Into One Structured Report
type: spec
updated: "2026-10-07T19:01:06Z"
verified:
  at: "2026-10-07T17:55:40Z"
  branch: main
  commit: fc08a201e17beae3e6195f94ec832bbb7c7127a9
  method: manual-review
  next_review_by: "2027-01-05"
  state: unverified
---

# SPEC-002: Write-Time Validation Failures Merge Into One Structured Report

> **For agents:** this SPEC is authority only while `status: active` AND
> `verified.state: live`. If your task conflicts with it, or any anchor fails
> to resolve at HEAD, stop and name the conflict. Do not silently work around
> it; propose an amendment or supersession instead.

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-S1 (RFC §7 carries the
merged error model and names this SPEC; §18 governs this record's lifecycle), researched
2026-10-07 in the research session ([[rfc-002-s1-log]]); drafting decisions (report surface; mockup-attribution
exclusion) ratified by Pete Hope 2026-10-07 via decision form.

## Why

A failed write today prints human-only stderr lines and returns a data-free sentinel, so an
agent driving akb programmatically cannot tell a schema violation from a CEL failure, nor
locate either without parsing prose. RFC-002 makes JSON Schema a co-equal write-time validator
and `write --json` the agent-facing path; both require one ordered, machine-readable failure
report instead of two divergent error vocabularies.

## Goals and Non-Goals

- Goal: a single ordered report covering both validators' write-time failures, with stable
  locators (schema keyword + JSON Pointer; CEL rule ID).
- Goal: the report serialized as one JSON object on stdout when `akb write --json` fails.
- Goal: envelope stability of the same grade as the frozen `lint --json` contract, so agent
  harnesses can gate on it.

- Non-goal: the sweep. `lint --json`'s envelope is frozen (RFC-002 §11, N5); the sweep-time
  schema checker emits `LintIssue` rows, not this report.
- Non-goal: `template write` mockup attribution and a `<!-- FAILS: schema: ... -->` parser —
  excluded by the owner 2026-10-07; [[ADR-006-schema-frontmatter-retirement|ADR-006]] N7
  already defers the parser.
- Non-goal: the schema validator's integration mechanics ([[ADR-003-santhosh-v6-format-assertion|ADR-003]]'s
  domain) and the `write --json` success envelope (RFC-002 §8.2/§15).
- Non-goal: any change to exit codes, sentinel types, or human-readable stderr wording.
- Non-goal: a markdown-mode report flag (rejected 2026-10-07; see Decisions).

## Current Behaviour

- `runTemplateValidations` (cmd/akb/write.go:603) evaluates a template's CEL `validations`,
  collects every failed and unevaluable rule as `cel.ValidationError{RuleID, Message, Line: 0,
  Severity: "error"}` (internal/cel/errors.go:11), prints each to stderr, and returns the bare
  `validationFailure{}` sentinel (cmd/akb/main.go:53). The first CEL *compile* failure aborts
  as `internalError` (exit 2) before any rule is evaluated.
- Call sites: cmd/akb/write.go:529 and cmd/akb/append.go:225. `template write` uses a separate
  `evaluateValidations` (cmd/akb/templates_write.go:363) returning bare rule-ID strings;
  it is untouched by this SPEC.
- A required-field pre-gate (cmd/akb/required_fields.go:18) refuses pages missing
  template-required keys before CEL runs; RFC-002 §7 folds this gate into the schema pass,
  whose `required` failures become schema rows of the report.
- No machine-readable failure output exists: `validationFailure` carries no payload,
  `cel.ValidationError` has no JSON tags, and `classifyExit` (cmd/akb/main.go:78) maps the
  sentinel to exit 1 and reports nothing further (print-then-return).
- `approve` reuses the sentinel: `approveAllDraftPages` (cmd/akb/approve.go:265) matches it
  with `errors.As` to treat a refused draft as non-fatal-per-page, and `requiredFieldsRefusal`
  (cmd/akb/approve.go:220) deliberately reuses write's refusal wording.
- The only structured-report precedent is `lint --json`: `{issues: [{check, rule_id?, message,
  path, severity}], summary: {total, pages_checked, by_check}}` (internal/lint/engine.go:17,28;
  cmd/akb/lint.go:220), frozen by N5.

## Requirements

REQ-001: WHEN `akb write` or `akb append` fails template validation, the system MUST collect
every JSON Schema violation and every failed or unevaluable CEL rule for the page into a
single ordered validation report before the command returns.

REQ-002: WHEN the schema pass fails for a page, the system MUST NOT evaluate that page's CEL
validations; the system MUST return a report whose rows come from the schema pass alone.

REQ-003: WHEN the schema pass fails, the system MUST report each violation with the failing
schema keyword, the JSON Pointer of the offending instance location, and a single-line
message.

REQ-004: WHEN CEL validations fail, the system MUST report each failed or unevaluable rule
with its declared rule ID and a single-line message; an unevaluable rule MUST appear as a
report row and MUST NOT abort report collection.

REQ-005: WHEN `akb write --json` fails validation, the system MUST write the report to stdout
as one JSON object and MUST exit with code 1.

REQ-006: The system MUST order report rows in evaluation order: schema rows before CEL rows,
and CEL rows in the template's rule declaration order.

REQ-007: The system MUST mark every write-time report row with severity "error".

REQ-008: The report MUST NOT attach a `line` field to schema-violation rows; a CEL row MUST
omit `line` unless the CEL layer supplied a nonzero line.

REQ-009: WHEN `akb write --json` fails validation, the system MUST print the same
human-readable failure lines to stderr that the markdown-mode write path prints.

The report's JSON shape (REQ-005), emitted only on failure:

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

`path` is the KB-relative page path (the stripped-`relPath` convention). Schema rows carry
`keyword` and `pointer` instead of `rule_id`. A write-time report is single-source by REQ-002
(schema failure skips CEL); the `summary` keeps both count fields so the envelope never
changes shape between failure modes.

## Acceptance Criteria

AC-001 (verifies REQ-001): WHEN a page failing two CEL rules is written, the command MUST
exit 1 and the report MUST contain both rule IDs — no failure is dropped within a layer.
  verify: { method: check, check: "go test ./test/ -test.run 'TestScript/write_report_aggregates'" }

AC-002 (verifies REQ-002): WHEN a written page violates the schema (missing required
`frontmatter.title`) and would also fail a CEL rule, the report MUST contain schema rows only
and MUST NOT contain any CEL rule-failure or unevaluable-rule row.
  verify: { method: check, check: "go test ./test/ -test.run 'TestScript/write_report_schema_short_circuits'" }

AC-003 (verifies REQ-003): WHEN a page missing required `frontmatter.title` fails the schema
pass, its report row MUST carry keyword `required` and pointer `/frontmatter`.
  verify: { method: check, check: "go test ./test/ -test.run 'TestScript/write_report_schema_rows'" }

AC-004 (verifies REQ-004): WHEN a CEL rule reading an absent optional key without `has()` is
evaluated, the report MUST include an error-severity row with that rule's ID, and every
later-declared rule MUST still be evaluated.
  verify: { method: check, check: "go test ./cmd/akb/ -run 'TestRunTemplateValidations'" }

AC-005 (verifies REQ-005): WHEN `akb write --json` fails validation, stdout MUST parse as
exactly one JSON object with `valid: false`, the page's KB-relative `path`, an `errors`
array, and `summary` counts whose `total` equals the array length; the exit code MUST be 1.
  verify: { method: check, check: "go test ./test/ -test.run 'TestScript/write_json_report'" }

AC-006 (verifies REQ-006): WHEN a page fails two CEL rules declared in order `[a, b]`, the
report's rows MUST appear in that same order.
  verify: { method: check, check: "go test ./cmd/akb/ -run 'TestRunTemplateValidations'" }

AC-007 (verifies REQ-007): WHEN any write-time report is emitted, every row MUST carry
`severity: "error"`.
  verify: { method: check, check: "go test ./test/ -test.run 'TestScript/write_json_report'" }

AC-008 (verifies REQ-008): WHEN a report contains schema rows, no row MUST carry a `line`
key; a CEL row serialized today MUST omit `line` (the write-time CEL layer supplies none).
  verify: { method: check, check: "go test ./test/ -test.run 'TestScript/write_json_report'" }

AC-009 (verifies REQ-009): WHEN `akb write --json` fails validation, stderr MUST contain the
same `[rule_id] message` lines that markdown-mode `akb write` prints for the same page.
  verify: { method: check, check: "go test ./test/ -test.run 'TestScript/write_json_report'" }

## Contract and Invariants

- The report MUST NOT alter the frozen `lint --json` envelope; consumers needing sweep
  results MUST use `akb lint --json` instead.
- Consumers MUST NOT gate on message text (validator or cel-go strings churn across
  upgrades, [[ADR-004-cel-go-upgrade-pinned-extensions|ADR-004]]); consumers MUST gate on
  `source`, `keyword`, `rule_id`, and `pointer` only.
- Engine faults (CEL compile errors, env construction, unreadable `old_page`) MUST NOT be
  folded into the report; they MUST remain exit-2 akb faults.
- The report JSON MUST NOT be written to stderr; the structured form MUST go to stdout only.
- Report rows MUST carry single-line messages; a multi-line validator message MUST be
  collapsed to its first line.
- Empty collections in the report MUST serialize as `[]`, never `null` (akb JSON convention,
  pinned by `raw_status_test.go`).
- A failed validation MUST NOT persist page bytes; the write path stores either a passing
  page or nothing (RFC-002 P6).
- Once `status: active`, the report envelope MUST evolve additively only; removing or
  renaming a field requires supersession of this SPEC.

## Preserved Behaviour

- `validationFailure` remains an `errors.As`-matchable sentinel; `classifyExit` maps it to
  exit 1 and prints nothing further, so failure output is never doubled.
- `approveAllDraftPages` continues to treat a refused page as non-fatal-per-page via the same
  sentinel match (cmd/akb/approve.go:265), and approve's refusal wording continues to match
  write's (cmd/akb/approve.go:220).
- Markdown-mode `write`/`append` failure output stays human-readable stderr lines only.
- The first CEL compile error still aborts as an exit-2 akb fault before any rule evaluation.
- Within the evaluated layer, every failed and every unevaluable rule is still collected and
  reported (all-errors aggregation).
- The `--json` / `--append` / `--frontmatter` mutual exclusion stays a usage error, exit 2.
- `lint --json`, `LintIssue`, and the lint text format are untouched.
- Write-path unevaluable rules remain blocking errors while the lint sweep degrades them to
  per-page issues — the deliberate divergence documented at cmd/akb/write.go:598-602.

## Decisions and Rejected Alternatives

- Report surface is RFC-literal (owner-ratified 2026-10-07 via decision form): the structured
  report exists only where `--json` exists. Rejected: a markdown-mode report flag — unratified
  flag surface under RFC-002 P-i. Do not re-propose unless agents demonstrably drive
  markdown-mode writes and need machine-readable failures.
- Mockup attribution excluded (owner-ratified 2026-10-07): the `<!-- FAILS: schema: ... -->`
  parser rides Phase-1's template-write schema obligations, per ADR-006 N7. Do not re-propose
  unless that Phase-1 work is folded into this SPEC's delivery.
- akb-owned envelope, not santhosh's `OutputUnit`/Basic/Detailed output: the library's JSON
  tags are quirky (a capitalized `AbsoluteKeywordLocation` tag diverges from the 2020-12
  output spec) and akb needs N5-grade freeze control. Do not re-propose unless akb commits to
  emitting the JSON Schema 2020-12 output formats for external tooling.
- Schema failure short-circuits CEL (RFC-001 §7's ordering, carried unchanged by RFC-002 §7):
  CEL rules are authored to read schema-required keys unguarded, so evaluating them against a
  shape-broken page yields misleading unguarded-key eval errors. Rejected: always evaluate
  both layers. Do not re-propose unless CEL rules lose the unguarded-read guarantee.
- Rows share no type with lint's `LintIssue`: the locators differ (JSON Pointer vs page path)
  and N5 freezes `LintIssue`. Do not re-propose unless N5's freeze is lifted.
- Severity is carried on every row and fixed to "error" at write time: the write path is
  fail-closed and has no warning class. Do not re-propose unless write-time validation gains
  a warning class.

## Assumptions and Open Questions

- [ASSUMPTION: santhosh-tekuri/jsonschema v6's error tree is
  `ValidationError{SchemaURL, InstanceLocation []string, ErrorKind, Causes}` with a public
  `kind` package — verified 2026-10-07 against the v6.0.3 module cache, to be re-verified
  against the version go.mod pins under ADR-003.]
- [ASSUMPTION: the schema pass replaces the required-field pre-gate with equivalent refusal
  coverage (RFC-002 §7); exact message text is implementation-owned, not contract.]
- Open: the Go package hosting the report type (`internal/cel` is CEL-only; a schema+CEL
  report wants a new home) — an implementation choice, not contract.
- Open: whether `approve` ever gains a `--json` surface; this SPEC specifies internals reuse
  only.

## Affected Surface and Ordering

- `cmd/akb/write.go` — `runTemplateValidations` widened to build and return the report;
  `runWrite` emits it (stdout under `--json`; stderr lines unchanged).
- `cmd/akb/append.go` — `runAppend` consumes the widened signature.
- `cmd/akb/approve.go` — sentinel semantics preserved; no surface change.
- `cmd/akb/main.go` — `validationFailure` / `classifyExit` unchanged but load-bearing.
- `internal/cel/errors.go` — `ValidationError` may gain JSON tags or be wrapped by the
  report row type.
- `test/testdata/` — new testscript cases named in Acceptance Criteria.
- The report type's package home is an implementation choice (see Open Questions).

depends_on:
- "ADR-003: santhosh v6 integration provides the schema pass and its error tree"
- "RFC-002 write --json surface (§8.2/§15): the stdout emission point lands with it; the
  widened internals land with the merged error model"

## Verification Plan

- `go test ./cmd/akb/...` — report assembly, ordering, JSON serialization (`[]` never
  `null`), sentinel compatibility.
- `go test ./internal/cel/...` — `ValidationError` tags/wrapping.
- `go test ./test/ -test.v` — the testscript cases named in Acceptance Criteria plus the
  existing write/append/approve suites, which must pass unchanged.
- Human-only residue: (1) owner ratification (`proposed` → `active`); (2) conformance of the
  emitted envelope against a real agent harness (the pi-subagents structured-output flow of
  RFC-002 §15) once `write --json` ships — not hermetically runnable.

## Drift Ledger

- 2026-10-07 · SPEC-002 · ∅ → proposed · authored from RFC-002 §7/§18 plus owner decision-form
  ratifications (surface, mockup scope) · evidence: [[rfc-002-s1-log]]

## Revisit Triggers

- `cmd/akb/write.go`, `cmd/akb/append.go` — the widening lands or the call sites move.
- `cmd/akb/main.go`, `cmd/akb/approve.go` — the sentinel or exit contract changes.
- `internal/lint/engine.go`, `cmd/akb/lint.go` — any change to the frozen envelope forces a
  boundary review.
- `go.mod` — the santhosh-tekuri/jsonschema pin lands or changes (re-verify the error-tree
  assumption).
- `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md` — amendments to §7 or §15,
  or the §18 S1 row moving to ratified (this SPEC's Origin paragraph must then cross-check).
- `docs/knowledge-base/.agent-kb/templates/spec.yaml` — grammar bump.

## Evidence Appendix

At `verified.commit` fc08a201e17beae3e6195f94ec832bbb7c7127a9:

- Write path: cmd/akb/write.go:603-651 (`runTemplateValidations`), :522-526 (required-field
  pre-gate), :529 (call site), :598-602 (deliberate lint/write divergence comment);
  cmd/akb/append.go:225 (call site); cmd/akb/required_fields.go:18,32.
- Sentinels and exit contract: cmd/akb/main.go:27 (codes), :53 (`validationFailure`), :78-91
  (`classifyExit`); cmd/akb/approve.go:220,233,265.
- CEL layer: internal/cel/errors.go:11-21; internal/cel/engine.go:41,64,86-102 (compile,
  evaluate, panic/cost recovery).
- Structured-report precedent: internal/lint/engine.go:17,28 (`LintIssue`/`LintReport`);
  cmd/akb/lint.go:164,200,220-246 (text/JSON rendering, exit policy).
- Mockup divergence: cmd/akb/templates_write.go:150-221,363; the `FAILS:` convention is
  documentary only (zero Go parsers).
- RFC authority: RFC-002 §7 (:332-335 — the carry that names this SPEC), §17 (:750 — Phase 1
  includes the merged error model), §18 (:775-776 — candidate drafting rule; :796 — the S1
  row). RFC-001 §7:223-240 (ordering + error merge), §11:329 (stdout binding), §15:373-374
  (the original candidate charter) — all carried by RFC-002 §3 and Appendix B, never cited as
  independent authority.
- Governing decisions: [[ADR-003-santhosh-v6-format-assertion|ADR-003]] (validator + format
  assertion), [[ADR-004-cel-go-upgrade-pinned-extensions|ADR-004]] (message-churn warning),
  [[ADR-005-date-bridge-determinism|ADR-005]] (I10/I11: messages name rule ID + path +
  remediation), [[ADR-006-schema-frontmatter-retirement|ADR-006]] (N7: `FAILS:` stays
  unparsed until its parser lands).
- santhosh v6.0.3 error tree, freshly verified 2026-10-07 from the local module cache:
  validator.go:960-977 (`ValidationError`/`ErrorKind`), kind/kind.go (~60 public kinds),
  output.go:114-166 (Flag/Basic/Detailed output; capitalized `AbsoluteKeywordLocation` tag).
- Discard log: library `OutputUnit` reuse (rejected — tag quirk plus freeze control);
  v5-shaped error field names from the original research brief (superseded by the v6.0.3
  verification); simultaneous schema+CEL rows in one write-time report (discarded — REQ-002's
  short-circuit makes each write-time report single-source; the envelope still carries both
  counts); a markdown-mode report flag and always-on stdout JSON (owner-rejected 2026-10-07).
