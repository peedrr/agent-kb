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
- internal/**
- cmd/akb/**
status: final
summary: "The run log and hub page for the json-cel research run: sessions 1-5, commissioned reports, owner decisions, and the RFC-001 to RFC-002 trajectory."
tags:
- run-log
- json-cel
- json-schema
- cel
title: Research run log — JSON Schema/CEL bridge (json-cel)
type: research
updated: "2026-09-27"
---
# JSON ⇄ CEL Bridge — Research Log

**Proposal:** side-by-side coexistence of JSON Schema (structure) and CEL (context/arithmetic)
validation in akb, plus JSON I/O flags. Owner: pete. Investigation: pi session + subagents.
**Status:** direction approved (coexistence); seam/boundary mapping in progress.

---

## 2026-09-26 — Session 1: Feasibility investigation

### Framing (owner's four questions)

1. Use a subagent's pre-defined `outputSchema` (JSON Schema) to define required CEL
   validation rules of a new akb template (JSON → CEL, scripted or built into akb).
2. Reverse: pre-defined CEL in a template → `outputSchema` requirements in agent definitions.
3. `akb read --json` output flag.
4. `akb write --json` — pipe structuredOutput into akb, passing CEL validation.

### Research commissioned

| Report | Agent | Contents |
|---|---|---|
| `akb-recon.md` | scout | akb internals: TemplateV2 format, CEL env, write/read/template-write flows, JSON precedents, constraint inventory, per-feature feasibility ratings |
| `schema-survey.md` | scout | `outputSchema` usage across pi-meta-config agents (4 of 7), pi-subagents validator (TypeBox, draft 3→2020-12), typed gates impl, feature inventory, gap analysis |
| `cel-bridge-research.md` | researcher (glm-5.3-flash) | External: no JSON-Schema→CEL compiler exists; K8s/protovalidate coexistence pattern; impedance map; cost-model analysis |
| `cel-bridge-research-b.md` | researcher (deepseek-v4.1-flash:high, duplicate per owner) | De-duped: adds decree CEL→schema AST extractor prior art, KEP-5073 "tags first, CEL escape hatch", CEL macro error-absorption semantics, cel2sql precedent |

### Key findings (session 1)

- akb `schema.frontmatter` type/enum fields are **dead** (only `required` enforced);
  owner confirms: known, fix in the works.
- akb CEL env = bare stdlib; cel-go ext libs (`ext.Strings/Lists/Sets/Math/Regex`) available
  but not enabled (one-line change each).
- No JSON-Schema→CEL compiler exists anywhere; ecosystem consensus (K8s, protovalidate,
  flux-schema, Datree) is **coexistence, not compilation**.
- CEL ⊃ JSON Schema validation vocabulary, except: recursion (`$ref` cycles) and
  `unevaluatedProperties`-style bookkeeping.
- Permanent fault lines for *translation* designs: regex dialect (ECMA-262 vs RE2),
  recursion, null-vs-absent, numeric model (int64/float64), error granularity.
  Permanent boundary for *JSON-only* designs: no arithmetic, no relative time, no
  cross-field comparison → `now`/`old_page`/AST rules are JSON-inexpressible by construction.
- Feature ratings: read --json **easy** (67-line passthrough + envelope); write --json
  **moderate** (synthesize frontmatter+body at write.go:266, reuse pipeline; must widen
  `runTemplateValidations` for structured failure output); schema→CEL **moderate** (hand-rolled
  subset generator, mockup synthesis is the real work); CEL→schema **lossy subset extractor,
  diagnostics-only** (decree prior art; akb's `CompileRule` discards the AST — needs new
  AST-returning entry point).

### Decision (owner, 2026-09-26)

**Sold on side-by-side coexistence.** Native JSON Schema for structure + CEL for
context/arithmetic. No translation layer as the primary mechanism.

### Open concerns carried into session 2

1. "JSON for structure, CEL for arithmetic" may be an over-simplification — the seam must be
   mapped against *everything akb does or plans to do* (roadmap:
   `/home/pete/code/projects/tools/pi/agent-memory/DESIGN.md`).
2. **Structured JSON I/O = metadata + content combined.** JSON mode must not be
   metadata-only; the JSON view of a page must represent the markdown content too (otherwise
   CEL content checks force metadata duplication — rejected as "crazy restriction").
3. Need per-command / per-subsystem boundary map: what each does, what JSON covers, what CEL
   covers, what neither covers.

---

## 2026-09-26 — Session 2: Seam/boundary mapping (in progress)

### Research commissioned

| Report | Agent | Contents |
|---|---|---|
| `command-inventory.md` | scout | Full inventory: 31 leaf commands, flags, I/O, subsystem map, data flows, DESIGN.md roadmap extraction, preliminary seam sketch |
| `seam-core-pipeline.md` | scout | Page model (every key + derivation), rule-by-rule expressibility table over all 26 shipped rules, body-parser purity, write --json assembly designs A/B, read --json shape, template format implications, SEAM VERDICT |
| `seam-periphery.md` | scout | Per-subsystem verdicts: search, linkgraph, all 10 lint checkers, index/log, raw/manifest, git, approve, config, kb-management skill doctrine; BOUNDARY TABLE + residual risks |

### Findings

**The seam rule (from seam-core-pipeline):** a rule is JSON Schema's iff its verdict is a pure
function of **one document at one instant**. `old_page` and `now` are the only non-page CEL
variables (`engine.go:31-35`) — everything needing either is CEL's, permanently.

**Empirical tally over the 26 shipped rules (adr + note templates):**
- 17/26 expressible in JSON Schema — but 8 of those 17 **only** if akb materializes derived
  views (`ast.headings/links/code_blocks`, `content.word_count`, `akb.*`) into the JSON document.
- 9/26 forever CEL: old_page state machine (#14) + immutability (#15), sibling-value comparison
  (#16), `now` arithmetic (#17, #22-25), cross-subtree existential (#21).

**Owner's "combined" constraint resolves as:**
- The validated JSON document MUST carry akb-materialized derived views; JSON Schema-authored
  rules may address derived paths but never `content.raw` text.
- On write: `content.raw` (markdown string) is the ONLY content input; derived keys are
  output-only and supplying them is a hard error. Design A (raw md in JSON) accepted;
  Design B (structured content → akb generates markdown) REJECTED (ast view is lossy by
  construction — paragraphs/lists/tables absent).
- `page` ≠ `document`: CEL map coerces date-like strings to `time.Time`; the JSON view must
  carry raw YAML values or `read --json | write --json` silently mutates dates.
- JSON→YAML assembly must use goccy `yaml.JSONToYAML` (or typed decode) — `encoding/json` →
  `yaml.Marshal` produces `count: 3.0` corruption.

**Template format shape:** new `schema.json` block (frontmatter structure + derived paths,
write-time before CEL + sweep-time), `validations[]` stays CEL (old_page/now/cross-doc),
`lint_rules[]` stays CEL (temporal), **`schema.frontmatter` retired** (its presence job moves
to JSON Schema `required`; invariant: the required set = exactly the keys CEL may read
unguarded). Dialect must be pinned (recommend 2020-12). No JSON Schema library in go.mod yet.

**Periphery:** search / linkgraph / index / log / raw / git = **neither** (downstream consumers
or exempt walkers, not validators). Lint splits: `required_fields` + `type_orphan` collapse
into a new schema checker; `provenance` + `cel_lint` stay CEL; sweep becomes two passes
(schema structural prefilter → cel_lint policy) sharing one `now` stamp. **Approve gap found:
approve never re-runs CEL validations today** — a CEL-invalid page can be published;
recommendation: approve = full schema pass + write-time validations with old_page/now.

**Roadmap (DESIGN.md) interactions:** JSON/JSON-Schema direction is NEW relative to the
ratified design (no mention). Strong synergies: DESIGN §4.3 typed gates already plan
`gate: {command: "akb lint --json", output: "json", schema}`; T1 (enforce decorative
type/enum) is subsumed by native JSON Schema; the six planned content templates are the
authorship target; D11 index-rebuild ingestion bypass makes the sweep-time schema checker
load-bearing. Risks: `akb lint --json` shape is gate-binding-stable; all walkers filter `.md`
(JSON I/O flags ≠ page file format change).

**kb-management skill doctrine requiring rewrite** (per seam-periphery §9):
TEMPLATE.md:60 ("Presence is all the schema enforces" — directly contradicted),
TEMPLATE.md:34-58 (two-layer anatomy), :102-119 (has() guard doctrine narrows to optional
keys + old_page), MAINTAIN.md:14-27 (new schema checker row), APPROVE.md:32, embedded
adr.yaml as exemplar.

**Known data-quality hazards to fix before publishing a JSON view:** duplicated exclusion
logic in pagebuilder vs internal/markdown (becomes public contract); `akb.annotations[].position`
offset means two different things on lint vs write paths; `char_count` = bytes not code points;
`file.dir` OS separator leak; `internal/cel/types.go` typed structs are dead code (natural home
for the JSON view contract).

### Status

Coverage complete: all 31 commands, all internal subsystems, roadmap. Seam map reviewed and
accepted by owner → RFC commissioned.

---

## 2026-09-26 — Session 3: Owner clarifications + RFC-001

### Owner clarifications (design constraints, ratified)

1. **Embedded adr/note templates are EXAMPLES ONLY** — adr is deliberately convoluted to
   showcase CEL's range. Neither is meant for use as-is. The 26-rule analysis in session 2 is
   a worked example validating the seam rule, NOT a coverage target.
2. **akb is permissive/unopinionated** about KB content; agents create whatever templates
   their job requires (the embedded kb-management SKILL points them at the examples).
3. **Coverage target = the upstream specs**: maximum JSON Schema 2020-12 + maximum latest
   cel-go — not the embedded rules.
4. Where a constraint is expressible in **both** languages, the SKILL must carry clear,
   low-token guidance on which to use and why.
5. Fine details ratified later via ADRs → tech-specs; immediate deliverable is an Extended
   RFC-style proposal.

### Gap check (this session)

- cel-go **v0.28.0** pinned in go.mod (flagged as missing evidence in session 1 research);
  full ext suite available in the module cache. ~~"Maximum CEL coverage" = ext enablement in
  `NewEnv`, no upgrade needed.~~ **CORRECTED same day (owner):** v0.28.0 is four releases
  behind — latest is **v0.32.0** (2026-08-19, per cel-expr/cel-go releases; assistant's
  training-data claim that the pin was "current" was wrong). v0.32.0 carries a BREAKING
  module-path migration to `cel.dev/cel-go` (#1413). Notable v0.29–v0.32 features for this
  design: JSON encoder ext (v0.29), network IP/CIDR (v0.29), `timestamp()` rejects
  non-RFC3339 (v0.30), native types in core + regex plan size controls + env COW (v0.31),
  timestamp parsing helpers + NativeToValue JSON support (v0.32). RFC §9/G2/ADR-3 updated
  accordingly.

### Outcome

**`../RFC-001-json-schema-cel-coexistence.md` written.** Key positions:
- Seam rule P1 (one document at one instant) is the formal boundary.
- `document` vs `page` artifact split (raw vs date-coerced) with one builder.
- Template format: `schema:` becomes a JSON Schema 2020-12 document over the document view;
  `schema.frontmatter` retired with loud rejection; `validations[]`/`lint_rules[]` unchanged.
- Validator library = coverage strategy (santhosh v6 recommended → ADR); Go validator makes
  both layers RE2 → dialect concern dissolves in practice.
- CEL ext enablement: Strings/Lists/Sets/Math/Bindings (additive).
- Overlap guidance: 3-bullet SKILL decision rule.
- `read --json` (two tiers) / `write --json` (Design A; derived keys output-only; round-trip
  invariant pinned as integration test).
- Approve gate hole closed (schema + write-time CEL re-run).
- 8 ADR/tech-spec candidates enumerated; 4-phase plan; risk table.

### Status

RFC-001 in Draft, awaiting owner ratification. Next: owner review → ADRs per RFC §15.

---

## 2026-09-26 — Session 4: Owner challenge on §5.2 "structured-content input rejected"

### Owner's challenge (verbatim intent)

Owner's target flow: pi-subagent ends turn with a gated `structured_output` tool call →
outputSchema maps to the akb template's JSON Schema → payload piped into akb → persisted as
markdown+frontmatter via the usual pipeline. Owner disputes the RFC's "structured-content input
is rejected" blocker: "markdown absolutely can travel as JSON". Owner asks whether CEL mutates
values (and if so, whether it should / whether mutation could be runtime-only).

### Findings (this session, gap-fill only — no re-research)

1. **Owner is right; the RFC sentence is mis-worded, not wrong.** The rejection targets
   *Design B* (akb renders markdown FROM a structured JSON content representation — headings/
   paragraphs arrays). The owner's flow is *Design A* (markdown travels as a JSON **string** in
   `content.raw`, authored by the agent) — already the accepted `write --json` design.
   §5.2 conflates "structured content" with "JSON transport"; must be rewritten.
2. **CEL never mutates.** Verified `internal/cel/pagebuilder.go:29-63`: `BuildPage` builds a
   **fresh** `fmMap`; `convertDateField` returns new values; `fm.Fields` untouched. Coercion is
   evaluation-time only, in-memory; disk always holds raw YAML. The only on-disk mutations are
   akb write-logic's managed frontmatter (`created` default, `is_draft`), which is why
   `write --json` echoes the effective document. No design change needed — RFC already says this
   (§5.1, §11) but doesn't say it *defensively* where a reader asks the question.
3. **pi-subagents gap-fill (verified in src):**
   - `outputSchema` is available **per-run** in the launch plan
     (`runs/shared/child-launch-plan.d.ts:11,22,56`; `shared/types.d.ts:2064`) — not only static
     agent frontmatter. Dynamic injection at spawn time is possible.
   - Gate/verify commands spawn with `stdio: ["ignore","pipe","pipe"]`
     (`runs/shared/acceptance.js:1381`) — a gate command does **not** receive the child's
     structured output on stdin. Gates validate the *command's own stdout*. So "gated" in the
     owner's flow = the TypeBox-gated `structured_output` tool call; the pipe into
     `akb write --json` is the **orchestrating parent's** job (`run.structuredOutput` → stdin).
   - (Prior research stands: TypeBox validator, drafts 3→2020-12, wrapper rewrite of local
     `$ref`s, 4096-byte bounded rejection errors, typed-gate/outputSchema mutual exclusion.)
4. **New design gap surfaced:** the agent-facing outputSchema validates the *input* shape
   (`{frontmatter, content.raw}`) while the template's `schema:` block validates the full
   *document* (incl. derived views). Two schemas, two objects — RFC must define how one source
   yields both (derivation vs explicit second block vs hand-maintained). → owner question.

### Status

Awaiting owner decisions (3 questions: schema-source strategy, input envelope shape, RFC scope
on the pi-subagents side) before RFC-001 refinement pass.

---

## 2026-09-27 — Session 5: OWNER PIVOT — openness audit commissioned

### Owner's independent research (input/output mutations)

Owner ran a live trace of `akb template write` / `akb write` against a throwaway KB. Key facts:
- `template write`: byte-for-byte copies of 3 inputs; zero injection (benign).
- `write`: conditional re-marshaller — injects `created`/`updated`, deletes `is_draft:true`,
  re-marshals whole frontmatter on any trigger (destroys comments, key order, quoting,
  indent); verbatim passthrough only when created+updated present and page is new; `--append`
  always re-marshals + bumps updated + `\n` blank-line bug with `--dated`; `--frontmatter`
  stores all values as strings; template `dir` rewrites page path; git commit + search.db +
  akb.lock side effects; index.md untouched.

### Owner's thesis (trajectory-changing)

akb's remaining mandates (obligatory title/type, managed-field mutations, fixed derived views,
fixed lint checker set, prescribed provenance/annotation vocabulary, date coercion, hand-rolled
wikilink parser duplicating ~70% of goldmark with two DRIFTED copies) are artifacts of the
pre-CEL restrictive design. If users define validity via JSON Schema + CEL, akb may not need to
prescribe ANY of it. Goal: **input = output** where possible; openness to maximum; agent-first
but NOT pi-subagents-specific. Owner questions: mandatory mutations, title/type, 10 shipped lint
checks (AddChecker exists but unreachable), fixed ast views (no tables/lists/blockquotes), date
coercion vs reject-at-write, provenance vocabulary prescription, hardcoded constants (CEL cost
100k, drift threshold 0.20, search fields, type→dir, is_draft lifecycle, commit messages),
custom CEL functions, goldmark-based rewrite of wikilink/exclusion scanners + runtime-exposed
goldmark capabilities in templating/linting.

### Owner's directives

1. Complete audit of akb-as-it-stands (subagents, recurse to fill gaps): where are we
   over-prescribing / forcing usage decisions?
2. Synthesize: where does RFC-001 not go far enough on openness?
3. Owner open to amending RFC-001 OR wholesale rejection + new RFC.
4. RFC text unchanged until owner ratifies; earlier 3 interview questions DEFERRED (pivot may
   reframe them).

### Status

Openness audit being commissioned (wave 1: 4 parallel scouts — write-path/frontmatter policy,
lint engine prescription, document model + parser architecture, config surface + subsystem
coupling). Synthesis → owner decision on amend-vs-new-RFC.

### Wave 1 complete (4/4 scouts, all completed, file-only outputs collected into research/)

- `openness-write-path.md` — 32 mandates (M1-M32) with consumer traces. Headline: `created`/`updated`
  have **ZERO internal consumers** (search `--after` reads index-time, not frontmatter); the
  whole-frontmatter re-marshal (M9) is the single largest input≠output defect; `is_draft:true`
  deletion is a free win; type→dir enforcement is internally contradictory (retype-without-move,
  silent cross-location duplicate); index.md is mechanically derivable → protected-file class could
  dissolve; approve bundles 3 jobs (flip is_draft, strip markers, presence gate) and never re-runs CEL.
- `openness-lint-engine.md` — engine is already an open registry (empty by default); the 10-checker
  mandate exists ONLY at cmd/akb/lint.go:157-167. Real openness bound: user rules can name only
  page-local facts — 4 cross-store checkers (broken_links, orphans, citations, index_consistency)
  depend on linkgraph/manifest/index exposed nowhere. 4 of 5 lint thresholds DEAD. Severity
  unvalidated (escape valve). `by_check` key set is the frozen contract to watch.
- `openness-document-model.md` — drift PROBED and confirmed: 3-class (pagebuilder) vs 5-class
  (markdown) vs 2-class (annotation) exclusion sets; live splits on indented code, escaped brackets,
  tilde-fence blind spot, frontmatter-vs-body annotation positions. 743/936 lines of wikilink.go are
  re-implemented CommonMark serving ~2 real consumers. Two fix shapes: (a) goldmark InlineParser
  rewrite (~120-150 lines), (a′) AST-derived exclusions (~40 lines, lower risk, fixes drift + tilde
  blind spot). Projections additive by construction (page is map(string,dyn)). Date coercion is
  name-agnostic + top-level-only; schema-declared `format: date` is the open regime.
- `openness-config-surface.md` — akb.yaml has 3 fields, 0 behavioral; unknown keys silently dropped.
  Full constants inventory with TRIVIAL/NEEDS-DESIGN/STRUCTURAL/ACCIDENT verdicts. Custom CEL
  functions: native bindings escape cost accounting → recommend closed akb-shipped function registry,
  template-selectable; cache key must become (envID, expr). Commit message format: 13 literal sites,
  zero consumers. Log operation enum is agent-workflow vocabulary (distill/plan) — harness-coupling
  smell. No git timeout anywhere (liveness risk, independent of openness).

### Gap-fill (parent, direct)

- `go.abhg.dev/goldmark/wikilink@v0.6.0` IS in the module cache: supports `[[target]]`,
  `[[target|label]]`, `[[target#fragment]]`, `![[embed]]` — but NOT akb's `[[display]](dest)`
  explicit-destination form nor the degenerate bracket-run semantics. ADR-level detail; option (a′)
  remains the lower-risk path regardless.

### Status

Wave-2 recon judged unnecessary before owner picks direction (remaining gaps are ADR-level).
Synthesis next → owner decision: amend RFC-001 vs new RFC, openness posture, index.md fate,
lint ownership model.

### Owner decisions (interview, 2026-09-27) — RFC-002 RATIFIED IN DIRECTION

- **D1: NEW RFC-002, supersedes RFC-001** (RFC-001 stays as research record).
- **D2: Maximal direction, phased delivery.**
- **D3: Demote managed fields to template-declared + input=output.**
- **D4: index.md becomes a fully derived artifact.**
- **D5: Lint = config registry + KB-facts CEL variable.**

### Owner's three embedded questions (answered in chat, this session)

1. (D2) "Why do untyped/invalid pages become possible?" → clarified: custom lint rules already
   exist (template lint_rules) and are central; the risk is write-path refusal relaxation
   (unknown-type pages, git-pull bypass) + built-in de-selection foot-gun. Sweep is the net
   because it's the only layer seeing ALL pages regardless of origin.
2. (D3) Date regime: verified cel-go v0.28 source — `timestamp()` = strict RFC3339
   (time.Parse(time.RFC3339)); `duration()` = **Go-style** time.ParseDuration ("1h30m"), NOT
   ISO 8601. JSON Schema `format: duration` (ISO 8601) ≠ CEL duration() vocabulary — real seam.
   Resolution direction: schema-assertion gate guarantees CEL preconditions for declared temporal
   fields; conversion is runtime-only/in-memory, restricted to schema-declared fields. Two open
   forks posed to owner: (a) date-only `2006-01-02` — drop vs declared bridge; (b) lint Phase B
   placement (early vs later phase).
3. (D5) Phasing rationale for KB-facts variable → explained: delivery order not commitment;
   Phase A (config registry) is ~15 lines; Phase B changes the lint execution model (new
   sweep-level rule class, KB-scale cost semantics, kb.* fact-surface + freshness design).

### Status

Awaiting owner's two fork decisions, then RFC-002 drafting (owner must explicitly greenlight
the write).

### Fork decisions (interview, 2026-09-27)

- **Date regime: schema-declared bridge.** Template declares temporal fields (format: date-time
  OR date); date-only coerced to midnight UTC in-memory only, declared fields only; schema
  assertion guarantees CEL preconditions; name-agnostic silent coercion dies. ADR candidates:
  midnight-UTC determinism spec, template-write static check (timestamp() on field lacking
  format assertion → warning), duration vocabulary mismatch (CEL duration() = Go-style,
  format: duration = ISO 8601) → document + optional closed-registry iso_duration() bridge.
- **kb.* phasing: core commitment, later phase** (after schema+CEL core + config registry).

### Status

All trajectory decisions ratified (D1-D5 + 2 forks). Remaining content decisions before
RFC-002 drafting: agent-facing input-schema source, untyped-page policy (type gate),
write --json input envelope, harness-integration scope. Then owner greenlight to draft.

### Content decisions (interview, 2026-09-27)

- **Input schema: akb derives + exports it** (`akb template get <name> --input-schema`;
  frontmatter + content only, projections stripped, content.additionalProperties:false,
  single-line strict JSON, explicit type on every node). Harness-agnostic.
- **Harness scope: akb contract + reference flows** (pi-subagents = one worked example).
- **Type gate: KEEP (selector only)** — with owner's amendment: type vs KB-layout coupling
  questioned; owner proposes optional `path` key, no path = flat. Under discussion (see below).
- **Envelope: PENDING** — owner challenges `content.raw` naming/shape; discussion in chat.

### Discussion points in flight

- type->dir: NOT hardcoded type=dir today; template `dir:` is optional (adr.yaml:11) but FORCED
  when set (write.go:418-441 strip/re-prefix). Owner's proposal ≈ 90% existing behavior minus
  the forcing. Open sub-question: dir as default vs hint-only vs constraint; layout expressible
  as CEL over page.file.path (maximally open option). Bare-filename resolution implications.
- Envelope: `raw` exists because page.content is an object carrying derived metrics
  ({raw, word_count, char_count}); user correctly identifies frontmatter/content as the
  canonical pair. Flatten proposal: content = string everywhere; derived views relocate to
  top-level output-only `derived` block. CEL churn verified: adr.yaml uses page.ast.* and
  page.content.word_count (mechanical renames under the already-breaking TemplateV2 revision).

### Final forks ratified (interview, 2026-09-27)

- **Layout: resolution-hint only.** `path`/`dir` does ONE job (bare-filename addressing);
  caller's path always verbatim; layout policy expressed as CEL/schema validation over
  page.file.path. M10/M11/M12 machinery deleted, not fixed.
- **Envelope: FLATTEN — content = markdown string everywhere** (envelopes, document view,
  CEL page map). Derived views relocate to top-level output-only `derived` block.
  CEL churn verified mechanical (adr.yaml page.ast.*, page.content.word_count).

### Derived-metrics decision (this session, facts verified)

- char_count: ZERO rule consumers, mislabeled (bytes vs CEL size() code points, cel-go
  string.go:176), redundant with schema minLength/maxLength on the flattened content string
  → SCRAP.
- word_count: one showcase consumer (adr.yaml:117); JSON Schema cannot express word-count
  (no keyword) so materialization only ever served CEL; ext.Strings split(" ") diverges
  from strings.Fields semantics → SCRAP as field, REPLACE as closed-registry CEL function
  word_count() (rides D8 registry). Design line: **the document carries facts, not
  computations** — derived block = parse facts only.

### Status

ALL decisions ratified except: (1) word_count function-vs-projection micro-fork (presented
with recommendation), (2) owner greenlight to draft RFC-002.

### FINAL decision register — ALL RATIFIED (interview, 2026-09-27)

| # | Decision | Ratified |
|---|---|---|
| D1 | RFC disposition | **New RFC-002**, supersedes RFC-001 (kept as research record) |
| D2 | Posture | **Maximal direction, phased delivery** |
| D3 | Managed fields | **Demote to template-declared + input=output** (created/updated injection removed from tool; is_draft → template-declared state field; byte-faithful writes) |
| D4 | index.md | **Fully derived artifact** (rebuild-only; protected-file class dissolves) |
| D5 | Lint | **Config registry + kb.* sweep-level CEL variable** (kb.* core commitment, later phase) |
| D6 | Parsers | **Both, sequenced**: (a′) AST-derived exclusions in parser phase; (a) full goldmark inline-parser rewrite as scheduled follow-on ADR |
| D7 | Dates | **Schema-declared bridge** (format: date-time|date; date-only → midnight UTC in-memory, declared fields only; schema assertion guarantees CEL preconditions) |
| D8 | CEL functions | **Closed registry ONLY** — akb-shipped, template-selectable; users CANNOT author functions; no exec escape hatch |
| D9 | akb.yaml | **Behavioral config home with unknown-key strictness** (hard error/warning on unknown keys) |
| D10 | Type gate | **Keep, selector only** (type = ordinary frontmatter key; schema enforces; engine only routes) |
| D11 | Layout | **Resolution-hint only** (path/dir for bare-filename addressing; layout policy = CEL/schema over page.file.path; caller path verbatim) |
| D12 | Envelope | **Flatten: content = markdown string everywhere**; derived views in top-level output-only `derived` block |
| D13 | Derived metrics | **char_count SCRAPPED; word_count = closed-registry CEL function**; document carries facts, not computations; derived block = parse facts only |
| D14 | Input schema | **akb derives + exports** (`akb template get --input-schema`) |
| D15 | Harness scope | **akb contract + reference flows** (pi-subagents = one worked example) |

### Status

ALL decisions ratified. Awaiting owner greenlight to draft RFC-002.

---

## 2026-09-27 — Session 6: RFC-002 DRAFTED (owner greenlight)

- **`.pi/subagents/proposals/json-cel/RFC-002-open-validation-engine.md` written in full.**
  19 sections + 3 appendices. D1–D15 woven through; RFC-001 spine carried per Appendix B;
  honesty list of surviving mandates at §16; 4-phase plan at §17; 18 ADR/tech-spec
  candidates at §18.
- **RFC-001 marked SUPERSEDED** (header only; content untouched as research record).

### Status

RFC-002 in Draft, awaiting owner section-by-section ratification. Next: owner review →
ADRs per RFC-002 §18.

### Transcript-recon gap check (owner-commissioned, low-reasoning agent) — findings processed

Verified all findings against session history. Clear omissions FIXED in RFC-002:
- G6 overlap-guidance SKILL decision rule restored (§6 bullet + Appendix B).
- D5 key-name declaration: sources/provenance checker keys configurable (§11).
- N2 restored to Appendix B.
- Audit dispositions added: §8.4 defect sweep (M23 encoding, M21 segment-wise .., H5 UTC,
  M25 predicate); §11 exit contract (A6/open-Q2), per-checker fault isolation (A4),
  10-checker disposition table, provenance re-expression route (registry fns), template
  delete exit-policy consolidation, dead PageData.Annotations; §5.3 code_blocks covers
  indented; §12 ManagedPage() predicate; §14 smalls (--now, discovery, modes, dead
  config.Created); §15 TYPED_VERIFY_OUTPUT_SCHEMA_CONFLICT, 4096B bound, $defs/$ref
  derivation survival; §16 honest list completed (git identity, search opinions B6/B8/
  B3-5, write coupling); §17 Phase 2 defect sweep; §19 known-unknowns row.

Positions taken under ratified principles (flagged to owner, not asked): exit-1-iff-
error-severity contract; per-checker degrade (P4-consistent); --orphans untouched.

NEW decisions surfaced → interview: D16 write-path coupling (M18: write requires git +
initialized search.db; read needs neither), D17 search --after semantics (B12: documented
as "creation date", reads index time).

### D16 + sequencing ratified (interview, 2026-09-27)

- **Sequencing: init-versioning spec lands FIRST, then RFC-002 phases.** Rationale (agreed
  with owner gut): spec is agreed-and-ready against current akb; RFC Phase 2 rewrites the
  same write-path call sites; spec gifts D9 its first behavioral config keys (versioning,
  git-author/git-email). RFC-002 §16 git lines to reference spec modes at ratification.
  ADR 16 flag: strict unknown-key config needs forward-compat story.
- **D16 (reduced to search.db): KEEP the coupling.** Write requires initialized search.db
  in both versioning modes; transactional write+index guarantee stays. §16 lists it
  structural. (Git half of D16 was pre-decided by the spec: versioning: git|none.)
- D17 scout commissioned: recon-search-after.md (mechanism + usage evidence).

### D17 recon landed (recon-search-after.md, probe-verified)

- Mechanism: SQLite-only filter `AND d.created >= ?` (lexicographic TEXT compare, no
  datetime functions). Column written in ONE place: IndexPageTx binds datetime('now')
  UTC; NO caller passes a date; frontmatter created never leaves the file.
- Rebuild AND any re-index (append/approve via INSERT OR REPLACE) reset the value →
  semantics = "last (re)indexed on/after X".
- "Broken and unused-looking" = ACCURATE both halves. Broken: RFC3339 input (the very
  format akb stamps) silently empty same-day (' '<'T'); garbage silently empty exit 0;
  rebuild makes every page match. Unused: zero tests, zero internal callers, one commit
  of history, docs-only. Bonus: documents.updated also index-time, read by nobody;
  --limit flag does not exist (hardcoded 10).
- Rewire cost: MECHANICAL — tags/summary lifting precedent; ~5 files, 2 packages, one
  extra param; store normalized UTC 'YYYY-MM-DD HH:MM:SS' to keep lexicographic compare;
  nullable + COALESCE fallback for undeclared pages; no index needed (post-FTS-join).

### D17 ratified + RFC-002 updated (2026-09-27)

- **D17: rewire --after to the declared creation field** (COALESCE fallback to index time;
  input validation + RFC3339 normalization fixed regardless; documents.updated dropped or
  documented; --limit becomes a real flag; creation-role declaration mechanics → ADR 19).
- RFC-002 edits: §16 git lines now reference init-versioning spec modes + D16 ratified
  coupling; new §14.1 search semantics repairs; §17 precondition (spec-first) + Phase 3
  gains search repairs; ADR 19 added; Appendix A gains D16/D17/D-seq rows; ADR 16 gains
  forward-compat note (strict keys vs spec-era keys).

### Status

RFC-002 draft complete and gap-checked (transcript recon processed; D16/D17/D-seq ratified
and woven in). Awaiting owner section-by-section ratification read.

### RFC-002 final edits applied (2026-09-27)

§16 git lines now reference init-versioning spec modes (versioning: git|none; spec §6
identity chain) + D16 ratified coupling in both modes; §14 "Stays structural" paragraph
restored to §14 (merge-conflict → merge preflight per spec); §14.1 search repairs in
place. Markdown clean.

### Status

RFC-002 COMPLETE as draft: all ratifications woven (D1–D17 + D-seq), transcript-recon
gaps closed, §16 honest list updated for the spec. Awaiting owner ratification read of
.pi/subagents/proposals/json-cel/RFC-002-open-validation-engine.md. Next after
ratification: implement init-versioning spec (D-seq), then ADRs per §18.

### Recon round 2 (owner-supplied miss list) — all fixed, no new decisions needed

1. Annotation body grammar (whitespace k=v, no spaces in values) → §13.3: grammar is
   part of the declaration, not just the prefix; ADR 10.
2. Git has no timeout (C5) → §8.4 bounded git invocation (context timeout).
3. Promised defensive "CEL never mutates" statement → §5.2 (evaluation-only; BuildPage
   copies; Evaluate read-only; old_page fresh read; only writer = write path, P6).
4. Silent page-skipping in derived-index path (B7/M5) → §12: derivation never silent
   (P7); rebuilds report skips; missing_frontmatter stays the error-severity net.
5. §16 non-git half expanded (FilesystemProvider, flock, --no-commit no-op, status line,
   identity preflight exit 2 — spec §4/§6).
6. AGENTS.md doctrine updates → §17: every phase ships its own (spec §10 precedent);
   enumerated the stale assertions.
Minor: smalls line extended (C4 retry values, D2 names, E4 naming, F6 manifest guard);
ADR 13 gains log.md heading grammar (F4); §13.3 full goldmark extension list; §6
temporal_note:null dead-YAML nit removed.

### Recon round 3 (owner-supplied miss list) — fixed; RFC MOVED

**RFC-002 relocated to `docs/kb/rfc/RFC-002-open-validation-engine.md`** (owner, in
preparation for akb dogfooding its own KB; RFC-001 moved alongside by owner).

Fixes applied at new location:
1. Template V3 unknown-key strictness → §6: hard-reject at `template write`, warn at
   load (forward-compat per ADR 16) — D9 principle extended to templates.
2. Round-trip/`file` contradiction → §15: `file` moved to the `--derived` tier;
   `read --json` = literally `{frontmatter, content}` = the write input; invariant is
   now "identity. Full stop." (no modulo at all). Position taken under D12/D14 —
   flagged to owner, object-if-disagree.
3. CEL projection paths → §5.3: `page.derived.ast.*`, `page.derived.markers|annotations`;
   embedded-template renames stated (page.ast.* → page.derived.ast.*;
   page.content.word_count → word_count(page.content)).
4. --tag match semantics → DEFERRED to search tech-spec (pointer in §14.1) — owner's
   pushback invitation accepted: this is spec/ADR territory, not RFC.
5. --json/--append/--frontmatter mutual exclusion → restored in §8.2 + Appendix B.
6. Appendix C gains recon-search-after.md; §11 notes per-template checker-enablement
   constraint (LintChecker.Check global signature).

### Recon round 4 (owner-commissioned independent check) — all 8 fixed, no new decisions

1. Round-3 file fix half-applied → §8.2 echo now mirrors `read --json` shape exactly
   ({frontmatter, content}; path is argv-known); §15 write row forbids derived OR file;
   §5.2 consumer line clarified (plain read = subset, --derived = full).
2. M20 + kb/ prefix contradiction → file.path/dir convention pinned to today's stripped
   relPath (example fixed: notes/foo.md, notes); §8.4 gains prefix-handling
   consolidation (5 copies → one internal/path normalization).
3. §8.3 index add mention removed (retired by §12).
4. Derivation rule generalized → §15: ANY keyword referencing stripped subtrees (ref,
   dependentRequired/Schemas, if/then/else, unevaluated*) = template-write error;
   mechanics to tech-spec.
5. Header table stale D1–D15 → D1–D17 + D-seq.
6. N1 explicit in Appendix B; §16 names the replaced -c user.name=akb fallback (M29/C3).
7. Replace-before-you-disable → §11 (Phase-4 SKILL doctrine + disabled-set reporting).
8. New-key defaults → §6: functions=[] empty; projections: none (minimal document, P5);
   state_field absent → approve = usage error pointing at lint.

Positions taken under ratified principles (flagged, object-if-disagree): stripped-relPath
convention kept (linkgraph/search/index keys unchanged); write echo = read shape;
approve-without-state_field = usage error.

### Recon round 5 — completions applied + D18/D19 ratified

Completions (parent, under ratified principles): #3 severity ceiling semantics in §11
(off/warning=cap/error=as-shipped; per-issue-class → ADR); #4 §19 stability row points
at ADR 18; #5 N4 deliberately superseded in Appendix B; secret-scanning roadmap note
carried; #6 one-parse consolidation assigned to ADR 7 in §13.1; #8 §16 gains exit-code
contract (H7) + DB schema/VerifySchema (B11) + M32 read-coupling note; smalls: M13
append join in §8.4, documents.created named in §14.1.

Ratified (interview):
- **D18: title = declared display field** (title_field:, default title — §12 summary
  precedent; undeclared → documented fallbacks: empty title column, filename entries).
- **D19: untyped pages in derived artifacts = store raw, group explicitly** ((untyped)
  index.md section; literal --type match; never silently dropped; type_orphan flags).

Both woven into §6, §12, Appendix A. RFC-002 register now D1–D19 + D-seq.

### Recon round 6 — all 9 fixed, no new decisions

1. **Date guarantee punctured + repaired (§5.2/§9):** RFC 3339 permits lowercase t/z +
   leap seconds; Go time.RFC3339 (cel-go's parser) rejects both (independently verified
   by recon). Fix: akb's date-time assertion is DEFINED as the CEL acceptance set
   (Go-strict profile), pinned ADR 1, documented as "akb date rules" in SKILL;
   format: time = string-only never coerced; bridge traversal scope = wherever declared
   (top-level-only limitation dies); mechanics ADR 4.
2. §7 pipeline gains explicit build-document step + ordering rationale (declaration
   load → materialize → validate; audit open Q4).
3-5. Staleness fixed: header register D1–D19; spec path → docs/kb/spec/ (owner moved);
   research count 12 → 13.
6. §5.2: old_page receives the IDENTICAL bridge under the CURRENT template's
   declarations (both sides of timestamp() must be bridged).
7. §10 registry seed gains marker_count + line_count/non_empty_line_count (contingent
   on a′) — the pair §11's provenance re-expression depends on.
8. §14 severity split clarified: code-shipped vs per-template lint_rules vs config
   ceiling override.
9. §12: index.md parse side dissolves (ReadIndex/parseIndex + heading grammar + em-dash
   entry contract die as dead code; renderer write-only, grammar ADR 12).

### Housekeeping flagged by recon (owner attention, not yet ratified)

Unratified position-taken set: stripped-relPath convention; write echo = read shape;
approve-without-state_field = usage error; file in --derived tier; exit-1-iff-error
contract; per-checker degrade; --orphans untouched. All flagged in LOG; owner to
confirm during ratification read.

### Recon round 7 — all 8 fixed + provisional register added

1. config description key → §14 seed (kept, live: init --description/discover). D9 strict
   loader no longer breaks existing KBs on it.
2. Derivation contradiction resolved → §15: explicit type on every ASSERTION node
   (inferred types added, e.g. enum-only); $ref-only nodes exempt by construction;
   $defs/refs still verbatim. Mechanics → tech-spec.
3. summary declaration mechanism → §12: summary_field:, default summary, same mechanism
   as title_field (D18); mechanics ADR 16/tech-spec.
4. §16 gains --only -- partial commits (completeness).
5. Dialect ownership → §13.3: akb.yaml = KB-global default; template set overrides per
   type.
6. Sweep ordering restated in §11 (schema before cel_lint; one now stamp) — RFC-002 now
   self-contained vs superseded RFC-001 §13.
7. Header research-base path → .pi/subagents/proposals/json-cel/research/.
8. §12 "existing repo lock" → write lock (repo lock git mode / flock non-git).

**Provisional positions register added to Appendix A (P-a..P-i)** — the 9 normative
positions taken under ratified principles are now marked IN THE DOCUMENT as pending
owner ratification (previously only flagged in LOG/chat): stripped-relPath, write echo
= read shape, approve-without-state_field = usage error, file in --derived tier,
exit-1-iff-error, per-checker degrade, --orphans untouched, log enum opened, no new
flags.

### P-register ratified (interview, 2026-09-27)

P-a…P-h ratified AS POSITIONED. P-f amendment: checker-error rows must be actionable
(checker, cause, remediation hint; self-diagnosis depth → tech-spec). **P-i OVERTURNED**:
flags allowed where they earn their keep (owner: sandboxed agents may not be able to
read the KB's config); precedence flag > akb.yaml > default. §11, §14, header, and
Appendix A updated — provisional table is now the ratified P-register with outcomes.

### Status

RFC-002: all content ratified or P-registered. Full register = D1–D19 + D-seq + P-a…P-i.
Awaiting owner's section-by-section ratification read (or declaration of ratification).
Next after ratification: implement init-versioning spec (D-seq), then ADRs per §18.
