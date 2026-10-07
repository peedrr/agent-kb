---
grammar: 1
type: research
title: "Research extract — RFC-002-S1 structured validation report"
status: final
provenance: agent-drafted
run: rfc-002-s1
as_of: "2026-10-07"
created: "2026-10-07"
updated: "2026-10-07"
scope:
  - "docs/knowledge-base/**"
  - "cmd/akb/**"
  - "internal/cel/**"
  - "internal/lint/**"
tags:
  - validation-report
  - research-extract
  - spec-002
  - prior-art
summary: "Verbatim extraction of the RFC-001 merged-error-model requirements and prior a1–a4 error findings that constrain the SPEC-002 report design, with gaps marked."
informs:
  - SPEC-002
---
# Research extract — RFC-002-S1 (structured validation report / widened `runTemplateValidations`)

Read-only scout. Repo: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb`.
Each section: anchors + verbatim operative sentences. Gaps marked **[GAP]**.

---

## 1. `docs/knowledge-base/kb/rfcs/RFC-001-json-schema-cel-coexistence.md` — merged error model

All line anchors are verbatim locations in RFC-001.

### 1.1 §7 "Validation pipeline (write/append)" (:223-240) — **the operative requirement**

Pipeline fence (:225-227):

> `merge-conflict check → required gate (from schema) → JSON Schema → CEL validations → write`

:229-230 (order):
> **Order:** structural first, CEL second — CEL may then read required keys unguarded, and a
> page failing shape never pays comprehension cost.

:231-234 (**Error merge**, the sentence S1 exists to implement):
> **Error merge:** schema violations (keyword + JSON Pointer) and CEL rule failures (rule ID)
> merge into one report, one exit-1 outcome. Human-readable stderr stays as today; a structured
> form requires widening `runTemplateValidations` (currently returns data-free
> `validationFailure{}`; two callers: write.go:521, append.go:222) → tech-spec.

:239 (flag decision that exists in RFC-001):
> **`--json` is mutually exclusive with `--append`/`--frontmatter`** (usage error, exit 2).

### 1.2 §11 "CLI surface" (:322-334) — stdout vs stderr + exit code

:329:
> | `akb write --json` failures | structured validation report on stdout (rule IDs + JSON Pointers), exit 1 | requires `runTemplateValidations` widening |

:331 (round-trip the report must not break):
> Round-trip invariant: `read --json | write --json` must be an effective no-op (modulo akb's
> managed mutations). This is guaranteed by §5.1 ... and is a pinned integration test.

### 1.3 §15 candidate list (:364-383) — the S1 charter, verbatim

:373-374:
> 4. **Tech-spec: structured validation report** — widening `runTemplateValidations`, merged
>    schema+CEL error model, stable IDs/JSON Pointers, stdout/stderr split.

### 1.4 §13 (lint side — the report must not churn the frozen envelope)

:351:
> - All issues flow into the existing `LintIssue`/`--json` envelope unchanged (N5).

N5 (:97-99):
> **N5.** `akb lint --json`'s envelope shape is **stable** — it is load-bearing for
> `agent-memory/DESIGN.md` §4.3 typed gates ... New checkers may add rows; the envelope does not change.

### 1.5 §3 Goals / §4 P4 (failure contract the report encodes)

G7 (:83-84):
> Preserve the three-surface failure contract: write fails closed, lint sweep degrades
> per page, `template write` proves mockups.

P4 (:109-110):
> **P4. Fail closed on write, degrade on sweep, prove at authoring.** Unchanged; extended to
> the schema layer.

### 1.6 Adjacent constraints a report must respect

- §5.3 (:144-155) — pre-publication hazards; hazard 5: "`line` numbering convention ... must be pinned" (CEL `ValidationError.Line` feeds the report).
- §6.2 (:214-221) — fail mockup convention extends to `<!-- FAILS: schema: /frontmatter/status -->`.
- §12 (:337-342) — approve re-runs schema + write-time validations; its failures reuse the same report.
- N2 (:92-93) JSON is not a storage format; N3 (:94) no schema for managed files.

**[GAP]** RFC-001 decides **no flag name**: the word "`--json-errors`" (or any error-only flag) occurs nowhere in `docs/` or `.pi/`. The only decided surface is "structured report on stdout when `akb write --json` fails" plus "human-readable stderr stays as today". Whether the report is emitted only under `--json` or always is **undecided** — spec must decide.
**[GAP]** No JSON envelope/schema shape is given: only "(rule IDs + JSON Pointers)" (:329) and "stable IDs/JSON Pointers" (:373-374). Field names, ordering, `file`, summary counts, and whether CEL failures carry `line` are all open.

---

## 2. `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md` — carrier (must optimize for both)

:332-335 (§7):
> RFC-001 §7 carries with two changes: the required-presence gate is the schema's own
> `required` (no separate pre-gate), and **no mutation step exists** — validation either
> passes and the bytes are written, or fails and nothing is written. Error merge (schema
> violations by keyword + JSON Pointer; CEL failures by rule ID; one exit-1 report) and the
> `runTemplateValidations` widening carry unchanged (**RFC-002-S1**).

:358 (§8.2, `write --json` output shape):
> stdout echoes the persisted document in **exactly `read --json`'s shape** (`{frontmatter, content}` ...)

:362-364 (§8.2) — mutual exclusion survives:
> **Mutual exclusion (RFC-001 §7 carries):** `--json` is incompatible with
> `--append`/`--frontmatter` (usage error, exit 2) ...

:796 (§18 row — S1 status `proposed`, no Resolves-as yet):
> | RFC-002-S1 | SPEC | Structured validation report (widened `runTemplateValidations`) | proposed | — |

:885 (Appendix B): "merged error model + `runTemplateValidations` widening; JSON→YAML assembly landmine" listed as carried from RFC-001 unchanged.
:12/§18 note: S1 IS old candidate ordinal 6 (`Renumbering note`, :814-816).
:750 (§17 Phase 1): "pipeline reorder; merged error model; cel-go v0.32.x + ext libraries" — S1 ships in Phase 1.

---

## 3. `.pi/research/rfc-002-a1-validator/` — what A1 established, and what it did NOT

### 3.1 LOG.md — settled decisions that constrain the report

- **D-A1** (LOG.md:44-49): library = `github.com/santhosh-tekuri/jsonschema/v6`; "only library with verified full 2020-12 (100% bowtie badge), overridable built-in format checkers (runtime-verified), active maintenance, Apache-2.0, RE2 default."
- **D-A5** (LOG.md:61-64): "register formats before Compile; one compiled schema cached per template ... compilation not thread-safe — akb is a serial CLI, non-issue; concurrent Validate on one schema unverified, akb validates serially."
- Format policy (LOG.md:98-101 of `research-web.md` Q3): `Compiler.AssertFormat()` enables 2020-12 assertions; `RegisterFormat` shadows built-ins (only `regex` un-overridable); OQ2 ratified 2026-10-06: **"assert everything declared"**.
- Integration point (`recon-code.md:96-105`, `[H]` confidence): schema pass slots **after** the required gate (`cmd/akb/write.go:523`) and **before** CEL (`write.go:529`); validates the RAW document view.
- `recon-code.md:100-105` — current shape of the function to widen:
> `runTemplateValidations(...)` — **`:603-651`** **[H]**. Compiles each `tmpl.Validations[].Rule` (`:605`), evaluates with `{page, old_page, now}` (`:616-622`), collects **all** failures + unevaluable rules into `[]cel.ValidationError`, prints to stderr, returns `validationFailure{}`. Fail-closed on compile errors (`&internalError`) and eval errors (collected → validation exit) **[H]** — the A1 consequence "schema pass fails the write with the same exit-1 shape" should mirror this aggregation + stderr contract.
- `research-web.md` Q5 (:219-232) — `Compiler.Compile` ≈2.8 ms, `(*Schema).Validate` per instance; "Whether a *single compiled* `*Schema` is safe for concurrent `Validate` is **not documented**".

### 3.2 **[GAP]** — santhosh v6 ERROR TYPES are NOT established anywhere in the research base

Exhaustive check of `.pi/research/rfc-002-a1-validator/*.md`, `-a2`, `-a3-celgo`, `-a4`, `docs/`, `.pi/subagents/`:
- `ValidationError`, `OutputUnit`, `BasicOutput`, `DetailedOutput`, `FlagOutput`, `InstanceLocation`, `KeywordLocation`, `AbsoluteKeywordLocation`, `Causes` — **0 hits in prose research** (grep across `.pi/**/*.md` and `docs/**/*.md`).
- `ADR-003` has **zero occurrences of the word "error"** — it pins library, format profile and I8 cache, nothing about error reporting, output verbosity, or `Validate`'s return type.
- No decision about error verbosity exists (grep "verbose|verbosity" in the RFC-001-era base → only `Seam`/`display` hits; no decision).

**Task premise correction:** the assumed field set `InstanceLocation/KeywordLocation/AbsoluteKeywordLocation/Message/Causes` is v5-shaped. v6 differs — see §3.3.

### 3.3 Supplement (verified at source, read-only): `jsonschema/v6.0.3`

Source: the module zip already in the local module cache (`/home/pete/go/pkg/mod/cache/download/github.com/santhosh-tekuri/jsonschema/v6/@v/v6.0.3.zip`; the file-backed `@v6.0.3` dir is absent), extracted to `/tmp/s1-santhosh/...` — **no repo writes**. Not in the research base; the spec author should cite it as freshly verified if used.

`validator.go:960-972` — the error tree:
```go
type ValidationError struct {
	// absolute, dereferenced schema location.
	SchemaURL string
	// location of the JSON value within the instance being validated.
	InstanceLocation []string
	// kind of error
	ErrorKind ErrorKind
	// holds nested errors
	Causes []*ValidationError
}
```
`validator.go:974-977`:
```go
type ErrorKind interface {
	KeywordPath() []string
	LocalizedString(*message.Printer) string
}
```
- **How to walk it:** recursive `Causes`; keyword comes from `ErrorKind.KeywordPath()`, message from `ErrorKind.LocalizedString(printer)`, instance pointer = join of `InstanceLocation` (already []string tokens), schema location = `SchemaURL`. `validator.go:15-46`: `(*Schema).Validate(v any) error` returns a **synthetic root** `*ValidationError{ErrorKind: &kind.Schema{Location: sch.Location}, InstanceLocation: nil, Causes: <errors>}` — the root's `Causes` is the real list (unwrap `*kind.Group` when present). Nested aggregation uses `vd.addErrors(errors, kind)` (`validator.go:842-847`).
- **Kinds** are a public package: `github.com/santhosh-tekuri/jsonschema/v6/kind` (single file `kind/kind.go`), ~60 types with data fields usable as report detail, e.g. `Type{Got string, Want []string}`, `Enum{Got any, Want []any}`, `Required{Missing []string}`, `AdditionalProperties{Properties []string}`, `Pattern{Got, Want string}`, `Format{Got any, Want string, Err error}`, `MinLength{Min, Got int}`, `RefCycle{URL, KeywordLocation1, KeywordLocation2}`. Kind names ≈ keyword names (each `KeywordPath()` returns e.g. `["type"]`, `["enum"]`, `["format"]`, `["required"]`, `["additionalProperties"]`, `["pattern"]`), but `KeywordPath()` may return `nil` (Schema/Group/InvalidJsonValue/RefCycle) and composite kinds (`AllOf/AnyOf/OneOf/Not`) nest via `Causes`.
- **JSON marshalling (the 2020-12 output formats), `output.go:114-166`:**
```go
type FlagOutput struct { Valid bool `json:"valid"` }                      // :115-117
func (e *ValidationError) FlagOutput() *FlagOutput                        // :120-122
type OutputUnit struct {
	Valid                   bool         `json:"valid"`
	KeywordLocation         string       `json:"keywordLocation"`
	AbsoluteKeywordLocation string       `json:"AbsoluteKeywordLocation,omitempty"`
	InstanceLocation        string       `json:"instanceLocation"`
	Error                   *OutputError `json:"error,omitempty"`
	Errors                  []OutputUnit `json:"errors,omitempty"`
}
func (e *ValidationError) BasicOutput() *OutputUnit                       // :149-151  (flat)
func (e *ValidationError) DetailedOutput() *OutputUnit                    // :159-161  (nested)
```
  - `OutputError` marshals to the **message string** (`MarshalJSON` → `Kind.LocalizedString`, `output.go:144-146`).
  - `KeywordLocation`/`AbsoluteKeywordLocation` are **computed**, not stored (`absoluteKeywordLocation()` :25-36, `jsonPtr` :103-110); `skip()` :38-44 collapses single-`Reference` causes.
  - Quirk: the struct's JSON tag is capitalized `AbsoluteKeywordLocation`, diverging from JSON Schema 2020-12's `absoluteKeywordLocation`; `valid` is always `false` on an error tree. Syntactic-only output (Basic/Detailed) is **not the v6 error struct** — if S1 wants the spec's output format verbatim it can reuse `DetailedOutput()`; if it wants akb's own envelope it should walk `Causes` directly.
  - Library-side verbosity knob is only `LocalizedError` (terse) vs `LocalizedGoString` (`display(..., verbose=true)`, `output.go:97-101`); there is no severity/limit/truncation API.

---

## 4. a2 / a3 / a4 — error-bearing findings

### 4.1 `.pi/research/rfc-002-a2/LOG.md` (schema.frontmatter retirement, ADR-006)

- F3 (LOG.md:24-27): rejection message doctrine —
> Recipe home = CHANGELOG of the shipping release (full recipe) + compact self-sufficient rejection message (mapping essence + CHANGELOG pointer); SKILL documents V3 only, never migration.
- LOG.md:26: "Rejection posture: clean break, no alias; unify the two divergent `detectOldFormat` copies into one helper/message."
- Key findings (recon-code): "`detectOldFormat` precedent: TWO divergent copies (template.go:73-82 capital/period vs templates_write.go:461-467 lower/semicolon) ... plain error → exit 1"; "testscript precedent: `test/testdata/cel_old_format.txt` asserts **both message spellings**".
  → A report-message change must account for pinned message strings in testscript.
- "`FAILS:` convention is NOT machine-parsed anywhere (0 .go hits) — documentary only — RFC-002's `<!-- FAILS: schema: /pointer -->` extension requires a NEW parser."
- Required-field gate: `cmd/akb/required_fields.go:16-35`, "exit 1, all missing listed sorted" (write/append/approve/templates_write); sweep twin `internal/lint/required_fields.go:83`. RFC-002 §7 folds this gate into schema `required`, so the report must absorb its failures.

### 4.2 `.pi/research/rfc-002-a3-celgo/LOG.md` (cel-go upgrade, ADR-004) — error strings that move

- LOG.md:83-86 (code recon, `[H]`-confidence anchors) — cel-go-owned error strings **pinned in akb tests**:
> `cel.CostLimit(uint64)` (`internal/cel/engine.go:61`) ...; `interpreter.EvalCancelledError` + `interpreter.CostLimitExceeded` (`engine.go:86,99`, `engine_test.go:314-315`); cel-go-owned error strings pinned in tests: `"no such key: ..."` (write_test.go:1090,2304; templates_write_test.go:734,845; template_test.go:443), `"overload"` substring (engine_test.go:285-286).
- LOG.md:98: cost canary — "cubic-cost rule over 2048 headings must abort <10s (engine_test.go:186-203)".
- LOG.md:103-105 / 108-116: module path `github.com/google/cel-go` → `cel.dev/cel-go` in v0.32.0 (#1413); read-only alias shim; **the only compile-break for akb's surface is imports**; `cel.CostLimit` doc-identical v0.28↔v0.32; panic-recovery shape unchanged ("`prog.Eval` recovers `EvalCancelledError` into an error return; akb already handles BOTH panic and returned-error paths (`engine.go:86-89,98-102`) → forward-compatible").
- LOG.md:136: ext cost accounting is real ("uncosted functions fall to default cost++ (=1)") — relevant if the report ever surfaces cost.
- **Net for S1:** no cel-go upgrade is claimed to change eval error *text*; the pinned strings are akb-test owned and must be preserved. `engine.go:86-102` error paths (panic-recovered "exceeded compute budget" vs returned `CostLimitExceeded`/`EvalCancelledError`) are two shapes a report must unify.
- a3 `RFC-002-edited.md` retained (draft copy) — not an authority; the KB page is.

### 4.3 `.pi/research/rfc-002-a4/LOG.md` (date bridge, ADR-005)

- D1 (Decisions made, `D1 (F1 — traversal semantics)`) — coercion-failure semantics:
> Parse failure leaves the raw string (only reachable on non-matching branches or sweep-bypassed pages).
- D2 (:54-58):
> syntactic AST walk at `template write` ... **Warnings (never errors)** for: `timestamp()` on undeclared path; `duration()` on format:duration path. Covers validations + lint_rules, `page.`/`old_page.` roots; non-path args silent (documented limitation ...).
- ADR-005 requirements the report/UX must satisfy (`kb/decisions/ADR-005-date-bridge-determinism.md`): **I10** — "MUST emit a stderr warning naming the rule ID, the attribute path, and the remediation"; **I11** (owner-amended) — warning must name "the rule ID, the attribute path, the vocabulary mismatch ... and the remediations"; **I12** — no warning for non-statically-resolvable args.
- **Precedent for S1's message contract:** akb already has a normative "message carries rule ID + path + remediation" requirement (I10/I11). S1's report rows should be consistent with it.
- a4 also records (Open questions): "Swept-side schema checker ownership ... left to implementation planning" — i.e. not settled; `lint/cel.go` eval errors degrade to a per-page `LintIssue` (deliberate divergence, RFC-001/AGENTS.md).

---

## 5. `.pi/subagents/proposals/json-cel/research/` — RFC-001-era error-model detail (not in RFC-001 text)

Files (13): `LOG.md`, `SYNTHESIS-openness.md`, `seam-core-pipeline.md`, `seam-periphery.md`, `openness-write-path.md`, `openness-lint-engine.md`, `openness-document-model.md`, `openness-config-surface.md`, `akb-recon.md`, `command-inventory.md`, `schema-survey.md`, `cel-bridge-research.md`, `cel-bridge-research-b.md`, `recon-search-after.md`.
No separately named error-model report exists — the material is in `seam-core-pipeline.md` §6 + `akb-recon.md` §3 + `SYNTHESIS-openness.md` §5.

### 5.1 `seam-core-pipeline.md:274-280` — the design origin of the merge (richest statement)
> **Ordering and error aggregation:** JSON Schema validation errors + CEL rule errors must be merged into one report and one exit-1 outcome (current: all failed rules printed to stderr, then `validationFailure{}` → exit 1 via main.go:86-90 `classifyExit`). A `--json` mode needs that merged report serialized to stdout with stable rule IDs (JSON Schema keyword + instance pointer, CEL rule `ID`).

Same file :286-293 — mockup obligations (`<!-- FAILS: schema: /frontmatter/status -->`), and: "a mockup whose only defect is schema-level (or whose only defect is CEL-level) must be attributable".

### 5.2 `akb-recon.md` — current machinery and blockers
- :116: "Rule failures are typed: `internal/cel/errors.go:11-16` `ValidationError{RuleID, Message, Line, Severity}` (`Error()` → `[id] message`, optional `(near line N)`)."
- :139: "`runTemplateValidations` (`:595-635`): compiles each rule (compile failure → `internalError` = exit 2), **collects every failed and every unevaluable rule**, prints each with `fmt.Fprintln(os.Stderr, ve.Error())`, returns the bare `validationFailure{}` (exit 1)."
- :182 / :198 (d): "**Validation results are not machine-readable** ... A structured `--json` write needs that function's signature widened (two callers: `write.go:521`, `append.go:222`)."
- :172 (JSON conventions): "results to **stdout** as a JSON envelope/array, failures stay human text on stderr with the normal exit codes; empty collections must serialize as `[]` not `null` (pinned by `raw_status_test.go:41-90`)." (This is the akb-wide convention S1's envelope should follow.)
- :151 precedent payloads: `list.go:22-25`, `search.go:95-119` (`json.Encoder` + `SetIndent`), `links.go:154-206` (envelope struct), `raw_status.go:170-180`.

### 5.3 `SYNTHESIS-openness.md`
- :96 (carry-over list): "Merged error model, structured validation report (tech-spec)".
- :51 (N5 risk): "Freeze must be stated at the *envelope* level with disabled checkers reporting 0, or gates break" — S1 must not depend on `by_check` keys it may alter.
- :31 (#3 of structural findings): the lint engine is "already an open registry" — a `schema` checker must emit `LintIssue{check,rule_id?,message,path,severity}` rows, not a new shape.

### 5.4 `command-inventory.md`
- :39 `lint` JSON today: `{issues:[{check,rule_id?,message,path,severity}],summary:{total,pages_checked,by_check}}` (`lint.go:232-246`); exit 1 iff any error-severity issue.
- :145-146: "**`lint`** — must aggregate both schema-type findings and CEL findings into one `LintIssue{check,rule_id?,message,path,severity}` list; the boundary is inside one command's output, not between commands."
- :47: today 6 of 31 commands emit JSON; no input command consumes JSON.

### 5.5 `openness-lint-engine.md`
- :63 sample envelope; :54-55 `Severity` is an unvalidated bare string; :20-23 no surface to add/re-severity checkers; A6 exit policy = "fail iff any issue has severity exactly `error`".

---

## 6. Gaps / open questions for the spec author

1. **[GAP] Flag name undecided.** No `--json-errors` (or equivalent) anywhere. RFC-001 only binds the report to `akb write --json` failures (stdout, exit 1) and says stderr stays human-readable. Decide: report on every exit-1 validation failure vs only under `--json`; a separate `--json-errors`; and whether `--json` output is a document (success) or a report (failure) — two shapes on one stream.
2. **[GAP] No report envelope schema.** Field names/ordering/detail level undecided; candidate inputs: `OutputUnit` (library), `LintIssue` (lint), `cel.ValidationError{RuleID,Message,Line,Severity}` (write). "stable IDs/JSON Pointers" only.
3. **[GAP] Library error types were never researched.** ADR-003 (accepted) is silent on them; the `santhosh v6` picture in §3.3 above is fresh, not ratified research. Spec author should treat §3.3 as new evidence to cite, and should reconcile the v5-shaped field names in the task brief.
4. **[GAP] `line` semantics.** CEL `ValidationError.Line` exists but RFC-001 §5.3 hazard 5 says line conventions are unpinned; schema errors have no line (JSON Pointer only). Report must define `line` as optional or drop it.
5. **[GAP] Verbosity decision absent** — no research or ADR chooses terse vs verbose (`LocalizedError` vs `LocalizedGoString`) or truncation limits (harness precedent: pi-subagents rejections bounded to 4096 bytes).
6. **[GAP] Aggregate-vs-first-failure** is implied ("collects every failed ... rule" today) but never stated as a requirement for the merged report.
7. **[GAP] Feeds lint?** `lint --json` envelope is frozen (N5) and per §13 "all issues flow into the existing `LintIssue` envelope unchanged" — but S1 never says whether the merged report shares a type with `LintIssue` or is write-only. `SYNTHESIS-openness.md:31` + `command-inventory.md:146` argue for one shape.
8. **Unpinned severity vocabulary** (`openness-lint-engine.md:54-55`): schema violations have no severity; CEL validations are always hard errors (`akb-recon.md:45`); only `lint_rules` carry severity. Report must fix the mapping.
9. **Exit-code interaction:** `classifyExit` (`cmd/akb/main.go:17-23,68-90`) — validation → 1; compile error → 2 (`internalError`); usage → 2. Widening must not change these.
