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
summary: "Run log for RFC-002-A1: closed validator-library and format-assertion questions, owner-ratified forks, and the decision trail behind ADR-003."
tags:
- research-log
- json-schema
- validation
- rfc-002
title: RFC-002-A1 Validator Research Run Log
type: research
updated: "2026-10-06"
---
# Research run: RFC-002-A1 — Validator library + `format` assertion policy

- **Started:** 2026-10-06T15:18Z
- **Mode:** start
- **Coordinator:** pi session (RESEARCH-THIS protocol)
- **Instructions (verbatim):** `@docs/knowledge-base/kb/rfcs/RFC-002-open-validation-engine.md **in full** and begin work on RFC-002-A1. Use subagents to research as per your protocol. If all auto-resolvable, create the required ADR using `akb` (in PATH, AKB_KB is set to correct KB). If questions surface requiring my input and ratification: owner-escalation (SKILL)`

## Question

Produce the ADR resolving RFC-002-A1 (per RFC-002 §18 lifecycle: draft as real ADR
with next free global number, header records `Origin: RFC-002 candidate RFC-002-A1`).
Two decisions:

1. **Validator library** — RFC-002 §9 recommends `github.com/santhosh-tekuri/jsonschema/v6`
   (full 2020-12, Go-native); final choice is A1's. Must verify: 2020-12 keyword
   coverage (`contains`/`minContains`, `if/then/else`, `$ref`/`$defs` incl. cycles,
   `dependentRequired`/`dependentSchemas`, `unevaluated*`), format-assertion API
   (can built-in format checkers be registered/overridden? — §5.2 requires
   overriding `date-time` to Go-strict RFC3339), maintenance status, performance,
   module path/license. RE2 pattern semantics (Go ⇒ RE2, ECMA-262/RE2 dialect
   concern dissolved).
2. **`format` assertion policy** — RFC-002 §9: assertion enabled at least for
   declared temporal fields (the §5.2 date-bridge precondition guarantee depends
   on it); default for other formats decided by A1. §5.2 additionally pins via
   A1: Go-strict RFC3339 profile for `date-time` (uppercase `T`, numeric offset
   or uppercase `Z`, no leap seconds — RFC 3339 permits lowercase t/z + leap
   seconds; Go `time.RFC3339` / cel-go `timestamp()` reject both — claimed
   "verified" in RFC-002, re-verify); `format: date` coerces to midnight UTC
   in-memory; `format: time` validates as string-only, never coerced.

## Agreed requirements

- Deliverable: ADR in the KB (`AKB_KB=docs/knowledge-base`, template `adr`),
  created via `akb write`, next free global ADR number, `Origin: RFC-002
  candidate RFC-002-A1`.
- Every claim anchored. External claims: recent + authoritative sources or
  convergence of fact; no training-data reliance for ecosystem state.
- Decision forks: auto-resolve if one right way / low stakes; else
  owner-escalation per skill (`global-harness/skills/owner-escalation`,
  `decision-forms` craft).

## Decisions made

- **D-A1: library = `github.com/santhosh-tekuri/jsonschema/v6`.** Auto-resolved: only
  library with verified full 2020-12 (100% bowtie badge), overridable built-in format
  checkers (runtime-verified), active maintenance, Apache-2.0, RE2 default. All
  alternatives disqualified (see findings). kaptinlin/jsonschema recorded as unverified
  challenger — do-not-revisit-unless condition goes in the ADR.
- **D-A2: `date-time` profile pinned to the strict pattern, NOT raw Go acceptance.**
  Probe (2026-10-06, Go 1.x, /tmp/rfc3339-probe): Go `time.Parse(time.RFC3339)` rejects
  lowercase t/z + leap seconds (RFC-002 §5.2 verified) but ACCEPTS comma fractions,
  single-digit hours, offsets +24:00/+03:60. cel-go ≥v0.30 `strictRFC3339Pattern`
  rejects those; akb pins cel-go v0.28.0 today (no gate). Profile = strict pattern
  (uppercase T; Z or ±hh:mm, hh≤23 mm≤59; seconds ≤59; optional dot-fraction;
  calendar-validated) — subset of BOTH cel-go v0.28 and v0.32 acceptance sets, so the
  §5.2 precondition guarantee survives the A3 upgrade. Refines RFC-002 §5.2 wording
  ("Go-strict") with verified evidence; intent ("assertion = CEL acceptance set")
  preserved.
- **D-A3: akb registers own checkers for `date-time` and `date`** (coercion-relevant);
  `time` keeps santhosh built-in (never coerced to CEL; no parse precondition).
- **D-A4: unknown format names → `template write`-time warning** against the known
  registry (assertions make an unknown format a silent no-op = trap); load stays
  lenient (forward-compat precedent, template.go:106-114).
- **D-A5: register formats before Compile; one compiled schema cached per template**
  (Compiler.Compile ≈2.8ms maintainer-measured; compilation not thread-safe — akb is
  a serial CLI, non-issue; concurrent Validate on one schema unverified, akb validates
  serially).

## Key findings

Full reports: `extract-prior-research.md`, `recon-code.md`, `research-web.md` (this dir).

**Library (research-web.md Q1/Q2/Q4):**
- santhosh v6.0.3 (2026-06-28), pushed 2026-09-21, Apache-2.0, Go 1.21 — maintained,
  single-maintainer, steady cadence. bowtie badge: **100% Passing draft 2020-12**
  (badge JSON fetched). Cycle detection explicit in validator.go.
- Alternatives: qri-io (draft 7/2019-09, dead since 2021), xeipuuv (unmaintained 2020,
  draft ≤7), invopop (generation only), google/jsonschema-go (2020-12 but **ignores
  `format` entirely** — disqualifying), kaptinlin (2020-12, RegisterFormat exists but
  no bowtie evidence; built-in-name override unverified).

