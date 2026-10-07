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
summary: "De-duplicated second survey: no library compiles JSON Schema into CEL; reverse extraction is provably partial; recommends a hand-rolled Go generator over a declared subset, display-only AST extractor."
tags:
- json-schema
- cel
- prior-art
- validation
title: JSON Schema ⇄ CEL bridge feasibility for akb template validation (duplicate survey)
type: research
updated: "2026-09-26"
---
# Research: JSON Schema ⇄ CEL bridge feasibility for akb template validation

Scope: (a) compile JSON Schema (e.g. pi-subagents `outputSchema`) into CEL validation
expressions; (b) extract a JSON Schema back out of existing CEL rules.

## Summary

No library compiles JSON Schema into CEL. The two ecosystems facing this problem
(Kubernetes CRDs, protovalidate) both keep a declarative vocabulary as the primary mechanism
and use CEL only for cross-field/arithmetic constraints, with CEL attached as an *extension
keyword* rather than generated from the schema ([K8s CEL][k8s-cel], [KEP-5073][kep5073],
[Protovalidate][pv-cel]). The only shipped generator that emits CEL generates it from typed
rule annotations, not JSON Schema. Reverse direction (b) has real prior art only for the
typed-annotation form ([cerbos][cerbos-js], [buf][buf-ps]); extracting schema from arbitrary
CEL is unsolved and provably partial. Recommendation: hand-rolled Go generator over a
declared subset for (a), rejecting out-of-subset keywords rather than approximating them, and
a display-only lossy AST extractor for (b).

## Findings

1. **Kubernetes adds CEL *beside* the schema, never derived from it.**
   `x-kubernetes-validations` is a schema extension keyword; official guidance says *"Use
   OpenAPIv3 value validations (`maxLength`, `maxItems`, `maxProperties`, `required`, `enum`,
   `minimum`, `maximum`, ..) and string formats where available"* and reserves CEL for
   cross-field work. **Sources:** [kubernetes.io CEL][k8s-cel], [CRD validation rules][k8s-crd].
   **Support:** direct evidence. **Confidence:** high.

2. **Kubernetes' newer declarative validation again chose typed tags over CEL.** KEP-5073's
   `validation-gen` uses `+k8s:minLength`, `+k8s:enum`, `+k8s:required`, `+k8s:minimum`, …
   and treats CEL as a last-resort escape hatch: *"The goal is to use dedicated IDL tags for
   the vast majority of validations, ensuring that CEL is reserved for exceptional cases if
   used at all."* Its tag→OpenAPI-keyword table is effectively a hand-maintained impedance
   map. **Sources:** [KEP-5073][kep5073], [declarative validation docs][dv-docs].
   **Support:** direct evidence. **Confidence:** high.

3. **The closest "declarative rules → CEL" system has no JSON Schema input.** protovalidate
   authors every standard rule as a CEL expression in `validate.proto`
   (`(predefined).cel.expression`), e.g. the UUID rule
   `"!rules.uuid || this == '' || this.matches('^[0-9a-fA-F]{8}-…$')"`, and compiles those at
   runtime. **Sources:** [Protovalidate — how CEL works][pv-cel],
   [protovalidate-go cel/library.go][pv-lib]. **Support:** direct evidence. **Confidence:** high.
   **Inference (mine):** its generation step is *message field → CEL string*, far narrower
   than *JSON Schema → CEL*.

