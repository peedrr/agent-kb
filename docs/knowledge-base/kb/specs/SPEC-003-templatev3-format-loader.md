---
anchors:
- checked_at: "2026-10-10T17:14:18Z"
  claim: REQ-004
  evidence: at 5a2ea67 one of two divergent retired-shape detectors (capital/period message); the second lives in cmd/akb/templates_write.go (lowercase/semicolon); ADR-006 I3 requires convergence into one shared helper
  kind: local
  path: internal/template/template.go
  state: live
  symbol: detectOldFormat
  verify:
    check: go test ./internal/template/...
    method: check
- checked_at: "2026-10-10T17:14:18Z"
  claim: REQ-001
  evidence: at 5a2ea67 the loader is a non-strict double unmarshal (goccy/go-yaml, no Strict/DisallowUnknownField); unknown keys silently ignored
  kind: flow
  path: internal/template/template.go
  state: live
  symbol: LoadTemplates
  verify:
    check: go test ./internal/template/...
    method: check
- checked_at: "2026-10-10T17:14:18Z"
  claim: REQ-009
  evidence: at 5a2ea67 the key loop routes type/title into ParsedFrontmatter.Type/.Title and everything else into Fields
  kind: local
  path: internal/frontmatter/frontmatter.go
  state: live
  symbol: Parse
  verify:
    check: go test ./internal/frontmatter/...
    method: check
- checked_at: "2026-10-10T17:14:18Z"
  claim: REQ-010
  evidence: at 5a2ea67 BuildPage copies fm.Fields then force-injects fmMap["type"]/["title"] — the compatibility shim un-routing removes
  kind: local
  path: internal/cel/pagebuilder.go
  state: live
  symbol: BuildPage
  verify:
    check: go test ./internal/cel/...
    method: check
- checked_at: "2026-10-10T17:14:18Z"
  claim: REQ-018
  evidence: at 5a2ea67 loadExampleTemplates unmarshals embedded YAML directly into template.Template, bypassing LoadTemplates and all its checks
  kind: local
  path: cmd/akb/template.go
  state: live
  symbol: loadExampleTemplates
  verify:
    check: go test ./cmd/akb/...
    method: check
