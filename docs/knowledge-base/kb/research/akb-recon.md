---
grammar: 1
type: research
title: akb feasibility recon — TemplateV2, CEL, write/read flows, JSON I/O options
status: final
provenance: agent-drafted
run: json-cel
as_of: 2026-09-26
created: 2026-09-26
updated: 2026-09-26
scope:
  - internal/**
  - cmd/akb/**
tags: [recon, akb, json-schema, cel]
summary: Read-only recon of akb internals (TemplateV2, CEL env, write/read/template-write flows, JSON precedents) rating feasibility of JSON Schema↔CEL bridging and read/write --json.
informs:
  - RFC-001
  - RFC-002
---
# akb feasibility recon (READ-ONLY)

Repo: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb` @ v0.21.0. Nothing was modified.
Questions: (a) JSON Schema → CEL rules, (b) CEL rules → JSON Schema, (c) `akb read --json`, (d) `akb write --json`.

---

## 1. TemplateV2 format

`internal/template/template.go:15-50` — the entire on-disk contract is four small structs; rule bodies are **raw strings**, never parsed at load time.

```go
// internal/template/template.go:15
type Schema struct { Frontmatter map[string]FieldSchema `yaml:"frontmatter"` }
// :23
type FieldSchema struct {
	Type     string   `yaml:"type"`
	Required bool     `yaml:"required"`
	Enum     []string `yaml:"enum,omitempty"`
}
// :26
type ValidationRule struct {
	ID          string `yaml:"id"`
	Rule        string `yaml:"rule"`
	Requirement string `yaml:"requirement,omitempty"`
	Expect      string `yaml:"expect"`
}
// :32
type LintRule struct {
	ID       string `yaml:"id"`
	Rule     string `yaml:"rule"`
	Severity string `yaml:"severity"` // "warning" | "error"
	Expect   string `yaml:"expect"`
}
// :41
type Template struct {
	Name, Description, Dir string
	Schema                 Schema           `yaml:"schema"`
	Validations            []ValidationRule `yaml:"validations"`
	LintRules              []LintRule       `yaml:"lint_rules"`
	filename               string
}
```

- No `expression`/`message` fields; the names are `rule` + `expect` (+ `requirement` on validations only). No per-rule severity on validations (always hard error); `severity` exists only on `lint_rules`.
- Loader: `LoadTemplates` (`template.go:78-146`) unmarshals into `map[string]any` first only to reject the pre-v2 shape (`required`/`optional`/`body` → "Template format has changed"), then unmarshals into `Template`; `name` empty and duplicate names are hard errors (`:129-141`). No `FieldSchema.Type`/`Enum` validation, no CEL compile at load time, no `field`-vs-AST (`page.ast.*`) declared type.
- Embedded showcase set: `internal/template/embedded/{adr,note}.yaml` + `_pass.md`/`_fail.md` (`//go:embed embedded/*`, `template.go:52-53`).

`internal/template/embedded/adr.yaml:9-48` (schema shape, `type: string|list` only):

```yaml
name: adr
dir: decisions
schema:
  frontmatter:
    title:   {type: string, required: true}
    tags:    {type: list,   required: true}
    status:
      type: string
      required: true
      enum: [proposed, accepted, deprecated, superseded]
    sources: {type: list,   required: false}
validations:                       # adr.yaml:49-178, 22 rules
  - id: title_style
    rule: 'page.frontmatter.title.matches("^[A-Z]") && size(page.frontmatter.title) <= 80'
    requirement: Title must start with a capital letter and stay within 80 characters
    expect: title must start with an uppercase letter and be at most 80 characters
lint_rules:                        # adr.yaml:179-198
  - id: adr_stale
    rule: '!has(page.frontmatter.updated) || now - timestamp(page.frontmatter.updated) < duration("4320h")'
    severity: warning
    expect: ADR hasn't been updated in 180 days
```

`internal/template/embedded/note.yaml:1-35` is the minimal counterexample (1 validation, 1 lint rule).
**Critical fact:** `FieldSchema.Type` and `.Enum` are read **nowhere** in non-test code; only `.Required` is consumed (`cmd/akb/required_fields.go:21`, `cmd/akb/templates_write.go:386`, `internal/lint/required_fields.go:83`). Declaration types/enums are decorative documentation; every *real* constraint lives in a CEL string.

Authoring doctrine (must be respected by any generator): `internal/skill/embedded/kb-management/references/TEMPLATE.md:41-77` — required keys read unguarded, optional keys always behind `has()`, every `old_page` read guarded, lint rules guard everything except `type`/`title`.

## 2. CEL engine and the `page` map

`internal/cel/engine.go:25-37` — the env is the **plain CEL standard library plus nothing else**: no extensions, no custom functions, no `cel.OptionalTypes`.

```go
func NewEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("page", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("old_page", cel.NullableType(cel.MapType(cel.StringType, cel.DynType))),
		cel.Variable("now", cel.TimestampType),
	)
}
```

Thus available: `has()`, `in`, `size()`, `matches()` (RE2, **unanchored**), `startsWith/endsWith/contains`, `.all/.exists/.filter`, `timestamp()`, `duration()`, arithmetic and comparisons. No `format()`, no URL/path helpers, no numeric helpers beyond arithmetic; RE2 has no lookaround/backrefs.

- Cost limit: `MaxCostLimit = 100000` (`engine.go:22`), applied via `cel.CostLimit` in `CompileRule` (`engine.go:43-61`); programs are cached in a package-level `sync.Map` keyed by the expression **string** (`engine.go:32`, `.Load`/`.Store` at `:44,:56`).
- `CompileRule` returns only a `cel.Program` — the checked AST is discarded, so nothing downstream can introspect a rule (matters for feature b).
- `Evaluate` (`engine.go:67-106`) recovers panics and maps cost cancellation to `ErrComputeBudget` (`engine.go:24`).

`internal/cel/pagebuilder.go:47-85` — `BuildPage` produces exactly four top-level keys:

```go
page["file"]        = {path, name, dir}
page["frontmatter"] = <every fm.Fields key> + type + title   // :58-65
page["content"]     = {raw, word_count, char_count}          // :68-73
page["ast"]         = {headings, links, code_blocks}         // :75-79
page["akb"]         = {provenance_markers, annotations}      // :81-84
```

- `type` and `title` are **always injected** from `ParsedFrontmatter.Type/.Title` even if absent from `Fields` (`pagebuilder.go:56-64`); every other frontmatter key keeps its YAML-decoded Go type (string/[]any/int64/float64/bool/map).
- Date coercion: `convertDateField` (`pagebuilder.go:26-45`) converts **top-level string field values only** to `time.Time` for RFC3339 *and* date-only `2006-01-02`. Strings nested inside lists/maps are not converted, so `timestamp(page.frontmatter.tags[0])` fails.
- AST element shapes: `internal/cel/types.go:7-24` = `{level,text,line}`, `{target,text,is_wikilink,line}`, `{language,line}`; built by `flattenHeadings`/`flattenLinks`/`flattenCodeBlocks` (`pagebuilder.go:139-...`, links merge wikilink tokens with goldmark links, sorted by offset for determinism — `pagebuilder.go:180-305`).
- `BuildOldPage` (`pagebuilder.go:88-113`) reads the on-disk page through `storage.Provider` and returns `nil` when the file is absent.
- Absent-key semantics: reading a missing `page.frontmatter.X` key is a CEL **evaluation error**, not false → the reason for the `has()` doctrine.

Rule failures are typed: `internal/cel/errors.go:11-16` `ValidationError{RuleID, Message, Line, Severity}` (`Error()` → `[id] message`, optional `(near line N)`).

## 3. Write flow (`cmd/akb/write.go`, 653 lines)

Single `runWrite` (`:86`) with three mutually exclusive input shapes: `--frontmatter` (`:127-254`), `--append` (`:270-357`), and plain stdin (`:358-...`).

- Stdin gate: rejects a TTY (`:257-265`), then `stdinContent, err := io.ReadAll(os.Stdin)` at `:266`.
- Plain path: `fm, body, err = frontmatter.Parse(stdinContent)` at `:359`; `ValidateType` `:365`, `ValidateTitle` `:370`; `writeContent = stdinContent` at `:374` (byte passthrough when nothing is stamped).
- Stamping: `created` default if missing and `is_draft` deletion set `needsReserialize` (`:376-386`); type-dir resolution from `tmpl.Dir` (`:389-397`); path guards `.md` suffix / `raw/` / managed `index.md`/`log.md` (`:399-421`); `ending reserialize` re-marshals with `type`/`title` re-merged into `allFields` (`:490-497`).
- Locks: repository lock held across read→validate (`:445-455`) and again across write→commit (`:527-533`).
- Validation, `:501-522`:

```go
tmpl, ok := templates[fm.Type]
md := goldmark.New(); astDoc := md.Parser().Parse(text.NewReader(body))
page := cel.BuildPage(relPath, fm, body, astDoc, body)
if missing := checkRequiredFields(tmpl, fm); len(missing) > 0 {
	fmt.Fprintln(os.Stderr, requiredFieldsMessage(tmpl, missing))
	return validationFailure{}
}
if err := runTemplateValidations(celEnv, tmpl, page, oldPage); err != nil { return err }
```

- `runTemplateValidations` (`:595-635`): compiles each rule (compile failure → `internalError` = exit 2), **collects every failed and every unevaluable rule**, prints each with `fmt.Fprintln(os.Stderr, ve.Error())`, returns the bare `validationFailure{}` (exit 1). Message helpers `unevaluableRuleMessage` (`:645`).
- Post-validation side effects: `store.Write` → one SQLite tx indexing FTS + link graph (`:536-565`) → human stdout (`:567-577`, `Written to kb/<rel>` + index nudge).
- `checkRequiredFields` (`cmd/akb/required_fields.go:16-25`): presence only, sorted; `frontmatterKeyPresent` (`templates_write.go:399-410`) knows `type`/`title` live outside `Fields`.
- Exit codes (`cmd/akb/main.go:17-23, 68-90`): 0 ok; 1 = `validationFailure`/`driftDetected` (message already printed, no extra report); 2 = `*usageError`/`*path.GuardError` ("usage: ...") or `*internalError` ("internal: ..."); anything else → 1 with `Error: ` prefix.
- `akb append` repeats the identical gate (`cmd/akb/append.go:217-223`) — a shared helper exists implicitly but is duplicated.

**Where `--json` slots in:** the cheapest correct insertion is at `:266-273` — unmarshal the piped JSON into `{frontmatter, body}` (or `{content}`), `yaml.Marshal` the frontmatter map, and synthesize `[]byte("---\n"+yamlBytes+"---\n"+body)` as `stdinContent`; everything downstream (`Parse`, type/title validation, required fields, CEL, old_page, locks, indexing) then runs **unchanged**. A second, deeper option is to construct `*frontmatter.ParsedFrontmatter` directly and skip `Parse`, which bypasses the `type`/`title` routing logic and is easier to get wrong (`Fields` must be non-nil).

## 4. Read flow (`cmd/akb/read.go`, 67 lines)

- Resolve KB + config (`:36-45`), `path.ResolveKBPath` (`:47`, with `ErrUseAKBRawWrite` → usage error), then `storage.NewFilesystemProvider(kbRoot).Read(...)` and `fmt.Print(string(data))` (`:57-65`). Byte-exact passthrough; no frontmatter parsing, no flags at all.
- A `--json` flag would parse via `internal/frontmatter.Parse` (`internal/frontmatter/frontmatter.go:27-88`) and emit `{path, type, title, frontmatter, body}`. Note `Parse` fails hard on pages with no `---` block (`no frontmatter found in content`, `internal/frontmatter/frontmatter.go:47`) and on `type`/`title` that are not strings (`:71-83`), and it moves `type`/`title` **out of** `Fields` (`:83-87`) so a JSON emitter must re-merge them.
- Precedent payloads: `cmd/akb/list.go:22-25` (`pageInfo{path,is_draft}`, plain `json.Marshal` `:54-63`), `lint.go:45`, `search.go:95-119` (`json.Encoder` with `SetIndent("", "  ")`), `links.go:154-206` (`backlinkJSON` + envelope struct), `raw_status.go:170-180` (`MarshalIndent`, empty slice kept non-nil so JSON renders `[]` — see `raw_status_test.go:41-90` pinning that), `discover.go:54-58`.

## 5. Template write path (`cmd/akb/templates_write.go`, 479 lines)

Validation gates before anything is written:

1. Name regex + flag requirements (`:53-77`): `--pass` and `--fail` are **mandatory for a new template** (`:78-81`, re-checked at `:141` and `:208`).
2. YAML parse into `map[string]any` for old-format rejection (`:80-88`) and into `template.Template` (`:90-101`); `name` must match the argument (`:104`).
3. CEL **syntax** check of every `validations[].rule` and `lint_rules[].rule` via `cel.CompileRule` (`:108-121`); errors are plain `fmt.Errorf` → exit 1.
4. Pass mockup must carry every schema-required field (`:152-157`) and must pass all validations as given (`:159-162`).
5. Optional-key-stripping proof: re-evaluate the mockup once per schema-optional key it supplies, with that key removed; an *evaluation error* gets a purpose-built message telling the author to guard with `has()` or mark the field `required: true` (`:165-181`).
6. Self-`old_page` no-op-update proof (`:183-187`).
7. Fail mockup must fail ≥1 validation (`:216-222`).
8. Overwrite: diff + page count + `--force` (`:224-259`) — `--force` bypasses the existence warning only.

Supporting helpers reusable by a generator: `buildTestPage` (`:341-345`), `evaluateValidations` (`:352-378`, returns failed IDs / the first erroring rule ID), `optionalKeysSuppliedByMockup` (`:384-397`), `withoutFrontmatterKey` (`:415-427`), `templateFilesMatch` (`:432-447`), atomic temp-dir + rename swap + single `akb: template write <name>` commit (`:261-337`).

`akb template get` (`cmd/akb/template.go:131-192`): default = Writer View (`name`, `description`, `schema`, `requirements[]` — the `requirement` strings, no rules); `--example` = pass mockup (with stale-mockup warnings, `:200-230`); `--full` = raw YAML. A JSON->CEL generator would most likely emit the `.yaml` and hand it to `template write --template <file> --pass ... --fail ...`.

## 6. Existing JSON precedent

Six flags, all `BoolVar(&x, "json", false, ...)`: `list.go:40`, `lint.go:45`, `search.go:38`, `links.go:56`, `raw_status.go:54`, `discover.go:18/43`. Conventions: results to **stdout** as a JSON envelope/array, failures stay human text on stderr with the normal exit codes; empty collections must serialize as `[]` not `null` (pinned by `raw_status_test.go:41-90`). Integration coverage: a single testscript line, `test/testdata/exit_codes.txt:106` (`akb raw status --json`). No `--json` on `write`, `read`, `append`, `template *`. No JSON Schema library or JSON-Schema handling exists anywhere in the repo (only a prose mention in `.sisyphus/plans/akb-p5-lint.md:1103`, describing lint's JSON output). No help-text snapshot tests, so adding flags breaks nothing.

## 7. Constraint inventory (blocks / complications per feature)

- **`type` is mandatory** on every page; `ValidateType` rejects unknown/absent type (`frontmatter.go:130-140`), and `akb write`/`append` cannot run without a matching template (`templates_write.go` doctrine, `TEMPLATE.md:16`). Any JSON-ingest path must supply or derive `type`.
- **`FieldSchema.Type`/`Enum` are never enforced** ⇒ generating a TemplateV2 schema block from a JSON Schema alone gives *weaker* guarantees than the JSON Schema promised; constraints must be emitted as CEL rules.
- **`page.frontmatter` key naming** is the frontmatter key verbatim (`pagebuilder.go:56-64`) — JSON Schema property names can contain characters that are illegal/fragile as bare CEL field selectors (`page.frontmatter.foo-bar`, `page.frontmatter.` + unicode). Index/`in` access forms would be required, which the hand-written templates never use and the guard doctrine does not cover.
- **Date coercion is shallow** (`pagebuilder.go:26-45` top-level strings only) ⇒ JSON Schema `format: date-time` on array items / nested objects cannot use `timestamp()`.
- **Regex dialect**: CEL `matches` = RE2, unanchored; JSON Schema `pattern` = ECMA-262, unanchored but with lookarounds/backrefs that RE2 rejects. A generator must anchor (`^...$` when the intent is full match) and reject unsupported constructs.
- **No numeric/bool constraints in the schema block**, and JSON numbers decoded from JSON into `map[string]any` become `float64` (via `encoding/json`) whereas the YAML path yields `int64` — a rule like `page.frontmatter.count == 3` may behave differently across the two ingest paths (CEL numeric equality across int/double is usually tolerant, but `%`/integer division and type-strict comparisons are worth an explicit test).
- **Validation results are not machine-readable**: `runTemplateValidations` prints to stderr and returns `validationFailure{}` (no data), and `checkRequiredFields` failures likewise (`write.go:515-518`, `:595-635`). A structured `--json` write needs that function's signature widened (two callers: `write.go:521`, `append.go:222`).
- **`--json` interacts with `--append`/`--frontmatter`/`--dated`**: those paths take raw markdown / `key=value` strings (`write.go:69-72`, `:127-135`); a JSON document plus `--frontmatter` must be rejected as a usage error (exit 2) rather than silently merged.
- **`writeContent` byte-passthrough**: on the plain path the exact stdin bytes are written unless reserialization triggers (`write.go:374`, `:482-497`); a JSON path always reserializes, so YAML key order must be deterministic (goccy sorts map keys; the existing reserialize path relies on it) or page diffs will churn.
- **`template write` requires mockups** (`templates_write.go:78-81`) whose pass mockup must carry every required field and pass every rule (`:152-222`) — a generated template needs generated mockups too.
- **Rule programs cannot be introspected** after `CompileRule` (only `cel.Program`, `engine.go:43-61`), so a CEL→JSON-Schema extractor must add a new entry point (e.g. return `*cel.Ast` / walk the checked AST) or fall back to string parsing.
- Locking/commit side effects mean `--json` write must go through the same lock sequence (no shortcut) — but that is already the case for any code inserted at `write.go:266`.

---

## FEASIBILITY

| Feature | Rating | Single biggest blocker |
|---|---|---|
| (a) JSON Schema → CEL validation rules | **moderate** | The template schema's `type`/`enum` are dead fields — nothing enforces them (`internal/template/template.go:24`, only `.Required` is read at `cmd/akb/required_fields.go:21`) — so the generator must translate *every* constraint into CEL, and CEL's fixed stdlib (no numeric keywords, RE2-only regex, no `allOf`/`oneOf`/`$ref`) plus `template write`'s mandatory `_pass.md`/`_fail.md` mockups (`templates_write.go:78-81,152-222`) make the generated artifacts, not the expressions, the real work. |
| (b) CEL rules → JSON Schema | **moderate/hard** | Rules are opaque strings compiled into a `cel.Program` with the AST discarded (`internal/cel/engine.go:43-61`), so extraction needs a new AST-exposing entry point; even then, rules over `page.ast.*`, `page.content.*`, `old_page.*` and `now` have no JSON Schema representation, so the output is best-effort and lossy by construction. |
| (c) `akb read --json` | **easy** | `read.go` is 67 lines of byte passthrough, but pages without a frontmatter block (or with non-string `type`/`title`) make `frontmatter.Parse` fail (`internal/frontmatter/frontmatter.go:47,71-83`) — the flag needs a defined fallback (null frontmatter + raw body) or it turns previously readable pages into errors. |
| (d) `akb write --json` | **moderate** | All validation can be reused by synthesizing `---\nyaml\n---\nbody` at `write.go:266` and letting the existing pipeline run — but validation *results* are unstructured: `runTemplateValidations` prints to stderr and returns a data-free `validationFailure{}` (`write.go:595-635`), so a JSON caller cannot learn which rules failed without scraping stderr and the function must be refactored (callers: `write.go:521`, `append.go:222`). |

Secondary notes for the parent: no JSON Schema dependency exists in `go.mod` (would be a new dep, or hand-rolled structural mapping); the "schema block is documentation-only" finding is the single most important design fact for both (a) and (d) — a `--json` write today validates only `required` presence plus whatever CEL rules the template author wrote, not the declared `type`/`enum`.
