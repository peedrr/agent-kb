---
grammar: 1
type: research
title: "RFC-002-S6 code recon — TemplateV2 loader, key inventory, un-routing blast radius"
status: final
provenance: agent-drafted
run: rfc-002-s6
as_of: "2026-10-09"
created: "2026-10-09"
updated: "2026-10-09"
scope:
- internal/template/**
- internal/frontmatter/**
- internal/cel/**
- cmd/akb/**
- internal/lint/**
- test/**
tags:
- rfc-002
- templatev3
- code-recon
summary: "Read-only recon of the current TemplateV2 loader, key-to-consumer inventory, embedded template corpus, type/title routing path, and test pinning, with file:line anchors."
informs:
- SPEC-003
---

# RFC-002-S6 Code Recon — TemplateV3 core format + loader

Read-only recon of the CURRENT TemplateV2 state that RFC-002-S6 (SPEC-003) will transform.
Repo: `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb`. All line anchors are file:line in that tree.

---

## 1. Current loader architecture

**Entry point:** `internal/template/template.go` (175 lines total).

- **Structs** — `internal/template/template.go:19-54`:
  - `Schema` (`:19-21`): `Frontmatter map[string]FieldSchema yaml:"frontmatter"`.
  - `FieldSchema` (`:24-28`): `Type string yaml:"type"`, `Required bool yaml:"required"`, `Enum []string yaml:"enum,omitempty"`.
  - `ValidationRule` (`:31-36`): `ID`, `Rule`, `Requirement,omitempty`, `Expect`.
  - `LintRule` (`:39-44`): `ID`, `Rule`, `Severity`, `Expect`.
  - `Template` (`:47-55`): `Name`, `Description`, `Dir`, `Schema`, `Validations []ValidationRule`, `LintRules []LintRule`, plus unexported `filename string` (`:54`).
- **No format-version key.** There is no `version:` or `format:` field anywhere; "V2" is implicit (the doc/AGENTS call it TemplateV2). A V3 needs a new discriminator.
- **Embedding** — `//go:embed embedded/*` at `template.go:59-60` (`DefaultFS embed.FS`); `init()` derives `DefaultTemplates fs.FS` via `fs.Sub(DefaultFS, "embedded")` at `:65-70`.
- **Loader** `LoadTemplates(dir) (map[string]Template, error)` — `template.go:84-130`:
  - `os.ReadDir`; `os.IsNotExist` ⇒ empty map, nil error (`:86-91`).
  - Only `*.yaml` non-dir entries (`:94-96`).
  - **Non-strict double unmarshal**: first `yaml.Unmarshal(data, &raw)` into `map[string]any` (`:106`) used *only* by `detectOldFormat` (`:109`); then `yaml.Unmarshal(data, &tmpl)` into the typed struct (`:114`). Library is `github.com/goccy/go-yaml` (`:12`). **No `yaml.Strict()` / `DisallowUnknownField()` / `KnownFields` anywhere** — unknown top-level keys and unknown `schema.frontmatter.<field>` sub-keys are silently ignored. (Confirmed by grep: no strict option in `internal/template/` or `cmd/akb/`.)
  - **Only three load-time checks**: `detectOldFormat` (`:73-80`), non-empty `name` (`:118-120`, error `missing name field in %s`), duplicate name (`:122-124`, error names both filenames).
  - **No type/enum validation, no CEL compile, no verify that `dir` is sane, no `name` vs filename check** at load.
- **`detectOldFormat(raw, filename)`** — `template.go:73-80`: rejects top-level `required`, `optional`, or `body`; message is `parse %s: Template format has changed. Please update to the new schema.` **There is a near-duplicate copy** in `cmd/akb/templates_write.go:461-469` with a lowercased/semicolon message (`...format has changed; please update to the new schema`). Both are keyed only on pre-V2 shape; neither knows V3.
- **`CopyDefaults(targetDir)`** — `template.go:134-175`: copies every embedded file (`.yaml` + `_pass.md` + `_fail.md`), skips existing targets, `MkdirAll 0750`, files `0600`.
- **Other loaders (not in the template package):**
  - `cmd/akb/template.go:104-127` `loadExampleTemplates()` reads `template.DefaultTemplates` and `yaml.Unmarshal`s each `.yaml` directly into `template.Template` — **also non-strict**, bypasses `LoadTemplates` (no `detectOldFormat`, no name/dup checks).
  - `LoadTemplates` production callers: `cmd/akb/write.go:125`, `cmd/akb/append.go:160`, `cmd/akb/approve.go:112`, `cmd/akb/lint.go:76`, `cmd/akb/index.go:166`, `cmd/akb/resolve_page.go:52`, `cmd/akb/template.go:87`.

**Confidence: high.**

---

## 2. Current key inventory (TemplateV2 YAML → Go field → consumers)

| YAML key | Go field | Consumers (non-test) |
|---|---|---|
| `name` | `Template.Name` | loader key + duplicate check (`template.go:118-124`); `frontmatter.ValidateType` (`frontmatter.go:130-140`); routing `templates[fm.Type]`; `index.go:187`; lint `type_orphan.go:42`, `cel.go:55`, `required_fields.go:52` |
| `description` | `Template.Description` | `template get` writer view (`template.go:169`), `template list` (`template.go:274`) |
| `dir` | `Template.Dir` | type-derived write path (`write.go:398-454`); `resolve_page.go:62` `declaredTypeDirs`; `index.go:172-175` candidate path |
| `schema.frontmatter.<field>.type` | `FieldSchema.Type` | **DEAD in non-test code** — read nowhere (only tests assert it; catalogued in `docs/knowledge-base/kb/research/` recon). Decorative. |
| `schema.frontmatter.<field>.required` | `FieldSchema.Required` | `cmd/akb/required_fields.go:19-29`; `internal/lint/required_fields.go:81-90`; `cmd/akb/templates_write.go:397` (`optionalKeysSuppliedByMockup`) |
| `schema.frontmatter.<field>.enum` | `FieldSchema.Enum` | **DEAD in non-test code** — enum is enforced only because templates restate it as a CEL rule (`adr.yaml`: `status in [...]`). Field itself is decorative. |
| `validations[].id` | `ValidationRule.ID` | write/append/mockup error reporting (`write.go:630`, `templates_write.go`) |
| `validations[].rule` | `ValidationRule.Rule` | CEL compile/eval at write time (`write.go:603-635`, `append.go`, `templates_write.go:113-217`, `template.go` mockup warnings) |
| `validations[].requirement` | `ValidationRule.Requirement` | writer view requirements list (`template.go:178-184`) |
| `validations[].expect` | `ValidationRule.Expect` | failure message `cel.ValidationError.Message` (`write.go:631`) |
| `lint_rules[].id` | `LintRule.ID` | `internal/lint/cel.go:64-66` |
| `lint_rules[].rule` | `LintRule.Rule` | lint sweep CEL eval (`internal/lint/cel.go:64-80`) |
| `lint_rules[].severity` | `LintRule.Severity` | issue severity (`internal/lint/cel.go` result→issue); asserted `template_test.go:100` |
| `lint_rules[].expect` | `LintRule.Expect` | lint issue message (`internal/lint/cel.go`) |
| (unexported) | `Template.filename` | duplicate-name error (`template.go:123-126`) |

Key-set notes for a strict V3 loader: the **live** consumers touch only `name`, `description`, `dir`, `schema.frontmatter.*.required`, `validations[]`, `lint_rules[]`. `FieldSchema.Type`/`.Enum` and `ValidationRule.Requirement` (writer-view only) are the soft/dead surface. No `version`, `title_field`, `summary_field`, `path`, `state_field`, `functions`, `projections` keys exist today (those are RFC-002 deferred keys; see `docs/knowledge-base/kb/research/rfc-002-phase1-scope-log.md:50-53`).

**Confidence: high.**

---

## 3. Embedded template inventory

Embedded via `//go:embed embedded/*` (`template.go:59`). All six files live in `internal/template/embedded/`. **No file carries an explicit format-version key**; all are TemplateV2 shape.

| File | Kind | Shape / keys actually used |
|---|---|---|
| `adr.yaml` | template | top-level: `name: adr`, `description`, `dir: decisions`, `schema`, `validations` (21 rules), `lint_rules` (3 rules). Frontmatter fields (10): `title`(req), `type`(req), `summary`(req), `tags`(req, list), `status`(req, enum 4), `deciders`(req), `created`(req), `updated`(req), `sources`(opt, list), `supersedes`(opt) |
| `note.yaml` | template | top-level: `name: note`, `description`, `dir: notes`, `schema`, `validations` (1: `require_title`), `lint_rules` (1: `note_stale`). Frontmatter fields (7): `title`(req), `type`(req), `summary`(opt), `tags`(opt, list), `created`(opt), `updated`(opt), `sources`(opt, list) |
| `adr_pass.md` | mockup | valid ADR page; frontmatter `type: adr`, `title`, `summary`, `tags`, `status: accepted`, `deciders`, `created`, `updated`, `sources` |
| `adr_fail.md` | mockup | frontmatter `type: adr`, `title`, `summary`, `tags`, `status: draft` (fails `valid_status` only) |
| `note_pass.md` | mockup | `type: note`, `title: Test Note` |
| `note_fail.md` | mockup | `type: note`, `title: ""` (fails `require_title`) |

`CopyDefaults` copies **all six** (the `.md` mockups included). `CopyDefaults` also feeds `TestEmbeddedTemplates` and the `template list/get --examples` empty-KB copy workflow.

**Confidence: high.**

---

## 4. `type`/`title` routing path (D10 un-routing target)

`type` and `title` are **special-cased out of the generic frontmatter map at parse time and re-injected at CEL-eval time**:

1. **Parse routes them into dedicated fields** — `internal/frontmatter/frontmatter.go:44-73`: `ParsedFrontmatter{Type string; Title string; Fields map[string]any}` (`:25-29`). In the key loop (`:61-72`) `type` and `title` become `fm.Type`/`fm.Title`; everything else lands in `fm.Fields`. Type asserts string (`:63-68`).
2. **Type routes a page to its template** — `frontmatter.ValidateType` (`frontmatter.go:130-140`) looks up `templates[fm.Type]`; missing/unknown type is refused. `akb write` calls it at `write.go:202`, `:243`, `:326`, `:373`; final routing `templates[fm.Type]` at `write.go:247-249`, `:398-402`, `:510-512`. `append.go` same at `:203-207`, `:224-227`. Unknown type ⇒ `unknown type %q`.
3. **Title is validated separately** — `frontmatter.ValidateTitle` (`frontmatter.go:143-148`, non-empty); `write.go:205`, `append.go` `validate title`.
4. **CEL pagebuilder re-injects both into the generic map** — `internal/cel/pagebuilder.go:60-66`: copies `fm.Fields` then forces `fmMap["type"] = fm.Type`, `fmMap["title"] = fm.Title`. So CEL always sees `page.frontmatter.type`/`.title`, even though they never lived in `Fields`.
5. **Required-field presence helpers special-case both keys** — `cmd/akb/required_fields.go:38-49` `frontmatterKeyPresent` and `internal/lint/required_fields.go` `frontmatterFieldPresent`: `type`/`title` checked against the dedicated fields, else `fm.Fields`.
6. **Mockup optional-key enumeration excludes both** — `templates_write.go:395-402` `optionalKeysSuppliedByMockup` skips `"type"`/`"title"` (comment: BuildPage always injects them). `withoutFrontmatterKey` (`:426-437`) copies them into the stripped struct.
7. **Re-serialization writes both back first** — `write.go:252-256`, `:352-356`, `:493-497`; `append.go:~245-253` build `allFields := {"type":fm.Type,"title":fm.Title}` then merge `fm.Fields`.
8. **Title is also a consumer of the search index and index.md** — `write.go:561` `IndexPageTx(..., fm.Title, ...)`; `index.go:209` `Title: fm.Title`.

**Implication for D10:** un-routing means `ParsedFrontmatter` stops pulling `type`/`title` out; routing then reads `fm.Fields["type"]` (or a derived field). Every anchor in items 1–8 must change together; the `pagebuilder.go:64-65` injection is the compatibility shim that currently masks the split.

**Confidence: high.**

---

## 5. Command surface consuming the loader

Loader hands these commands a `map[string]template.Template` plus the raw bytes (for `--full`/mockups):

- **`cmd/akb/template.go`** — `resolveTemplateSource()` (`:63-101`): with `--examples`, `loadExampleTemplates()` (`:104-127`, direct unmarshal into `template.Template`); otherwise `path.ResolveKB` → `assertTemplateFilesContained` → `template.LoadTemplates` (`:87`). `runTemplateGet` (`:140-197`): writer view from `Name/Description/Schema` + `Validations[].Requirement` (`:169-184`); `--full` dumps the raw `name.yaml`; `--example` reads `name_pass.md` and runs `warnMockupValidation` (`:200-285`). `runTemplateList` (`:270-287`). Names gated by `templateNameRe` (`cmd/akb/template_delete.go:44`).
- **`cmd/akb/templates_write.go`** — `runTemplatesWrite` (`:46-...`) re-reads the template YAML **itself** with `yaml.Unmarshal` (`:93-107`), runs the second `detectOldFormat` (`:464`), checks `tmpl.Name == name`, compiles all `Validations[].Rule` + `LintRules[].Rule` (`:113-121`), then mockup validation: required-field gate (`checkRequiredFields`), evaluate pass (`:159`), per-optional-key variant (`:170-178`), self/no-op variant (`:183-190`), fail-must-fail (`:216-219`). Atomic temp-file + rename swap; commit scoped to the three files. **This is the load-bearing consumer of `Schema.Frontmatter.Required` and `Validations`.**
- **`cmd/akb/template_delete.go`** — `runTemplateDelete` (`:55-...`): impact analysis via `SELECT COUNT(*) FROM documents WHERE type = ?`, `templateCommitPaths` (`:47-53`, the three files), `--force`, symlink-contained remove, then `RunLint`. Does not consume the loader structurally.
- **Consumers of `Template.Dir`:** `cmd/akb/resolve_page.go:31-70` (`typeDirsFromTemplates`/`declaredTypeDirs`), `write.go:398-454`, `index.go:166-189`.
- **Lint consumers:** `internal/lint/cel.go:55-80` (`LintRules`), `internal/lint/required_fields.go:52-90` (`Schema.Frontmatter.Required`), `internal/lint/type_orphan.go:42` (template-name set), `internal/lint/engine.go:52-53` (`KB.Templates`).

**Confidence: high** (mockup-validation internals flagged as another spec's scope).

---

## 6. Test pinning inventory

**Unit — loader (`internal/template/template_test.go`):**
- `TestLoadTemplates` (`:13-283`): loads valid new-format template incl. enum (`:39-144`), rejects old format `required`/`optional`/`body` asserting substring `Template format has changed` (`:146-208`), rejects duplicate name naming both files (`:210-233`), rejects missing name filing filename (`:235-251`), rejects malformed YAML (`:253-270`), empty map for nonexistent dir (`:272-282`).
- `TestCopyDefaults` (`:286-361`): copies `.yaml` + `.md`, skips existing, creates nested dir.
- `TestEmbeddedTemplates` (`:363-...`): `CopyDefaults`→`LoadTemplates` yields **exactly 2 templates** (`:378-380`), asserts `note.Dir=="notes"`, non-empty schema, `adr.Dir=="decisions"`, `status.Enum[0]=="proposed"`. **Pins the embedded count at 2.**

**Unit — CLI:**
- `cmd/akb/template_test.go`: writer view (`:155,207`), `--example` warnings incl. stripped/no-op/unevaluable/each-broken (`:244-575`), `--full` (`:576`), missing template/mockup (`:604,619`), list (`:648,672`), delete commit scope (`:700`), symlink guards (`:751-829`), `TestEmbeddedADRFailMockupFailsExactlyValidStatus` (`:831`), in-tree symlink (`:874`).
- `cmd/akb/templates_write_test.go`: accept/reject CEL syntax, pass-fails/fail-passes, atomic write, commit scope, unchanged-needs-no-identity, symlink guards, **guarded-vs-unguarded optional key reads** (`:691,752`), no-op-update rules, `optionalKeysSuppliedByMockup` (`:931`), `frontmatterKeyPresent` (`:958`), schema-optional `title`/`type` (`:992`), required-field missing on pass mockup (`:1154`), fail-mockup exempt (`:1211`).
- `cmd/akb/write_test.go`: type routing/unknown/missing type/title (`:276,296,310`), type-dir write (`:222,256`), `TestTypeDirsFromTemplates` (`:1916`), required-fields (`:2048,2149,2176,2210`).
- `internal/lint/required_fields_test.go:18-38`, `internal/lint/cel_test.go:34-330`, `internal/frontmatter/frontmatter_test.go:116` construct `template.Template` literals — a struct-shape change breaks them.

**Testscript (`test/testdata/`):**
- `template_examples.txt` — pins `--examples` listing sorted `adr:`/`note:`, writer view contains `requirements:`/`Title must not be empty`, `--full` contains `lint_rules:`/`note_stale`, `--example` contains `type: note`/`This is a valid note.` and no WARNING, no-KB errors, the empty-KB copy workflow (`template get --full --examples` piped to `template write`), and the `unknown type 'note'` refusal text naming `Author .agent-kb/templates/note.yaml`, `kb-management skill`, `akb template list --examples`.
- `template_write_noop.txt` — pins byte-identical `--force` re-run ⇒ `Template "note" unchanged` no new commit, `--no-commit` no-op, changed pass ⇒ real write, planted template refusal `pass mockup no longer validates`.
- `cel_old_format.txt` — pins the **load-time** message `Template format has changed` (lint + write) AND the **write-time** lowercase `template format has changed` (three spellings pinned in one file).
- Page-level template behavior also pinned in `cel_validation.txt`, `cel_lint*.txt`, `lint_required_fields.txt`, `lint_tests_adr_write.txt`, `superseded_adr.txt`, `write_errors.txt`, `exit_codes.txt`, `full_workflow.txt`.

**Confidence: high.**

---

## 7. Surprises / fragility a strict-key V3 loader must handle

1. **Two divergent `detectOldFormat` copies** with different message spellings — `template.go:73-80` (capital, period) vs `templates_write.go:461-469` (lowercase, semicolon). `cel_old_format.txt` pins both. A V3 strict loader must reconcile or keep both spellings; a single canonical message breaks the testscript.
2. **Non-strict unmarshal is load-time *and* authoring-time.** RFC-002 wants `template write` strict, load lenient (`RFC-002-open-validation-engine.md:299-302`). `write` currently re-parses with plain `yaml.Unmarshal` (`templates_write.go:100`) — the strict check must be added there, and the message shape pinned.
3. **`loadExampleTemplates` bypasses the loader entirely** (`cmd/akb/template.go:104-127`) — a V3-aware loader won't be exercised by `--examples` unless that path is changed; `template list --examples`/`get --examples` would silently accept a malformed embedded template.
4. **`FieldSchema.Type` and `.Enum` are dead** in non-test code; only `.Required` is read. A V3 `schema:` (JSON-Schema-style) block that keeps field-level `type`/`enum` must wire new consumers or those keys stay inert. `.Enum` is currently restated by hand as a CEL rule inside the embedded templates.
5. **`name` vs filename is unvalidated at load** (only `write` checks `tmpl.Name == name`). A template file renamed without editing `name` (or an embedded file) loads under the `name` key, not the filename; `copy`/`delete`/`get --full` all assume `name.yaml`.
6. **No format-version discriminator exists.** A V3 loader must decide how to tell V2 from V3 (presence of `schema:`? a new `version:` key?) with no existing field to key on; the only precedent is the shape-sniffing `detectOldFormat`.
7. **`type`/`title` dual representation is the widest blast radius** (section 4): parse (`frontmatter.go:61-72`), CEL re-injection (`pagebuilder.go:64-65`), two presence helpers (`required_fields.go:38-49`, `internal/lint/required_fields.go:90-102`), mockup enumeration (`templates_write.go:398`), re-serialization in three write paths, plus `index.go:209`/search `fm.Title`. Any "un-routing" is a cross-package change.
8. **`--full` serializes `tmpl.Schema` verbatim** — actually `--full` dumps the **raw file bytes** (`cmd/akb/template.go:174-181`), not a re-marshal; the writer view (`:169-176`) builds `{name, description, schema}` by marshalling `tmpl.Schema`. A V3 `schema:` block must flow through the writer view marshal (`:186`).
9. **Tests pin the embedded count at exactly 2** (`template_test.go:378-380`) and pin `status.Enum[0]`; adding/renaming embedded templates or dropping `enum` breaks `TestEmbeddedTemplates` and `template_examples.txt`.
10. **`CopyDefaults` copies all embedded files unconditionally** (`*.yaml`, `*_pass.md`, `*_fail.md`) — a V3 file-naming convention change ripples into the empty-KB copy workflow test.
11. **Skill embedding is separate** — `internal/skill/skill.go:17` has its own `//go:embed embedded/*` (`internal/skill/embedded/kb-management/`, holding `SKILL.md` + `references/`). It does **not** share the template embed; no coupling to change, but skill docs (`kb-management/references/TEMPLATE.md`) document the V2 template format and would go stale under V3.

**Confidence: high** for items 1–7, 9–11; **medium** for item 8 (behavior verified by reading, not by running).

---

## GAPS

- **G1:** Did not execute the binary or tests; behavior is read from source (version `VERSION` = `0.23.1` per AGENTS.md). No runtime confirmation that goccy/go-yaml silently ignores unknown keys *in this version* — inferred from API absence, matches prior recon in `docs/knowledge-base/kb/research/rfc-002-a1-recon-code.md:90-96`.
- **G2:** Did not inventory the `kb-management/references/TEMPLATE.md` skill doc content (would document V2 shape; flagged as stale risk, not read).
- **G3:** Did not read `internal/lint/cel.go` result→issue mapping in full (severity/expect plumbing) — only the template-facing loop (`:55-80`).
- **G4:** Mockup-validation mechanics inside `templates_write.go` are summarized only to the level the task asked (loader hand-off); exact rule-variant logic belongs to another spec.
- **G5:** Testscript files that only *incidentally* use templates (e.g. `exit_codes.txt`, `full_workflow.txt`) were listed by grep but not read line-by-line; their template assertions are not itemized.
- **G6:** No feature-branch / in-progress V3 code found; recon reflects current `HEAD` working tree only.
