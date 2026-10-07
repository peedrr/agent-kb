---
as_of: "2026-10-06"
created: "2026-10-06"
grammar: 1
informs:
- ADR-006
is_draft: false
provenance: agent-drafted
run: rfc-002-a2
scope:
- internal/template/**
- cmd/akb/**
- internal/skill/**
status: final
summary: Read-only code recon of the TemplateV2 loader, write-path required-field gate, embedded templates, and in-tree inventory for the RFC-002-A2 migration.
tags:
- code-recon
- templates
- migration
- rfc-002
title: TemplateV2 to V3 Migration Recon (RFC-002-A2)
type: research
updated: "2026-10-06"
---
# Recon: TemplateV2 → V3 migration (RFC-002-A2)

Read-only code recon. All anchors `file:line`. Confidence stated per section.
Repo root: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb`.

---

## 1. TemplateV2 shape + loader

**Files:** `internal/template/template.go:14-55` (types), `:73-82` (`detectOldFormat`), `:84-131` (`LoadTemplates`).

**Types** (`template.go:14-55`):
- `Schema struct { Frontmatter map[string]FieldSchema `yaml:"frontmatter"` }` — `:15-17`
- `FieldSchema struct { Type string; Required bool; Enum []string }` — `:20-24`. `Type`/`Enum` are **declared but never enforced** by the loader or the write gate (only `Required` is read — see §3).
- `ValidationRule { ID, Rule, Requirement, Expect }` — `:27-32`
- `LintRule { ID, Rule, Severity, Expect }` — `:35-40`; `Severity` is a free string, default interpreted by lint engine.
- `Template { Name, Description, Dir, Schema, Validations, LintRules, filename }` — `:43-51`

**Loader validates today** (`LoadTemplates`, `:84-131`): only
1. decode each `.yaml` into `map[string]any` (`:105-108`),
2. `detectOldFormat(raw, name)` hard-reject (`:109-111`, §2),
3. decode into `Template` (`:113-116`),
4. `Name` non-empty (`:118-120`),
5. duplicate name detection across files (`:122-124`).
No `Type`/`Enum` checking, no CEL compile at load, no unknown-key rejection.

**Non-strict unmarshal `template.go:106-114` — CONFIRMED.** `yaml.Unmarshal` (goccy/go-yaml) is used with **no `yaml.Strict()`/`KnownFields` option**; grep for `Strict|DisallowUnknownFields|KnownFields` across `internal/template/`, `cmd/akb/templates_write.go`, `internal/config/` returns nothing. Unknown **top-level** keys and unknown **`schema.frontmatter.<field>` sub-keys** are silently ignored. (In-tree evidence: `docs/knowledge-base/.agent-kb/templates/rfc.yaml:12-14` explicitly documents "V2's schema declares known keys but cannot reject unknown ones, so a typo in a purely-optional key passes silently".) RFC-002 §"Unknown template keys" (`docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md:288-291`) cites this exact line as the trap.

**Confidence:** high. **Gaps:** none — behavior is unambiguous.

---

## 2. `detectOldFormat` precedent (v1→v2 rejection)

**Two copies, both live:**
- `internal/template/template.go:73-82` — called from `LoadTemplates` (`:109`). Message:
  `parse <filename>: Template format has changed. Please update to the new schema.`
- `cmd/akb/templates_write.go:461-467` — duplicate for `akb template write` (`:93`). Message:
  `parse <filename>: template format has changed; please update to the new schema` (lowercase, `;` not `.`).

Both reject if any top-level key **`required` / `optional` / `body`** is present (`:74`, `:463`).

**Exact message shape** (the precedent RFC-002 cites), quoting `internal/template/template.go:77`:
```
parse %s: Template format has changed. Please update to the new schema.
```

**Exit code:** these return a plain `error` (not `usageError`/`internalError`/`validationFailure`); `classifyExit` (`cmd/akb/main.go:79-98`) default branch → **exit 1**, stderr `Error: execute command: load templates: parse old.yaml: Template format has changed. Please update to the new schema.` (observed in `.sisyphus/evidence/final-qa/old-format.txt:19`). Confirm `exitFailure=1` at `cmd/akb/main.go:27`.

**Testscript coverage:** `test/testdata/cel_old_format.txt` (whole file, 61 lines). Covers three paths:
- lint: `! exec akb lint` + `stderr 'Template format has changed'` (`:14-15`)
- write: `! exec akb write legacy-page.md` + `stderr 'Template format has changed'` (`:22-23`)
- template write: `! exec akb template write legacy ...` + `stderr 'template format has changed'` (`:29-30`); fixture `old-format-template.yaml` at `:31-40`.
Wired automatically: `test/integration_test.go:44-46` runs `testscript.Run` with `Dir: "testdata"` (glob picks up every `.txt`).

**Confidence:** high. **Gaps:** none.

---

## 3. Write-path required-field gate

**Implementation:** `cmd/akb/required_fields.go:16-29` (`checkRequiredFields`) + `:31-35` (`requiredFieldsMessage`). Iterates `tmpl.Schema.Frontmatter`, collects keys with `Required==true` and `!frontmatterKeyPresent`, **`sort.Strings`** (`:27`), returns all missing.

**Call sites (all four use the same helper):**
- `cmd/akb/write.go:520-526` — prints `requiredFieldsMessage` to stderr, returns `validationFailure{}`
- `cmd/akb/append.go:220-222`
- `cmd/akb/approve.go:227-233`
- `cmd/akb/templates_write.go:152-156` (pass-mockup gate)

**Exact behavior confirmed:** all missing fields listed in one message; `validationFailure` → `classifyExit` `errors.As` (`main.go:89`) → **exit 1**, no extra report (command already printed). Message shape (`required_fields.go:33-35`):
```
missing required frontmatter field(s): a, b (declared required by template "rfc")
```
Only **presence** is checked; values/types/enums are left to CEL rules (comment `required_fields.go:18-19`). Gate runs **before** CEL validations (`write.go:520-527`, then `runTemplateValidations`).

**Lint-time twin:** `internal/lint/required_fields.go:83` reads `tmpl.Schema.Frontmatter` for the sweep check (`lint_required_fields`).

**Confidence:** high. **Gaps:** none.

---

## 4. `frontmatter.Parse` type/title routing

**File:** `internal/frontmatter/frontmatter.go:18-74`.
- `ParsedFrontmatter { Type string; Title string; Fields map[string]any }` — `:18-22`
- `Parse` decodes raw map (`:44-51`), then loops (`:61-72`): key `"type"` → `fm.Type` (string-checked, `:63-68`), key `"title"` → `fm.Title` (`:69-70`), **everything else → `fm.Fields`** (`:71-72`). Non-string type/title → error (`:66`, `:70`).

**Consumers that assume routing:**
- `cmd/akb/required_fields.go:38-49` `frontmatterKeyPresent` special-cases `type`/`title` before the `Fields` lookup (comment `:36-37`).
- `cmd/akb/templates_write.go:411-426` `optionalKeysSuppliedByMockup`/`frontmatterKeyPresent` — same special-case; `optionalKeysSuppliedByMockup` **excludes** `type`/`title` (`:399-401`) because `cel.BuildPage` always injects them.
- `cmd/akb/templates_write.go:436-447` `withoutFrontmatterKey` copies `Type`/`Title` separately, strips only `Fields`.

**What changes when they become ordinary keys (RFC-002 D10, `RFC-002-...md:278-282`):** `frontmatter.Parse` stops routing; `type`/`title` land in `Fields`; the special-cases in `required_fields.go:38-49` and `templates_write.go:411-447` must be deleted; `title` additionally becomes the declared display field (`title_field:`, default `title`) feeding search title column / `index.md` / log / CEL page map. `cel.BuildPage:60-66` currently injects `fmMap["type"]=fm.Type; fmMap["title"]=fm.Title` — that injection becomes redundant once they are in `Fields` (must avoid double-write).

**Confidence:** high on routing; medium on downstream rewrite details (design, not code).

---

## 5. Embedded templates (`internal/template/embedded/*`)

Files: `adr.yaml`, `adr_pass.md`, `adr_fail.md`, `note.yaml`, `note_pass.md`, `note_fail.md`.
`internal/skill/embedded/` has **no** template YAMLs — only `kb-management/references/TEMPLATE.md` (authoring reference) and sibling skill docs.

### `embedded/adr.yaml`
- **schema.frontmatter** (`:11-47`): `title` req str, `type` req str, `summary` req str, `tags` req list, `status` req str enum[proposed,accepted,deprecated,superseded], `deciders` req str, `created` req str, `updated` req str, `sources` opt list, `supersedes` opt str.
- **date-like fields:** `created`, `updated` (both `type: string`, required).
- **validations calling timestamp():** `updated_not_before_created` `timestamp(page.frontmatter.updated) >= timestamp(page.frontmatter.created)` (`:143`); `temporal_created` `timestamp(page.frontmatter.created) <= now` (`:149`).
- **lint_rules calling timestamp()/duration():** `adr_stale` `!has(updated) || now - timestamp(updated) < duration("4320h")` warn (`:183`); `deprecated_flagged` `... now - timestamp(updated) < duration("720h")` error (`:190`); `proposed_too_long` `... now - timestamp(updated) < duration("2160h")` warn (`:197`).
- **other CEL frontmatter reads:** `title`, `status`, `tags`, `supersedes`, `created`, `updated`.

### `embedded/note.yaml`
- **schema.frontmatter** (`:4-25`): `title` req str, `type` req str, `summary` opt str, `tags` opt list, `created` opt str, `updated` opt str, `sources` opt list.
- **date-like:** `created`, `updated` (optional).
- **validations:** only `require_title` `page.frontmatter.title != ""` (`:27`); no timestamp.
- **lint_rules:** `note_stale` `!has(updated) || now - timestamp(updated) < duration("2160h")` warn (`:34`).

### `embedded/adr_fail.md` / `note_fail.md`
FAILS convention is documentary comments, not machine-parsed (see §6).

**Confidence:** high (fully read).

---

## 6. `template write` mockup validation flow

**File:** `cmd/akb/templates_write.go:64-380` (`runTemplatesWrite`).
Order:
1. name regex (`:65-68`), resolve KB/templates dir (`:70-82`); new template requires `--pass`+`--fail` (`:86-90`).
2. read + `detectOldFormat` duplicate (`:87-95`) + decode `Template` + name-match check (`:97-107`).
3. CEL env + compile every validation and lint rule (`:109-124`).
4. pass mockup read (flag, else reuse existing `_pass.md`, `:126-145`); parse (`:147-150`).
5. **required-fields gate on pass mockup** (`:152-156`) via `checkRequiredFields`/`requiredFieldsMessage`; failure embeds the whole mockup and says `Provide updated mockup with --pass <path>`.
6. `buildTestPage` + evaluate validations vs `old_page=nil` (`:158-165`).
7. **optional-key-stripping proof** (`:167-181`): for each schema-optional key the mockup supplies (`optionalKeysSuppliedByMockup`, `:397-410`), re-evaluate once with that key removed (`withoutFrontmatterKey`, `:436-447`). Rule must survive absent optional keys; if a rule **errors** (unguarded access) the message tells the author to guard with `has()` or mark the field `required: true` (`:176-178`).
8. **self-`old_page` proof** (`:183-192`): re-evaluate pass page against itself as `old_page`; error tells author to guard with `has()`.
9. fail mockup read/parse (`:194-222`); evaluate; **require at least one failure** (`:224-230`, "expected at least one validation to fail, but all passed").
10. overwrite confirmation/diff/page-count (`:232-263`), no-op detection (`:280-289`), preflight (`:292-296`), atomic temp+rename swap (`:298-345`), commit (`:347-351`).

**FAILS convention parsing:** **not parsed by any Go code.** `grep -rn "FAILS" --include=*.go` = zero hits. In mockups it is a human-readable HTML comment consumed only by the *test harness author* to know which rule the fail mockup intends to trip; the code only asserts `len(failFailed) > 0`. Examples: `internal/template/embedded/adr_fail.md:16` `<!-- FAILS: valid_status — ... -->`, `embedded/note_fail.md:8`, `docs/knowledge-base/.agent-kb/templates/spec_fail.md:30`. RFC-002 proposes extending this to `<!-- FAILS: schema: /frontmatter/status -->` (`RFC-002-...md:299-300`) — **but the existing convention is declarative only, so a V3 implementer must add the parser.**

**Confidence:** high. **Gaps:** none.

---

## 7. In-tree TemplateV2 inventory

Complete set of template YAMLs (all under `internal/template/embedded/` and `docs/knowledge-base/.agent-kb/templates/`). No other template YAMLs exist anywhere (find/grep). No test/testdata template YAMLs — testdata uses inline fixtures.

| Template | File | required | optional | date-ish (type: string) | timestamp()/duration() rules |
|---|---|---|---|---|---|
| **adr** (embedded) | `internal/template/embedded/adr.yaml` | title,type,summary,tags,status,deciders,created,updated | sources,supersedes | created,updated | validations `updated_not_before_created`(:143), `temporal_created`(:149); lints `adr_stale`(:183), `deprecated_flagged`(:190), `proposed_too_long`(:197) |
| **note** (embedded) | `internal/template/embedded/note.yaml` | title,type | summary,tags,created,updated,sources | created,updated | lint `note_stale`(:34) |
| **adr** (KB) | `docs/knowledge-base/.agent-kb/templates/adr.yaml` | grammar(int),type,id,title,status,created,updated,provenance,scope,tags,summary | deciders,supersedes,superseded_by,revisit | created,updated | validations `updated_not_before_created`(:150), `temporal_created`(:155); lints `adr_stale`(:220), `proposed_too_long`(:225), `deprecated_flagged`(:230) |
| **rfc** (KB) | `docs/knowledge-base/.agent-kb/templates/rfc.yaml` | grammar(int),type,id,title,status,provenance,author,steward,created,updated,scope,tags,summary | review_by,decided_by,decided_at,resolution,supersedes,superseded_by,spawns,disposition_requested_at | created,updated,review_by,decided_at,disposition_requested_at | validations `review_by_future`(:209), `decided_at_not_future`(:241), `disposition_requested_at_not_future`(:247), `updated_not_before_created`(:298-299), `temporal_created`(:304); lints `rfc_at_risk`(:322-323), `rfc_stale`(:331), `rfc_dormant`(:339) |
| **spec** (KB) | `docs/knowledge-base/.agent-kb/templates/spec.yaml` | grammar(int),type,id,title,kind(enum),status(enum),provenance(enum),created,updated,scope,verified(map) | ratified_by,supersedes,superseded_by,anchors | created,updated, nested `verified.at` / `verified.next_review_by` (map, never declared date) | validations `updated_not_before_created`(:162), `temporal_created`(:167); lints `spec_stale`(:257), `proposed_too_long`(:262) |

**Field types declared by CEL:** only `int`, `string`, `list`, `map` exist in `FieldSchema.Type` usage; there is **no `format`/date type today**. The "date bridge" is implicit: `internal/cel/pagebuilder.go:38-46` `convertDateField` coerces any string that parses as RFC3339 or `2006-01-02` into `time.Time` before CEL sees it.

**Which break/warn under a V3 schema-declared date bridge:**
- Templates that declare `created`/`updated`/`review_by`/`decided_at`/`disposition_requested_at` as plain `type: string` and then call `timestamp()` will **break** if the V3 bridge only coerces fields the schema declares with a date `format` — or **warn** if the bridge keeps the heuristic. Every one of the 5 YAMLs above calls `timestamp()` on `created`/`updated`, so all 5 are affected.
- `spec.yaml` is the sharpest: `verified.at`/`verified.next_review_by` are **nested keys inside a `map` field** (guard doctrine `spec.yaml:6-8`: required gate checks top-level only). A V3 schema is JSON Schema, which *can* declare nested date formats — today the heuristic catches them incidentally. `anchors[].*` is a list-of-maps with no dates.
- `rfc.yaml:20-31` documents two V2 engine interactions the V3 bridge must re-specify: auto-bump of `updated` after validation, and the bridge coercing date-like strings so `matches()` has no overload on date fields (format checks must go via `timestamp()`).
- All KB templates read `status`, `id`, `grammar`, `scope`, `tags`, `summary`, `provenance` unguarded in validations — these must become schema-required keys under V3 or the guard invariant breaks (`RFC-002-...md:279-281`).

**Confidence:** high (all five YAMLs read fully or schema region fully read). **Gaps:** exact V3 date-format vocabulary is a design decision, not yet in code.

---

## 8. Migration machinery

**None exists.** `grep -rni "migrate"` over `cmd/akb/` and `internal/template/` = **zero hits**. No `akb template migrate` subcommand, no conversion code. Plan-level notes only (`.sisyphus/drafts/cel-validation-plan.md:235-240` records the v1→v2 decision as **hard break, no migration tool, no dual format**; `.sisyphus/plans/cel-validation.md:113` "NO backwards compatibility / migration tooling (hard break)"). RFC-002 (`docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md:277`) calls the V3 retirement a "hard-rejection migration path (precedent: detectOldFormat)".

**Hook point for `akb template migrate`:** `cmd/akb/template.go:26-28` defines `templateCmd`; subcommands self-register in `init()`:
- `template.go:53-56` adds `templateGetCmd`, `templateListCmd`, `RootCmd.AddCommand(templateCmd)`
- `templates_write.go:44-50` adds `templateWriteCmd`
- `template_delete.go` adds `templateDeleteCmd` (`:41`)
A new `cmd/akb/template_migrate.go` would define the cobra command and `templateCmd.AddCommand(...)` in its own `init()`, mirroring `template_delete.go`. It would read via `template.LoadTemplates` / raw YAML (`internal/template/template.go:84`) and rewrite through the same temp+rename+commit path as `templates_write.go:298-351`.

**Confidence:** high (absence verified by grep). **Gaps:** ADR for retirement + migration mechanics is a downstream decision (a2 LOG flags it).

---

## 9. Testscript tests for old-format rejection / v1→v2 messaging

**One file:** `test/testdata/cel_old_format.txt` (read fully, 61 lines). Sections:
- §1 lint refuses old-format template (`:7-15`) — `stderr 'Template format has changed'`
- §2 write refuses (`:17-23`)
- §3 template write refuses (`:25-30`) — lowercase `'template format has changed'`
Fixtures inline: `old-format-template.yaml` (`:31-40`), `legacy-page.md` (`:42-47`), `pass.md` (`:49-...`).

**How v2's rejection was messaged to users (the precedent to copy):** a single, format-agnostic sentence naming the file and telling the author the format changed and to update to the new schema — `parse <file>: Template format has changed. Please update to the new schema.` (`internal/template/template.go:77`), surfaced with the CLI `Error:` prefix at exit 1. No automated conversion offered; the message points forward only. The `.sisyphus/plans/cel-validation.md:1630,1654` acceptance text pins the substring "Template format has changed. Please update to the new schema." and asserts non-zero exit.

**Confidence:** high. **Gaps:** no test asserts the *exact* full string beyond the substring; no dedicated Go unit test for `detectOldFormat` in `internal/template/template_test.go` found by grep (testscript only) — verify before relying on unit coverage.

---

## Cross-cutting notes / risks for the a2 research

1. **Two divergent `detectOldFormat` messages** (`template.go:77` capital/period vs `templates_write.go:464` lower/semicolon). A V3 hard-reject should pick one shape; the existing testdata asserts both spellings in different sections.
2. **`FieldSchema.Type`/`Enum` are dead weight today** — only `Required` is read (`required_fields.go:21-27`, `internal/lint/required_fields.go:83`). A JSON Schema `schema:` block replaces all of `Schema`; the required-set read sites are `cmd/akb/required_fields.go:21`, `internal/lint/required_fields.go:83`, `cmd/akb/templates_write.go:397`.
3. **Non-strict unmarshal is load-time *and* `template write`-time** (`template.go:113`, `templates_write.go:100`). RFC-002 wants `template write` strict, load lenient (`RFC-002-...md:288-291`).
4. **`type`/`title` dual representation** is the widest blast radius: `frontmatter.Parse:61-72`, `cel.BuildPage:60-66`, `required_fields.go:38-49`, `templates_write.go:399-401,411-447`.
5. **`--full` output** (`cmd/akb/template.go:174`) serializes `tmpl.Schema` verbatim — the V3 `schema:` block must flow through this path (`template get --full`).
6. **Date bridge is heuristic, not schema-declared** (`internal/cel/pagebuilder.go:38-46`); all 5 in-tree templates depend on it for `timestamp()`.
