---
grammar: 1
type: research
title: "cel-go v0.28.0 → v0.32.x upgrade research (long-form)"
status: final
provenance: agent-drafted
run: rfc-002-a3-celgo
as_of: "2026-10-06"
created: "2026-10-06"
updated: "2026-10-06"
scope:
  - "go.mod"
  - "internal/cel/**"
tags:
  - cel-go
  - dependencies
  - upstream
  - adr-004
summary: "Source-verified survey of the cel-go v0.28.0→v0.32.x upgrade: module-path migration, v0.30 timestamp strictness, extension library costs, and advisories, with evidence labels and open gaps."
informs:
  - ADR-004
---
# cel-go v0.28.0 → v0.32.x upgrade research (long-form)

Scope: `github.com/google/cel-go v0.28.0` (with `cel.dev/expr v0.25.1` indirect) → `cel.dev/cel-go v0.32.x`.
Research date: **2026-10-06**. All claims below are from fetched primary sources (GitHub release pages, PRs,
issues, raw source at tags, pkg.go.dev, Go vuln DB) — not from training data.

Evidence labels used:
- **[DIRECT]** = quoted/paraphrased straight from an authoritative source I fetched.
- **[INFER]** = researcher inference from reading source code; not stated by the source.
- **[SINGLE]** = only one source found; treat as unconfirmed.
- **[GAP]** = not verified.

---

## 1. MODULE PATH MIGRATION

