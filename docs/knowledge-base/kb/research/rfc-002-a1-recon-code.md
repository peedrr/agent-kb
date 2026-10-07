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
- cmd/akb/**
status: final
summary: Read-only code recon mapping where RFC-002-A1's validator-library and format-assertion decisions touch the akb codebase, with file:line anchors.
tags:
- code-recon
- json-schema
- validation
- rfc-002
title: RFC-002-A1 Integration-Point Map (recon-code)
type: research
updated: "2026-10-06"
---
# RFC-002-A1 Validator — Integration-Point Map (recon-code)

Repo root: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb` (cwd).
Scope: where RFC-002-A1's decisions (JSON Schema library + format-assertion policy) touch the
current akb codebase. Read-only reconnaissance; anchors are `file:line` in the current tree.

Confidence legend: **[H]** high (read the code directly), **[M]** medium (inferred from code +
RFC text), **[L]** low (needs confirmation).

---

## 1. `internal/cel/pagebuilder.go` — current date/time coercion

The whole current coercion story is one small helper plus one call site.

- `convertDateField(value any) any` — **`internal/cel/pagebuilder.go:29-46`** **[H]**. Accepts a
  `string`; tries `time.Parse(time.RFC3339, str)` first (`:36`), else `time.Parse("2006-01-02", str)`
  (`:41`); returns the original value unchanged if neither parses (including non-strings, `:32-34`,
  `:45`). Exactly the CEL acceptance set RFC-002 §5.2 calls "Go-strict RFC3339" — because it uses
  Go's `time.Parse`, it rejects lowercase `t`/`z` and leap seconds by construction.
- `BuildPage(...)` — **`internal/cel/pagebuilder.go:51-88`**. Iterates `fm.Fields` and coerces
  **every** value name-agnostically: `for k, v := range fm.Fields { fmMap[k] = convertDateField(v) }`
  **`:60-63`** **[H]**. `frontmatter` map gets `type`/`title` assigned separately (`:64-65`) and those
  two are **not** coerced. So today: (a) coercion is **name-agnostic** (any key whose value looks like a
  date), (b) **top-level only** (only direct children of `frontmatter`; nested maps/arrays are never
  walked), (c) it does not distinguish `format: date` vs `format: date-time` — both collapse to
  `time.Time`.
- `BuildOldPage(...)` — **`internal/cel/pagebuilder.go:91-117`** — reads the on-disk page and calls
  `BuildPage`, so the same bridge applies to `old_page` **[H]**. RFC-002 D7 requires undeclared fields
  never coerced + declaration-driven (schema-declared `format: date-time`/`format: date`) at arbitrary
  depth; that is a rewrite of the `:60-63` loop (traversal mechanics deferred to RFC-002-A4) and a
  signature change to accept the schema/declaration set.
- Constants/limits for the new format checker: none exist here. The CEL-side parse set is implicit in
  Go's `time.RFC3339`; RFC-002-A1 pins a validator format checker to that same set **[M]** (RFC §5.2).

## 2. `internal/cel/engine.go` — CEL env, cost limit, cache, pipeline position

- `NewEnv()` — **`internal/cel/engine.go:35-46`** **[H]**. Builds `cel.NewEnv` with three variables:
  `page` (`map(string,dyn)`), `old_page` (nullable same), `now` (timestamp). No custom functions,
  no declaration hooks — a JSON Schema layer sits **beside** this, not inside it.
- `MaxCostLimit = 100000` — **`:22`**; applied via `cel.CostLimit(MaxCostLimit)` in `CompileRule`
  — **`:61`** **[H]**.
- `CompileRule(env, expr)` — **`:51-69`** **[H]**. Cache is `programCache sync.Map` keyed by **the
  expression string alone** (`:30`, load `:52`, store `:66`) — no env identity, no template identity.
  Relevant to A1 only if a schema-derived rule text is ever fed through CEL (RFC forbids translation,
  so probably not), but worth flagging as a pre-existing collision risk.
- `Evaluate(...)` — **`:74-105`**; read-only; recovers panics, maps cost cancellation to
  `ErrComputeBudget` **[H]**.
- Pipeline position: RFC-002 §4 order is `… required gate (from schema) → JSON Schema → CEL
  validations → write`. In code that means a new validate call must sit **between** the
  required-field gate and `runTemplateValidations` (see §4). CEL is currently the only thing the
  env touches; JSON Schema is a separate pass over a separate `document` view.

## 3. `internal/template/template.go` — structs, loader, the `schema:` collision

- `type Schema struct { Frontmatter map[string]FieldSchema \`yaml:"schema:frontmatter"\` }` —
  **`internal/template/template.go:18-21`** **[H]**; `FieldSchema {Type, Required, Enum}` —
  **`:23-27`**. `Template.Schema` is `yaml:"schema"` — **`:51`** **[H]**.
- **Critical collision:** the current `schema:` key is the homegrown `schema.frontmatter` block
  (presence/type/enum). RFC-002 §6 **reuses the same `schema:` key** for a JSON Schema 2020-12
  document over the whole `document` view, and retires the homegrown block. So a new top-level
  `schema:` is **not additive** — it is a type/meaning change of an existing key, with a hard-reject
  migration path. Precedent for that rejection already exists: `detectOldFormat(raw, filename)` —
  **`:73-82`**, called at **`:109`** — returns "Template format has changed" for `required`/`optional`/
  `body`. RFC §6 explicitly cites this precedent for the `schema.frontmatter` retirement **[H]**.
- Loader `LoadTemplates(dir)` — **`:84-138`** **[H]**. Two unmarshals:
  `yaml.Unmarshal(data, &raw)` at **`:106`** (into `map[string]any`, used only by `detectOldFormat`),
  then `yaml.Unmarshal(data, &tmpl)` into the typed struct at **`:114`**. Both are **non-strict**
  (goccy/go-yaml ignores unknown keys unless `yaml.Strict()`/`yaml.DisallowUnknownField()` is passed)
  — the "loader's non-strict unmarshal, `template.go:106-114`" the RFC names as the typo trap. A1's
  "unknown template keys rejected at authoring time (`template write`), warned at load" lands here:
  load path must stay lenient, `template write` must become strict.
- New JSON Schema block would be a field on `Template` (likely typed, e.g. `Schema json.RawMessage`
  or a decoded struct) replacing/renaming the current `Schema Schema`. Cross-refs: `cmd/akb/
  templates_write.go` (authoring validation + mockups) and `cmd/akb/template.go` are the authoring
  side an A1 static check would touch **[M]** — not read in this pass.

## 4. `cmd/akb/write.go` — write pipeline order + `runTemplateValidations`

- Entry `runWrite` — **`cmd/akb/write.go:89`** **[H]**. Three input branches share one validation tail:
  - `--frontmatter` update branch **`:151-300`**; `--append` branch **`:301-410`**; create/stdin branch
    **`:411-510`**.
- Per-branch sequence: `frontmatter.Parse` (`:197`, `:320`, `:367`) → `frontmatter.ValidateType`
  (`:202`, `:243`, `:326`, `:373`) → `frontmatter.ValidateTitle` (`:205`, `:331`, `:378`) →
  `cel.BuildOldPage` for overwrites (`:328`, `:479`) → frontmatter re-serialized via
  `yaml.Marshal` with `type`+`title`+`Fields` merged (`:255-260`, `:344-355`, `:491-500`) **[H]**.
- Shared tail (the slot A1 modifies), **`:512-556`** **[H]**:
  1. `tmpl := templates[fm.Type]` (`:513-516`)
  2. `astDoc := goldmark parse(body)`; `page := cel.BuildPage(relPath, fm, body, astDoc, body)`
     — **`:517-518`**
  3. required-field gate: `if missing := checkRequiredFields(tmpl, fm); len(missing) > 0 { …;
     return validationFailure{} }` — **`:523-526`**
  4. `runTemplateValidations(celEnv, tmpl, page, oldPage)` — **`:529`**
  5. `search.ExtractTags/ExtractSummary` (`:531-532`) → `store.Write` (`:547`) → index tx
     (`:551-575`).
- **JSON Schema slot:** RFC §4 puts JSON Schema **after** the required gate (`:523`) and **before**
  CEL (`:529`). It validates the `document` view (raw frontmatter, no date coercion), not the coerced
  `page` view built at `:518`. Ordering matters: the required gate must stay first (guard invariant),
  and BuildPage's coercion must not leak into the schema-validated document.
- `runTemplateValidations(...)` — **`:603-651`** **[H]**. Compiles each `tmpl.Validations[].Rule`
  (`:605`), evaluates with `{page, old_page, now}` (`:616-622`), collects **all** failures +
  unevaluable rules into `[]cel.ValidationError`, prints to stderr, returns `validationFailure{}`.
  Fail-closed on compile errors (`&internalError`) and eval errors (collected → validation exit)
  **[H]** — the A1 consequence "schema pass fails the write with the same exit-1 shape" should mirror
  this aggregation + stderr contract.
- Duplicate required-check helper: `checkRequiredFields` — **`cmd/akb/required_fields.go:15-27`**,
  `requiredFieldsMessage` — **`:31-33`** **[H]**. Presence-only against `tmpl.Schema.Frontmatter`.

## 5. `go.mod` — dependencies

- Module `github.com/peedrr/agent-kb`; `go 1.26.1` — **`go.mod:1-3`** **[H]**.
- Direct: `github.com/spf13/cobra v1.10.2`, `modernc.org/sqlite v1.48.2`; second block:
  `github.com/goccy/go-yaml v1.19.2`, **`github.com/google/cel-go v0.28.0`**,
  `github.com/rogpeppe/go-internal v1.14.1`, `github.com/yuin/goldmark v1.8.2`,
  `go.abhg.dev/goldmark/frontmatter v0.3.0`, `golang.org/x/sys v0.43.0` — **`:7-18`** **[H]**.
- **No JSON Schema library is present today** **[H]**. A repo-wide (case-insensitive) grep for
  `json schema|jsonschema|JSONSchema` matches only `docs/knowledge-base/.../RFC-001*`, `RFC-002*`,
  `kb/index.md`, and `.git/logs` — zero Go-source or `go.mod` hits. The RFC recommends
  `github.com/santhosh-tekuri/jsonschema/v6` (§9) but it is not yet a dependency.
- Note: `gopkg.in/yaml.v3 v3.0.1` is present but **indirect** (`:26`) and the project's
  anti-pattern list forbids direct use — an A1 schema library choice should not pull it in.
  `cel.dev/expr v0.25.1`, ANTLR, etc. are cel-go's transitive deps (`:21-28`).

## 6. `internal/frontmatter/frontmatter.go` — the document view a schema would validate

- `type ParsedFrontmatter { Type string; Title string; Fields map[string]any }` —
  **`internal/frontmatter/frontmatter.go:22-26`** **[H]**.
- `Parse(content)` — **`:29-96`** **[H]**. Decodes YAML frontmatter into `map[string]any`, then
  **routes `type` and `title` out of `Fields`** into dedicated struct fields (`switch key`, `:79-93`);
  everything else stays in `Fields`. **This is the split RFC-002 D10 removes**: `type`/`title` must
  become ordinary keys visible to both validators, and `frontmatter.Parse` must stop routing them out.
  A JSON Schema over the "raw frontmatter values as parsed from YAML" needs the *whole* key set
  including `type`/`title` in one map — today it is split, so the document builder must re-merge or
  `Parse` must change.
- `ValidateType` — **`:130-139`**; `ValidateTitle` — **`:141-147`**; both are the surviving
  door-bouncers per D10 (`type` stays the template selector; `title` becomes a declared display field).
  `IsDraft(fields)` — **`:151-164`** reads `is_draft` out of `Fields` (relevant to D3 `state_field`).
- YAML decode is `goccy/go-yaml` via `go.abhg.dev/goldmark/frontmatter` (`Unmarshal: yaml.Unmarshal`,
  `:33-37`); raw map values are Go-typed (`string`, `bool`, etc.) — a JSON Schema validator will need
  a JSON-model conversion step from this map (RFC-002 mentions the document view is shared with
  `read --json`) **[M]**.

## 7. `internal/lint/` — what validates frontmatter formats today (one note)

- `RequiredFieldsChecker` — **`internal/lint/required_fields.go:31-73`**; `missingRequiredFields`
  — **`:80-93`**; `frontmatterFieldPresent` — **`:98-106`**. It checks **presence only** against
  `tmpl.Schema.Frontmatter`; comment at `:91-93` explicitly says "a value is never inspected and type
  and enum constraints stay with the template's CEL rules". **No format/type assertion happens in lint
  today** — there is no schema sweep checker; RFC-002 P4 requires a new sweep-time schema checker for
  pages that bypass write (git pull). The existing sweep-time value checking is `CELLintChecker` —
  **`internal/lint/cel.go:34-107`** (evaluates `tmpl.LintRules`, degrades eval errors per-page).
  `PageData`/`KB` types live in **`internal/lint/engine.go:17-59`**.

---

## Pipeline order today (grounded)

`ResolveKB` → `db.OpenKB` → `config.Load` → `template.LoadTemplates` → `storage.OpenStore` →
`cel.NewEnv` → branch (frontmatter-update | append | create) → `frontmatter.Parse` →
`ValidateType` → `ValidateTitle` → `BuildOldPage` (overwrites) → re-serialize frontmatter →
`cel.BuildPage` → **`checkRequiredFields` (presence gate)** → `runTemplateValidations` (CEL) →
`ExtractTags/Summary` → `store.Write` → index+linkgraph tx → commit.

A1 inserts: a JSON Schema pass validating the raw `document` view, **between** the required gate
(`write.go:523`) and `runTemplateValidations` (`write.go:529`); plus a minimal format-assertion
checker (`date`/`date-time`/`time`) whose `date-time` set equals Go's `time.RFC3339`; plus a
sweep-time schema checker in `internal/lint/`.

## Three most important integration constraints

1. **`schema:` is not a new key — it is the existing `Schema yaml:"schema"` (currently
   `schema.frontmatter`).** Adding JSON Schema means repurposing/renaming that key with a
   hard-reject migration modeled on `detectOldFormat` (`template.go:73-82`), and making
   `template write` strict while `LoadTemplates` stays lenient (`template.go:106-114`).
2. **The date bridge is name-agnostic and top-level-only (`pagebuilder.go:60-63`) and must become
   declaration-driven (schema `format: date`/`date-time`) at arbitrary depth, for both `page` and
   `old_page`.** Its acceptance set is exactly Go `time.RFC3339` (`pagebuilder.go:36`) + `2006-01-02`
   (`:41`); A1's format checker must be pinned to that same set.
3. **JSON Schema must validate the raw `document` view, but `frontmatter.Parse` currently splits
   `type`/`title` out of `Fields` (`frontmatter.go:79-93`).** Either `Parse` changes (D10) or the
   document builder re-merges; and the schema pass must slot after the required gate and before CEL
   (`write.go:523`/`:529`) so CEL's guard invariant and fail-closed exit-1 aggregation
   (`runTemplateValidations`, `:603-651`) are preserved.

## Explicit gaps / unverified

- **[G1]** Did not read `cmd/akb/templates_write.go`, `cmd/akb/template.go`, `cmd/akb/read.go`,
  or `internal/search` — the authoring-side strict-unknown-key check and any `read --json`
  document-view code live there and are A1-adjacent.
- **[G2]** RFC-002-A4 (depth traversal spec) is out of scope; the exact new signature of
  `BuildPage`/`BuildOldPage` is undetermined here.
- **[G3]** No code confirms which JSON Schema library A1 will ratify — only RFC §9's recommendation
  (`santhosh-tekuri/jsonschema/v6`). `go.mod` has no schema dep today (verified).
- **[G4]** The `duration`/`time` format seam (RFC §5.2) has no current code anchor; no duration
  parsing exists in the repo.
- **[G5]** Did not enumerate `internal/lint/engine.go`'s full checker list / registration; only the
  two frontmatter-relevant checkers were inspected.
