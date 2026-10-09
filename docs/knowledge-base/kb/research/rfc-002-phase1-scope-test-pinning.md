---
as_of: "2026-10-09"
created: "2026-10-09"
grammar: 1
informs:
- ADR-008
- ADR-009
- ADR-010
is_draft: false
provenance: agent-drafted
run: rfc-002-phase1-scope
scope:
- test/**
- internal/**
- cmd/**
- go.mod
status: final
summary: Recon of flatten-vs-current-shape test churn, goccy frontmatter value types, the complete BuildPage/BuildOldPage call-site list, and the approve-wording testscript coverage gap.
tags:
- rfc-002
- phase-1
- recon
- test-churn
title: Phase-1 scope recon — test pinning, frontmatter types, call sites
type: research
updated: "2026-10-09"
---

# RFC-002 Phase-1 Follow-up Recon — Test Pinning, Frontmatter Types, BuildPage Call Sites

Repo: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb` @ `a316f0a` (v0.23.1, main). Read-only recon.

---

## Q1 — TEST-PINNING INVENTORY (flatten / V3 migration churn)

Method: `grep -rn` over `test/testdata/*.txt` and `*_test.go` in `internal/cel/`, `cmd/akb/`, `internal/template/`, `internal/lint/`; counts are command invocations / test funcs / assertion lines as noted.

### 1A. `word_count` / `char_count` — the `page.content` projection

| Location | Pins | Count |
|---|---|---|
| `internal/cel/pagebuilder_test.go:322-349` `TestBuildPageWordCount` | `content.word_count == 5`, `content.char_count == len(body)` | 1 test func, 2 asserts |
| `internal/cel/pagebuilder_test.go:100-110` `TestBuildPage` | `content.raw == string(body)`, `content.char_count == len(body)` | 1 test func, 2 asserts |
| `internal/lint/cel_test.go:94` (`TestCELLintCheckerEvaluationErrorDoesNotAbortSweep`) | rule fixture `page.content.word_count > 5` | 1 |
| `internal/lint/cel_test.go:222` (`TestCELLintChecker`) | rule fixture `page.content.word_count > 10` | 1 |
| `test/testdata/exit_codes.txt:40` | `stderr 'min_word_count'` (rule id in all-rules-fail ADR write) | 1 positive assertion |
| `test/testdata/superseded_adr.txt:65` | `! stderr 'min_word_count'` (negative: no extra rule reported) | 1 negative assertion |
| `internal/template/embedded/adr.yaml:117-118` | template rule `page.content.word_count >= 50` | definition (not a test) |
| `docs/knowledge-base/.agent-kb/templates/{adr,spec,rfc,research}.yaml` | 4 more `min_word_count` rule copies | definitions |

`min_word_count` is the rule **id**; the body is `page.content.word_count >= 50` (`internal/template/embedded/adr.yaml:118`).

### 1B. `content.raw`

| Location | Pins | Count |
|---|---|---|
| `internal/cel/pagebuilder_test.go:104-105` | `content.raw == string(body)` | only repo-wide hit |
| `internal/template/embedded/adr_pass.md:65` | prose documenting `page.content.word_count` as the body metric | doc |

### 1C. `page.ast`

| Location | Pins | Count |
|---|---|---|
| `internal/cel/pagebuilder_test.go:113,434,894` (+ 8 flatten-links tests at 467,604,618,628,643,674,700,742,772,811,843,923,988) | `page["ast"]` map shape: `headings[].{level,text,line}`, `links[].{target,text,is_wikilink,line}`, `code_blocks[].{language,line}` | 21 test funcs in file |
| `internal/cel/engine_test.go:26` (`TestNewEnv`) | compiles `page.ast.headings.exists(h, h.level == 2)` | 1 |
| `internal/cel/engine_test.go:154` (`tripleHeadingsRule`, used by `TestCompileRuleCostLimit`) | `page.ast.headings.exists(...)` | 1 |
| `cmd/akb/write_test.go:1119` (`TestWriteComputeBudgetRuleGetsBudgetRemedy`) | `page.ast.headings` cost-limit rule fixture | 1 |
| `cmd/akb/template_test.go:93` | `page.ast.headings.exists(h, h.level == 2 && h.text == "Context")` template fixture | 1 |
| `test/testdata/cel_validation.txt:167` | custom template rule `page.ast.headings.exists(h, h.level == 1)` | 1 testscript |
| `test/testdata/wikilink_forms.txt:146` | `page.ast.links.exists(l, l.is_wikilink && l.target == page.frontmatter.expect_target)` | 1 testscript |

### 1D. `page.akb`

- Literal `page.akb` / `page["akb"]` **zero hits** repo-wide (`.go`/`.txt`/`.yaml`, excluding `.sisyphus/`).
- The `page["akb"]` map is pinned **structurally only** in `internal/cel/pagebuilder_test.go:153-171` (`TestBuildPage`) and `:196` (`TestBuildPageAnnotationExclusions`) — asserts `provenance_markers` and `annotations` are `[]any` of length 1. No CEL rule reads `page.akb`, so a flatten would only break these two structural asserts.

### 1E. `missing required frontmatter field` / `required_fields`

| Location | Pins | Count |
|---|---|---|
| `cmd/akb/write_test.go:2165,2194,2228` | exact wording `missing required frontmatter field(s): … (declared required by template "adr")` | 3 funcs |
| `cmd/akb/approve_test.go:673,725` | same wording for `akb approve` refusal | 2 funcs |
| `cmd/akb/append_test.go:770` | same wording for `akb append` | 1 |
| `cmd/akb/templates_write_test.go:1192` | same wording for pass-mockup refusal (`template "exemplar"`) | 1 |
| `internal/lint/required_fields_test.go:87,106,159` | `required_fields` lint message wording | 3 |
| `test/testdata/lint_required_fields.txt:28` | exact regex of the lint message `…: deciders, status (declared required by template "adr")` | 1 testscript |
| `test/testdata/lint_tests_adr_write.txt:18` | `^missing required frontmatter field\(s\): status \(…\)$` | 1 testscript |

Message source (single definition, two callers): `cmd/akb/required_fields.go:34` (`requiredFieldsMessage`) reused by `cmd/akb/approve.go:221` (`requiredFieldsRefusal`, wording parity contract) and `internal/lint/required_fields.go:67`.

### 1F. `Template format has changed` (V2 loader guard)

| Location | Pins | Count |
|---|---|---|
| `internal/template/template.go:72-80` `detectOldFormat` (keys `required`/`optional`/`body`) | message source, line 77 | def |
| `internal/template/template_test.go:163,184,205` | 3 subtests: old `required`, `optional`, `body` each refused | 3 |
| `test/testdata/cel_old_format.txt:17,22,29` | lint, write, and `template write` all refused | 1 scenario, 3 command asserts |

### 1G. `template write` mockup-flow testscript files

| File | `template write` invocations |
|---|---|
| `test/testdata/template_write_noop.txt` | 4 |
| `test/testdata/cel_validation.txt` | 2 |
| `test/testdata/template_examples.txt` | 1 |
| `test/testdata/cel_old_format.txt` | 1 |

Go-side mockup coverage: `cmd/akb/templates_write_test.go` (24 funcs, incl. `TestTemplatesWrite_RejectsPassFailsRules:172`, `..._RejectFailPassesRules:244`, `..._RejectsPassMockupMissingRequiredField:1154`, `..._FailMockupIsExemptFromRequiredFields:1211`) and `cmd/akb/template_test.go` (~21 funcs, esp. `--example` variant warnings at 291-575).

### 1H. `approve` refusal wording testscript coverage

- **Zero testscript hits** for `cannot approve` (`grep -rn "cannot approve" test/testdata/` → none).
- Wording parity is pinned **only in Go**: `cmd/akb/approve_test.go:651-748` (`TestApproveRefusesDraftMissingRequiredField:651`, `TestApproveAllDraftsRefusesOnlyIncompleteDrafts:694`).
- **Gap/risk:** the `approve.go:221` wording-parity contract has no integration-test (testscript) guard; only unit tests.

### 1I. Churn cost table — flatten-now vs validate-current-shape

| Area | Pin count (Go / testscript) | Flatten-now churn | Validate-current-shape churn |
|---|---|---|---|
| `page.content` (`raw`/`word_count`/`char_count`) | 4 funcs / 2 asserts | Rewrite `TestBuildPage`, `TestBuildPageWordCount`, 2 lint rule fixtures, 2 testscript rule asserts; re-path 5 template `min_word_count` rules | ~0 (add new tests) |
| `page.ast` (`headings`/`links`/`code_blocks`) | 21 funcs / 2 testscript | Rewrite 21 pagebuilder funcs + 3 Go rule fixtures + 2 testscript rules | ~0 |
| `page.akb` (markers/annotations) | 2 structural asserts / 0 | Rewrite 2 asserts only | ~0 |
| `missing required frontmatter field` wording | 10 funcs / 2 testscript | Unaffected if the required-fields check stays a pre-CEL gate (it does not read the page map) | ~0 |
| `Template format has changed` (V2 loader) | 3 subtests / 3 asserts | Unaffected unless V3 replaces V2 loading | Unaffected |
| `template write` mockup flows | 24+21 funcs / 8 invocations | Only if `buildTestPage` shape changes | ~0 |
| `approve` refusal wording | 2 funcs / **0** | Unaffected | Unaffected |

**Bottom line:** flatten-now concentrates ~27 test funcs + 4 testscript rule assertions + 5 template rule re-paths into the `page.content`/`page.ast` rewrite; validate-current-shape costs ≈0 churn and only adds new coverage. The `required_fields` and old-format guards are orthogonal to the page-map shape.

---

## Q2 — FRONTMATTER PARSE TYPES

Read `internal/frontmatter/frontmatter.go` (167 lines) in full; decode behavior verified empirically against the pinned module (`github.com/goccy/go-yaml v1.19.2`, `go.mod:11`) via a scratch module outside the repo.

### Exact decode call and target types

- `internal/frontmatter/frontmatter.go:26-30` — the format's unmarshaler is goccy:
  ```go
  yamlFmt := frontmatter.Format{ Name: "YAML", Delim: '-', Unmarshal: yaml.Unmarshal }
  ```
  (`yaml` = `github.com/goccy/go-yaml`, import `frontmatter.go:13`.)
- `internal/frontmatter/frontmatter.go:44-46`:
  ```go
  var raw map[string]any
  if err := data.Decode(&raw); err != nil { … }
  ```
  `data` is `go.abhg.dev/goldmark/frontmatter`'s `Data` (`frontmatter.go:44`); `Decode` dispatches to the `Format.Unmarshal` above.
- Resulting types: `ParsedFrontmatter{ Type string; Title string; Fields map[string]any }` (`frontmatter.go:18-22`); `type`/`title` are string-asserted (`frontmatter.go:63-75`), all other keys land verbatim in `Fields` (`frontmatter.go:76`).

### Empirically observed value types (goccy v1.19.2)

| YAML input | Go type produced |
|---|---|
| `created: 2026-01-01` | `string` |
| `stamp: 2026-01-01T10:00:00Z` | `string` |
| `when: "2026-01-01"` (quoted) | `string` |
| `a: 2026-01-01 10:00:00`, `b: 2026-1-2` | `string` |
| `b: !!timestamp 2026-01-01` (explicit tag) | `time.Time` (`decode.go:380-382`) |
| `count: 42`, `0x10` | `uint64` |
| `a: -5` | `int64` |
| `ratio: 1.5` | `float64` |
| `e: 1e3` | `string` (not float) |
| `tags: [go, testing]` | `[]any` (elements recursively decoded) |
| `nested: {a: 1, b: {c: x}}` | `map[string]any` (never `map[interface{}]interface{}`) |
| `objlist: [{name: a, n: 1}]` | `[]any` of `map[string]any` |
| `flag: true` / `nothing: null` | `bool` / `nil` |

Confirmation in the pinned module: `nodeToValue` maps `*ast.TagNode`+`TimestampTag` → `time.Time` (`decode.go:380-382`), `*ast.IntegerNode`/`*ast.FloatNode` → native (`decode.go:352-359`), mapping nodes → `map[string]interface{}` (`decode.go:492,516`), interface targets go through `nodeToValue` (`decode.go:938-948`). Plain date scalars carry **no** timestamp tag (tag only arises from an explicit `!!timestamp`), so they stay strings — which is exactly why `internal/cel/pagebuilder.go:29-47` `convertDateField` exists as an explicit string→`time.Time` normalization.

### Implication for a JSON-Schema document view

- **Structural normalization NOT needed:** `map[string]any` and `[]any` are directly JSON-encodable; no `map[interface{}]interface{}` cleanup pass.
- **Value normalization IS needed** if the schema is to be strict:
  1. date strings → `format: date-time`/`date` (must reuse `convertDateField` logic or equivalent);
  2. integers arrive as `uint64` (positive) or `int64` (negative) — a schema expecting `integer` must accept both; JSON has no unsigned type;
  3. `1e3` stays a string, so numeric coercion is not guaranteed.
- The CEL page map already applies (1) via `convertDateField`; a document view built from **raw `ParsedFrontmatter.Fields`** would need the same pass to match CEL's view.

---

## Q3 — BUILDPAGE / BUILDOLDPAGE CALL SITES (complete)

`grep -rn "BuildPage\|BuildOldPage" --include="*.go"` (non-test), plus test callers.

### Direct production callers

| Site | Call | What the caller does with the result |
|---|---|---|
| `cmd/akb/write.go:518` | `cel.BuildPage(relPath, fm, body, astDoc, body)` | Builds the **post-modification** `page` map for `akb write` (create/overwrite/frontmatter paths) and passes it to `runTemplateValidations` (`write.go:529`). |
| `cmd/akb/write.go:210` | `cel.BuildOldPage(relPath, store)` | `--frontmatter` update path: `old_page` from the on-disk page before mutation. |
| `cmd/akb/write.go:336` | `cel.BuildOldPage(relPath, store)` | append-inside-write path: `old_page` pre-bump. |
| `cmd/akb/write.go:477` | `cel.BuildOldPage(relPath, store)` | overwrite path: `old_page` from on-disk state (guarded by `os.Stat`, comment `write.go:459`). |
| `cmd/akb/append.go:173` | `cel.BuildOldPage(relPath, store)` | `akb append`: pre-append `old_page`. |
| `cmd/akb/append.go:210` | `cel.BuildPage(relPath, fm, newBodyBytes, astDoc, newBodyBytes)` | post-append `page` map for validation via `runTemplateValidations` (`append.go:225`). |
| `cmd/akb/templates_write.go:355` | `cel.BuildPage(...)` inside `buildTestPage` | builds mockup `page` maps for `akb template write` rule validation. |
| `internal/lint/cel.go:62` | `cel.BuildPage(page.RelPath, page.Frontmatter, page.Body, doc, page.Body)` | builds the per-page map for the `cel_lint` sweep (`page` var only; `old_page` deliberately not injected). |

### Indirect callers (wrappers)

- `buildTestPage` — `cmd/akb/templates_write.go:352` (wraps `BuildPage`). Call sites: `templates_write.go:157,171,215` and `cmd/akb/template.go:213,218`.
- `evaluateValidations` — `cmd/akb/templates_write.go:363` (consumes page maps; not a BuildPage wrapper). Call sites: `templates_write.go:159,172,183,216`.
- `runTemplateValidations` — `cmd/akb/write.go:603` (consumes `BuildPage` output). Call sites: `write.go:529`, `append.go:225`.
- `BuildOldPage` → `BuildPage` — internal delegation at `internal/cel/pagebuilder.go:115`.
- `warnMockupValidation` / `warnMockupVariant` — `cmd/akb/template.go:200-225` drive `buildTestPage` for `akb template get --example` warnings.

### Test callers

- `internal/cel/pagebuilder_test.go` — 21 funcs; `BuildPage` at `:68,194,257,340,892,988`; `BuildOldPage` at `:383,407`.
- Rule-compilation fixtures referencing the produced map: `internal/cel/engine_test.go:26,154`; `cmd/akb/write_test.go:1119`; `cmd/akb/template_test.go:93`.

**Single chokepoint:** every production `page` map is produced by `cel.BuildPage` (`internal/cel/pagebuilder.go:51`); every `old_page` map by `cel.BuildOldPage` (`:91`). Changing the page-map shape touches only these two functions plus the 8 production call sites and the wrappers listed above.

---

## SUMMARY

1. **Flatten churn is concentrated in `page.content` + `page.ast`**: ~27 Go test funcs (21 in `internal/cel/pagebuilder_test.go`) + 4 testscript rule assertions (`cel_validation.txt:167`, `wikilink_forms.txt:146`, `exit_codes.txt:40`, `superseded_adr.txt:65`) + 5 template `min_word_count` rule re-paths; validate-current-shape costs ≈0 test churn.
2. **`page.akb` has no rule consumers** — only 2 structural asserts in `pagebuilder_test.go:153,196`; the required-fields gate and V2 old-format guard are orthogonal to the page-map shape.
3. **Frontmatter types:** goccy leaves date-like scalars as `string` (only `!!timestamp` → `time.Time`), ints as `uint64`/`int64`, floats `float64`, maps `map[string]any`, lists `[]any` — so a JSON-Schema document view needs a **value** normalization pass (dates, int signedness) but **no structural** pass.
4. **`cel.BuildPage`/`BuildOldPage` are the sole chokepoints** (`internal/cel/pagebuilder.go:51,91`) with 8 production call sites across `write.go`, `append.go`, `templates_write.go`, `lint/cel.go` and 3 wrapper layers (`buildTestPage`, `evaluateValidations`, `runTemplateValidations`).
5. **Coverage gap:** the `approve.go:221` refusal-wording parity contract has no testscript guard — only `cmd/akb/approve_test.go:651,694`.
