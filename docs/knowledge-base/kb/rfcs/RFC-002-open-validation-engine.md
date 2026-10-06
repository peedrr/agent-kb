---
author: pete
created: "2026-09-27"
grammar: 1
id: RFC-002
provenance: agent-drafted
scope:
- "**"
spawns:
- RFC-002-A1
- RFC-002-A2
- RFC-002-A3
- RFC-002-A4
- RFC-002-A5
- RFC-002-A6
- RFC-002-A7
- RFC-002-A8
- RFC-002-A9
- RFC-002-A10
- RFC-002-A11
- RFC-002-A12
- RFC-002-A13
- RFC-002-A14
- RFC-002-A15
- RFC-002-A16
- RFC-002-A17
- RFC-002-A18
- RFC-002-S1
- RFC-002-S2
- RFC-002-S3
- RFC-002-S4
- RFC-002-S5
status: draft
steward: pete
summary: Make akb an unopinionated validation-and-persistence engine — JSON Schema 2020-12 and CEL as its only validators, byte-faithful writes, every other mandate removed, demoted, or made configurable.
tags:
- validation
- json-schema
- cel
- openness
title: The Open Validation Engine — JSON Schema + CEL as akb's only validators
type: rfc
updated: "2026-10-06T17:16:30Z"
---

# RFC-002: The Open Validation Engine — JSON Schema + CEL as akb's only validators

| | |
|---|---|
| **Status** | Draft — awaiting owner ratification |
| **Date** | 2026-09-27 |
| **Authors** | pete (owner), pi session + subagent research team |
| **Supersedes** | [[RFC-001-json-schema-cel-coexistence|RFC-001]] (`RFC-001-json-schema-cel-coexistence.md`) — retained as research record |
| **Decision register** | Appendix A (D1–D19 + D-seq + P-a…P-i, all owner-ratified 2026-09-26/27) |
| **Research base** | `.pi/subagents/proposals/json-cel/research/` — 13 reports incl. the openness audit (`openness-*.md`) and `SYNTHESIS-openness.md` |
| **Downstream artifacts** | ADRs + tech-specs spawned as RFC-local candidates per §18 (`RFC-002-A*`/`RFC-002-S*` — nothing here is implementation-final) |

---

## 1. Abstract

akb becomes an **unopinionated validation-and-persistence engine**. Every page-level
mandate akb enforces is exactly one of:

1. **Structural** — required for the tool to function safely (KB boundary, path
   containment, concurrency, cost bounds);
2. **Derived** — mechanically regenerable from page content (search index, link graph,
   `index.md`);
3. **User-declared** — expressed in the user's own templates (JSON Schema + CEL +
   template declarations) or KB config (`akb.yaml`).

Everything else — and the audit (`research/openness-*.md`) found ~20 such mandates —
is removed, demoted, or made configurable. Validation is two co-equal layers the user
authors: **JSON Schema 2020-12** over a materialized **document** view, and **CEL** for
everything schema cannot express (`old_page`, `now`, cross-value comparison, arithmetic).
JSON I/O (`read --json`, `write --json`, `--input-schema` export) makes pages pipeable
to and from agent structured output. **The write path is byte-faithful: input = output.**

RFC-001 proposed adding JSON Schema *alongside* akb's existing prescriptive core. The
openness audit demonstrated that the prescriptive core itself — managed-field injection,
whole-frontmatter re-marshalling, a fixed lint set, hand-rolled markdown scanners,
hardcoded vocabularies and constants — is largely historical accident. This RFC therefore
reframes: **schema + CEL do not augment the prescriptive core; they replace it.**

## 2. Motivation

RFC-001 §2 stands (agents speak JSON Schema natively; no translation design is sound;
coexistence is the ecosystem-proven pattern). The openness audit added three findings
that changed the trajectory:

1. **The input≠output gap is one function deep.** `created`/`updated` injection has
   *zero* internal consumers (search `--after` reads index time, never the frontmatter
   field); `is_draft:true` deletion is a semantic no-op. Remove those and the entire
   fidelity gap reduces to a serialization accident (map round-trip re-marshal).
2. **The lint mandate is one call site.** `NewLintEngine()` returns an empty engine; the
   10-checker set exists only at `cmd/akb/lint.go:157-167`. The real openness bound is
   that user rules can name only page-local facts.
3. **The markdown layer is the deepest accident.** 743/936 lines of `wikilink.go`
   re-implement CommonMark structure goldmark already computes; a second, *drifted*
   copy in `pagebuilder.go` produces probe-verified live inconsistencies (a marker in an
   indented code block is counted by CEL and ignored by lint, on the same page).

akb is agent-first and harness-agnostic: it provides the hands; users decide what a
valid KB entry is, how entries are compared, and how their KB is laid out.

## 3. Relationship to RFC-001

**Carries over unchanged** (Appendix B): the P1 seam rule, coexistence-over-translation
rationale and external research, JSON Schema 2020-12 + library-first coverage (G1),
cel-go upgrade + extension libraries (G2), the three-surface failure contract (write
fails closed, sweep degrades per page, `template write` proves mockups), JSON I/O
commands, the merged error model, `lint --json` envelope stability, and the phasing
skeleton.

**Superseded by this RFC:** the fixed `document` shape (→ flattened, projections
declared), the "modulo managed mutations" round-trip caveat (→ identity), "export-and-
share the exclusion scanners" (→ delete them), the prescribed `akb.*` vocabularies
(→ declared), `created`/`updated`/`is_draft` as tool behavior (→ template-declared),
approve as bundled gate (→ unbundled), `index.md` as managed registry (→ derived),
and the silent config surface (→ behavioral `akb.yaml` with strict keys).

## 4. Design principles

- **P1. The seam rule** (unchanged). A constraint belongs to JSON Schema iff its verdict
  is a pure function of **one document at one instant**. Anything needing a second input
  — `old_page`, `now`, or a second value — belongs to CEL.
- **P2. Capabilities and examples, never policy.** akb ships validators, projections,
  functions, and example templates. It does not ship opinions about what a KB contains.
  The embedded adr/note templates are showcases, deliberately exercisable, never prescriptive.
- **P3. One canonical author per fact.** Every fact is validated by exactly one layer
  and materialized at most once.
- **P4. Fail closed on write, degrade on sweep, prove at authoring.** Unchanged;
  extended to every new layer.
- **P5. The document carries facts, not computations.** The materialized view contains
  what neither validator can compute (parse facts). Anything computable from content is
  a CEL expression or a closed-registry function — never a pre-computed field.
- **P6. input = output.** The write path performs no mutation the caller did not request
  or declare. What akb persists is what it was given.
- **P7. Refuse at the door you configure; report everywhere else.** Write-time
  validation guards pages that pass through `akb write`. The sweep is the safety net for
  all pages regardless of origin (git pull, direct edits, ingestion). Both layers are
  user-configured; neither may be silently absent (§11, §16).

## 5. The document model (D12, D13, D7)

### 5.1 Shape

