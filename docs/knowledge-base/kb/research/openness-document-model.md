---
as_of: "2026-09-27"
created: "2026-09-27"
grammar: 1
informs:
- RFC-001
- RFC-002
is_draft: false
provenance: agent-drafted
run: json-cel
scope:
- internal/cel/**
- internal/markdown/**
- internal/frontmatter/**
status: final
summary: "Empirically probe-verified audit of the document model: frontmatter/file-identity mandates, derived page-map projections, and markdown parsing drift across goldmark call sites."
tags:
- openness
- document-model
- markdown
- audit
title: Openness audit — document model & markdown parsing architecture
type: research
updated: "2026-10-09T16:36:23Z"
---
# OPENNESS AUDIT — Seam: Document Model & Markdown Parsing Architecture

**Scout:** openness-audit / document-model seam (read-only recon; no repo file modified — verified `git status` shows only a pre-existing `.pi/exec-profile.md` modification).
**Method:** read `internal/cel/{pagebuilder,engine,types}.go`, `internal/markdown/*`, `internal/frontmatter`, all `goldmark.New()` call sites, `go.mod`; **empirically verified** every drift claim by running probes in a throwaway copy of the tree (`/tmp/akbcopy`, `internal/probe/*_test.go`, deleted-on-exit; the repo itself was never written). Probe outputs are quoted verbatim below.
**Not re-derived (cited instead):** page-map key inventory and derived-view semantics — prior research: `.pi/subagents/proposals/json-cel/research/seam-core-pipeline.md:15-56` (table), `:43-48` (char_count bytes, is_wikilink heuristic), `:139,330` (position semantics split), `command-inventory.md:62-65`, `akb-recon.md:103-112`, `seam-periphery.md:20,27,59,115,150`.

---

## Findings

### F1. Document-level mandates (frontmatter / file identity)

| # | WHAT it forces | anchor | WHO consumes it | Could it be optional/configurable? | verdict |
|---|---|---|---|---|---|
| M1 | Page **must** carry YAML frontmatter delimited by a line starting `---` at line 1 and a closing `---`; body is what follows | `internal/frontmatter/frontmatter.go:89-126` (`extractBody`); error `no frontmatter found` at `:47`, `:52` | every read path: `write.go:189,312,359`, `append.go:149-158`, `approve.go:139-142`, `cmd/akb/lint.go:120-124` (lint tolerates: `hasFrontmatter=false`, `body=content`) | Yes — a frontmatter-less page is already legal to *lint*; only write/append/approve refuse. Making `type`/`title` schema properties (not engine keys) is the owner's thesis. | **HISTORICAL ACCIDENT** |
| M2 | `type` must be a non-empty **string** and must name an existing template | `frontmatter.go:57-60` (string coercion), `:130-138` (`ValidateType`); callers `write.go:194,235,318,365`, `append.go:162` | `templates[fm.Type]` lookup (`write.go:502`), `search.IndexPageTx` `type` column (`internal/search/sqlite.go:56-78`), index grouping, `type_orphan` lint, `template delete` impact count | Yes in principle: a JSON-Schema/CEL KB needs only "the schema that applies". Today `type` is the *selector* for both schema and directory. | **HISTORICAL ACCIDENT** (pre-CEL type system) |
| M3 | `title` must be a non-empty **string** | `frontmatter.go:61-64`, `:141-147`; callers `write.go:197,323,370`, `append.go:165` | `search` title column, `index.md` headings, `akb list` | Yes — `adr.yaml:13-17` already declares `title` required; the engine check duplicates the template check. | **HISTORICAL ACCIDENT** |
| M4 | `type`/`title` are **reserved CEL keys**: `page.frontmatter.type`/`.title` are re-injected from the parsed struct | `pagebuilder.go:64-66` | CEL rules (`adr.yaml:13-17`, skill `TEMPLATE.md:83`) | Yes (it's convenience), but removing it breaks every shipped rule that reads `page.frontmatter.type`. | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** |
| M5 | `is_draft` lifecycle key with lossy coercion: absent ⇒ draft; **only** literal `false` (or string `"false"`) approves | `frontmatter.go:150-168` (`IsDraft`) | `approve.go:144`, `write.go:382-386` (deletes `is_draft: true` on write!), `list.go:24` (`is_draft` JSON field) | State is opinionated-and-load-bearing; the *coercion* (string `"false"`, any other type ⇒ draft) is arbitrary. | state: **LOAD-BEARING**; coercion: **HISTORICAL ACCIDENT** |
| M6 | akb **injects/rewrites** `created` and `updated` (RFC3339) and re-serializes the whole frontmatter | `write.go:377-379` (`created` if absent), `:231`, `:335`, `:478-481` (`updated`), `:485-495` (goccy re-marshal), `append.go:176` | search `updated`? (no — `documents.created` is `datetime('now')`, see F6), git diffs, CEL `timestamp(page.frontmatter.updated)` (`adr.yaml:143-149`, `note.yaml:34`) | Could be per-template declared (e.g. `auto_fields:`), or off. **Blast radius: every new page is byte-re-serialized** (see F6/§input=output). | **GENUINELY OPINIONATED-BUT-LOAD-BEARING** |
| M7 | Template `dir:` prescribes the page's directory; write strips/re-adds the prefix | `template.go:47`; `write.go:396,423-431,433-447` | path resolution, `index.md` rendering, `type_orphan` lint | Yes — a path is already in the page; the type→dir binding is a filing convention. | **GENUINELY OPINIONATED** (weakly load-bearing) |

### F2. Derived page-model projections (`page.*` keys) — shape mandates

Full key inventory is prior work (`seam-core-pipeline.md:15-56`; `akb-recon.md:103-112`). Gaps and prescription details this scout verified:

| # | Mandate | anchor | Consumer | Openable? | verdict |
|---|---|---|---|---|---|
| M8 | `page.akb.*.position` is a **byte offset**, while `page.ast.*.line` is a **line number** — two unit systems inside one document model | `pagebuilder.go:363,392` vs `:156,259,272,294,338` | CEL rules (offset is nearly unusable in CEL: no substring/slice op in core; only ordering/equality) | Yes — unify on `line` (or drop) with no rule rewrite, since shipped templates never read `position` | **HISTORICAL ACCIDENT** |
| M9 | `content.char_count` = **bytes**; `content.word_count` = whitespace-split fields | `pagebuilder.go:70,71` | CEL rules (`seam-core-pipeline.md:31-32,43`) | Yes — but any change silently shifts existing `maxLength`/`minimum` thresholds | **HISTORICAL ACCIDENT** (mislabeled; byte-vs-rune undecided) |
| M10 | `ast.code_blocks[]` counts **fenced blocks only** (`*ast.FencedCodeBlock`) | `pagebuilder.go:322-345` (type-switch at `:328`) | `code_blocks_have_language` rule (`seam-core-pipeline.md:87`) | Yes — add `*ast.CodeBlock`; today indented code is invisible (probe A: `code_blocks: nil`) | **HISTORICAL ACCIDENT** |
| M11 | `ast.links[].is_wikilink` is decided by **sniffing the source** (`source[pos]=='[' && source[pos+1]=='['`), not by the parse | `pagebuilder.go:281` | `require_wikilink` rule (`seam-core-pipeline.md:94`) | Yes — after an inline-parser rewrite it becomes a parse fact | **HISTORICAL ACCIDENT** |
| M12 | `ast.*` exposes only headings / links / code_blocks; blockquotes, lists, tables, emphasis, images, autolinks, raw HTML are **not projected** | `pagebuilder.go:75-79`; `types.go:7-24` (structs with **zero consumers** — verified by grep) | nothing today | Purely additive: CEL env declares `page` as `map(string,dyn)` (`engine.go:36-40`), so **new keys need no env change** | **OPINIONATED** (projection surface frozen early) |

### F3. Markdown dialect mandates (parse-level)

| # | Mandate | anchor | Notes |
|---|---|---|---|
| M13 | Dialect = **CommonMark core only**. Every parse site is bare `goldmark.New()` with **no extensions**: `pagebuilder.go:112`, `write.go:508`, `append.go:205`, `templates_write.go:342`, `lint/cel.go:60`; only `frontmatter.go:36-43` adds an extension (`frontmatter.Extender`) | grep: no `goldmark.WithExtensions` except `frontmatter.go:37` | goldmark v1.8.2 (`go.mod:14`) offers `extension.Table|Strikethrough|Linkify|TaskList|Footnote|DefinitionList|Typographer|CJK` (`extension/package.go`) — all OFF. Probe: `| a | b |` parses as a **Paragraph** (no `Table` node), `~~x~~`/task lists likewise. So tables are plain text — neither "supported" nor rejected; nothing tells the user. |
| M14 | Wikilinks are a hand-rolled grammar grafted onto CommonMark: `[[target]]`, `[[target\|display]]`, `[[target#heading]]`, `[[display]](dest)` with destination-normalization (`./` and `.md` stripped) | `wikilink.go:37-153`, `:350-409` (`NormalizeBareDestination`/`normalizeDest`), tests `wikilink_test.go:282-961` | The grammar (degenerate bracket runs, suppression inside titles/paths) is genuinely akb-specific: **keep**. |
| M15 | All *structure* detection around that grammar is **re-implemented by hand** rather than read from the goldmark AST the same process already built | `wikilink.go:410-935` + `pagebuilder.go:415-524` | See F4. **HISTORICAL ACCIDENT** — akb parses with goldmark *and* re-derives CommonMark block structure byte-by-byte. |

### F4. THE DRIFT — verified, with exact divergent classes

Owner's claim verified and made precise. `internal/markdown/wikilink.go` is **936 lines**; the classification below counts **only function bodies** (doc comments are heavy in this file, so the owner's ~658 figure sits between the two bounds I measured):

| category | functions (anchors) | lines incl. doc comments | non-comment lines |
|---|---|---|---|
| **A. block-structure/exclusion scanners** (goldmark already computes these) | `computeExclusions` 428-451, `indentedCodeBlockRanges` 452-582, `lineOverlapsRanges` 583-591, `leadingWhitespaceColumn` 592-609, `scanListMarker` 620-666, `isThematicBreak` 667-683, `startsParagraphInterrupt` 684-770, `hasItemContent` 771-792, `escapedBracketRanges` 793-806, `isEscapedByte` 807-816, `fencedCodeBlockRanges` 817-840, `fencedCodeBlockInteriors` 841-867, `inlineCodeRanges` 868-892, `countBackticks` 893-900, `findClosingBackticks` 901-912, `htmlCommentRanges` 913-935, + `exclusion`/`exclusionSet`/`isExcluded` 410-427 | 547 | 366 |
| **B. CommonMark inline-link tail** (goldmark's inline lexer does this) | `scanParenDestination` 163-214, `isValidLinkDestination` 215-224, `parseLinkDestination` 225-251, `scanDestination` 252-285, `scanAngleDestination` 286-304, `scanTitle` 305-325, `skipSpaces` 326-333, `isSpace` 334-339, `isPunct` 340-349 | 196 | 138 |
| **C. akb-specific wikilink grammar** (keep) | `scanWikilinkTokens` 44-80, `countRun` 81-88, `ParseWikilinks` 90-153, `NormalizeBareDestination` 350-364, `normalizeDest` 365-374, `parseWikilinkInner` 379-409 | 193 | 132 |

**A+B = 743/936 (79%) incl. doc comments, 504/936 (54%) excluding them. The owner's ~658/935 is inside the range; the point stands either way.**

**The second copy, driftier.** `internal/cel/pagebuilder.go` (524 lines) carries its own exclusion machinery, self-admittedly: comment at `:118-120` ("The exclusion helpers below duplicate internal/markdown's unexported equivalents"). Exact ranges: `exclusion`/`exclusionSet`/`isExcluded` `:415-431`; `computeExclusions` `:433-441`; `fencedCodeBlockRanges` `:441-462`; `inlineCodeRanges` `:463-483`; `countBackticks` `:484-491`; `findClosingBackticks` `:492-503`; `htmlCommentRanges` `:504-524`. Also duplicated parser bodies: `parseProvenanceMarkers` `:347-367` ≈ `provenance.go:22-47`, `parseAnnotations` `:371-395` ≈ `annotation.go:24-70`, `parseAnnotationFields` `:398-413` ≈ `annotation.go:47-69` (verified near-identical after stripping comments; `parseAnnotationFields` differs only in comment text — 23 vs 16 non-comment lines).

**Exact divergent classes** (machine-diffed the shared function bodies; only `computeExclusions` differs):

| exclusion class | `markdown/wikilink.go:428-437` | `cel/pagebuilder.go:433-441` | body identical? |
|---|---|---|---|
| fenced code, backtick | ✔ `:817` | ✔ `:441` | **yes** (comment-stripped) |
| **indented code** (CommonMark, tab-aware, list-content-column-aware) | ✔ `:452-582` | **✘ MISSING** | — |
| inline code | ✔ `:868` | ✔ `:463` | **yes** |
| HTML comment | ✔ `:913` | ✔ `:504` | **yes** |
| **escaped brackets** | ✔ `:793-806` | **✘ MISSING** | — |
| tilde fence `~~~` | **✘ missing** | **✘ missing** | shared blind spot |

**Consumers of the shared 5-class set:** `provenance.go:27` (`ParseProvenanceMarkers`), `provenance.go:67` (`StripProvenanceMarkers`).
**Consumers of the 2-class subset (fence + inline only):** `annotation.go:32` (`ParseAnnotations`), `annotation.go:84` (`StripAnnotations`), `pagebuilder.go:378-379` (`parseAnnotations`).

**Live inconsistencies (all probed; probe output verbatim):**

1. **`^[inferred]` in a 4-space indented code block** — the divergence the owner predicted, confirmed:
   ```text
   markdown.ParseProvenanceMarkers -> []
   CEL page.akb                    -> {"provenance_markers":[{position:15,"type":"inferred"}], ...}
   ```
   Same page, two answers. Practical effect: the `provenance` lint checker counts markers via `cmd/akb/lint.go:129` → `internal/lint/provenance.go:51`, while a CEL lint rule reading `page.akb.provenance_markers` counts the same page differently → drift warnings (`provenance.go:69-77`) disagree with CEL rules on one page.
2. **Tab-indented code** — same split (`\t^[inferred]`): `[]` vs `[{inferred,7}]` (goldmark: `CodeBlock(7..7)`).
3. **Escaped bracket** — `\[[^[inferred]]]`: markdown `[]` vs CEL `[{inferred,3}]` (the escaped class only bites when the marker lies inside the token's inner span; `a \[[x]] b` alone shows no split).
4. **Tilde fence** — both parsers report `^[inferred]` inside `~~~...~~~`, while goldmark yields `FencedCodeBlock(4..16)`. Both wrong, jointly.
5. **Annotation inside an indented code block** — **both** parsers report it (`markdown.ParseAnnotations -> [{olw-auto ... Position:83}]`, CEL agrees) although goldmark renders that line as `CodeBlock` — a shared false positive, not a drift.
6. **Annotation inside the YAML frontmatter** — lint path passes the *full content* (`cmd/akb/lint.go:130`), CEL path the *body* (`pagebuilder.go:83`):
   ```text
   lint path  ParseAnnotations(full content) -> [{in_frontmatter:true pos=38} {in_body:true pos=114}]
   CEL  path  parseAnnotations(body)        -> [{in_body:true pos=25}]
   ```
   Two consequences: a marker in frontmatter is an annotation on one path only; identical annotations get **different positions** (114 vs 25). Prior research flagged the position-semantics split (`seam-core-pipeline.md:139,330`); the frontmatter case is new. Blast radius is currently **latent**: `lint.PageData.Annotations` (`internal/lint/engine.go:66`) is written at `cmd/akb/lint.go:139` and **never read** by any checker (verified by grep) — dead data.

### F5. Prescribed vocabularies

| # | Vocabulary | anchors | Consumers | If template/config-defined or optional | verdict |
|---|---|---|---|---|---|
| M16 | `^[type]` provenance markers, **open** on parse (`[^\]]+?`, any type) | `provenance.go:18,22-47`; duplicate `pagebuilder.go:345,347-367` | `lint/provenance.go:51` (via `CountMarkersByType`, `provenance.go:49-55`), CEL `akb.provenance_markers`, counts vs frontmatter `provenance` map (`provenance.go:36`) | Parse side is already open — good. |
| M17 | **Strip** vocabulary is **closed to 3 types**: `(inferred\|ambiguous\|extracted)` | `provenance.go:57` (`stripProvenanceRe`), used `:62-84` | `cmd/akb/approve.go:153` | A template-declared 4th marker type is counted by CEL and drift lint, then **survives `akb approve` forever** — a real asymmetry between "open parse" and "closed strip". | **HISTORICAL ACCIDENT** (should derive from config/template, or strip any `^[...]`) |
| M18 | `<!-- olw-auto: k=v k=v -->` annotation prefix, hardcoded **twice** per copy (regex + literal `Type`) | `annotation.go:20` & `:45`; duplicate `pagebuilder.go:369` & `:390`; strip `annotation.go:87`; approve `approve.go:152` | CEL `akb.annotations`; `approve` strips; lint parses but ignores | The prefix is a residue of a prior tool ("olw"); `Type` is *not* derived from the match, so a second namespace (e.g. `akb-auto:`) is impossible without code change. Field grammar is whitespace-tokenized `key=value` (`annotation.go:47-69`) → **values can never contain spaces**. | **HISTORICAL ACCIDENT** |
| M19 | Hardcoded frontmatter key names: `provenance` (drift checker), `sources` (citations checker), `is_draft`, `created`/`updated` | `lint/provenance.go:36`, `lint/citations.go:37`, `frontmatter.go:155`, `write.go:377,478`; thresholds `lint/thresholds.go:11-16` (drift 0.20) | lint checkers, CEL rules, search | Yes — vocabulary could be template/config-declared; today these three checkers fire on *any* type's pages regardless of template | **HISTORICAL ACCIDENT**-ish / opinionated for `sources` |

### F6. Date coercion

`convertDateField` (`pagebuilder.go:29-45`, applied `:60-64`) rewrites **every top-level string** that parses as RFC3339 **or** `2006-01-02` into `time.Time`, *regardless of name and without opt-out*.

- Verified (`probe4`): goccy decodes `created: 2024-01-01` as **`string`** — so the `time.Time` CEL sees comes **only** from this function (matches `seam-core-pipeline.md:215`).
- Verified asymmetry: nested values are **not** coerced (`nested: [2024-01-01]` ⇒ `string`, `map.when` ⇒ `string`), so `timestamp(page.frontmatter.nested[0])` fails while `timestamp(page.frontmatter.created)` succeeds (also `akb-recon.md:111`).
- **Consumers that depend on coercion**: shipped templates — `adr.yaml:143,149,183,190,197` (`timestamp(...)`, duration math), `note.yaml:34`; documented in `skill/.../TEMPLATE.md:83`, `internal/cel/AGENTS.md:70`; test `pagebuilder_test.go:236-320` (non-`created`/`updated` field names `valid_from`, `starts_at` are coerced too — i.e. the coercion is *deliberately* name-agnostic).
- **Who relies on date-only** `2006-01-02`: the shipped mockups (`adr_pass.md:27` "created (2026-01-05)"), since `write.go:377-379` injects RFC3339 but hand-written pages use date-only.
- Every **other** date site is unrelated string formatting and does **not** share this parser: `log.go:93` (`2006-01-02` for log headings), `manifest.go:141,177`, `raw_sync.go:111,122`, `init.go:68`, `append.go:57,176`, `write.go:231,335,378,480`. Search's `--after` is documented as "filter by creation date (YYYY-MM-DD)" (`search.go:41`) but compares against `documents.created`, which is `datetime('now')` **at index time** (`search/sqlite.go:56-62,176-179`) — the frontmatter `created` **never reaches search**.
- Three candidate regimes and their consumer impact:
  1. **Coerce silently (today):** `version: 2026-01-01` becomes a timestamp — a *string* field silently changes CEL type; string comparison on such a field is impossible/broken; nested date strings behave differently from top-level (no rule can rely on uniformity).
  2. **Reject non-RFC3339 at write:** breaks every date-only page and mockup above (`adr_pass.md:27`), and would refuse a human-typed `created: 2026-01-05`; also write-time-only, so `lint` over pre-existing pages still sees whatever the coercion does.
  3. **Template declares date fields (`format: date`):** coercion becomes schema-driven and name-aware — CEL's `timestamp()` then needs the same knowledge or the value must already be a timestamp; this is exactly the JSON-Schema seam (`seam-core-pipeline.md:256`, `format`/`oneOf` for dates). Cost: the schema must exist for a page to be coerced, so lint-time handling of untyped pages needs a fallback.

---

## Consumer trace

```text
body (post-frontmatter) ──┬─→ frontmatter.Parse (goldmark+frontmatter ext)  [write.go:189,312,359; append.go:149; lint.go:120; approve.go:139]
                          └─→ goldmark.New().Parser().Parse(body)            [write.go:508; append.go:205; templates_write.go:342; lint/cel.go:60; pagebuilder.go:112]
                                  └─ cel.BuildPage ─→ page map ─→ CEL rules (validations write-time; lint_rules sweep-time)
markdown pkg (7 scanners) ─┬─→ linkgraph.UpdatePageLinksTx   [linkgraph/sqlite.go:78]  ← uses ONLY Target/Display
                          ├─→ page.ast.links (merge, offsets) [pagebuilder.go:205]     ← uses Start/Destination/Display
                          ├─→ lint provenance checker        [lint.go:129 → lint/provenance.go:51]
                          └─→ approve strip                  [approve.go:152-153]
page.akb.provenance_markers/annotations (cel-local copies) ─→ CEL rules ONLY
lint.PageData.Annotations ─→ NO CONSUMER (dead field, engine.go:66)
```

Precise consumer facts that bound the blast radius:

- **Link graph needs almost nothing from the parser**: `linkgraph/sqlite.go:111` uses `wl.Target` + `wl.Display`; `dedupeWikilinksByTarget` dedupes by target. `Start`/`End`/`Destination`/`HasHeading` are consumed **only** by `pagebuilder.go:233,239,250,255,259,268,272` (offset merge + `line` numbers). So ~659 lines of block-structure scanning exist to serve **one** function's offset merge and one lint marker count.
- **Date coercion has exactly one consumer**: the CEL page map (`pagebuilder.go:62`); nothing else reads `convertDateField`.
- **`page.ast.code_blocks`** feeds only CEL rules (`seam-core-pipeline.md:87`).
- **`cel/types.go:7-24` (`Heading`/`Link`/`CodeBlock`) has zero references** anywhere in the tree (grep). The map shapes are defined twice (structs + map literals) and consumed zero times — the CEL boundary is `map[string]any` (`engine.go:36-40`).
- **Annotations on the lint path are dead** (§F4.6), so fixing the position-semantics split is cheap *today* but will bite the moment a JSON/derived view exposes annotations on both paths.

---

## Blast-radius / verdicts

**STRUCTURALLY REQUIRED (keep):**
1. `page.content.raw` + `content.word_count/char_count` — shipped rules depend on them; a body must exist (`seam-core-pipeline.md:86-87`).
2. `page.frontmatter.*` as raw (coerced) input, `old_page` from disk, `now` injection.
3. The **wikilink grammar itself** (`[[target|display]]#heading` + `(dest)` normalization) — it is the one thing embedding akb's "agent-first KB" semantics, and the link graph's target spelling must stay aligned with it (`pagebuilder_test.go:696-740`).
4. `is_draft` state semantics (draft → approve), `sources`↔raw-manifest citations link, provenance-ratio drift *concept* (`thresholds.go:11-16` constant is not).

**HISTORICAL ACCIDENTS (openable / removable):**
- `type`/`title` as engine-level keys with engine-level errors (M2/M3) — the single biggest openness lever in this seam.
- The whole hand-rolled CommonMark structure layer: **743 of 936 lines** in `wikilink.go` (A+B) + **110 lines** of drifted duplicate scanners in `pagebuilder.go` (`:415-524`) + 3 duplicated parser bodies; two live inconsistencies (`indented`, `escaped`) plus one shared blind spot (tilde fence).
- The **3-class vs 5-class vs 2-class** inconsistency: `akb.provenance_markers` (CEL) ≠ `provenance` lint marker count on indented-code/escaped-bracket content.
- Closed strip vocabulary (`stripProvenanceRe`) vs open parse vocabulary.
- `char_count` = bytes; `akb.*.position` = byte offset vs `ast.*.line`; `ast.code_blocks` fenced-only; `is_wikilink` source-sniffed; `cel/types.go` dead structs.
- Silent name-agnostic date coercion of *any* top-level string (and its top-level/nested asymmetry).
- Per-page `goldmark.New()` in 5 places, + a second full parse in `frontmatter.Parse` per write ⇒ ≥2 goldmark parses/page/write and one more for `old_page` (`pagebuilder.go:112`) — a parse-cost artifact, not a semantic mandate.

**GENUINELY OPINIONATED-BUT-LOAD-BEARING (keep, make configurable at most):**
- `is_draft`/approve lifecycle; `created`/`updated` auto-fields (they are what makes "input = output" false on every new page, M6); template `dir:` filing; CommonMark-only dialect (no GFM) — this is a defensible stance for a KB, but it should be *documented* since tables silently degrade to paragraphs.

**input = output precision (measured, not asserted):** on the stdin path `writeContent = stdinContent` **verbatim** (`write.go:376`, `:489-495` re-serializes only when `needsReserialize`). For a **new page** `created` is always missing (`write.go:377-379`), so `needsReserialize` is always true ⇒ the frontmatter is re-marshalled through goccy YAML (key order/formatting/comments not preserved). Only an overwrite of a page that already carries `created` **and** `updated` and no `is_draft: true` round-trips byte-identically. `--append` additionally injects `"\n" + appended` (`write.go:342`). **The body itself is never mutated** — the body round-trip is already clean.

---

## Feasibility sketches

**(a) Goldmark inline-parser rewrite — the owner's ~120-150-line target.** Register a `parser.InlineParser` for trigger `[` that consumes `[[...]]` (+ optional adjacent `(dest)`), emitting nodes (or a custom `ast.Node` with `Destination`/`Display`/`Heading`). Feasible inputs, measured: goldmark v1.8.2 exposes `parser.WithInlineParsers`, `parser.WithASTTransformers`, `ast.KindText`/`Segment` (source offsets), and `ast.Text.Segment` covers escapes verbatim (`a \[[x]] b` ⇒ `Text seg=[0,5) "a \[["`), so pos/line for `ast.links[].line` and for `Wikilink.Start/End` remain derivable. Deleted as a consequence: **all** of category A (547 lines incl. `indentedCodeBlockRanges` 132/`startsParagraphInterrupt` 85/`scanListMarker` 51/`isThematicBreak` 22/`hasItemContent` 23/`lineOverlapsRanges`+`leadingWhitespaceColumn`/fence+backtick+comment scanners), all of category B's role in `scanParenDestination`'s exclusion duty, the cel copies (`pagebuilder.go:415-524`), the merge in `flattenLinks` (`pagebuilder.go:177-315`, incl. the `containsAnother` malformed-case fixups `:219-247`), and `is_wikilink` sniffing. Residual cost: ~40-60 lines of token grammar (category C) + API change so `linkgraph/sqlite.go:78` and `pagebuilder.go:205` take the parsed node list instead of re-parsing strings. Risk: `wikilink_test.go:282-961` encodes close to 100 cases of the *current* degenerate-bracket/suppression semantics; a rewrite must re-home them (they are the real spec, not the scanners).
*(Also worth a look, unverified here: `go.abhg.dev/goldmark/wikilink@v0.6.0` is present in the local module cache — a third-party goldmark wikilink extension exists; whether its grammar matches akb's `[[display]](dest)`/degenerate runs was not tested.)*

**(a′)** **Lower-risk variant — AST-derived exclusions.** Keep the string scanners but build the exclusion set from the already-parsed `ast.Node` the callers already have (BuildPage receives `astDoc`: `pagebuilder.go:51`). Ranges are all available, measured:
`CodeBlock.Lines()` `[10,31)` (indented), `FencedCodeBlock.Lines()` `[4,22)` (backtick **and** tilde), `CodeSpan` child `Text` segments `[6,23)` (inline code), `HTMLBlock.Lines()` `[0,27)` / `RawHTML.Segments` `[5,32)` (comments), escape presence from `source[start-1]=='\\'`. That is ~40 lines replacing ~500 (A + cel copies), and it *fixes* the tilde-fence blind spot and the CEL/markdown split at once. Cost: `markdown.ParseWikilinks/ParseProvenanceMarkers/ParseAnnotations` currently take `string` only (`linkgraph/sqlite.go:78`, `cmd/akb/lint.go:129-130`) — they would take a range set or ask for one, and lint would have to parse bodies with goldmark (it already does at `lint/cel.go:60`).

**(b) Richer / template-defined AST projections.** Cheap by construction: the CEL env types `page` as `map(string, dyn)` (`engine.go:36-40`), so **any** new key (`ast.blockquotes`, `ast.tables`, `ast.emphasis`, `ast.images`, `ast.autolinks`, `ast.html_blocks`, `ast.list_items`) works in rules with **zero** env/type changes — only `BuildPage` (`pagebuilder.go:75-79`) and the JSON view need to grow. Tables/lists/blockquotes need `extension.Table` etc. enabled to have nodes at all; a *template-defined* projection list would additionally need a new `template.Template` field (`template.go:43-54`; the loader is **non-strict** `yaml.Unmarshal`, `template.go:106-114`, so an unknown `views:` key is currently ignored silently — decide strict vs. silent) plus docs in `skill/.../TEMPLATE.md:67-83`. Blast radius of *adding* keys: none for existing rules; the only sharp edge is `seam-core-pipeline.md:86-94`'s table — rules that encode `ast.*` shapes must keep working (they do: keys are additive).

**(c) Date regime migration sketch.** `format: date` in `FieldSchema` (`template.go:20-25`) → `convertDateField` becomes schema-aware (coerce only declared date fields), CEL rules unchanged for declared fields; undeclared date-only strings stay strings, so `adr.yaml`'s rules keep working because `created`/`updated` are declared. Cost: `BuildPage` would need the template (it currently takes only `fm`/`body`/`doc`; `write.go:510` has `tmpl` in scope, `lint/cel.go:62` has it too, `templates_write.go:344`/`append.go:207` as well — all call sites already hold the template, so the signature change is mechanical, not architectural).

---

## Open questions for synthesis

1. **Who owns "what is a valid marker"?** With JSON Schema + CEL, should `akb.provenance_markers` exist at all, or should the projection be *declared* (e.g. a template-level `projections:` list naming marker vocab, annotation namespace, and date fields)? That single declaration would retire M16-M19 and make F4's drift impossible by construction.
2. **Is `type` load-bearing as an engine key or can it become a schema property?** Consumers found: template selection, search `documents.type`, index grouping, `type_orphan`, `template delete` impact count (`command-inventory.md:40`, `seam-periphery.md:164-186`). If it becomes a property, who selects the schema? (Filename? `schema:` in frontmatter? A KB-wide schema?)
3. **Dialect declaration:** should akb document "CommonMark, no GFM" as doctrine, or expose `markdown: {tables: true}` in `akb.yaml`/template? Today a user's GFM tables are silently paragraphs — with no lint to tell them.
4. **Reorder write-time validation:** `BuildPage` (`write.go:510`) runs **before** required-field checks (`write.go:515`) and before the CEL rules; if projections become schema-declared, the declared-vs-parsed order matters (a schema-driven view needs the schema loaded earlier than today's `checkRequiredFields`).
5. **One parse or five?** Five `goldmark.New()` sites + the frontmatter extender parity (≥2 parses/page/write). If a single parse-with-extensions becomes the front door, does `frontmatter.Parse` keep its own `extractBody` (`frontmatter.go:89-126`), which re-derives the body split the extender already computed?
6. **`position` vs `line`:** unify before a JSON/derived view freezes the byte-offset semantics (and the frontmatter-vs-body offset split, F4.6) into a public contract.
7. **Should strip vocabulary follow parse vocabulary?** (`^[custom]` is counted but never stripped by `akb approve` — a silent data-retention bug for template-defined types.)
