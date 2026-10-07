---
as_of: "2026-10-06"
created: "2026-10-06"
grammar: 1
informs:
- ADR-004
is_draft: false
provenance: agent-drafted
run: rfc-002-a3-celgo
scope:
- internal/cel/**
- go.mod
- cmd/akb/**
status: final
summary: "Run log for the RFC-002-A3 cel-go upgrade research: question, RFC constraints, subagent inventory, and the interview forks that resolved ADR-004's pin and regex-limit decisions."
tags:
- cel-go
- validation
- adr-004
- research-log
title: RFC-002-A3 research log — cel-go upgrade and pinned extension set
type: research
updated: "2026-10-06"
---
# RESEARCH LOG — RFC-002-A3: cel-go upgrade + extension set

## Question

Draft the ADR resolving RFC-002 candidate **RFC-002-A3**: cel-go upgrade
(v0.28.0 → v0.32.x, module-path migration to `cel.dev/cel-go`) + extension set
(`ext.Strings`, `ext.Lists`, `ext.Sets`, `ext.Math`, `ext.Bindings` in `NewEnv`).
On ratification it takes the next free global ADR number (**ADR-004** — ADR-003 is
highest extant) with `Origin: RFC-002 candidate RFC-002-A3`.

- MODE: start · SUBJECT: rfc-002-a3 · Date: 2026-10-06
- Owner instructions: consume RFC-002 in full (done); research via subagents per
  protocol; if auto-resolvable create the ADR with `akb` (AKB_KB set) per
  docs-writer skill; escalate via interview if owner input needed.
- Final artifact: ADR (destination granted: KB decisions dir via akb, per
  instructions — A1 precedent resolved as ADR-003).

## RFC-given constraints (anchors: RFC-002 §10, §5.2, §17 Phase 1, App. B)

- Upgrade v0.28.0 → v0.32.x; breaking module-path migration `github.com/google/cel-go` → `cel.dev/cel-go`.
- Enable `ext.Strings`, `ext.Lists`, `ext.Sets`, `ext.Math`, `ext.Bindings` in `NewEnv`. Additive; every existing rule keeps working.
- v0.30 `timestamp()` strictness noted; §5.2 date bridge designed around it.
- ADR-003 I4 pins the strict date-time pattern as the intersection of cel-go v0.28 and v0.32 acceptance sets — must survive the upgrade intact.
- Cache-key change `(env identity, expression)` is **RFC-002-A5's** scope, not A3.
- Phase 1 placement: ships with santhosh v6 + TemplateV3 work, but ADR itself is sequenced now (A1 precedent).
- Appendix B carries a "cel-go release analysis (v0.29–v0.32)" from RFC-001-era research — extract before re-researching.

## Current state (trivial checks, verified)

- go.mod:12 `github.com/google/cel-go v0.28.0` (direct); `cel.dev/expr v0.25.1` (indirect).
- KB decisions: ADR-001, ADR-002, ADR-003 → next free = ADR-004.

## Decisions made

- **D-a3.1 — Pin `cel.dev/cel-go v0.32.0`** (only v0.32.x tag extant; nothing newer). One right way; logged.
- **D-a3.2 — Single-change require+import rewrite** (old path = read-only alias shim; `replace` cannot rescue a version-only bump). One right way; logged.
- **D-a3.3 — Pin explicit ext versions** at the highest shipping in v0.32.0 (Strings v5 incl. cost estimators + precision cap; Math v2; Lists v2; Sets/Bindings single-version). **Owner-ratified 2026-10-06 via interview** (conviction slight; alternatives: library defaults, pin-Strings-only). Rationale: write-gate determinism — a `go get -u` must never silently change rule semantics; same posture as ADR-003 I4.
- **D-a3.4 — `cel.RegexProgramSizeLimit` left unbounded**, recorded as considered-and-declined with explicit revisit trigger (untrusted template ingestion). **Owner-ratified 2026-10-06 via interview** after a clarification round (owner asked what "adversarial rule ingestion" meant; answer: templates are the only regex source, pages are never compiled, akb is a one-shot CLI).
- **D-a3.5 — ADR-003 untouched**: I4 subset argument verified against v0.32 source; `convertDateField` pre-conversion means the frontmatter path never hits the v0.30 gate.
- **D-a3.6 — No cache-key change** (A5 scope); global ext enablement keeps one env shape, latency unchanged.
- **D-a3.7 — Defer** OptionalTypes / ext.Regex / TwoVarComprehensions / Encoders / Native per RFC-001 §9; do not adopt v0.32's `ParseTimestamp`/`NativeToValue` helpers.
- **D-a3.8 — Interview scope panel raised no disputes** → requirements AGREED; proceed to ADR-004 draft.

## Key findings

### Prior research base (scout efc5a2cb — full report in its context.md artifact)

- **Premise correction:** the "cel-go release analysis (v0.29–v0.32)" lives in
  `RFC-001-json-schema-cel-coexistence.md:277-301` (§9), NOT in cel-bridge-research*.md.
  The A1 audit flagged the four-row feature table `[STALE/unverified]` (training-data-derived,
  `rfc-002-a1-validator/extract-prior-research.md:81-86`) → treat as hedged; web child must
  re-verify each row.
- **VERIFIED at primary source (A1 run, `research-web.md` Q6c):** cel-go `timestamp()` =
  `time.Parse(time.RFC3339)` gated by `strictRFC3339Pattern` (PR #1338, first shipped
  v0.30.0); the gate accepts lowercase `t`/`z` (lowercase `z` then dies in time.Parse);
  v0.31 rejects out-of-range offset hours (#1391). v0.32.0 released 2026-08-19; repo moved
  to `github.com/cel-expr/cel-go`; module path → `cel.dev/cel-go` in v0.32 (#1413).
- **ADR-003 interlocks (non-negotiable for A3):** header `revisit` trigger = re-check
  timestamp() acceptance at every cel-go upgrade; I4 strict profile is a subset of BOTH
  v0.28 and v0.32 acceptance sets so the upgrade needs no page re-validation; N3 explains
  raw-Go parse laxness under v0.28 (comma fractions, single-digit hours, +24:00); rejected
  alternative notes the raw-Go profile could only be adopted if akb pins cel-go ≥ v0.30
  permanently — after A3 that precondition holds (candidate simplification, flag in ADR).
- **Ratified decisions binding A3:** D8 (closed function registry, no user functions);
  cache-key change is A5's scope; D13 word_count = registry function because
  `ext.Strings` `split(" ")` ≠ `strings.Fields` semantics (LOG.md:399-401).
- **RFC-001 ext deferral list:** ext.Regex (needs OptionalTypes), optional types,
  two-var comprehensions, ext.Encoders/ext.Native — out of A3 scope.
- **Env mechanics:** `(*Env).Extend`, `cel.Function`, `cel.FunctionBinding` all exist at
  v0.28 → per-template envs don't need v0.31 COW; ext enablement is one line per lib.
- **Gaps confirmed as genuine A3 work:** full breaking-change inventory v0.28→v0.32;
  exact patch target / v0.33+ existence; ext cost accounting vs 100k bound; whether
  #1414 timestamp helpers / #1402 NativeToValue replace akb code; no v0.32 compile check.

### Code recon (scout 8dd0b5fa — full report in its context.md artifact)

- **cel-go surface is tiny:** 7 files, 4 subpackages (`cel`, `common/types`,
  `common/types/ref`, `interpreter`), ~15 call sites. No `ext`/`parser`/`checker`
  imports. `cel.dev/expr v0.25.1` + `antlr4-go/antlr/v4 v4.13.1` are indirect,
  pulled only by cel-go (`go mod graph` verified).
- **The exact version-sensitive identifiers:** `cel.CostLimit(uint64)`
  (`internal/cel/engine.go:61` — NOT ActualCostLimit, which doesn't exist at v0.28);
  `interpreter.EvalCancelledError` + `interpreter.CostLimitExceeded`
  (`engine.go:86,99`, `engine_test.go:314-315`); cel-go-owned error strings pinned in
  tests: `"no such key: ..."` (write_test.go:1090,2304; templates_write_test.go:734,845;
  template_test.go:443), `"overload"` substring (engine_test.go:285-286).
- **Cache key is expression-only** (`engine.go:53,63`) — the known A5 issue; one global
  sync.Map across envs. Confirmed A3 must NOT fix this (A5 scope) but must not make it
  worse: enabling ext globally in NewEnv keeps one env shape, so the bug stays latent.
- **Date path:** `convertDateField` (pagebuilder.go:26-46) pre-converts to `time.Time`
  BEFORE CEL sees values → `timestamp(field)` bypasses cel-go's string gate entirely;
  only literal `timestamp("...")` in rules hits the v0.30 strict gate. Embedded
  templates use `timestamp()` + `duration("720h"/"2160h"/"4320h")` (adr.yaml,
  note.yaml) — Go-style durations, no ISO 8601.
- **`internal/cel/types.go` structs (Heading/Link/CodeBlock with cel tags) are
  vestigial** — never used to build maps. Note for ADR as incidental cleanup, not A3.
- **Cost test pins wall-clock:** cubic-cost rule over 2048 headings must abort <10s
  (engine_test.go:186-203) — the canary if v0.32 cost accounting changes behavior.
- **go.mod `go 1.26.1`** — web child must confirm v0.32.x min-Go compatibility.

### Upstream web research (researcher 78939e4b — full detail: `.pi/research/rfc-002-a3-celgo/web-cel-go-releases.md`)

- **Module path:** v0.32.0 (2026-08-19, PR #1413) moved module to `cel.dev/cel-go`;
  `github.com/google/cel-go` is now a READ-ONLY alias shim, last release v0.31.0. One
  change must rewrite require + all imports; `replace` does not rescue a version-only
  bump (4 downstream PRs corroborate: heimdall#3483, kromgo#371, cilium#48341,
  gitlab client-go!3010). VERIFIED.
- **Target:** "v0.32.x" today = exactly **v0.32.0** — no v0.32.1 tag, no v0.33
  (releases.atom + pkg.go.dev, 2026-10-06). Go floor 1.23.0 (repo: go 1.26.1 ✓).
  `cel.dev/expr` stays v0.25.1; antlr4-go/antlr/v4 v4.13.1 retained. VERIFIED.
- **API stability for akb's surface:** `cel.CostLimit` doc-identical v0.28↔v0.32
  (source compared at both tags). **No `cel.ActualCostLimit` exists** — RFC/recon
  premise corrected. Panic-recovery shape unchanged: v0.32 `prog.Eval` recovers
  `EvalCancelledError` into an error return; akb already handles BOTH panic and
  returned-error paths (`engine.go:86-89,98-102`) → forward-compatible. The one
  compile-break surface (v0.29 `InterpretableV2`/`ExecutionFrame`) only affects custom
  `Interpretable` implementors — akb has none. VERIFIED.
- **Cost NUMBERS moved in v0.32** (issue #1476): `optional.*` macro estimates up ~60%
  (18→29, 22→31); partial revert is post-v0.32.0 (PR #1487). akb impact ≈ nil:
  OptionalTypes not enabled, cost test is behavioral (abort <10s), not numeric.
- **v0.30 new hard defaults:** expression ≤100k code points, ≤100k AST nodes, depth
  ≤250 — compile-time, aligns with akb's fail-closed posture; rules are tiny.
- **timestamp() gate (VERIFIED at source, v0.32.0 `common/types/timestamp.go`):**
  `isStrictRFC3339` grammar — uppercase/lowercase T and z pass the REGEX but lowercase
  z + leap seconds then die in `time.Parse`; rejects comma fractions, single-digit
  fields, out-of-range offsets (+24:01, +00:60) that v0.28 silently accepted.
  Resolves A1's "contradiction #1" as far as the ADR needs (net rejection in both
  versions; only the error message differs). Combined with code-recon: akb's
  convertDateField pre-converts to time.Time → frontmatter path never hits the gate;
  only literal `timestamp("...")` rules do. ADR-003 I4 subset argument HOLDS.
- **ext libs all present at v0.32.0** (VERIFIED, `ext/README.md` at tag). Version
  gates matter: `ext.Strings` v≥1 format/quote, v≥2 join, v≥3 reverse, v4=spec-v2
  format, **v≥5 installs cost estimators/trackers + precision cap 100**;
  `ext.Strings()` default = MaxUint32 (everything). ext.Math v2 (sqrt), ext.Lists v2
  (distinct/sort/sortBy), ext.Sets, ext.Bindings (cel.bind macro).
- **ext cost accounting (resolves open question):** ext libs register static
  `checker.OverloadCostEstimate` + runtime `interpreter.OverloadCostTracker` where
  costed (sets, strings≥5, lists, math — cost PR series #1352-54/#1360); uncosted
  functions fall to default cost++ (=1). RFC-001:300's "costed like stdlib" claim is
  substantively VERIFIED, with the mechanism now documented.
- **Regex:** RE2 linear-time, runtime cost = strCost×regexCost; v0.31 added OPT-IN
  `cel.RegexProgramSizeLimit` — default unbounded (applied only when >0).
  `lists.range` has an OOM cap (v0.29, `ListsMaxRangeSize`, default value = GAP).
- **Security:** current v0.28.0 pin is EXPOSED to CVE-2026-83530 (alloc-before-limit,
  fixed v0.29.0) and GHSA-gcjh-h69q-9w9g (NativeTypes json:"-" exposure, fixed
  v0.30.0). Both moot at v0.32.0 → strengthens upgrade motivation. VERIFIED.
- **Deferred-list confirmation:** OptionalTypes / ext.Regex / TwoVarComprehensions /
  ext.Encoders / ext.Native all exist at v0.32.0 but stay deferred per RFC-001 §9.

## Open questions (resolved unless noted)

- ~~Exact v0.32.x patch target; whether v0.33+ exists~~ → v0.32.0 exactly; nothing newer.
- ~~Breaking-change inventory~~ → compiled above; only InterpretableV2 + module path,
  neither touches akb.
- ~~ext cost accounting~~ → verified above.
- ext global enablement → yes, "in NewEnv" per RFC; single env shape keeps the
  expression-only cache key latency (A5) unchanged. Confirmed.
- REMAINING (implementation-phase verification, ADR will note as such):
  `StringsVersion(5)` option symbol existence at v0.32.0; `ListsMaxRangeSize` default;
  go build after path rewrite; runtime timestamp-table check (medium-high confidence
  inference).


## Parked / deferred

- None parked during the run. Re-emitted residual items (ranked):
  1. **RFC-002 §18 row amendment** — on owner ratification, A3 row → `ratified` as
     ADR-004, body references rewritten (RFC-002 §10 anchor + the ADR-003 inline
     mention). Also resolves ADR-004's orphan lint warning (inbound link from RFC-002).
  2. **Implementation-phase verifications** (carried in ADR-004 Verification):
     `StringsVersion(5)`/`ListsVersion(2)`/`MathVersion(2)` symbol existence at
     v0.32.0 (fallback recorded in ADR); `ListsMaxRangeSize` default; runtime
     timestamp-table check (medium-high confidence inference); go build after rewrite.
  3. **`internal/cel/types.go` vestigial structs** (Heading/Link/CodeBlock, never
     used) — incidental cleanup, belongs to RFC-002-A6's builder rewrite, not A3.
  4. **A1-run dispute #1 residue** (lowercase-t/z net rejection) — resolved to
     ADR-sufficient confidence by source reading; runtime check rides item 2.

## Final artifact

- **ADR-004 RATIFIED 2026-10-06 by Pete Hope** (post-review; owner raised the
  RFC-001-reference self-containment question, answered below):
  `kb/decisions/ADR-004-cel-go-upgrade-pinned-extensions.md` — status `accepted`,
  `deciders: [Pete Hope]`, ratification line added to Context per ADR-003 precedent.
- **RFC-002 amended per §18 lifecycle** (same pass): A3 row → `ratified` with
  Resolves-as [[ADR-004-cel-go-upgrade-pinned-extensions|ADR-004]]; body refs
  rewritten (§10 `RFC-002-A3)` → wikilink, §5.2 "the A3 upgrade" → "the ADR-004
  upgrade"); frontmatter `updated` bumped. `spawns:` list keeps candidate IDs
  (A1 precedent). Both writes passed template CEL validation.
- Lint post-amendment: 2 orphan warnings (ADR-001, RFC-003 — both pre-existing);
  ADR-004's orphan warning cleared by the RFC-002 inbound link. 0 errors.
- Draft/amendment copies retained: `ADR-004-draft.md`, `RFC-002-edited.md`.
- **Self-containment amendment APPLIED 2026-10-06** (owner approved in review):
  RFC-002 §10 gains a "Deferred extension families" bullet restating RFC-001 §9's
  deferral list (ext.Regex/OptionalTypes/TwoVarComprehensions/Encoders/NativeTypes +
  the per-capability-ADR route, linking ADR-004 N3); RFC-002 §11's self-containment
  precedent is the pattern. ADR-004's N3, Decision Drivers, Alternatives, and
  References re-pointed from RFC-001 §9 to RFC-002 §10; RFC-001 §9 now cited in
  ADR-004's References as "historical research record only". Both `updated` stamps
  bumped; both writes passed CEL validation.
- Verified landed state: RFC-002 §10:438-442 (new bullet); ADR-004 N3:55, Drivers:87,
  Alternatives:97, Refs:117; `RFC-002-A3` now only in RFC-002 `spawns:` (line 12) and
  the ratified §18 row (line 787) — A1 precedent. Lint: 2 pre-existing orphans only.
- RUN COMPLETE (A3).

## Follow-up: RFC-002 Appendix B self-containment audit (owner-requested 2026-10-06)

| Run | Agent | Task | Output |
|---|---|---|---|
| bce683f0 | scout | Audit every Appendix B "carried from RFC-001" item: restated in RFC-002? load-bearing? RFC-001 home? Verdict per item (self-contained / load-bearing gap / record-only / stale) + ranked fixes | table inline + `.pi/research/rfc-002-a3-celgo/selfcontainment-audit.md` |

### Audit result (20 Appendix B items)

- **15 SELF-CONTAINED** (substance already in RFC-002 §4/§6/§7/§8.2/§9/§10/§11/§15/§17).
- **3 RECORD-ONLY** (coexistence/external research, DESIGN.md roadmap, cel-go release
  analysis) — RFC-001 is their correct home per ADR-004:117.
- **1 LOAD-BEARING GAP — #14 mockup obligations:** RFC-002 §6 name-drops "mockup
  obligations carry from RFC-001 §6.2" without defining them, while P4 "prove at
  authoring" (§4:132-133) and Phase 1 `template write` depend on them. RFC-001:214-221
  verified verbatim (pass mockup satisfies schema block + every CEL rule; fail mockup
  fails the intended validator via `<!-- FAILS: rule_id — … -->`, extended to
  `<!-- FAILS: schema: … -->`; optional-key-stripping + self-`old_page` proofs stay CEL-side).
- **1 STALE/DANGLING — #17 "§4.3 secret-scanning write gate":** no `DESIGN.md` exists
  in this repo (verified by find) and "§4.3" resolves in neither RFC (§4 = principles in
  both); substance is self-contained in App B itself, only the label dangles.
- **Contradictions:** C1 §3/App B call JSON I/O "unchanged" while §15 re-specifies the
  envelope (body already flags the supersession at §3:115-116); C3 §11 labels the
  `required_fields`/`schema` rows "RFC-001 §13 carries" while D10 deliberately keeps
  `type_orphan` (RFC-001 §13's subsumption half is superseded — a label, not a decision,
  defect); C2/C4 cosmetic (phase re-partition; App B header says "unchanged" above a
  row it labels superseded).
- **Anchor hygiene:** App B "(§7)" should be §8.2:355-358; N1/N2/N3 exist only as
  Appendix B one-liners (child rates them marginal).
- **ADR reliance: clean** — no ratified ADR depends on an RFC-001-only carried item;
  ADR-003 relies on RFC-002 §5.2/§9/§18, ADR-004 on §10 (restated last pass).
- Proposed fixes ranked; owner sign-off requested before editing RFC-002 body content.

### Self-containment fixes APPLIED 2026-10-06 (owner approved: all five)

1. **§6 mockup obligations restated** (RFC-002:295-301) — RFC-001 §6.2's three clauses
   inlined verbatim-in-substance; overlap-rejection sentence split out; the load-bearing
   gap is closed.
2. **Dangling "§4.3" label dropped** (Appendix B:892) — now "the secret-scanning write
   gate (a pre-RFC design-doc gate)"; substance unchanged.
3. **C1 wording sharpened** (§3:112) — "JSON I/O commands (envelopes re-specified in §15)"
   so "carries unchanged" no longer contradicts §15.
4. **C3 clarified** (§11 table:490) — `type_orphan` row now states D10 supersedes
   RFC-001 §13's claim that the schema checker subsumes it.
5. **Anchor fixed** (Appendix B:890) — "mutual exclusion (§8.2)"; the mockup-obligations
   row now reads "(now defined in §6)".

- Both writes re-validated (template CEL passed); lint unchanged: 2 pre-existing
  orphans (ADR-001, RFC-003), 0 errors.
- **Incidental live finding:** the shipped akb INJECTS `updated` on `akb write` —
  authored timestamps were replaced by the write-time clock (RFC-002:43 now reads
  19:22:34Z; same for ADR-004:29). This is exactly the RFC-002 §8.1 behavior
  ("`created`/`updated` injection") that Phase 2 removes as part of input=output;
  recorded here as a live confirmation of the superseded behavior, not a defect to fix
  in this run.

## Child-run index

| Run | Agent | Task | Output |
|---|---|---|---|
| 8dd0b5fa | scout | Code recon: cel-go usage inventory (imports, NewEnv/cost APIs, template timestamp() use, version-pinned tests) | inline result |
| efc5a2cb | scout | Prior-research extraction: `.pi/subagents/proposals/json-cel/research/` cel analysis, ADR-003 cel interlocks, A1 run residue | inline result |
| 78939e4b | researcher | Web: cel-go v0.29–v0.32 releases, module-path migration, timestamp strictness, ext lib inventory + cost accounting, current version/advisories | summary inline + `.pi/research/rfc-002-a3-celgo/web-cel-go-releases.md` |

## Decision-fork record

- Interview 1 (2 questions, 2026-10-06): ext-versioning → **Pin explicit ext versions**;
  regex-limit → owner requested clarification (what is "adversarial rule ingestion").
- Clarification delivered in chat (regex source = templates only; pages never compiled;
  one-shot CLI; v0.30 compile caps bound literal size).
- Interview 2 (1 question): regex-limit → **Leave unbounded** (considered-and-declined
  with revisit trigger).
