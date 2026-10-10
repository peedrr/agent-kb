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
- go.mod
status: final
summary: "Module-source and runtime-probe verification of the santhosh v6.0.3 API surface SPEC-004 codes against: error tree, compiler, formats, Validate type acceptance."
tags:
- rfc-002
- spec-004
- json-schema
- santhosh
title: RFC-002-S7 libs-verify — santhosh-tekuri/jsonschema v6.0.3 API surface verification
type: research
updated: "2026-10-10"
---
# Code Context — rfc-002-s7: santhosh-tekuri/jsonschema v6 API surface verification (libs-verify)

**Scope:** runtime + source verification of the exact API SPEC-004 will code against.
**Method:** read the **extracted module source in GOMODCACHE** (not web docs, not memory) and **executed probes** against a throwaway module in `/tmp/jsv6smoke` (outside the repo). Repo left untouched (`git status --porcelain` empty; `go.mod`/`go.sum` unmodified).
**Prior work cited, not redone:** `docs/knowledge-base/kb/research/rfc-002-s6-gapfill-libs.md` §2.2–§2.3 (v6.0.3 compile surface, JSON-only `UnmarshalJSON`, default draft, "unknown schema keywords not rejected").

---

## Files Retrieved

Module source root (abbreviated `$M` below):
`/home/pete/go/pkg/mod/github.com/santhosh-tekuri/jsonschema/v6@v6.0.3/`

1. `$M/validator.go` (lines 15-51, 70-180, 816-977) — `Validate` entry point, root-error wrapping, invalidType gate, format assertion site, `ValidationError` + `ErrorKind`.
2. `$M/compiler.go` (lines 9-215) — `Compiler`, every register/assert method with its doc comment, `AddResource`, `Compile`/`MustCompile`, cache write.
3. `$M/output.go` (lines 15-215) — `Error()/GoString()`, `jsonPtr`/`escape` rendering, `OutputUnit`/`OutputError`, `FlagOutput`/`BasicOutput`/`DetailedOutput`.
4. `$M/util.go` (lines 53-101, 260-275, 407-467) — `escape()` (~0/~1), `jsonPointer` type, `isInteger`, `SchemaValidationError`, `LocalizableError`.
5. `$M/schema.go` (lines 26-58, 116-176) — `Schema` (public `DraftVersion`, `Location`), `jsonType`/`typeOf` (what Go types `Validate` accepts).
6. `$M/loader.go` (lines 142-163, 238-266) — `LoadURLError`, `ResourceExistsError`, `UnmarshalJSON`.
7. `$M/format.go` (lines 11-39, 378-385, 440-513) — `Format` type, built-in format table, `validateDate`/`validateTime`/`validateDateTime` laxity.
8. `$M/vocab.go` (lines 3-99) — `Vocabulary`, `SchemaExt`, `ValidatorContext`, `AddError`.
9. `$M/kind/kind.go` (651 lines; 39 exported kinds) — the public `kind` package; `:13 InvalidJsonValue`, `:27 Schema`, `:41 Group`, `:207 Format`.
10. `$M/go.mod` — `go 1.21`; requires `golang.org/x/text v0.14.0` (runtime) and `dlclark/regexp2 v1.11.0 // used for testing`.
11. `docs/knowledge-base/kb/specs/SPEC-002-structured-validation-report.md` (lines 269-274, 354-360) and `docs/knowledge-base/kb/decisions/ADR-003-santhosh-v6-format-assertion.md` (lines 43-66, 102-114) — the assumption sets cross-checked below.

## Pinned version — dependency is ABSENT

