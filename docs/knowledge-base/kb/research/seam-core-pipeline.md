---
as_of: "2026-09-26"
created: "2026-09-26"
grammar: 1
informs:
- RFC-001
- RFC-002
is_draft: false
provenance: agent-drafted
run: json-cel
scope:
- internal/cel/**
- internal/frontmatter/**
- cmd/akb/**
status: final
summary: Deep-dive on where a combined frontmatter+body JSON view and schema validation can run; 8 of 22 shipped CEL rules read body-derived views, so derived keys must stay output-only on write.
tags:
- json-schema
- cel
- page-model
- recon
title: "JSON ⇄ CEL seam: core pipeline deep-dive"
type: research
updated: "2026-09-26"
---
# JSON ⇄ CEL Seam: Core Pipeline Deep-Dive (READ-ONLY recon)

Scope: where a JSON view + JSON Schema validation can and cannot run, given the owner's
constraint that **structured JSON I/O = frontmatter metadata + markdown content COMBINED**
(metadata-only JSON is rejected). All anchors are `file:line` at commit `073f275`.

Verdict up front: the constraint is satisfiable, but only if the JSON document akb validates
carries **akb-materialized derived views** (`content.word_count`, `ast.headings/links/code_blocks`,
`akb.*`) because 8 of the 22 shipped validation rules read body-derived facts. And it only stays
honest if those derived keys are **output-only on write** — the caller supplies `content.raw`
(markdown text) + `frontmatter`, never `ast`.

---

## 1. THE PAGE MODEL — every key, its Go type, and how it is derived

Producer: `internal/cel/pagebuilder.go:51-86` (`BuildPage`). It returns a plain
`map[string]any` — **not** the typed structs in `internal/cel/types.go:6-24`, which are dead in
production (only referenced from `internal/cel/types_test.go:12`). The CEL env is
`MapType(StringType, DynType)` (`internal/cel/engine.go:31-33`) precisely so no adapter is needed.

| Page key | Go type as built | Source | Class |
|---|---|---|---|
| `file.path` | `string` (`filepath`-joined, write.go:447 `ToSlash`) | CLI arg + type-dir derivation | caller/session input (path) |
| `file.name` | `string` = `filepath.Base(relPath)` (pagebuilder.go:56) | derived from path | path-derived |
| `file.dir` | `string` = `filepath.Dir(relPath)` (pagebuilder.go:57) | derived from path | path-derived, **OS separator leak** |
| `frontmatter.<k>` | `any` — YAML-decoded value; **date-like top-level strings coerced to `time.Time`** (`convertDateField`, pagebuilder.go:29-45, applied at 60-64) | `fm.Fields` | **raw input** (coerced) |
| `frontmatter.type` | `string` (pagebuilder.go:65) | `fm.Type` (frontmatter.go:64-70 routes `type` out of Fields) | raw input |
| `frontmatter.title` | `string` (pagebuilder.go:66) | `fm.Title` (frontmatter.go:71-76) | raw input |
| `content.raw` | `string` = body verbatim (pagebuilder.go:67-68) | body bytes (frontmatter.go:90-127 `extractBody`) | **raw input** |
| `content.word_count` | `int` = `len(strings.Fields(body))` (pagebuilder.go:69) | body | **DERIVED from body** |
| `content.char_count` | `int` = `len(body)` — **bytes, not characters** (pagebuilder.go:70) | body | **DERIVED from body** |
| `ast.headings[]` | `[]map[string]any` of `{level int, text string, line int}` (pagebuilder.go:143-161) | goldmark AST walk over `source` | **DERIVED from body** |
| `ast.links[]` | `[]map[string]any` of `{target, text string, is_wikilink bool, line int}` (pagebuilder.go:177-312) | goldmark `ast.Link` **merged with** `markdown.ParseWikilinks` (pagebuilder.go:205), sorted by source offset (pagebuilder.go:302) | **DERIVED from body** |
| `ast.code_blocks[]` | `[]map[string]any` of `{language string, line int}` (pagebuilder.go:322-343) | fenced code blocks only | **DERIVED from body** |
| `akb.provenance_markers[]` | `[]any` of `{type string, position int}` (pagebuilder.go:347-367) | regex `\^\[...\]` minus code/inline-code/comment ranges | **DERIVED from body** |
| `akb.annotations[]` | `[]any` of `{type "olw-auto", fields map[string]string, position int}` (pagebuilder.go:371-392) | regex `<!--\s*olw-auto:...-->` minus code ranges only | **DERIVED from body** |
| `old_page` | `map[string]any` or `nil` (`BuildOldPage`, pagebuilder.go:91-116) | **on-disk file** re-read + re-parsed | **external context** |
| `now` | `time.Time` (injected at evaluate time: write.go:611, lint/cel.go:74) | clock | **external context** |

Notes that matter for the seam:

- `char_count` is a **byte** count (pagebuilder.go:70) while `word_count` is whitespace-split
  (markdown syntax counts as words). Neither equals anything JSON Schema's `minLength`/`maxLength`
  (code points) can compute.
- `ast.*` are **not reproducible from markdown by a caller without reimplementing akb**: `links[]`
  is a merge of goldmark and a hand-written wikilink parser with documented
  suppression/normalization rules (pagebuilder.go:158-313), and `is_wikilink` is decided from raw
  bytes at the node position (pagebuilder.go:268-276).
- The `akb.*` extraction **duplicates** `internal/markdown`'s unexported exclusion logic
  (acknowledged at pagebuilder.go:117-120; copies at 345, 369, 415-524). The JSON view would
  promote this duplication to a public contract — see Risks.
- `BuildPage`'s `source` parameter is not pinned: production passes `body` for both `body` and
  `source` (write.go:508-510, append.go:205-207, lint/cel.go:61-62, templates_write.go:341-344),
  but `pagebuilder_test.go:257` passes the **full** document including frontmatter. If `line` is
  published in JSON, a convention must be fixed first.

## 2. JSON-VIEW FEASIBILITY — rule-by-rule over the shipped templates

Assumed rich view (call it `document`):

```json
{ "file": {...},
  "frontmatter": { ...as written... },
  "content": { "raw": "<markdown>", "word_count": 0, "char_count": 0 },
  "ast": { "headings": [...], "links": [...], "code_blocks": [...] },
  "akb": { "provenance_markers": [...], "annotations": [...] } }
```

All 22 validations + 4 lint rules: `adr.yaml` 21+3 (`adr.yaml:51-177`, `179-201`),
`note.yaml` 1+1 (`note.yaml:28`, `33`).

| # | Rule (anchor) | Reads | JSON Schema over the rich view? | Why / why not |
|---|---|---|---|---|
| 1 | `require_title` adr:51 / note:28 | frontmatter.title | **YES** | `required:["title"]` + `properties.title.minLength:1` |
| 2 | `title_no_period` adr:57 | frontmatter.title | **YES** | `not:{pattern:"\\.$"}` |
| 3 | `title_style` adr:63 | frontmatter.title | **YES** | `pattern:"^[A-Z]"` + `maxLength:80`. CEL `size()` and JSON Schema `maxLength` both count code points |
| 4 | `valid_status` adr:69 | frontmatter.status | **YES** | `enum: [...]` — duplicates the **dead** `schema.frontmatter.status.enum` (template.go:24-29 `FieldSchema.Enum`, never read; only `.Required` is used) |
| 5 | `tags_nonempty` adr:75 | frontmatter.tags | **YES** | `minItems:1` |
| 6 | `tags_format` adr:81 | frontmatter.tags[] | **YES** | `items:{type:"string",pattern:"^[a-z0-9-]+$"}`. Dialect note only (see below) |
| 7 | `require_context` adr:87 | `ast.headings` | **YES, ONLY over materialized `ast`** | `contains:{properties:{level:{const:2},text:{const:"Context"}},required:["level","text"]}`. **Inexpressible over `content.raw`** — JSON Schema cannot parse markdown |
| 8 | `require_decision` adr:93 | `ast.headings` | **YES (same as 7)** | Same |
| 9 | `require_consequences` adr:99 | `ast.headings` | **YES (same as 7)** | Same |
| 10 | `disallow_options` adr:105 | `ast.headings` | **YES, only over materialized `ast`** | `not:{contains:{...}}`. Expressible, but error text is coarse ("must NOT be valid") |
| 11 | `disallow_pros_cons` adr:111 | `ast.headings` | **YES (same as 10)** | Same |
| 12 | `min_word_count` adr:117 | `content.word_count` | **YES, only over materialized `content.word_count`** | `minimum:50`. From `content.raw` alone: impossible (no word counting) |
| 13 | `code_blocks_have_language` adr:123 | `ast.code_blocks[]` | **YES, only over materialized `ast`** | `items:{required:["language"],properties:{language:{minLength:1}}}` |
| 14 | `valid_state_transition` adr:129 | `old_page.frontmatter.status` × `frontmatter.status` | **NO** | Two documents. An instance-scoped validator has no second instance. `if/then` can gate on `status` but cannot see the previous value; AJV's non-standard `$data` cannot express a transition matrix either |
| 15 | `created_immutable` adr:136 | `old_page` × `frontmatter.created` | **NO** | Equality with a value outside the document |
| 16 | `updated_not_before_created` adr:142 | two sibling strings | **NO** | No value-to-value comparison in standard JSON Schema (no `$data`); `format` is annotation-only, and even with format-assertion it checks shape, not order |
| 17 | `temporal_created` adr:148 | `frontmatter.created` × `now` | **NO** | `now` is not a property of the instance; JSON Schema has no clock |
| 18 | `superseded_requires_field` adr:155 | frontmatter.status → supersedes | **YES** | `if:{properties:{status:{const:"superseded"}},required:["status"]},then:{required:["supersedes"],properties:{supersedes:{minLength:1}}}` |
| 19 | `proposed_has_no_supersedes` adr:162 | frontmatter.status → supersedes | **YES** | `if: status const "proposed", then: {not:{required:["supersedes"]}}`. Semantics match CEL `!has()` (both treat "key present with any value" as present) |
| 20 | `require_wikilink` adr:168 | `ast.links[]` | **YES, only over materialized `ast`** | `contains:{properties:{is_wikilink:{const:true}}}`. A `pattern:"\\[\\["` on `content.raw` is **not** equivalent (false positives in code blocks/inline code, which `ParseWikilinks` excludes, wikilink.go:428-451) |
| 21 | `superseded_requires_link` adr:175 | frontmatter.supersedes × `ast.links[].target` | **NO** | Cross-subtree existential: needs *some* array element compared against a sibling root property. `contains` cannot reference `frontmatter.supersedes` |
| 22 | `adr_stale` adr:182 (lint) | `updated` × `now` × `duration` | **NO** | Clock + arithmetic (`now - updated < 4320h`) |
| 23 | `deprecated_flagged` adr:189 (lint) | status + `updated` × `now` | **NO** | Clock + conditional |
| 24 | `proposed_too_long` adr:196 (lint) | status + `updated` × `now` | **NO** | Clock + conditional |
| 25 | `note_stale` note:33 (lint) | `updated` × `now` | **NO** | Clock |

25 rows above cover the 26 shipped rules (row 1 covers both `adr:51` and `note:28`).

**Tally: 17 of 26 expressible in JSON Schema; 8 of those 17 require akb-materialized derived
views (#7-13, #20); 9 are forever CEL (#14-17, #21-25).**

Two dialect caveats that do not change the tally but must be pinned in docs:

- **Regex dialect**: CEL `matches` = RE2, JSON Schema `pattern` = ECMA-262. `^[A-Z]`,
  `^[a-z0-9-]+$`, `\\.$` are portable; lookarounds/backrefs are ECMA-only and must be banned by
  the template authoring rules (they already are, de facto, since RE2 rejects them in CEL).
- **`not` + `contains`** (#10, #11) is draft-06+; `contains` item-level `properties` needs
  draft-07+; `if/then/else` (#18, #19) needs draft-07+. `minContains`/`maxContains` are 2019-09.
  Pin a dialect in the template format (recommend 2020-12).

## 3. BODY PARSERS — is the derived view a pure function of the markdown?

Yes, with one caveat, for all four extractors:

- `markdown.ParseWikilinks(content string)` (`internal/markdown/wikilink.go:90-161`) — pure
  function of the content string. No KB state, no file access, no DB. Deterministic ordering
  (scan left to right, `scanWikilinkTokens` wikilink.go:44-79; one entry per token; suppression
  rules documented at wikilink.go:101-108). Output `Wikilink{Target, Display, HasHeading, Heading,
  Destination, Start, End}` (wikilink.go:13-27).
- `markdown.ParseProvenanceMarkers(content string)` (provenance.go:22-46) — pure; exclusions from
  `computeExclusions` (fenced code + inline code + HTML comments, wikilink.go:428-451).
- `markdown.ParseAnnotations(content string)` (annotation.go:24-46) — pure; excludes fenced and
  inline code **only** (an annotation *is* an HTML comment, annotation.go:33-36).
- goldmark parse for `ast.*` (pagebuilder.go:143-343) — pure function of the bytes.

Caveats:

1. **Two implementations of the same exclusion semantics.** `internal/cel/pagebuilder.go` carries
   its own copies of fenced-code/inline-code/comment range logic (pagebuilder.go:345, 369,
   415-524) `internal/markdown`'s unexported equivalents. Today they are described as duplicates
   (pagebuilder.go:117-120); publishing the derived view makes any divergence a user-visible
   contract break. Recommend `internal/markdown` export the three parsers and have `BuildPage`
   (or the JSON view builder) call them; then `akb.provenance_markers` and
   `markdown.ParseProvenanceMarkers` cannot drift.
2. `akb.annotations[].position` is an offset into the **full** page content in the lint path
   (`cmd/akb/lint.go:130` parses `string(content)`) but into the **body** in the page map
   (pagebuilder.go:82 uses `bodyStr`). Two different offsets for the same annotation. A published
   view must pick one and normalize.
3. What is **not** in the view and could tempt a JSON consumer: link **resolution**
   (`resolved` lives only in the link graph, `internal/linkgraph/sqlite.go:78`), page backlinks,
   and search rank. Those need the DB; `akb read` today does not open it (read.go:55 uses
   `storage.NewFilesystemProvider`, unlike write.go:121 `NewGitProvider` + db at write.go:99).
   Any read `--json` that surfaced them would add a DB dependency + a new "run `akb index
   rebuild`" failure mode on a command that currently cannot fail that way.

## 4. WRITE `--json` ASSEMBLY

### Today's assembly points

- New page from stdin: `writeContent = stdinContent` passthrough (write.go:374); re-serialization
  only when akb mutates frontmatter (`created` default write.go:376-379, `is_draft:true` strip
  381-386, `updated` bump 476-483) → `allFields` map + `yaml.Marshal` (write.go:484-496).
- `--frontmatter`: read page → mutate → re-serialize (write.go:244-255).
- `--append`: read page → `body + "\n" + appended` (write.go:344-355; append.go:181-196 identical
  logic); `appended` comes from raw stdin.
- Then one pipeline for all three: `BuildPage` (write.go:510) → required-field gate
  (write.go:515-518 → required_fields.go:19-33) → CEL validations (write.go:521 →
  `runTemplateValidations` write.go:595-643) → `store.Write` (write.go:539) → index + linkgraph in
  one tx (write.go:527-561, relPath, `string(body)`, tags, summary, type).

### Design A — raw markdown as a JSON string (RECOMMENDED)

```json
{ "frontmatter": { "type": "adr", "title": "...", "status": "accepted", ... },
  "content": { "raw": "# Title\n\n## Context\n..." } }
```

- `frontmatter` → `ParsedFrontmatter` (type/title routed as in frontmatter.go:64-77);
  `content.raw` → `body`. Everything else in the view (`content.word_count/char_count`, `ast.*`,
  `akb.*`, `file.*`) is **output-only**. Supplying any of them must be a hard error (exit 2,
  structured), never silently ignored: a caller who sends `ast.headings:[{level:2,text:"Context"}]`
  with a body lacking `## Context` must not believe the derived fact was accepted.
- akb still owns the managed mutations and must echo the **effective** document: `created` default
  (write.go:376-379), `is_draft` normalization (381-386 + `frontmatter.IsDraft` frontmatter.go:151-166),
  `updated` bump (476-483). Input doc ≠ output doc is fine as long as the response carries the
  document that was validated and written.
- Path: keep the CLI arg authoritative (as today). `file.path` in the payload is advisory; if it is
  honoured, the type-dir derivation (write.go:414-448) and the guards (`index.md`/`log.md`,
  `raw/`, `..`, absolute: write.go:426-436 + `internal/path.ResolveKBPath` path.go:90-109) must
  still apply.
- Duplication/consistency risk: **none between metadata and body** under this design, because body
  is opaque text and every derived fact is computed once, by akb. The only residual cross-check is
  `frontmatter.type` vs the template dir (write.go:438-444) — already enforced.
- Cost: JSON writes always take the re-serialize path, so the output frontmatter is normalized
  (sorted keys + quoted date-like strings). Document it; it is already what `--frontmatter` and
  `--append` do today.

### Design B — structured content, akb generates markdown (REJECT)

- The structured view is **lossy by construction**: `ast` holds only headings, links, code blocks
  (pagebuilder.go:73-77). Paragraphs, emphasis, lists, tables, blockquotes, images, HTML are
  absent. Rendering markdown from it would delete most of the document.
- If both `raw` and structured content are accepted, they can disagree; whichever wins, the fact
  is stated twice — precisely the duplication the owner rejects for metadata.
- Rules would then run over akb's own rendering, so a failure message ("no `## Context`") would
  point at a document the author never wrote, and writes would stop being byte-faithful (which is
  what `git diff` review and `--append`'s read-modify-write depend on).
- Verdict: **`content.raw` is required and authoritative on write.** Structured content is a
  read-side view plus an optional future syntax-sugar preprocessor, never the write source.

### JSON→YAML type fidelity (measured, goccy/go-yaml v1.19.2)

| Behavior | Result | Consequence |
|---|---|---|
| `yaml.Marshal(map[string]any)` key order | **lexicographically sorted** (`encode.go:682-692`, `fmt.Sprint` keys) | every JSON write reorders frontmatter; existing behavior, not new |
| YAML ints on decode | `uint64` (probe) | `count: 3` stays `3` |
| `json.Unmarshal` into `map[string]any` then `yaml.Marshal` | `"count":3` → **`count: 3.0`** | **do not** hand `encoding/json` output to `yaml.Marshal` |
| `json.Decoder` + `UseNumber` then `yaml.Marshal` | `json.Number` → **`"3"` (quoted string)** | also wrong |
| `yaml.JSONToYAML` (goccy) | `3` → `3`, `1.0` → `1.0`, `"2024-01-01"` → `created: "2024-01-01"`, `n: null` key quoted | correct types; **preserves JSON document key order** (not akb's sorted order) |
| `yaml.YAMLToJSON` | round-trips the above losslessly | usable for the read view |
| unquoted `created: 2024-01-01` on YAML decode | stays `string` (probe) — NOT auto-`time.Time` | the `time.Time` seen by CEL comes only from `convertDateField` (pagebuilder.go:29-45) |
| date-like strings on `yaml.Marshal` | emitted quoted: `created: "2024-01-01"` | cosmetic diff churn on any re-serialize; already the case today |

Recommendation: one assembly function for JSON writes that (a) converts JSON→YAML with `JSONToYAML`
or an equivalent typed decode, (b) re-parses it through `frontmatter.Parse` so the exact write
pipeline runs unchanged, (c) re-marshals from akb's own map so key order/quoting is akb's single
canonical form. Do **not** add a second writer.

## 5. READ `--json` SHAPE

- Today: `akb read` prints the file verbatim from a `FilesystemProvider` (read.go:34-66). No
  frontmatter parse, no goldmark, no DB.
- **Derived views**: producing `ast`/`akb`/`word_count` costs one goldmark parse + one wikilink
  scan + two regex scans per page (pagebuilder.go:73-83). That is exactly the per-page cost `akb
  lint` already pays (lint/cel.go:60-62), so it is cheap per invocation and measurable per page —
  but a script looping over a 1000-page KB pays it 1000 times, and `read` is the command agents
  call most.
- **Consumer value of the derived views is high and non-substitutable**: they are the *only* way a
  caller can see what CEL actually sees (there is no `akb explain`/`akb eval` command today). Two
  concrete uses: pre-flight a draft before writing; self-repair a failing rule without guessing how
  akb counted words or normalized a wikilink target.
- Recommendation: two tiers.
  - Default `akb read --json` → `{file, frontmatter (raw as parsed), content:{raw}, is_draft}`
    (one file read + one YAML parse; no goldmark).
  - `akb read --json --derived` (or `--full`) → adds `content.word_count`, `content.char_count`,
    `ast`, `akb` from the same builder the write path uses.
- **Round-trip invariant (critical):** the JSON view must carry frontmatter values **as parsed from
  YAML**, never the CEL-coerced map. `BuildPage` rewrites date-like strings to `time.Time`
  (pagebuilder.go:60-64), which JSON-marshals as `2026-01-05T00:00:00Z` while the file says
  `2026-01-05`. If `read --json` dumped the page map, `read --json | write --json` would silently
  mutate every date field. Name the two things separately: **`page`** = CEL evaluation input
  (internal, coerced), **`document`** = the JSON view (raw + derived block, CLI boundary).
- Also normalize `file.dir` to `/` (pagebuilder.go:57 uses `filepath.Dir`) and emit an
  `akb_version`/view `schema_version` so consumers can detect shape changes.

## 6. TEMPLATE IMPLICATIONS

Proposed split, with no overlap:

| Block | Holds | Evaluated |
|---|---|---|
| `schema.json` (new; e.g. `json_schema:`) | Structural constraints over the **combined document**: `frontmatter` properties/required/enum/pattern/length/items, plus the derived paths it can address (`content.word_count.minimum`, `ast.headings` `contains`, `ast.code_blocks.items`, `ast.links` `contains`) | write time, **before** CEL; and sweep time |
| `validations[]` (`template.go:31-37`) | CEL over context/arithmetic/cross-document: `old_page` state machines, immutability, sibling-value ordering, frontmatter↔derived cross-references, `now` | write time (fail closed, write.go:595-643) |
| `lint_rules[]` (`template.go:39-45`) | CEL over time/clock drift; sweep-only | `akb lint` (degrade per page, lint/cel.go:76-95) |
| `schema.frontmatter` (`template.go:19-29`) | **Retire.** `type`/`enum` are dead today (only `.Required` is read: required_fields.go:21, lint/required_fields.go:83, templates_write.go:386). Its remaining job — presence — is `required` inside the JSON Schema block | — |

Migration notes:

- **Presence is load-bearing.** The write path refuses a page missing a `schema`-required field
  *before* any rule runs (write.go:513-518) so CEL may read required keys unguarded (the guard
  doctrine in `adr.yaml:1-7`). The JSON Schema block must therefore run at the same point and
  supply the same information. Cleanest: keep one accessor (`checkRequiredFields`,
  required_fields.go:19-33) but source the set from the JSON Schema's `properties.frontmatter.required`;
  the sweep mirror (lint/required_fields.go:81-95) must follow, and `templates_write.go:386-395`
  (`optionalKeysSuppliedByMockup`) too. Write the invariant down: **the required set in the JSON
  Schema is exactly the set of keys CEL may read unguarded.**
- **CEL field-selector caveat:** a frontmatter key with `-`/unicode is legal in JSON Schema
  `properties` but not as a bare CEL selector (`page.frontmatter.foo-bar`). Coexistence does not
  fix that; it just moves it. Templates should keep keys CEL-identifier-safe.
- **Ordering and error aggregation:** JSON Schema validation errors + CEL rule errors must be
  merged into one report and one exit-1 outcome (current: all failed rules printed to stderr, then
  `validationFailure{}` → exit 1 via main.go:86-90 `classifyExit`). A `--json` mode needs that
  merged report serialized to stdout with stable rule IDs (JSON Schema keyword + instance pointer,
  CEL rule `ID`).
- **`akb template write` mockup obligations** (`templates_write.go:74-227`) grow by one test:
  - The pass mockup must satisfy the JSON Schema block **as well as** every CEL rule. Today it is
    only checked for required presence (templates_write.go:143-150) and CEL pass/optional-key/stale-old_page
    variants (155-188). A satisfiability check on the materialized mockup is the cheapest way to
    catch has-no-solution schemas (e.g. `content.word_count.minimum` above the mockup's actual word
    count — note rule #12 means the adr showcase's own mockup must clear 50 words).
  - The fail mockup must fail **the intended validator**. Today `templates_write.go:219-227`
    demands "at least one validation fails" over `tmpl.Validations` only. If the schema block is
    added, a mockup whose only defect is schema-level (or whose only defect is CEL-level) must be
    attributable; otherwise the `_fail.md` proof silently passes for the wrong reason. The existing
    machine-readable comment convention (`<!-- FAILS: valid_status — ... -->`, `adr_fail.md:19`)
    should be extended to `<!-- FAILS: schema: /frontmatter/status -->`.
  - The "no-op update against itself" old_page check (templates_write.go:179-188) is CEL-only and
    stays that way — JSON Schema has no `old_page`.
- **Surfacing:** `template get <name>` emits a Writer View of `{name, description, schema,
  requirements[]}` (template.go:163-190), YAML-marshalled; the `schema` value is `tmpl.Schema`
  (template.go:19-21), so a new block appears automatically once it is a field on `Template`.
  `--full` (template.go:154-159) dumps the YAML, so no extra work there.
- **Template file parsing is strict-ish:** `detectOldFormat` (template.go:73-82) already rejects
  `required`/`optional`/`body`; a new required block should get the same treatment (reject an
  unknown top-level block, or at least fail when both `schema.frontmatter` and `schema.json`
  declare `required` for the same key).

## 7. ARCHITECTURE — how the pieces connect (data flow)

```text
file bytes ──frontmatter.Parse(frontmatter.go:29)──> {Type,Title,Fields} + body
                                  │
        goldmark parse ───────────┴──> cel.BuildPage(pagebuilder.go:51) ──> page map
                                  │                                          │
        on-disk re-read ──> cel.BuildOldPage(pagebuilder.go:91) ──> old_page ─┤
        clock ─────────────────────────────────────────────────> now ───────┤
                                                                           ▼
                        [required gate] → [NEW: JSON Schema] → [CEL validations] → write
                                                                           │
                                          lint sweep uses the SAME BuildPage (lint/cel.go:62)
                                          with NO old_page and a shared `now` (lint/cel.go:74)
```

Four call sites build the page map: write.go:510, append.go:207, lint/cel.go:62,
templates_write.go:344. Four call sites read `schema.required`: required_fields.go:21,
lint/required_fields.go:83, templates_write.go:386, plus the writer view. Any new JSON layer must
plug into these five seams or it will diverge between write and sweep — which is exactly the
divergence the fail-closed/degrade split (write.go:588-594) exists to manage.

## 8. RISKS / OPEN QUESTIONS

1. `internal/cel/types.go` is dead code; the JSON view is the natural place for it to become the
   real contract (typed `Heading`/`Link`/`CodeBlock` instead of anonymous maps).
2. Duplicated exclusion logic (pagebuilder.go:117-120, 345-524 vs internal/markdown) becomes a
   published contract the moment derived views are emitted.
3. `akb.annotations[].position` means two different things on two paths (lint/cel bodies, see §3).
4. `char_count` = bytes vs `word_count` = whitespace fields: both need documented definitions
   before anyone writes a JSON Schema against them.
5. `file.dir` OS separator; `line` numbering convention (`source` parameter is passed
   inconsistently in tests, pagebuilder_test.go:257).
6. Link `resolved`/backlinks are DB-only; adding them to `read --json` changes read's failure modes.
7. New dependency: **no JSON Schema library exists in `go.mod`** today. Choice of library fixes the
   dialect (2020-12 vs draft-07) and the regex engine (ECMA vs RE2) — pick one and pin it in the
   template format docs.
8. A JSON Schema block makes validation possible on pages that never passed `akb write` (git pull,
   index rebuild ingestion) — same as lint today; the sweep must run schema checks per page and
   degrade like CEL does (lint/cel.go:84-95), not abort the sweep.

## SEAM VERDICT

**Rule for the line:** a rule is JSON Schema's iff its verdict is a pure function of **one
document at one instant**. Everything that needs a second input — `old_page` or `now` — is CEL's.
That is not an arbitrary split; it is visible in the code, because those two are the only non-page
CEL variables (`engine.go:31-35`).

- **JSON Schema owns:** frontmatter structure and value constraints (type, enum, pattern, length,
  list size/items, required, `if/then/else` on frontmatter values) — 9 of 26 shipped rules — plus
  every constraint over the **materialized derived views** (headings-present/absent, code-block
  language, wikilink existence, word-count floor) — 8 more.
- **CEL owns:** `old_page` (immutability #15, state machine #14), `now` (all 4 lint rules, #17),
  sibling-value comparison (#16), frontmatter↔derived cross-reference (#21), and any future
  arithmetic or disjunction-heavy predicate.
- **Where the owner's "combined" constraint forces derived data into the checked document:**
  rules #7-13 and #20 cannot be moved into frontmatter without inventing metadata that duplicates
  the body — exactly the rejected design. So the JSON document that akb validates **must** contain
  `content.word_count`, `ast.headings/links/code_blocks` (and `akb.*` for future rules), computed
  by akb from `content.raw`. Consequences to accept explicitly:
  1. On **write**, `content.raw` is the only content input; derived keys are akb-authored,
     output-only, and supplying them is an error.
  2. On **read**, the same builder must produce them; the raw frontmatter values, not the
     CEL-coerced `time.Time` view, are what cross the boundary (`page` ≠ `document`).
  3. JSON Schema-authored constraints may address derived paths but never `content.raw` text —
     i.e. the schema is checked against **akb's materialized document**, never against the caller's
     bytes.
  4. The presence gate (`required`) and the JSON Schema check run before CEL, preserving the guard
     doctrine that lets CEL read required keys unguarded.

## Start Here

Open `internal/cel/pagebuilder.go:51-86` (`BuildPage`). It is the single definition of what a page
*is*, and every design decision above follows from the fact that four of its five top-level keys are
derived from the body rather than supplied by the caller.

### Files retrieved

- `internal/cel/pagebuilder.go:29-116, 143-343, 345-524` — page map, old_page, derived views
- `internal/cel/types.go:6-24` — dead typed AST structs
- `internal/cel/engine.go:25-107` — env (`page`/`old_page`/`now`), compile cache, cost limit, panic recovery
- `internal/frontmatter/frontmatter.go:22-166` — parse, body extraction, type/title routing, IsDraft
- `internal/template/template.go:19-56` — Schema/FieldSchema/ValidationRule/LintRule/Template
- `internal/template/embedded/adr.yaml:1-201`, `note.yaml:1-34` — the 26 shipped rules
- `cmd/akb/write.go:86-561, 595-653` — write pipeline, re-serialization paths, validations
- `cmd/akb/append.go:68-260` — append pipeline (same gates, same re-serialization)
- `cmd/akb/read.go:34-66` — verbatim read, filesystem provider, no DB
- `cmd/akb/required_fields.go:19-39`, `cmd/akb/templates_write.go:74-227, 380-411` — presence gate + mockup validation
- `cmd/akb/main.go:17-107` — error classes and exit-code mapping
- `internal/lint/cel.go:41-100`, `internal/lint/required_fields.go:44-108`, `internal/lint/engine.go:45-67`
- `internal/markdown/wikilink.go:13-161`, `annotation.go:10-103`, `provenance.go:9-84`
