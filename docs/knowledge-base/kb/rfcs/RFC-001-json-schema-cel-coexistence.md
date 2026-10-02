---
grammar: 1
type: rfc
id: RFC-001
title: JSON Schema / CEL Coexistence and JSON I/O in akb
status: superseded
provenance: agent-drafted
author: pete
steward: pete
created: 2026-09-26
updated: 2026-10-02
scope: ["**"]
tags: [validation, json-schema, cel]
summary: Add JSON Schema 2020-12 alongside CEL plus JSON I/O for agent pipelines; superseded by RFC-002, which replaces akb's prescriptive core rather than augmenting it.
superseded_by: RFC-002
---

# RFC-001: JSON Schema / CEL Coexistence and JSON I/O in akb

| | |
|---|---|
| **Status** | **SUPERSEDED by RFC-002** (`RFC-002-open-validation-engine.md`, 2026-09-27) — retained as research record; its technical spine carries into RFC-002 (see RFC-002 §3 and Appendix B) |
| **Date** | 2026-09-26 |
| **Authors** | pete (owner), pi session + subagent research team |
| **Research base** | `.pi/subagents/proposals/json-cel/research/` (7 reports + LOG.md) |
| **Supersedes / amends** | `agent-memory/DESIGN.md` §4.2/§4.3/§5 (amends; JSON direction is new relative to that ratified design) |
| **Downstream artifacts** | ADRs + tech-specs per §15 (nothing here is implementation-final) |

---

## 1. Abstract

akb validates knowledge-base pages with CEL rules in typed templates. This RFC proposes
adding **native JSON Schema 2020-12 validation as a first-class, co-equal validation layer**,
plus JSON input/output on the read/write paths. The two validators operate **side by side** —
no translation, no compilation between them. JSON Schema validates a single materialized
**document** (frontmatter + akb-derived views of the markdown body); CEL keeps everything that
needs a second input (`old_page`, `now`), cross-field value comparison, or general arithmetic.
CLI gains `akb read --json` and `akb write --json`, making akb pages pipeable to and from
agent structured output (e.g. pi-subagents `outputSchema` / typed gates).

## 2. Motivation

akb is an agent-first tool. Agentic harnesses already speak JSON Schema natively — tool
definitions, structured output, gate schemas (pi-subagents validates `outputSchema` and typed
gates with TypeBox, full draft 3→2020-12). Today an agent cannot:

1. read a page as structured data (`akb read` emits raw markdown bytes only),
2. write a page from structured data (`akb write` accepts markdown-over-stdin only),
3. declare structural constraints (types, enums, shapes) in the ecosystem-standard vocabulary —
   akb's `schema.frontmatter` block is a homegrown mini-schema whose `type`/`enum` fields are
   currently unenforced (known; fix in progress), and every real constraint must be hand-written
   as a CEL string.

External research (two independent surveys; see `research/cel-bridge-research*.md`) established
that no JSON-Schema→CEL compiler exists anywhere, and that every production system facing this
exact pairing — Kubernetes CRDs (`x-kubernetes-validations`), protovalidate, flux-schema,
Datree — converged on **coexistence**: declarative schema for structure, CEL for the residue.
Translation designs carry permanent fault lines (regex dialect, recursion, null-vs-absent,
numeric model, error granularity); JSON-only designs are impossible for a KB (no arithmetic,
no relative time, no cross-field comparison). Coexistence has no structural fault.

## 3. Goals / Non-goals

### Goals

- **G1.** Native JSON Schema **2020-12** validation of pages, with **maximum coverage** of the
  validation vocabulary (see §8). Coverage is a library-choice commitment, not hand-rolled code.
