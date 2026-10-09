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
- internal/**
- cmd/**
- docs/knowledge-base/.agent-kb/templates/**
status: final
summary: Recon of today's CEL page-map shape, word_count/char_count/raw consumers, required-gate coupling, loader leniency, lint interface, template inventory, and write/read flags.
tags:
- rfc-002
- phase-1
- recon
- document-model
title: Phase-1 scope recon — code surface and rule consumers
type: research
updated: "2026-10-09"
---

# RFC-002 Phase-1 Recon — Code Surface

Repo: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb` @ `a316f0a0bd5029710eee16cf4bbfcadd88391598` (main). READ-ONLY.

---

## Q1 — Current CEL page-map shape (`internal/cel/pagebuilder.go`)

`BuildPage` (`internal/cel/pagebuilder.go:52-85`) builds a nested `map[string]any`. Top-level keys, verbatim:

- `page["file"]` = `{path, name, dir}` — `pagebuilder.go:54-58` (`filepath.Base`/`filepath.Dir` of relPath).
- `page["frontmatter"]` = one map of every FM field plus `type`/`title`; each string value passed through `convertDateField` — `pagebuilder.go:60-66`.
- `page["content"]` = `{"raw": bodyStr /* string */, "word_count": len(strings.Fields(bodyStr)), "char_count": len(body)}` — `pagebuilder.go:69-73`.
  - **`content` is an OBJECT with `.raw`, NOT a string.** `word_count`/`char_count` live INSIDE `content`, not at page root.
  - `char_count` = **bytes** (`len(body)`); `word_count` = whitespace fields. Definitions are inline; there are no helper functions.
- `page["ast"]` = `{headings, links, code_blocks}` — `pagebuilder.go:75-79`; element shapes at `pagebuilder.go:139-157` (`{level,text,line}`), `:200-310` (`{target,text,is_wikilink,line}`), `:322-341` (`{language,line}`).
- `page["akb"]` = `{provenance_markers, annotations}` — `pagebuilder.go:81-83`; entries carry `{type,position}` / `{type,fields,position}` (`:344-370`, `:374-401`).

`BuildOldPage` (`pagebuilder.go:87-117`) reads the on-disk file, parses FM, runs goldmark, then **delegates to `BuildPage`** (`:116`), so `old_page` has an **identical shape**. It returns `(nil, nil)` when the file does not exist (`:92-95`) — this is what `has(old_page.frontmatter)` guards in templates.

`cel/types.go:8-29` holds `Heading`/`Link`/`CodeBlock` structs with `cel:` tags, but note `BuildPage` passes plain `map[string]any` (comment `engine.go:33-34`); the structs appear unused by the builder — dead/legacy (see `.pi/.../openness-document-model.md:177` "cel/types.go dead structs"). Confirm before relying on them.

### Consumers of `word_count` / `char_count` / `content.raw` / `page.ast` (whole repo)

**`char_count`: ZERO rule consumers.** Only the definition site `internal/cel/pagebuilder.go:72`, the skill doc `internal/skill/embedded/kb-management/references/TEMPLATE.md:81`, and research/RFC prose. No Go reader, no template rule anywhere. (Matches RFC-002 D13 "char_count scrapped".)

**`word_count` in Go:** none (opaque map key; only `pagebuilder.go:71` defines it).

**`word_count` in shipped/embedded templates (all `page.content.word_count >= 50`):**
- `internal/template/embedded/adr.yaml:118` — `rule: 'page.content.word_count >= 50'` (id `min_word_count`)
- `docs/knowledge-base/.agent-kb/templates/adr.yaml:210`
- `docs/knowledge-base/.agent-kb/templates/rfc.yaml:309`
- `docs/knowledge-base/.agent-kb/templates/spec.yaml:247`
- `docs/knowledge-base/.agent-kb/templates/research.yaml:189`
- `internal/template/embedded/note.yaml` — none.
- Docs: `internal/template/embedded/adr_pass.md:28,65`; skill `TEMPLATE.md:81`.

**`page.content.raw` rule consumers (all dogfood KB templates; NO embedded template uses it):**
- `adr.yaml:205` `'page.content.raw.contains("if your task conflicts with this ADR, stop and name the conflict")'`; `:236-237` `page.content.raw.contains(".pi/")`/`".sisyphus/"`
- `rfc.yaml:358` `'!page.content.raw.contains(".pi/") && !page.content.raw.contains(".sisyphus/")'`
- `spec.yaml:172,222,227,232,237,282` — `contains`/`matches` (incl. `raw.matches("REQ-[0-9]{3}")`, `raw.matches("AC-...")`)

**`page.ast.*` rule consumers:**
- Embedded: `internal/template/embedded/adr.yaml:88,94,100,106,112` (`ast.headings`), `:124` (`ast.code_blocks`), `:169,176` (`ast.links`, incl. `l.target == page.frontmatter.supersedes`).
- Dogfood `adr.yaml:160-200` (headings), `:215` (code_blocks); `spec.yaml:177-237` (headings), `:252` (code_blocks); `rfc.yaml:315` (code_blocks); `research.yaml:207` (code_blocks).
- `page.akb.*` rules: **no template reads `page.akb`** — the only consumers are prose docs. (`.pi/.../openness-lint-engine.md:135` notes the provenance checker can't be re-expressed in CEL.)

Loop-back: `internal/lint/cel.go:60-62` is the **second production caller** of `cel.BuildPage` — it rebuilds the page map itself from `page.Body` for lint sweeps (no `old_page`). Any change to `BuildPage`'s shape hits both the write path (`cmd/akb/write.go:516` `cel.BuildPage`) and the lint path here.

**Confidence:** high (grep-verified). **Gap:** `is_wikilink` is a source-sniff heuristic in `flattenLinks` (`pagebuilder.go:277-278`); semantics of `position` (byte offset) vs `line` differ — out of scope for shape but relevant if key names move.

---

## Q2 — Required-field pre-gate (`cmd/akb/required_fields.go`)

Two tiny helpers plus one refusal wrapper:

- `checkRequiredFields(tmpl template.Template, fm *frontmatter.ParsedFrontmatter) []string` — `required_fields.go:19-29`. Iterates **`tmpl.Schema.Frontmatter`** (a `map[string]FieldSchema`), keeps keys where `field.Required && !frontmatterKeyPresent(fm, key)`, sorts. **Presence only** — never inspects value/type/enum (documented `:16-18`).
- `requiredFieldsMessage(tmpl, missing) string` — `required_fields.go:33-36` → `missing required frontmatter field(s): %s (declared required by template %q)`.
- `frontmatterKeyPresent` lives in **`cmd/akb/templates_write.go:410-419`** (not in required_fields.go): special-cases `type`→`fm.Type != ""`, `title`→`fm.Title != ""`, else `fm.Fields[key]`.

**Call sites (all in `cmd/akb/`):**
- `write.go:523-526` — after `cel.BuildPage`, before CEL rules; prints message, returns `validationFailure{}`. Comment `:521-522`: "A field the schema marks required must be present before the rules run: the rules guard on key presence, so an absent key would make them vacuously true."
- `append.go:220-222` — same refusal on append.
- `approve.go:151` via `requiredFieldsRefusal(...)` (`approve.go:221-234`). It wraps the same two helpers, prints `cannot approve '<path>': <requiredFieldsMessage>`, returns `validationFailure{}`. `approveAllDraftPages` treats that sentinel as per-page refuse-and-continue (`approve.go:262-272`).
- `templates_write.go:152-156` — reuse to check the PASS mockup carries every required field (before CEL mockup proofs).

**What breaks if `schema.frontmatter` disappears:** `tmpl.Schema.Frontmatter` is a plain nilable map; `LoadTemplates` (`internal/template/template.go:84-137`) does **not** validate it is present/non-empty. So:
1. `checkRequiredFields` returns `nil` → the write/append/approve refusal becomes a **silent no-op** (no error, gate just never fires).
2. `internal/lint/required_fields.go:76-88` `missingRequiredFields` likewise stops reporting.
3. `templates_write.go` PASS-mockup required check and `optionalKeysSuppliedByMockup` (`:395-409`) silently no-op.
4. **Worse:** the templates' CEL rules read `page.frontmatter.<required>` **unguarded** — this is the documented contract (`internal/template/embedded/adr.yaml:2-4`: "`page.frontmatter.<required>` — read UNGUARDED: akb write refuses a page that leaves out a …"). With the schema gone, an absent required key makes those rules resolve to a missing-key error, which write-time treats as a hard failure (fails closed) and lint-time degrades to a per-page issue. So flattening away the homegrown schema without a replacement required-list regresses from "clean refusal" to "CEL evaluation errors".

**Confidence:** high. **Gap:** none material.

---

## Q3 — Lint checker registration & interface

Registration: `cmd/akb/lint.go:157-167`, one `engine.AddChecker(lint.New…Checker())` per checker, after `lint.NewLintEngine()`:
```text
157 engine := lint.NewLintEngine()
158 BrokenLinks  159 Orphans  160 EmptyPages  161 MissingFrontmatter
162 RequiredFields  163 IndexConsistency  164 Citations  165 Provenance
166 CEL(CELLint)  167 TypeOrphan
```
(Engine + context built just above: `kb := &lint.KB{Root, LinkGraph, Templates, Manifest, Pages}` at `lint.go:149-155`.)

**Interface** (`internal/lint/engine.go:37-40`):
```go
type LintChecker interface {
    Name() string
    Check(ctx context.Context, kb *KB) ([]LintIssue, error)
}
```
`KB` (`engine.go:45-51`): `Root string; LinkGraph *linkgraph.SQLiteLinkGraph; Templates map[string]template.Template; Manifest []manifest.Entry; Pages []PageData`.
`PageData` (`engine.go:56-64`): `RelPath string; Content []byte; Body []byte; Frontmatter *frontmatter.ParsedFrontmatter; ProvenanceMarkers []; Annotations []; HasFrontmatter bool`. **No derived `content`/`ast` map is carried** — a checker wanting `content.word_count`/`ast.*` must rebuild via `cel.BuildPage` (as `internal/lint/cel.go:60-62` does) or the engine must be extended.
`LintIssue` (`engine.go:17-23`): `Type`(json `check`), `RuleID`, `Message`, `Path`, `Severity`.

**Checker files in `internal/lint/`:** `broken_links.go`, `orphans.go`, `empty_pages.go`, `missing_frontmatter.go`, `required_fields.go`, `index_consistency.go`, `citations.go`, `provenance.go`, `cel.go`, `type_orphan.go` (10), plus `engine.go`, `thresholds.go` (`thresholds.go:9-15` constants). AGENTS index in `internal/lint/AGENTS.md`.

**`required_fields` checker** (`internal/lint/required_fields.go`): `Name()` → `"required_fields"` (`:36-38`); `Check` (`:44-73`) skips pages lacking FM or `Type`, skips types with no template, calls `missingRequiredFields(tmpl, page.Frontmatter)` (`:76-88`) — same `tmpl.Schema.Frontmatter` walk as the write path — and emits an `error` with the **identical** wording (comment `:66-67`: "The wording matches the write path's required-field refusal"). `frontmatterFieldPresent` (`:92-103`) mirrors `frontmatterKeyPresent`.

**Interface surface a new built-in `schema` checker would need (no code, just what it must satisfy):**
- A struct implementing `Name() string` and `Check(ctx, *KB) ([]LintIssue, error)`; register with `engine.AddChecker` at `lint.go:157-167` and add a constructor `lint.NewSchemaChecker()`.
- Reads available today: `kb.Templates` (map name→`template.Template`; note `Template.Schema.Frontmatter` is the only schema today — a JSON-schema field would be a new struct field), `kb.Pages[].Frontmatter` + `.Body`/`.Content`.
- To validate over derived paths (`content.word_count`, `ast.*`) it needs the derived map: either extend `PageData` to carry `content`/`ast`, or call `cel.BuildPage(page.RelPath, page.Frontmatter, page.Body, goldmarkParse(page.Body), page.Body)` inside `Check` (pattern at `internal/lint/cel.go:42,60-70`). `PageData` today carries no such field.
- Must emit `[]LintIssue` with `Type`/`RuleID`/`Path`/`Severity`; engine aggregates `ByCheck[Name()]`.

**Confidence:** high. **Gap:** whether the checker should be CEL-backed (reuse `internal/lint/cel.go`) or a separate JSON-schema evaluator is a design choice not fixed by the code.

---

## Q4 — Template loader (`internal/template/template.go`) + `templates_write.go` flow

**`Template` struct** (`template.go:43-51`) YAML tags:
```text
name; description; dir; schema; validations; lint_rules; (unexported filename)
```
- `Schema` (`:15-17`): `Frontmatter map[string]FieldSchema`.
- `FieldSchema` (`:20-24`): `Type string`, `Required bool`, `Enum []string`.
- `ValidationRule` (`:27-32`): `ID`, `Rule`, `Requirement,omitempty`, `Expect`.
- `LintRule` (`:35-39`): `ID`, `Rule`, `Severity`, `Expect`.
(Note the RFC/AGENTS shorthand "TemplateV2" = this `Template` struct; there is no type literally named `TemplateV2`.)

**Unknown-key strictness: NOT strict today.** `LoadTemplates` (`:84-137`) does `yaml.Unmarshal(data, &raw)` into `map[string]any` (`:106`) purely to run `detectOldFormat`, then `yaml.Unmarshal(data, &tmpl)` into `Template` (`:114`). No `goccy/go-yaml` option such as `yaml.DisallowUnknownField()` is passed anywhere (grep: none), so **unknown top-level keys are silently ignored**. A template that flips to a new key (e.g. a JSON-schema block) would load without error and be silently dropped by this loader.

**`detectOldFormat` mechanics:** `template.go:73-78` loops `[]string{"required","optional","body"}`; any top-level key match → error "Template format has changed. Please update to the new schema." There is a **second, near-duplicate** `detectOldFormat` in `cmd/akb/templates_write.go:461-469` (lowercase message "template format has changed; please update to the new schema") used by the write command. Both are keyed on the pre-v2 shape only. Load-time checks beyond this: non-empty `name` (`:129-131`) and duplicate name detection (`:133-136`). No CEL compile at load time; no `FieldSchema.Type`/`Enum` validation.

`CopyDefaults` (`:141-172`) extracts embedded `embedded/*` into a base's templates dir (skip-if-exists).

**`templates_write.go` validation flow (mockup proof), high level:**
1. name regex `templateNameRe` (`:49-51`); resolve KB/templates dir and target YAML path (`:57-73`).
2. `templateExists` via `os.Stat`; a **new** template requires both `--pass` and `--fail` (`:80-84`).
3. Read template YAML → `detectOldFormat` (`:93`) → unmarshal `Template` (`:96-98`) → require `name` present and equal to the CLI arg (`:99-104`).
4. `cel.NewEnv()`; **compile every** `validations[].rule` (`:113-116`) and `lint_rules[].rule` (`:118-121`) — a syntax error aborts.
5. Resolve PASS mockup: `--pass` flag else existing `<name>_pass.md` (`:134-150`).
6. Parse mockup FM; **required-field pre-gate** on the PASS mockup (`:152-156`, via `checkRequiredFields`).
7. `buildTestPage` (`:352-356`, uses `cel.BuildPage`) and `evaluateValidations(..., nil old_page)` (`:359`) — every rule must pass.
8. For each schema-**optional** key the mockup supplies (`optionalKeysSuppliedByMockup` `:395-409`), strip it and re-evaluate (`:170-181`); an evaluation *error* must be guarded with `has()`.
9. Self-update proof: evaluate PASS page against itself as `old_page` (`:183-190`).
10. Resolve FAIL mockup (`:192-214`); **at least one** rule must fail (`:216-222`).
11. Existing template without `--force`: print `diffStrings` + page count and refuse (`:224-258`); with `--force` or new, `storage.OpenStore` + write YAML + both mockups (`:256+`).

**Confidence:** high. **Gap:** `--force` on `template write` is documented as bypassing only the existence warning, never mockup validation — confirmed by flow (force is only read at `:224`).

---

## Q5 — Template inventory (what a flatten breaks)

### (a) Embedded templates — `//go:embed embedded/*` at `internal/template/template.go:59`
Files: `adr.yaml`, `adr_pass.md`, `adr_fail.md`, `note.yaml`, `note_pass.md`, `note_fail.md`.
(Skills embed separately: `internal/skill/skill.go:17` `//go:embed embedded/*` → `kb-management/…`; docs only, no rules.)

**`adr.yaml`** rules referencing the derived/document surface — verbatim:
- `:88` `page.ast.headings.exists(h, h.level == 2 && h.text == "Context")`
- `:94` `page.ast.headings.exists(h, h.level == 2 && h.text == "Decision")`
- `:100` `page.ast.headings.exists(h, h.level == 2 && h.text == "Consequences")`
- `:106` `!page.ast.headings.exists(h, h.level == 2 && h.text == "Options")`
- `:112` `!page.ast.headings.exists(h, h.level == 2 && h.text == "Pros and Cons")`
- `:117-118` id `min_word_count` `page.content.word_count >= 50`
- `:124` `page.ast.code_blocks.all(c, c.language != "")`
- `:169` `page.ast.links.exists(l, l.is_wikilink)`
- `:176` `... page.ast.links.exists(l, l.is_wikilink && l.target == page.frontmatter.supersedes)`
- No `page.content.raw`, no `char_count`, no `page.akb.*`.
- `adr_pass.md:55,65,66,69` are prose restating the same rules (documentation, not enforced).

**`note.yaml`**: validations only `page.frontmatter.title != ""` (`:24-27`); lint rule `note_stale` uses `now`/`timestamp` only (`:29-32`). **No derived/document-surface references** — a flatten does not touch `note.yaml`.

### (b) Dogfood KB templates — `docs/knowledge-base/.agent-kb/templates/*.yaml`
Files: `adr.yaml`, `rfc.yaml`, `spec.yaml`, `research.yaml` (+ `_pass.md`/`_fail.md` each).

Derived/document-surface rules:
- **`adr.yaml`**: `page.ast.headings` ×9 (`:160,165,170,175,180,185,190,195,200`); `page.content.raw` (`:205,236,237`); `page.content.word_count >= 50` (`:210`); `page.ast.code_blocks` (`:215`).
- **`rfc.yaml`**: `page.content.word_count >= 50` (`:309`); `page.ast.code_blocks` (`:315`); `page.content.raw` (`:358`).
- **`spec.yaml`**: `page.content.raw` (`:172,222,227,232,237,282`); `page.ast.headings` ×9 incl. a `kind`-guarded one (`:177,182,187,192,197,202,207,212,217`); `page.content.word_count` (`:247`); `page.ast.code_blocks` (`:252`).
- **`research.yaml`**: `page.content.word_count >= 50` (`:189`); `page.ast.code_blocks` (`:207`). (No `page.content.raw`, no `page.ast.headings`.)
- **`char_count`: referenced by NO template anywhere.** `page.akb.*`: referenced by NO template anywhere.

`schema.frontmatter` required/optional fields (as declared; `type`/`title` always injected by `BuildPage`):
- **adr.yaml** required: `grammar, type, id, title, status(enum proposed|accepted|deprecated|superseded|rejected), created, updated, provenance(human|agent-drafted), scope(list), tags(list), summary`. optional: `deciders, supersedes, superseded_by, revisit`.
- **rfc.yaml** required: `grammar, type, id, title, status(enum draft|review|accepted|implemented|rejected|superseded|abandoned), provenance, author, steward, created, updated, scope, tags, summary`. optional: `review_by, decided_by, decided_at, resolution, supersedes, superseded_by, spawns, disposition_requested_at`.
- **spec.yaml** required: `grammar, type, id, title, kind(enum change|capability), status(enum draft|proposed|active|completed|superseded|withdrawn), provenance, created, updated, scope, verified(map)`. optional: `ratified_by, supersedes, superseded_by, anchors`.
- **research.yaml** required: `grammar, type, title, status(enum draft|final|superseded), provenance, run, as_of, created, updated, scope, tags, summary`. optional: `informs, supersedes, superseded_by`.
- Also note `FieldSchema.Type` is declared (`int`/`list`/`map`/`string`) but **never enforced anywhere** — only `Required` and `Enum`-by-CEL are actually used.

**Flatten impact:** 4 embedded-template rules + 26 dogfood-template rules reference `page.content.*`/`page.ast.*`; all are mechanical renames under RFC-002 §5–§7 (`page.content.raw`→`body`, `page.ast.*`→`page.derived.ast.*`, `page.content.word_count`→`word_count(page.content)` registry fn). `char_count` has zero consumers → safe to scrap. `page.akb.*` has zero rule consumers.

**Confidence:** high. **Gap:** whether `word_count` becomes a registry function (RFC-002 D13) or stays a projection determines whether the 5 `min_word_count` rules are rewritten or just repath'd; that's a design verdict, not visible in code.

---

## Q6 — Write/read flags & stdout

**`cmd/akb/write.go` — no `--json` today.** Flags registered in `init()` at `write.go:69-73`: `--append` (bool), `--dated` (bool, append-only), `--frontmatter` (StringArray). Nothing else. Grep for `json` in `write.go`/`read.go`/`append.go`: **zero hits**.

**stdout output** is plain human text at `write.go:575-582` (single `switch`):
```text
len(writeFrontmatter)>0 → "Updated frontmatter for %s\n"
writeAppend           → "Appended to %s\n"
default               → "Written to %s\n" + "Don't forget to update the index! `akb index add %s <summary>`\n"
```
Errors/refusals go to `os.Stderr` (e.g. `write.go:524`). A future `--json` on write would have to add a flag next to `:69-73` and branch this output block.

**`cmd/akb/read.go` — no `--json`.** No `init()` flag registration for `readCmd` at all; output is `fmt.Print(string(data))` at `read.go:65` (raw page bytes). Confirms RFC's "read --json" is future.

`--json` today exists on other commands only (pattern to follow): `cmd/akb/discover.go:43`, `links.go:56`, `lint.go:45`, `list.go:40`, `raw_status.go:54`, `search.go:38`.

**Confidence:** high (grep-verified, no false negatives). **Gap:** none.

---

## SUMMARY (decision-relevant)
1. Page map today: `page.content = {raw:string, word_count:int, char_count:int}` (object, `pagebuilder.go:69-73`), `page.ast = {headings,links,code_blocks}` (`:75-79`); **no planar `body`/`content` string key exists** — a flatten is a real shape change, and `BuildOldPage` delegates so `old_page` moves with it.
2. `char_count` has **zero rule consumers** (only `pagebuilder.go:72` + docs) → safely scrap; `word_count` has **5 template consumers**, all `page.content.word_count >= 50` (embedded adr + 4 dogfood).
3. `page.content.raw` (8 dogfood rules) and `page.ast.*` (4 embedded + ~18 dogfood rules) are the real rename blast radius; **`page.akb.*` has zero rule consumers**.
4. Required-field gate is fed solely by the homegrown `tmpl.Schema.Frontmatter` map (`required_fields.go:19-29`, `templates_write.go:410-419`); if that block vanishes it **silently no-ops** and the templates' unguarded `page.frontmatter.<required>` reads turn into CEL eval errors — a replacement required-list must ship with any flatten.
5. Loader is **lenient** (`template.go:106,114`, no `DisallowUnknownField`), so a new schema key would be silently dropped; a new built-in `schema` checker only needs `Name()`/`Check(ctx,*KB)` (`engine.go:37-40`) + one `AddChecker` line at `cmd/akb/lint.go:157-167`, but `PageData` carries no derived content/ast map today.