4. **cel-go itself has no schema support.** No `jsonschema` package in the repo tree;
   extensions are Bindings / Encoders / Math / Protos / Optionals / Lists / Sets / Strings /
   Regex. PGV is the superseded predecessor of protovalidate (*"the next generation of
   protoc-gen-validate"*), not a CEL compiler. **Sources:** [cel-go][celgo],
   [ext README][celgo-ext], [protovalidate README][pv-readme]. **Support:** direct evidence
   (repo tree, ext headings, README sentence). **Confidence:** high.

5. **Adjacent designs worth borrowing from, none of which compiles schema → CEL.**
   kube-cel (Rust) evaluates CRD `x-kubernetes-validations` client-side and uses the schema
   only to *type-convert JSON values* (`format: date-time` → Timestamp); flux-schema extracts
   the same rules from JSON Schemas and evaluates them; helm-cel replaces
   `values.schema.json` with hand-written CEL, and `helm cel generate` scaffolds rules from a
   *values file*, not a schema; Datree embeds CEL inside JSON Schema as a `CELDefinition`
   keyword combined with `if`/`then`. **Sources:** [kube-cel][kube-cel],
   [flux-schema][flux-schema], [helm-cel][helm-cel], [Datree][datree]. **Support:** direct
   evidence from each README (item iii's flag semantics read, not executed). **Confidence:** high.

6. **CEL macros, typing, error behaviour.** `has(e.f)`, `all`, `exists`, `exists_one`, `map`
   (2- and 3-arg), `filter`, ternary, `size()` on string/bytes/list/map are standard, and
   macros are enabled by default in cel-go. Error semantics differ: `all` combines with `&&`
   and `exists` with `||`, so they *absorb* element errors; `filter`/`map` *raise* on any
   element error; `exists_one` deliberately does not short-circuit.
   **Sources:** [cel-spec langdef][langdef], [cel-go README][celgo]. **Support:** direct
   evidence (verbatim quotes). **Confidence:** high. **Inference (mine):** generated rules
   should prefer `all`/`exists` (with an explicit `has()` guard inside the predicate) over
   `filter`/`map`, because absent keys in akb's `page` map evaluate to errors.

7. **CEL strings are RE2 and unanchored, like JSON Schema `pattern`.** *"Regular expressions
   follow the RE2 syntax. Regular expression matches succeed if they match a substring of the
   argument."* JSON Schema says the same about anchoring, but its dialect is ECMA-262.
   **Sources:** [langdef][langdef], [JSON Schema 2020-12][jsv2020]. **Support:** direct
   evidence. **Confidence:** high.

8. **CEL typing: `dyn` + gradual typing; typed nested objects need a custom `TypeProvider`.**
   *"The type checker also introduces the `dyn` type, which is the union of all other
   types"*; `dyn(x)` is a no-op at runtime. **Sources:** [langdef][langdef]. **Support:**
   direct evidence. **Confidence:** high. **Corroboration:** decree's design doc records that
   typing a nested `self` object "would need a custom `cel.TypeProvider`", so its env stays
   `dyn` and type checking is a hand-rolled AST pass ([decree][decree]) — akb's `page`-map
   situation exactly.

9. **CEL ext covers the mechanical cases but not formats.** `ext.Strings`, `ext.Lists`
   (`distinct`, `first`/`last`), `ext.Sets` (`contains`/`equivalent`/`intersects`),
   `ext.Math` (`greatest`, `least`, `abs`, `sign`, `isNaN`, `isInf`, `ceil`, `round`…),
   `ext.Regex` (`regex.replace` etc., `regex` namespace, needs `cel.OptionalTypes()`).
   **Sources:** [ext README][celgo-ext]. **Support:** direct evidence. **Confidence:** high.
   **Gap:** no stdlib uuid/email/uri/hostname validator — protovalidate had to add
   `isHostname`, `isNan`, `isInf`, `unique`, `getField` itself ([pv-lib][pv-lib]).

10. **Cost is specified and enforceable.** cel-spec gives complexity classes (`matches` and
    `contains` O(|a|·|b|); list `in` O(|a|·|b|); map index/select O(n·m); conditional
    evaluates one branch; list index constant). cel-go adds a runtime tracker plus a static
    estimator, exposed as `cel.CostLimit(uint64)`, `cel.CostTracking`,
    `cel.CostTrackerOptions`, `cel.CostSizingStrategy`; the estimator's comprehension rule is
    `iterRange + accuInit + size(iterRange) × (loopCondition + loopStep) + result`.
    **Sources:** [langdef][langdef], [runtimecost.go][runtimecost], [checker/cost.go][checkercost],
    [cel-go options.go][celgo-opt]. **Support:** direct evidence (identifiers and formula
    quoted from source). **Confidence:** high.

### Impedance map (synthesis; verdicts are my judgment on the evidence above)

**D** direct · **C** possible with caveats · **N** no reasonable equivalent

| JSON Schema keyword | Verdict | CEL shape / caveat |
|---|---|---|
| `type` | C | `type(x) == string`, but values arrive as `dyn` from YAML (int64 vs float64 depends on decode), a `null` value is indistinguishable from an absent key, and there is no reflective "is object shaped like X". |
| `enum` / `const` | D | `x in ['a','b']` (with `has()` guard; list `in` is O(n)) / `x == 'literal'`. |
| `required` | C | `has(page.x)` — usually redundant, since akb enforces template `required: true` before CEL runs. |
| `properties` | C | Generate path-qualified rules (`page.a.b`); no schema recursion. |
| `additionalProperties: false` | C | `page.keys().all(k, k in [...])` per level; needs the sibling key set known at generation time. |
| `items` (single schema) | C | `page.tags.all(i, i.size() <= 20)` — `all` absorbs errors, which is what you want. |
| `prefixItems` / draft-07 tuple `items` | C | Index selectors plus a `size() == N` guard. |
| `additionalItems: false` | C | `size(x) <= N`. |
| `minItems`/`maxItems`, `minimum`/`maximum` | D | `size(x) >= n`, `x <= n`. |
| `minLength`/`maxLength` | C | `size(x)` — verify cel-go's string unit against JSON Schema's "number of its characters per RFC 8259" (Missing evidence). |
| `pattern` | C | `x.matches('…')`; RE2 ≠ ECMA-262: lookaround/backreferences must be rejected at generation time, not reinterpreted. Anchoring matches on both sides. |
| `exclusiveMinimum`/`exclusiveMaximum` | D | `x > n` / `x < n` in 2020-12's numeric form; draft-04's boolean form has no standalone equivalent (C). |
| `multipleOf` | C / N | Integers: `x % n == 0` (CEL `%` is int/uint only). Non-integer `multipleOf` (0.1): N. |
| `$ref`/`$defs` | C / N | Flatten at generation time (rule duplication); recursion is **N** — KEP-2876: CEL *"lacks support recursive data types (OpenAPIV3 & CRD structural schemas are not possible to validate with CEL)"*. `cel.bind` gives intra-expression reuse only. |
| `allOf` / `anyOf` | D | `&&` / `\|\|` of translated subschemas. |
| `oneOf` | C | `exists_one` over subschema predicates, but "exactly one" interacts with `additionalProperties`/`unevaluatedProperties`; naive translation is unsound (cerbos documented needing `unevaluatedProperties` for its `oneof` case). |
| `not` | D | `!(…)` for pure-assertion subschemas. |
| `if`/`then`/`else` | D | implication: `(ifC ? thenC : true) && (!ifC ? elseC : true)`; missing branches default to `true`. |
| `dependentRequired` | D | `!has(page.b) \|\| has(page.a)` — decree maps exactly this shape (`has(x) implies has(y)`). |
| `dependentSchemas` | C | Conditionally `&&`-in the translated subschema on the containing object. |
| `propertyNames` | C | `page.keys().all(k, k.matches('…'))`. |
| `contains` (+`minContains`/`maxContains`) | C | `x.filter(e, P).size() >= n`, but `filter` raises on element errors and costs O(n·|P|); prefer `exists`/counting via `all`. |
| `uniqueItems` | C | `x.size() == x.distinct().size()` (needs `ext.Lists`) or `sets.equivalent`. |
| `format: date-time` | C | Only if the value is a CEL timestamp (akb converts date-parseable frontmatter to `time.Time`); otherwise regex. Kubernetes maps it to a Timestamp *type*, not a check. |
| `format: email/uri/uuid/hostname/ipv4` | N / C | No stdlib support; regex approximations are not the spec (protovalidate had to add custom functions). |
| `format` as annotation | — | Spec: annotation-only unless enabled, and *"MUST be disabled by default"* — turning `format` into a hard rule changes semantics. |
| `unevaluatedProperties`/`unevaluatedItems` | N | Needs subschema-application accounting; Kubernetes refused the analogous class for CEL. |
| `default`, `title`, `description`, `examples`, `readOnly`, `writeOnly`, `$comment` | — | Annotation-only; drop, don't emulate. |

**Cost vs akb's 100000 limit.** Kubernetes sets `StaticEstimatedCostLimit = 10000000` *per
expression* and `StaticEstimatedCRDCostLimit = 100000000` per CRD — akb's runtime
`cel.CostLimit` of 100000 is 100× tighter than even K8s' per-expression budget (direct
evidence, [apiextensions validation.go][k8s-cost]). Comprehension cost scales as
`size(iterRange) × (loopCondition + loopStep)`, so nested comprehensions over unbounded lists
are the failure mode; Kubernetes' remedy is to *add schema size bounds*, because *"the cost
system is aware of size limits declared in the CRD's schema"*, and without them the worst case
is assumed and rules get rejected ([K8s CRD CEL][k8s-crd], [k8s#121162][k8s121162]).
**Inference (mine):** (i) having the JSON Schema is an *asset* here — it supplies the
`maxItems`/`maxLength` bounds that keep estimated cost low; (ii) a leading size guard stops a
*failing* page from paying comprehension cost (short-circuit) but does not reduce a *passing*
page's cost, so it is not a substitute for bounded lists; (iii) gate generated rules with
cel-go's static estimate at template-write time, which fits akb's fail-closed write policy.

### Reverse direction: CEL → JSON Schema

11. **The real prior art converts typed annotations → JSON Schema and drops CEL.** cerbos
    `protoc-gen-jsonschema` reads `buf.validate.FieldRules` and emits a closed vocabulary:
    objects (`maxProperties`, `minProperties`, `required`, `additionalProperties`,
    `properties`, `propertyNames`) and strings (`const`, `enum`, `maxLength`, `minLength`,
    `pattern`, `format`); `schemaForOneOf` fires only when `rules.Required == true`. Nothing
    in `internal/module/message.go` reads `rules.GetCel()`. Stated limitation: *"the schemas
    don't attempt to validate that only one field of a `oneof` is set; to do so efficiently
    whilst rejecting unknown fields requires `unevaluatedProperties`"*.
    **Sources:** [cerbos repo][cerbos-js] (`internal/jsonschema/*.go`, `internal/module/message.go`),
    [PR #727][cerbos-pr]. **Support:** direct evidence from source; "no CEL handling" is an
    observation about the file read. **Confidence:** high.

12. **buf's `protoschema-plugins` does the same from buf.validate standard rules**
    (`required` → `required`, `gte`/`lte` → `minimum`/`maximum`, `finite`), with an open
    discussion on `oneof` → `oneOf` that names the gap: *"is the issue that we have 'at most
    one' in proto and 'exactly one' in JSON schema?"* **Sources:** [protoschema-plugins][buf-ps],
    [issue #109][buf-ps-109]. **Support:** direct evidence (README output, issue quote).
    **Confidence:** high.

13. **A constrained AST-pattern extractor is a proven design, with its limits stated.**
    decree's schema importer walks the CEL AST and recognises substitutable shapes:
    `field == literal` and `field == lit || field == lit …` → `const`/`enum`; single-field
    `<>=` chains → `minimum`/`maximum`; `has(x) implies has(y)` → `dependentRequired`; and
    *"Anything else passes"*. It states *"SMT is out of scope"*, that false negatives are
    acceptable and false positives are not, and that it is a hand-rolled AST pass precisely
    because a typed env would need a custom `cel.TypeProvider`. **Sources:**
    [decree `cel-validation.md`][decree]. **Support:** direct evidence (quoted), but a
    design/plan document, i.e. asserted intended behaviour rather than an audit of shipped
    code. **Confidence:** medium-high.

14. **The CEL AST is fully walkable in Go, so an extractor is mechanical.** `common/ast`
    exposes `NavigateAST(ast) NavigableExpr`, `NavigableExpr` (`Kind`, `Type`, `Parent`,
    `Children`, `Depth`, typed `AsCall()`/`AsSelect()`/`AsComprehension()`), and
    `PreOrderVisit`/`PostOrderVisit`/`MatchDescendants`. cel-spec's `syntax.proto` `Expr` is
    the wire form; precedent for "AST → another declarative language" is `cel2sql`, which
    turns a compiled CEL `ast` into a SQL WHERE clause and ships `WithMaxDepth`/
    `WithMaxOutputLength` guard rails. **Sources:** [cel-go common/ast][celgo-ast],
    [langdef][langdef], [cel2sql][cel2sql]. **Support:** direct evidence (identifiers and
    options quoted from source/docs). **Confidence:** high.

15. **(b) assessment: invertible only for a constrained subset.** Single-field
    comparisons/ranges, enum/const membership, `size()` bounds, `has()` implications and
    single-predicate comprehensions over one field are recoverable. Arbitrary boolean logic,
    multi-field arithmetic, ternaries across paths and nested comprehensions are not: the
    result would be an opaque "unrepresentable rule" marker or an unreadable
    `anyOf`-of-`if/then` blob. **Support:** researcher inference, grounded in the fact that
    both Kubernetes and decree keep CEL *alongside* declarative keywords rather than
    translating between them ([kep5073][kep5073], [decree][decree]). **Confidence:** medium-high.

### Recommendation for (a): hand-rolled generator over a declared subset

16. **Claim:** no library to adopt; a hand-rolled Go generator is the sensible path.
    **Support:** researcher inference from Findings 1–5 and 11–12 — both shipped rule→CEL
    generators source from typed annotations, cel-go has no schema package, and akb's `page`
    shape/rule storage differ from every surveyed target. Searched, not proven absent.
    **Confidence:** medium-high.

Design opinions (mine, not source claims):

- **Two layers, not one.** Emit structural keywords into akb's template `schema.frontmatter`
  (already enforced before CEL), and generate CEL only for cross-field arithmetic,
  conditionals, per-element bounds, enum/pattern and exclusive selection — KEP-5073's
  "dedicated tags first, CEL escape hatch" conclusion.
- **Declared subset with hard failure.** Reject `unevaluatedProperties`, recursive `$ref`,
  ECMA-262 lookaround and non-integer `multipleOf` with a diagnostic naming keyword+path, and
  skip that subtree (cerbos/buf accept and document partial support; decree warns and skips).
- **Presence-guard every dereference** — `has(page.x) ? <rules> : true` — since akb's
  write-time CEL errors fail closed.
- **Cost gate at template-write time** using cel-go's static estimate against the 100000
  budget; raise the limit toward K8s' per-expression 10000000 only if generated
  comprehensions are unavoidable.
- **Consider not compiling at all.** Because a JSON Schema is already declarative, an
  `x-cel-validations`-style extension keyword on `outputSchema` (Kubernetes
  `x-kubernetes-validations`, Datree `CELDefinition`) delivers the same capability with no
  compiler, no dialect translation and no lossy reverse direction.

## Contradictions

- **"Kubernetes uses CEL for CRD validation" vs "Kubernetes uses JSON Schema keywords."**
  Both true and complementary: OpenAPI keywords carry structure, `x-kubernetes-validations`
  carries CEL. This is why schema→CEL compilation has no upstream precedent.
- **protovalidate both "is CEL rules" and "is consumed as typed annotations."** Resolution:
  the CEL is a compiled artifact of the annotation message, so tooling targets the source of
  truth — consistent, and it strengthens Findings 11–13.
- **`format` semantics differ by ecosystem:** JSON Schema annotation-only by default;
  protovalidate treats `string.uuid` as a hard assertion. A bridge must choose per format.

## Missing evidence

- **No JSON-Schema→CEL implementation found** — but only keyword/LLM-Exa searches were run;
  GitHub code search (`jsonschema cel path:*.go`) and registry sweeps were not. Treat as
  "searched and not found".
- **cel-go `size(string)` unit** vs JSON Schema's character count: unverified; `minLength`
  could be off for non-BMP characters. Needs a fixture test.
- **`has()` on raw CEL map keys:** langdef documents message-field presence and calls other
  cases errors; K8s and protovalidate use `has(...)` on structural/object fields, so it works
  in practice — no spec sentence found that blesses plain maps. Verify in cel-go.
- **cel-go's default `CostSizingStrategy`** and how much an unbounded list inflates the
  static estimate in akb's env (K8s' 10⁷/10⁸ are Kubernetes policy, not cel-go defaults).
- **Whether akb's CEL env enables `ext.Lists`/`ext.Sets`/`ext.Regex`** — `uniqueItems` and
  `regex.*` rows depend on it.
- **pi-subagents `outputSchema` draft** (2020-12 vs draft-07) assumed, not verified;
  `prefixItems`/`$defs`/`dependentRequired` presence changes which rows matter most.
- **`source_check` was unavailable this run** (not registered by the provider), so
  decision-critical claims were validated by fetching and reading primary sources directly
  (specs, Go source, KEP text) instead of an independent checker.

## Sources

- Kept: [cel-spec langdef][langdef] (macros, RE2/substring, cost classes, gradual typing) ·
  [cel-go][celgo], [ext README][celgo-ext], [options.go][celgo-opt], [common/ast][celgo-ast],
  [runtimecost.go][runtimecost], [checker/cost.go][checkercost] · [K8s CEL][k8s-cel],
  [CRD rules][k8s-crd], [KEP-2876][kep2876], [KEP-5073][kep5073],
  [declarative validation][dv-docs], [cost constants][k8s-cost], [k8s#121162][k8s121162] ·
  [Protovalidate CEL][pv-cel], [README][pv-readme], [protovalidate-go cel/library.go][pv-lib] ·
  [cerbos protoc-gen-jsonschema][cerbos-js] + [PR #727][cerbos-pr] ·
  [buf protoschema-plugins][buf-ps] + [issue #109][buf-ps-109] ·
  [decree cel-validation.md][decree] · [JSON Schema 2020-12][jsv2020] · [cel2sql][cel2sql].
- Rejected/deprioritized: `@cleverbrush/schema-json`, `json-schema-to-zod`,
  `traversable/schema-to-json-schema`, `schema-gen` (JSON Schema ⇄ TS/Rust builders, no CEL);
  Datree hub docs (used only for the "CEL inside JSON Schema" precedent; stale vendor docs);
  `cel-python`, `cel-php`, JIT-CEL blog, `protoc-gen-cel-validate` (other runtimes/IDLs, not
  used for any decision-relevant claim).

## Next steps

1. **Spike the subset generator** on akb's template model: 8–10 keywords (`type`, `enum`,
   `required`, `min/maxLength`, `pattern`, `min/maxItems`, `minimum`/`maximum`,
   `dependentRequired`, `if`/`then`/`else`), then run cel-go's static estimator against real
   pages and the 100000 budget — settles cost empirically, not by analogy. Include fixture
   tests for the three unverified primitives (`size()` unit for non-BMP strings, `has()` on
   `page` keys, RE2 rejection of ECMA-262 lookaround).
2. **Decide the product shape before building (a):** union format (`outputSchema` +
   `x-cel-validations`) versus true compilation. Only the union format has ecosystem
   precedent.
3. **(b) later, diagnostics-only:** AST subset extractor for `akb template get`/docs output
   with an explicit `unrepresentable` list; never authoritative, never round-tripped.

[k8s-cel]: https://kubernetes.io/docs/reference/using-api/cel/
[k8s-crd]: https://kubernetes.io/blog/2022/09/23/crd-validation-rules-beta/
[k8s-cost]: https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apis/apiextensions/validation/validation.go
[k8s121162]: https://github.com/kubernetes/kubernetes/issues/121162
[kep2876]: https://www.kubernetes.dev/resources/keps/2876/
[kep5073]: https://www.kubernetes.dev/resources/keps/5073/
[dv-docs]: https://www.kubernetes.dev/docs/code/declarative-validation/
[langdef]: https://github.com/cel-expr/cel-spec/blob/master/doc/langdef.md
[celgo]: https://github.com/cel-expr/cel-go
[celgo-ext]: https://github.com/cel-expr/cel-go/blob/master/ext/README.md
[celgo-opt]: https://github.com/cel-expr/cel-go/blob/master/cel/options.go
[celgo-ast]: https://github.com/cel-expr/cel-go/blob/master/common/ast/navigable.go
[runtimecost]: https://github.com/google/cel-go/blob/933f926a7fbc21d664a2894388e7a7811ae2ffcb/interpreter/runtimecost.go
[checkercost]: https://github.com/google/cel-go/blob/933f926a7fbc21d664a2894388e7a7811ae2ffcb/checker/cost.go
[pv-cel]: https://protovalidate.com/cel/how-cel-works/
[pv-readme]: https://github.com/bufbuild/protovalidate
[pv-lib]: https://github.com/bufbuild/protovalidate-go/blob/main/cel/library.go
[cerbos-js]: https://github.com/cerbos/protoc-gen-jsonschema
[cerbos-pr]: https://github.com/cerbos/cerbos/pull/727
[buf-ps]: https://github.com/bufbuild/protoschema-plugins
[buf-ps-109]: https://github.com/bufbuild/protoschema-plugins/issues/109
[decree]: https://github.com/opendecree/decree/blob/main/.agents/context/cel-validation.md
[jsv2020]: https://json-schema.org/draft/2020-12/json-schema-validation
[cel2sql]: https://github.com/SPANDigital/cel2sql
[kube-cel]: https://docs.rs/kube-cel/latest/kube_cel/
[flux-schema]: https://github.com/fluxcd/flux-schema
[helm-cel]: https://github.com/idsulik/helm-cel/
[datree]: https://hub.datree.io/custom-rules/cel-support
