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
- internal/template/**
status: final
summary: "External prior-art survey: no JSON-Schema→CEL compiler exists; K8s/protovalidate use coexistence; a declared-subset Go generator is the sensible path for schema→CEL, AST extraction only partial."
tags:
- json-schema
- cel
- prior-art
- validation
title: Feasibility of bridging JSON Schema ⇄ CEL for akb page validation
type: research
updated: "2026-09-26"
---
# Research: Feasibility of bridging JSON Schema <-> CEL for akb page validation

## Summary
No known tool compiles JSON Schema into CEL; the closest prior art (Kubernetes CRD validation, bufbuild/protovalidate) treats CEL as the *target* language with schemas either kept adjacent (K8s attaches CEL rules alongside the OpenAPI structural validation) or defined natively in proto with CEL used only for the "custom rules" escape hatch. Compiling a *declared subset* of JSON Schema (primitives, enum, required, bounds, pattern) into CEL is straightforward and well-supported by cel-go's stdlib + ext libraries; the reverse extraction from CEL ASTs is feasible only for a small invertible subset (comparisons, enum membership, size checks) using cel-go's walkable `cel.Ast` / protobuf `Expr`. A hand-rolled Go generator over a declared subset is the sensible path for direction (a); no drop-in library exists.

## Findings

### 1. Prior art for JSON Schema -> CEL

1. **Claim:** No published tool compiles JSON Schema into CEL. Multiple independent search rounds (GitHub/npm/PyPI/Go angle) returned only adjacent projects, none a jsonschema→cel compiler.
   **Sources:** Exa searches for "jsonschema to CEL expression compiler tool" and "json schema to CEL compiler" (results included [cel-python](https://github.com/cloud-custodian/cel-python), [protoc-json-schema](https://gitlab.com/andrewn/protoc-json-schema), [sourcemeta/jsonschema](https://github.com/sourcemeta/jsonschema) — all adjacent, none compiling schema→CEL). **Support:** direct evidence of absence (bounded — two query rounds; absence claims are weaker than presence claims). **Confidence:** medium.

2. **Claim:** Kubernetes chose CEL *alongside* the OpenAPI v3 schema (which carries most JSON-Schema-like structural validation) rather than compiling one into the other: `x-kubernetes-validations[].rule` CEL expressions provide cross-field/semantic validation against `self` (and `oldSelf` on updates). There is no KEP or doc that compiles the structural schema into CEL.
   **Sources:** [Common Expression Language in Kubernetes](https://kubernetes.io/docs/reference/using-api/cel/), [KEP-2876: CRD Validation Expression Language](https://www.kubernetes.dev/resources/keps/2876/), [Kubernetes 1.25 blog](https://kubernetes.io/blog/2022/09/23/crd-validation-rules-beta/). **Support:** direct evidence. **Confidence:** high.

3. **Claim:** protovalidate (Buf), successor to protoc-gen-validate, stores standard rules as proto metadata handled by fast native code in Go/Java/TS and uses CEL only for `(buf.validate.field).cel` / `(buf.validate.message).cel` custom rules referencing `this` — i.e., hand-written CEL expressions, not CEL compiled from a schema. protoc-gen-validate had the same shape.
   **Sources:** [bufbuild/protovalidate](https://github.com/bufbuild/protovalidate), [protovalidate custom CEL rules](https://protovalidate.com/schemas/custom-rules/), [A faster Protovalidate (Buf blog)](https://buf.build/blog/faster-protovalidate). **Support:** direct evidence. **Confidence:** high.

4. **Claim:** Ecosystem prior art for *generating* CEL rules from other sources exists — kube-rs `derive(CELSchema)` generates CEL validation from Rust type annotations; fluxcd/flux-schema validates Kubernetes YAML against JSON Schema and CEL rules side by side — but nothing generated from a standalone JSON Schema document.
   **Sources:** [kube-rs kube commit b104472d (derive CELSchema)](https://github.com/kube-rs/kube/commit/b104472d2672455771a2306f05614accbd700560), [fluxcd/flux-schema](https://github.com/fluxcd/flux-schema). **Support:** direct evidence (commit title / repo description). **Confidence:** medium-high.

5. **Claim:** cel-go ships ext libraries covering the main helper surfaces: `ext.Strings()` (charAt, indexOf, replace, split, substring, trim, join), `ext.Sets()` (sets.contains, sets.equivalent, sets.intersects), `ext.Math()` (math.greatest, math.least, bitwise, rounding), `ext.Lists()` (flatten, sort, reverse, distinct, range), enabled via `cel.NewEnv(ext.Strings(), ext.Sets(), ext.Math(), ext.Lists())`.
   **Sources:** [cel-go ext/README.md](https://github.com/google/cel-go/blob/master/ext/README.md), [ext/strings.go at v0.21.0](https://github.com/google/cel-go/blob/v0.21.0/ext/strings.go). **Support:** direct evidence. **Confidence:** high.

6. **Claim:** kube-cel (Rust) exposes `json_to_cel_with_schema(value, schema)` utilities that convert a JSON value into CEL-compatible value types respecting `format` hints (date-time, duration) — prior art for the *value coercion* half of a bridge, but it evaluates CEL; it does not *compile* a schema into CEL expressions.
   **Source:** [kube-cel values.rs](https://docs.rs/kube-cel/latest/src/kube_cel/validation/values.rs.html). **Support:** direct evidence via search summary of the source file. **Confidence:** medium.

### 2. CEL language capabilities relevant to schema validation

All from [cel-spec doc/langdef.md](https://github.com/google/cel-spec/blob/v0.13.0/doc/langdef.md) unless noted.

1. **Regex:** `matches` uses RE2 syntax with *substring* semantics: "Regular expressions follow the RE2 syntax. Regular expression matches succeed if they match a substring of the argument. Use explicit anchors (`^`/`$`) ... to force full-string matching." **Confidence:** high. Relevance: JSON Schema `pattern` is also substring search but in the EcmaScript dialect — see Contradictions.

2. **`size()` overloads:** `(string)`, `(bytes)`, `(list(A))`, `(map(A,B))` all → int. Directly maps `minLength`/`maxLength` and `minItems`/`maxItems`. **Confidence:** high.

3. **`has(e.f)` works on maps:** "If `e` evaluates to a map, then `has(e.f)` indicates whether the string `f` is a key in the map (note that `f` must syntactically be an identifier)." This is the direct equivalent of `required` / `dependentRequired`. Dynamic keys need `keys(m)` comprehensions instead. **Confidence:** high.

4. **Macro comprehensions:** `all(x,p)`, `exists(x,p)`, `exists_one(x,p)`, `map(x,t)`, `filter(x,p)` operate over list elements or map keys; `map`/`filter` are "not supported" when `e` is a map. Cel-spec documents single-variable comprehensions; **two-variable comprehensions are an extension** present in cel-go and Kubernetes 1.33+ (["CEL TwoVarComprehensions — Kubernetes versions 1.33+"](https://kubernetes.io/docs/reference/using-api/cel/)), not core CEL. Macros expand to Comprehension AST nodes at parse time, so their cost is size × predicate. **Confidence:** high.

5. **`dyn` typing:** "`dyn` is the union of all other types"; the `dyn()` stdlib function has no runtime effect and signals the type checker to widen types — essential for heterogeneous map validation like akb's `page`. **Confidence:** high.

6. **Timestamps/durations:** first-class abstract types with ordering (`<`, etc.), arithmetic (`Timestamp + Duration -> Timestamp`, `Duration - Duration -> Duration`), and range limits ("Duration ... roughly ±290 years"; timestamps limited to what is serializable as a string). Supports `format: date-time` partially — see impedance table. **Confidence:** high.

7. **Ternary `?:`:** `_?_:_ (bool, A, A) -> A`; "Will evaluate the test and only one of the remaining sub-expressions." Covers `if/then/else` and `not` via `!`. **Confidence:** high.

8. **Cost unit basis** (mirrors cel-go's estimator): comparisons cost 1; "list literal declarations have a fixed base cost of 40 cost units"; regex ops ≈ `length(regexString) * length(inputString)` (~0.25 factor per celbyexample's cel-go table); string functions ≈ length × 0.1; conditional = condition + max(branches). **Sources:** [Kubernetes CEL docs](https://kubernetes.io/docs/reference/using-api/cel/), [celbyexample execution-cost](https://celbyexample.com/execution-cost/). **Support:** direct quotes (K8s) + secondary corroboration. **Confidence:** high on model, medium on exact celbyexample coefficients (secondary source).

### 3. Impedance mapping table (researcher synthesis grounded in the cited docs)

Legend: **direct** = clean one-line CEL; **caveats** = expressible but with generator/cost/dialect concerns; **none** = no reasonable equivalent in cel-go stdlib+ext.

| JSON Schema keyword | CEL equivalent | Verdict |
|---|---|---|
| `type` | explicit runtime type checks emitted by generator (e.g. `type(page.f) == "string"` style helpers for scalars, dynamic checks / `size(page.f)` shape checks for map/list) | caveats (CEL's dyn typing means the generator must emit explicit runtime type checks; exact helper API varies across cel-go versions) |
| `enum` | `page.f in ["a","b"]` | **direct** |
| `const` | `page.f == "x"` | **direct** |
| `required` | `has(page.f)` (akb template `required` fields already checked at write time per project KB; CEL adds conditioned/dependent variants) | **direct** |
| `properties` | one predicate per property, joined with `&&` | **direct** |
| `additionalProperties: false` | `!keys(page).exists(k, !(k in ["f1","f2"]))` or `keys(page).all(k, k in [...])` | caveats (comprehension over `keys(page)`; cost-limit sensitive; should fuse with other key-set checks) |
| `items` | `page.list.all(x, <predicate per items schema>)` | caveats (nested comprehension if items itself has array/object constraints) |
| `minItems` / `maxItems` | `size(page.list) >= N` / `<= N` | **direct** |
| `minLength` / `maxLength` | `size(page.f) >= N` / `<= N` | **direct** |
| `pattern` | `page.f.matches("<pattern>")` with anchors added by generator | caveats (RE2 vs EcmaScript dialect: no lookaround in RE2, some Unicode behavior differs; both are substring-based so the anchoring half ports cleanly, but not all regexes themselves do — validate each ported pattern against Go RE2) |
| `minimum` / `maximum` (+ exclusive*) | `page.f >= N`, `<= N`, `> N`, `< N` | **direct** (generator concern: ensure numeric typing before comparison) |
| `$ref` / `$defs` | none — CEL has no function abstraction; generator must inline the referenced schema at each use site | caveats (expression size blowup; must detect recursive schemas and refuse) |
| `allOf` | conjunction `&&` of sub-predicates | **direct** |
| `anyOf` | disjunction `||` | **direct** |
| `oneOf` | arity-checked: `((a) ? 1 : 0) + ((b) ? 1 : 0) == 1`, or a `filter(...).size() == 1` comprehension | caveats (comprehension form is more cost-expensive) |
| `not` | `!(sub-expr)` | **direct** |
| `if` / `then` / `else` | `(cond) ? (then-expr) : (else-expr)`; when no `else`: `!(cond) || (then-expr)` | **direct** |
| `format: date-time` | no "validate parse cleanly" primitive in CEL; options: rely on akb's frontmatter RFC3339 coercion (per project KB, date-shaped strings become `time.Time` before CEL) plus `timestamp(page.f)` used inside some inequality as a canary, or enforce the format at parse time outside CEL | caveats (best done as akb frontmatter-stage coercion, not CEL) |
| `format: uri` | no stdlib; `ext.Strings` has no URI validator; Kubernetes ships a separate URL library not in core cel-go | **none** in cel-go core (regex approximation is lossy; better handled outside CEL — akb already parses wikilinks structurally) |
| `dependentRequired` (if `p` present then `q` required) | `!has(page.p) || has(page.q)`, chained per dependency pair (equivalent ternary form: `has(page.p) ? has(page.q) : true`) | **direct** |

**Cost-limit interaction with nesting (important against akb's 100000 cap).** Comprehension/macro cost ≈ collection size × predicate cost, and nested comprehensions multiply that — so `items.all(x, <predicate containing another comprehension>)` is the pattern most likely to blow a fixed budget. Without size bounds, cel-go's cost estimator must assume a worst-case input size; Kubernetes solved this by requiring CRD authors to provide `maxLength` "on all variable length data types that they iterate across in CEL expressions so we know exactly how to compute cost" ([KEP-2876](https://www.kubernetes.dev/resources/keps/2876/)), and the K8s docs state that defining `maxItems`/`maxProperties`/`maxLength` gives the cost estimator the bounds it needs ([Kubernetes CEL docs](https://kubernetes.io/docs/reference/using-api/cel/)). For comparison, Kubernetes enforces `StaticEstimatedCostLimit = 10_000_000` per expression and `StaticEstimatedCRDCostLimit = 100_000_000` per CRD ([apiextensions validation.go](https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apis/apiextensions/validation/validation.go)) — akb's 100000 cap is 100× tighter per rule, which is fine for simple frontmatter pages (few fields, short strings: e.g. a regex costing `len*patlen*0.25` on a 50-char string with a 30-char pattern ≈ 375 cost units) but leaves little headroom for comprehension-heavy patterns (`oneOf`-as-`filter`, `additionalProperties` over `keys(page)`); mitigation is for the generator to fuse per-keyword comprehensions where possible (e.g., one `keys(page).all(...)` handling both required and no-extras). Comprehension *nesting depth* has no default limit in cel-go — enforcement exists only via the opt-in `cel.validator.comprehension_nesting_limit` validator ([cel-go issue #1050](https://github.com/google/cel-go/issues/1050), [PR #1196](https://github.com/cel-expr/cel-go/pull/1196)); with akb's cost cap coverage, akb likely doesn't need it separately.

### 4. Reverse direction (CEL -> JSON Schema)

1. **Claim:** cel-go's parsed representation is walkable, making a subset-extractor an AST walker rather than a language tool. `cel.Ast` carries the expression (via `NativeRep`/`common.ast.Ast` since the native-AST migration); `cel.AstToParsedExpr` / `cel.AstToCheckedExpr` (cel/io.go) expose the `cel.expr.conformance.v1.ParsedExpr` protobuf whose `Expr` node has an `ExprKind` oneof (`IdentExpr`, `SelectExpr`, `CallExpr`, `ListExpr`, `StructExpr`, `ComprehensionExpr`) plus `SourceInfo` for positions.
   **Sources:** [cel-go cel/io.go](https://github.com/google/cel-go/blob/beb275fdd166ff94174913b5ee805dfe5b4216a3/cel/io.go), [cel-spec syntax.pb.go](https://github.com/google/cel-spec/blob/master/syntax.pb.go), [cel-go issue #789 (native AST migration)](https://github.com/google/cel-go/issues/789). **Support:** direct evidence for the AST being walkable; the *inversion tool* is researcher inference. **Confidence:** high.

2. **Claim:** No prior art found for CEL→JSON Schema extraction tools. The nearest prior art is Kubernetes' [Declarative Validation KEP-4153](https://www.kubernetes.dev/resources/keps/4153/), which generates both CEL rules and equivalent validation logic from a single Go-struct-tag source — a shared-source approach, not parsing CEL back.
   **Source:** [KEP-4153](https://www.kubernetes.dev/resources/keps/4153/). **Support:** direct evidence of KEP-4153; interpretation that no back-parser exists. **Confidence:** medium.

3. **Claim:** Only a constrained subset of CEL is invertible; arbitrary CEL is not. Division (researcher analysis, grounded in the walkable AST above):
   - **Invertible:** `page.f == <literal>` → `const`; `page.f in [...]` → `enum`; type checks → `type`; `page.f >= / <= / > / <` with literals → `minimum` / `maximum` (+ exclusive); `size(page.f) >= / <= N` → `minLength` / `maxItems` depending on f's apparent type; `page.f.matches(...)` → `pattern` (dialect caveat); `has(page.f)` / conjunctions of the above → `required` / `properties`.
   - **Not invertible in general:** free-form boolean algebra mixing unrelated concerns; ternaries encoding `if/then/else`; string equality used as a flow-control guard; cross-field comparisons (`page.a <= page.b` — JSON Schema has no pairwise predicate); comprehensions over computed derivations; `matches` used as a semantic rule (e.g., branch-name shape); references to `old_page` (akb's transition style); any ext-library call. An extractor over these must either emit lossy `x-` extension annotations or refuse.
   **Support:** researcher inference. **Confidence:** high on the general claim (CEL is a general-purpose expression language over maps; it can express predicates JSON Schema cannot name); medium on any specific keyword map until prototyped.

4. **Claim (inference):** For akb, CEL→Schema extraction is worth doing only as a *reporting/lint aid* ("this rule is also expressible as these schema keywords"), not as the canonical source of truth — CEL rules are strictly more expressive than JSON Schema, and extraction silently drops anything outside the whitelist.
   **Confidence:** medium (design judgment; flagged as researcher inference).

### 5. Recommendation for direction (a)

1. **Claim:** A hand-rolled Go generator over a declared subset is the sensible path; no library exists to drop in. The ecosystem's architectural analogy supports this: protovalidate handles standard rules natively and reserves CEL for hand-written custom rules; Kubernetes keeps the structural schema and CEL rules side by side rather than compiling one into the other — the consensus design is coexistence, not compilation.
   **Sources:** absence-of-prior-art evidence in §1; [protovalidate](https://github.com/bufbuild/protovalidate); [KEP-2876](https://www.kubernetes.dev/resources/keps/2876/). **Support:** interpretation. **Confidence:** high on the recommendation's shape, medium on the effort estimate.

2. **Recommendation detail (researcher reasoning, labeled):**
   - Cover the subset only: `type`, `const`, `enum`, `required`, `properties`, `items` (one level), `min/maxItems`, `min/maxLength`, `pattern`, `minimum`/`maximum` (+exclusive), `allOf`, `anyOf`, `not`, `if/then/else`, `dependentRequired`. Refuse-with-clear-error on: recursive `$defs`; regex features RE2 doesn't support (lookaround; validate each ported pattern with Go's RE2-aware `regexp.Compile` and reject on failure); exotic `format`s like `uri` (fall through to akb lint checks rather than CEL); deeply nested items-in-items (or support one level given the 100000 cost cap).
   - Guard direction (b) the same way: declare an invertible whitelist — comparisons, `in`, `size()`, `matches`, `has` on `page.*` identifiers with literal or list-literal operands — and refuse everything else with "cannot express as JSON Schema", mirroring akb's fail-closed write-time CEL eval error philosophy (project KB).
   - Reuse cel-go machinery rather than hand-building: `cel.Ast`/`AstToParsedExpr` for (b); akb's existing env + `CompileRule`/`Evaluate` for (a). For budget safety, replicate the protovalidate/K8s pattern of checking `EstimatedCost` on the generated rule at authoring time, then enforcing `cel.CostTracking()` + `cel.CostLimit(...)` at runtime (akb already enforces a 100000 cost cap in `CompileRule`).
   - Optional shape worth considering, labeled **inference**: instead of a full JSON-Schema→CEL compiler, akb could extend TemplateV2 with a `schema:` block natively understood by the CEL engine — schema-like ergonomics without a full compiler. No cited tool supports this; it's a design suggestion.

## Contradictions
- **RE2 vs JSON Schema `pattern` dialect:** cel-spec confirms RE2 with substring semantics for `matches`; JSON Schema `pattern` uses EcmaScript regex with substring semantics. Both are substring, but the dialects differ (RE2 has no lookaround; Unicode behavior differs). A generator that copies patterns verbatim without an RE2 compatibility check will produce subtly-broken validations. Not a contradiction in the evidence, but a direct, decision-relevant caveat.
- **`has(e.f)` on maps:** the langdef is explicit that `has(e.f)` tests map key presence when `e` evaluates to a map (key must be a syntactic identifier). No contradiction found; flagged as a targeted integration test for akb's `page` map (K8s' later "CEL Optional Values" feature changes some edge semantics, worth one integration test).
- None otherwise: protovalidate and K8s docs consistently frame CEL as an *additional* validation system, not a JSON Schema compiler.

## Missing evidence
- Whether `bufbuild/protoschema-plugins` (which generates JSON Schema from proto/protovalidate rules) emits protovalidate CEL rules into the generated schema or drops them — unverified; if it emits them, it would be direct prior art for direction (b).
- Exact per-keyword cost benchmarks against akb's 100000 cap for representative templates — not measured; needs a small prototype harness using cel-go's `EstimatedCost` API.
- Precise cel-go version pinned in akb's go.mod and whether it includes two-var comprehensions / current ext libs — not verified in this research run.
- Exhaustive survey for a lesser-known jsonschema→cel tool beyond two search rounds — absence evidence is bounded.

## Sources
- Kept: [cel-spec doc/langdef.md](https://github.com/google/cel-spec/blob/v0.13.0/doc/langdef.md) — normative source for macros, RE2 `matches`, `size()`, `has()`, `dyn`, timestamp/duration, ternary.
- Kept: [cel-go ext/README.md](https://github.com/google/cel-go/blob/master/ext/README.md) — canonical reference for ext.Strings/Sets/Math/Lists.
- Kept: [Common Expression Language in Kubernetes](https://kubernetes.io/docs/reference/using-api/cel/) — production deployment story, feature table, cost-unit model (comparisons = 1, list literal base = 40, regex ≈ len(regex)×len(input)).
- Kept: [KEP-2876: CRD Validation Expression Language](https://www.kubernetes.dev/resources/keps/2876/) — design rationale for adjacent schema+CEL; why `maxLength`/`maxItems` are needed for cost tractability.
- Kept: [apiextensions validation.go](https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apis/apiextensions/validation/validation.go) — exact constants `StaticEstimatedCostLimit = 10000000`, `StaticEstimatedCRDCostLimit = 100000000`.
- Kept: [bufbuild/protovalidate](https://github.com/bufbuild/protovalidate) and [custom CEL rules](https://protovalidate.com/schemas/custom-rules/) — successor design to protoc-gen-validate; reference architectural pattern of "native standard rules + CEL escape hatch".
- Kept: [cel-go issue #1050](https://github.com/google/cel-go/issues/1050) and [PR #1196](https://github.com/cel-expr/cel-go/pull/1196) — comprehension nesting is not limited by default; opt-in nesting validator exists.
- Kept: [celbyexample execution-cost](https://celbyexample.com/execution-cost/) — concrete cost figures per operator family, corroborating the cost model (secondary source).
- Kept: [Kubernetes 1.25 CRD validation blog](https://kubernetes.io/blog/2022/09/23/crd-validation-rules-beta/) — concise framing of CEL rules living alongside JSON Schema.
- Kept: [kube-rs derive(CELSchema) commit](https://github.com/kube-rs/kube/commit/b104472d2672455771a2306f05614accbd700560) — nearest prior art for generating CEL rules from another source-of-truth.
- Kept: [cel-go cel/io.go](https://github.com/google/cel-go/blob/beb275fdd166ff94174913b5ee805dfe5b4216a3/cel/io.go) and [cel-go issue #789](https://github.com/google/cel-go/issues/789) — the `AstToParsedExpr`/protobuf machinery a reverse extractor would target.
- Kept: [fluxcd/flux-schema](https://github.com/fluxcd/flux-schema) — a validator consuming JSON Schema and CEL side by side; evidence the coexistence pattern is the ecosystem default.
- Kept: [kube-cel values.rs](https://docs.rs/kube-cel/latest/src/kube_cel/validation/values.rs.html) — prior art for JSON value ↔ CEL value bridging with format awareness.
- Rejected/deprioritized: [DeepWiki cel-expr/cel-spec](https://deepwiki.com/cel-expr/cel-spec) — secondary AI-generated aggregation of the normative spec; cited primary langdef.md instead.
- Rejected/deprioritized: [JIT-compiling CEL (Taichi's Blog)](https://taichimaeda.github.io/posts/jit-compiling-cel/) — performance context, unrelated to feasibility.
- Rejected/deprioritized: [cloud-custodian/cel-python](https://github.com/cloud-custodian/cel-python) — CEL evaluator in Python, not a JSON Schema compiler; kept only as "adjacent prior art exists" citation.

## Next steps
- Prototype the forward mapping on one akb template: hand-write CEL for a small outputSchema, then implement the declared-subset generator and diff against the hand-written rules, running each through akb's `CompileRule` to confirm the 100000 cost cap holds.
- Check `bufbuild/protoschema-plugins` line-by-line for whether protovalidate CEL is emitted into the generated JSON Schema — direct prior art for direction (b) if yes.
- Stress-test the 100000 cost limit with a worst-case generated rule (nested `items` + `oneOf` + `additionalProperties: false`) via cel-go `EstimatedCost` before committing to a generator design.
