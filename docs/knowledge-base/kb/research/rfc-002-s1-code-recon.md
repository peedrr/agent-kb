---
as_of: "2026-10-07"
created: "2026-10-07"
grammar: 1
informs:
- SPEC-002
is_draft: false
provenance: agent-drafted
run: rfc-002-s1
scope:
- cmd/akb/**
- internal/cel/**
- internal/template/**
status: final
summary: "Read-only recon of akb's validation error machinery: runTemplateValidations aggregation, ValidationError, write-path sentinels, and classifyExit, with file:line anchors and risks."
tags:
- validation
- code-recon
- spec-002
- akb
title: Code recon — current validation error/report machinery (RFC-002-S1)
type: research
updated: "2026-10-07"
---
# Code Context — current validation error/report machinery (RFC-002 S1 recon)

READ-ONLY. Repo `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb`. All anchors are
`file:line`. Scope: the "structured validation report (widened `runTemplateValidations`)".

## 1. `runTemplateValidations` — signature, aggregation, call sites

`cmd/akb/write.go:603` (body to `:644`):
```go
func runTemplateValidations(celEnv *gocel.Env, tmpl template.Template, page, oldPage map[string]any) error {
	var validationErrors []cel.ValidationError
```
- Aggregation: loops `tmpl.Validations`; each rule → `cel.CompileRule` then `cel.Evaluate`.
  - Compile failure → **returns immediately** as an akb fault: `return &internalError{err: fmt.Errorf("CEL engine error: compile rule %s: %w", rule.ID, err)}` (`write.go:607`).
  - Eval failure → appends `cel.ValidationError{RuleID, Message: unevaluableRuleMessage(...), Line: 0, Severity: "error"}` (`write.go:620`), `continue` (does NOT abort).
  - `result != types.True` → appends `cel.ValidationError{RuleID, Message: rule.Expect, Line: 0, Severity: "error"}` (`write.go:629`).
- Emission (`write.go:638-644`): if no errors `return nil`; else `for _, ve := range validationErrors { fmt.Fprintln(os.Stderr, ve.Error()) }` then `return validationFailure{}`.
- **All-errors semantics:** every false/unevaluable rule is collected and printed (all-errors), but the FIRST *compile* failure aborts early (fault path). `Line` is recorded as `0` here; `Severity` is always `"error"`.
- Helper `unevaluableRuleMessage` at `write.go:653` — special-cases `cel.ErrComputeBudget` ("too expensive…simplify") vs. other ("guard optional frontmatter keys with has()").
- Return type is the unexported sentinel `validationFailure` (defined `cmd/akb/main.go:53`), NOT a structured type.

Call sites (exhaustive, non-test):
- `cmd/akb/write.go:529` — in `runWrite`, after required-field gate, before storage write.
- `cmd/akb/append.go:225` — in `runAppend`, after required-field gate, before `store.WriteWithCommitMsg`.
- template write/mockup path does **not** call it: `cmd/akb/templates_write.go` uses a separate `evaluateValidations` (`templates_write.go:363`).

## 2. `internal/cel` — ValidationError, error conversion, aggregation

`internal/cel/errors.go:11`:
```go
type ValidationError struct {
	RuleID   string
	Message  string
	Line     int
	Severity string // "error" for validations, "warning" for lint_rules
}
```
`internal/cel/errors.go:18` `Error()`:
```go
if e.Line > 0 { return fmt.Sprintf("[%s] %s (near line %d)", e.RuleID, e.Message, e.Line) }
return fmt.Sprintf("[%s] %s", e.RuleID, e.Message)
```
- **No JSON tags** on `ValidationError`; it is plain (kind = struct, not an `error` interface impl of note — it does implement `error`).
- Errors → ValidationErrors: `CompileRule` (`internal/cel/engine.go:41`) returns wrapped `error` (never a ValidationError); `Evaluate` (`engine.go:64`) returns `(ref.Val, error)`, translating panic/`EvalCancelledError` into `ErrComputeBudget` (`engine.go:26`). The caller (`runTemplateValidations`) is the ONLY place that maps these into `cel.ValidationError`. `CompileRule` compile errors are NOT converted — they become `internalError` faults.
- Aggregation happens in the caller, not in `internal/cel` — there is no package-level "validate all rules" function; the `[]cel.ValidationError` slice is built in `cmd/akb/write.go`.
- Lint equivalent: `internal/lint/cel.go:47` `CELLintChecker.Check` builds `[]LintIssue` per page (eval error → `Severity:"error"`, false → `Severity: rule.Severity`), never a `ValidationError`.

## 3. Write-path pipeline order + exact failure text + exit codes

`cmd/akb/write.go` `runWrite` order (create/new-page branch shown; `--append`, `--frontmatter` branches mirror it):
1. `path.ResolveKB` (`:92`), `db.OpenKB` (`:98`), `config.Load` (`:110`).
2. Load templates (`:117-125`); `storage.OpenStore` (`:130`); `cel.NewEnv` (`:139`).
3. Per-branch path guards, `frontmatter.ValidateType`/`ValidateTitle`, `cel.BuildOldPage` (`:337`/`:472`), stamp `updated`, resolve `fullPath`/`relPath`, acquire `store.Lock` (`:463`), build `page := cel.BuildPage(...)` (`:518`).
4. **Required-field pre-gate** (`write.go:522-526`):
   ```go
   if missing := checkRequiredFields(tmpl, fm); len(missing) > 0 {
       fmt.Fprintln(os.Stderr, requiredFieldsMessage(tmpl, missing))
       return validationFailure{}
   }
   ```
   `checkRequiredFields`/`requiredFieldsMessage` in `cmd/akb/required_fields.go:18,32`. Message: `missing required frontmatter field(s): %s (declared required by template %q)`.
5. **CEL validation** `runTemplateValidations` (`:529`) → returns `validationFailure{}`; every failed rule already printed to stderr as `[rule_id] message`.
6. **Storage write/commit**: `store.Write(...)` (`:546`) → IndexPage tx (`:572`) → UpdatePageLinks tx (`:586`) → `tx.Commit()` (`:596`). Commit-conflict/identity results surface as `usage:`/`Error:` via `classifyExit`.
7. Output on success: `Written to %s` + index reminder (`:604-610`).

Exit codes: validation failures → `validationFailure` → **exit 1**, `classifyExit` returns empty report (already printed). CEL engine/compile/env/old_page faults → `internalError` → **exit 2** `internal:`.

## 4. Exit-code contract — `classifyExit` (`cmd/akb/main.go:78`)

```go
const ( exitSuccess = 0; exitFailure = 1; exitFault = 2 )  // main.go:27
```
Mapping (`main.go:91`):
- `err == nil` → `(0, "")`.
- `errors.As(err, &validationErr) | &driftErr` → `(1, "")` — silent, command already reported.
- `usageError | *path.GuardError | storage.ErrNoCommitIdentity | storage.ErrMergeInProgress` → `(2, "usage: "+commandMessage(err))`.
- `*internalError` → `(2, "internal: "+internalErr.Error())`.
- default → `(1, "Error: "+err.Error())` (e.g. git errors).
Sentinel types: `usageError` (`main.go:33`), `internalError` (`main.go:42`), `validationFailure` (`main.go:53`), `driftDetected` (`main.go:59`), `commandFailure` (`main.go:67`). `main()` (`main.go:108`) is sole reporter; Cobra `SilenceErrors/SilenceUsage = true`.

## 5. `template write` mockup validation + `<!-- FAILS: rule_id -->`

`cmd/akb/templates_write.go` `runTemplatesWrite`:
- Compile gate for `tmpl.Validations` and `tmpl.LintRules` (`:111-118`): `compile validation rule %q: %w`.
- Required-field gate on PASS mockup (`:150`): `pass mockup %s\n\n--- pass mockup ---\n%s\n\nProvide updated mockup with --pass <path>`.
- PASS validations via `evaluateValidations` (`:159`, def `:363` — returns `([]string failedIDs, string ruleID, error)`):
  - `pass mockup: %w` on eval error (`:161`).
  - `pass mockup no longer validates: %v\n\n--- pass mockup ---\n...` (`:164`) where `%v` is the `[]string` of rule IDs (e.g. `[title_not_hello]` — asserted `templates_write_test.go:232`).
  - Optional-key-strip loop (`:171-179`): `pass mockup no longer validates without optional key %s: %v`.
  - Self-`old_page` check (`:183-189`): `pass mockup no longer validates as an update of itself: %v`.
- FAIL mockup (`:216-221`): only asserts `len(failFailed) > 0`; message `fail mockup no longer validates: expected at least one validation to fail, but all passed\n\n--- fail mockup ---\n...`.
- `<!-- FAILS: rule_id — … -->` convention: **documentary only, never machine-parsed.** `grep -rn "FAILS" --include=*.go .` = **0 hits**. Examples: `internal/template/embedded/adr_fail.md:16`, `embedded/note_fail.md:8`. The FAIL mockup passes iff at least one rule is false; the comment is not read.
- `akb template get --example` reuses the same variant proofs as stderr WARNINGS via `warnMockupValidation`/`warnMockupVariant` (`cmd/akb/template.go` ~`:188-255`) — `WARNING: Could not validate mockup: CEL error in rule '%s': %v` and a headline + failed IDs.

## 6. `lint --json` envelope (machine-readable precedent)

`internal/lint/engine.go:17,28`:
```go
type LintIssue struct {
	Type string `json:"check"`; RuleID string `json:"rule_id,omitempty"`
	Message string `json:"message"`; Path string `json:"path"`; Severity string `json:"severity"`
}
type LintReport struct {
	Issues []LintIssue `json:"issues"`; PagesChecked int `json:"pages_checked"`; ByCheck map[string]int `json:"by_check"`
}
```
`cmd/akb/lint.go:220` `printLintJSON` wraps it in a local `jsonOutput` with a `summary` object:
```go
type jsonOutput struct { Issues []lint.LintIssue `json:"issues"`; Summary struct{ Total int `json:"total"`; PagesChecked int `json:"pages_checked"`; ByCheck map[string]int `json:"by_check"` } `json:"summary"` }
```
`enc.SetIndent("", "  ")`; nil Issues → `[]`. Both text and JSON return `errLintIssues` (→ exit 1) when any issue has `Severity=="error"` (`lint.go:200,246`). Text format (`printLintText`, `lint.go:164`): `=== path ===` / `  [type] rule_id=<id> severity: message`. This is the only structured report precedent in the repo.

## 7. JSON/machine-readable output for write-time validation failures?

**Confirmed: none.** `grep -rn "json" cmd/akb/write.go cmd/akb/append.go` = 0 hits; no `--json` flag on `write`/`append`; `validationFailure` carries no payload (empty struct, `Error()` = `"page validation failed"`); `cel.ValidationError` has no JSON tags. Write-time failures are stderr text + exit code only.

## 8. Other error-report facts a widened model must preserve

- **Sentinel identity is load-bearing:** `validationFailure` is matched two ways — `classifyExit` (`main.go:82`) and `approveAllDraftPages` (`cmd/akb/approve.go:265` uses `errors.As(err, &validationErr)` to treat a refused draft as non-fatal-per-page). A widened report must keep `errors.As`-matchable sentinel semantics or update BOTH sites.
- **`approve.go` reuses write-path wording:** `requiredFieldsRefusal` (`approve.go:220`) prints `cannot approve '%s': <requiredFieldsMessage>` and `return validationFailure{}` (`:233`), explicitly "so the wording and the exit code match `akb write`".
- **Fault vs. result split:** CEL engine faults (compile failure, `NewEnv`, `BuildOldPage`) are `internalError` → exit 2; content failures are exit 1. Widening must not silently fold faults into results.
- **Print-then-return contract:** the command prints ALL messages to stderr and returns the bare sentinel; `classifyExit` reports nothing further (`report == ""`). A structured report must preserve "already reported" so stderr isn't doubled.
- **`Line`/`Severity` already exist but are unused at write time** (`Line:0`, `Severity:"error"`); `Severity:"warning"` is used only by lint_rules.
- **Lint vs. write divergence is deliberate** (`write.go:598-602` comment): write blocks on unevaluable rules; lint degrades to a per-page `cel_lint` issue. Preserve.
- `internal/lint/required_fields.go:67` duplicates the required-field message into a `LintIssue`, so a JSON schema change to that wording touches two packages.

## Gaps / risks

- **No shared report type:** write-time errors live only as stderr strings; there is no `ValidationReport` struct to widen — the spec must introduce one and thread it through `runTemplateValidations` → `runWrite`/`runAppend` while keeping the `validationFailure` sentinel and `classifyExit` exit-1/silent contract.
- **Compile-error early-return** means a widened "report all" model cannot simply collect compile errors too without changing the akb-fault (exit 2) semantics.
- **FAILS convention has no parser** — extending it to `<!-- FAILS: schema: ... -->` requires new machinery (ADR-006 N7 keeps it unparsed until then); a widened mockup report that consumes it is net-new.
- **`evaluateValidations` vs `runTemplateValidations` divergence:** template-write returns bare rule-ID strings (no messages/line/severity), write returns `cel.ValidationError`. Unifying them is a design choice the spec must make explicit.
- **JSON tags absent** on `cel.ValidationError`/`validationFailure`; any JSON output is additive and must not break lint's existing `--json` shape.
