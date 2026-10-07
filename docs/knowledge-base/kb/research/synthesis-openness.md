---
grammar: 1
type: research
title: "Synthesis — openness audit vs RFC-001"
status: final
provenance: agent-drafted
run: json-cel
as_of: 2026-09-27
created: 2026-09-27
updated: 2026-09-27
scope:
  - internal/**
  - cmd/akb/**
tags: [openness, synthesis, rfc-002, audit]
summary: "Synthesis of the four openness-audit seams: classifies every mandate as structurally required, opinionated-but-load-bearing, or historical accident, and reframes RFC-001 into RFC-002."
informs:
  - RFC-001
  - RFC-002
---
# SYNTHESIS — Openness audit vs RFC-001

**Date:** 2026-09-27 · **Inputs:** `openness-write-path.md` (32 mandates, M1–M32),
`openness-lint-engine.md`, `openness-document-model.md`, `openness-config-surface.md`
+ prior seam research · **Question:** where does RFC-001 adopt current behavior as
*required* when the evidence says it is *prescribed* — and what does maximal openness
look like?

---

## 1. What the audit established (digest)

Across the four seams, every mandate was classified **SR** (structurally required),
**OL** (opinionated-but-load-bearing), or **HA** (historical accident):

| Verdict | Count (approx., deduped) | Examples |
|---|---|---|
| SR — keep, no debate | ~8 | symlink containment, merge-conflict block, template-name regex, single-conn WAL DB, `--only --` partial commits, managed-file *protection as a concept*, cost limit *as a bound* |
| OL — keep deliberately, configure at most | ~12 | type-as-template-selector, draft lifecycle, auto-commit, body-verbatim indexing, FTS5-syntax stripping, kb/raw roots |
| HA — demote, derive, or delete | ~20 | **created/updated injection (zero internal consumers)**, whole-frontmatter re-marshal, `is_draft:true` deletion, hardcoded type/title gates (duplicated by schema `required`), `..`-anywhere rejection, 4 of 5 lint thresholds (dead), `olw-auto` prefix, UTC-vs-local clock split, naive index pluralization, closed strip vocabulary, commit-message format (13 literal sites, zero readers) |

**The three structural findings that reframe the RFC:**

1. **The input≠output gap is one function deep.** `created`/`updated` injection has *no*
   internal consumer (search `--after` reads index time); `is_draft:true` deletion is a
   no-op semantically. Remove those and the only remaining fidelity defect is the
   whole-frontmatter re-marshal (M9) — a serialization accident, not a policy.
   **input = output is achievable** on the write path with a document-level round-trip.
2. **The lint engine is already an open registry; the mandate is one call site**
   (`cmd/akb/lint.go:157-167`). The real openness bound is not the registry — it's that
   user rules can name only *page-local* facts, while 4 of 10 checkers depend on
   cross-store state (linkgraph, manifest, index.md) exposed to no rule language.
3. **The markdown layer is the deepest historical accident.** 743/936 lines of
   `wikilink.go` re-implement CommonMark structure goldmark already computes; a second,
   *drifted* copy lives in `pagebuilder.go` (3 vs 5 exclusion classes); probe-verified
   live inconsistencies (indented code, escaped brackets, tilde-fence blind spot,
   frontmatter-vs-body annotation offsets). The link graph itself needs almost none of
   it (`Target`+`Display` only).

## 2. Where RFC-001 as drafted does not go far enough

| RFC-001 says | Audit says | Gap |
|---|---|---|
| §11: round-trip invariant "modulo akb's managed mutations" (created/updated/is_draft echo) | Those mutations have zero internal consumers; they exist to serve user CEL rules that templates could declare themselves | RFC *enshrines* the mutation layer instead of questioning it. "Modulo managed mutations" should shrink to nothing or to template-declared auto-fields |
| §5.1: `document`/`page` split exists because `convertDateField` coerces name-agnostically | Coercion is name-agnostic, top-level-only, silently retypes strings; schema-declared `format: date` is the open regime | RFC treats coercion as a fixed fact to design around; audit makes it a candidate for schema-driven declaration |
| §5.2: derived views `ast.headings/links/code_blocks`, `akb.*` are akb-authored, fixed | Projections are additive by construction (`page` is `map(string,dyn)`); tables/lists/blockquotes/emphasis unreachable; vocabularies prescribed (olw-auto residue; strip-list closed to 3 while parse is open) | RFC freezes the projection surface into the public contract at the exact moment it should become *declared* (template-level `views:`/`projections:`) |
| §5.3 hazard 1: "publish means export-and-share" the duplicated exclusion logic | Don't share the hand-rolled scanners — delete them. AST-derived exclusions (~40 lines, fixes drift + tilde blind spot) or full goldmark inline-parser rewrite | RFC would cement the wrong primitive as a public contract |
| §6: retires `schema.frontmatter` but keeps hardcoded `type`/`title` engine gates | M1/M2/M4 are duplicate gates (both embedded templates already declare them required); `type` survives only as the *template selector* | RFC keeps two engine-level mandates it could demote to the schema layer (title entirely; type as selector only) |
| §12: approve gate fix (re-run validations) | Approve bundles 3 jobs (flip is_draft, strip markers, presence gate); stripping uses a *closed* vocabulary that silently retains custom markers | RFC fixes the hole but keeps the bundling; openness wants the jobs unbundled and the strip vocabulary declared |
| §13: sweep gains a schema checker; required_fields/type_orphan "retire or thin-wrap" | The whole checker set is one call site; the real question is checker ownership (config-driven enable/disable/severity) + exposing KB-level facts to rules | RFC shuffles the prescribed set but keeps it prescribed |
| N5: `lint --json` envelope frozen | `by_check` key set enumerates the registered checkers — checker enable/disable changes the key set | Freeze must be stated at the *envelope* level with disabled checkers reporting 0, or gates break |
| Non-goals: no mention of akb.yaml | akb.yaml has 3 fields, 0 behavioral, silently ignores unknown keys | The natural home for the openness policy (lint selection, thresholds, budget, commit format, indexed fields, dialect) is absent from the RFC |
| G2: enable cel-go ext libs | Ext libs are additive; *custom* functions escape cost accounting → closed akb-shipped registry, template-selectable; cache key must become (envID, expr) | RFC's CEL coverage story is right but silent on user-named functions and the cache-key consequence of per-template envs |

## 3. The reframe

RFC-001's **spine survives intact**: coexistence over translation, the P1 seam rule,
library-first coverage, `read/write --json`, the three-surface failure contract, the
phasing. None of the audit findings contradict it.

What changes is the **mission statement**: RFC-001 is framed as *"add JSON Schema
alongside akb's validation"*. The audit says the correct frame is:

> **akb becomes an unopinionated validation-and-persistence engine.** Every page-level
> mandate is either (a) structural (security, concurrency, the KB boundary), (b) derived
> (index.md, link graph, search index), or (c) user-declared (schema + CEL + template
> + akb.yaml). The tool ships *capabilities and examples*, never *policy*. input = output
> on the write path; anything akb adds to a page is either requested or declared.

Under that frame, JSON Schema + CEL are not an addition to the prescriptive core — they
are what *replaces* it. The embedded templates stop being "the shape of a valid page"
and become what the owner already says they are: examples.

## 4. Decision register (owner)

| # | Decision | Options | Recommendation |
|---|---|---|---|
| D1 | RFC disposition | Amend RFC-001 / **new RFC-002 superseding** | **New RFC-002.** The mission statement changes; ~50% of RFC-001's technical content carries over, but bolting openness onto a "coexistence" frame produces a self-contradictory document. RFC-001 stays as research record, marked superseded. |
| D2 | Openness posture | Maximal demotion as RFC *direction*, phased delivery / config-knobs-only / case-by-case | **Maximal direction, phased.** The audit shows most demotions are mechanically small (HA verdicts); declaring the direction once prevents each from becoming its own debate. |
| D3 | `created`/`updated`/`is_draft` | Remove tool injection entirely / demote to template-declared auto-fields / keep | **Demote to template-declared.** Zero internal consumers; templates that want timestamps declare `format: date-time` + CEL immutability rules. `is_draft` lifecycle is the one genuinely load-bearing piece (approve/list) — candidate for a template-declared *state field* rather than a hardcoded key. |
| D4 | `index.md` | Keep managed registry / **make derived** (`index rebuild` always; no `index add` hand-entries) / delete the concept | **Derived.** It is mechanically regenerable; derivation dissolves the protected-file class, the `index_consistency` checker, the M26/M27 asymmetries, and 8 copies of the exclusion set. |
| D5 | Lint ownership | Config-driven built-ins + expose KB facts to rules / built-ins stay privileged / demote page-local only | **Config-driven + KB-level CEL variable** (`kb.*` naming linkgraph/manifest/index facts), phased: config first (cheap), KB-facts second (design), demotion third. |
| D6 | Goldmark pivot | Full inline-parser rewrite (a) / **AST-derived exclusions (a′) first** / export-and-share as RFC-001 says | **(a′) now, (a) as ADR.** a′ is ~40 lines, kills the drift *and* the tilde blind spot, and doesn't re-home ~100 grammar test cases under time pressure. The third-party `goldmark/wikilink` ext covers `[[t]]`, `[[t|d]]`, `[[t#f]]`, `![[embed]]` but **not** `[[display]](dest)` — grammar gap makes (a) an ADR, not a given. |
| D7 | Date regime | Silent coercion (today) / reject non-RFC3339 at write / **schema-declared `format: date` drives coercion** | **Schema-declared.** Name-aware, template-owned, kills the top-level/nested asymmetry; `BuildPage` already has the template in scope at all call sites. |
| D8 | Custom CEL functions | Closed akb-shipped registry, template-selectable / free native bindings / none | **Closed registry** (ADR). Free bindings escape cost accounting = DoS on the fail-closed write path. |
| D9 | akb.yaml as behavioral config | **Yes, with unknown-key warnings** / flags-per-invocation (the `--no-commit` precedent) | **akb.yaml.** Flags are agent-hostile; one edit, many invocations. Unknown-key rejection/warning required (today's silent drop is a trap). |
| D10 | `type` as selector | Keep as the schema selector / untyped escape hatch | **Keep.** "Users decide what a valid entry is" is satisfied *by* the template being the user's declaration. Untyped pages = no validation driver; that's a product decision to defer, not an openness gap. |

## 5. What survives RFC-001 unchanged (carry-over list for RFC-002)

- P1 seam rule, P3 one-author-per-fact, P4 failure contract, G1/G2 coverage commitments
- JSON Schema 2020-12 + santhosh-tekuri recommendation (ADR 1)
- cel-go v0.32.x upgrade + ext libraries (ADR 3)
- Coexistence-over-translation rationale and all external research
- `read --json` / `write --json` CLI surface (extended: input=output replaces "modulo managed mutations")
- Merged error model, structured validation report (tech-spec)
- Round-trip invariant as pinned integration test
- The agent-flow answers from session 4 (Design A transport, CEL-never-mutates, two-schema derivation) — reframed under declared projections

## 6. Risks of the reframe

- **Scope explosion.** RFC-002 touches write path, lint, config, parsers. Mitigation: the
  phase plan must keep Phase 1 (schema+CEL coexistence core) independently shippable;
  openness demotions slot into phases by blast radius, not by enthusiasm.
- **Contract churn.** `lint --json` `by_check`, `is_draft` in `list --json`, index.md
  format — all potentially load-bearing for downstream harnesses (DESIGN.md typed gates).
  Each needs a stability statement in RFC-002.
- **"Permissive" vs "safe".** Demoting gates means untyped/invalid pages become possible;
  the sweep (lint) becomes the only net. D11 (index-rebuild ingestion bypass) already
  made the sweep load-bearing; RFC-002 must own that consequence explicitly.

## 7. Residual gaps (deferred, ADR-level)

- Embedded skill docs (kb-management) prose consumers of `created`/`updated` not audited.
- DESIGN.md typed-gate dependency cited secondhand (file not at cited path).
- `goldmark/wikilink` third-party ext grammar match for `[[display]](dest)` — verified
  absent; a fork/extension would be needed for option (a).
- testscript integration-suite churn for each demotion (not enumerated).