**Format API (research-web.md Q3 + runtime probe /tmp/santhosh-probe, 2026-10-06):**
- `Compiler.AssertFormat()` enables assertions for 2020-12 (default: annotation-only).
- `Compiler.RegisterFormat(&Format{Name:"date-time", Validate: f})` SHADOWS the
  built-in (compiler-registered wins; only `regex` un-overridable). **Verified
  end-to-end**: overridden checker rejected lowercase t/z + leap second; built-in
  accepted all three. Register BEFORE Compile (compile-time resolution).
- santhosh built-in `date-time` accepts lowercase t, lowercase z, leap second 23:59:60
  → override is NECESSARY, not optional (research-web.md Q3c, format.go quoted).

**RFC 3339 facts (research-web.md Q6 + probe):**
- RFC 3339 §5.6 NOTE: lowercase t/z permitted by ABNF; consuming spec "MAY further
  limit" to uppercase — akb's narrowing is licensed by the RFC itself.
- Leap seconds: RFC permits :60 at Jun/Dec endpoints; Go rejects ("second out of
  range"); cel-go rejects via the Go step.
- cel-go `timestamp()`: `time.Parse(time.RFC3339)` gated by `strictRFC3339Pattern`
  since v0.30.0 (PR #1338); pattern allows [Tt]/[Zz]/:60 but Go parse then rejects
  lowercase + leap. akb pins v0.28.0 (go.mod:12) = NO gate = laxer acceptance.
- JSON Schema 2020-12 §7.2.1: format assertion "MUST be disabled by default",
  implementations MAY enable with option; §7.2.2: Format-Assertion vocabulary optional.

**Integration (recon-code.md):**
- Schema pass slots between required gate (cmd/akb/write.go:523) and CEL
  (write.go:529); validates RAW document view (no date coercion).
- `schema:` key collision: current `schema.frontmatter` homegrown block — V3 repurposes
  with hard-reject migration (precedent detectOldFormat, template.go:73-82).
- frontmatter.Parse splits type/title out of Fields (frontmatter.go:79-93) — D10
  re-merge needed for the document view.
- Current coercion: name-agnostic, top-level only (pagebuilder.go:60-63); acceptance
  = time.RFC3339 then "2006-01-02" (pagebuilder.go:36,41).
- go.mod has NO JSON Schema dep today (verified); cel-go v0.28.0.

## Open questions

- ~~OQ1~~ CLOSED: yes — RegisterFormat shadows built-ins; verified source + runtime.
- ~~OQ2~~ CLOSED — owner ratified F1 in conversation 2026-10-06: **assert everything
  declared** ("Template writers own their mistakes"). Owner also confirmed the D-A2
  reading (strict pattern ⇒ cel-go unpin/upgrade path stays safe) and directed the
  RFC-002 amendment to follow ADR-003's outcome.
- ~~OQ3~~ CLOSED: no verified better option; kaptinlin only challenger, unverified.

## Parked/deferred

- Remaining orphans ADR-001 + RFC-003 — out of scope for this run; clear them when a
  related doc lands or via a dedicated linking pass.
- kb-management skill (ships with akb): add NON-prescriptive linking guidance
  (wikilink forms, 3-step resolution, dedupe-by-target, what orphan/broken-link
  lint measures) so agent-users can choose a style — owner-directed, rides the
  RFC-002 Phase 4 doctrine rewrite (§17), not a piecemeal edit now.
- Sweep-time `format: time` uses the santhosh built-in (accepts lowercase z, :60) —
  deliberate, since time values are never coerced; noted in ADR-003 I6.

## Child-run index

| Child | Task | Output |
|---|---|---|
| extract-prior | Extract JSON-Schema-library findings from RFC-001-era research base | `.pi/research/rfc-002-a1-validator/extract-prior-research.md` |
| recon-code | Map current akb date-coercion + validation integration points | `.pi/research/rfc-002-a1-validator/recon-code.md` |
| research-web | Current ecosystem state: santhosh v6 + alternatives, format APIs, RFC3339 facts | `.pi/research/rfc-002-a1-validator/research-web.md` |
| coordinator probes | Go time.RFC3339 acceptance matrix; santhosh RegisterFormat override end-to-end | /tmp/rfc3339-probe, /tmp/santhosh-probe (ephemeral); results inlined in D-A2 + findings |

## Final artifact — RUN COMPLETE 2026-10-06

- **ADR-003 `accepted`** (deciders: [Pete Hope]) →
  `docs/knowledge-base/kb/decisions/ADR-003-santhosh-v6-format-assertion.md`.
  Origin: RFC-002 candidate RFC-002-A1.
- **RFC-002 amended per §18**: row A1 → ratified/ADR-003 (linked), §5.2 profile
  pinned to ADR-003 I4 strict pattern + tightening note, §9 format policy resolved
  (assert-all), Supersedes row links RFC-001. Stays `draft` until all rows resolve.
- **Linking convention established** (owner chose harness-level record over a
  project-local ADR-004): canonical homes always link; first prose mention links,
  later plain; form = `[[full-page-name|Short ID]]`; quote zones + frontmatter
  exempt. Recorded in global docs-writer skill — mechanics in SKILL.md router,
  per-type homes in references/{adr,rfc,spec}.md — committed pi-meta-config
  `f4086d3`.
- Verified post-state: `akb lint` orphans 5 → 2 (ADR-001, RFC-003 remain; parked);
  `akb links` shows both directions resolving, no broken/ambiguous.
- Convention nuance for future linkers: repeat edges add no graph value
  (linkgraph dedupes by target, internal/linkgraph/sqlite.go:78); explicit-path
  form `[[display]](dest)` reserved for basename collisions; heading anchors only
  for stable sections.
