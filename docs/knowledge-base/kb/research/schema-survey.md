---
grammar: 1
type: research
title: JSON-Schema side of the pi-subagents ↔ akb-CEL bridge — survey
status: final
provenance: agent-drafted
run: json-cel
as_of: 2026-09-26
created: 2026-09-26
updated: 2026-09-26
scope:
  - internal/cel/**
  - internal/template/**
tags: [json-schema, pi-subagents, survey, gap-analysis]
summary: Survey of outputSchema usage across pi-meta-config agents (4 of 7), pi-subagents validation (TypeBox, draft 3→2020-12), typed gates, and the feature/gap analysis against CEL.
informs:
  - RFC-001
  - RFC-002
---
# JSON-Schema side of the pi-subagents ↔ akb-CEL bridge — survey

Scope: (1) `outputSchema` frontmatter in pi-meta-config agents, (2) pi-subagents consumption/validation,
(3) typed-gate schema path, (4) feature inventory + GAP ANALYSIS. All anchors are exact file:line.

Repo roots:
- Agents: `/home/pete/code/projects/tools/pi/pi-meta-config/global-harness/agents/`
- Package: `/home/pete/.pi/agent/npm/node_modules/pi-subagents/` (version **0.71.0**, `package.json:4`)

---

## 1. Agent definitions carrying `outputSchema`

Exactly **4 of 7** agents in `global-harness/agents/` declare `outputSchema:` frontmatter
(confirmed by `grep -rn outputSchema global-harness/agents/`):

| Agent | file:line | Complexity |
| --- | --- | --- |
| task-sizer | `task-sizer.md:10` | large, `allOf`+`if/then`, no `$defs`/`$ref` |
| fix-designer | `fix-designer.md:13` | largest, `$defs`+`$ref` + 3× top-level `if/then` + nested `if/then` |
| spec-builder | `spec-builder.md:13` | large, `$defs`+`$ref` + 2× `if/then` |
| run-coroner | `run-coroner.md:12` | medium, flat, `additionalProperties:false`, no `allOf`/`$ref` |

No `outputSchema` in the other three: `context-synthesizer.md`, `council-agy.md`, `wave-reviewer.md`
(verified via frontmatter scan; `wave-reviewer.md:11` is `acceptanceRole: read-only`, no schema).
Project-tier agents: none under `.pi/agents/` carry one (only `.pi/agents/ext-sync-synthesizer.md`
mentions the concept in prose). `global-harness/skills/execute-plan/SKILL.md` references designer
schemas but defines none.

### 1a. Representative schemas

**run-coroner** (`run-coroner.md:12`) — simplest structured shape. Top-level `type:object`,
`required:["failureClasses","confidence","recommendedAction","paramPatch","evidence","summary",
"harnessSuspects","ownerDecisions"]`. Uses `array`+`minItems:1`+`items:{enum:[...]}`,
`enum`, `object`+`properties`+`additionalProperties:false` (only use of `additionalProperties`
in the entire set — `paramPatch`). `failureClasses.items` is a **typeless enum-only node**
(`"items":{"enum":[...]}`), i.e. `enum` without `type`.

**spec-builder** (`spec-builder.md:13`) — `$defs` (`scopeClass`,`authority`,`riskItem`,`anchor`) +
15× `$ref` (`#/$defs/...`), nested object arrays (`tasks[].anchors[].symbol`), `minItems:1`,
`minLength:1`, `enum`, `const`, and a **top-level `allOf` of 2 `if/then`** conditionals:
(i) `if workerWaiting.const == true → then required:[workerReply], workerReply.minLength:1`;
(ii) `if mode.const == "consult" → then properties.tasks.items.required:[scopeClass,authority]`.

**fix-designer** (`fix-designer.md:13`) — same `$defs` family plus **3 top-level `if/then`**
(workerWaiting, mode∈{adjudicate,consult,obs-adjudication}, mode=="backlog-triage"), a **nested
`allOf`/`if`/`then` inside `dispositions.items`** (`if disposition.const=="accept" → then
required:[fixSpec]`), deep nesting `triage.ranked[]/groups[]/dropped[]`, `minLength:1` on ids.
This is the richest schema and the hardest to convert.

**task-sizer** (`task-sizer.md:10`) — `allOf`/`if`/`then` inside `tasks.items`:
`if decision.enum∈{split-serial,split-parallel} → then required:[splits],
splits.minItems:1`. Note the `if` branch carries `enum` **without** an explicit `type`.
`task-sizer.md:13-21` is an HTML comment documenting the anyOf→if/then revert and the
"every schema node MUST carry an explicit type" transport rule (opencode-go grammar compiler
400s on typeless nodes, incident e53cac92).

### 1b. Duplication vs sharing

Schemas are **self-contained per file** — there is no cross-file `$ref` mechanism in frontmatter
(each agent's schema is a standalone JSON document; the runtime treats it as the root). Two
schemas are near-clones by copy-paste, not by reference:

- `$defs` shared between fix-designer and spec-builder: `scopeClass`, `authority`, `riskItem`
  are **byte-identical**; `anchor` **diverges** — spec-builder has `minLength:1` on
  `symbol`/`path`, fix-designer does not.
- `ownerDecisions` item schema and the `workerWaiting→workerReply` `if/then` are byte-identical
  across fix-designer and spec-builder.
- task-sizer and run-coroner are each independent/unshared.

Consequence for a converter: no schema dedup exists today; any "shared rule library" must be
introduced, and the fix-designer/spec-builder drift (anchor) is exactly the kind of
copy-paste divergence a shared source would eliminate.

---

## 2. pi-subagents `outputSchema` consumption

**Validator = TypeBox `Compile`, not Ajv.** No `ajv` anywhere in the tree
(`find … -name 'ajv*'` empty). Dependency is `typebox@1.1.38`
(`pi-subagents/package.json:29`, resolved at `pi-subagents/node_modules/typebox/`). The old
recon prose referencing `@sinclair/typebox` is stale.

Flow:
1. Launch-time shape check only: `assertJsonSchemaObject` (`structured-output.js:340`) rejects
   non-object/array. No keyword-level preflight.
2. The agent schema is **wrapped** before being handed to the model as a tool:
   `createStructuredOutputToolParameters` (`structured-output.js:140`) emits
   `{type:"object", properties:{value:<schema>, acceptanceReport?:{...}}, required:["value",...],
   additionalProperties:false}`. Local JSON-Pointer refs are rewritten one level down —
   `rewriteLocalJsonPointerRefs` (`structured-output.js:94`) maps `#`, `#/...`, `$defs`,
   `definitions`, `properties`, `items`, `allOf/anyOf/oneOf`, `if/then/else`, etc.
   (`structured-output.js:91-93` keyword tables). So `#/$defs/x` inside the agent schema still
   resolves after wrapping (`#/properties/value/$defs/x`).
3. Runtime writes schema to disk (`createStructuredOutputRuntime`, `structured-output.js:345`;
   consumed at `subagent-runner.js:497`, `async-execution.js:937/1579`). The child must finish by
   calling the `structured_output` tool (`MISSING_STRUCTURED_OUTPUT_CALL_ERROR`,
   `structured-output.js:15`).
4. **The actual validation**: `validateStructuredOutputValue` (`structured-output.js:362`) —
   `const compile = await loadCompile(); validator = compile(schema); validator.Check(value)`.
   `loadCompile` (`structured-output.js:~200`) imports **`typebox/compile`**'s `Compile`
   (with fallbacks; `STRUCTURED_OUTPUT_VALIDATOR_UNAVAILABLE_ERROR`, `structured-output.js:19`).
5. Errors: `validator.Errors(value)` expanded via `expandedErrorMessages` /
   `expandConditionalError` (`structured-output.js:~300-335`) which re-compiles the failing
   `then`/`else` branch into a diagnostic schema to produce **rooted field paths** for `if/then`
   failures (max 8 messages). Surfaced as `Structured output validation failed: <msg>`
   (`readStructuredOutput`, `structured-output.js:~390`); repackaged/sanitized and **bounded to
   4096 bytes** by `formatStructuredOutputRejectionError` (`structured-output.js:~65`,
   `MAX_STRUCTURED_OUTPUT_REJECTION_ERROR_BYTES=4096`). The child sees the rejection as a tool
   error and may re-emit once (agent-side circuit breaker; see `task-sizer.md:94`).

**Supported draft/subset**: TypeBox `Compile` supports JSON Schema **drafts 3 → 2020-12**
(`pi-subagents/node_modules/typebox/readme.md:194`, coverage table `readme.md:227+`).
Practically, anything a standard draft-07/2020-12 validator accepts is accepted here — including
`allOf/anyOf/oneOf/not`, `if/then/else`, `$ref`/`$defs`, `const`, `enum`, numeric bounds,
`pattern`/`format`, `prefixItems`, `dependentRequired/Schemas`, `unevaluated*`. Only partials:
`dynamicRef`, `unevaluatedItems`/`unevaluatedProperties` (not full 2020-12 coverage).

---

## 3. Typed gates

Docs section: `docs/tool-reference.md:413` "### Typed gates"; stdout bound stated at
`docs/tool-reference.md:426` ("at most 12,000 characters"); `outputSchema` conflict at
`docs/tool-reference.md:429`.

Implementation (`src/runs/shared/acceptance.js`):
- Gate shape parsed by `parseGateInput` (`acceptance.js:135`); allowed keys
  `GATE_OBJECT_KEYS = {command, output, schema, timeoutMs}` (`acceptance.js:132`).
  `gate.schema` must be a JSON-Schema object; requires `output:"json"` (`acceptance.js:154-159`).
- The task named `typedVerifyOutput` (`acceptance.js:1256`) is only a **reader** — returns the
  first verify run's `structuredOutput`. The real gate logic is `applyTypedVerifyOutput`
  (`acceptance.js:1229`): on a passing command it (a) fails on empty stdout, (b) fails if the
  output ends with the `...[truncated]` marker, (c) `JSON.parse` (fails on invalid JSON),
  (d) if `command.schema` given, calls **the same** `validateStructuredOutputValue(command.schema,
  value)` — i.e. **identical TypeBox validator and the full draft-3→2020-12 subset**, no reduced
  gate grammar.
- **Stdout size limit**: `trimOutput` (`acceptance.js:1134`) truncates to **12 000 characters**
  and appends `\n...[truncated]`; `TYPED_VERIFY_OUTPUT_MAX_BYTES = 12_000`
  (`acceptance.js:1221`). Truncation is detected by the marker (not a parse guess) and fails the
  gate: "stdout exceeded 12000 characters and was truncated." (`acceptance.js:1240`).
- Failure surface: the run is flipped to `status:"failed"` with `run.structuredOutputError`
  set (`acceptance.js:1231-1245`); `acceptance.js:1609-1610` turns that into
  `Acceptance verification '<id>' failed: <structuredOutputError>`.
- Typed gates are **never memoized** (`acceptance.js:1258+`, "typed gates read inputs … that can
  change without the tracked tree changing").
- One structured-output source per child: typed gate (or any typed `acceptance.verify`) **cannot**
  combine with an `outputSchema` — `TYPED_VERIFY_OUTPUT_SCHEMA_CONFLICT` (`acceptance.js:174`),
  enforced in `validateAcceptanceReportMode` (`acceptance.js:429`).

---

## 4. JSON-Schema feature inventory (in use vs never used)

Counts from parsing all 4 frontmatter schemas (property *names* excluded).

**In use (converter MUST support):**
- `type` — string/number/boolean/object/array; **all** nodes in 3 schemas carry explicit type.
- `properties`, `required` (arrays of property names; **no** `required` at root of `$defs` defs is
  conditional except via `if/then`).
- `items` (schema form only; **never** tuple/array form), `minItems`.
- `enum` (25×), `const` (5×) — enum often **without** adjacent `type`.
- `minLength` (20×, always `1`).
- `additionalProperties` (1×, value `false`, run-coroner only).
- `$defs` (2 agents) + `$ref` (15×, all local `#/$defs/...`).
- `allOf` (4×) always as `allOf:[{if,then}]`; `if`/`then` (7× each). **No `else`, no bare
  `allOf` of schemas, no `anyOf`/`oneOf`/`not`.**
- `description` (65×, annotation only, non-validating).

**Used by the validator but NOT in any agent schema** (converter may ignore for now, but the
validator accepts them): `anyOf`, `oneOf`, `not`, `else`, `pattern`, `format`, numeric bounds
(`minimum/maximum/exclusive*`), `multipleOf`, `maxItems/maxLength/minProperties/maxProperties`,
`uniqueItems`, `patternProperties`, `propertyNames`, `contains`, `prefixItems`, tuple `items`,
`dependencies`/`dependentRequired`/`dependentSchemas`, `$schema`, `$id`, `definitions`,
`unevaluated*`, `content*`.

**Never used anywhere (safe to reject/ignore in a converter):** the 2019-09/2020-12 additions
above, plus boolean-root schemas (`true`/`false` as a whole schema), `$dynamicRef`/`$recursiveRef`.

Note the **"explicit type on every node" transport rule** (`task-sizer.md:13-21`): the runtime
wraps the schema into an LLM tool signature, and opencode-go's grammar compiler rejects typeless
nodes. A converter emitting schemas must add explicit `type` everywhere (even on enum-only nodes)
or the launch 400s (incident e53cac92). This is a real production constraint, not style.

---

## GAP ANALYSIS: JSON Schema → CEL (akb page-template validation rules)

CEL is an expression language, not a schema language; a converter must lower schema structure into
CEL predicates over a single message/document. Difficulty tracks how much structural context a
keyword needs.

**Easy (direct CEL lowering):**
- `type` → `type(x) == string/…` or `has()`+dyn checks per node (CEL types are static-ish; a
  dynamic JSON message usually needs runtime `type()` guards).
- `required:["a","b"]` → `has(m.a) && has(m.b)`.
- `enum:["x","y"]` → `m.field in ["x","y"]` (`const` → `m.field == "x"`).
- `minLength:1` → `size(m.field) >= 1`.
- `minItems:1` → `size(m.field) >= 1`.
- `additionalProperties:false` → `!has(m.field)` for each undeclared key, or exclude-based check.

**Medium (needs per-item iteration / type dispatch):**
- `items:{schema}` + `minItems` on typed arrays → `m.list.all(x, <lowered item>)` (CEL `all`/`exists`
  macros), with a `type(x) == map`/`type(x) == string` guard per element kind.
- Nested objects/arrays → path chains `m.tasks.all(t, t.splits…)`; a path-aware emitter is required.
- `enum` on `items` without `type` → element-level `in` check (no type assertion to emit).

**Hard (no clean 1:1 CEL form):**
- `if/then` conditionals (7×, the crux of 3 of 4 agents) → `!(cond) || consequent`; correct but
  conditional requirements inside a `then` that also *adds properties* (spec-builder `then` adds
  `tasks[].required`) do not lower to a flat implication — you need the full antecedent guard plus
  a per-element `all()` that re-checks the required set only where matched. Doable but the
  schema→CEL compiler must detect `then`-scoped `properties`/`required` and scope them.
- `$ref`/`$defs` (15 refs) → CEL has no reference/rule-graph primitive; must be **inlined**
  (and inlined refs can blow up expression size — fix-designer's riskItem/anchor are reused 6×
  across nested arrays). Requires a ref-resolution/inlining pass before lowering.
- Nested `anyOf`/`oneOf` (validator supports them; schemas currently avoid them, but a generic
  converter must handle discriminator unions) → CEL has no schema-level union; must lower to a
  disjunction over branch predicates, and CEL's map/message typing makes branch-typed fields
  (`oneOf` where each branch has different properties) awkward without `has()` guards everywhere.
- `allOf` of *schemas* (as opposed to the `allOf:[{if,then}]` idiom actually used) → conjunction;
  fine, but combined with `$ref` inlining it multiplies expression size.

**Cross-cutting risks / open questions:**
1. **Rooted error paths.** pi-subagents' `if/then` diagnostics deliberately emit
   `tasks[0].splits: is required`-style paths (`structured-output.js:299-335`). A CEL lowering
   returns a boolean/first-failure message at whatever granularity the akb template emits — parity
   of actionable field paths is *not* free and is the main UX regression risk when moving validation
   from outputSchema to CEL rules.
2. **Validator ≠ subset.** The gate and outputSchema share TypeBox (full draft-3→2020-12). A CEL
   converter supporting only the in-use feature set is sufficient *today* but will silently diverge
   if a new agent uses `anyOf`/`pattern`/numeric bounds. Recommend the converter explicitly reject
   unsupported keywords rather than ignore them.
3. **Explicit-type rule** (opencode-go) is an LLM-transport artifact with no CEL analogue; the
   direction of conversion matters — JSON-Schema→CEL never needs it, but CEL→JSON-Schema does if the
   result feeds an agent launch.
4. **Frontmatter is single-line strict JSON.** Any generated schema must survive one-line
   serialization (parse sanity is a documented failure mode; cf. `.pi/subagents/specs/archive-2026-09-21/plan.json`).
5. **No shared schema source today** — fix-designer/spec-builder already drift (`anchor`). A bridge
   is a good forcing function to introduce one canonical source per schema and generate both sides.
6. **Wrapper rewrite** (`structured-output.js:94-149`) means a converter reading agent frontmatter
   sees the *unwrapped* schema, but the runtime validates the *wrapped* one; local `$ref`s are
   rewritten. A converter should resolve refs against the unwrapped root (as this survey did).

---

## Start Here

For the bridge work, open **`global-harness/agents/fix-designer.md:13`** — it exercises every
in-use feature at once (`$defs`/`$ref`, nested `if/then`, deep object/array nesting, `minLength`,
`enum`/`const`). Then read
**`/home/pete/.pi/agent/npm/node_modules/pi-subagents/src/runs/shared/structured-output.js:362`**
(the shared validator both outputSchema and typed gates call) to confirm parity targets.