```text
$ grep -n santhosh go.mod go.sum      → no matches
$ go list -m -versions .../jsonschema/v6
  v6.0.0-alpha.1 v6.0.0-beta1 v6.0.0 v6.0.1 v6.0.2 v6.0.3
```
`github.com/santhosh-tekuri/jsonschema/v6` is **not** in `go.mod` or `go.sum` — consistent with SPEC-003 REQ-007 (no validator dep in the loader phase). **VERIFIED.**
The local module cache already holds **v6.0.3** (extracted dir + zip):
`$M/../@v/v6.0.3.info` → `{"Version":"v6.0.3","Time":"2026-06-28T17:38:00Z","Origin":{...,"Hash":"b0fc661f4939578bc429f408c18202573dbafcc4","Ref":"refs/tags/v6.0.3"}}`.
v6.0.3 is the newest v6 ⇒ this report verifies **v6.0.3**. (Note: the `.info` `Time` is the tag/commit time 2026-06-28; the prior report's 2026-08-06 was the GitHub *release* publication date — both true, different clocks. The module README still self-titles "jsonschema v6.0.2"; stale title, not tree evidence.)

---

## 3a. Error tree — **VERIFIED**

`$M/validator.go:960-977` verbatim:
```go
type ValidationError struct {
	// absolute, dereferenced schema location.
	SchemaURL string

	// location of the JSON value within the instance being validated.
	InstanceLocation []string

	// kind of error
	ErrorKind ErrorKind

	// holds nested errors
	Causes []*ValidationError
}

type ErrorKind interface {
	KeywordPath() []string
	LocalizedString(*message.Printer) string
}
```
There is **no `Message`, `KeywordLocation`, or `AbsoluteKeywordLocation` field** on the struct. Keywords come from `ErrorKind.KeywordPath()`; keyword locations are computed only in `output.go:25-36` (`absoluteKeywordLocation`, unexported) and in `OutputUnit`.

Public `kind` package: **YES** — `$M/kind/kind.go`, `package kind`, **39 exported ErrorKind implementations** (not "~60"). Classes: `InvalidJsonValue`, `Schema`, `Group`, `Not/AllOf/AnyOf/OneOf/FalseSchema`, `RefCycle/Reference`, `Type/Enum/Const/Format`, per-keyword kinds (`MinLength`, `Pattern`, `Required`, `Maximum`, …). This matters: `errors.As`/type-switch on `*kind.X` is a supported classification mechanism.

**Tree walk (probe-verified).** `Validate` returns a synthetic root:
```go
// $M/validator.go:34-47
if _, err := vd.validate(); err != nil {
    verr := err.(*ValidationError)
    var causes []*ValidationError
    if _, ok := verr.ErrorKind.(*kind.Group); ok { causes = verr.Causes } else { causes = []*ValidationError{verr} }
    return &ValidationError{SchemaURL: sch.Location, InstanceLocation: nil,
        ErrorKind: &kind.Schema{Location: sch.Location}, Causes: causes}
}
```
Probe (schema with two failing properties + `required`):
```text
root kind=*kind.Schema causes=3 instLoc=[]
  *kind.Required kw=[required] inst=[]                          <LEAF
  *kind.Group    kw=[]         inst=[a]                          (intermediate)
    *kind.MinLength kw=[minLength] inst=[a]                      <LEAF
    *kind.Pattern   kw=[pattern]  inst=[a]                       <LEAF
  *kind.Type      kw=[type]     inst=[b]                         <LEAF
```
⇒ **Leaves = nodes with `len(Causes)==0`**; intermediate nodes are `*kind.Group`/`*kind.Reference` (empty `KeywordPath()`). Enumerate by recursive `Causes` walk. `BasicOutput()` returns a flattened list (it emits one unit for the root plus one per leaf: probe showed 4 leaves ↔ 5 flat units) — convenient but it flattens the tree into `OutputUnit`s and cannot be re-shredded.

Other presentation helpers (`output.go:83-101,120,149,159`): `Error()`, `GoString()`, `LocalizedError(p)`, `LocalizedGoString(p)`, `FlagOutput()`, `BasicOutput()`, `DetailedOutput()`.

**Gotcha (probe-reproduced SIGSEGV): `ErrorKind.LocalizedString(nil)` panics.** `kind/kind.go:22,218` call `p.Sprintf`; `golang.org/x/text/message.(*Printer).Sprintf` dereferences the printer. To render a row message you must either pass a real `*message.Printer` (`message.NewPrinter(language.English)`), making `golang.org/x/text/message` a direct dep, or read `BasicOutput().Error.String()` (which closes over the package's private `defaultPrinter`). SPEC-002's row shape should store `ErrorKind` (or a pre-rendered string), **not** a `Message` field the library does not have.

## 3b. InstanceLocation → RFC 6901 JSON Pointer — **VERIFIED, with a gap**

`InstanceLocation` **is `[]string`** (`validator.go:965`). Tokens are **raw / unescaped**. Probe: property `"a/b"` → `InstanceLocation = []string{"deep", "a/b"}`.

No exported rendering helper exists. The library's own renderer is private:
```go
// $M/output.go:103-110
func jsonPtr(tokens []string) string {
	var sb strings.Builder
	for _, tok := range tokens { sb.WriteByte('/'); sb.WriteString(escape(tok)) }
	return sb.String()
}
// $M/util.go:55-59 — the escaping rule, also private
func escape(tok string) string {
	tok = strings.ReplaceAll(tok, "~", "~0")
	tok = strings.ReplaceAll(tok, "/", "~1")
	return tok
}
```
⇒ akb must re-implement those 3 lines (order matters: `~` before `/`) or take `InstanceLocation` from `BasicOutput()`/`DetailedOutput()`, where it is already escaped. Probe: `BasicOutput()` emitted `InstanceLocation="/deep/a~1b"`, `KeywordLocation="/properties/deep/properties/a~1b/minLength"`. The root error's `InstanceLocation` is `nil` (→ `""` as a pointer).

## 3c. Compiler API — **VERIFIED (signatures + verbatim doc comments)**

Public surface (exhaustive: `^func [A-Z]` over non-test files yields only `NewCompiler`, `UnmarshalJSON`, `LocalizableError`; everything else is a method):
```go
func NewCompiler() *Compiler                                  // compiler.go:21
func UnmarshalJSON(r io.Reader) (any, error)                  // loader.go:255
func LocalizableError(format string, args ...any) error       // util.go:452

func (c *Compiler) DefaultDraft(d *Draft)                     // compiler.go:43
func (c *Compiler) AssertFormat()                             // compiler.go:53
func (c *Compiler) AssertContent()                            // compiler.go:65
func (c *Compiler) RegisterFormat(f *Format)                  // compiler.go:75
func (c *Compiler) RegisterContentEncoding(d *Decoder)       // compiler.go:85
func (c *Compiler) RegisterContentMediaType(mt *MediaType)    // compiler.go:93
func (c *Compiler) RegisterVocabulary(vocab *Vocabulary)      // compiler.go:102
func (c *Compiler) AssertVocabs()                             // compiler.go:112
func (c *Compiler) AddResource(url string, doc any) error     // compiler.go:121
func (c *Compiler) UseLoader(loader URLLoader)                // compiler.go:137
func (c *Compiler) UseRegexpEngine(engine RegexpEngine)       // compiler.go:146
func (c *Compiler) Compile(loc string) (*Schema, error)       // compiler.go:178
func (c *Compiler) MustCompile(loc string) *Schema            // compiler.go:169
func (sch *Schema) Validate(v any) error                      // validator.go:15
```
There is **no `CompileURL`** — `Compile` is the only compile entry (prior report §2.2 agrees). `AddResource` doc, verbatim (`compiler.go:116-120`):
```text
// AddResource adds schema resource which gets used later in reference
// resolution.
//
// The argument url can be file path or url. Any fragment in url is ignored.
// The argument doc must be valid json value.
```
`AssertFormat` doc, verbatim (`compiler.go:47-52`):
```text
// AssertFormat always enables format assertions.
//
// Default Behavior:
// for draft-07: enabled.
// for draft/2019-09: disabled unless metaschema says `format` vocabulary is required.
// for draft/2020-12: disabled unless metaschema says `format-assertion` vocabulary is required.
```
`RegisterFormat` doc, verbatim (`compiler.go:69-74`):
```text
// RegisterFormat registers custom format.
//
// NOTE:
//   - "regex" format can not be overridden
//   - format assertions are disabled for draft >= 2019-09
//     see [Compiler.AssertFormat]
```
`AssertVocabs` doc (`compiler.go:106-111`): "AssertVocabs always enables user-defined vocabularies assertions." `DefaultDraft` (`compiler.go:33-42`) warns the default "will not stay the same overtime" — 2020-12 is the current default (prior report §2.2).

**Thread-safety — NO DOC STATEMENT EXISTS.** `grep -rin "concurren|threadsafe|thread-safe|goroutine|mutex"` over all module `.go` files **and** `README.md` ⇒ **0 hits**. `Schema.Validate` has **no doc comment at all** (`validator.go:13-17`); `Compiler.Compile`'s entire doc comment is `// Compile compiles json-schema at given loc.` (`compiler.go:177`). So:
- "Compile is not thread-safe" — **UNVERIFIED as documentation**; it is *evidenced* by source-visible shared-map mutation on the same `*Compiler` (`c.schemas[sch.up] = sch` at `compiler.go:211`, read at `compiler.go:154`; loader `docs` map write at `loader.go:138`) and, per the prior report §2.3, attributed to upstream issue #207. Separate `*Compiler` instances do not share state.
- "concurrent `Validate` on a compiled `*Schema` is undocumented" — **VERIFIED as undocumented**; it is also *safe by source reading*: `validate` builds a fresh local `validator` and I found **no write to any `*Schema` field anywhere in `validator.go`** (`grep 'sch\.[A-Za-z]* *='` → only a comparison at `:807`).
- Probe: 50 goroutines calling `Validate` on one shared `*Schema` completed with no fatal error. **`-race` could NOT run** (no gcc/CGO in this environment) ⇒ race-level proof is missing.

`Compile` is cached: probe `c.Compile("a.json")` twice returned the **same `*Schema` pointer**; a second `AddResource` with the same URL returns `*ResourceExistsError`.

## 3d. Malformed-schema detection — **VERIFIED (callers can distinguish)**

`AddResource` returns `nil` for content errors; they surface from `Compile` (probe). Typed errors, probe-confirmed with `errors.As`/`%T`:
| input defect | returned type | source |
|---|---|---|
| `{"type":123}`, `{"minLength":"3"}`, bad `pattern` regex | `*jsonschema.SchemaValidationError` | `util.go:440-447` |
| dangling `$ref: "#/nope"` | `*jsonschema.JSONPointerNotFoundError` | `util.go:430-436` |
| unknown `$schema` URL / unresolvable location | `*jsonschema.LoadURLError` | `loader.go:222-231` |
| duplicate `AddResource` URL | `*jsonschema.ResourceExistsError` | `loader.go:243-249` |
| non-JSON doc passed to `AddResource` (`map[any]any`) | `*jsonschema.SchemaValidationError` (`invalid jsonType map[interface {}]interface {}`) | `util.go:440` |

`SchemaValidationError.Error()` = `"<url>" is not valid against metaschema: <err>`; it carries `URL string` + `Err error`. Other compile faults: `UnsupportedDraftError`, `MetaSchemaMismatchError`, `MetaSchemaCycleError`, `ParseURLError`, `ParseIDError`, `ParseAnchorError`, `DuplicateIDError`, `DuplicateAnchorError`, `InvalidRegexError`, `InvalidMetaPointerError`, `InvalidMetaSchemaURLError`, `UnsupportedURLSchemeError`, `UnsupportedVocabularyError`, `AnchorNotFoundError`.
⇒ **VERIFIED:** a caller can classify "schema malformed" (`*SchemaValidationError` plus the structural types above) with a type switch; there is no single sentinel and no positional information (confirmed by prior report §2.2 — `position.go` is the vocabulary `SchemaPath` abstraction, not source offsets; re-confirmed: `position.go:9-83`).

## 3e. Custom format checkers (`Format`) — **VERIFIED**

```go
// $M/format.go:11-18, verbatim
// Format defined specific format.
type Format struct {
	// Name of format.
	Name string

	// Validate checks if given value is of this format.
	Validate func(v any) error
}
```
Registration: `Compiler.RegisterFormat(&jsonschema.Format{Name: "...", Validate: fn})`; `regex` cannot be overridden (`compiler.go:76-78`). **Overriding built-ins such as `date-time`/`date` works** — probe: the registered checker was invoked (receiving a `string`) and its rejection surfaced as a `*kind.Format` leaf with `KeywordPath()==["format"]`.

Probe-verified semantics for ADR-003 I2/I4/I5:
- **`AssertFormat()` is mandatory for 2020-12**: with a checker registered but `AssertFormat` not called, the format keyword is not compiled at all and no check runs (probe: `"not-a-date"` passed).
- **An unregistered/typo'd format name is silently not asserted** (probe: `format:"totally-unknown"` passed) ⇒ akb's format names must be exact; no error for an unknown `format` value.
- `Validate func(v any) error` receives the **raw Go value**, so a strict checker must handle non-string inputs itself (return `nil` to skip or a shaped error).
- Checker errors are wrapped as `kind.Format{Got any; Want string; Err error}` (`kind/kind.go:207-219`) and rendered `'<got>' is not valid <name>: <err>` (probe: `"'nope' is not valid date-time: parsing time ..."`). Passing a checker's own `jsonschema.LocalizableError(...)` keeps the inner text localizable.

Built-in laxity re-confirmed from source **and** probe (substantiates ADR-003 N2):
- `validateDateTime` accepts `'t'` or `'T'` (`format.go:503`) and both `"z"`/`"Z"` (`format.go:440`); leap second allowed when the offset-adjusted time is `23:59:60` (`format.go:484`). Probe: `2023-10-10t09:19:21Z`, `…z` and `1990-12-31T23:59:60Z` all **accepted**.
- Built-in `date` already rejects `2023-02-30` (Go `time.Parse` "day out of range"), so the reason to override `date` is the strict `full-date` profile (I5), not a Feb-30 laxity.

Vocabulary API also exists for the "unknown keyword" problem: `RegisterVocabulary(*Vocabulary)` + `AssertVocabs()` (`compiler.go:97-114`), with `Vocabulary{URL, Schema, Subschemas, Compile}` (`vocab.go:22-39`), `SchemaExt.Validate(ctx *ValidatorContext, v any)` (`vocab.go:44-48`), and `ValidatorContext.AddError(k ErrorKind)`/`AddErrors` (`vocab.go:84-91`) — i.e. **akb can contribute its own `ErrorKind` into the same tree**. `Schema.Extensions []SchemaExt` (`schema.go:101`) is the compiled hook.

## 3f. Validate entry point / accepted document types — **CRITICAL, VERIFIED by probe**

`func (sch *Schema) Validate(v any) error` — no doc comment. Acceptance is decided by `typeOf` (`schema.go:129-156`):
```go
case nil: nullType
case bool: booleanType
case float64, float32: numberType        // NaN/Inf ⇒ invalidType
case json.Number, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64: numberType
case string: stringType
case []any: arrayType
case map[string]any: objectType
default: invalidType
```
Any other Go type becomes `invalidType` and, at `validator.go:92-95`, a **leaf failure** `&kind.InvalidJsonValue{Value: v}` ("invalid jsonType <T>"). Probe results against a typed object schema:
```text
time.Time, json.RawMessage, []byte, map[string]string, map[any]any, struct{...}  → FAIL *kind.InvalidJsonValue
int / int64 / float64                                                            → OK
```
⇒ **`Validate` does NOT accept `time.Time`, `json.RawMessage`, or goccy's native YAML values directly.** akb's frontmatter values must be lowered to JSON shapes (`map[string]any`, `[]any`, `string`, `bool`, `nil`, numeric) before `Validate`; a raw `time.Time` is reported as a validation failure, never coerced. Numbers are tolerant: `type: integer` accepts `float64(3.0)` and `int64(3)` (`isInteger` via `big.Rat` over `fmt.Sprint`, `util.go:272-275`) and rejects `3.5`.

**One exception (probe-confirmed):** a schema compiling to boolean `true` — literally `{}` or `true` — short-circuits at `validator.go:75-78` and **never runs the invalidType check** (`{}` + `time.Time` → `nil`). A `{"$schema": ...}`-only schema is *not* boolean-true (`compileObject` only sets `Bool` for `len(obj)==0`, `compiler.go:264-268`), so the check **does** fire there (`{$schema}` + `time.Time` → `*kind.InvalidJsonValue`).

Document materialization: `UnmarshalJSON(r io.Reader) (any, error)` (`loader.go:253-266`) is `encoding/json` + `UseNumber` + a trailing-content guard; probe raw `{"n": 3}` → `json.Number`. There is **no YAML entry point**; the prior report §2.2's route stands: goccy `yaml.YAMLToJSON([]byte) ([]byte, error)` (`$GOMODCACHE/github.com/goccy/go-yaml@v1.19.2/yaml.go:241`, the repo-pinned goccy) → `jsonschema.UnmarshalJSON`.

---

## Cross-check: SPEC-002 assumption (`SPEC-002…md:271-274`)

> `[ASSUMPTION: santhosh-tekuri/jsonschema v6's error tree is ValidationError{SchemaURL, InstanceLocation []string, ErrorKind, Causes} with a public kind package …]`

**VERIFIED, with two additions and one correction.**
| SPEC-002 claim | Verdict |
|---|---|
| `ValidationError{SchemaURL, InstanceLocation []string, ErrorKind, Causes}` | **VERIFIED verbatim** — same names, order, types (`validator.go:960-972`). |
| `InstanceLocation` is `[]string` | **VERIFIED** (`validator.go:965`) — raw, unescaped tokens. |
| public `kind` package | **VERIFIED** (`$M/kind/kind.go`). |
| "~60 public kinds" (`SPEC-002:355`) | **CORRECTED:** **39** exported ErrorKind types. |
| error tree is directly renderable into rows | **ADDITION 1:** no exported JSON-Pointer helper for `InstanceLocation []string` (`jsonPtr`/`escape` are private) — re-implement `~`→`~0` then `/`→`~1`, or reuse `BasicOutput()` strings. |
| (implied) rows can carry a message | **ADDITION 2:** the struct has **no `Message`**; `ErrorKind.LocalizedString(nil)` **panics**, so rendering needs a `*message.Printer` (new direct `golang.org/x/text/message` dep) or `BasicOutput().Error.String()`. |
| `output.go:114-166 … capitalized AbsoluteKeywordLocation tag` (`SPEC-002:356`) | **VERIFIED** — `output.go:129`: `AbsoluteKeywordLocation string \`json:"AbsoluteKeywordLocation,omitempty"\``. |
| "to be re-verified against the version go.mod pins under ADR-003" | The repo pins **nothing**; verified against **v6.0.3** (latest v6, the only v6 in the local cache). When ADR-003 lands, pin exactly `v6.0.3`. |

**ADR-003 cross-check (adjacent claims that this verification touches):**
- I2 `Compiler.AssertFormat()` — **VERIFIED** and **necessary** (probe: without it, no format assertion for 2020-12).
- I4/I5 + N2 (must override built-in `date-time`; built-in accepts lowercase `t`/`z` and leap seconds) — **VERIFIED** from `format.go:440,484,503` and by probe.
- I8 (`RegisterFormat` before compilation; cache compiled schema per template) — **VERIFIED** for caching (same `*Schema` pointer on re-`Compile`). The "register before compile" ordering is a **high-confidence inference**, not directly probed: formats are looked up during `compileValue`, and a cache hit returns the already-compiled `*Schema` without re-compiling.
- `ADR-003:106` "`Compiler.Compile` is **documented** as not thread-safe" — **CORRECTED:** the module source and README contain **no** concurrency statement of any kind; the compile-unsafety is *evidenced* by shared-map writes (`compiler.go:211`, `loader.go:138`) and externally by upstream issue #207. The second half ("concurrent `Validate` … undocumented") is **VERIFIED**.

---

## Architecture (how akb will use it)

```text
akb template write / lint sweep
  └─ schema document (YAML)  ──goccy YAMLToJSON──▶ []byte JSON
        └─ jsonschema.UnmarshalJSON(r)  → any (json.Number)      [loader.go:255]
        └─ Compiler: NewCompiler → DefaultDraft(Draft2020)
                     → RegisterFormat(akb date-time, date)   [I4/I5, BEFORE compile]
                     → AssertFormat()                        [I2, mandatory 2020-12]
                     → AddResource(url, doc)                 [content errors surface at Compile]
                     → Compile(url) → *Schema                [cached; NOT goroutine-safe on one Compiler]
  └─ page frontmatter  ──lower to JSON shapes──▶ map[string]any
        └─ (*Schema).Validate(doc) → *ValidationError (tree)
              root = *kind.Schema{Location} with Causes = real failures   [validator.go:42-47]
              leaves = len(Causes)==0; classify via ErrorKind.KeywordPath() / *kind.X type switch
              locator = InstanceLocation []string  → escape(~0,~1) → RFC 6901 pointer
              message = ErrorKind.LocalizedString(printer)  (nil printer PANICS) or BasicOutput()
```
Two invariants SPEC-004 must design around: (i) the document must already be JSON-shaped, (ii) the report row cannot carry a library-supplied `Message` field — it must carry `ErrorKind` + a locator, and render on demand.

## Start Here

Open `$M/compiler.go:9-215` first (the whole compile-side contract in one file, doc comments intact), then `$M/validator.go:15-51` (entry + root wrapping) and `$M/validator.go:960-977` (the report contract), then `$M/output.go:103-215` (pointer rendering + output units) and `$M/kind/kind.go` (classification surface).

## Confidence and Gaps

**Confidence — high** on: the `ValidationError`/`ErrorKind`/`kind` shapes (source + probe), `InstanceLocation` type + unescaped tokens, the absence of any exported pointer helper, every Compiler method signature and the quoted doc comments, `AssertFormat` necessity, built-in format override, unknown-format silence, malformed-schema error types, `Validate` type acceptance (`typeOf`), the boolean-`true` short-circuit, `UnmarshalJSON` JSON-only, and Compile caching.
**Confidence — medium-high** on: "register a format after Compile has no effect" (source reading + cache-hit observation, not a direct probe); Validate concurrency safety (no `*Schema` write in `validator.go` + 50-goroutine probe, but no `-race`).

**Gaps / could not verify:**
1. **No `-race` run** — gcc/CGO unavailable. Concurrent `Validate` on one `*Schema` is unproven at the race level (strongly suggested safe by source).
2. **`Compiler` goroutine-safety is undocumented** in the module; the unsafe-shared-map-write reading is mine. Test concurrent `Compile` on one `Compiler` under `-race` before relying on any concurrency.
3. Formats registered **after** a successful `Compile` are not re-read for cache hits — inferred, not directly probed.
4. Override reach into `$ref`-ed/nested subschemas and into `oneOf`/`anyOf` branches was not probed (only top-level `properties`).
5. `AddResource` does not validate the doc's shape up front (probe: `map[any]any` surfaced only at `Compile`) ⇒ the YAML→JSON conversion must happen before `AddResource`; **YAML shape survival** (non-string keys, anchors/aliases, `!!timestamp`) through `YAMLToJSON` is untested here and remains the prior report's §2.2 gap.
6. Verified against the **local module cache**, not a fresh `go mod download` from the proxy in this run; the module zip hash (`ziphash`) and `go.sum` entries were not compared to the public checksum DB. Pin `v6.0.3` and let `go mod tidy` record the hashes.
7. **Version caveat for SPEC-004:** the repo pins nothing; if a different v6.x is chosen later, re-run probes 3a/3c/3e/3f against it. This report is valid for **v6.0.3 only**.
8. Probe harness lives at `/tmp/jsv6smoke` (throwaway module, `GOPROXY=off`, `-mod=mod`); it is outside the repo and was not committed. Repo is unmodified. No supervisor coordination was needed.

```acceptance-report
{
  "criteriaSatisfied": [
    {"id": "criterion-1", "status": "satisfied", "evidence": "Full verification report written to .pi/research/rfc-002-s7/libs-verify.md (<=350 lines) with module-source file:line citations for every claim, verbatim doc comments, a VERIFIED/CORRECTED/UNVERIFIED verdict per item, a per-claim cross-check of SPEC-002's assumed error-tree shape and of ADR-003 I2/I4/I5/I8/N2, a Confidence section, an explicit Gaps list, and residual risks."}
  ],
  "changedFiles": [
    ".pi/research/rfc-002-s7/libs-verify.md"
  ],
  "testsAddedOrUpdated": [
  ],
  "commandsRun": [
    {"command": "grep -n santhosh go.mod go.sum", "result": "passed", "summary": "no matches: jsonschema v6 is absent from go.mod and go.sum (consistent with SPEC-003 REQ-007)"},
    {"command": "go list -m -versions github.com/santhosh-tekuri/jsonschema/v6", "result": "passed", "summary": "v6.0.0-alpha.1, v6.0.0-beta1, v6.0.0, v6.0.1, v6.0.2, v6.0.3 -> verified against v6.0.3, the latest v6 and the only v6 in the local cache"},
    {"command": "read module source in GOMODCACHE (validator.go, compiler.go, output.go, util.go, schema.go, loader.go, format.go, vocab.go, kind/kind.go, go.mod) at v6.0.3", "result": "passed", "summary": "ValidationError/ErrorKind structs, all Compiler signatures with doc comments, pointer escaping, typeOf acceptance, format table, error types all read with exact line numbers"},
    {"command": "grep -rin 'concurren|threadsafe|thread-safe|goroutine|mutex' over module .go + README", "result": "passed", "summary": "0 hits: no thread-safety statement exists anywhere in the v6.0.3 module source or README; Validate has no doc comment at all"},
    {"command": "go run . in /tmp/jsv6smoke (probe: error tree, non-JSON Go types, format override, malformed schemas, cache)", "result": "passed", "summary": "confirmed tree shape and leaf walking, []string raw InstanceLocation requiring ~0/~1 escaping, time.Time/RawMessage/[]byte/map[string]string rejection, AssertFormat necessity, built-in date-time override, SchemaValidationError/JSONPointerNotFoundError/LoadURLError classification, same-pointer Compile cache, {} boolean-true short-circuit, LocalizedString(nil) panic"}
  ],
  "validationOutput": [
    "validator.go:960-972 verbatim: type ValidationError struct { SchemaURL string; InstanceLocation []string; ErrorKind ErrorKind; Causes []*ValidationError } — exactly SPEC-002's assumed shape; no Message/KeywordLocation/AbsoluteKeywordLocation fields.",
    "validator.go:974-977: type ErrorKind interface { KeywordPath() []string; LocalizedString(*message.Printer) string }; public kind package at kind/kind.go with exactly 39 exported ErrorKind types (SPEC-002's '~60' corrected).",
    "Probe tree: root *kind.Schema (InstanceLocation nil) -> Causes = real failures; leaves have len(Causes)==0; intermediates are *kind.Group with empty KeywordPath.",
    "output.go:103-110 jsonPtr + util.go:55-59 escape (~ -> ~0 then / -> ~1) are both unexported; BasicOutput renders InstanceLocation '/deep/a~1b' from []string{\"deep\",\"a/b\"}.",
    "compiler.go:47-52 AssertFormat doc verbatim (2020-12 format assertions disabled unless metaschema requires format-assertion); probe showed no format check runs without AssertFormat even with a registered checker.",
    "compiler.go:69-74 RegisterFormat doc verbatim ('regex' cannot be overridden); probe confirmed override of built-in date-time/date works and an unknown format name is silently unasserted.",
    "schema.go:129-156 typeOf accepts only nil/bool/float32/float64/json.Number/all ints/string/[]any/map[string]any; probe showed time.Time, json.RawMessage, []byte, map[string]string, map[any]any, struct{} all produce *kind.InvalidJsonValue.",
    "validator.go:75-78: a schema compiled to boolean true ({} or true) returns nil before the invalidType check; probe confirmed {} + time.Time == nil, while {$schema}-only still fires the check.",
    "No write to any *Schema field in validator.go (grep 'sch.X =' -> comparison only at :807); 50-goroutine shared-Schema Validate probe clean (no -race available).",
    "kind/kind.go:22,218 call p.Sprintf on the supplied *message.Printer: ErrorKind.LocalizedString(nil) reproduced a nil-pointer SIGSEGV."
  ],
  "residualRisks": [
    "No -race run (gcc/CGO unavailable): concurrent Validate on a shared compiled *Schema is strongly suggested safe by source (no *Schema writes in validator.go; 50-goroutine probe clean) but unproven at the race-detector level.",
    "Compiler goroutine-safety is undocumented in the module; the claim rests on source-visible shared-map mutation (compiler.go:211, loader.go:138) plus upstream issue #207, not on any quoted doc statement.",
    "RegisterFormat called after a successful Compile is inferred to have no effect for cache hits (formats are looked up during compileValue and a cache hit reuses the *Schema) — not directly probed.",
    "Format override reach into $ref-ed or oneOf/anyOf subschemas was not probed; only top-level properties were exercised.",
    "ErrorKind.LocalizedString(nil) panics; rendering a row message forces either a direct golang.org/x/text/message dependency or reuse of BasicOutput().Error.String(). SPEC-002's report row must not assume a Message field exists.",
    "Validate rejects non-JSON Go values (time.Time, json.RawMessage, []byte, map[string]any-incompatible maps) with a validation failure rather than coercing them: akb must lower goccy frontmatter values to JSON shapes before Validate, and no such conversion was implemented or tested here.",
    "AddResource accepts arbitrary any without validating doc shape (map[any]any surfaced only at Compile), and YAML shape survival through goccy YAMLToJSON (non-string keys, anchors, !!timestamp) remains unverified."
  ],
  "reviewFindings": [
    "no blockers: every claim is anchored to module source file:line or to a reproduced probe result, with VERIFIED/CORRECTED/UNVERIFIED marked per item",
    "correction: SPEC-002:355 states 'kind/kind.go (~60 public kinds)' — the actual count is 39 exported ErrorKind types",
    "correction: ADR-003:106 says Compiler.Compile is 'documented as not thread-safe' — no such documentation exists in the v6.0.3 module source or README; the claim is evidenced by shared-map writes and upstream issue #207 only",
    "addition: the report row cannot carry a Message; ErrorKind has none, LocalizedString(nil) panics, and no exported InstanceLocation->JSON-Pointer helper exists"
  ],
  "noStagedFiles": true,
  "diffSummary": "Added one research file, .pi/research/rfc-002-s7/libs-verify.md, containing the module-source + runtime verification of the santhosh-tekuri/jsonschema v6.0.3 API surface. No repository source, go.mod, or go.sum was modified (git status --porcelain empty); the Go probe ran in a throwaway module at /tmp/jsv6smoke, outside the repo.",
  "manualNotes": "Headline: SPEC-002's assumed error-tree shape is VERIFIED verbatim at v6.0.3 (validator.go:960-972: SchemaURL string, InstanceLocation []string, ErrorKind ErrorKind, Causes []*ValidationError) and kind/kind.go is a genuine public package (39 kinds, not ~60). Three things SPEC-004 must design around that the assumption does not cover: (1) no exported helper renders InstanceLocation []string to an RFC 6901 pointer — jsonPtr and escape are both unexported (output.go:103, util.go:55, escape order ~ then /); (2) the struct has no Message field and ErrorKind.LocalizedString(nil) panics, so rendering needs a *message.Printer or BasicOutput(); (3) Validate only accepts JSON-shaped Go values — typeOf (schema.go:129-156) rejects time.Time, json.RawMessage, []byte, map[string]string and map[any]any with a *kind.InvalidJsonValue leaf, while accepting int/int64/float64 (float64 3.0 satisfies type:integer). The one escape hatch is a schema that compiles to boolean true ({} or true), which returns nil before the invalidType gate. Also: dependencies are absent from go.mod/go.sum as SPEC-003 requires; the cache holds v6.0.3 only; AddResource defers content errors to Compile; malformed schemas are classifiable via *SchemaValidationError / *JSONPointerNotFoundError / *LoadURLError / *ResourceExistsError; AssertFormat is mandatory for 2020-12 and an unregistered format name is silently unasserted; RegisterFormat does override built-in date-time (probe) whose looseness (lowercase t/z, 23:59:60) is confirmed in format.go, substantiating ADR-003 N2. The one ADR-003 correction: its claim that thread-unsafety is 'documented' is not supported by the module — there is zero concurrency text in the source or README; the previous report's thread-safety evidence (issue #207) was therefore external, and the -race confirmation is still outstanding because gcc/CGO is unavailable in this environment."
}
```
