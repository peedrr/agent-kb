---
as_of: "2026-10-10"
created: "2026-10-10"
grammar: 1
informs:
- SPEC-004
is_draft: false
provenance: agent-drafted
run: rfc-002-s7
scope:
- docs/knowledge-base/kb/**
status: final
summary: Coverage map of all 46 research pages plus the specs/ADR/RFC landscape, so the RFC-002-S7 (schema-pass integration + sweep checker, draft SPEC-004) run does not repeat prior work.
tags:
- rfc-002
- prior-research
- coverage-map
title: RFC-002-S7 prior-research coverage map — all existing RFC-002 validation-engine research
type: research
updated: "2026-10-10"
---

# RFC-002-S7 Prior-Research Coverage Map

Read-only recon. Repo `agent-kb/agent-kb`. S7 = **Schema-pass integration**:
document construction, pipeline placement, compiled-schema cache, error-tree → SPEC-002 report
mapping, pre-gate removal, sweep-time schema checker (ADR-009); drafts as SPEC-004
(RFC-002 §18 row `:834`). This page maps everything already researched that S7 touches.

## What S7 is (anchor)

RFC-002 §18 row (`rfcs/RFC-002-open-validation-engine.md:834`):
> RFC-002-S7 | SPEC | 1 | Schema-pass integration: document construction, pipeline placement,
> compiled-schema cache, error-tree → SPEC-002 report mapping, pre-gate removal, sweep checker
> (ADR-009) — drafts as SPEC-004.

S6's requirement-extraction run produced the sharpest S6/S7 boundary table
(`rfc-002-s6-rfc-extract.md:67-89`) and left S7 ownership gaps open (see Cluster D).

## Coverage map — research pages (46)

`run` = frontmatter `run`; `as_of` = frontmatter. One line each = topic covered.

### Cluster A — json-cel run (2026-09-26/27): RFC-001 openness audit → grounded RFC-002

| path | run | as_of | coverage (topic) |
|---|---|---|---|
| `research/akb-recon.md` | json-cel | 2026-09-26 | Read-only akb internals recon (TemplateV2, CEL env, write/read/template-write flows, JSON precedents); feasibility of JSON↔CEL bridge. |
| `research/command-inventory.md` | json-cel | 2026-09-26 | Inventory of 31 akb leaf commands: purpose, I/O, exit codes, subsystems, which already expose `--json`. |
| `research/schema-survey.md` | json-cel | 2026-09-26 | Survey of `outputSchema` usage in pi-meta-config agents + pi-subagents validation (TypeBox draft 3→2020-12) vs CEL. |
| `research/seam-core-pipeline.md` | json-cel | 2026-09-26 | Where a combined frontmatter+body JSON view and schema validation can run; 8/22 CEL rules read body-derived views → derived keys output-only. |
| `research/seam-periphery.md` | json-cel | 2026-09-26 | Subsystems outside core pipeline (search, link graph, lint, raw, index, log, git): where a JSON-Schema/CEL seam can attach. |
| `research/cel-bridge-research.md` | json-cel | 2026-09-26 | External prior art: no JSON-Schema→CEL compiler exists; declared-subset Go generator is the path; K8s/protovalidate use coexistence. |
| `research/cel-bridge-research-b.md` | json-cel | 2026-09-26 | De-dup second survey: reverse CEL extraction provably partial; recommends hand-rolled Go generator + display-only AST extractor. |
| `research/recon-search-after.md` | json-cel | 2026-09-27 | `search --after` filters on documents.created (index time), reset by rebuild/append/approve — not page creation date. |
| `research/openness-config-surface.md` | json-cel | 2026-09-27 | akb.yaml holds 3 fields, 2 read, 0 behavioral; every knob worth exposing (compute budget, drift threshold, indexed fields, layout) is a Go constant. |
| `research/openness-document-model.md` | json-cel | 2026-09-27 | Probe-verified document model: frontmatter/file-identity mandates, derived page-map projections, markdown parsing drift across goldmark call sites. |
| `research/openness-lint-engine.md` | json-cel | 2026-09-27 | Lint engine: checker registry open in code/closed in CLI; all 10 built-in checkers classified by what they force. |
| `research/openness-write-path.md` | json-cel | 2026-09-27 | Write path prescriptions (created/updated injection, is_draft deletion, re-marshal, type-dir prefix) classified structural/load-bearing/accidental. |
| `research/synthesis-openness.md` | json-cel | 2026-09-27 | Synthesis of the four openness-audit seams; reframes RFC-001 into RFC-002 (mandates classified + openness ladder). |
| `research/json-cel-log.md` | json-cel | 2026-09-27 | Hub/log for the json-cel run: sessions 1-5, commissioned reports, owner decisions, RFC-001→RFC-002 trajectory. |

### Cluster B — RFC-002-A1..A4 (2026-10-06): validator lib, schema.frontmatter retirement, cel-go, date bridge

| path | run | as_of | coverage (topic) |
|---|---|---|---|
| `research/rfc-002-a1-extract-prior-research.md` | rfc-002-a1-validator | 2026-10-06 | Audit of RFC-001-era research on Go JSON Schema validator libs and date/format handling; verified/unverified/stale claims. |
| `research/rfc-002-a1-recon-code.md` | rfc-002-a1-validator | 2026-10-06 | Code recon: where validator-lib and format-assertion decisions touch akb, with file:line anchors. |
| `research/rfc-002-a1-research-web.md` | rfc-002-a1-validator | 2026-10-06 | Sourced survey of Go JSON Schema 2020-12 validators + RFC 3339 parsing; closes format-assertion questions behind ADR-003. |
| `research/rfc-002-a1-log.md` | rfc-002-a1-validator | 2026-10-06 | Run log: closed validator-library/format-assertion questions, owner forks, decision trail to ADR-003. |
| `research/rfc-002-a2-extract-prior-research.md` | rfc-002-a2 | 2026-10-06 | Audit of openness/coexistence base on schema.frontmatter retirement, migration mechanics, deferrals. |
| `research/rfc-002-a2-recon-code.md` | rfc-002-a2 | 2026-10-06 | Code recon of TemplateV2 loader, write-path required-field gate, embedded templates, in-tree inventory for migration. |
| `research/rfc-002-a2-log.md` | rfc-002-a2 | 2026-10-06 | Run log: owner forks on schema.frontmatter retirement/migration; trail to ADR-006. |
| `research/rfc-002-a3-web-cel-go-releases.md` | rfc-002-a3-celgo | 2026-10-06 | cel-go v0.28.0→v0.32.x upgrade survey: module-path migration, v0.30 timestamp strictness, extension costs, advisories. |
| `research/rfc-002-a3-log.md` | rfc-002-a3-celgo | 2026-10-06 | Run log: cel-go upgrade question, constraints, interview forks resolving ADR-004 pin/regex-limit. |
| `research/rfc-002-a4-log.md` | rfc-002-a4 | 2026-10-06 | Run log: date-bridge determinism spec, template-write static temporal check, duration seam; trail to ADR-005. |

### Cluster C — RFC-002-S1 (2026-10-07): structured validation report → SPEC-002

| path | run | as_of | coverage (topic) |
|---|---|---|---|
| `research/rfc-002-s1-research-extract.md` | rfc-002-s1 | 2026-10-07 | Verbatim RFC-001 merged-error-model requirements + prior a1–a4 error findings constraining the SPEC-002 report. |
| `research/rfc-002-s1-code-recon.md` | rfc-002-s1 | 2026-10-07 | Code recon of validation error machinery: runTemplateValidations aggregation, ValidationError, write-path sentinels, classifyExit. |
| `research/rfc-002-s1-adr-conventions.md` | rfc-002-s1 | 2026-10-07 | Conventions + governing decisions (ADR-003..006, SPEC-001, template contract) for authoring SPEC-002. |
| `research/rfc-002-s1-log.md` | rfc-002-s1 | 2026-10-07 | Run log: SPEC-002 scope, format authority, decision-fork strategy, child-run index. |

### Cluster D — RFC-002 Phase-1 scope + S6 (2026-10-09): document shape, sweep checker, TemplateV3 loader

Note: `RFC-002-S6` run pages carry the S6/S7 boundary and the explicit S7 gap list — **most load-bearing cluster for S7**.

| path | run | as_of | coverage (topic) |
|---|---|---|---|
| `research/rfc-002-phase1-scope-code-surface.md` | rfc-002-phase1-scope | 2026-10-09 | Recon of today's CEL page-map shape, word_count/char_count/raw consumers, required-gate coupling, loader leniency, lint interface, write/read flags. |
| `research/rfc-002-phase1-scope-test-pinning.md` | rfc-002-phase1-scope | 2026-10-09 | Recon of flatten-vs-current-shape test churn, goccy frontmatter value types, full BuildPage/BuildOldPage call-site list, approve-wording gap. |
| `research/rfc-002-phase1-scope-log.md` | rfc-002-phase1-scope | 2026-10-09 | Run log: four owner-ratified scope decisions → ADR-008/009/010; evidence index + parked items. |
| `research/rfc-002-s6-rfc-extract.md` | rfc-002-s6 | 2026-10-09 | Every RFC-002 requirement on S6 with section anchors, **S6/S7/S8 boundary table**, ADR-010 verbatim invariants, SPEC-003 page contract. |
| `research/rfc-002-s6-prior-research.md` | rfc-002-s6 | 2026-10-09 | Extraction of ratified prior decisions for S6 + **open gap list** incl. schema-compile site (S6 vs S7) and `type`/`title` un-routing scope. |
| `research/rfc-002-s6-code-recon.md` | rfc-002-s6 | 2026-10-09 | Current TemplateV2 loader, key-to-consumer inventory, embedded template corpus, type/title routing path, test pinning, file:line anchors. |
| `research/rfc-002-s6-gapfill-libs.md` | rfc-002-s6 | 2026-10-09 | Pinned-version verification of goccy unknown-key detection + santhosh jsonschema v6 compile surface; feasibility verdicts. |
| `research/rfc-002-s6-log.md` | rfc-002-s6 | 2026-10-09 | Run log: recon fanout, owner-ratified forks, filing record for SPEC-003 (incl. F2: compilation/cache = S7). |

### Cluster E — rfc-templates run (2026-10-01/02): RFC authoring format (tangential to S7)

| path | run | as_of | coverage (topic) |
|---|---|---|---|
| `research/rfc-templates-recon-local.md` | rfc-templates | 2026-10-01 | Local recon: owner's RFC-001/002 exemplars, the RFC doc type's missing committed home, harness consumption. |
| `research/rfc-templates-recon-landscape.md` | rfc-templates | 2026-10-01 | Web recon: published RFC/proposal templates mapped by revision lifecycle; where parseable status/disposition live. |
| `research/rfc-templates-lane-agent-era.md` | rfc-templates | 2026-10-01 | Lane on 2026 agent-era RFC literature: machine-extractable header, alternatives slot, pre-flight rubric, prose argument. |
| `research/rfc-templates-lane-deliberation.md` | rfc-templates | 2026-10-01 | Lane on deliberation/ratification: comment-disposition processes, machine-readable precedents, human-only ratification gate. |
| `research/rfc-templates-lane-lifecycle.md` | rfc-templates | 2026-10-01 | Lane on proposal lifecycle: IETF expiry/abolition, status rot, derived staleness postures vs clock kills. |
| `research/rfc-templates-lane-verify.md` | rfc-templates | 2026-10-01 | Verification lane: four IETF/KEP items re-checked against primaries; corrects I-D-stage premise. |
| `research/rfc-templates-synthesis.md` | rfc-templates | 2026-10-02 | Coordinator design analysis of the RFC as process unit; machine contract, ratification gate, lifecycle; ratified forks F1–F5. |
| `research/rfc-templates-report.md` | rfc-templates | 2026-10-02 | Findings report: the RFC authoring contract for agent-consumed design proposals. |
| `research/rfc-templates-audit-report.md` | rfc-templates | 2026-10-02 | Independent audit: all 18 claims verified (18 supported, 0 corrected/killed) with caveats. |
| `research/rfc-templates-log.md` | rfc-templates | 2026-10-02 | Run log: question, method, owner forks F1–F5, lane activity, decision trail. |

## Artifact landscape

### RFCs (`kb/rfcs/`)
| id | title | one-line subject |
|---|---|---|
| RFC-001 | JSON Schema / CEL Coexistence and JSON I/O in akb | status `superseded`; the predecessor that RFC-002 replaces (research record). |
| RFC-002 | The Open Validation Engine | status `draft`; akb as unopinionated validation+persistence engine, JSON Schema + CEL only validators. |
| RFC-003 | Give akb a Declared Actor Identity So Rules Can Assert Who Acted | status `draft`; expose resolved actor identity to the rule engine; roster policy stays in the KB. |

### Specs (`kb/specs/`)
| id | title | status | one-line subject |
|---|---|---|---|
| SPEC-001 | Init Selects the KB Versioning Mode and Resolves Commit Identity | completed | init-versioning precondition (RFC-002 §17 D-seq, lands before Phase 1). |
| SPEC-002 | Write-Time Validation Failures Merge Into One Structured Report | active | merged error model / structured report (S1 output); S7 maps schema error trees into this. |
| SPEC-003 | Templates Adopt a Strict Six-Key V3 Format With a JSON Schema Block | active | TemplateV3 format + loader (S6 output); S7 owns compile/cache/pass of its `schema:` block. |

### ADRs (`kb/decisions/`) — note: dir is `decisions/`, not `adrs/`
| id | title | status | one-line subject |
|---|---|---|---|
| ADR-001 | KB Diagnosis and Repair Live in a Dedicated akb Doctor Command | proposed | new `akb doctor` owns diagnosis/repair; remediation text names akb commands only. |
| ADR-002 | Init Author Flags Record a Durable Commit Identity in akb.yaml | accepted | complete `--author-*` pair recorded as git-author/git-email; ambient identity never recorded. |
| ADR-003 | JSON Schema Validation Uses santhosh v6 and Asserts Every Declared Format | accepted | santhosh-tekuri/jsonschema v6; every declared format asserts; akb date-time/date checkers pin strict RFC 3339. **S7 consumes.** |
| ADR-004 | CEL Runs on cel.dev/cel-go v0.32 with a Pinned Extension Set | accepted | cel-go v0.32.0 one-shot rewrite; five ext libs pinned highest; regex plan-size unbounded. |
| ADR-005 | The Date Bridge Coerces Only Schema-Declared Fields and Is Statically Checked | accepted | date bridge coerces only schema-declared temporal fields; template-write warns on unbacked calls; duration bridging deferred. |
| ADR-006 | TemplateV3 Retires schema.frontmatter — Migration Is a Recipe, Not a Command | accepted | hard-reject retired block with self-sufficient message; migration via CHANGELOG recipe. |
| ADR-007 | Research Reports Are First-Class Pages, Not Raw Sources or Chat Residue | accepted | agent research/logs are first-class `kb/research/` pages; raw/ holds external sources only. |
| ADR-008 | The RFC-002 Flattened Document Model Is Not Phase-1 Scope | accepted | Phase 1 validates today's shape (content.raw, ast.*); flatten/renames/word_count/write --json deferred. **S7 bound.** |
| ADR-009 | Phase 1 Ships the Sweep-Time Schema Checker and Keeps required_fields | accepted | built-in schema lint checker in Phase 1 (before cel_lint, one shared `now`); required_fields kept until Phase-3 registry. **S7 owns.** |
| ADR-010 | Phase-1 TemplateV3 Recognizes Six Keys and Hard-Rejects the Rest | accepted | V3 = name, description, dir, schema, validations, lint_rules; `template write` hard-rejects others, load warns. |

### RFC-002 phase map (for scoping S7)
- **Phase 1 — validation core:** santhosh v6 integration, TemplateV3 `schema:`, pipeline reorder, merged error model, cel-go v0.32.x, `template write` schema obligations, schema.frontmatter hard rejection, sweep schema checker (ADR-009). Tech-specs S6/S7/S8.
- **Phase 2 — write-path honesty + JSON I/O:** input=output, approve unbundled, `read/write --json`, `--input-schema`, round-trip tests, date bridge, §8.4 defect sweep.
- **Phase 3 — openness machinery:** akb.yaml config surface, lint registry, index.md derivation, parser pivot (a′), vocabularies, path resolution hints, §14.1 search repairs.
- **Phase 4 — extensibility + doctrine:** closed function registry (+word_count), `kb.*` sweep rules, inline-parser rewrite (a), SKILL/doctrine rewrite.

## Cluster flags — already deeply researched

| Topic area | Depth | Where |
|---|---|---|
| **Write-path + required-gate internals** | DEEP (write path audit, sentinels, classifyExit, pre-gate coupling, call-site list) | Cluster A `openness-write-path`, Cluster C `s1-code-recon`, Cluster D `phase1-scope-code-surface`/`-test-pinning` |
| **CEL engine / rule evaluation** | DEEP (env, ext libs, cost limit, page-map shape, consumers) | Cluster A `seam-core-pipeline`, Cluster B a3, Cluster D `phase1-scope-code-surface` |
| **Lint engine + checker registry / sweep checker** | DEEP (all 10 checkers classified; registry open/closed) | Cluster A `openness-lint-engine`; ADR-009 constrains S7's checker |
| **Template system (V2→V3 loader, keys, strictness)** | DEEP | Cluster B a2, Cluster D all S6 pages, ADR-006/010 |
| **Validator library + format assertion (santhosh v6)** | DEEP (lib survey, compile surface, format policy) | Cluster B a1 (all 4), Cluster D `s6-gapfill-libs`, ADR-003 |
| **Structured report / merged error model** | DEEP (source of the SPEC-002 contract) | Cluster C (all 4), SPEC-002 |
| **Document model / page-map construction** | DEEP | Cluster A `openness-document-model`, `seam-core-pipeline`, Cluster D `phase1-scope-*`, ADR-008 |

## Gaps (what S7 research does NOT yet have)

1. **Compiled-schema cache design.** No page specifies cache key/lifetime/invalidation, concurrency vs the single-connection DB rule, or where compilation sits (loader S6 vs schema pass S7) — flagged open in `rfc-002-s6-prior-research.md:151-155` and `rfc-002-s6-log.md:41` (F2: "compilation/cache/malformed-schema = S7"). **No dedicated cache page exists.**
2. **Error-tree → SPEC-002 report mapping.** SPEC-002 defines the report; no page maps santhosh v6 `*ValidationError` trees (instanceLocation/keywordLocation/causes) onto it. Closest: `rfc-002-s1-code-recon.md` (CEL-side only) and a1's validator survey.
3. **Pipeline placement of the schema pass.** `seam-core-pipeline.md` covers JSON↔CEL seam placement generically, but not the concrete S7 reorder relative to the required-gate/CEL steps under Phase-1 shape.
4. **Sweep-time schema checker internals.** ADR-009 fixes its existence/ordering/`now`; no page details per-page schema compile reuse, issue mapping, or `by_check` key stability implementation.
5. **Pre-gate removal mechanics.** Phase-1 keeps `required_fields` (ADR-009) while the JSON Schema pre-gate must be removed — the interaction (ordering, double-reporting) is not researched.
6. **Dogfood conversion + test fallout for the schema pass.** S6 enumerated its own fallout; S7's test impact (santhosh integration points, required-gate tests, cel_old_format) is not enumerated.
7. **Cache/malformed-schema error surface.** Malformed embedded `schema:` at load vs compile time — ownership (S6 vs S7) explicitly undrawn (`rfc-002-s6-prior-research.md:152-153`).

## Confidence + method

- **Confidence HIGH** that the 46-page list is complete: `find docs/knowledge-base/kb -type d` shows only `decisions/`, `research/`, `rfcs/`, `specs/`; all 46 `kb/research/*.md` frontmatters + summaries read; no nested research dirs.
- **Confidence HIGH** on run/as_of/status: taken from frontmatter, not prose.
- **Confidence MEDIUM** on one-liners: condensed from each page's own `summary:` (author-written), spot-checked against bodies for the key S6/S7 pages.
- **Gaps confidence MEDIUM-HIGH**: the "no dedicated cache page" claim rests on title/summary scoping; a page could mention caching in body — not exhaustively grepped.
- **Out of scope / not inspected:** `docs/knowledge-base/raw/` (external sources per ADR-007), git history of research pages, and the RFC-002 Appendix A decision register (D-register) beyond what research pages quote.
- **Note:** rfc-templates frontmatter lists `informs: RFC-003` although the run is about RFC authoring format — likely a linkage quirk, but noted as observed.