created: "2026-10-09"
grammar: 1
id: SPEC-003
is_draft: false
kind: change
provenance: agent-drafted
ratified_by:
- Pete Hope
scope:
- internal/template/**
- internal/frontmatter/**
- internal/cel/**
- internal/lint/**
- internal/index/**
- internal/search/**
- cmd/akb/**
- docs/knowledge-base/.agent-kb/templates/**
- internal/skill/embedded/**
- test/**
status: active
title: Templates Adopt a Strict Six-Key V3 Format With a JSON Schema Block
type: spec
updated: "2026-10-10T17:15:23Z"
verified:
  at: "2026-10-10T17:14:18Z"
  branch: main
  commit: 5a2ea674e91b1c976901ab6f16f4bcf9bad826bc
  method: manual-review
  next_review_by: "2027-01-09"
  state: live
---

# SPEC-003: Templates Adopt a Strict Six-Key V3 Format With a JSON Schema Block

> **For agents:** this SPEC is authority only while `status: active` AND
> `verified.state: live`. If your task conflicts with it, or any anchor fails
> to resolve at HEAD, stop and name the conflict. Do not silently work around
> it; propose an amendment or supersession instead.

Origin: [[RFC-002-open-validation-engine|RFC-002]] candidate RFC-002-S6 (RFC §6 carries the
V3 format, §17/§18 phase and govern this record's lifecycle), researched 2026-10-09 in the
research session ([[rfc-002-s6-log]]); the three scope decisions (six-template corpus, opaque
schema passthrough, format-only conversion) were ratified by Pete Hope 2026-10-09 via decision
form. Ratified 2026-10-10 by Pete Hope; per the RFC-002 §18 lifecycle the RFC's S6 row moves
to ratified as SPEC-003 and its body reference is rewritten on that ratification.

## Why

The template loader silently ignores misspelled and retired keys (plain unmarshal, no
strictness), and the presence schema lives in a homegrown `schema.frontmatter` block whose
`type`/`enum` fields are dead in non-test code — an unenforceable declaration is a trap.
[[ADR-010-templatev3-minimal-key-set|ADR-010]] ratified the Phase-1 answer: a strict, minimal
TemplateV3 with exactly six top-level keys and a JSON Schema 2020-12 `schema` block. This SPEC
is the executable boundary for that change; the schema *pass* that consumes the block is
SPEC-004, and `template write` schema obligations are SPEC-005.

## Goals and Non-Goals

Goals:
- Strict TemplateV3: six recognized top-level keys, refused at authoring, warned at load.
- One shared retired-shape refusal (ADR-006) with one self-sufficient message.
- D10 un-routing: `type`/`title` become ordinary frontmatter keys; `type` survives only as the
  template selector.
- Convert all six templates this repo ships/uses (2 embedded + 4 KB) to V3, format-only.

Non-goals:
- Compiling, caching, or evaluating the `schema` block (SPEC-004 / [[ADR-009-phase1-sweep-schema-checker|ADR-009]]).
- Mockup-vs-schema proofs, the `FAILS:` parser, overlap warnings (SPEC-005; ADR-006 N7).
- The `dir`→`path` rename, `state_field`, `title_field`, `functions`, `projections` (later phases; ADR-010 N1/N3).
- The document-shape flatten (`derived.*`, `word_count()` registry) and `write --json` ([[ADR-008-phase1-current-document-shape|ADR-008]]).
- A V2→V3 migration command or alias (ADR-006: none exists, none is added).

## Current Behaviour

At `verified.commit` the loader (`internal/template/template.go`, `LoadTemplates`) unmarshals
each `*.yaml` twice with goccy/go-yaml and no strict option, so unknown top-level keys and
unknown `schema.frontmatter` sub-keys are silently dropped. Two divergent `detectOldFormat`
copies (template.go; cmd/akb/templates_write.go) reject only the pre-V2 `required`/`optional`/
`body` shape, with different message spellings both pinned by `test/testdata/cel_old_format.txt`.
`frontmatter.Parse` routes `type`/`title` into dedicated struct fields; `BuildPage` re-injects
them into the CEL page map; presence helpers, mockup enumeration, re-serialization, search,
and `index.md` all read the dedicated fields. `loadExampleTemplates` bypasses `LoadTemplates`
entirely. `FieldSchema.Type`/`.Enum` have no non-test consumer.

## Requirements

REQ-001: The template loader MUST recognize exactly the top-level keys `name`, `description`,
`dir`, `schema`, `validations`, and `lint_rules`.

REQ-002: WHEN `akb template write` reads a template carrying any other top-level key, the
command MUST refuse with exit 1, and the refusal MUST name every offending key and the full
supported set.

REQ-003: WHEN the load path reads an otherwise-V3 template carrying an unrecognized top-level
key, the loader MUST warn on stderr naming each offending key and MUST continue loading the
recognized keys.

REQ-004: WHEN a template file carries the retired `schema.frontmatter` shape, the load path
and `akb template write` MUST refuse it through one shared detection helper emitting one
message text, and the refusal MUST apply before any unknown-key judgment.

REQ-005: The `schema` value MUST be a JSON Schema 2020-12 document whose paths name the
current document shape (`frontmatter.*`, `content.raw`, `ast.headings`, `akb.*`, `file.*`).

REQ-006: WHEN a template lacks a `schema` mapping, the loader MUST refuse it as it refuses a
missing `name`.

REQ-007: The Phase-1 loader MUST NOT compile or evaluate the `schema` block; it MUST preserve
the parsed payload uncompiled for the schema pass (SPEC-004).

REQ-008: In every shipped template, the schema `required` set under `properties.frontmatter`
MUST equal the set of frontmatter keys the template's CEL rules read unguarded.

REQ-009: `frontmatter.Parse` MUST place every frontmatter key, including `type` and `title`,
into the generic fields map; it MUST NOT route individual keys into dedicated struct fields.

REQ-010: The CEL page builder MUST NOT re-inject `type` or `title` into the page map; it MUST
expose each exactly once, sourced from the parsed fields.

REQ-011: WHEN a page's frontmatter `type` names no loaded template, `akb write` and
`akb append` MUST refuse the page; `type` MUST NOT gate anything else.

REQ-012: The display field for the search title column, `index.md`, and the log MUST be the
literal frontmatter `title` key; Phase 1 ships no `title_field` machinery.

REQ-013: `dir` MUST keep its current placement semantics; the loader, embedded templates, and
KB templates MUST NOT rename it to `path`, which arrives with Phase 3 (D11).

REQ-014: The embedded `adr`/`note` templates and this KB's `adr`/`rfc`/`spec`/`research`
templates MUST ship in V3 form in the same change as the loader.

REQ-015: WHEN converting a template to V3, the conversion MUST move every `required: true`
declaration into the schema `required` set, MUST move `enum` declarations into schema `enum`,
and MUST NOT redesign any CEL rule — rule texts move verbatim.

REQ-016: WHEN a converted template's rules call `timestamp()` on a field, the schema MUST
declare that field with the `format` (`date` or `date-time`) matching the values the existing
corpus actually carries.

REQ-017: Unknown-key detection on both paths MUST use one shared map-diff over the parsed YAML
AST that reports every offending key with its line and column; the implementation MUST NOT rely
on `yaml.Strict()` alone, which reports one nondeterministic key and cannot aggregate.

REQ-018: The `--examples` template source MUST load through the same V3 machinery as on-disk
templates; it MUST NOT bypass the format checks.

## Acceptance Criteria

AC-001 (verifies REQ-001): WHEN the shipped corpus is loaded, every template carries only keys
from the six-key set.
  verify: { method: check, check: "go test ./internal/template/..." }

AC-002 (verifies REQ-002): WHEN `template write` is run against fixtures carrying each deferred
key (`path`, `state_field`, `functions`, `projections`, `title_field`, `summary_field`,
`strip_on_approve`) plus a typo key, each refusal exits 1 and names the offending key(s) and
all six supported keys.
  verify: { method: check, check: "go test ./test/ (new testscript: template_v3_strict_keys)" }

AC-003 (verifies REQ-003): WHEN a page is written against an otherwise-V3 template carrying an
unknown top-level key, the write succeeds and stderr names the unknown key.
  verify: { method: check, check: "go test ./test/ (new testscript: template_v3_strict_keys)" }

AC-004 (verifies REQ-004): WHEN a retired-shape template is loaded or written, both paths emit
the identical message, and a mixed retired/V3 file draws the retired-shape refusal, never the
unknown-key refusal.
  verify: { method: check, check: "go test ./internal/template/... ./cmd/akb/... ./test/" }

AC-005 (verifies REQ-005): WHEN the shipped V3 schemas are inspected, every `properties` path
names a current-shape location and no `derived.*` path appears.
  verify: { method: check, check: "rg -n 'derived\\.' internal/template/embedded docs/knowledge-base/.agent-kb/templates; expect 0 hits" }

AC-006 (verifies REQ-006): WHEN a template omits `schema`, load and `template write` refuse it
naming the missing key.
  verify: { method: check, check: "go test ./internal/template/... (new case)" }

AC-007 (verifies REQ-007): WHEN the Phase-1 diff is inspected, `internal/template/` imports no
JSON Schema compiler; `go.mod` gains no validator dependency.
  verify: { method: check, check: "go list -deps ./internal/template | rg jsonschema; expect 0 hits" }

AC-008 (verifies REQ-008): WHEN a shipped template's CEL rule reads a frontmatter key without a
`has()` guard, that key appears in the schema `required` set, and conversely.
  verify: { method: ask, ask: "audit the six converted templates rule-by-rule against the guard invariant (RFC-002 §6)" }

AC-009 (verifies REQ-009): WHEN `frontmatter.Parse` parses `type: adr` and `title: X`, both
appear in the fields map and no dedicated struct field is populated.
  verify: { method: check, check: "go test ./internal/frontmatter/..." }

AC-010 (verifies REQ-010): WHEN `BuildPage` builds the page map, `frontmatter.type` and
`frontmatter.title` appear exactly once with the parsed values.
  verify: { method: check, check: "go test ./internal/cel/..." }

AC-011 (verifies REQ-011): WHEN a page declares an unknown `type`, write refuses naming the
type; a page of known `type` with unusual `title` values behaves as the template's rules dictate.
  verify: { method: check, check: "go test ./cmd/akb/... ./test/" }

AC-012 (verifies REQ-012): WHEN a page carries `title: X`, the search title column, index
entry, and log entry render `X`, with no `title_field` key consulted.
  verify: { method: check, check: "go test ./test/ (existing workflow scripts, updated)" }

AC-013 (verifies REQ-013): WHEN the corpus and loader are inspected, no `path:` template key
exists and `dir` placement behaviour is unchanged.
  verify: { method: check, check: "rg -n '^path:' internal/template/embedded docs/knowledge-base/.agent-kb/templates; expect 0 hits" }

AC-014 (verifies REQ-014): WHEN the change lands, all six templates parse as V3 and their
pass/fail mockups keep their prior verdicts (each fail mockup fails exactly its original rule).
  verify: { method: check, check: "go test ./internal/template/... ./cmd/akb/..." }

AC-015 (verifies REQ-015): WHEN a converted template is diffed against its V2 form, CEL rule
texts are byte-identical and presence/enum declarations live only in the schema block.
  verify: { method: ask, ask: "diff review of the six template conversions" }

AC-016 (verifies REQ-016): WHEN `created`/`updated` formats are inspected in converted
templates, date-only fields declare `format: date` and RFC3339 fields `format: date-time`.
  verify: { method: ask, ask: "human review of each format declaration against corpus values (ADR-006 N3: author judgment)" }

AC-017 (verifies REQ-017): WHEN a fixture carries two unknown top-level keys, the write
refusal lists both with line and column, deterministically across runs.
  verify: { method: check, check: "go test ./internal/template/... (new aggregation case)" }

AC-018 (verifies REQ-018): WHEN `template list/get --examples` runs, it exercises the same V3
format checks as on-disk loading.
  verify: { method: check, check: "go test ./cmd/akb/... ./test/ (template_examples.txt, updated)" }

AC-019 (verifies REQ-002): WHEN the refusal message is read cold, it states what was found,
the supported set, and where the format is documented, without consulting other output.
  verify: { method: ask, ask: "human review of message self-sufficiency (ADR-010 human-only residue)" }

## Contract and Invariants

- The loader MUST NOT recognize-then-ignore a deferred key (`path`, `state_field`,
  `functions`, `projections`, `title_field`, `summary_field`, `strip_on_approve`) on any
  Phase-1 path; each is refused at authoring and warned at load like any unknown key.
- The format MUST NOT gain a `version:` discriminator key; V3 is told apart structurally —
  retired shape refused first, six-key posture otherwise — keeping the recognized set at six.
- Unknown keywords *inside* the `schema` block are not policed in Phase 1; JSON Schema is
  vocabulary-open, and any closure rides SPEC-004's compiler integration.
- Refusal and warning texts SHOULD be single-sourced constants shared by both paths.
- `frontmatter.Parse` MUST NOT retain the `Type`/`Title` struct fields as a compatibility
  shim; downstream consumers read the fields map, so stale dual sources cannot diverge.

## Preserved Behaviour

- Unknown-`type` refusal text and the write/append routing behavior stay as today (P7).
- CEL `validations` (write-time, fail closed) and `lint_rules` (sweep-time) semantics are
  unchanged; rule evaluation order and error aggregation are untouched.
- `template get --full` continues to dump the raw template file bytes; the writer view keeps
  its `{name, description, schema}` shape with the V3 schema block flowing through.
- Duplicate-name and missing-name load checks, `CopyDefaults` semantics, and the
  `akb: `-prefixed commit behavior are unchanged.
- For well-formed pages, write/append/search/index/log observable behavior is unchanged;
  un-routing is representation-internal.

## Decisions and Rejected Alternatives

The conversion corpus is all six templates (2 embedded + 4 KB), owner-ratified 2026-10-09,
because the strict loader refuses the V2 shape and the repo's own KB must never be unloadable
at HEAD. Rejected: embedded-only conversion. Do not re-propose unless the loader gains a
V2-tolerance mode, which ADR-006 forbids.
The loader hands `schema` through uncompiled, owner-ratified 2026-10-09, keeping all santhosh
v6 coupling inside SPEC-004's compiled-schema cache design. Rejected: compile-at-load. Do not
re-propose unless SPEC-004's design proves pass-through unworkable.
The conversion is format-only, owner-ratified 2026-10-09; RFC §17's embedded-example
restructure stays in Phase 4. Rejected: showcase rewrite in Phase 1. Do not re-propose unless
the schema pass (SPEC-004) has landed.
Unknown-key detection is one shared AST map-diff helper. Rejected: `yaml.Strict()` alone —
goccy v1.19.2 reports a single nondeterministic key and panics on structs with inline maps
(upstream issue #810, fix PR #849 unmerged). Do not re-propose unless goccy ships a collecting
strict mode.
`title` stays a literal default key in Phase 1. Rejected: shipping `title_field` early — a
deferred key under ADR-010 N1. Do not re-propose unless Phase 3 scope moves into Phase 1.

## Assumptions and Open Questions

- [ASSUMPTION: goccy v1.19.2's `parser`/`ast` packages expose per-key line:column positions
  sufficient for REQ-017 — verified from pinned source, not executed; the implementation
  smoke-tests the helper before wiring it in.]
- [ASSUMPTION: un-routing lands as one cohort across all consumer sites; the build plus the
  pinned tests fail loudly at any missed site.]
- Open question (non-blocking): exact wording of the REQ-002 refusal and REQ-003 warning is
  pinned by the implementation and human-reviewed (ADR-010 records wording as human-only
  residue; AC-019 gates it).

## Affected Surface and Ordering

- `internal/template/template.go` — `LoadTemplates`, `detectOldFormat` (converges), `Template`
  struct (`Schema` re-homed), `CopyDefaults`.
- `cmd/akb/templates_write.go` — second `detectOldFormat` (deleted), authoring-time strict
  check; `cmd/akb/template.go` — `loadExampleTemplates` routes through the loader.
- `internal/frontmatter/frontmatter.go` — `Parse`, `ParsedFrontmatter`, `ValidateType`,
  `ValidateTitle`; `internal/cel/pagebuilder.go` — `BuildPage`/`BuildOldPage`.
- Presence helpers (`cmd/akb/required_fields.go`, `internal/lint/required_fields.go`), mockup
  enumeration (`cmd/akb/templates_write.go`), re-serialization (`cmd/akb/write.go`,
  `cmd/akb/append.go`), title readers (`internal/index`, `internal/search`, `cmd/akb/index.go`).
- Corpus: `internal/template/embedded/{adr,note}.yaml` (+4 mockups),
  `docs/knowledge-base/.agent-kb/templates/{adr,rfc,spec,research}.yaml` (+8 mockups) — this
  page's own template converts in this wave.
- Doctrine (RFC §17 standing instruction): AGENTS.md TemplateV2 claims;
  `internal/skill/embedded/kb-management/references/TEMPLATE.md` (documents the V2 format).
- Tests: `internal/template/template_test.go` (embedded-count and enum pins),
  `cmd/akb/{template,templates_write,write}_test.go`, `internal/lint/*_test.go`,
  `internal/frontmatter/frontmatter_test.go`; testscript: `cel_old_format.txt` (message
  spellings), `template_examples.txt`, `template_write_noop.txt`, `write_errors.txt`,
  `exit_codes.txt`, `full_workflow.txt`, `cel_validation.txt`, `cel_lint*.txt`,
  `lint_required_fields.txt`, `lint_tests_adr_write.txt`, `superseded_adr.txt`.

depends_on: [ADR-006, ADR-008, ADR-010]
blocks: [SPEC-004, SPEC-005]

## Verification Plan

- `just build && just test` and `go test ./test/ -test.v` green, including the new
  `template_v3_strict_keys` testscript matrix and every updated pin listed above.
- Zero-hit greps: `rg -n '^path:' <corpus>`; `rg -n 'schema:'`-adjacent `frontmatter:` retired
  shape; `rg -n 'derived\.' <corpus>`.
- `go list -deps ./internal/template` shows no JSON Schema compiler.
- Human-only residue: refusal/warning message self-sufficiency (AC-019); per-template
  `format: date|date-time` declarations (AC-016); guard-invariant audit (AC-008); conversion
  diff review (AC-015).

## Drift Ledger

- 2026-10-09 · all · — → unverified · authored from recon at 5a2ea67 · evidence: research
  session [[rfc-002-s6-log]] and its filed reports.
- 2026-10-10 · all anchors · unverified → live · human ratification (Pete Hope) after
  re-resolution of every anchor symbol at 5a2ea67 · evidence: symbol resolution of
  detectOldFormat, LoadTemplates, Parse, BuildPage, loadExampleTemplates.

## Revisit Triggers

- `go.mod` — goccy/go-yaml version change (REQ-017's mechanism is version-pinned).
- `internal/template/**`, `docs/knowledge-base/.agent-kb/templates/**` — format or corpus drift.
- `docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md` — §6/§17 amendment.
- SPEC-004 ratification — re-check REQ-005/REQ-007 boundary text.

## Evidence Appendix

Forensics at `5a2ea67` (line numbers valid at that commit only):
- `internal/template/template.go:106,114` — non-strict double unmarshal; `:73-80` —
  `detectOldFormat` copy A; `:43-51` — six-key struct shape already present.
- `cmd/akb/templates_write.go:461-469` — `detectOldFormat` copy B (divergent message).
- `internal/frontmatter/frontmatter.go:61-72` — type/title routing switch;
  `internal/cel/pagebuilder.go:60-66` — re-injection shim; `cmd/akb/template.go:104-127` —
  `--examples` bypass; `internal/template/template_test.go:378-380` — embedded-count pin of 2.
- Library feasibility (goccy v1.19.2: no warn mode, single-key strict error, inline-map panic;
  santhosh v6.0.3: compile API, non-thread-safe compile, no YAML entry point) is filed as
  [[rfc-002-s6-gapfill-libs]] with upstream sources; code state as [[rfc-002-s6-code-recon]];
  RFC/ADR requirements as [[rfc-002-s6-rfc-extract]]; prior-research deltas as
  [[rfc-002-s6-prior-research]].
- Governing decisions: [[ADR-006-schema-frontmatter-retirement|ADR-006]],
  [[ADR-008-phase1-current-document-shape|ADR-008]], ADR-009, ADR-010; origin
  RFC-002 §6/§17/§18.

Discard log:
- pi-subagents' TypeBox validator findings (`schema-survey` research) — bridge-side, not akb;
  akb's validator is santhosh v6 per ADR-003. Not imported.
- `rfc-templates-*` research run — concerns the RFC process document format, not akb page
  templates. Not imported.
- `yaml.Strict()` as the sole strictness mechanism — rejected (see Decisions).
- RFC §6's flattened-shape schema example — superseded by ADR-008's current-shape bound; not
  copied into this SPEC.