```jsonc
{
  "file": { "path": "notes/foo.md", "name": "foo.md", "dir": "notes" },
  "frontmatter": { "type": "note", "title": "..." },   // caller-supplied, raw YAML values
  "content": "# Body\n\nMarkdown as a string.",          // caller-supplied
  "derived": {                                           // akb-authored, output-only,
    "ast": { "headings": [...], "links": [...],          //   template-declared projections
             "code_blocks": [...] }                      //   (parse facts only)
  }
}
```

- **`content` is the markdown body, as a string, everywhere** — JSON envelopes, the
  document view, and the CEL `page` map. The `content.raw` object shape (an internal
  container that leaked into the public contract) is gone.
- **`derived` holds parse facts only** — projections of the goldmark AST that neither
  validator can compute itself. `word_count` and `char_count` are **scrapped**:
  `char_count` had zero rule consumers, counted bytes (mislabeled), and is covered
  natively by CEL `size(page.content)` and schema `minLength`/`maxLength` (both code
  points); `word_count` becomes a closed-registry CEL function (§10) because JSON Schema
  has no word-count keyword — materializing it only ever served CEL anyway.
- **`file` is derived from the page path** (caller's path, verbatim — §8); paths are
  relative to the KB with the `kb/` prefix **stripped** — today's `relPath` convention,
  shared with the search/linkgraph/index keys, so `file.path` is `notes/foo.md` and
  `file.dir` is `notes` (normalized to `/` separators). Surfaced in the `--derived`
  tier only. `line`-based position conventions throughout (byte-offset `position`
  fields die with the old scanners, §13).

### 5.2 Two artifacts, one builder (unchanged mechanism, new justification)

| Artifact | Definition | Consumers |
|---|---|---|
| **`document`** | Raw frontmatter values as parsed from YAML | JSON Schema validation; `read --json` (plain = caller-supplied subset; `--derived` = full view); agent tooling |
| **`page`** | Same shape; **schema-declared** temporal fields coerced in-memory for CEL | CEL validations + lint rules |

**CEL is evaluation-only.** Nothing in the validation layers mutates input, in memory
or on disk: `BuildPage` copies values into a fresh map (the §5.2 coercion included),
`Evaluate` is read-only, and `old_page` is built from a fresh on-disk read. The only
writer of page bytes is the write path persisting what it was given (P6).

**Date bridge (D7).** Silent, name-agnostic coercion dies. A temporal field is one the
template's schema declares with `format: date-time` or `format: date` (at whatever
depth it is declared — the old top-level-only limitation dies with name-agnostic
coercion; traversal mechanics → RFC-002-A4). Write-time schema validation **guarantees
CEL's preconditions** — a value reaching `timestamp()` is always parseable — because
akb's `date-time` assertion is **defined as the CEL acceptance set**, not the full
RFC 3339 grammar: RFC 3339 permits lowercase `t`/`z` and leap seconds, Go's
`time.RFC3339` (what cel-go's `timestamp()` parses with) rejects both — verified. So
the assertion profile is Go-strict RFC3339 (uppercase `T`, numeric offset or uppercase
`Z`, no leap seconds), pinned by [[ADR-003-santhosh-v6-format-assertion|ADR-003]] (the validator's format checker is
registered/overridden accordingly — the library's built-in accepts lowercase `t`/`z` and leap
seconds, and ADR-003 tightened the profile beyond raw Go acceptance: under the pinned cel-go
v0.28, `time.Parse(time.RFC3339)` additionally admits comma fractions, single-digit hours, and
out-of-range offsets, so ADR-003 I4 pins the strict pattern — the intersection of the cel-go
v0.28 and v0.32 acceptance sets — keeping this guarantee intact across the A3 upgrade) and documented in the SKILL as *akb's date rules*
(the same one-sentence treatment as "akb is RE2 everywhere"). `format: time` has no
CEL counterpart (no time type) — it validates as a string format only, never coerced.
For `format: date` fields, akb coerces date-only to **midnight UTC, in memory only,
never on disk** — the single akb-defined conversion, explicit and declared. Undeclared
fields are never coerced. **`old_page` receives the identical bridge**, built from the
fresh on-disk read coerced under the *current* template's declarations — every
immutability/comparison rule calls `timestamp()` on both sides, so both sides must be
bridged. Pages bypassing write
(git pull) are flagged by the sweep's schema checker per-page (P4 degrade).
RFC-002-A4 scope: bridge determinism spec; `template write`-time static check (a rule
calling `timestamp(x)` where `x` lacks a format assertion → warning); the duration
vocabulary mismatch (CEL `duration()` is Go-style `time.ParseDuration`; JSON Schema
`format: duration` is ISO 8601 — document the seam; an `iso_duration()` registry
function is the optional bridge).

### 5.3 Declared projections

Templates declare which AST projections akb materializes (`headings`, `links`,
`code_blocks`, and — once the dialect enables them — `tables`, `lists`, `blockquotes`,
etc.). Projection keys are additive by construction (`page` is `map(string, dyn)`);
the declaration syntax and registry are candidate RFC-002-A17. The `code_blocks` projection covers
**fenced and indented** code blocks alike — today's fenced-only view makes indented
code invisible to both validators (audit M10, probe-confirmed); (a′) makes the fix
automatic since ranges derive from the AST (§13.1). Marker/annotation projections exist
**only if the template declares their vocabularies** (§13.3) — akb no longer prescribes
`^[inferred]`/`^[ambiguous]`/`^[extracted]` or the foreign `olw-auto:` prefix.

**The CEL surface renames with the flatten** (stated explicitly — it is the
breaking-change inventory): `page.content` is the string; projections live at
`page.derived.ast.*`; declared marker/annotation projections live at
`page.derived.markers` / `page.derived.annotations`. The embedded templates' rules
rename mechanically — `page.ast.headings/code_blocks/links` → `page.derived.ast.*`,
`page.content.word_count >= 50` → `word_count(page.content) >= 50` — riding the
already-breaking V3 revision.

## 6. Template format (V3 — breaking, loudly rejected like the v1/v2 rejections)

```yaml
name: note
path: notes                 # OPTIONAL resolution hint for bare-filename addressing (D11).
                            # NOT a placement rule: caller paths are always verbatim (§8).
description: ...
state_field: is_draft       # OPTIONAL; declares the draft-lifecycle key akb's approve/list
                            # commands operate on (D3). Absent = no draft lifecycle for this type.
schema:                     # JSON Schema 2020-12 over the whole document (§5)
  $schema: "https://json-schema.org/draft/2020-12/schema"
  type: object
  properties:
    frontmatter:
      type: object
      required: [type, title]
      properties:
        type:    {const: note}        # D10: type is an ordinary key; schema enforces it
        title:   {type: string, minLength: 1}
        created: {type: string, format: date-time}   # D3/D7: temporal fields are
        updated: {type: string, format: date-time}   # template-declared, never tool-injected
    content: {type: string, minLength: 200}          # the body itself, schema-constrainable
    derived:
      type: object
      properties:
        ast:
          type: object
          properties:
            headings:
              type: array
              contains: {type: object, required: [level, text],
                         properties: {level: {const: 2}, text: {const: "Context"}}}
functions: [word_count]     # OPTIONAL subset of akb's closed registry (§10)
projections: [headings, links, code_blocks]   # OPTIONAL declared derived views (§5.3)
validations: []             # UNCHANGED: CEL, write-time, fail closed (old_page/now/cross-value)
lint_rules: []              # UNCHANGED: CEL, sweep-time; gains kb.* in Phase 4 (§11)
```

- **`schema.frontmatter` (homegrown block) is retired** with the RFC-001 hard-rejection
  migration path (precedent: `detectOldFormat`). Its presence job moves to schema
  `required`; the guard invariant stands: *the schema `required` set under
  `properties.frontmatter` is exactly the set of keys CEL rules may read unguarded.*
- **`type`/`title` stop being engine gates** (D10). `type` survives only as the
  **template selector**: a page's `type` names its template; a page naming an unknown
  template is refused at write (the one surviving door-bouncer, P7) and flagged by
  `type_orphan` at sweep. `frontmatter.Parse` stops routing `type`/`title` out of
  `Fields`; they become ordinary keys visible to both validators. `title` additionally
  becomes the **declared display field** (D18): the key the search title column,
  `index.md` entries, log entries, and the CEL page map read is declared (`title_field:`,
  default `title` — the §12 `summary` precedent, since prescribing the key would
  violate P2); undeclared → documented fallbacks (empty title column — BM25 title
  weight inert, content still ranks; filename as the index entry text).
- **Unknown template keys are rejected at authoring time** (the D9 principle applied
  to templates — the audit's trap: the loader's non-strict unmarshal,
  `template.go:106-114`, would silently ignore a typo'd `projection:`): `template
  write` hard-rejects unknown keys; *load* time only warns, so a template authored by
  a newer akb never bricks an older binary (same forward-compat story as RFC-002-A13).
- **Overlap rejection** (warning, not error — permissive P2) and mockup obligations
  carry from RFC-001 §6.2, extended to `<!-- FAILS: schema: /frontmatter/status -->`.
- **Overlap guidance (G6, carries from RFC-001 §10 — owner-ratified session 3):** the
  kb-management SKILL carries a one-screen, low-token "which validator?" decision rule
  — *checkable on this page alone, right now (type, presence, enum, pattern, length,
  list contents, a derived view) → `schema:`; needs the previous version (`old_page`),
  the clock (`now`), arithmetic, or value-to-value comparison → CEL; never state the
  same fact in both layers.* Placement: `TEMPLATE.md` (anatomy + narrowed guard
  doctrine), `MAINTAIN.md` (checker table incl. the schema checker), `APPROVE.md`
  (unbundled approve semantics) — all rewritten in Phase 4 (§17).
- **Defaults for the new keys:** `functions:` defaults to empty (no registry
  functions); `projections:` defaults to none (no `derived` block materialized — the
  minimal document, per P5); `state_field:` absent means no lifecycle for the type —
  `approve` on such a page is a usage error pointing at `lint` (there is no state to
  flip; validation is the sweep's job).
- **`dir` is renamed/re-specified as `path`** with resolution-hint semantics (D11, §8.3).

## 7. Validation pipeline (write/append)

```text
merge-conflict check → selector gate (type names a template; template loaded with its
  projections/functions declarations) → build document (frontmatter + content parse,
  goldmark parse, declared projections materialized) → JSON Schema → CEL validations → write
```

The build step is explicit because the schema may constrain the materialized document
(`contains` over `derived.ast.headings`) — validation cannot precede materialization,
and materialization cannot precede the declaration load (document-model audit, open
question 4).

RFC-001 §7 carries with two changes: the required-presence gate is the schema's own
`required` (no separate pre-gate), and **no mutation step exists** — validation either
passes and the bytes are written, or fails and nothing is written. Error merge (schema
violations by keyword + JSON Pointer; CEL failures by rule ID; one exit-1 report) and the
`runTemplateValidations` widening carry unchanged (RFC-002-S1).

## 8. The write path: input = output (D3, P6)

### 8.1 What dies

- **`created`/`updated` injection.** Zero internal consumers (audit-verified: search
  `--after` reads index time). Templates that want timestamps declare them
  (`format: date-time`, `required`, plus CEL immutability/comparison rules — the embedded
  examples will model this). akb never writes a timestamp a caller didn't write.
- **`is_draft` as tool vocabulary.** No deletion of `is_draft: true`, no implicit-draft
  semantics owned by the engine. A template declares `state_field:`; `akb approve`/`list`
  operate on the declared key (default `is_draft` for the embedded examples). AKB reads
  the field; it never normalizes it behind the author's back.
- **Whole-frontmatter re-marshal.** The map round-trip that destroyed comments, key
  order, quoting, and indentation is deleted. Markdown-stdin writes are **byte-verbatim
  always** (the `needsReserialize` conditional disappears because its triggers are gone).
- **`updated` bumps on append/--frontmatter.** Gone with the injection machinery.

### 8.2 What remains, transformed

- **`write --json`**: stdin `{frontmatter, content}` → JSON→YAML assembly (goccy
  `yaml.JSONToYAML` or typed decode — the `count: 3.0` landmine from RFC-001 stands) →
  full pipeline → stdout echoes the persisted document in **exactly `read --json`'s
  shape** (`{frontmatter, content}` — the caller passed the path as argv and already
  knows it), so write output is itself pipeable into write/read chains.
- **Mutual exclusion (RFC-001 §7 carries):** `--json` is incompatible with
  `--append`/`--frontmatter` (usage error, exit 2) — the partial-update paths stay
  markdown-native.
- **`--append`**: caller-requested body concatenation only (the `--dated` heading option
  is a caller request; its blank-line bug is fixed as ordinary bugfix work). Frontmatter
  is untouched.
- **`--frontmatter`**: values parsed as YAML scalars (today's string-only behavior
  corrupts lists); surgical, comment-preserving edit; deprecation in favor of
  `write --json` partial updates is candidate RFC-002-A11.
- **`approve` unbundled.** Three formerly-bundled jobs become explicit:
  (1) **full re-validation** — schema pass + write-time `validations` with
  `old_page` = on-disk page, `now` = approval time (closes the pre-existing hole);
  (2) **state flip** on the declared `state_field` (and nothing else);
  (3) **body stripping** only if the template declares it (e.g. `strip_on_approve:
  [markers, annotations]`) against its declared vocabularies — the closed 3-type strip
  list and `olw-auto` prefix die.

### 8.3 Paths (D11)

The caller's path is used **verbatim** — no strip/re-prefix machinery (M10/M11/M12 die
with it). The template `path:` key does exactly one job: tell bare-filename addressing
(`append foo.md`, `--frontmatter`) where to look first; resolution falls
back to the search index, and basename collisions across directories are an ambiguity
error naming the candidates. A template that wants a layout **policy** expresses it as
validation — CEL `page.file.dir == "notes"` or a schema `pattern` on `file.path` —
fail-closed at write like any other rule.

### 8.4 Defect repairs riding Phase 2 (no policy content)

- **Encoding hardening (audit M23):** BOM tolerance, UTF-8 validation, CRLF
  normalization, and a page size bound — today's BOM'd file fails with a confusing
  "no frontmatter found".
- **Segment-wise `..` check (M21):** the any-substring rejection (`my..draft.md` and
  `v1.2..3.md` are refused today for the wrong reason) is replaced by the segment-wise
  `filepath.Clean` + containment rule that already exists.
- **Clock unification (H5):** `--dated` used the local clock while injected timestamps
  were UTC; UTC everywhere.
- **Append join semantics (M13):** `--append` joins bodies with a deliberate newline
  policy — today a bare `"\n"`, which double-blanks a body already ending in a newline
  (the `--dated` blank-line symptom in §8.2 is the same defect).
- **Single managed-file predicate (M25):** the eight copied managed-file exclusion
  sites collapse behind one `path.ManagedPage()` predicate (§12 removes index.md from
  the set; `log.md` and `raw/files.log` remain).
- **Prefix handling consolidated (M20):** the five copies of `kb/`-strip /`raw/`-reject
  logic collapse into one `internal/path` normalization function; the convention
  itself (KB-relative paths, `kb/` stripped on page keys) is unchanged.
- **Bounded git invocation (C5):** every git subprocess gains a `context` timeout —
  today a hung git (credential prompt, hook, network FS) hangs akb forever, and the
  index-lock retry loop *waits* rather than bounds (zero `context.WithTimeout` in the
  tree).

## 9. JSON Schema coverage (G1 — carries from RFC-001)

Library-first: `github.com/santhosh-tekuri/jsonschema/v6` recommended (full 2020-12,
Go-native; ratified as the validator by ADR-003). Full keyword coverage over the flattened document
(`contains`/`minContains`, `if/then/else`, `$ref`/`$defs` including cycles,
`dependentRequired`/`dependentSchemas`, `unevaluated*`, etc.). Go validator ⇒ RE2
everywhere ⇒ the ECMA-262/RE2 dialect concern stays dissolved. **`format` policy (resolved by ADR-003):** assertion enabled for every declared format — a
declared format always asserts, at write time and at sweep; akb registers strict-profile
`date-time`/`date` checkers per §5.2; unknown format names warn at `template write`. What schema can
never express is unchanged and permanent: `old_page`, `now`, value-to-value comparison,
arithmetic, and body-text constraints beyond `pattern`.

## 10. CEL coverage (G2 + D8)

- **cel-go v0.28.0 → v0.32.x** (breaking module-path migration to `cel.dev/cel-go`;
  RFC-002-A3), then enable `ext.Strings`, `ext.Lists`, `ext.Sets`, `ext.Math`,
  `ext.Bindings` in `NewEnv`. Additive; every existing rule keeps working. Note the
  v0.30 `timestamp()` strictness — the §5.2 bridge is designed around it.
- **Closed function registry (D8).** akb ships a curated, documented, O(content)-bounded
  function set — seed: `word_count(string)`; candidates: `iso_duration(string)`,
  `url_host(string)`, plus the pair §11's `provenance` re-expression depends on:
  exclusion-aware `marker_count(content, type)` and `line_count(content)`/
  `non_empty_line_count(content)` (contingent on (a′)'s AST-derived exclusions) —
  which templates opt into via `functions: []`. **Users cannot
  author functions** (no free native bindings, no subprocess escape hatch): a native
  binding executes outside CEL's cost accounting, and an unbounded one is a DoS on the
  fail-closed write path. CEL rule strings remain the user-extensibility surface — the
  interpreter bounds them; the registry is how akb keeps that guarantee while still
  offering reusable helpers. Registry contents and per-function cost treatment → RFC-002-A5.
- **Cache key becomes `(env identity, expression)`** — required once function selection
  makes envs vary per template (today's expression-only key would silently return stale
  programs).

## 11. Lint ownership (D5)

- **Config registry (Phase A).** `akb.yaml` gains `lint.checkers.<name>: off|warning|
  error` over the built-in set (all on by default): `off` disables; `warning` **caps**
  every issue of that checker at warning severity; `error` preserves the checker's
  as-shipped severity mix. Per-issue-class granularity (e.g. silencing only
  `broken_links`' ambiguous-link warnings while keeping its errors) requires issue
  typing the checkers don't have — candidate RFC-002-A18 (audit G4: NEEDS DESIGN). Disabled
  checkers **keep their
  `by_check` key at 0** — the frozen `lint --json` envelope never loses keys (load-
  bearing for downstream typed gates; N5 carries). `lint_rules[].severity` becomes
  validated (today an unknown severity silently never fails). Dead thresholds
  (`FreshnessHalfLifeDays`, `FreshnessScoreThreshold`, `SummaryMinLength`,
  `SummaryMaxLength` — referenced only by a test pinning their literals) are deleted;
  `ProvenanceDriftThreshold` moves to config.
- **Built-ins become defaults, not privileges.** The registry mandate at
  `cmd/akb/lint.go:157-167` is the only thing that ever made them special.
- **Replace-before-you-disable** is the taught pattern (Phase-4 SKILL doctrine):
  author a replacement rule — or accept the gap deliberately — before turning a
  built-in off, and `akb lint` output always reports the disabled set so no KB is
  silently unguarded (P7).
- **Exit contract stated.** `akb lint` exits 1 iff any reported issue has severity
  `error`. Severities are KB-owned: downgrading a built-in to `warning` makes the
  sweep exit 0 with issues present — that is the ratified registry's purpose, and the
  contract says so plainly (answers audit open question 2).
- **Per-checker fault isolation.** One checker's error degrades to a per-checker issue
  instead of aborting the whole sweep (audit A4) — aligning built-ins with cel_lint's
  established degrade posture (P4). The error row is **actionable** (owner amendment):
  it names the checker, the cause (template/rule/page), and a remediation hint —
  degrade diagnostics guide repair, not just report (self-diagnosis depth → tech-spec).
  The triplicated exit-policy logic (`lint.go` ×2,
  `template_delete.go`) consolidates to one site; the dead `lint.PageData.Annotations`
  field is deleted.
- **Disposition of the 10 built-ins:**

  | Checker | Disposition |
  |---|---|
  | `missing_frontmatter` | Built-in (structural given the page model); scope/severity configurable |
  | `empty_pages` | Demotable to a shipped example `lint_rule` (`size(page.content)`-expressible) |
  | `required_fields` | Subsumed by the sweep-time schema checker (RFC-001 §13 carries) |
  | `type_orphan` | Built-in — the selector's sweep counterpart (D10) |
  | `cel_lint` | Already user-owned |
  | `broken_links`, `orphans` | Built-in until `kb.*`; then user-re-expressible (`delete --orphans`, a user-invoked convenience, is untouched) |
  | `citations` | Built-in; `sources` key name configurable; `kb.*`-re-expressible later |
  | `index_consistency` | **Dissolves** with §12 |
  | `provenance` | Built-in; threshold + `provenance` key configurable; user-re-expressible once the registry ships exclusion-aware marker-count and line-count functions (D8 + (a′)) — P5 forbids materializing the ratio, so the functions are the route |
  | `schema` (new) | Sweep-time structural pass (RFC-001 §13 carries) |

  Per-template checker *enablement* is not claimed: today's `LintChecker.Check(ctx,
  kb)` signature is global (audit feasibility 3), so demoted built-ins travel as
  copyable example `lint_rules`, not enableable units; a per-template enable surface
  is a `kb.*`-era question for RFC-002-S3.

- **Checker key names stop being prescribed (ratified with D5).** The hardcoded
  frontmatter keys the built-ins read — `sources` (citations), `provenance`
  (provenance) — become configurable (per-KB config or per-template declaration;
  mechanics → tech-spec). Marker/annotation *vocabularies* are template-declared per
  §13.3.
- **Sweep ordering (restated from RFC-001 §13 for self-containment):** the sweep-time
  schema checker runs **before** `cel_lint` — CEL never evaluates pages with
  known-missing required keys — and both passes share **one `now` stamp** per sweep.
- **`kb.*` — sweep-level CEL rules (Phase B; core commitment, later phase).** A new rule
  class evaluated per-sweep (not per-page) with a KB-level variable exposing cross-store
  facts (link graph, manifest, index) — the machinery that lets users *author* what
  today only built-ins can check (broken links, orphans, citations). Execution-model
  design (fact surface + freshness, KB-scale cost semantics, issue mapping) → RFC-002-S3.
  Phase A ships first because without it, `kb.*` authors couldn't disable the built-ins
  their rules replace.
- **Draft-awareness** of the sweep (today drafts can fail with `error`) → RFC-002-A12.

## 12. index.md becomes a derived artifact (D4)

`index.md` is mechanically regenerable from page frontmatter; this RFC makes derivation
its only mode. `akb index rebuild` (plus automatic post-write regeneration under the
write lock — the repo lock in git mode, the `.agent-kb/akb.lock` flock in non-git mode)
is the sole writer; `index add`/`index remove` and their argv-summary
path are retired (the summary source is a declared frontmatter field — `summary_field:`,
default `summary`, same declaration mechanism as D18's `title_field:`; mechanics →
RFC-002-A13/tech-spec — since prescribing the key would violate P2; entry titles likewise read
  the declared display field, D18, filename fallback). Pages with absent or unknown
  `type` are **not** skipped: `documents.type` stores the raw value, `index.md` groups
  them under an explicit `(untyped)` section, and `--type` matches literally (D19) —
  derived artifacts never lie, never crash, never silently drop (P7); `type_orphan`
  remains the sweep-time flag. Consequences, all intended:
the managed-file protection for `index.md` dissolves (one of the eight copied exclusion
sites' reasons to exist), the `index_consistency` checker dissolves, the write/delete
maintenance asymmetry (M26) and the two-sources-of-truth for summaries (M27) dissolve.
Derivation is **never silent** (P7): index and search rebuilds report every page they
skip for unparsable frontmatter — today both skip silently (`sqlite.go:275-278`,
`index.go:204-207`), which under pure derivation would let a broken page vanish from
the artifacts unnoticed; `missing_frontmatter` remains the error-severity net at sweep.
The naive pluralization heading grammar (`analysis`→`Analysiss`) is fixed or made
irrelevant in the renderer. The **parse side dissolves entirely**: with `index add`/
`remove` and `index_consistency` gone, nothing reads `index.md` back — `ReadIndex`/
`parseIndex`, the heading grammar, and the `- [Title](Path) — Summary` em-dash entry
contract (audit F3) die as dead code; the renderer is write-only (grammar → RFC-002-A9).
The remaining managed-file guards consolidate behind one
`path.ManagedPage()` predicate (§8.4). `log.md` stays managed (append-only, not derivable); its
closed operation enum (`ingest|distill|plan|...` — agent-workflow vocabulary, a
harness-coupling smell) becomes free-form or configurable (RFC-002-A10).

## 13. The markdown parsing pivot (D6)

### 13.1 Phase (a′): AST-derived exclusions — committed, parser phase

Build exclusion ranges from the goldmark AST akb already parses (`CodeBlock`/
`FencedCodeBlock` line spans, `CodeSpan` segments, `HTMLBlock`, escape presence) instead
of hand-rolled byte scanners. RFC-002-A6 (document builder) also owns consolidating the five
`goldmark.New()` call sites (≥2 parses/page/write) into one parse pipeline, and
`frontmatter.Parse`'s `extractBody` parity (document-model audit, open question 5).
~40 lines replacing ~500; deletes **both** exclusion-
scanner copies (`wikilink.go:410-935` category A, `pagebuilder.go:415-524`) and the
three duplicated parser bodies (provenance/annotation ×2 paths). Fixes, in one move:
the 3-class-vs-5-class drift (probe-verified: `^[inferred]` in indented code counted by
CEL, ignored by lint), the escaped-bracket split, the shared tilde-fence blind spot, and
the frontmatter-vs-body annotation offset split. Position semantics unify on `line`.
`is_wikilink` stops being source-sniffed and becomes a parse fact.

### 13.2 Phase (a): full goldmark inline parser — committed, follow-on ADR (RFC-002-A8)

The ~120–150-line `parser.InlineParser` for `[[...]]`, deleting the remaining string-
token grammar's exclusion duties. The ~100 `wikilink_test.go` cases are the real spec
and must be re-homed; the third-party `go.abhg.dev/goldmark/wikilink` extension covers
`[[target]]`, `[[target|display]]`, `[[target#fragment]]`, `![[embed]]` but **not**
akb's `[[display]](dest)` explicit-destination form — grammar work is required either
way (extend, fork, or from-scratch). Sequenced after (a′) deliberately: with the
scanners gone, (a) has exactly one job left — the grammar.

### 13.3 Vocabularies and dialect

Marker/annotation vocabularies become **template-declared** (a template names its marker
types and annotation comment prefix; parse is already open — `\^\[([^\]]+?)\]` — and the
strip side follows the declaration). The annotation **body grammar** is part of the
declaration, not just the prefix: today's whitespace-tokenized `k=v`
(`annotation.go:47-69`) means values can never contain spaces — declared grammars may
define quoting, and akb's default grammar is documented rather than accidental (RFC-002-A8).
The markdown dialect (CommonMark core today; tables silently parse as paragraphs)
becomes declarable via `akb.yaml`/template goldmark extension selection (`Table`,
`Strikethrough`, `Linkify`, `TaskList`, `Footnote`, `DefinitionList`, `Typographer`,
`CJK`) — required anyway for `tables`/`lists` projections to have nodes to project.
Ownership: `akb.yaml`'s `markdown.extensions` is the KB-global default; a template's
declared set overrides it for its pages (per-type dialects are legitimate — a type
whose content uses tables declares `Table`).

## 14. Config surface (D9)

`akb.yaml` becomes the behavioral config home (today: 3 fields, 0 behavioral, unknown
keys silently dropped — a trap this RFC closes with **strict unknown-key rejection**).
Seed schema (final shape → RFC-002-A13):

```yaml
name: mykb
description: ...                 # kept — live, consumed by init --description/discover
lint:
  checkers: {orphans: off, citations: warning}
  provenance_drift: 0.20
validation:
  cost_limit: 100000        # bound stays structural; the value becomes tunable
git:
  commit_message: "akb: {verb} {path}"   # 13 literal sites, zero parsers — free to open
search:
  index_fields: [tags, summary]          # needs DDL + rebuild; design → RFC-002-S4
markdown:
  extensions: [Table, Strikethrough]
```

The severity split is three-layered: as-shipped built-in severities live in the
checkers' code, `lint_rules[].severity` is per-template, and the `lint.checkers.*`
config value is a **ceiling override** (§11). Genuinely per-template concerns
(projections, layout policy, temporal fields, function selection, state field) live
in templates, not here. `akb.yaml` is the primary config home, **but per-invocation
flag overrides are permitted where they earn their keep** (P-i, ratified overturned:
sandboxed agents may be unable to read the KB's config — overrides such as `--now`
are sometimes necessary). Precedence: flag > `akb.yaml` > default.
The smalls ride RFC-002-A13: a `--now` flag for reproducible evaluation, discovery bounds,
file modes, deletion of the write-only `config.Created` field (dead since init), git
index-lock retry values, `templates/`/`search.db` naming, `<name>.yaml` template
naming, and the manifest `|`-in-filename corruption guard.
Strict unknown-key rejection needs a forward-compatibility story (spec-era `versioning:`/
`git-author:` keys must never hard-fail older binaries) — part of RFC-002-A13.
**Stays structural (not configurable):** the `.agent-kb/akb.yaml` marker itself, `kb/`/
`raw/` roots, path containment, template-name regex, WAL/single-connection DB, the
existence of a cost bound, merge preflight, `--only --` partial commits (in git mode).

### 14.1 Search semantics repairs (D17)

`akb search --after` is **rewired to the template-declared creation field**: the
indexer lifts it (precedent: `tags`/`summary` lifting, `sqlite.go:306-339`), normalized
to UTC `YYYY-MM-DD HH:MM:SS` so the lexicographic TEXT comparison stays sane, with a
`COALESCE` fallback to index time for pages whose type declares none. Today `documents.created`
is always `datetime('now')` at index time — reset by every rebuild *and* every
append/approve re-index — while the help text promises "creation date" (probe-verified,
`research/recon-search-after.md`). Also fixed: `--after` input validation (garbage
currently returns silent empty, exit 0) and RFC3339-input normalization (currently
silently wrong same-day — `' '` < `'T'`). `documents.updated` (index-time, zero
readers) is dropped or documented (RFC-002-S4); `--limit` becomes a real flag. The
creation-role declaration mechanics (schema annotation vs config key) → RFC-002-A16.
`--tag` matching semantics (today `LIKE '%x%'` substring over a flat space-joined
string — token vs substring) is decided with `search.index_fields` in
RFC-002-S4, not prescribed here.

## 15. JSON I/O and agent integration (D12, D14, D15)

| Surface | Shape |
|---|---|
| `akb read --json` | `{frontmatter, content}` — **literally** the write input |
| `akb read --json --derived` | adds `file` and `derived` (the only way an agent sees what CEL/schema see) |
| `akb write --json` | stdin `{frontmatter, content}`; `derived` or `file` keys in input = exit 2 |
| `akb template get <name> --input-schema` | the derived agent-facing schema (below) |
| `akb template get <name> --json` | machine-readable full template view |

**Round-trip invariant:** `read --json | write --json` is an identity. Full stop —
`file` is akb-derived and lives in the `--derived` tier, so the plain read emits
exactly the write input and the input schema's `additionalProperties: false` never
rejects a literal pipe; P6 removes RFC-001's "modulo managed mutations" caveat.
Pinned integration test carries.

**Input-schema derivation (D14).** From the template's `schema:` block akb mechanically
emits the caller-facing contract: the `frontmatter` subtree as-is; `content` as
`{type: string, ...}`; `derived` and `file` stripped; `required` restricted to
caller-supplied keys; `additionalProperties: false` at the top level (supplying derived
facts is rejected at the *agent* boundary too, not just at write). Emitted as single-line
strict JSON with explicit `type` on every node (agent-harness transport rules). One
source of truth; zero drift. The transport rule is explicit `type` on every
**assertion node** (adding inferred types where the harness grammar requires them,
e.g. `enum`-only nodes); `$ref`-only nodes are exempt by construction — they carry no
assertion (normalization mechanics → RFC-002-S5). **`$defs` and local `$ref`s survive derivation verbatim**
— only the `derived`/`file` property subtrees are stripped; a schema whose `$ref`
resolves into a stripped subtree is a `template write`-time error, not a silent rewrite.
The rule generalizes to every keyword: **any** reference from the retained schema into
a stripped subtree — `$ref`, `dependentRequired`/`dependentSchemas`, `if/then/else`,
`unevaluated*` — is a `template write`-time error (keyword-walk mechanics → RFC-002-S5).
Only then does "one source of truth; zero drift" hold across the full 2020-12 surface.

**Reference flow (D15) — harness-agnostic, with pi-subagents as the worked example.**
Markdown travels **as a JSON string** inside `content`; akb never generates markdown
from structured data (that design — rejected in RFC-001 §5.2 and re-affirmed here — is
the only thing that was ever rejected; the derived view is lossy by construction, so
akb-rendered markdown would destroy documents on round-trip). The flow:

1. Template authored → `akb template get --input-schema` exported.
2. Injected into any structured-output mechanism (pi-subagents `outputSchema` — static
   frontmatter or per-run launch plan; validated in-process by TypeBox, drafts
   3→2020-12; the gated `structured_output` tool call rejects with re-emittable
   diagnostics). Other harnesses wire their own equivalent.
3. The validated payload is piped by the orchestrator to `akb write --json` (verified:
   pi-subagents acceptance gates receive no stdin, so piping is the parent's job).
4. akb re-validates (schema + CEL — authoritative, covers what TypeBox can't see:
   `old_page`, `now`, cross-value rules) and persists bytes via the usual pipeline.

Verified pi-subagents constraints the flow must respect (schema-survey.md): a typed
acceptance gate **cannot** combine with an `outputSchema` on the same child
(`TYPED_VERIFY_OUTPUT_SCHEMA_CONFLICT`) — so a child emitting structured output and a
step running an `akb lint --json` typed gate are different children; rejection
diagnostics are bounded at 4096 bytes; and the runtime wraps the agent schema one level
down, rewriting local `$ref`s (the derivation above emits the *unwrapped* root, so
refs resolve correctly for both harness and akb).

Double validation is intended: the harness gate gives the agent immediate feedback;
akb is the system of record, and the sweep (§11) covers everything that bypassed both.

## 16. What akb still prescribes (the honest, complete list)

After all demotions, akb mandates exactly: the `.agent-kb/akb.yaml` marker; `kb/` +
`raw/` roots and the `.md` page extension; path containment (symlinks, segment-wise
`..`, absolute); the template-name regex; a page must carry frontmatter naming a known
template (the selector — the one write-time refusal); git versioning per the
**init-versioning spec** (`docs/kb/spec/init-versioning-spec.md`, lands before
this RFC's phases — D-seq): `versioning: git|none`, and in git mode auto-commit with
`--no-commit` escape, merge preflight (conflicted *and* clean-merge detection), and the
spec §6 identity chain (env `AKB_AUTHOR_*` → `akb.yaml` `git-author`/`git-email` →
git-native; repo config is never written — replacing today's `-c user.name=akb
-c user.email=akb@local` per-invocation fallback, audit M29/C3); in **non-git mode**
`FilesystemProvider`
(byte-verbatim), flock on `.agent-kb/akb.lock`, `--no-commit` a no-op, `akb status`
reporting `Versioning: none`, and an unresolvable commit identity a preflight exit 2
(never a raw git fatal) — spec §4/§6; the write path's coupling to an initialized
`.agent-kb/search.db` in **both** versioning modes (D16, ratified: the transactional
write+index guarantee is what makes search and the link graph trustworthy —
structural); `akb read` stays coupled to neither git nor the DB (byte-verbatim
filesystem read, audit M32); WAL/single-connection SQLite with its fixed schema +
`VerifySchema` (indexed-field changes go through DDL migration + rebuild, §14); the
0/1/2 exit-code + output-prefix contract (incl. `mutatingCommands` gating, audit H7);
the existence (not value) of a CEL cost bound; `--only -- <paths>` partial commits in
git mode (akb commits only what akb touched); the `lint --json`
envelope shape; machine-ownership of `log.md` and `raw/files.log`; and the search
layer's three deliberate opinions: body-verbatim indexing (markers and annotations
included — search finds what the author wrote), FTS5-syntax stripping (agents get
literal search, not operator injection), and fixed BM25/snippet/limit defaults
(tunable via RFC-002-A13). Everything else is user-declared, derived, or configured. Each
entry here carries its structural justification; nothing else may accrete without one.

## 17. Phasing (sizing via tech-specs; each phase independently shippable)

**Precondition (D-seq, ratified):** the init-versioning spec lands **before** Phase 1 —
it is agreed and implementation-ready against current akb; Phase 2 would otherwise force
its storage plumbing to be designed twice, and it gifts this RFC the first behavioral
`akb.yaml` keys (`versioning`, `git-author`/`git-email`) for §14's strict loader.

- **Phase 1 — validation core:** santhosh v6 integration; TemplateV3 `schema:` block;
  pipeline reorder; merged error model; cel-go v0.32.x + ext libraries; `template write`
  schema obligations; `schema.frontmatter` hard rejection.
- **Phase 2 — write-path honesty + JSON I/O:** input=output (injection/re-marshal/state
  deletion removed); approve unbundled; `read/write --json`; `--input-schema` export;
  round-trip invariant tests; date bridge (declared temporal coercion); the §8.4
  defect sweep.
- **Phase 3 — openness machinery:** akb.yaml config surface + strict keys; lint config
  registry (+ `by_check` stability); `index.md` derivation; parser pivot (a′);
  vocabulary declarations; `path:` resolution-hint semantics; §14.1 search repairs
  (D17 rewire, `--after` validation, `--limit` flag).
- **Phase 4 — extensibility + doctrine:** closed function registry (+ `word_count`);
  `kb.*` sweep-level rules; inline-parser rewrite (a) per RFC-002-A8; SKILL/kb-management
  doctrine rewrite; embedded example restructure.
- **Every phase also ships its own doctrine updates** (init-versioning spec §10
  precedent): the repo `AGENTS.md` — loaded into every session — currently asserts
  TemplateV2, engine-level `type`/`title` gates, silent name-agnostic date coercion,
  "10 lint checks", and "thresholds not configurable in v1"; each phase corrects what
  it invalidates rather than leaving stale doctrine in agents' contexts.

## 18. Spawned ADR / tech-spec candidates

The documents below are **candidates proposed by this RFC, not standing decisions**.
They carry RFC-local IDs — `RFC-002-A<n>` for ADRs, `RFC-002-S<n>` for tech-specs —
disjoint from the KB's global `ADR-NNN`/`SPEC-NNN` namespaces by construction, so no
candidate reference can be mistaken for, or collide with, a real decision document.
Candidate IDs exist only within this RFC; no file is ever created under a candidate
name. Lifecycle:

1. A candidate is drafted as a real ADR/SPEC, taking the next free global number; its
   header records `Origin: RFC-002 candidate <ID>`.
2. On ratification this RFC is amended: the row moves to `ratified` with its global
   number, and every body reference is rewritten from the candidate ID to the
   standing document.
3. A candidate rejected in drafting stays in the table as `dropped` (with reason) —
   IDs are never reused, so the record of the road not taken survives.

This RFC remains `status: draft` until every row below is `ratified` or `dropped` —
it is the last document approved, and this table is its progress tracker.

| ID | Kind | Candidate | Status | Resolves-as |
|---|---|---|---|---|
| RFC-002-A1 | ADR | Validator library (santhosh v6 recommended) + `format` assertion policy | ratified | [[ADR-003-santhosh-v6-format-assertion|ADR-003]] |
| RFC-002-A2 | ADR | `schema.frontmatter` retirement + migration (`akb template migrate`?) | proposed | — |
| RFC-002-A3 | ADR | cel-go upgrade (v0.28.0 → v0.32.x, `cel.dev/cel-go`) + extension set | proposed | — |
| RFC-002-A4 | ADR | Date bridge determinism + `template write` static temporal check; duration seam | proposed | — |
| RFC-002-A5 | ADR | Closed function registry: contents, selection syntax, per-function cost treatment; cache key | proposed | — |
| RFC-002-S1 | SPEC | Structured validation report (widened `runTemplateValidations`) | proposed | — |
| RFC-002-A6 | ADR | Document builder rewrite (typed structs, one parse, line conventions) | proposed | — |
| RFC-002-A7 | ADR | Derived-view versioning (`schema_version` marker) | proposed | — |
| RFC-002-S2 | SPEC | (a′) AST-derived exclusions — implementation tech-spec | proposed | — |
| RFC-002-A8 | ADR | (a) inline parser + wikilink test re-homing | proposed | — |
| RFC-002-S3 | SPEC | `kb.*` execution model (fact surface, freshness, cost semantics) | proposed | — |
| RFC-002-A9 | ADR | `index.md` retirement mechanics + renderer grammar | proposed | — |
| RFC-002-A10 | ADR | Log operation enum + `log.md` heading grammar (the parse contract, distinct from the enum) | proposed | — |
| RFC-002-A11 | ADR | `--frontmatter` fate | proposed | — |
| RFC-002-A12 | ADR | Sweep draft-awareness | proposed | — |
| RFC-002-A13 | ADR | Config schema (full field list, strictness mode) | proposed | — |
| RFC-002-A14 | ADR | CEL→schema portability report (optional, diagnostics-only) | proposed | — |
| RFC-002-A15 | ADR | `state_field` semantics (approve/list/`--all-drafts` over a configurable key) | proposed | — |
| RFC-002-A16 | ADR | D17 mechanics: creation-role declaration (schema annotation vs config key), normalization, `COALESCE` fallback, mixed-declaration KB semantics | proposed | — |
| RFC-002-A17 | ADR | Projection declaration syntax + registry (§5.3) | proposed | — |
| RFC-002-A18 | ADR | Lint per-issue-class granularity (issue typing for checkers; audit G4) | proposed | — |
| RFC-002-S4 | SPEC | Search: `search.index_fields` DDL + rebuild, `--tag` matching semantics, `documents.updated` disposition | proposed | — |
| RFC-002-S5 | SPEC | Input-schema derivation mechanics (normalization, reference keyword-walk) | proposed | — |

Renumbering note: items 6, 9, and 11 of the original flat 1–19 list were tech-specs,
so the IDs above are not the old ordinals (old 6 → S1, 9 → S2, 11 → S3; old 7+
shift accordingly). Rows A17, A18, S4, and S5 were added in the same pass to
enumerate deferred decisions the body previously mentioned only generically
(§5.3 projections, §11 issue-class granularity, §14/§14.1 search, §15 input-schema
derivation); S3's scope was extended to cover the per-template checker enable
surface. Every body reference was rewritten accordingly.

## 19. Risks

| Risk | Mitigation |
|---|---|
| Contract churn (`by_check`, `list --json.is_draft`, `index.md` consumers) | Stability statements per surface: `by_check` keys (§11), `index.md` (§12), `list --json` draft field (RFC-002-A15); disabled checkers keep keys; N5 carries |
| Permissive ⇒ invalid pages possible | P7 layering made explicit; selector gate stays; sweep is load-bearing (D11 ingestion already made it so); lint reports the disabled-checker set |
| Scope explosion | Four independently-shippable phases; openness demotions slotted by audited blast radius, not enthusiasm |
| Migration pain for existing KBs | Template V3 hard rejection with migration message (precedent: v1); existing pages' `created`/`updated` stay on disk (they're just data now); guarded temporal rules degrade to vacuous until templates declare |
| `kb.*` becomes permanent vaporware | Committed in-register (D5); Phase-4 placement is scheduling, not optionality; tech-spec follows ratification |
| Known unknowns (unaudited residue) | Embedded-skill prose consumers of `created`/`updated` never audited; DESIGN.md typed-gate dependency cited secondhand; testscript integration churn per demotion unenumerated — each phase's tech-spec must enumerate its test fallout and re-verify these (tracked in research/LOG.md) |
| Date bridge becomes a new silent-coercion | It is schema-declared, assertion-gated, in-memory-only, and the only conversion akb performs — specified by RFC-002-A4 with a `template write`-time static check |

## Appendix A — Decision register (owner-ratified)

| # | Decision | Ratified (2026-09) |
|---|---|---|
| D1 | RFC disposition | New RFC-002 supersedes RFC-001 (26/27) |
| D2 | Posture | Maximal direction, phased delivery (27) |
| D3 | Managed fields | Template-declared + input=output (27) |
| D4 | index.md | Fully derived artifact (27) |
| D5 | Lint | Config registry + `kb.*` (core commitment, later phase) (27) |
| D6 | Parsers | Both, sequenced: (a′) in parser phase; (a) follow-on ADR RFC-002-A8 (27) |
| D7 | Dates | Schema-declared bridge (27) |
| D8 | CEL functions | Closed registry only; no user-authored functions (27) |
| D9 | akb.yaml | Behavioral config home + unknown-key strictness (27) |
| D10 | Type gate | Keep, selector only (27) |
| D11 | Layout | `path:` resolution-hint only; caller path verbatim; policy via CEL/schema (27) |
| D12 | Envelope | Flatten: `content` = markdown string everywhere; `derived` output-only (27) |
| D13 | Metrics | `char_count` scrapped; `word_count` = registry function; facts-not-computations (27) |
| D14 | Input schema | akb derives + exports (`--input-schema`) (27) |
| D15 | Harness scope | akb contract + reference flows (27) |
| D16 | Write coupling | search.db coupling **kept** (transactional write+index is structural); git half pre-decided by init-versioning spec (`versioning: git\|none`) (27) |
| D17 | `search --after` | Rewire to declared creation field; `COALESCE` fallback to index time (27) |
| D18 | `title` | Declared display field (`title_field:`, default `title`); documented fallbacks when undeclared (27) |
| D19 | Untyped pages in artifacts | Store raw, group explicitly (`(untyped)` section; literal `--type` match); never skip silently (27) |
| D-seq | Sequencing | init-versioning spec lands **before** RFC-002 phases (27) |

### P-register — drafting positions, owner-ratified 2026-09-27

Taken under ratified principles during drafting; ratified via interview (P-a…P-h as
positioned; **P-i overturned** — flags permitted where they earn their keep; P-f with
the actionable-feedback amendment):

| # | Position | Outcome | Where |
|---|---|---|---|
| P-a | Stripped-`relPath` convention kept (`file.path: "notes/foo.md"`) | **Ratified** | §5.1 |
| P-b | `write --json` echo = `read --json` shape | **Ratified** | §8.2 |
| P-c | `approve` without `state_field` = usage error pointing at `lint` | **Ratified** | §6 |
| P-d | `file` lives in the `--derived` tier | **Ratified** | §15 |
| P-e | Lint exits 1 iff any issue has severity `error` (severities KB-owned) | **Ratified** | §11 |
| P-f | Per-checker fault isolation (degrade, not abort) | **Ratified** + amendment: error rows are actionable (checker, cause, remediation hint) | §11 |
| P-g | `delete --orphans` untouched | **Ratified** | §11 |
| P-h | `log.md` operation enum opened (free-form or configurable) | **Ratified** | §12 / RFC-002-A10 |
| P-i | No new per-invocation flags | **OVERTURNED** — flags allowed where they earn their keep (sandboxed agents may not read the KB config); precedence flag > akb.yaml > default | §14 |

## Appendix B — Carried from RFC-001 unchanged

P1 seam rule; coexistence-over-translation rationale + external research (K8s CRDs,
protovalidate, flux-schema, Datree); G1/G2 coverage commitments; G6 overlap guidance
(now §6); N1 (no JSON-Schema→CEL or CEL→JSON-Schema translation as a validation
mechanism — the portability report stays diagnostics-only, RFC-002-A14); N2 (no `.json`
page format — JSON is transport/validation view, never storage); P4 failure contract;
merged error model + `runTemplateValidations` widening; JSON→YAML assembly landmine
(goccy `yaml.JSONToYAML`); N3/N5 non-goals; **N4 deliberately superseded** — §11
(`kb.*`), §12 (derivation), and §14.1 (indexer declarations) touch search/linkgraph/
index, but as data sources for user rules and derived artifacts, never as validators;
no subsystem gains validation authority; overlap warning at `template write`;
mockup obligations; the `--json`/`--append`/`--frontmatter` mutual exclusion (§7);
DESIGN.md roadmap interactions (T1 subsumption, typed-gate synergy,
D11 sweep criticality); the §4.3 secret-scanning write gate — unchanged, a third,
separate, opt-in gate that is *not* a validator (so §1's "only validators" claim is
unaffected); cel-go release analysis (v0.29–v0.32).

## Appendix C — Research base

`research/`: `LOG.md` (full session log + decision history); RFC-001-era: `akb-recon.md`,
`schema-survey.md`, `cel-bridge-research.md`, `cel-bridge-research-b.md`,
`command-inventory.md`, `seam-core-pipeline.md`, `seam-periphery.md`; openness audit:
`openness-write-path.md`, `openness-lint-engine.md`, `openness-document-model.md`,
`openness-config-surface.md`; `recon-search-after.md`; `SYNTHESIS-openness.md`.