- **G2.** Maximum CEL coverage: track a **current** cel-go (v0.32.0 at time of writing;
  akb pins v0.28.0 — four releases behind) and enable the official extension libraries
  (`ext.Strings`, `ext.Lists`, `ext.Sets`, `ext.Math`, `ext.Bindings`; `ext.Regex`/optionals
  deferred to an ADR). The upgrade is non-trivial: v0.32.0 migrates the module path to
  `cel.dev/cel-go` (breaking, cel-expr/cel-go#1413) → ADR 3.
- **G3.** One materialized **document** per page combining metadata and content-derived views;
  JSON Schema validates that document. A metadata-only JSON mode is rejected (it would force
  duplicating content facts into metadata so rules can see them).
- **G4.** `akb read --json` and `akb write --json` — pages pipeable to/from agent structured
  output. Write path reuses the existing pipeline (locks, git, indexing) unchanged.
- **G5.** Zero-preference permissiveness: akb ships *examples*, not opinions. Agents author
  whatever templates their job needs; the embedded adr/note templates remain showcases only
  (the adr template is deliberately convoluted to demonstrate CEL's range).
- **G6.** Where a constraint is expressible in **both** languages, the kb-management SKILL
  gives one clear, low-token decision rule (§10).
- **G7.** Preserve the three-surface failure contract: write fails closed, lint sweep degrades
  per page, `template write` proves mockups.

### Non-goals

- **N1.** No JSON-Schema→CEL or CEL→JSON-Schema translation as a validation mechanism.
  (A lossy CEL→schema *portability report* remains an optional, diagnostics-only extra — ADR
  candidate, not part of this RFC.)
- **N2.** No `.json` page file format. Pages remain markdown + YAML frontmatter on disk. JSON
  is a *transport and validation view*, not a storage format. (All page walkers filter `.md`.)
- **N3.** No schema for akb-managed files (`index.md`, `log.md`, `raw/files.log`) — machine-owned.
- **N4.** No validation change for search, linkgraph, raw, or git subsystems — they are
  downstream consumers or exempt walkers, not validators.
- **N5.** `akb lint --json`'s envelope shape is **stable** — it is load-bearing for
  `agent-memory/DESIGN.md` §4.3 typed gates (`gate: {command: "akb lint --json", output: "json",
  schema}`). New checkers may add rows; the envelope does not change.

## 4. Design principles

- **P1. The seam rule.** A constraint belongs to JSON Schema iff its verdict is a pure function
  of **one document at one instant**. Anything needing a second input — `old_page`, `now`, or a
  second value to compare against — belongs to CEL. This is not arbitrary: `old_page` and `now`
  are the only non-page CEL variables (`internal/cel/engine.go:31-35`).
- **P2. Examples, not prescriptions.** The embedded templates exercise the machinery; they do
  not define the coverage target. Coverage targets are the two upstream specifications
  (JSON Schema 2020-12, cel-go latest), per G1/G2.
- **P3. One canonical author per fact.** Every page fact is validated by exactly one layer.
  The current duplication of `schema.frontmatter.status.enum` into a hand-written CEL
  `in [...]` rule (adr.yaml:29-32 vs 68-72) is the anti-pattern this RFC eliminates.
- **P4. Fail closed on write, degrade on sweep, prove at authoring.** Unchanged; extended to
  the schema layer.
- **P5. `page` ≠ `document`.** The CEL input map (date-coerced) and the JSON view (raw values +
  derived block) are two named artifacts with one builder. Neither leaks into the other's
  role (§5).

## 5. The document model

### 5.1 Two artifacts

| Artifact | Definition | Consumers |
|---|---|---|
| **`document`** | `{file, frontmatter, content:{raw, word_count, char_count}, ast:{headings, links, code_blocks}, akb:{provenance_markers, annotations}}` — frontmatter values **as parsed from YAML** (no date coercion) | JSON Schema validation; `read --json` output; agent tooling |
| **`page`** | the CEL evaluation map — same keys, but date-like top-level frontmatter strings coerced to `time.Time` (`convertDateField`, pagebuilder.go:29-45) | CEL validations + lint rules |

Rationale: `BuildPage` rewrites `created: 2026-01-05` into a `time.Time`, which JSON-marshals
as `2026-01-05T00:00:00Z`. If the JSON view carried coerced values, `read --json | write --json`
would silently mutate every date field. Both artifacts come from **one builder** so they cannot
drift (the dead typed structs in `internal/cel/types.go` are the natural home for this contract).

### 5.2 Derived views are akb-authored, output-only

`content.word_count/char_count`, `ast.*`, `akb.*` are deterministic pure functions of the
markdown body (goldmark + wikilink/annotation/provenance parsers — all verified pure,
`research/seam-core-pipeline.md` §3). On **write**, the caller supplies exactly
`{frontmatter, content.raw}`; **supplying any derived key is a usage error (exit 2)**, never
silently ignored — a caller must not be able to assert AST facts that contradict the body.
Structured-content input (akb generates markdown from JSON) is **rejected**: the derived view
is lossy by construction (paragraphs, lists, tables, emphasis are absent), so round-tripping
would destroy documents.

### 5.3 Pre-publication hazards (must fix before the view is a contract)

1. Duplicated code-exclusion logic in `pagebuilder.go` (copies at :345, :369, :415-524) vs
   `internal/markdown`'s unexported equivalents — publish means export-and-share, else
   divergence becomes a user-visible contract break.
2. `akb.annotations[].position` is an offset into the full page on the lint path but into the
   body on the write path — pick one.
3. `char_count` counts **bytes**, not code points; `word_count` counts whitespace fields
   including markdown syntax. Both need documented definitions before a schema can reference
   them (JSON Schema `minLength` counts code points).
4. `file.dir` leaks OS separators (`filepath.Dir`); normalize to `/`.
5. `line` numbering convention (`source` parameter passed inconsistently) must be pinned.

## 6. Template format

### 6.1 Shape (TemplateV2 revision — breaking, loudly rejected like the v1 rejection)

```yaml
name: example
dir: examples
description: ...
schema:                # NEW: a JSON Schema 2020-12 document over the `document` view
  $schema: "https://json-schema.org/draft/2020-12/schema"
  type: object
  properties:
    frontmatter:
      type: object
      required: [title, status]
      properties:
        title:  {type: string, minLength: 1, maxLength: 80}
        status: {type: string, enum: [proposed, accepted]}
        tags:   {type: array, minItems: 1, items: {type: string, pattern: "^[a-z0-9-]+$"}}
    content:
      type: object
      properties:
        word_count: {type: integer, minimum: 50}
    ast:
      type: object
      properties:
        headings:
          type: array
          contains: {type: object, required: [level, text],
                     properties: {level: {const: 2}, text: {const: "Context"}}}
validations:           # UNCHANGED: CEL, write-time, fail closed — old_page/now/cross-value
  - id: valid_state_transition
    rule: '...'
    expect: '...'
lint_rules:            # UNCHANGED: CEL, sweep-time temporal policy
  - id: example_stale
    rule: '...'
    severity: warning
    expect: '...'
```

- **`schema.frontmatter` (the homegrown block) is retired.** Its one enforced job — presence —
  moves to JSON Schema `required` under `properties.frontmatter`. Templates using it get a hard
  rejection with a migration message (precedent: `detectOldFormat`, template.go:73-82).
  Migration path details → ADR.
- **The invariant that preserves the guard doctrine:** *the JSON Schema `required` set under
  `properties.frontmatter` is exactly the set of keys CEL rules may read unguarded.* The
  presence gate (`checkRequiredFields`) is re-sourced from the schema; its three consumers
  (write, lint, template-write mockups) follow.
- **`validations[]` / `lint_rules[]` are unchanged in syntax and semantics.** CEL gains
  capability only via the newly enabled ext libraries (G2) — additive, non-breaking.
- **Dialect pin:** 2020-12 (needed anyway: `contains` is draft-06+, `if/then` draft-07+).
- **Overlap rejection:** a template that constrains the same fact in both layers (e.g.
  `status.enum` in the schema block *and* a CEL `in [...]` rule) is a lint warning at
  `template write` time — warning, not error, because akb stays permissive (G5); but the SKILL
  teaches against it (P3).

### 6.2 Mockup obligations (extend `akb template write`)

- Pass mockup must satisfy the schema block **and** all CEL rules (cheapest satisfiability
  check for has-no-solution schemas).
- Fail mockup must fail the **intended** validator; the existing machine-readable convention
  (`<!-- FAILS: rule_id — ... -->`) extends to `<!-- FAILS: schema: /frontmatter/status -->`.
- The optional-key-stripping and self-`old_page` proofs stay CEL-side; "declared optional keys"
  is re-derived from the schema block.

## 7. Validation pipeline (write/append)

```text
merge-conflict check → required gate (from schema) → JSON Schema → CEL validations → write
```

- **Order:** structural first, CEL second — CEL may then read required keys unguarded, and a
  page failing shape never pays comprehension cost.
- **Error merge:** schema violations (keyword + JSON Pointer) and CEL rule failures (rule ID)
  merge into one report, one exit-1 outcome. Human-readable stderr stays as today; a structured
  form requires widening `runTemplateValidations` (currently returns data-free
  `validationFailure{}`; two callers: write.go:521, append.go:222) → tech-spec.
- **JSON→YAML assembly (measured landmine):** `encoding/json` → `yaml.Marshal` turns `"count":3`
  into `count: 3.0`; `json.Decoder.UseNumber` yields quoted strings. Assembly must use goccy
  `yaml.JSONToYAML` (or a typed decode), then re-parse through `frontmatter.Parse` so the
  **entire existing pipeline runs unchanged** — one writer, one canonical form.
- **`--json` is mutually exclusive with `--append`/`--frontmatter`** (usage error, exit 2).

## 8. JSON Schema coverage commitment (G1)

Under native validation, coverage = the validator library's 2020-12 conformance; akb hand-rolls
nothing. Recommended library: **`github.com/santhosh-tekuri/jsonschema/v6`** (full 2020-12,
optional format assertions; Go-native). Final choice → ADR (candidates: santhosh v6,
invopop/jsonschema for generation, qri-io/jsonschema).

Consequences of the library-first approach:

- **Full keyword coverage over the `document`**: `type`, `enum`, `const`, `required`,
  `properties`, `additionalProperties`, `items`, `prefixItems`, `contains`/`minContains`,
  `min/maxItems`, `min/maxLength`, `pattern`, numeric bounds, `multipleOf`, `$ref`/`$defs`
  (including local reuse), `allOf/anyOf/oneOf/not`, `if/then/else`, `dependentRequired`,
  `dependentSchemas`, `propertyNames`, `unevaluatedProperties/Items`, `format`
  (annotation-default; assertion opt-in per template — details → ADR).
- **Regex dialect simplification:** a Go validator compiles `pattern` with Go's RE2 — the same
  dialect as CEL `matches`. The ECMA-262-vs-RE2 divergence documented in research becomes a
  non-issue in practice: **akb is RE2 everywhere**, and the SKILL says so in one sentence.
- **What no schema can express (permanent, by construction — this is CEL's territory):**
  `old_page` (transitions, immutability), `now` (staleness, deadlines), value-to-value
  comparison (`updated >= created`), cross-subtree existentials ("frontmatter X must appear as
  a wikilink target"), arithmetic, and anything about `content.raw` *text* (schemas see the
  derived views, not the markdown).
- **Recursion** (`$ref` cycles) works natively — it was only impossible in the rejected
  compile-to-CEL design.

### 8.1 Worked example (validates the seam, not a coverage target)

Against the 26 rules of the embedded showcase templates — which exist to demonstrate range,
not to prescribe — the seam rule sorts **17 to schema, 9 to CEL**, with zero ambiguity:
schema gets title/status/tags rules, heading presence (`contains` over materialized
`ast.headings`), code-block language, word-count floor, wikilink existence, and the
`if/then` status↔supersedes pair; CEL gets the state machine, immutability, sibling
comparison, the cross-subtree `supersedes`↔wikilink rule, and all four temporal lint rules.
Full table: `research/seam-core-pipeline.md` §2.

## 9. CEL coverage commitment (G2)

akb pins cel-go **v0.28.0**; latest is **v0.32.0** (2026-08-19). The four missed releases carry
features directly relevant to this design:

| Release | Relevant changes |
|---|---|
| v0.29.0 | JSON encoder ext library; `network.IP`/`CIDR` support upstreamed from Kubernetes; `has()` unknown-propagation fix; `lists.range` OOM guard |
| v0.30.0 | **`timestamp()` now rejects non-RFC3339 strings** (interacts with akb's date-only `2006-01-02` coercion — pre-conversion in `convertDateField` remains required); expression node limits for parser/checker; `cel.bind` nesting validation; cost-tracking accuracy fixes |
| v0.31.0 | Native Go type support moved into core (`cel.NativeTypes`); self-describing struct types; **regex program plan size controls** (DoS hardening for `matches`); env copy-on-write (near-zero cost for per-invocation env extension) |
| v0.32.0 | **BREAKING: module path → `cel.dev/cel-go`** (#1413); timestamp parsing helpers for varied formats (#1414 — could simplify `convertDateField`); `NativeToValue` Go-based JSON type support (#1402 — relevant to the document builder); aggregate size computations; JWT/HMAC libraries |

Upgrade to v0.32.x, then enable in `NewEnv` (`internal/cel/engine.go:25-37`):

| Extension | Unlocks | Example rule it enables |
|---|---|---|
| `ext.Strings()` | trim/split/join/replace/substring | normalized-title comparisons |
| `ext.Lists()` | distinct/sort/range/flatten | `uniqueItems`-style checks |
| `ext.Sets()` | contains/equivalent/intersects | tag-set algebra |
| `ext.Math()` | greatest/least/abs/ceil/round | thresholds over counts |
| `ext.Bindings()` | `cel.bind` | readable multi-step rules without repetition |

Deferred to ADR: `ext.Regex` (needs `cel.OptionalTypes()`), optional types themselves,
two-variable comprehensions, `ext.Encoders`/`ext.Native`. The cost limit (100k) stays; ext
functions are costed like stdlib ones. This is **additive**: every existing template keeps
working unchanged.

## 10. Overlap guidance (G6) — the SKILL decision rule

Many constraints are expressible in both layers. Agent guidance must be one screen, low token:

> **Which validator?**
> - Can the rule be checked by looking at **this page alone, right now** — a field's type,
>   presence, enum, pattern, length, list contents, or a derived view (headings, word count,
>   links, code blocks)? → **`schema:` block**.
> - Does it need the **previous version** (`old_page`), the **clock** (`now`), **arithmetic**,
>   or **comparing two values** against each other? → **CEL `validations:`** (write gate) or
>   **`lint_rules:`** (sweep-time drift policy).
> - Never state the same fact in both. Schema first; CEL for what schema can't see.

Placement: rewrite of `internal/skill/embedded/kb-management/references/TEMPLATE.md` (the
agent-facing anatomy + guard doctrine), a new structural-checker row in `MAINTAIN.md`, and
`APPROVE.md` wording. The embedded adr/note showcases get restructured to model the split
(adr keeps its convoluted CEL as the *showcase of the CEL half*).

## 11. CLI surface

| Change | Shape | Notes |
|---|---|---|
| `akb read --json` | `{file, frontmatter (raw), content:{raw}, is_draft}` | one file read + YAML parse; **no goldmark** |
| `akb read --json --derived` | adds `content.word_count/char_count`, `ast`, `akb` | the only way an agent can see what CEL sees; enables pre-flight/self-repair loops |
| `akb write --json` | stdin `{frontmatter, content.raw}` → full pipeline; stdout echoes the **effective document** (after `created`/`updated`/`is_draft` mutations) | derived keys in input = exit 2 |
| `akb write --json` failures | structured validation report on stdout (rule IDs + JSON Pointers), exit 1 | requires `runTemplateValidations` widening |
| `akb template get --json` | machine-readable Writer View incl. schema block | complements existing YAML views |
| `links`, `orphans`, `status`, `index show`, `log show`, `raw list` | `--json` additions (output-only) | opportunistic, independent of this RFC's core |

Round-trip invariant: `read --json | write --json` must be an effective no-op (modulo akb's
managed mutations). This is guaranteed by §5.1 (raw values cross the boundary, never the
CEL-coerced map) and is a pinned integration test.

## 12. Approve gate fix (pre-existing hole, closed by this design)

`akb approve` today checks required-field presence only and **never re-runs CEL validations** —
a CEL-invalid page can be published. Under coexistence, approve becomes: **full schema pass +
write-time `validations` re-run with `old_page` = on-disk page, `now` = approval time**;
`lint_rules` stay out (temporal policy, not a publish gate). Unknown-type handling unchanged.

## 13. Sweep design (lint)

- New `schema` checker: per-page structural validation, degrades per page like `cel_lint`
  (rescues pages that bypassed write — git pull, `index rebuild` ingestion, D11).
- Subsumes `required_fields` and `type_orphan` checkers (retire or thin-wrap → ADR).
- Ordering: schema checker before `cel_lint` so CEL never evaluates pages with known-missing
  required keys. Both passes share one `now` stamp.
- All issues flow into the existing `LintIssue`/`--json` envelope unchanged (N5).

## 14. Roadmap interactions (agent-memory/DESIGN.md)

| DESIGN.md item | Interaction |
|---|---|
| §4.3 typed gates (`akb lint --json` + gate schema) | Direct synergy — stable envelope is N5; this RFC completes the loop the design anticipated |
| T1 (enforce decorative `schema.frontmatter` type/enum) | **Subsumed** — native JSON Schema replaces the fix |
| §4.2 six content templates + bi-temporal fields | Authorship targets for the split; `valid_from/valid_to` comparisons are CEL-side (value-to-value) |
| §4.2 provenance classes (`source` enum, `confidence`) | `source` enum is schema-native; confidence-vs-marker ratio stays CEL (`provenance` checker) |
| D11 (index-rebuild ingestion bypass) | Makes the sweep-time schema checker load-bearing, not optional |
| §4.3 secret scanning | Unchanged — a third, separate write gate (neither schema nor CEL) |

## 15. Open questions → ADR / tech-spec candidates

1. **ADR: validator library choice** (santhosh v6 recommended) + `format` assertion policy.
2. **ADR: `schema.frontmatter` retirement mechanics** — clean break vs one-release alias;
   migration error text; in-place migration command (`akb template migrate`?) vs manual.
3. **ADR: cel-go upgrade + ext library set** — v0.28.0 → v0.32.x (module path migration
   `github.com/google/cel-go` → `cel.dev/cel-go`, all imports); final ext list;
   `ext.Regex`/optionals/two-var comprehensions; whether #1414 timestamp helpers replace part
   of `convertDateField`.
4. **Tech-spec: structured validation report** — widening `runTemplateValidations`, merged
   schema+CEL error model, stable IDs/JSON Pointers, stdout/stderr split.
5. **Tech-spec: document builder** — unify `page`/`document` behind typed structs (revive
   `internal/cel/types.go`), fix §5.3 hazards, export shared exclusion logic.
6. **ADR: derived-view versioning** — `document` gets a `schema_version`/akb version marker so
   agents can detect shape changes.
7. **ADR (optional, diagnostics-only): CEL→JSON-Schema portability report** — lossy subset
   extractor answering "which of my KB guarantees travel to a downstream harness?" (decree-style
   AST walker; never authoritative, never round-tripped).
8. **Tech-spec: approve gate** (§12) — re-validation scope, unknown-field policy on
   re-serialization.

## 16. Phasing (high level; sizing via tech-specs)

- **Phase 1 — foundations:** document builder + hazards (§5.3), validator integration, `schema`
  block in templates, pipeline order + merged errors, `template write` schema obligations.
- **Phase 2 — JSON I/O:** `read --json` (two tiers), `write --json`, round-trip invariant tests.
- **Phase 3 — lifecycle:** sweep schema checker, approve gate fix, `schema.frontmatter`
  retirement + migration path.
- **Phase 4 — doctrine:** SKILL rewrites, showcase restructure, overlap guidance, CEL ext
  enablement (can pull forward — it's independent and additive).

## 17. Risks

| Risk | Mitigation |
|---|---|
| Derived view becomes a public contract prematurely | §5.3 fixes gate Phase 1; `schema_version` marker (ADR 6) |
| `akb lint --json` shape churn breaks planned typed gates | N5 freeze; additive rows only |
| Agents duplicate facts across both layers | P3 + `template write` overlap warning + SKILL guidance |
| JSON I/O confused with a page file-format change | N2 stated everywhere; walkers untouched |
| Cost-limit regression from ext-enabled CEL or schema-first ordering | Schema-first ordering *reduces* CEL work; cost gate at `template write` via static estimate |
| `format: date-time` semantics diverge from `convertDateField` coercion | ADR 1 picks assertion policy; SKILL documents akb's date rules in one place |

## Appendix A — Research base

All in `.pi/subagents/proposals/json-cel/research/`: `LOG.md` (session log + decisions),
`akb-recon.md`, `schema-survey.md`, `cel-bridge-research.md`, `cel-bridge-research-b.md`,
`command-inventory.md`, `seam-core-pipeline.md`, `seam-periphery.md`.

## Appendix B — Glossary

- **document** — the materialized JSON view of a page (raw frontmatter values + derived views);
  the JSON Schema validation target and the `read --json` boundary object.
- **page** — the CEL evaluation map; same shape as `document` but with date coercion.
- **derived views** — `content.word_count/char_count`, `ast.*`, `akb.*`; pure functions of the
  markdown body, akb-authored, output-only on write.
- **seam rule** — P1: schema owns single-document/single-instant predicates; CEL owns the rest.