### 1.1 Which release moved the module
- **v0.32.0** (published 2026-08-19T19:43:14Z) is the release that switched the module path. Release notes list
  under **Breaking Changes**: "**Switch module and import paths to cel.dev/cel-go** — #1413"
  ([release v0.32.0](https://github.com/cel-expr/cel-go/releases/tag/v0.32.0)).
- The tag `v0.32.0` is the merge commit of PR #1413 ("Merge pull request #1413 from cel-expr/vanity-cel-go-import
  / Switch module and import paths to cel.dev/cel-go") ([tags page](https://github.com/cel-expr/cel-go/tags)).
- Repo rename (separate, earlier event): `google/cel-go` → `cel-expr/cel-go`, announced in pinned issue #1329
  ("On June 16, 2026, this repository will move to github.com/cel-expr/cel-go!")
  ([issue #1329](https://github.com/cel-expr/cel-go/issues/1329)). Git traffic on the old URL redirects; the module
  path change is NOT the same thing as the repo rename.
- pkg.go.dev states the split explicitly:
  "**cel.dev/cel-go**: Canonical path for **v0.32.0** and newer. **github.com/google/cel-go**: Path for versions
  **v0.2.0** through **v0.31.0**. **github.com/cel-expr/cel-go**: Only contains **v0.1.0**"
  ([pkg.go.dev/cel.dev/cel-go](https://pkg.go.dev/cel.dev/cel-go)).
  → **v0.31.0 is the last version under `github.com/google/cel-go`; v0.32.0+ only exists at `cel.dev/cel-go`.**

### 1.2 Migration mechanics
- **go.mod**: replace `github.com/google/cel-go v0.28.0` with `cel.dev/cel-go v0.32.0`.
- **Imports**: every import must be rewritten, e.g.
  `github.com/google/cel-go/cel` → `cel.dev/cel-go/cel`,
  `github.com/google/cel-go/common/types` → `cel.dev/cel-go/common/types`,
  `github.com/google/cel-go/checker`, `/parser`, `/interpreter`, `/ext` likewise. This is a whole-repo rename, not a
  single-package move.
- **No `replace` directive is required** for a normal mono-consumer upgrade — but `replace` will NOT work as a way to
  keep the old path at v0.32.0. The failure mode seen across the ecosystem is a resolution error, not a version error:
  ```
  go: github.com/google/cel-go@v0.32.0: parsing go.mod:
      module declares its path as: cel.dev/cel-go
              but was required as: github.com/google/cel-go
  ```
  ([heimdall PR #3483](https://github.com/dadrus/heimdall/pull/3483),
  [kromgo PR #371](https://github.com/home-operations/kromgo/pull/371),
  [cilium PR #48341](https://github.com/cilium/cilium/pull/48341),
  [GitLab client-go MR !3010](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/3010)).
  In practice Renovate/Dependabot PRs that only bumped the version string fail; the import rewrite has to be part of
  the same change (cilium did exactly this: "updates the import path accordingly here, also updates the module from
  v0.31.0 to v0.32.0").
- **Dual-publish period / old path updates**: there is no dual publish of v0.32+. `github.com/google/cel-go` is now a
  **read-only compatibility repository** that "contains compatibility aliases (Go type aliases, function forwarders,
  and variable bindings) that forward directly to their canonical counterparts in `cel.dev/cel-go`" for
  "diamond-dependency scenarios", and is marked "temporary and will eventually be removed"
  ([pkg.go.dev/github.com/google/cel-go](https://pkg.go.dev/github.com/google/cel-go), also mirrored in the repo README).
  Consequence: if any transitive dependency still imports `github.com/google/cel-go` while we import `cel.dev/cel-go`,
  both compile and the types are identical (aliases), so no `replace` is needed for that case either.
- **Submodules also re-tagged at v0.32.0**: `repl/v0.32.0`, `tools/v0.32.0`, `policy/v0.32.0`,
  `conformance/v0.32.0`, `codelab/...`, `ext/security/v0.32.0` (releases.atom entries dated 2026-08-19).

### 1.3 cel.dev/expr pairing
- `cel.dev/cel-go v0.32.0`'s own `go.mod` (tag v0.32.0, fetched raw):
  ```
  module cel.dev/cel-go
  go 1.23.0
  require (
      cel.dev/expr v0.25.1
      github.com/antlr4-go/antlr/v4 v4.13.1
      github.com/google/go-cmp v0.7.0
      go.yaml.in/yaml/v3 v3.0.4
      google.golang.org/genproto/googleapis/api v0.0.0-20240826202546-f6391c0de4c7
      google.golang.org/protobuf v1.36.10
  )
  require (
      golang.org/x/exp v0.0.0-20240823005443-9b4947da3948 // indirect
      golang.org/x/text v0.22.0
      google.golang.org/genproto/googleapis/rpc v0.0.0-20240826202546-f6391c0de4c7 // indirect
  )
  ```
  ([go.mod @ v0.32.0](https://raw.githubusercontent.com/cel-expr/cel-go/v0.32.0/go.mod)).
- **cel.dev/expr stays at v0.25.1** — the same version the project already pins indirectly. No expr bump is implied
  by the cel-go upgrade. (pkg.go.dev's dependency table for v0.32.0 agrees: `cel.dev/expr v0.25.1`.)
- Notable non-cel-go dependency changes vs a v0.28-era tree: `github.com/google/go-cmp v0.7.0`, `go.yaml.in/yaml/v3`,
  `golang.org/x/text v0.22.0`, `google.golang.org/protobuf v1.36.10`, `antlr4-go/antlr/v4 v4.13.1`. (Some of these
  changes may already have landed at v0.30/v0.31; not diffed per release — see GAPS.)

---

## 2. RELEASE-BY-RELEASE: v0.29 → v0.32 (API-relevant changes)

### v0.29.0 — 2026-07-03 ([release](https://github.com/cel-expr/cel-go/releases/tag/v0.29.0))
- New: JSON encoder `ext` library (#1340); `network.IP`/`CIDR` upstreamed from Kubernetes (#1238).
- **Managed execution frame / InterpretableV2** (#1316, #1344): "Plugs the `ExecutionFrame` into the CEL interpreter
  and program APIs and introduces a new `InterpretableV2` for calling program steps with the `ExecutionFrame` rather
  than the `Activation`" ([PR #1344](https://github.com/cel-expr/cel-go/pull/1344)). Files touched include
  `interpreter/interpretable.go` (+346/−157) and new `interpreter/frame.go`
  ([compare v0.28.0...v0.29.0](https://github.com/google/cel-go/compare/v0.28.0...v0.29.0)).
  → **Breaking for anyone implementing `interpreter.Interpretable`/`InterpretableCall` or writing custom
  `InterpretableDecorator`s.** A compat path exists (`interpreter.CustomDecorator` adapts a legacy decorator via
  `adaptToV2`, plus a new `interpreter.CustomDecoratorV2`) — so `cel.CustomDecorator` still compiles.
- Security fix #1302: "Enforce expression size limit before source construction" → this is the fix for
  **CVE-2026-83530**.
- Input-validation tightenings: timezone-offset minutes (#1336), `indexOf`/`lastIndexOf` empty-offset (#1335),
  int32/uint32 map-key narrowing (#1337), `ext/lists` `genRange()` max size to prevent OOM (#1310).
- Cost work: receiver/global cost agreement (#1350), `startsWith`/`endsWith` runtime-vs-checked agreement (#1351),
  "Avoid repeated construction of cost tracker" (#1357 — cost tracker now built once per program and cloned per eval;
  see program.go comment below).

### v0.29.1 (2026-07-03) / v0.29.2 (2026-07-08)
- Patch releases only: CRLF/LF in `cel/prompt.go` (#1359), `ext/lists` runtime cost calculator update (#1360).
  There is **no v0.29.3 tag** (release URL 404s; the v0.30.0 release-notes "Full Changelog" link to
  `v0.29.2...v0.29.3` is a release-notes artifact, not a tag).

### v0.30.0 — 2026-07-26 ([release](https://github.com/cel-expr/cel-go/releases/tag/v0.30.0))
- **Timestamp strictness** (#1338) — see §3.
- **Loaded-AST depth validation** (#1334): `ParsedExprToAst`/`CheckedExprToAst` now depth-check loaded ASTs;
  default `ast.MaxNestingDepth = 250` mirroring the parser; error text
  `input exceeds maximum expression nesting depth: 250`. Configurable via a `NewEnv` option
  ([PR #1334](https://github.com/cel-expr/cel-go/pull/1334); limit ID `cel.limit.max_ast_depth`,
  `cel.ExpressionNestingDepthLimit(n)` in v0.32 `cel/options.go`).
- `Prevent indexing on sentinel json field '-'` (#1349) — fix for GHSA-gcjh-h69q-9w9g.
- Async internals (#1355, #1364, #1367, #1369): `DrainStrategy`, `DrainAction`, `Observer`, `Call` types;
  `ConcurrentEval` with `context.Context`-bound functions.
- Constant-folding correctness: #1371, #1374 (`x in [x]` NaN), #1380.
- `cel.bind()` nesting limits (#1378); "Allow validators to be reconfigured" (#1379).
- **Expression node limits for parser and checker (#1386)**: "In addition to limiting parser recursion depth and ast
  recursion depth, also limit the number of AST nodes permitted in an expression generated by the parser or accepted
  by the checker." Touches `cel/env.go`, `cel/options.go`, `parser/parser.go`, `common/ast/ast.go`
  ([PR #1386](https://github.com/cel-expr/cel-go/pull/1386)).
  In v0.32 `cel/env.go`: `configuredExpressionSizeLimit()` → default `100_000` code points;
  `configuredExpressionNodeLimit()` → default `100_000` nodes. **New hard defaults where v0.28 had no node cap.**
- Cost tracking added/changed for: encoders (#1354), `ext/lists` (#1352), `ext/math` (#1353), plus "dummy 'infinite'
  costs" for non-deterministic JSON encoding (#1365).

### v0.31.0 — 2026-08-07 ([release](https://github.com/cel-expr/cel-go/releases/tag/v0.31.0))
- **Env copy-on-write (#1405)**: "Introduces Copy-on-Write (COW) mechanics for environment definitions… The COW
  optimization prevents copying underlying maps of variables, functions, and adapters." Also "Fix race-related issue
  with copy-on-write mutability check". Benchmarks: `BenchmarkEnvExtension/10_vars 3.56µs → 0.15µs`, allocs 32 → 1.
  In v0.32 `cel/env.go` this shows up as `funcsShared`/`limitsShared`/`libsShared` flags plus
  `ensureMutableFunctions()`/`ensureMutableLimits()`/…, a `parent *Env` link, and a per-env
  `sharedDispatcher` that "caches a dispatcher populated with the env's function bindings, built once and reused
  across every Program() constructed from this env".
- **Native types moved to core**: `Move native type support into the Core CEL library (#1396)`; release notes say
  "There are no breaking changes, but all `ext.NativeTypes` features are now available via `cel.NativeTypes`".
  Verified at v0.32.0: `ext/native.go` still exists and is a thin alias layer
  (`type NativeTypesOption = types.NativeTypeOption`, `ParseStructTag = types.ParseStructTag`,
  `ParseStructTags = types.ParseStructTags`, `ParseStructField = types.ParseStructField`, `NativeTypes(...)`
  reimplemented on `types.ComposeTypes`), and `NativeTypesVersion` is a documented **no-op**:
  "Deprecated: NativeTypesVersion is a no-op and will be removed in a future release."
  ([ext/native.go @ v0.32.0](https://raw.githubusercontent.com/cel-expr/cel-go/v0.32.0/ext/native.go)).
- **Regex program plan size controls (#1383)**: "mirrors the logic used in RE2 within cel-java and cel-cpp…
  provides a means to control the plan size per-platform" ([PR #1383](https://github.com/cel-expr/cel-go/pull/1383)).
  It is **opt-in**: `cel.RegexProgramSizeLimit(limit)` — "A negative or zero value means unbounded" — and when
  `limit > 0` it also installs `cel.ValidateRegexProgramSizeLimit(limit)` as an AST validator
  ([cel/options.go @ v0.32.0](https://raw.githubusercontent.com/cel-expr/cel-go/raw/v0.32.0/cel/options.go)).
  Program construction only applies it when positive:
  `if limit := p.limits[limitRegexProgramSize]; limit > 0 { plannerOptions = append(plannerOptions, interpreter.RegexProgramSizeLimit(limit)) }`
  ([cel/program.go @ v0.32.0](https://raw.githubusercontent.com/cel-expr/cel-go/v0.32.0/cel/program.go)).
  **[INFER]** the default is therefore unbounded unless the embedder sets it — the "critical size constraint" headline
  is a capability, not an on-by-default cap.
- Optimizer changes (#1406 list concat consolidation, #1387 optional macros) → smaller program plans.
- `reject out-of-range hours in timezone offset parsing` (#1391) — this is about `timestamp.getHours('+24:00')`
  *timezone-string* arguments, not `timestamp("...")` parsing.

### v0.32.0 — 2026-08-19 ([release](https://github.com/cel-expr/cel-go/releases/tag/v0.32.0))
- **Breaking**: module/import path switch (#1413) — §1.
- Features: JWT data types + claim helpers (#1415), HMAC verify/compute library (#1416), Go-based JSON type support in
  `NativeToValue` (#1402), new timestamp parsing helper (#1414), aggregate size computations over list/map/struct
  (#1404), policy aggregate semantics (#1408), "Report every evaluation step to every observer" (#1419).
- Fixes: nil-valued struct pointer panic in native traversal (#1417), shorthand type specifier parsing now allows
  newlines/tabs (#1411), **"Scale sizes for strings and bytes" (#1421)** and **"Consolidate saturating cost arithmetic
  into common/cost" (#1420)** → cost numbers move again (see §2.4), plus cost observability fix "when combined with
  state tracking or exhaustive eval".

### 2.4 Impact on this CLI's three risk areas

**(a) Env with a cost limit — API stable, NUMBERS not.**
- `cel.CostLimit` is byte-for-byte the same option in v0.28.0 and v0.32.0 (same doc comment, same
  `p.costLimit = &costLimit; p.evalOpts |= OptTrackCost`), and `cel.CostTracking(estimator)` is unchanged
  ([v0.28.0 cel/options.go](https://raw.githubusercontent.com/cel-expr/cel-go/v0.28.0/cel/options.go),
  [v0.32.0 cel/options.go](https://raw.githubusercontent.com/cel-expr/cel-go/v0.32.0/cel/options.go)).
- **The RFC's term `cel.ActualCostLimit` does not exist.** A full read of v0.32.0 `cel/options.go` finds no such
  symbol; GitHub code search for "ActualCostLimit" returns no cel-go hit. The runtime limit is `cel.CostLimit`; the
  runtime *observable* is `cel.EvalDetails.ActualCost() *uint64` via `cel.EvalOptions(cel.OptTrackCost)`; the
  injection points are `cel.CostTracking(interpreter.ActualCostEstimator)`,
  `cel.CostTrackerOptions(...interpreter.CostTrackerOption)` and
  `interpreter.CostTrackerLimit(n)` / `interpreter.OverloadCostTracker(overloadID, fn)`.
- Enforcement mechanism unchanged: exceeding the limit panics `interpreter.EvalCancelledError{Cause: CostLimitExceeded}`
  and is recovered by `prog.Eval` into a returned error.
- **Risk:** the *cost model* changed in v0.29–v0.32 (lists #1352/#1360, math #1353, encoders #1354, strings/bytes
  scaling #1421, saturating arithmetic #1420, optional `optMap`/`optFlatMap` macro shape). Issue **#1476
  "Non backward compatible cost changes in release 0.32.0"** documents Kubernetes-visible regressions:
  `optional.of('a').optMap(v, v == 'value').hasValue()` estimated 18 → **29**; `self.l[?0].optMap(v, v == 'a').hasValue()`
  22 → **31**; runtime `optional.of('a').optMap(...)` 8 → 9 (some exist/map cases went *down*, e.g. 24 → 22).
  The maintainer replied: "These two are fixed at HEAD… The others are as a result of a macro change for safety;
  however, I will revert the change and introduce the fixed macros in a bump of the optional library," referencing
  PR #1487 "Make the optMap / optFlatMap fixes versioned and backwards compatible"
  ([issue #1476](https://github.com/cel-expr/cel-go/issues/1476)). #1487 is **not** in v0.32.0 (opened after it) —
  so a limit/estimate tuned on v0.28 may now trip on optional-heavy rules. Re-baseline any stored/expected costs.

**(b) sync.Map program cache.**
- `cel.Program` remains the immutable-ish, concurrently evaluable unit; caching `Program` values keyed by rule
  source is still valid. Nothing in the v0.29–v0.32 notes changes `cel.NewEnv`/`env.Program`/`Program.Eval`
  signatures.
- New in v0.30: `Program.ConcurrentEval(ctx, input)` (channel-based, for async calls) and **`Eval` now refuses
  programs containing async calls**: "Asynchronous calls cannot be resolved by a single-pass evaluation. Reject before
  doing any work (this also covers ContextEval, which delegates here); ConcurrentEval does not call Eval"
  ([cel/program.go @ v0.32.0](https://raw.githubusercontent.com/cel-expr/cel-go/v0.32.0/cel/program.go)).
  Only relevant if extension libraries introduce async functions (ours do not).
- **COW caution (#1405)**: an `Env` now shares maps with its parent and clones lazily on mutation. Sharing a base
  `Env` across goroutines while calling `env.Extend(...)`/`EnvOption`s from multiple goroutines is the race the COW
  check guards; build envs in one goroutine (or per goroutine) and only share *programs*. Programs built from a shared
  base env reuse a cached dispatcher (`sharedDispatcher`), which is the intended read-only sharing path.
- Cost-tracker construction moved: v0.32 builds one `interpreter.CostTracker` at program-construction time and
  clones it per evaluation ("Creating a new cost tracker for each evaluation causes significant work… Therefore it
  gets constructed once now and later a cheap clone is used"), matching v0.29 #1357.

**(c) panic recovery around `Evaluate`.**
- Unchanged shape in v0.32.0: `prog.Eval` installs a deferred recover that maps
  `interpreter.EvalCancelledError` → returned error, and anything else → `fmt.Errorf("internal error: %v", r)`.
  `ConcurrentEval` has the same recover in its goroutine. So keeping our own `recover()` as a belt-and-braces wrapper
  still works and is still wise; but note that cel-go already converts CEL-runtime panics into `err`/`ref.Val`,
  meaning a panic reaching our recover now indicates either an `EvalCancelledError` propagating out of a nested
  evaluation or a genuine cel-go/internal bug.
- New error surface in v0.30+: async-related errors (`errAsyncRequiresConcurrentEval`, `EvalResult{Err: ...}`),
  which would surface through `ConcurrentEval`, not `Eval`.
- `types.Provider`: `cel.CustomTypeProvider(provider any)` still accepts `types.Provider` or the deprecated
  `ref.TypeProvider`, normalising via `maybeInteropProvider`. No change found between v0.28 and v0.32 (the
  `ref.TypeProvider` deprecation predates v0.28). **[INFER/SINGLE]** — I compared the v0.32 source and the v0.28
  option list only for the symbols cited; a full symbol-level API diff was not run.
- `types.Provider`'s interface itself appears unchanged; v0.31's `types.ComposeTypes`/native-type code paths are
  additive.

---

## 3. TIMESTAMP STRICTNESS (the v0.30 change)

### What changed
- The change is **PR #1338 "reject non-RFC3339 timestamp strings in timestamp() conversion"**, merged 2026-07-15,
  shipped in **v0.30.0** (release notes list it; v0.29.2 was tagged 2026-07-08, i.e. before the merge)
  ([PR #1338](https://github.com/cel-expr/cel-go/pull/1338),
  [release v0.30.0](https://github.com/cel-expr/cel-go/releases/tag/v0.30.0)).
- PR body (direct): "Repro: `timestamp("2025-01-17T01:00:00,001Z")`, `timestamp("2025-01-17T1:00:00Z")`,
  `timestamp("2025-01-18T01:01:01.001+24:01")` and `timestamp("2025-01-17T01:01:01.001+00:60")` all evaluate to a value
  instead of erroring. Cause: the `string`-to-timestamp conversion parses with `time.Parse(time.RFC3339, ...)`, which
  accepts inputs RFC 3339 forbids: a `,` fractional separator, single-digit time fields, and offset hours past `23`
  or minutes past `59`. … Fix: gate the conversion on a strict RFC 3339 pattern in the callee before `time.Parse` runs."
- **v0.28.0 code (before)**: `String.ConvertToType(TimestampType)` did
  `if t, err := time.Parse(time.RFC3339, s.Value().(string)); err == nil { … }` with only the
  `minUnixTime`/`maxUnixTime` overflow check — i.e. pure Go leniency
  ([common/types/string.go @ v0.28.0](https://raw.githubusercontent.com/cel-expr/cel-go/v0.28.0/common/types/string.go)).
- **v0.32.0 code (after)**: same function now
  ```
  case TimestampType:
      str := s.Value().(string)
      if !isStrictRFC3339(str) {
          return NewErr("invalid RFC 3339 timestamp %q", str)
      }
      if t, err := time.Parse(time.RFC3339, str); err == nil { … }
  ```
  ([common/types/string.go @ v0.32.0](https://raw.githubusercontent.com/cel-expr/cel-go/v0.32.0/common/types/string.go)).
- Implementation `isStrictRFC3339` is a hand-rolled, allocation-free scan (the regex `strictRFC3339Pattern` is kept as
  the test reference). PR review notes the perf tradeoff: ~600 ns/op with regex → ~197 ns/op with `strconv.ParseUint`
  helpers, gate allocation-free ([PR #1338 review threads]).
- v0.32.0 also adds an exported helper `types.ParseTimestamp(val any)` (PR #1414) for embedders, accepting
  `time.Time`/`Timestamp`/`*timestamppb.Timestamp`, RFC 3339/3339Nano strings, unix epoch ints/floats and `json.Number`;
  string inputs go through the same `isStrictRFC3339` gate.

### The precise v0.32.0 acceptance set (`isStrictRFC3339`, verbatim logic)
Grammar enforced by the scan (all fields fixed-width):
`YYYY-MM-DD ['T'|'t'] HH:MM:SS [ '.' 1*DIGIT ] ( ['Z'|'z'] | ('+'|'-') HH:MM )`
- length ≥ 20; year 4 digits (`0000`–`9999`); `-`; month `01`–`12`; `-`; day `01`–`31`; `T`/`t`;
  hour `00`–`23`; `:`; minute `00`–`59`; `:`; **second `00`–`60` (60 allowed = leap second)**.
- fractional part: a **period** `.` followed by **1 or more digits**, any length; then
- zone: exactly one char `Z`/`z`, **or** exactly `±HH:MM` with offset hour `00`–`23` and minute `00`–`59`.
After the gate passes, `time.Parse(time.RFC3339, s)` must also succeed (calendar validation: day-of-month vs month,
leap years), then `t.Unix()` must be within `[minUnixTime, maxUnixTime]` = `[-62135596800, 253402300799]`
(year 1 … year 9999-12-31T23:59:59Z), else `celErrTimestampOverflow`.

Concrete cases (accepted/rejected) — the "net" column is gate ∧ `time.Parse`:
| Input | v0.28 | v0.32 gate | v0.32 net | Note |
|---|---|---|---|---|
| `2025-01-17T01:00:00.001Z` | accept | accept | **accept** | listed `valid` in v0.32 test |
| `2025-01-01T12:34:56Z` | accept | accept | **accept** | |
| `2025-01-01T12:34:56.123456789Z` | accept | accept | **accept** | benchmark input |
| `2025-01-01T12:34:56+05:30` / `-08:00` / `+14:00` | accept | accept | **accept** | `valid` list |
| `2025-01-17T01:00:00,001Z` (comma fraction) | **accept** (Go leniency) | **reject** | **reject** | listed `invalid` |
| `2025-01-17T1:00:00Z` (single-digit hour) | **accept** | **reject** | **reject** | listed `invalid` |
| `2025-01-17T01:5:00Z` (single-digit minute) | **accept** | **reject** | **reject** | listed `invalid` |
| `2025-01-18T01:01:01.001+24:01` (offset hour >23) | **accept** (silently shifts instant) | **reject** | **reject** | listed `invalid` |
| `2025-01-17T01:01:01.001+00:60` (offset min >59) | **accept** | **reject** | **reject** | listed `invalid` |
| `2025-01-01T12:34:56` (no zone) | reject | **reject** | reject | in scan-vs-regex case list |
| `2025-01-01 12:34:56Z` (space) | reject | **reject** | reject | |
| `2025-1-01T12:34:56Z` (1-digit year) | reject | **reject** | reject | |
| `2025-01-01T12:34:56.Z` (dot, no digits) | reject | **reject** | reject | |
| `2025-01-01T12:34:56+0530` (no colon) | reject | **reject** | reject | |
| `2025-01-01T12:60:00Z` (minute 60) | reject | **reject** | reject | |
| `2025-01-01T12:34:61Z` (second 61) | reject | **reject** | reject | |
| `2025-01-01T24:00:00Z` (hour 24) | reject | **reject** | reject | |
| `2025-01-01t12:34:56Z` (lowercase t) | reject | **accept** | **reject** [INFER] | see below |
| `2025-01-01T12:34:56z` (lowercase z) | reject | **accept** | **reject** [INFER] | see below |
| `2025-01-01T12:34:60Z` (leap second) | reject [INFER] | **accept** | **reject** [INFER] | Go: "does not parse leap seconds" |
| `…T12:34:56.1234567890Z` (>9 frac digits) | reject [INFER] | accept | **reject** [INFER] | Go `parseNanoseconds` range error |

Evidence for the rows marked [INFER]:
- cel-go side is direct: the scan and the reference regex both contain `[Tt]`/`[Zz]` and `([0-5]\d|60)`, and
  `TestIsStrictRFC3339MatchesPattern` asserts scan==regex on "2025-01-01T12:34:56z", "2025-01-01t12:34:56Z"
  and leap-second/boundary cases ([common/types/timestamp.go](https://raw.githubusercontent.com/cel-expr/cel-go/v0.32.0/common/types/timestamp.go),
  [common/types/timestamp_test.go](https://raw.githubusercontent.com/cel-expr/cel-go/v0.32.0/common/types/timestamp_test.go)).
- Go side: `time.Parse(time.RFC3339, …)` fast-paths `parseRFC3339`, which hard-requires `s[10] == 'T'` and `s[0] == 'Z'`
  (uppercase), and bounds seconds to `0..59`; the fallback `parse()` matches the layout literal `"T"` case-sensitively
  (`skip()` compares bytes with `value[0] != prefix[0]`) and only accepts `'Z'` uppercase for the `stdISO8601ColonTZ`
  branch ([format_rfc3339.go](https://raw.githubusercontent.com/golang/go/master/src/time/format_rfc3339.go),
  [format.go](https://raw.githubusercontent.com/golang/go/master/src/time/format.go)). golang/go#20869 records the
  deviation as a known, deliberately-unfixed bug: "The format is case sensitive… Given that generators should generate
  T and Z, not t and z, I'm not particularly inclined to try to fix this."
- Therefore in v0.32 a lowercase `t`/`z` or a leap second passes the CEL gate and then fails `time.Parse`, falling
  out of the `switch` to the generic `type conversion error from 'string' to 'timestamp'` message (rather than the new
  `invalid RFC 3339 timestamp "…"` message). Net behaviour = rejected, i.e. **no regression vs v0.28** for those two
  shapes; only the error *message* differs.

**Bottom line for the KB's date fields:** the RFC's assumption that v0.30 made `timestamp()` stricter is correct, but
the only *previously-accepted* forms that now fail are: comma fractions, single-digit time fields, out-of-range offset
hour/minute (and, per the gate, nothing else). Lowercase `t`/`z` and leap seconds were already rejected in v0.28 and
remain rejected. A CEL date field produced by `time.RFC3339Nano` (`Format(time.RFC3339Nano)`) round-trips fine:
uppercase `T`, `.` fraction with ≤9 digits, `Z` or `±HH:MM`.

---

## 4. EXTENSION LIBRARIES IN v0.32.0

All of `ext.Strings`, `ext.Lists`, `ext.Sets`, `ext.Math`, `ext.Bindings` exist in v0.32.0 (verified in the
v0.32.0 `ext/README.md` and per-file source). v0.32.0 `ext` package contents include: `bindings.go`,
`comprehensions.go`, `encoders.go`, `guards.go`, `lists.go`, `math.go`, `native.go`, `protos.go`, `regex.go`,
`sets.go`, `strings.go`, `formatting*.go`, `extension_option_factory.go`.

| Library | Constructors/options | Introduced | Function inventory (high level) |
|---|---|---|---|
| `ext.Bindings()` | – | before v0.13.0 (present at v0.12.5) | macro `cel.bind(varName, initExpr, resultExpr)` |
| `ext.Strings()` | `StringsVersion(uint32)`, `StringsLocale(string)` (ignored at version ≥4), `StringsMaxPrecision(int)`, deprecated no-op `StringsValidateFormatCalls(bool)` | early (pre-v0.13.0) | `charAt`, `indexOf`, `lastIndexOf`, `lowerAscii`, `replace` (optional count), `split` (optional count), `substring`, `trim`, `upperAscii`, `strings.quote` (v1), `join`/`join(sep)` (v2), `reverse` (v3), `format` (v1; **v4 rewrites semantics** to spec v2) |
| `ext.Lists()` | `ListsVersion(uint32)`, `ListsMaxRangeSize(int64)` | **added v0.20.x–v0.21.0** ("Add functions to the lists extension. (#1037)") | `slice` (v0), `flatten`/`flatten(depth)` (v1), `distinct`, `lists.range(n)`, `reverse`, `sort`, `sortBy` macro (v2) |
| `ext.Sets()` | `SetsVersion(uint32)` | ~v0.20/0.21 (REPL support commit "Add sets … (#1005)") | `sets.contains`, `sets.equivalent`, `sets.intersects` |
| `ext.Math()` | `MathVersion(uint32)` | **v0.13.0** (`ext/math.go` added in v0.12.5...v0.13.0) | macros `math.greatest`, `math.least` (v0; cost support v3); v1: `ceil/floor/round/trunc/isInf/isNaN/isFinite/abs/sign/bitAnd/bitOr/bitXor/bitNot/bitShiftLeft/bitShiftRight`; v2: `sqrt` |
| also present | `ext.Encoders()` (base64.encode/decode v0, json.encode v1), `ext.Protos()`, `ext.TwoVarComprehensions()` (all/exists/existsOne/transformList/transformMap/transformMapEntry), `ext.Regex()` (regex.replace; needs `cel.OptionalTypes()`), `ext.NativeTypes()` (now alias layer over `cel`/`types`), `ext.Formatting()` | v0.13.0 (encoders/protos/native), later for the rest | – |

Notable inventory details from v0.32 `ext/README.md`: `Lists`: "**Distinct** Introduced in version: 2 (cost support
in version 3)", "**Range** Introduced in version: 2 (cost support in version 3)", "**Slice** Introduced in
version: 0 (cost support in version 3)", "**Flatten** Introduced in version: 1 (cost support in version 3)";
`Math`: "**Math.Greatest** Introduced in version 0 (cost support in version 3)"; `Encoders`:
"**Base64.Encode** Introduced in version 0 (cost support in version 1)", "**JSON.Encode** Introduced at version: 1".

### ext.Strings version split (the one the RFC cares about)
From the v0.32.0 source (`ext/strings.go`, fetched at tag):
- version gates: `if lib.version >= 1 { … format (v1 semantics) + strings.quote … }`,
  `>= 2 { join, join(sep) }`, `>= 3 { reverse }`, and inside the format branch
  `if lib.version >= 4 { … parseFormatStringV2 … } else { … parseFormatString … }`.
- The doc comment on the format function: "Introduced at version: 4 (cost support in version 5). Formatting updated
  to adhere to https://github.com/google/cel-spec/blob/master/doc/extensions/strings.md."
- `StringsLocale` is documented as "If StringsVersion is greater than or equal to 4, this option is ignored."
- **v5 is the cost-accounting version**: `StringsMaxPrecision` doc — "If not set, the default is 100 for version >= 5,
  and no limit for earlier versions" (source comment: "maxPrecision is unbounded (0) for versions < 5 to maintain
  backward compatibility"); and only `>= 5` installs `cel.CostEstimatorOptions(...)` and
  `cel.CostTrackerOptions(interpreter.OverloadCostTracker(...))` for `string_char_at_int`,
  `string_index_of_string[_int]`, `string_last_index_of_string[_int]`, `list_join[_string]`, plus fixed-transform
  estimates (`string_replace`, `string_split*`, `string_substring*`, `string_lower_ascii`, `string_upper_ascii`,
  `string_trim`, `string_reverse`).
- **Important default**: `ext.Strings()` with no options sets `version: math.MaxUint32`
  (`s := &stringLib{version: math.MaxUint32}`), i.e. "all functions are available" *and* the ≥5 behaviours apply
  (default max precision 100, cost estimators/trackers on). Passing an explicit `StringsVersion(4)` opts **out** of
  the cost accounting and out of the precision cap. If our CLI currently calls `ext.Strings()`, behaviour on upgrade
  is: extra cost tracking + a 100-precision default for `format` (a semantic change only for precision >100 clauses).

### How ext functions interact with cost accounting
- The ext libraries are cost-aware by construction, not by default-on magic. Each library installs
  `cel.CostEstimatorOptions(...)` (static, `checker.OverloadCostEstimate`) and, for runtime counting,
  `cel.CostTrackerOptions(interpreter.OverloadCostTracker(overloadID, fn))` — the latter shown in `ext/strings.go`,
  and the same pattern appears across the cost PR series: **encoders #1354**, **ext/lists #1352 / #1360**,
  **ext/math #1353**, plus `ext/sets` static estimates visible in source
  (`checker.OverloadCostEstimate("list_sets_contains_list", estimateSetsCost(1))`, `…intersects…(1)`,
  `…equivalent…(2)` — "equivalence requires potentially two m*n comparisons").
- Consequence: **`cel.CostLimit` and `EvalDetails.ActualCost()` do account for ext function calls** — but only for the
  overloads that carry an estimator/tracker. Anything without one gets the interpreter's default `cost++`
  (`interpreter/runtimecost.go` `default:` branch: "Any functions that don't have a declared cost either here or in
  provided ActualCostEstimator" → cost 1). That under-counting is the practical risk for us: an ext function that is
  genuinely O(n·m) but has no estimator costs 1 at runtime.
- Known heavy/unbounded-cost surfaces and their mitigations:
  - `matches` (regex): runtime tracker computes `strCost * regexCost` with
    `regexCost = ceil(len(pattern) * RegexStringLengthCostFactor)` and a comment that RE2 guarantees linear time:
    "https://swtch.com/~rsc/russ/russ/regexp1.html applies to RE2 implementation supported by CEL" (see
    `interpreter/runtimecost.go`). Regex *constant* compilation is now also boundable via
    `cel.RegexProgramSizeLimit(n)` (v0.31 #1383) plus the `cel.ValidateRegexProgramSizeLimit` AST validator, and
    `cel.OptimizeRegex`/`interpreter.CompileRegexConstants` compile literal patterns at program build time.
    **[INFER]** unbounded by default: with no `RegexProgramSizeLimit` configured, a pathological literal pattern is
    only bounded by the string-length-based cost estimate and the new 100k expression-node/100k-code-point limits.
  - `ext/lists` `lists.range(n)`: OOM guard added v0.29.0 (#1310, `genRange()` max size) and
    `ListsMaxRangeSize(int64)` exposed; the v0.31 note "Correct documented default max value for lists.range (#1392)"
    indicates doc drift was fixed, so the default cap value should be read from source/README before relying on it
    **[GAP: exact default not extracted]**.
  - `ext.Strings.format` precision: capped at 100 by default only at version ≥5 (see above).
  - `sort`/`sortBy`/`distinct`/`flatten`/`join`: static estimators exist at version ≥3 (`list_*_sort`,
    `list_*_sortByAssociatedKeys`, `list_distinct`, `list_flatten`, `list_join`, …).
  - `sets.equivalent` is explicitly costed as 2×m·n comparisons.
  - Non-deterministic `json.encode` was given a **dummy "infinite" cost** in v0.30.0 (#1365: "Fix non-determinism in
    json encoding, add dummy 'infinite' costs") — a rule using it may now be rejected by static/estimated cost
    checks. **[INFER from release-note wording; the exact sentinel value was not read from source — GAP]**

---

## 5. CURRENT STATE (as of 2026-10-06)

- **Latest tagged release is v0.32.0 (2026-08-19).** Evidence: the releases Atom feed's newest entry is v0.32.0
  (`<updated>2026-08-19T19:43:14Z</updated>`, feed `<updated>2026-08-19T17:33:13Z</updated>`) and the tags page's newest
  tag is v0.32.0; pkg.go.dev marks `v0.32.0` as **Default** for `cel.dev/cel-go`.
  ([releases.atom](https://github.com/cel-expr/cel-go/releases.atom), [tags](https://github.com/cel-expr/cel-go/tags),
  [pkg.go.dev/cel.dev/cel-go](https://pkg.go.dev/cel.dev/cel-go)).
- **No v0.33.x and no v0.32.1 tag have shipped.** pkg.go.dev's version list shows only *pseudo-versions* past v0.32.0
  (`v0.32.1-0.20260819210734-…` through `v0.32.1-0.20260903181306-…`, plus `v0.0.0-20260922191641-e390c89c6ef7`),
  which are untagged master commits (master has moved since). So "v0.32.x" currently means exactly **v0.32.0**;
  anything newer would be a pinned pseudo-version (not advisable).
- **Minimum Go version for v0.32.0: `go 1.23.0`** (from the module's own `go.mod`; pkg.go.dev also reports
  "Go: 1.23.0"). No `toolchain` directive in the fetched go.mod.
- **Security advisories in the v0.28–v0.32 window (both already fixed before v0.32):**
  1. **GHSA-gcjh-h69q-9w9g / GO-2026-6094** — "JSON Private Fields Exposed via NativeTypes and ParseStructTag":
     `ext.NativeTypes(ParseStructTag("json"))` did not honour `json:"-"`, registering a readable CEL field literally
     named `"-"` (`dyn(obj)["-"]`), and `newNativeTypes` silently registered nested structs. Weakness CWE-495,
     CVSS 3.1 `5.3 MEDIUM` (`AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:N/A:N`).
     Affected: "v0.22.0 - v0.29.2", patched "v0.30.0" per the GitHub advisory; the Go vulnerability database entry says
     "from v0.22.0 before v0.30.0"; v0.30.0's release notes contain the fix ("Prevent indexing on sentinel json field
     '-'", #1349). **Contradiction recorded**: GitLab Advisory DB and Kodem both state "fixed in 0.29.0" / affected
     "≤ 0.28.1".
     ([GHSA-gcjh-h69q-9w9g](https://github.com/cel-expr/cel-go/security/advisories/GHSA-gcjh-h69q-9w9g),
     [GO-2026-6094](https://pkg.go.dev/vuln/GO-2026-6094),
     [GitLab advisory](https://advisories.gitlab.com/golang/github.com/google/cel-go/GHSA-gcjh-h69q-9w9g/)).
  2. **CVE-2026-83530** — "Uncontrolled Memory Allocation in cel-go" (CWE-789): "A user could provide an expression
     whose string length is longer than the `ParserExpressionSizeLimit()` configured on the CEL environment, and a
     memory allocation would occur proportional to the size of the input before the limit would be checked/enforced."
     CVSS v4.0 6.9 Medium; published 2026-09-09; affected `>= 0, < 0.29.0`; fixed by PR #1302, shipped in **v0.29.0**
     ("Enforce expression size limit before source construction").
     ([OpenCVE](https://app.opencve.io/cve/CVE-2026-83530),
     [Go vuln reference GO-2026-… via PR #1302](https://github.com/cel-expr/cel-go/pull/1302)).
- **Net**: v0.32.0 is clear of both advisories (both patched in v0.29.0/v0.30.0). The current pin at v0.28.0 is exposed
  to both, which is itself an upgrade argument (though our CLI does not appear to use `ext.NativeTypes`).
- No advisory was found that affects v0.30–v0.32 only. **[SINGLE]** — sweeps were GitHub advisories, Go vuln DB,
  OpenCVE and GitLab advisory search results; not exhaustive over every CNA.

---

## 6. MIGRATION GOTCHAS REPORTED BY THE COMMUNITY

1. **Renovate/Dependabot version-only bumps hard-fail** on the module-path change. Real reports:
   - heimdall PR #3483: `go: github.com/google/cel-go@v0.32.0: parsing go.mod: module declares its path as:
     cel.dev/cel-go but was required as: github.com/google/cel-go`
   - kromgo PR #371 (artifact/go.sum failure with the same error)
   - cilium PR #48341 — explicitly unblocked Renovate #48320 by rewriting the import path in the same change
   - GitLab `api/client-go` MR !3010 — same failure, `-d flag is deprecated` noise included.
   Lesson: the go.mod `require` line and every import must move in one commit; don't let a bot split them.
2. **Trying `github.com/cel-expr/cel-go` as a module path also fails** — discussed in issue #1329 before the vanity
   path was chosen: `github.com/cel-expr/cel-go@v0.30.0: parsing go.mod: module declares its path as:
   github.com/google/cel-go but was required as: github.com/cel-expr/cel-go`, plus "pkg.go.dev/github.com/cel-expr/cel-go
   is non-functional… It shows v0.1.0 but none of the fresh ones."
3. **Cost regressions on v0.32.0** — issue #1476 "Non backward compatible cost changes in release 0.32.0", filed with
   Kubernetes-style consumers in mind: "unexpected cost increases break backward compatibility for downstream
   consumers like Kubernetes, causing valid existing policies to fail validation checks" with the estimate deltas
   quoted in §2.4. The maintainer's reply commits to reverting the macro-shape change and shipping the optMap/optFlatMap
   fixes as a *versioned* optional-library bump (PR #1487, post-v0.32.0). Practical consequence for us: **re-derive
   any expected-cost assertions / cost limits after the upgrade, especially for `optional.*` rules.**
4. **Copy-on-write envs (v0.31)**: the feature deliberately shares declaration maps between an env and its parent and
   clones on mutation, and its own commit log lists "Fix race-related issue with copy-on-write mutability check". No
   user-facing breakage reports found, but it is the main new concurrency surface for a caching CLI. **[SINGLE]**
   (release/tag notes only; no independent bug report found).
5. **No `Eval`-side breakage reports** were found for `Program.Eval` / panic recovery / `cel.CostLimit` usage between
   v0.28 and v0.32. The most likely non-mechanical breakage for embedders is the v0.29 `InterpretableV2` /
   `ExecutionFrame` rework: "introduces a new `InterpretableV2` for calling program steps with the `ExecutionFrame`
   rather than the `Activation`" — only matters if you implement `interpreter.Interpretable` yourself
   (`cel.CustomDecorator` is adapted for you via `adaptToV2`; `cel.CustomDecoratorV2` exists for the new shape).
   **[SINGLE on the compat claim; verified only by reading the v0.32 interpreter source snippet for
   `CustomDecorator`/`CustomDecoratorV2`.]**

---

## CONFIDENCE

- §1 module path / mechanics / expr pairing: **high** (official release page + pkg.go.dev + raw go.mod + four
  independent downstream PRs).
- §2 release-by-release inventory: **high** for what each release contains; **medium** for "is this a compile-time
  break for us" (no symbol-level diff of the whole public API was run).
- §3 timestamp acceptance set: **high** for the CEL-side gate (read verbatim from source at tag + the PR's own
  invalid/valid test lists); **medium-high** for the "net" rows that depend on Go's `time.Parse` internals (read from
  Go master source + golang/go#20869, not executed). No Go toolchain was available in this run to execute test cases.
- §4 ext libraries/cost interaction: **high** for existence, version gates and estimator wiring; **medium** for the
  full function-by-function cost table (only the sampled overloads above were read).
- §5 current state / advisories / Go floor: **high** for latest-release and Go floor; **medium** for "no other
  advisories exist".
- §6 community gotchas: **high** for the Renovate/module-path failure class (four independent reports) and for the
  cost-regression issue; **low** for "nobody hit anything else" (absence of evidence).

## GAPS (unverified / follow-up)

1. Exact default value of `ListsMaxRangeSize` (the `lists.range` cap) and the exact sentinel value used for the
   "dummy infinite cost" on `json.encode` (#1365) — read `ext/lists.go` / the encoders cost code.
2. No execution-level verification of the timestamp table: the lowercase `t`/`z` and leap-second rows are source
   inference (Go `parseRFC3339` requires uppercase `T`/`Z` and `0..59` seconds; the layout literal `T` is matched
   case-sensitively). Confirm with a tiny Go program against the pinned Go toolchain before relying on the exact
   error *message* (v0.32 returns the generic conversion error for those shapes, not the new
   `invalid RFC 3339 timestamp` message).
3. Whether `cel-go` v0.28→v0.32 changed any *exported* symbol our package actually imports (`cel.NewEnv` options,
   `interpreter.InterpretableDecorator` signature, `types.Provider`) — needs a symbol-level diff / `go build` after
   the path rewrite, not just release-note reading.
4. Dependency-graph deltas per intermediate release (`antlr4-go/antlr/v4`, `google.golang.org/protobuf`,
   `golang.org/x/text`, `genproto`) — compared only v0.32.0's go.mod; an intermediate bump could interact with the
   project's other pins.
5. The GHSA affected-range contradiction (`< v0.30.0` per GHSA/Go vuln DB vs "fixed in 0.29.0" per GitLab/Kodem) is
   recorded but not adjudicated further; it does not affect v0.32.0.
6. `cel.RegexProgramSizeLimit` default: source shows the limit is only applied when `> 0`; whether any *other* code
   path (e.g. the `interpreter` decorator) imposes a default was not fully traced.
