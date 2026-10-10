---
as_of: "2026-10-10"
created: "2026-10-10"
grammar: 1
informs:
- SPEC-004
is_draft: false
provenance: agent-drafted
run: rfc-002-s7
scope:
- cmd/akb/**
- internal/cel/**
- internal/lint/**
- internal/template/**
- internal/frontmatter/**
- internal/storage/**
- test/**
status: final
summary: "Current-tree attachment points for the SPEC-004 schema pass: write/append pipeline order, document shape, pre-gate anatomy, lint registry model, test fallout, deltas vs prior recon."
tags:
- rfc-002
- code-recon
- spec-004
- schema-pass
title: RFC-002-S7 code recon — write/append pipeline, document construction, pre-gate, lint attachment
type: research
updated: "2026-10-10"
---

# RFC-002-S7 Code Recon — where the schema pass attaches

Repo `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb` @ `c8bac2e` (v0.23.1, clean tree).
Read-only. All anchors are `file:line` in the current tree.

Prior pages read first and NOT re-derived here (cite, don't repeat):
`rfc-002-s1-code-recon.md` (validation error machinery, `classifyExit`, sentinels),
`rfc-002-phase1-scope-code-surface.md` (page-map shape, rule consumers, loader, lint interface),
`rfc-002-phase1-scope-test-pinning.md` (test churn table, goccy value types, BuildPage call sites),
`rfc-002-s6-code-recon.md` (TemplateV2 loader, key→consumer inventory, type/title routing).
Ratified bounds: `decisions/ADR-008-phase1-current-document-shape.md` (Phase 1 validates the *current*
shape; no flatten), `ADR-009-phase1-sweep-schema-checker.md` (built-in `schema` checker, before
`cel_lint`, shared `now`, `required_fields` kept), `ADR-010` (six V3 keys).

---

## 0. Deltas vs the prior recon pages (verify-first findings)

| # | Prior claim | Current tree | Note |
|---|---|---|---|
| D1 | `required_fields.go:18` / `:19-29` | `checkRequiredFields` func at **:19**; doc comment :16-18; `requiredFieldsMessage` at **:33** | Anchor in the S7 brief (`:18`) names the comment line. No functional change. |
| D2 | `frontmatterKeyPresent` at `templates_write.go:410-419` | func at **:413-421** (comment :410-412) | Trivial drift. |
| D3 | `BuildPage` `pagebuilder.go:52-85`, `BuildOldPage :87-117` | `BuildPage` **:51-84**, `BuildOldPage` **:91-117** | Trivial drift; key anchors inside unchanged (`content` :69-73, `ast` :75-79, `akb` :81-83). |
| D4 | required-gate `write.go:522-526`, `append.go:220-222` | `write.go:521-526` (comment :521-522, `if` :523), `append.go:218-222` (comment :218-219, `if` :220) | Trivial drift. |
| D5 | **NEW** — not in prior recon | **The merge-conflict check is NOT early in write/append.** `runWrite`/`runAppend` never call `store.Preflight()`; the merge + commit-identity preflight runs *inside* `store.Write` (`mode.go:155-159`) at `write.go:547` / `store.WriteWithCommitMsg` (`mode.go:165-170`) at `append.go:230` — i.e. **after** all validation. Other mutating commands call `store.Preflight()` explicitly *before* their first mutation (`raw_write.go:93`, `index.go:146,297,352`, `log.go:148`, `delete.go:208`, `template_write.go:289`, `template_delete.go:116`, `raw_sync.go:150`). | RFC §7's `merge-conflict check → …` reorder needs an explicit `store.Preflight()` added to `runWrite`/`runAppend`; otherwise merge/identity faults stay *after* schema/CEL validation and precedence changes (exit 2 merge/identity only if validation passes). No prior page flags this. |
| D6 | **NEW** | `page["frontmatter"]` is **date-coerced** (`convertDateField`, `pagebuilder.go:29,63`): a string that parses as RFC3339 or `2006-01-02` becomes `time.Time`. ADR-008 I1 + ADR-005 N3 want the *document* view to hold **raw** YAML values. | The schema document must be built from `fm.Fields`/`fm.Type`/`fm.Title` **without** `convertDateField` — do not alias `page["frontmatter"]`. See §2. |
| D7 | **NEW** | `now` is injected **per caller**, not shared: write uses `time.Now()` per rule evaluation (`write.go:616`); `cel_lint` uses one `now := time.Now()` per `Check` call (`cel.go:44`). | ADR-009 I2 ("schema pass and CEL lint pass share one `now` stamp") has no mechanism today; SPEC-004 must thread a single clock (e.g. into `KB`/checker ctor). |
| D8 | s6: loader non-strict | confirmed: `template.go:106,114`, no `DisallowUnknownField`; `go.mod` has **no** `santhosh-tekuri/jsonschema` (grep ABSENT) | The schema dep is net-new in `go.mod`. |

---

## 1. Write/append pipeline as-is, with insertion points

### 1a. `akb write` — `runWrite` (`cmd/akb/write.go:89-585`)

Three branches converge on one tail. Setup is shared:

| Step | Anchor |
|---|---|
| `path.ResolveKB` | `write.go:98` |
| `db.OpenKB` | `write.go:103` |
| `config.Load` | `write.go:112` |
| templates dir assertions + `template.LoadTemplates` | `write.go:117-127` (`LoadTemplates` :125) |
| `storage.OpenStore` | `write.go:133` |
| `cel.NewEnv` | `write.go:139` |

Branch A — `--frontmatter` update (`:145-290`): lock `:182` → `resolveExistingPage` `:196` → `frontmatter.Parse` `:197` → `ValidateType` `:202` → `ValidateTitle` `:205` → `cel.BuildOldPage` `:210` → apply key=val `:217-241` → `ValidateType` again `:243` → re-serialize `:252-262`.

Branch B — `--append` in-file (`:297-360`): lock `:305` → resolve/parse `:320` → `ValidateType` `:326` → `ValidateTitle` `:331` → `cel.BuildOldPage` `:336` → stamp `updated` `:352` → re-serialize `:354-360`.

Branch C — create/overwrite from stdin (`:361-506`): `frontmatter.Parse` `:367` → `ValidateType` `:373` → `ValidateTitle` `:378` → `created` default `:391`, `is_draft` strip `:396` → **selector gate** `tmpl, ok := templates[fm.Type]` `:399-403` → path guards + type-dir resolution `:405-455` → lock `:462` → `os.Stat` `:471` → `cel.BuildOldPage` `:477` → stamp `updated` `:484-487` → re-serialize `:490-500`.

Converged tail (all branches):
1. **Selector gate (second time)** `tmpl, ok := templates[fm.Type]` — `write.go:507-510`; unknown → `unknown type %q`.
2. **goldmark parse + document/page build** — `write.go:513-518`:
   ```go
   md := goldmark.New()
   astDoc := md.Parser().Parse(text.NewReader(body))
   page := cel.BuildPage(relPath, fm, body, astDoc, body)
   ```
3. **Required-field pre-gate** — `write.go:521-526` (`checkRequiredFields` → stderr + `validationFailure{}`).
4. **CEL validations** — `runTemplateValidations` `write.go:529` (def `:603-644`).
5. **Lock, then storage write/commit** — `store.Lock` `:539`, `store.Write(ctx, fullPath, writeContent)` `:547` (this call runs `Store.Preflight()` → merge check + identity, `mode.go:155-159`).
6. **search + linkgraph in one tx** — `:552-582` (`IndexPageTx` `:562`, `UpdatePageLinksTx` `:566`, `tx.Commit` `:575`).
7. stdout `:578-583`.

### 1b. `akb append` — `runAppend` (`cmd/akb/append.go:68-259`)

`ResolveKB` :84 → `OpenKB` :89 → `config.Load` :98 → guards :101-112 → `OpenStore` :118 → **repo lock up front** :127 → `resolveExistingPage` :136 (`typeDirsFromDisk`) → `frontmatter.Parse` :144 → `LoadTemplates` :160 → `ValidateType` :165 → `ValidateTitle` :168 → `cel.BuildOldPage` :173 → stamp `updated` :179 → re-serialize :186-198 → selector gate :203-207 → goldmark + `cel.BuildPage` :207-210 → **required pre-gate** :218-222 → `cel.NewEnv` :212 / `runTemplateValidations` :225 → `store.WriteWithCommitMsg` :230 (Preflight inside) → search/linkgraph tx :245-255 → stdout :257.

### 1c. Insertion points for SPEC-004

| What | Where | Shape |
|---|---|---|
| (a) **document construction** | `write.go:518` (right after `cel.BuildPage`, or beside it) and `append.go:210` | A new `buildDocument(relPath, fm, body, astDoc)` producing the raw-value document of ADR-008 I1. `fm`, `body`, `astDoc`, `relPath` are all in scope at both sites. Do NOT reuse `page["frontmatter"]` (D6). |
| (b) **schema pass** | `write.go:519-528` — between the build and the pre-gate/CEL; `append.go:211-224` | Runs after goldmark (schema may constrain `content.raw`/`ast.*`), before `runTemplateValidations`; returns the SPEC-002-shaped report → maps to the existing `validationFailure{}` sentinel for write/append. |
| (c) **pre-gate removal** | delete `write.go:521-526` and `append.go:218-222`. Keep `internal/lint/required_fields.go` (ADR-009 I5). `approve.go:151/221-234` retires too per ADR-009 (wording parity preserved). | `checkRequiredFields`/`requiredFieldsMessage` stay exported for `templates_write.go:152` (mockup obligations, SPEC-005/S8) until that path moves. |
| (d) **merge-conflict-first reorder** (D5) | add `store.Preflight()` in `runWrite` after `OpenStore` (:133) and in `runAppend` after `OpenStore` (:118); wrap as `fmt.Errorf("commit preflight: %w", err)` like the other commands | Pure reorder; nothing is written before it. |
| (e) **timeout/identity** | unchanged; `Preflight` covers both merge (`checkMergeState`) and identity (`resolveIdentity`) — `mode.go:111-129` | — |

**Ordering vs search/linkgraph/git:** validation (schema + CEL) always precedes `store.Write`/`WriteWithCommitMsg`; the index/link tx runs *after* the commit funnel, so a schema refusal writes nothing.

---

## 2. Document construction

### 2a. `BuildPage` (`internal/cel/pagebuilder.go:51-84`) — the sole `page`-map chokepoint

Produces `map[string]any` with exactly five top-level keys:

- `file` = `{path, name, dir}` — `:54-58` (from `relPath`).
- `frontmatter` = `fm.Fields` **each through `convertDateField`**, then `type`/`title` force-injected — `:60-66`.
- `content` = `{raw: string, word_count: int, char_count: int(bytes)}` — `:69-73`. `content` is an **object**, not a string.
- `ast` = `{headings, links, code_blocks}` — `:75-79`; element fields at `:139+` (`{level,text,line}`), `:200+` (`{target,text,is_wikilink,line}`), `:322+` (`{language,line}`).
- `akb` = `{provenance_markers, annotations}` — `:81-83`.

`BuildOldPage` (`:91-117`) reads on-disk bytes, parses FM, goldmark-parses, and **delegates to `BuildPage`** (`:116`); returns `(nil,nil)` when absent (`:96-98`).

### 2b. Raw vs derived

| Part | Source | Raw or derived |
|---|---|---|
| `frontmatter.*` | `fm.Fields` (+ `type`,`title` from dedicated fields) | **normalized** — dates coerced to `time.Time` (`:63`) |
| `content.raw` | `body` bytes | derived projection of the body |
| `content.word_count` / `char_count` | computed (`:71-72`) | derived |
| `ast.*`, `akb.*` | goldmark AST + markdown parsers (`:75-83`) | derived |

### 2c. What a JSON-Schema-validatable document is

Per **ADR-008 I1** the document is the current shape: `{frontmatter, content{raw,word_count,char_count}, ast{headings,links,code_blocks}, akb{...}, file{...}}`; per **ADR-008 I2** a Phase-1 `schema:` block references `/frontmatter/title`, `content.raw`, `ast.headings`. Per RFC-002 §5 (`open-validation-engine.md:187`) the `document` row is "**Raw frontmatter values as parsed from YAML**", and per **ADR-005 N3** coerced values never enter the document view.

**Consequence (D6):** the document's `frontmatter` must be built from `fm.Fields` + `fm.Type`/`fm.Title` **without** `convertDateField`. RFC §6/§7's worked example (`content: {type: string}` + `derived.ast`) is the *post-flatten* shape and is explicitly out of Phase-1 scope (ADR-008 N1/N2); Phase-1 schemas say `content.raw`, `ast.*`.

### 2d. YAML → JSON-typed values (goccy `v1.19.2`, `go.mod:11`)

Verified empirically by the test-pinning recon (`rfc-002-phase1-scope-test-pinning.md` §Q2, cite — not re-run here):
date-like scalars stay **`string`** (only explicit `!!timestamp` → `time.Time`); integers → **`uint64`** (positive) **or `int64`** (negative); floats → `float64`; `1e3` stays `string`; lists → `[]any`; maps → `map[string]any` (never `map[interface{}]interface{}`); bool/nil native.
Parse call: `frontmatter.Parse` (`internal/frontmatter/frontmatter.go:29-87`) decodes into `map[string]any` at `:44-46` via goccy (`:16`), routes `type`/`title` into dedicated fields (`:61-73`), everything else verbatim into `Fields` (`:76`).

**UNVERIFIED:** whether `santhosh-tekuri/jsonschema v6` accepts Go `uint64`/`int64`/`float64` values for `"type":"integer"`/`"number"` without a JSON round-trip, and how it handles a bare `time.Time`. This is `libs-verify`'s item; flag it as the one normalization decision the document builder may need (signedness only — structural normalization is unnecessary).

---

## 3. Pre-gate mechanics (`cmd/akb/required_fields.go`)

Full anatomy (36 lines):

- `checkRequiredFields(tmpl template.Template, fm *frontmatter.ParsedFrontmatter) []string` — `:19-29`. Iterates `tmpl.Schema.Frontmatter` (`map[string]FieldSchema`), keeps keys where `field.Required && !frontmatterKeyPresent(fm, key)`, sorts. **Presence only** (`:16-18` doc): value never inspected; type/enum stay with CEL.
- `requiredFieldsMessage(tmpl, missing) string` — `:33-35`:
  `missing required frontmatter field(s): %s (declared required by template %q)`.
- `frontmatterKeyPresent(fm, key)` lives in **`cmd/akb/templates_write.go:413-421`**: `type`→`fm.Type != ""`, `title`→`fm.Title != ""`, else `_, ok := fm.Fields[key]`.

**Call sites (non-test, complete):**

| Site | Line | Behaviour |
|---|---|---|
| `akb write` | `write.go:523-526` | print stderr, `return validationFailure{}` |
| `akb append` | `append.go:220-222` | same |
| `akb approve` | `approve.go:227` inside `requiredFieldsRefusal` (`:221-234`), called at `:151` | prints `cannot approve '<inputPath>': <message>`, returns `validationFailure{}` |
| `akb template write` | `templates_write.go:152-155` | refusal names the PASS mockup |

**Error/exit path:** prints all messages to stderr itself and returns the bare `validationFailure` sentinel → `classifyExit` maps it to **exit 1** with `report == ""` (`main.go:79-97`; sentinel `main.go:53-55`). Compile/engine faults are `internalError` → exit 2.

**What breaks when removed from the write path:**
1. A page omitting a template-required field is no longer refused *before* CEL. The templates' rules read `page.frontmatter.<required>` unguarded (e.g. `internal/template/embedded/adr.yaml:2-4` contract), so the failure mode becomes `rule <id> could not be evaluated: no such key …` (`write.go:653-661`) → still exit 1, but a worse, author-directed message instead of a user refusal. The JSON Schema pass must own presence via `properties.frontmatter.required` (RFC §7 `open-validation-engine.md:331-338`).
2. `approve` loses its gate too (ADR-009 Consequences: "approve's refusal-wording parity then follows the schema pass's `required` message"); the schema pass MUST emit the same sentence to keep `approve_test.go:673,725` and `append_test.go:770` green.
3. `template write`'s PASS-mockup required check (`templates_write.go:152`) and `optionalKeysSuppliedByMockup` (`:395-409`, still reads `field.Required`) are **not** write-path; they linger until SPEC-005/S8 owns mockup obligations per the V3 `schema:` block. Removing the *helpers* outright breaks them — keep the functions.

**Interaction with missing-`type` handling:** independent. `frontmatter.ValidateType` (`internal/frontmatter/frontmatter.go:130-140`, unknown → `unknown type %q`) and `ValidateTitle` (`:141-148`) run **before** the pre-gate in every branch (`write.go:202/205`, `:326/331`, `:373/378`; `append.go:165/168`). A page with no `type` is refused earlier with `missing required field 'type'`/`unknown type` — the pre-gate never sees it. `frontmatterKeyPresent`'s `type`/`title` special-cases are effectively dead in the write path for that reason. Removing the pre-gate therefore does **not** change missing-`type`/`title` behaviour (`write_errors.txt:20-27`, `core_io_verification.txt:105,110` stay green).

---

## 4. Lint engine attachment (model for the new `schema` checker)

**Registry / interface** — `internal/lint/engine.go`:
- `LintChecker` interface `:37-40` (`Name() string`, `Check(ctx, *KB) ([]LintIssue, error)`).
- `KB` `:45-51`: `Root, LinkGraph, Templates map[string]template.Template, Manifest, Pages []PageData`.
- `PageData` `:56-64`: `RelPath, Content, Body, Frontmatter, ProvenanceMarkers, Annotations, HasFrontmatter`. **No derived content/ast map** — a checker rebuilds it (see cel_lint).
- `LintIssue` `:17-23` (`Type` json `check`, `RuleID` json `rule_id,omitempty`, `Message`, `Path`, `Severity`); `LintReport` `:28-32`.
- `AddChecker` `:87-89`; `Run` `:94-107` — executes in registration order, **any checker error aborts the whole sweep** (`:100-103`), `ByCheck[Name()] = len(issues)` (`:106`).

**Registration site** — `cmd/akb/lint.go:157-167`; insert `engine.AddChecker(lint.NewSchemaChecker())` **before** `NewCELLintChecker()` (`:166`) to satisfy ADR-009 I2. `KB` is built at `lint.go:149-155` from the walk at `:167-...`.

**`cel_lint` as the template** (`internal/lint/cel.go`):
- `Name()` → `"cel_lint"` `:38`.
- Per page: skip if no frontmatter or empty `Type` `:52-53`; skip if no template `:55-58`.
- **Rebuild the page map inside `Check`**: goldmark parse + `cel.BuildPage(page.RelPath, page.Frontmatter, page.Body, doc, page.Body)` `:60-62` — the only precedent for derived-view access from a checker.
- Compile each `tmpl.LintRules[].Rule` `:64-67`; a **compile** failure returns a hard error → aborts the sweep (`:66-68`).
- Eval error → **degrade to a per-page issue** and `continue` `:70-82` (`Severity: "error"`); `result != types.True` → issue `Severity: rule.Severity`, `Message: rule.Expect` `:84-91`.
- `now`: `now := time.Now()` once per `Check` `:44`, injected as `"now": now` `:73`. **Per-checker, not shared** (D7).

**For a `schema` checker (ADR-009 I1-I6):** same `Name()`/`Check` shape; issue `Type: "schema"`; must **degrade** (never `return nil, err`) on per-page schema faults (I3/N3); must emit only `LintIssue` rows, not the SPEC-002 write-time report (I4/N2); registration order before `cel_lint` (I2); `required_fields` stays registered (I5). `LintIssue` shape is frozen (N2). The sweep degradation divergence vs. the write path is deliberate (comment `write.go:594-602`).

---

## 5. Test fallout enumeration

### 5a. Pre-gate wording / behaviour pins (write, append, approve, mockup)

| Test | Assert | Pins |
|---|---|---|
| `cmd/akb/write_test.go:2048` `TestCheckRequiredFields` | `:2116` direct `checkRequiredFields` call | presence-only, sorted, empty-value = set. Helper's own unit test — survives only if the helper stays (mockup path). |
| `cmd/akb/write_test.go:2149` `TestWriteMissingRequiredFieldsFailsValidation` | `:2165` `missing …: deciders, status (… "adr")` | write pre-gate wording + exit 1 |
| `cmd/akb/write_test.go:2176` `…FrontmatterUpdateMissingRequiredField…` | `:2194` | `--frontmatter` path |
| `cmd/akb/write_test.go:2210` `…AppendMissingRequiredField…` | `:2228` | `--append` path |
| `cmd/akb/append_test.go:752` `TestAppendMissingRequiredFieldFailsValidation` | `:770` | `akb append` refusal + page untouched |
| `cmd/akb/approve_test.go:651` `TestApproveRefusesDraftMissingRequiredField` | `:673` (`cannot approve '…'` at :672) | approve refusal wording + exit 1 + no commit |
| `cmd/akb/approve_test.go:694` `TestApproveAllDraftsRefusesOnlyIncompleteDrafts` | `:725` | batch semantics of the gate |
| `cmd/akb/approve_test.go:750` `TestApproveAcceptsPageTypeWithoutTemplate` | — | gate boundary (unknown type ≠ refusal) |
| `cmd/akb/templates_write_test.go:1154` `…RejectsPassMockupMissingRequiredField` | `:1192` | PASS-mockup required check (mockup scope, S8) |
| `cmd/akb/templates_write_test.go:1211` `…FailMockupIsExemptFromRequiredFields` | — | FAIL mockup exempt |
| `internal/lint/required_fields_test.go:54,60` | `:87,106,159` | **lint** checker wording — survives (ADR-009 I5) |

Phase-1-log:195 tracked `approve_test.go:651,694`; current-tree assertions are at `:673`/`:725` (function stubs at 651/694). Coverage gap confirmed: **no testscript pins `cannot approve`** (grep over `test/testdata/` → 0 hits).

### 5b. Testscript pins

| File | Lines | Pins | Effect of SPEC-004 |
|---|---|---|---|
| `test/testdata/lint_tests_adr_write.txt` | `:1-4` comment, `:18` | write-time pre-gate **"refused at write time, before any validation rule runs"**, exit 1, `! exists` | **must be rewritten** to the schema-`required` message/ordering |
| `test/testdata/append_validation.txt` | `:31-35` | append runs the write pipeline; all-rules failure reported, page byte-identical (`:37 cmp`) | ordering assertion adjacent to the new schema step |
| `test/testdata/exit_codes.txt` | `:33-42` | all-rules-fail ADR write: exit 1, six `[rule]` stderr lines (`stderr -count=6 '^\['` `:41`) | if schema **short-circuits** CEL (RFC §7), the exact count/all-rules contract this pins changes |
| `test/testdata/lint_required_fields.txt` | `:28,36,42` | lint wording + `"required_fields": 1` / `:0` by_check keys; `stdout -count=1 'error:'` `:33` | checker survives; the `error:` count can break when the new `schema` checker lands on the same page |
| `test/testdata/lint_tests.txt` | `:1-2` ("nine of the ten checkers"), `:107` `"by_check"` | sweep coverage; envelope | header/coverage comment stale once an 11th checker exists |
| `test/testdata/write_errors.txt` | `:20-27` | `missing required field 'type'`/`'title'` (`ValidateType`, not the pre-gate) | **unaffected** |
| `test/testdata/core_io_verification.txt` | `:103-111` | same `missing required field` (ValidateType) | **unaffected** |
| `test/testdata/cel_validation.txt` | `:150-215` | V2 templates with `required: true` + CEL rules | template fixtures may need V3 `schema:` blocks |
| `test/testdata/superseded_adr.txt` | `:65` | `! stderr 'min_word_count'` | rule-set assertion; changes if schema short-circuits CEL |

### 5c. Model tests to mirror for the new checker

`internal/lint/cel_test.go:16,81,173` (guarded-absent-key, degrade-does-not-abort, per-page issues) and `internal/lint/engine_test.go:56,73,105,133,163` (`AddChecker`, aggregation, `ByCheck` map, `PagesChecked`, checker-error abort).

---

## 6. Template struct readiness (current state only)

`internal/template/template.go`:
- `Schema` `:19-21` = `{ Frontmatter map[string]FieldSchema yaml:"frontmatter" }`.
- `FieldSchema` `:24-28` = `{ Type string; Required bool; Enum []string }` (only `.Required` is live in non-test code; `.Type`/`.Enum` are decorative — s6 recon §2).
- `Template` `:47-55` = `{ Name, Description, Dir, Schema, Validations []ValidationRule, LintRules []LintRule, filename }`.
- Loader `LoadTemplates` `:84-130`: non-strict double-unmarshal (`:106` raw for `detectOldFormat`, `:114` typed); no strict option; only checks are old-format rejection `:73-80`, non-empty name `:118-120`, duplicate name `:122-124`.

**There is no JSON-schema field today.** SPEC-003 (`.pi/research/rfc-002-s6/SPEC-003-templatev3-format-loader.md`) introduces the V3 six-key format whose `schema:` value is a JSON Schema 2020-12 document, and its REQ-007 requires the loader to pass `schema` **uncompiled** (compile-site boundary S6/S7). SPEC-004 therefore consumes a parsed-but-uncompiled schema value from `template.Template`; the struct field does not exist yet (S6 lands it).

---

## Confidence + Gaps

**Confidence: high** for §1–§6 (all read from the current tree at `c8bac2e`; call-site lists are grep-exhaustive). **Medium** for the ordering-precedence consequence in D5 (read from source, not executed).

**Gaps / open items for SPEC-004:**
- **G1** (D5): whether SPEC-004 owns moving the merge/identity preflight ahead of validation, or whether RFC §7's "merge-conflict check → …" is merely nominal. Code today puts it last in write/append only.
- **G2** (D6): the raw-vs-coerced `frontmatter` split — the document builder must not reuse `page["frontmatter"]`; the spec must state this explicitly or the schema will see `time.Time` where ADR-005/008 want raw strings.
- **G3** (D7): no shared-`now` mechanism exists; ADR-009 I2 needs one.
- **G4**: `santhosh v6` handling of `uint64`/`int64`/`time.Time` Go values is UNVERIFIED here (libs-verify's item).
- **G5**: whether the schema pass short-circuits CEL on a schema failure — RFC §7 says "schema short-circuits CEL"; `exit_codes.txt:41`'s `-count=6` and `superseded_adr.txt:65` pin the *current* all-CEL-rules contract, so this is a concrete test-churn decision.
- **G6**: `approve`'s refusal wording parity (approve.go:221) has no testscript guard (phase1-scope-log:195); ADR-009 requires the schema pass to preserve the wording.
- **G7**: exact SPEC-002 report type home (F1 in the s7 LOG) is unresolved; this recon only anchors the `validationFailure{}`/`classifyExit` seam it must feed.
