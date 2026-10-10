---
grammar: 1
type: research
title: "Library feasibility for a strict TemplateV3 loader — goccy v1.19.2 and santhosh v6.0.3"
status: final
provenance: agent-drafted
run: rfc-002-s6
as_of: "2026-10-09"
created: "2026-10-09"
updated: "2026-10-09"
scope:
- go.mod
- internal/template/**
tags:
- rfc-002
- templatev3
- libraries
summary: "Pinned-version verification of goccy/go-yaml unknown-key detection (no warn mode; Strict() single-key and inline-map panic) and santhosh jsonschema v6 compile surface, with feasibility verdicts."
informs:
- SPEC-003
---

# Research: Library feasibility for a strict TemplateV3 loader (goccy/go-yaml v1.19.2 + santhosh-tekuri/jsonschema v6)

Scope: verify two design facts against **current upstream documentation/source**, not training data.
Repo state read directly: `go.mod` / `go.sum` in `/home/pete/code/projects/tools/llm-wiki/agent-kb/agent-kb`.

Evidence method: source read at the **exact git tag** via `raw.githubusercontent.com/.../v<tag>/...`, plus the `pkg.go.dev` package index for the same version, plus GitHub release/PR APIs for dates and status. No Go code was compiled or executed in this run (no shell tool was available), so every "verified" item below is a documentation/source fact, not a runtime observation. That limitation is carried into the verdict.

| Library | In `go.mod`? | Version researched | Release date |
|---|---|---|---|
| `github.com/goccy/go-yaml` | Yes — direct require, `v1.19.2` | **v1.19.2** (== current latest release) | published `2026-01-08T01:13:06Z` |
| `github.com/santhosh-tekuri/jsonschema/v6` | **No** — absent from both `go.mod` and `go.sum` | **v6.0.3** (current latest) | published `2026-08-06T09:16:56Z` (created `2026-06-28T17:38:00Z`) |

Note on the jsonschema version: v6.0.3 is the newest v6 release; the previous release v6.0.2 was published `2025-05-23T14:56:19Z`. I read the v6.0.3 tree as the authoritative surface. `compiler.go` at v6.0.3 and v6.0.2 are byte-identical in size (7942 chars, distinct blob SHAs but same length), and every other file I cite here was read from the v6.0.3 tag — so the API described is the current one, not a stale one. Because the library is **not yet pinned anywhere in this repo**, the exact API is still a free choice; a pin is recommended (see Missing evidence).

---

## 1. goccy/go-yaml at v1.19.2

### 1.1 Version / date evidence

- `go.mod` pins `github.com/goccy/go-yaml v1.19.2` (direct require block, alongside `cel-go`, `goldmark`, `goldmark/frontmatter`).
- Release `v1.19.2` published `2026-01-08T01:13:06Z` — https://github.com/goccy/go-yaml/releases/tag/v1.19.2 (release body: "Fix anchor reference regression in nested structures", PR #839). API: https://api.github.com/repos/goccy/go-yaml/releases/tags/v1.19.2
- The 3 most recent releases are v1.19.2 (2026-01-08), v1.19.1 (2025-12-17), v1.19.0 (2025-11-29) — so **v1.19.2 is the current latest release**, and the pinned version is exactly the documented one. No "docs only cover latest, pinned is older" caveat is needed here; the inverse holds.
- All API claims below were read at the pinned tag, e.g. https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/option.go, and cross-checked against the v1.19.2 package index: https://pkg.go.dev/github.com/goccy/go-yaml@v1.19.2

### 1.2 Unknown / undeclared field detection — what exists

- **`yaml.Strict()` and `yaml.DisallowUnknownField()` are the same behaviour.** Both option functions set the identical private flag; `Strict()`'s entire doc comment is "Strict enable DisallowUnknownField".
  ```go
  // Strict enable DisallowUnknownField
  func Strict() DecodeOption {
      return func(d *Decoder) error {
          d.disallowUnknownField = true
          return nil
      }
  }
  // DisallowUnknownField causes the Decoder to return an error when the destination
  // is a struct and the input contains object keys which do not match any
  // non-ignored, exported fields in the destination.
  func DisallowUnknownField() DecodeOption { ... d.disallowUnknownField = true ... }
  ```
  https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/option.go
- **There is no plural `DisallowUnknownFields`, no `yaml.Decoder.DisallowUnknownFields(...)` method, and no `KnownFields(bool)` equivalent** (that is `gopkg.in/yaml.v3`'s API, not goccy's). The v1.19.2 index lists exactly four `Decoder` methods — `Decode`, `DecodeContext`, `DecodeFromNode`, `DecodeFromNodeContext` — and the `DecodeOption` constructors are a closed set: `AllowDuplicateMapKey`, `AllowFieldPrefixes`, `CommentToMap`, `CustomUnmarshaler`, `CustomUnmarshalerContext`, `DisallowUnknownField`, `RecursiveDir`, `ReferenceDirs`, `ReferenceFiles`, `ReferenceReaders`, `Strict`, `UseJSONUnmarshaler`, `UseOrderedMap`, `Validator`.
  https://pkg.go.dev/github.com/goccy/go-yaml@v1.19.2 (index of `type Decoder` / `type DecodeOption`); https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/option.go
  The maintainer-adjacent confusion about "two options doing exactly the same without the documentation saying so" is acknowledged in issue #810's footnote — https://github.com/goccy/go-yaml/issues/810
- **Semantics: error, not collect.** In `decodeStruct`, after all known fields are consumed, remaining keys are checked and the function **returns on the first offending key**:
  ```go
  if len(unknownFields) != 0 && d.disallowUnknownField && src.GetToken() != nil {
      for key, node := range unknownFields {      // <-- range over a Go map
          var ok bool
          for _, prefix := range d.allowedFieldPrefixes {
              if strings.HasPrefix(key, prefix) { ok = true; break }
          }
          if !ok {
              return errors.ErrUnknownField(fmt.Sprintf(`unknown field "%s"`, key), node.GetToken())
          }
      }
  }
  ```
  https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/decode.go
  **Researcher inference (medium-high confidence, from this source text):** because the loop `range`s a `map[string]ast.Node`, when *two or more* keys are unknown the reported key is whichever the runtime happens to visit first — i.e. the key named in the error is **non-deterministic across runs**, and only **one** unknown key is ever reported per struct level. I did not execute code to confirm the ordering effect; the non-determinism follows from Go's specified map iteration randomness.
- **Exact error type and message.** The option produces a `*yaml.UnknownFieldError`:
  ```go
  type UnknownFieldError struct { Message string; Token *token.Token }
  func ErrUnknownField(msg string, tk *token.Token) *UnknownFieldError
  func (e *UnknownFieldError) GetToken() *token.Token
  func (e *UnknownFieldError) GetMessage() string
  func (e *UnknownFieldError) Error() string  // == e.FormatError(false, /*inclSource*/ true)
  ```
  https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/internal/errors/error.go ; re-exported in the `yaml` package as an alias — `UnknownFieldError = errors.UnknownFieldError` in https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/error.go
  The message text is literally ``unknown field "%s"``, so the offending key **is** named. The rendered form is documented by the package's own example output:
  ```
  [3:1] unknown field "unknown"
     1 | ---
     2 | simple: string
  >  3 | unknown: string
         ^
  ```
  https://pkg.go.dev/github.com/goccy/go-yaml@v1.19.2 (`Decoder.Decode (DisallowUnknownField)` example)
- **Scope of the check.** It applies only when the decode target is a **struct** (a map target is not checked) and is applied **per struct level**, descending recursively because `decodeStruct` recurses; it is suppressed when the source node has a nil token (inline-struct case, per the in-source comment "Ignore unknown fields when parsing an inline struct"), and `yaml.AllowFieldPrefixes(prefixes...)` exempts keys by prefix.
  https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/decode.go ; https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/option.go
- **⚠ Known defect at this exact version: `Strict()` + an `,inline` map field panics.** `deleteStructKeys` (called from the inline branch when `disallowUnknownField` is set) unguards-then-calls `structFieldMap(structType)` → `structType.NumField()`, which panics for a map kind. It is present in the pinned source:
  ```go
  func (d *Decoder) deleteStructKeys(structType reflect.Type, unknownFields map[string]ast.Node) error {
      if structType.Kind() == reflect.Ptr { structType = structType.Elem() }   // no map guard
      structFieldMap, err := structFieldMap(structType)                        // -> NumField() -> panic on map
  ```
  https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/decode.go
  Reported as issue #810 with the `gov1.18.0` stack trace (`reflect: NumField of non-struct type map[string]interface {}`), and the fix (PR #849, "Fix panic with DisallowUnknownField and inlined maps") is **still open and unmerged** (`"state":"open"`, `"merged":false`, `"merged_at":null`, updated 2026-06-23) — https://github.com/goccy/go-yaml/issues/810 , https://api.github.com/repos/goccy/go-yaml/pulls/849
  **Design consequence:** a TemplateV3 Go struct must not contain a `yaml:",inline"` map field if it will be decoded with `Strict()`. (The current `internal/template.Template`/`Schema` structs use named fields only — `Frontmatter map[string]FieldSchema` is a named field, not inline — so the defect is not currently triggered. Source read: `internal/template/template.go`.)

### 1.3 Can unknown-key detection WARN while continuing? — No.

**No option collects unknown fields as warnings.** Direct evidence: the complete `DecodeOption` constructor set at v1.19.2 (enumerated in §1.2) contains nothing that reports without failing, and the source path for an unknown key is an unconditional `return errors.ErrUnknownField(...)`. The only nearby knob, `AllowFieldPrefixes`, *whitelists* prefixes (suppressing the error) rather than reporting them.
https://pkg.go.dev/github.com/goccy/go-yaml@v1.19.2 ; https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/decode.go

**So the map-diff fallback is required for warn-only-at-load. Confirmed feasible, and the AST does carry line numbers.**

Supporting API, all verified at v1.19.2:

```go
// package github.com/goccy/go-yaml
func UnmarshalWithOptions(data []byte, v interface{}, opts ...DecodeOption) error
func NodeToValue(node ast.Node, v interface{}, opts ...DecodeOption) error
func FormatError(e error, colored, inclSource bool) string
func (d *Decoder) DecodeFromNode(node ast.Node, v interface{}) error
func (d *Decoder) DecodeFromNodeContext(ctx context.Context, node ast.Node, v interface{}) error

type Error interface {              // re-export of internal/errors.Error
    error
    GetToken() *token.Token
    GetMessage() string
    FormatError(bool, bool) string
}

// package github.com/goccy/go-yaml/parser
func ParseBytes(bytes []byte, mode Mode, opts ...Option) (*ast.File, error)
func ParseFile(filename string, mode Mode, opts ...Option) (*ast.File, error)
func Parse(tokens token.Tokens, mode Mode, opts ...Option) (*ast.File, error)
const ParseComments Mode = 1 << iota

// package github.com/goccy/go-yaml/token
type Position struct { Line int; Column int; Offset int; IndentNum int; IndentLevel int }
// token.Token has .Position token.Position among its fields

// package github.com/goccy/go-yaml/ast
type Node interface {
    GetToken() *token.Token
    Type() NodeType
    GetPath() string
    // ... String()/io.Reader, AddColumn, SetComment, GetComment, SetPath
}
type File struct { Name string; Docs []*DocumentNode }
type DocumentNode struct { *BaseNode; Start *token.Token; End *token.Token; Body Node }
type MappingNode struct { *BaseNode; Start *token.Token; End *token.Token; IsFlowStyle bool; Values []*MappingValueNode; FootComment *CommentGroupNode }
type MappingValueNode struct { *BaseNode; Start *token.Token; Key MapKeyNode; Value Node; ... }
func (n *MappingNode) GetToken() *token.Token          // returns n.Start
func (n *MappingValueNode) GetToken() *token.Token     // returns n.Start
```
Sources: https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/yaml.go , https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/parser/parser.go , https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/token/token.go , https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/ast/ast.go , https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/internal/errors/error.go

Two facts worth calling out:

- **There is no `yaml.Node` type.** `gopkg.in/yaml.v3`'s `yaml.Node` has no goccy equivalent; the node types live in the separate `ast` package. The v1.19.2 index for the `yaml` package contains no `type Node`. (The `ast` package *is* reachable to callers, and the package's own `NodeToValue` example uses `f.Docs[0].Body`.) https://pkg.go.dev/github.com/goccy/go-yaml@v1.19.2
- **Top-level key → line/column is a two-line walk:** `parser.ParseBytes(src, 0)` → `f.Docs[0].Body` as `*ast.MappingNode` → for each `*ast.MappingValueNode`, `mv.Key.GetToken().Position.Line` / `.Column`. (The `MappingNode` even derives a start position from `Values[0].Key.GetToken().Position` internally — confirming the key token carries the position. https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/ast/ast.go)
  An alternative for locating a *specific* key by path is `yaml.PathString("foo")` + `(*Path).FilterNode(ast.Node)` / `(*Path).ReadNode(io.Reader)` / `(*Path).AnnotateSource(source, colored)` — https://pkg.go.dev/github.com/goccy/go-yaml@v1.19.2 (`type Path`).
- **A third, in-library route to warn-and-continue** exists without the AST package: implement `yaml.NodeUnmarshaler` on the config struct —
  ```go
  type NodeUnmarshaler interface { UnmarshalYAML(ast.Node) error }
  ```
  which hands your code the node for the struct so you can collect unknown keys and return `nil`. https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/yaml.go — **Researcher interpretation:** this is viable but couples your struct to the `ast` package and bypasses goccy's own field mapping, so the map-diff + AST walk is the cleaner fallback for a loader that must also keep loading.

### 1.4 Duplicate-key detection

**Supported, and on by default — twice over (parser level and decoder level).**

- Parser level (runs first, so this is what a duplicate key actually surfaces as during `yaml.Unmarshal`):
  ```go
  func (p *parser) validateMapKey(ctx *context, tk *token.Token, keyPath string, colonTk *Token) error {
      if !p.allowDuplicateMapKey {
          if n, exists := p.pathMap[keyPath]; exists {
              pos := n.GetToken().Position
              return errors.ErrSyntax(
                  fmt.Sprintf("mapping key %q already defined at [%d:%d]", tk.Value, pos.Line, pos.Column), tk)
          }
      }
      ...
  }
  ```
  → `*yaml.SyntaxError`, message names the key **and the line:column of the earlier definition**. https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/parser/parser.go
- Decoder level:
  ```go
  func (d *Decoder) validateDuplicateKey(keyMap map[string]struct{}, key interface{}, keyNode ast.Node) error {
      if !d.allowDuplicateMapKey {
          if _, exists := keyMap[k]; exists {
              return errors.ErrDuplicateKey(fmt.Sprintf(`duplicate key "%s"`, k), keyNode.GetToken())
          }
      }
      ...
  }
  ```
  → `*yaml.DuplicateKeyError` (struct `{Message string; Token *token.Token}`). https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/decode.go , https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/internal/errors/error.go
- **Opt-out:** `yaml.AllowDuplicateMapKey()` (and `parser.AllowDuplicateMapKey()` at the parser level). Defaults are `allowDuplicateMapKey: false` in `NewDecoder`. The historical `yaml.DisallowDuplicateKey` was **removed** (not deprecated) in PR #531 / v1.14.0 — do not reference it. https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/option.go , https://github.com/goccy/go-yaml/issues/226 , https://github.com/goccy/go-yaml/pull/609
- **Researcher inference (not a source statement):** because the parser check precedes decoding, a duplicate key in a TemplateV3 document will most likely surface as `*yaml.SyntaxError` ("mapping key ... already defined at [l:c]") rather than `*yaml.DuplicateKeyError`, even though both paths are armed by default. I did not run code to confirm which error reaches the caller.

**CONFIDENCE — section 1 (goccy/go-yaml):** **high** on the option set, exact signatures, error types, message text, duplicate-key defaults, and the absence of any warn-and-continue mode (all read at the exact pinned tag and corroborated by the v1.19.2 index). **Medium-high** on the two inferences: non-deterministic first-unknown-key reporting, and `SyntaxError` (not `DuplicateKeyError`) being the duplicate-key error that surfaces in practice. **Medium** on the inline-map panic *at v1.19.2*: the unguarded `NumField` call is present in the pinned source and the fix PR is unmerged, but I could not execute the reproducer.

---

## 2. santhosh-tekuri/jsonschema v6 at v6.0.3

### 2.1 Version / date evidence

- **Not currently a dependency.** `github.com/santhosh-tekuri/jsonschema` appears in neither the direct nor the indirect require blocks of `go.mod`, and has no entry in `go.sum`. Adopting it is a new pin.
- Latest v6 release: **v6.0.3**, `published_at 2026-08-06T09:16:56Z` (`created_at 2026-06-28T17:38:00Z`) — https://api.github.com/repos/santhosh-tekuri/jsonschema/releases?per_page=12
  Prior: v6.0.2 `2025-05-23T14:56:19Z`, v6.0.1 `2024-06-06T18:00:53Z`, v6.0.0 `2024-06-04T17:04:25Z`. (Older v5.x and `cmd/jv/v*` tags appear later in the list.)
- The repo's `README.md` **at the v6.0.3 tag still carries the title "jsonschema v6.0.2"** — a stale doc title, not evidence that the tree is v6.0.2. I verified the API at the v6.0.3 tag directly. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/README.md
- The v6.0.3 file list contains **no `errors.go` and no `validation.go`**; compile errors live in `roots.go`/`loader.go`/`compiler.go` and validation errors in `validator.go` + `output.go`. (Useful when grepping upstream.) https://api.github.com/repos/santhosh-tekuri/jsonschema/contents?ref=v6.0.3

### 2.2 Compile API surface

```go
// package jsonschema  (github.com/santhosh-tekuri/jsonschema/v6)

func NewCompiler() *Compiler

func UnmarshalJSON(r io.Reader) (any, error)          // JSON ONLY — see below
func LocalizableError(format string, args ...any) error

func (c *Compiler) AddResource(url string, doc any) error
func (c *Compiler) Compile(loc string) (*Schema, error)
func (c *Compiler) MustCompile(loc string) *Schema

func (c *Compiler) DefaultDraft(d *Draft)             // fallback draft when "$schema" is missing
func (c *Compiler) AssertFormat()                     // always enable format assertions
func (c *Compiler) AssertVocabs()                     // always enable user-defined vocabularies
func (c *Compiler) AssertContent()                    // enable contentEncoding/contentMediaType/contentSchema
func (c *Compiler) RegisterFormat(f *Format)
func (c *Compiler) RegisterContentEncoding(d *Decoder)
func (c *Compiler) RegisterContentMediaType(mt *MediaType)
func (c *Compiler) RegisterVocabulary(vocab *Vocabulary)
func (c *Compiler) UseLoader(loader URLLoader)
func (c *Compiler) UseRegexpEngine(engine RegexpEngine)

var Draft4, Draft6, Draft7, Draft2019, Draft2020 *Draft   // package-level exported drafts

func (sch *Schema) Validate(v any) error
```
Sources: https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/compiler.go , https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/loader.go , https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/validator.go , https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6@v6.0.3

**Parsing a schema from raw bytes — the critical constraint: `UnmarshalJSON` is JSON-only.**
```go
// UnmarshalJSON unmarshals into [any] without losing
// number precision using [json.Number].
func UnmarshalJSON(r io.Reader) (any, error) {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	...
}
```
https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/loader.go — it is `encoding/json` with `UseNumber`, and it additionally rejects trailing content. **There is no YAML entry point in the library.** The `jv` CLI's "support both json and yaml files" is a CLI-level feature, not a library one. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/README.md
→ For a YAML schema document you must convert first: either `yaml.YAMLToJSON(bytes)` from the already-pinned goccy library, or decode to `map[string]any` and re-marshal as JSON, then `jsonschema.UnmarshalJSON`.

Canonical in-memory flow, from the library's own examples:
```go
schema, err := jsonschema.UnmarshalJSON(strings.NewReader(`{"type":"string","format":"palindrome"}`))
c := jsonschema.NewCompiler()
c.RegisterFormat(&jsonschema.Format{Name: "palindrome", Validate: validatePalindrome})
c.AssertFormat()
if err := c.AddResource("schema.json", schema); err != nil { ... }   // doc must be a valid JSON value
sch, err := c.Compile("schema.json")
err = sch.Validate(inst)
```
https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/example_test.go
`AddResource(url, doc)` notes: "The argument url can be file path or url. Any fragment in url is ignored. The argument doc must be valid json value." Re-adding the same URL returns `*ResourceExistsError`. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/compiler.go , https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/loader.go

**Default draft — yes, 2020-12 is the default, and it is explicitly "latest".**
- Docs: "If `$schema` property is missing, it uses latest draft which currently implemented by this library." plus `Compiler.DefaultDraft`'s comment: "By default, this library uses the latest draft supported. The use of this option is HIGHLY encouraged... The current default value will not stay the same overtime." https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/schema.go , https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/compiler.go
- Source: `draftLatest = Draft2020`, and `Draft2020 = &Draft{version: 2020, url: "https://json-schema.org/draft/2020-12/schema", ...}`. `draftFromURL("json-schema.org/schema")` also resolves to `draftLatest`. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/draft.go
- CLI corroboration of intent: `jv -d, --draft version ... (default 2020)`. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/README.md

**How to assert a specific draft.** `DefaultDraft(jsonschema.Draft2020)` sets only the *fallback* used when `$schema` is absent — a document that declares `$schema` overrides it. There is **no** "force this draft regardless of `$schema`" option in the Compiler API. **Researcher inference (high confidence, from source):** the supported way to assert a draft is post-compile, using the exported field — `compileValue` sets `sch.DraftVersion = res.dialect.draft.version` (2020 for 2020-12), so a loader can `Compile` then refuse unless `sch.DraftVersion == 2020`. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/compiler.go , https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/schema.go (`Schema.DraftVersion int`, exported)

**Format / vocabulary assertion controls.**
- `AssertFormat()` — "AssertFormat always enables format assertions. Default Behavior: for draft-07: enabled. for draft/2019-09: disabled unless metaschema says `format` vocabulary is required. for draft/2020-12: disabled unless metaschema says `format-assertion` vocabulary is required." https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/compiler.go
- `AssertVocabs()` — "AssertVocabs always enables user-defined vocabularies assertions." Same draft-conditional defaults. Same URL.
- `AssertContent()` — content assertions are "always disabled" by default. Same URL.
- There is **no** v6 `AssertVocabs`-vs-v5 divergence to worry about: v5's `Compiler.AssertFormat`/`AssertVocabs` design carried into v6 unchanged, and `AssertContent` is added. (v5's API differed chiefly in `AddResource(url string, r io.Reader)`, a `Draft *Draft` struct field, and a `CompileString` helper — none of which exist in v6. Any snippet you find with `AddResource(url, io.Reader)` is v5 and will not compile against v6.)

**Shape of errors.**

Compile errors — a concrete type plus typed sentinels:
```go
type SchemaValidationError struct { URL string; Err error }
func (e *SchemaValidationError) Error() string
// also, among others:
//   LoadURLError{URL string; Err error}          UnsupportedDraftError{URL string}
//   MetaSchemaCycleError{URL string}             MetaSchemaMismatchError{URL string}
//   DuplicateAnchorError{...}                    DuplicateIDError{...}
//   AnchorNotFoundError{URL, Reference string}   InvalidMetaSchemaURLError{URL string; Err error}
//   InvalidRegexError / InvalidJsonPointerError / JSONPointerNotFoundError / ParseIDError
//   ParseAnchorError / ParseURLError / ResourceExistsError / UnsupportedVocabularyError
//   UnsupportedURLSchemeError
```
https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6@v6.0.3 ; https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/roots.go ; https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/loader.go
`Compile` returns these as a plain `error` (the docs note "Returned error can be `*SchemaError`" in older text; the v6.0.3 surface names it `*SchemaValidationError{URL, Err}`). **No positional information is attached to compile errors** — `position.go` in this repo is unrelated (`Position`/`Prop`/`Item`/`SchemaPath` are a *schema-path* abstraction for vocabulary subschema discovery, not source positions). https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/position.go

Validation errors — a **tree**, not a flat list:
```go
type ValidationError struct {
    SchemaURL        string             // absolute, dereferenced schema location
    InstanceLocation []string           // path of the JSON value within the instance
    ErrorKind        ErrorKind
    Causes           []*ValidationError // nested errors
}
type ErrorKind interface {
    KeywordPath() []string
    LocalizedString(*message.Printer) string
}
// presentation
func (e *ValidationError) Error() string
func (e *ValidationError) GoString() string
func (e *ValidationError) LocalizedError(p *message.Printer) string
func (e *ValidationError) LocalizedGoString(p *message.Printer) string
// machine-readable outputs (JSON-Schema output formats)
func (e *ValidationError) FlagOutput() *FlagOutput
func (e *ValidationError) BasicOutput() *OutputUnit
func (e *ValidationError) DetailedOutput() *OutputUnit

type OutputUnit struct {
    Valid                   bool         `json:"valid"`
    KeywordLocation         string       `json:"keywordLocation"`
    AbsoluteKeywordLocation string       `json:"AbsoluteKeywordLocation,omitempty"`
    InstanceLocation        string       `json:"instanceLocation"`
    Error                   *OutputError `json:"error,omitempty"`
    Errors                  []OutputUnit `json:"errors,omitempty"`
}
```
Sources: https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/validator.go , https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/output.go
Practical note: `Validate` wraps a group failure in a synthetic `&kind.Schema{}` root whose `Causes` are the real errors (visible in `validator.go`'s `validate`), so a caller wanting "all failures" should walk `Causes` (or call `BasicOutput()` for the flattened list) rather than reading only the top-level `ErrorKind`.

**Does compiling reject unknown/undeclared schema keywords? No.**
Two independent pieces of evidence:
1. Compilation is vocabulary-driven: a keyword is handled only if an active vocabulary claims it (`Vocabulary.Schema`, `Vocabulary.Subschemas`, `Vocabulary.Compile`); otherwise it is simply not compiled, with no error. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/vocab.go
2. Subschemas *are* meta-validated at compile time (`roots.ensureSubschema` → `rr.validate(rClone, v, up.ptr)`), which is why `*SchemaValidationError`/`MetaSchemaMismatchError` exist — but the embedded, official 2020-12 metaschemas contain **no** `unevaluatedProperties: false` (nor a top-level `additionalProperties: false`), so unknown keywords are not caught by that route either. Verified by reading the embedded metaschemas: `meta/core` (properties list only), `meta/validation` (no `unevaluatedProperties` match at all), `meta/unevaluated` (defines only the two `unevaluated*` keywords). The outer 2020-12 schema composes them via `allOf` and likewise adds no blanket closure. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/roots.go , and `metaschemas/draft/2020-12/schema`, `.../meta/core`, `.../meta/validation`, `.../meta/unevaluated`
→ **A strict TemplateV3 schema check needs a custom metaschema/vocabulary** (e.g. a dialect whose metaschema sets `unevaluatedProperties: false`, or `Compiler.RegisterVocabulary` + `AssertVocabs`), or an independent top-level key check. This must not be assumed away.

### 2.3 Is compilation cheap / cacheable? Is `CompiledSchema` reuse sync-safe?

- **Reuse of a compiled `*Schema` is the intended pattern and is cheap.** `Schema` is a plain struct of pointers/values (`DraftVersion`, `Location`, `Properties map[string]*Schema`, `Ref`, ...); `Validate` builds a fresh local `validator` per call and does not mutate the schema:
  ```go
  func (sch *Schema) Validate(v any) error { return sch.validate(v, nil, nil, nil, false, nil) }
  ```
  https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/validator.go , https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/schema.go
  **Researcher inference (medium-high):** `Schema.Validate` should therefore be safe to call concurrently — `validator`/`scope`/`uneval` are per-call locals and no write to `*Schema` occurs in the validate path I read. I found **no upstream statement** promising concurrency safety for validation, so treat this as an inference to be smoke-tested under `-race`, not a documented guarantee.
- **Compilation is explicitly NOT thread-safe.** Maintainer response to a `fatal error: concurrent map read and map write` in `compiler.go`'s `enqueue`: "Compilation is not threadsafe. So you need to use mutex". Issue #207 is closed with the `wontfix` label (closed `2025-05-23`). https://github.com/santhosh-tekuri/jsonschema/issues/207
  → Compile schemas single-threaded (e.g. at process start / under a mutex) and share the resulting `*Schema`s.
- **Compilation is a real, repeated cost, and the `Compiler` accumulates state.** `doCompile` maintains an internal `queue` and caches results into `c.schemas[up]`, and `AddResource` refuses a URL already present (`ResourceExistsError`) — so one `Compiler` per schema-set, reused, is the natural pattern; re-`Compile`-ing the same `loc` hits the `c.schemas` cache via `enqueue`. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/compiler.go
- Package initialization itself compiles the metaschemas once in `init()` (`d.sch = c.MustCompile(d.url)` for all five drafts) — a fixed one-time cost. https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/v6.0.3/draft.go

**CONFIDENCE — section 2 (jsonschema v6):** **high** on the API surface, signatures, default draft (2020-12), error types and the `ValidationError` tree, the JSON-only `UnmarshalJSON`, the absence of a force-a-draft option, and the "unknown keywords are not rejected" finding (read at the v6.0.3 tag + embedded metaschemas). **High** on "Compilation is not threadsafe" (direct maintainer statement). **Medium** on validation-path concurrency safety (researcher inference from source; upstream is silent). **Medium-high** on `Schema.DraftVersion` being the practical draft-assertion mechanism (inference from two exported/source facts).

---

## FEASIBILITY VERDICT

### (a) Strict-refuse-at-write with key-naming messages — **feasible, but `yaml.Strict()` alone is not sufficient**

- `yaml.Strict()` (≡ `yaml.DisallowUnknownField()`) does produce a **key-naming** error (`unknown field "x"`) of type `*yaml.UnknownFieldError`, rendered with `[line:col]` plus a caret-annotated source snippet. For a fail-fast write path, this is one line of code:
  ```go
  err := yaml.UnmarshalWithOptions(src, &tmpl, yaml.Strict())
  ```
  (or `yaml.NewDecoder(r, yaml.Strict())`), then `errors.As(err, &yaml.UnknownFieldError)` to special-case the message.
- **Two caveats make it insufficient as the whole mechanism:**
  1. It reports **only one** unknown key per struct level, and *which* one is non-deterministic when several are unknown (map-iteration order). The project's own convention is the opposite — `AGENTS.md`: "All-errors aggregation on write: ALL failed rules reported". goccy's strict path cannot satisfy that; the map-diff path can, deterministically and with a line number per key.
  2. `Strict()` + an `,inline` map field **panics** at v1.19.2 (fix PR unmerged) — so the refusals must not be allowed to depend on a struct shape containing inline maps.
- **Recommended shape:** enforce strictness with the **map-diff + AST walk** (same machinery the warn-only path needs), and optionally keep `yaml.Strict()` as defence-in-depth on the typed struct:
  1. `parser.ParseBytes(src, 0)` → `f.Docs[0].Body` → `*ast.MappingNode`.
  2. Decode into `map[string]any` (or the typed struct) as today.
  3. Compare `map` keys against the allowed top-level key set; for each offending key, find its `*ast.MappingValueNode` and read `Key.GetToken().Position.Line`/`.Column`.
  4. Refuse (exit 1) listing **every** offending key with its line:column — a superset of what `Strict()` provides, plus deterministic output.
- **`yaml.Strict()` vs map-diff, in API terms:** `Strict()` = 1 option, one error, no line list, no aggregation, panic risk with inline maps. Map-diff = ~15 lines (`parser.ParseBytes` + `ast.MappingNode.Values` + `token.Position`), all offenders, deterministic, no panic risk, and it *also* powers (b). The marginal cost of map-diff is small and it strictly dominates.

### (b) Warn-only-at-load — **feasible, and only via the map-diff fallback**

Confirmed: **the library cannot warn-and-continue.** All unknown-key handling terminates in `return errors.ErrUnknownField(...)`; the complete `DecodeOption` set at v1.19.2 offers no collect/report mode.
→ Implement it with the map-diff approach in (a): decode normally (no `Strict()`), then independently parse the AST and diff keys, emitting warnings for unknown keys with their line:column. Because the AST walk yields per-key positions, the warning text can be as precise as the error text — e.g. `warning: unknown field "x" at line 7, column 1; ignored`. Same code path for both policies; the only differences are whether the command exits non-zero and whether the offending keys are dropped or retained.

### (c) Compile-at-load vs opaque-passthrough — cost in API terms

**Compile-at-load (jsonschema v6):**
- Cost is **3 API calls plus one conversion**: `jsonschema.UnmarshalJSON(bytes.NewReader(jsonBytes))` → `c := jsonschema.NewCompiler(); c.DefaultDraft(jsonschema.Draft2020); c.AddResource(url, doc)` → `sch, err := c.Compile(url)`.
- **You must supply the JSON conversion yourself** — the library has no YAML entry point (`UnmarshalJSON` is `encoding/json` + `UseNumber`). Cheapest path with what is already pinned: `yaml.YAMLToJSON(schemaBytes)` (goccy) then `jsonschema.UnmarshalJSON`.
- **You get, for free:** a validated schema (subschemas meta-validated against their dialect metaschema at compile) with typed, `errors.As`-able compile errors (`*SchemaValidationError`, `*LoadURLError`, `*UnsupportedDraftError`, `*MetaSchemaMismatchError`, ...); a reusable `*Schema` (`sch.Validate(inst)`, tree-shaped `*ValidationError` with `Causes`, plus `BasicOutput`/`DetailedOutput`/`FlagOutput` for machine output); and `sch.DraftVersion` as the draft assertion.
- **You do NOT get:** unknown-keyword rejection (vocabulary-driven; the embedded 2020-12 metaschemas set no `unevaluatedProperties: false`) — a strict schema check needs a custom metaschema/vocabulary or your own key check; nor source positions in compile errors.
- **Operational constraints:** compile once, single-threaded (compilation is documented as not thread-safe, `wontfix`); one `Compiler` per schema set (it caches internally and rejects duplicate `AddResource` URLs); validation is plausibly concurrent but undocumented (verify under `-race`).
- **New dependency surface:** `jsonschema/v6` is pinned nowhere in this repo yet, and it pulls `golang.org/x/text/message` (already indirect here as `golang.org/x/text v0.22.0`) — so the marginal module cost is small, but it is a new dependency and a new API-stability bet.

**Opaque-passthrough (hand the schema through unparsed, e.g. as `map[string]any` or raw bytes):**
- Cost is ~0 API: no new dependency, no conversion, no draft semantics to pin, no metaschema fetch at load.
- But it buys **no** validation: no compile-time error for a malformed schema, no metaschema check, no `$ref` resolution, no draft pinning, and every consumer of the schema (or every page-write that consults it) must re-implement or defer the check. The strictness promised by TemplateV3 becomes unenforced and drifts.
- The break-even is decision-theoretic, not API-volume: compile-at-load costs 3 calls + 1 conversion **once per process**, versus permanent deferred enforcement. For a CLI whose `akb template write` already refuses invalid templates with all-errors aggregation, compile-at-load is the closer analogue of existing behaviour; opaque-passthrough only looks cheaper because it moves the work out of the loader entirely.

**Net:** both goccy requirements are satisfiable, with (a) and (b) sharing one mechanism (map-diff over the goccy AST) since goccy has no warn mode and cannot aggregate unknown keys on the strict path. jsonschema v6 is fully capable for (c), with three caveats that must be designed for rather than assumed away: **YAML must be converted to JSON first**, **unknown schema keywords are not rejected**, and **compilation is not thread-safe**.

---

## Missing evidence / could not verify

1. **No code was compiled or run.** No shell tool was available in this run, so nothing here is a runtime observation. The inline-map panic and the concurrency inferences in particular should be settled by a short smoke test before the design depends on them.
2. **`Strict()` + inline map panic at v1.19.2** — the unguarded `NumField` call is present in the pinned source and the fix is unmerged, but I did not execute the reproducer (reported against v1.18.0, and `master` may or may not differ). Untested at exactly v1.19.2.
3. **Which unknown key gets named when several are unknown** — non-determinism follows from the source's `range` over a map; not observed. Treat the message text as stable but the *selection* as unspecified.
4. **Duplicate-key error identity in practice** — both parser (`*yaml.SyntaxError`) and decoder (`*yaml.DuplicateKeyError`) paths are armed by default; which one reaches the caller depends on order, which I inferred from source rather than observed. Code that does `errors.As(err, &yaml.DuplicateKeyError{})` may therefore miss the common case.
5. **Concurrent `Schema.Validate` safety** — no upstream statement found; issue #207 covers compilation only. Inference from source; needs `-race` confirmation.
6. **`Schema.DraftVersion` as the draft-assertion mechanism** — inferred from `Schema.DraftVersion` being exported plus `compileValue` assigning it. Not documented as an assertion API.
7. **Exhaustive metaschema grep** — I read `metaschemas/draft/2020-12/schema`, `meta/core`, `meta/validation`, `meta/unevaluated` and found no blanket closure, and no other directory was inspected. The "unknown keywords are not rejected" conclusion is also independently supported by the vocabulary-driven compile path, but the metaschema check was not exhaustive across every draft/vocabulary.
8. **jsonschema v6 API stability** — v6.0.3 landed ~2 months before v6.0.2 (2025-05-23 → 2026-08-06) after a year of only patch releases, so the surface looks settled; but this repo pins nothing yet. Pin an exact v6.0.3 in `go.mod` rather than floating, since the API changed **incompatibly** from v5 (`AddResource(url, io.Reader)` → `AddResource(url, any)`; `Draft` field → `DefaultDraft()` method; no `CompileString`). Any older tutorial snippet will not compile.
9. **goccy `Parser.parseTokens` ordering** — the claim that the parser's duplicate-key check fires before the decoder's is read from the two code sites, not observed.
10. **Performance** — no benchmarks were run for either library; "cheap/cacheable" for jsonschema is argued from the code structure (cache keyed on `urlPtr`, `init()`-time metaschema compilation) and the maintainer's mutex guidance, not measured.

---

## Sources

**Kept**
- `goccy/go-yaml` `option.go` @ v1.19.2 (https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/option.go) — the authoritative option set at the pinned version; proves `Strict()`≡`DisallowUnknownField()`, proves the absence of any warn/collect option and of a plural `DisallowUnknownFields`.
- `goccy/go-yaml` `decode.go` @ v1.19.2 (https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/decode.go) — unknown-field return-on-first-from-map, `deleteStructKeys` inline-map defect, duplicate-key decoder check, `DecodeFromNode`.
- `goccy/go-yaml` `internal/errors/error.go` + `error.go` @ v1.19.2 (https://raw.githubusercontent.com/goccy/go-yaml/v1.19.2/internal/errors/error.go) — exact error structs/aliases (`UnknownFieldError`, `DuplicateKeyError`, `Error` interface) and the `[line:col]` message format.
- `goccy/go-yaml` `parser/parser.go`, `token/token.go`, `ast/ast.go`, `yaml.go` @ v1.19.2 — `ParseBytes`/`ParseFile`, `token.Position`, `ast.Node`/`File`/`MappingNode`/`MappingValueNode`, `FormatError`, `NodeToValue`, `NodeUnmarshaler`.
- pkg.go.dev index for `goccy/go-yaml@v1.19.2` (https://pkg.go.dev/github.com/goccy/go-yaml@v1.19.2) — closes the question of what is exported (no `yaml.Node`; full `DecodeOption` list; official unknown-field error rendering).
- GitHub release API for goccy (https://api.github.com/repos/goccy/go-yaml/releases/tags/v1.19.2) and issue #810 / PR #849 (https://github.com/goccy/go-yaml/issues/810, https://api.github.com/repos/goccy/go-yaml/pulls/849) — date evidence and the unmerged inline-map fix.
- `santhosh-tekuri/jsonschema` `compiler.go`, `loader.go`, `validator.go`, `output.go`, `schema.go`, `draft.go`, `roots.go`, `vocab.go`, `position.go`, `example_test.go` @ v6.0.3 — the complete compile/validate/error surface and the JSON-only `UnmarshalJSON`.
- pkg.go.dev index for `jsonschema/v6@v6.0.3` (https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6@v6.0.3) — confirms exported types (`SchemaValidationError`, `ValidationError`, `OutputUnit`, `AssertVocabs`, ...) and that no errors.go/validation.go exists.
- GitHub releases API for jsonschema (https://api.github.com/repos/santhosh-tekuri/jsonschema/releases?per_page=12) — v6.0.3 latest, dated.
- Embedded metaschemas at v6.0.3 (`metaschemas/draft/2020-12/schema`, `.../meta/core`, `.../meta/validation`, `.../meta/unevaluated`) — evidence that unknown keywords are not closed off by meta-validation.
- Issue #207 (https://github.com/santhosh-tekuri/jsonschema/issues/207) — maintainer: "Compilation is not threadsafe", closed `wontfix`.
- Repo files read for pinning/context: `go.mod`, `go.sum`, `internal/template/template.go`, `AGENTS.md`.

**Rejected / deprioritized**
- `goccy/go-yaml` `master` docs and the raw pkg.go.dev "latest" page served before I pinned the tag — superseded by the v1.19.2 tag reads; master code (e.g. `AllowFieldPrefixes` context) can drift from the pinned version.
- Search snippets presenting `jsonschema` v5 API (`AddResource(url, io.Reader)`, `CompileString`, `Draft *Draft` field, `SchemaError`) — wrong major version; would not compile against v6.
- `newreleases.io` goccy v1.14.0 page and PR #531 / #609 chatter — retained only as history of the duplicate-key default and the *removed* `DisallowDuplicateKey` option; not authoritative for v1.19.2.
- `json-structure.org` Go SDK "Thread Safety" section — different library; its concurrency statement does **not** apply to `santhosh-tekuri/jsonschema`.
- The `jsonschema` README's v6.0.2 title and the general "errors: introspectable/hierarchy" feature checklist — marketing-level and stale-titled; replaced by direct source reads.
- `position.go` at face value — misleadingly named; it is the vocabulary `SchemaPath` abstraction, not source positions. Kept only to record that compile errors carry no positions.

## Next steps

1. **Smoke-test the three inferences in one small Go test** (highest value, lowest cost): (i) `yaml.Strict()` against a struct with an `,inline` map field at v1.19.2 to confirm/refute the panic; (ii) two unknown top-level keys, run the decode several times, to confirm the named key is non-deterministic and that only one is reported; (iii) duplicate key → `errors.As` for both `*yaml.SyntaxError` and `*yaml.DuplicateKeyError` to see which arrives.
2. **Prototype the map-diff + `ast` walk** (parse → `MappingNode.Values` → `Key.GetToken().Position`) as a single helper that can emit either the aggregated refusal list or the warning list, and confirm it yields line:column for every offending top-level key.
3. **Confirm jsonschema validation concurrency** under `-race`, and settle the YAML→JSON conversion route (`yaml.YAMLToJSON` vs decode-and-re-marshal) including how YAML-specific shapes (non-string keys, aliases, `!!timestamp`) survive the conversion.
4. **Decide the unknown-schema-keyword policy explicitly** — since v6 will not reject them, choose between a custom metaschema with `unevaluatedProperties: false` (plus `AssertVocabs`) and an independent top-level key check; this is a design decision, not a library capability.
5. **Pin `github.com/santhosh-tekuri/jsonschema/v6 v6.0.3`** once adopted, and record that v5 API examples are incompatible.

---

## Supervisor coordination

None needed — no blocking decision arose, no supervisor input was required, and no assumptions were substituted for a designer verdict. Research was bounded to the two assigned libraries and completed from primary sources at the exact pinned/tagged versions.

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Report written to .pi/research/rfc-002-s6/gapfill-libs.md with two library sections (goccy/go-yaml v1.19.2, jsonschema v6.0.3), each carrying verified facts with source URL anchors, exact Go signatures, version/date evidence, and a CONFIDENCE line, plus a three-part FEASIBILITY VERDICT (strict-refuse, warn-only-at-load, compile-at-load vs passthrough) and explicit residual risks (inline-map panic with unmerged fix PR #849, non-deterministic unknown-key naming, undocumented Validate concurrency, jsonschema absent from go.mod, YAML requires conversion, unknown schema keywords not rejected)."
    }
  ],
  "changedFiles": [
    ".pi/research/rfc-002-s6/gapfill-libs.md"
  ],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "read go.mod / go.sum / internal/template/template.go (repo pinning + context)",
      "result": "passed",
      "summary": "goccy/go-yaml pinned v1.19.2 (latest release); santhosh-tekuri/jsonschema absent from go.mod and go.sum, so a new dependency"
    },
    {
      "command": "fetch tagged sources: goccy v1.19.2 option.go/decode.go/struct.go/error.go/internal-errors/parser.go/token.go/ast.go/yaml.go/path.go + pkg.go.dev@v1.19.2 index",
      "result": "passed",
      "summary": "confirmed Strict()==DisallowUnknownField(), no warn/collect option, *yaml.UnknownFieldError with key-naming [line:col] message, no yaml.Node type, ast/token position APIs, duplicate-key default-on"
    },
    {
      "command": "fetch tagged sources: jsonschema v6.0.3 compiler.go/loader.go/validator.go/output.go/schema.go/draft.go/roots.go/vocab.go/position.go/example_test.go/README + pkg.go.dev@v6.0.3 index",
      "result": "passed",
      "summary": "confirmed UnmarshalJSON is JSON-only, AddResource(url, any)+Compile, draftLatest=Draft2020, SchemaValidationError/ValidationError tree shapes, no force-a-draft option (Schema.DraftVersion is the assertion hook)"
    },
    {
      "command": "fetch release APIs: goccy v1.19.2 (2026-01-08) and jsonschema v6.0.3 (2026-08-06); PR #849 state; issue #207 maintainer comment; embedded 2020-12 metaschemas",
      "result": "passed",
      "summary": "date evidence gathered; PR #849 unmerged; 'Compilation is not threadsafe' (wontfix); metaschemas set no unevaluatedProperties:false so unknown schema keywords are not rejected"
    },
    {
      "command": "go build / go test / runtime reproducer",
      "result": "not-run",
      "summary": "no shell or execution tool available in this run; all findings are documentation/source-level and the three runtime inferences are flagged for smoke testing"
    }
  ],
  "validationOutput": [
    "goccy v1.19.2 source read at exact tag: option.go shows Strict() and DisallowUnknownField() both setting d.disallowUnknownField=true; the full DecodeOption set contains no warning/collect mode.",
    "goccy decode.go v1.19.2: `for key, node := range unknownFields { ... return errors.ErrUnknownField(fmt.Sprintf(`unknown field \"%s\"`, key), node.GetToken()) }` -> first-match, map-order, non-deterministic key selection.",
    "goccy deleteStructKeys v1.19.2 has no map-kind guard before structFieldMap/NumField; issue #810 + PR #849 (open, merged:false) -> inline-map panic risk live at the pinned version.",
    "pkg.go.dev@v1.19.2 documents the rendered error: [3:1] unknown field \"unknown\" with caret-annotated source; and lists no `type Node` in the yaml package.",
    "jsonschema v6.0.3 loader.go: UnmarshalJSON uses encoding/json with UseNumber -> no YAML entry point; conversion required.",
    "jsonschema v6.0.3 draft.go: `draftLatest = Draft2020` -> 2020-12 is the default when $schema is missing.",
    "jsonschema v6.0.3 validator.go/output.go: ValidationError{SchemaURL, InstanceLocation, ErrorKind, Causes} tree plus BasicOutput/DetailedOutput/FlagOutput.",
    "jsonschema v6.0.3 issue #207: maintainer 'Compilation is not threadsafe. So you need to use mutex' (closed wontfix).",
    "jsonschema v6.0.3 embedded metaschemas (schema, meta/core, meta/validation, meta/unevaluated): no unevaluatedProperties:false -> unknown schema keywords pass compilation."
  ],
  "residualRisks": [
    "No code was compiled or executed (no shell tool in this run): every finding is documentation/source-level; the inline-map panic, unknown-key non-determinism, and concurrency claims are unconfirmed at runtime.",
    "goccy Strict() + an ',inline' map field is expected to panic (reflect: NumField of non-struct type) at v1.19.2; upstream fix PR #849 is still open and unmerged, so no release fixes it. TemplateV3 structs must avoid inline map fields or the strict path must avoid calling Strict() on them.",
    "goccy Strict() reports only ONE unknown key per struct level and which one is unspecified when multiple are unknown - it cannot satisfy akb's 'ALL failed rules reported' convention; the map-diff path must own aggregation.",
    "Duplicate-key errors may surface as *yaml.SyntaxError (parser, armed first) rather than *yaml.DuplicateKeyError, so errors.As on DuplicateKeyError may miss the common case (inferred, not observed).",
    "jsonschema ValidationError concurrency: Compile is documented as NOT thread-safe (wontfix); Validate concurrency is a researcher inference from source with no upstream guarantee - needs -race verification.",
    "santhosh-tekuri/jsonschema/v6 is not pinned in go.mod/go.sum; the v6 API is incompatible with v5 (AddResource(url, any) vs (url, io.Reader), DefaultDraft() vs Draft field, no CompileString), so unversioned snippets or a floating pin are hazardous.",
    "jsonschema does NOT reject unknown schema keywords (vocabulary-driven compile; embedded 2020-12 metaschemas set no unevaluatedProperties:false) - strictness needs a custom metaschema/vocabulary or an independent key check.",
    "jsonschema has no YAML input path, so a YAML->JSON conversion step is required; YAML-specific shapes (non-string keys, anchors/aliases, tags) may not survive it - unverified.",
    "No benchmarks were run; 'compile is cheap/cacheable' is argued from code structure and maintainer guidance, not measured.",
    "Metaschema grep was not exhaustive across every draft/vocabulary directory (read schema + meta/core + meta/validation + meta/unevaluated only)."
  ],
  "noStagedFiles": true,
  "diffSummary": "Added one research file, .pi/research/rfc-002-s6/gapfill-libs.md, containing the library-feasibility report (verified facts with source anchors, exact Go API signatures, version/date evidence, CONFIDENCE lines, feasibility verdict, missing evidence, sources, next steps, and the acceptance report). The repository's Go sources, go.mod, and go.sum were read only and not modified.",
  "reviewFindings": [
    "no blockers: report is source-backed at exact tags and distinguishes direct evidence from interpretation throughout",
    "risk: .pi/research/rfc-002-s6/gapfill-libs.md relies on three unexecuted inferences (inline-map panic, unknown-key selection order, Validate concurrency) that should be settled by a smoke test before the spec depends on them",
    "risk: the strict-refuse design cannot be `yaml.Strict()` alone if akb is to report all offending keys deterministically; the report recommends the map-diff + ast walk as the single shared mechanism for both the refuse and warn paths"
  ],
  "manualNotes": "Key corrections to training-data assumptions, all verified at the pinned/tagged versions: (1) goccy has NO plural DisallowUnknownFields and NO Decoder method - only the two identical DecodeOptions Strict() and DisallowUnknownField(), and v1.19.2 is itself the latest release, so no stale-docs caveat applies; (2) goccy cannot warn-and-continue on unknown keys at all, and its strict path names only one non-deterministic key, so the map-diff fallback is not merely a fallback but the recommended primary mechanism - goccy's ast package does carry key-level token.Position, and there is no yaml.Node type (nodes live in ast); (3) santhosh-tekuri/jsonschema v6 latest is v6.0.3 (published 2026-08-06), not v6.0.2, and it is NOT currently in go.mod/go.sum; (4) v6's UnmarshalJSON is JSON-only, so YAML needs an explicit conversion step; (5) 2020-12 IS the default draft (draftLatest = Draft2020) but DefaultDraft only sets the fallback - asserting a draft requires checking the exported Schema.DraftVersion after Compile; (6) v6 does not reject unknown schema keywords despite meta-validating subschemas; (7) compilation is explicitly not thread-safe per the maintainer (issue #207, wontfix), while Validate concurrency is undocumented. Also flagged: the jsonschema README at the v6.0.3 tag still self-titles as 'jsonschema v6.0.2', which is a stale title rather than tree evidence, and the repo has no errors.go/validation.go (compile errors live in roots.go/loader.go, validation errors in validator.go/output.go)."
}
```
