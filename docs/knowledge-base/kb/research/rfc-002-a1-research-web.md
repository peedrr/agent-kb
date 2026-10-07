---
grammar: 1
type: research
title: Go JSON Schema 2020-12 Validators and RFC 3339 Parsing Facts (RFC-002-A1)
summary: "Sourced ecosystem survey of Go JSON Schema 2020-12 validators and RFC 3339 parsing behavior, closing the format-assertion questions behind ADR-003."
status: final
provenance: agent-drafted
created: "2026-10-06"
updated: "2026-10-06"
as_of: "2026-10-06"
run: rfc-002-a1-validator
informs:
- ADR-003
scope:
- "internal/cel/**"
- "internal/template/**"
tags:
- json-schema
- validation
- ecosystem-survey
- rfc-002
---
# Research: Go JSON Schema 2020-12 validators + RFC 3339 parsing facts (for RFC-002-A1 ADR)

Retrieved: live web fetches of primary sources (GitHub raw source, GitHub API, json-schema.org, rfc-editor.org, go.dev/src, pkg.go.dev).
Evidence freshness: repo metadata and releases indexed up to **2026-09-30** (santhosh-tekuri/jsonschema `updated_at`), cel-go **v0.32.0** (2026-08-19).
Validation limitation: no `source_check` tool was available in this run's toolset. Instead every decision-critical claim below is backed by a
fetched primary source with an exact quote and a URL; nothing is taken from search-result prose alone except where explicitly labelled
"interpretation".

---

## Q1. santhosh-tekuri/jsonschema — version, maintenance, module path, license

| Fact | Value | Source |
|---|---|---|
| Module path | `github.com/santhosh-tekuri/jsonschema/v6` | https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6 |
| Latest tagged release | **v6.0.3**, published 2026-06-28T17:38:00Z | https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6 (version list: v6.0.0 2024-06-04, v6.0.1 2024-06-06, v6.0.2 2025-05-23, v6.0.3 2026-06-28) |
| Newest published module version | pseudo-version `v6.0.4-0.20260806092148-ec6106e5f0a3` (untagged, 2026-08-06) → v6 line is still iterating | https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6@v6.0.4-0.20260806092148-ec6106e5f0a3 |
| Go version / license | Go **1.21**; **Apache-2.0** | same pkg.go.dev page; GitHub API `license.spdx_id = "Apache-2.0"` |
| Repo activity | `pushed_at` **2026-09-21**, `updated_at` **2026-09-30**, `archived: false`, default branch **`boon`** | https://api.github.com/repos/santhosh-tekuri/jsonschema |
| Backlog / adoption | `open_issues_count` **44** (GitHub's count includes PRs), 1278 stars, 140 forks, Discussions enabled | same API response |
| Release delta | 16 commits / 20 files between v6.0.2 → v6.0.3 | https://github.com/santhosh-tekuri/jsonschema/compare/v6.0.2...v6.0.3 |
| CLI is versioned separately | README says `cmd/jv/v0.7.0`; a third-party release tracker shows `cmd/jv/0.8.0` | https://github.com/santhosh-tekuri/jsonschema/blob/boon/README.md ; https://newreleases.io/project/github/santhosh-tekuri/jsonschema/release/v6.0.3 |

**Verdict (Q1):** maintained, single-maintainer, low-but-steady release cadence, v6 is the current major line; no v7 exists.
Confidence: **high** (primary repo metadata + pkg.go.dev version index).
*Minor uncertainty:* exact current CLI tag (`cmd/jv/v0.7.0` vs `0.8.0`) — the README and the release tracker disagree; CLI version is irrelevant to library use.

---

## Q2. 2020-12 coverage of santhosh v6

- **Claim:** passes the JSON-Schema-Test-Suite (except optional tests), with a **100% bowtie compliance badge for draft 2020-12**.
  - README: "pass [JSON-Schema-Test-Suite](...) excluding optional(compare with other impls at bowtie)" with badges for draft-04/06/07/2019-09/2020-12 — https://github.com/santhosh-tekuri/jsonschema/blob/boon/README.md
  - Badge endpoint payload: `{"schemaVersion": 1, "label": "Draft 2020-12", "message": "100% Passing", "color": "006400"}` — https://bowtie.report/badges/go-jsonschema/compliance/draft2020-12.json
  - **Support:** direct evidence (badge payload fetched). **Confidence: high** for overall 2020-12 keyword coverage (the bowtie suite covers `contains`/`minContains`, `if`/`then`/`else`, `$ref`/`$defs` (incl. recursive/cyclic), `dependentRequired`/`dependentSchemas`, `unevaluatedProperties`/`unevaluatedItems` — a 100% score cannot be achieved without them).
- **Claim:** recursive/cyclic `$ref` handling is explicit, not incidental.
  - README feature list: "detect infinite loop traps — `$schema` cycle, **validation cycle**" (README, above).
  - `validator.go` contains a scope-cycle check: `if scp := vd.scp.checkCycle(); scp != nil { return nil, vd.error(&kind.RefCycle{...}) }` — https://github.com/santhosh-tekuri/jsonschema/blob/v6.0.2/validator.go
  - **Support:** direct evidence (source quote). **Confidence: high.**
- **Other relevant 2020-12 capabilities advertised:** vocabulary-based validation, custom `$schema` URL, **custom regex engine**, format assertions, content assertions, custom vocabularies, mixed-dialect support (README, above).
- **Caveat / gap:** the README does **not** enumerate individual keywords, and I did not read every keyword code path (e.g. `dependentRequired`, `unevaluatedItems`) in source. My assurance for those specific keywords rests on the 100% bowtie 2020-12 score. Confidence for keyword-by-keyword completeness: **medium-high (interpretation on top of direct badge evidence)**.
- **Minor deviation observed in source:** an unknown `format` name resolves to `nil` and is silently unchecked even when `AssertFormat()` is on (`objcompiler.go`, quote in Q3b), whereas 2020-12 §7.2.3 says an implementation using the Format-Assertion vocabulary "MUST fail upon encountering unknown formats." Not tested end-to-end; flag as a low-confidence nit, not a blocker for akb (akb controls its own format names).

---

## Q3. CRITICAL — santhosh v6 format-assertion API

### (a) How assertions are enabled
`compiler.go` (branch `boon`) — https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/boon/compiler.go

```go
// AssertFormat always enables format assertions.
//
// Default Behavior:
// for draft-07: enabled.
// for draft/2019-09: disabled unless metaschema says `format` vocabulary is required.
// for draft/2020-12: disabled unless metaschema says `format-assertion` vocabulary is required.
func (c *Compiler) AssertFormat() {
	c.assertFormat = true
}
```

Activation predicate in `objcompiler.go` — https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/boon/objcompiler.go

```go
func (c *objCompiler) assertFormat(draftVersion int) bool {
	if c.c.assertFormat || draftVersion < 2019 {
		return true
	}
	if draftVersion == 2019 {
		return c.hasVocab("format")
	} else {
		return c.hasVocab("format-assertion")
	}
}
```
**Confidence: high** (two independent files fetched from the default branch).

### (b) Can callers override the built-in `date-time` checker? — **YES**
Signature and guard (`compiler.go`):

```go
// RegisterFormat registers custom format.
//
// NOTE:
//   - "regex" format can not be overridden
//   - format assertions are disabled for draft >= 2019-09
//     see [Compiler.AssertFormat]
func (c *Compiler) RegisterFormat(f *Format) {
	if f.Name != "regex" {
		c.formats[f.Name] = f
	}
}
```
`Format` type (`format.go`): `type Format struct { Name string; Validate func(v any) error }`.

Lookup order at compile time (`objcompiler.go`) — **compiler-registered formats win over built-ins**:

```go
// format --
if c.assertFormat(s.DraftVersion) {
	if f := c.strVal("format"); f != nil {
		if *f == "regex" {
			s.Format = &Format{
				Name:     "regex",
				Validate: c.c.roots.regexpEngine.validate,
			}
		} else {
			s.Format = c.c.formats[*f]
			if s.Format == nil {
				s.Format = formats[*f]
			}
		}
	}
}
```
So `c.RegisterFormat(&jsonschema.Format{Name: "date-time", Validate: goStrictRFC3339Fn})` replaces the built-in
`date-time` for every schema compiled by that `Compiler`. Only `regex` is un-overridable.
Because the compiled `*Schema` caches the resolved `Validate` func, **the override must be registered before `Compile`**
(consistent with the historical v5 report that `Schema` pre-caches the format function — https://github.com/santhosh-tekuri/jsonschema/issues/93 ).
**Confidence: high (source-direct); medium for the "must register before Compile" timing inference** (implied by the compile-time resolution above; not stated in docs).

Note the surrounding behaviour: with assertions **off**, `format` is not even resolved, which is why users see `Schema.Format == nil` (see discussion https://github.com/santhosh-tekuri/jsonschema/discussions/172 ).

### (c) What the BUILT-IN `date-time` checker accepts — exact source
File: `format.go`, branch `boon` (current v6 line) — https://raw.githubusercontent.com/santhosh-tekuri/jsonschema/boon/format.go
(byte-identical to the vendored copy in skopeo 1.23.0 — https://fossies.org/linux/misc/skopeo-1.23.0.tar.gz/skopeo-1.23.0/vendor/github.com/santhosh-tekuri/jsonschema/v6/format.go )

```go
// see https://datatracker.ietf.org/doc/html/rfc3339#section-5.6
// NOTE: golang time package does not support leap seconds.
func validateTime(v any) error {
	...
	h, m, s := hms[0], hms[1], hms[2]
	if h > 23 || m > 59 || s > 60 {
		return LocalizableError("hour/min/sec out of range")
	}
	...
	if str != "z" && str != "Z" {
		// parse time-numoffset
		if len(str) != 6 {
			return LocalizableError("offset must be 6 characters long")
		}
	...
	// check leap second
	if s >= 60 && (h != 23 || m != 59) {
		return LocalizableError("invalid leap second")
	}

	return nil
}

// see https://datatracker.ietf.org/doc/html/rfc3339#section-5.6
func validateDateTime(v any) error {
	s, ok := v.(string)
	if !ok {
		return nil
	}

	// min: yyyy-mm-ddThh:mm:ssZ
	if len(s) < 20 {
		return LocalizableError("less than 20 characters long")
	}

	if s[10] != 't' && s[10] != 'T' {
		return LocalizableError("11th character must be t or T")
	}
	if err := validateDate(s[:10]); err != nil {
		return fmt.Errorf("invalid date element: %v", err)
	}
	if err := validateTime(s[11:]); err != nil {
		return fmt.Errorf("invalid time element: %v", err)
	}
	return nil
}
```

**Direct answers (confidence: high, source-direct):**
- **Lowercase `t`: ACCEPTED** — `if s[10] != 't' && s[10] != 'T'` (explicitly allows `t`).
- **Lowercase `z`: ACCEPTED** — `if str != "z" && str != "Z"` (explicitly allows `z` for the zone).
- **Leap second `:60`: ACCEPTED** when hour/minute are `23:59` (`if s >= 60 && (h != 23 || m != 59) { error }` → `23:59:60` passes).
- **Numeric offset: `+hh:mm` only**, exactly 6 chars (`offset must be 6 characters long`); `+0330` (no colon) rejected.
- Date element is delegated to `time.Parse("2006-01-02", s)` (`validateDate`), i.e. calendar validation via Go.

⇒ The built-in checker is **looser than the Go-strict RFC 3339 the ADR wants** on all three axes: it accepts lowercase `t`, lowercase `z`, and leap seconds. Overriding it (Q3b) is therefore necessary *and* possible. (Also note it is *stricter* than Go in one way: it requires the colon in the numeric offset.)

### (d) Default behaviour: assertion or annotation?
**Annotation-only / not evaluated by default for draft 2020-12** unless (i) the metaschema requires the `format-assertion` vocabulary, or (ii) the caller calls `Compiler.AssertFormat()` (quotes in Q3a).
**Confidence: high.**

---

## Q4. Alternatives (2026 state)

| Library | Draft 2020-12 | Maintenance | Format assertion | Verdict |
|---|---|---|---|---|
| **santhosh-tekuri/jsonschema/v6** | Yes, 100% bowtie 2020-12 | Active (pushed 2026-09-21) | `AssertFormat()` + `RegisterFormat` **override**; Go stdlib RE2 by default, swappable engine | **Best fit** |
| **qri-io/jsonschema** | **No** — "golang implementation of https://json-schema.org drafts 7 & 2019-09" | Dead-ish: latest release **v0.2.1 (Mar 29 2021)**; deps.dev: "0 commit(s) and 0 issue activity found in the last 90 days — score normalized to 0" | Custom keyword/format machinery exists, but wrong draft | Reject |
| **xeipuuv/gojsonschema** | **No** — "An implementation of JSON Schema, draft v4 v6 & v7"; last master commit **Oct 2020**; maintainer-status issue #344 still open (updated 2025-06-04) | Unmaintained | best-effort format.json (32 draft-7 format failures in vearutop benchmark) | Reject |
| **invopop/jsonschema** | Emits 2020-12 **schemas only** — "This package can be used to generate JSON Schemas from Go types through reflection"; Google: "invopop/jsonschema provides inference, but not validation." | Active | n/a (no validation) | Generation only |
| **google/jsonschema-go** (announced Jan 2026, MIT, stdlib-only) | Yes (2020-12 + draft-07 only) | Google-backed, new | **Impossible**: "The 'format' keyword ... is recorded in the Schema, but is ignored during validation. It does not even produce annotations." | Reject for this use case |
| **kaptinlin/jsonschema** (MIT, ~226 stars, v0.8.0 2026-06-05, last push 2026-06-08) | Yes: 2020-12/2019-09/7/6/4, `$schema`-driven, default 2020-12 | Active | `SetAssertFormat(true)` + `RegisterFormat("customer-id", func(v any) bool, "string")`; annotations by default | **Credible alternative, unverified** |
| others seen and deprioritised: `seeadoog/jsonschema` (zero-alloc claim, draft-07 era, self-reported benchmarks), `open-circle/schema-benchmarks` (TypeScript-focused), `ad3n/jsonschema` (fork mirror of kaptinlin docs) | — | — | — | Not evaluated |

Sources: https://github.com/qri-io/jsonschema · https://deps.dev/project/github/qri-io%2fjsonschema · https://github.com/xeipuuv/gojsonschema/issues/344 · https://raw.githubusercontent.com/invopop/jsonschema/main/README.md · https://opensource.googleblog.com/2026/01/a-json-schema-package-for-go.html · https://pkg.go.dev/github.com/google/jsonschema-go/jsonschema · https://raw.githubusercontent.com/kaptinlin/jsonschema/main/README.md

**Direct answer to "does anything beat santhosh v6 for this use case (embedded validator in a Go CLI, full 2020-12, overridable format checkers, RE2 patterns)?"**
**No verified library beats it.** The two capability combinations akb needs — (1) 2020-12 with suite-verified completeness and (2) a
*per-format override hook for a built-in like `date-time`* — are satisfied only by santhosh v6 as far as I could verify:
google/jsonschema-go explicitly ignores `format`; qri/xeipuuv do not do 2020-12; invopop does not validate.
`kaptinlin/jsonschema` is the only real challenger but I found **no bowtie/JSON-Schema-Test-Suite compliance evidence** for it and no
documentation that `RegisterFormat` can override a *built-in* name (it is only shown with a custom name).
**Confidence: high for "no verified better option"; medium for the claim that kaptinlin could not do it (unverified, not disproven).**

RE2 detail (santhosh): `compiler.go` — "UseRegexpEngine changes the regexp-engine used. By default it uses regexp package from go standard library." and `func goRegexpCompile(s string) (Regexp, error) { return regexp.Compile(s) }` → Go's stdlib `regexp` = RE2 semantics (no backreferences/lookaround). Confidence: high.

---

## Q5. Performance and compile-once/validate-many

- **Compilation caching:** v6 separates `Compiler.Compile` (expensive, once) from `(*Schema).Validate` (per instance). Maintainer-measured compile for a small config schema on v6: **≈2.8 ms** vs **≈37 ms** on v5, with the explicit caveat that v6 lets you keep JSON parsing out of compile time.
  - https://github.com/santhosh-tekuri/jsonschema/issues/185 and https://github.com/santhosh-tekuri/jsonschema/discussions/189
  - **Support:** maintainer statements + user benchmark code (`BenchmarkCompile-10 396 2857313 ns/op`). Confidence: **medium** (3rd-party hardware, no benchstat).
- **Thread-safety caveat (important for a CLI/daemon):** "Compilation is not threadsafe. So you need to use mutex" (maintainer, issue #207, in response to a `concurrent map read and map write` panic during concurrent `Compile`). Whether a *single compiled* `*Schema` is safe for concurrent `Validate` is **not documented** anywhere I found — treat as unverified.
  - https://github.com/santhosh-tekuri/jsonschema/issues/207
- **Comparative benchmarks:** the credible public ones are **stale for 2020-12**:
  - vearutop (draft-07 + ajv suites): santhosh ≈2× faster than xeipuuv on complex schemas; draft-7 geo-mean 2.66 µs (santhosh) vs 4.03 µs (xeipuuv) vs 4.92 µs (qri). https://dev.to/vearutop/benchmarking-correctness-and-performance-of-go-json-schema-validators-3247
  - ucarion (draft-07, "realistic" schema): santhosh and xeipuuv ~30% slower than `json-schema-go` but ~50% more memory-efficient. https://github.com/ucarion/json-schema-go-benchmark
  - **Gap:** no credible 2025–2026 head-to-head benchmark of 2020-12 Go validators (santhosh vs kaptinlin vs google/jsonschema-go) was found.

---

## Q6. RFC 3339 facts

### (a) What RFC 3339 permits — exact text
Source: https://www.rfc-editor.org/rfc/rfc3339.html (§5.6, §5.7, Appendix D; same text at https://datatracker.ietf.org/doc/html/rfc3339)

```
      date-fullyear   = 4DIGIT
      date-month      = 2DIGIT  ; 01-12
      date-mday       = 2DIGIT  ; 01-28, 01-29, 01-30, 01-31 based on
                                ; month/year
      time-hour       = 2DIGIT  ; 00-23
      time-minute     = 2DIGIT  ; 00-59
      time-second     = 2DIGIT  ; 00-58, 00-59, 00-60 based on leap second
                                ; rules
      time-secfrac    = "." 1*DIGIT
      time-numoffset  = ("+" / "-") time-hour ":" time-minute
      time-offset     = "Z" / time-numoffset

      partial-time    = time-hour ":" time-minute ":" time-second [time-secfrac]
      full-date       = date-fullyear "-" date-month "-" date-mday
      full-time       = partial-time time-offset

      date-time       = full-date "T" full-time

      NOTE: Per [ABNF] and ISO8601, the "T" and "Z" characters in this
      syntax may alternatively be lower case "t" or "z" respectively.

      This date/time format may be used in some environments or contexts
      that distinguish between the upper- and lower-case letters 'A'-'Z'
      and 'a'-'z' (e.g. XML).  Specifications that use this format in
      such environments MAY further limit the date/time syntax so that
      the letters 'T' and 'Z' used in the date/time syntax must always
      be upper case.  Applications that generate this format SHOULD use
      upper case letters.
```
and §5.7:
```
   The grammar element time-second may have the value "60" at the end of
   months in which a leap second occurs -- to date: June (XXXX-06-
   30T23:59:60Z) or December (XXXX-12-31T23:59:60Z); see Appendix D ...
   It is also possible for a leap second to be subtracted, at which times
   the maximum value of time-second is "58". At all other times the
   maximum value of time-second is "59".
```
- **Confirmed:** lowercase `t`/`z` are permitted *by the ABNF* (the RFC says so explicitly, via ABNF's case-insensitive terminals), **and** the RFC explicitly permits a consuming spec to *further limit* to uppercase. Confidence: **high** (primary RFC text).
- **Confirmed:** `time-second = 00-60` is permitted at leap-second moments (June/December endpoints). Confidence: **high**.
- *Not separately fetched:* RFC 5234 §2.3 (ABNF literal case-insensitivity rule itself). The RFC 3339 NOTE already states the consequence, so this is not decision-relevant.

### (b) Go `time.RFC3339` / `time.Parse`
Layout constants and the documented caveat (source: https://go.dev/src/time/format.go , https://pkg.go.dev/time ):
```go
RFC3339     = "2006-01-02T15:04:05Z07:00"
RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
```
```go
// [RFC3339], [RFC822], [RFC822Z], [RFC1123], and [RFC1123Z] are useful for formatting;
// when used with time.Parse they do not accept all the time formats
// permitted by the RFCs and they do accept time formats not formally defined.
```
```go
// Timestamps representing leap seconds (second 60) cannot be parsed.
// These are not representable by [Time].
```
- **Leap seconds: REJECTED.** `time.Parse(time.RFC3339, "1990-12-31T23:59:60Z")` → `parsing time "1990-12-31T23:59:60Z": second out of range`.
  Sources: Go source doc above; https://github.com/golang/go/issues/77072 (closed) ; https://github.com/golang/go/issues/50888 (closed) ; https://github.com/golang/go/issues/77163 (2026-01-13, closed as duplicate, with the playground repro). Confidence: **high**.
- **Lowercase `z`: REJECTED.** Source (go.dev/src/time/format.go):
```go
case stdISO8601TZ, stdISO8601ShortTZ, stdISO8601ColonTZ, stdISO8601SecondsTZ, stdISO8601ColonSecondsTZ:
	if len(value) >= 1 && value[0] == 'Z' {
		value = value[1:]
		z = UTC
		break
	}
	fallthrough
case stdNumTZ, ...:
```
  (uppercase-only test; the `fallthrough` then demands a `+`/`-` offset). Corroborated by Go's own commit message: "include RFC3339 in the list of layouts that do not accept all the time formats allowed by RFCs (**lowercase z**)" — https://avyl.ru/go/go/commit/3e887ff7ea4f1e0d17a7a67e906bef9eec00ed1d and the closed issue https://github.com/golang/go/issues/20869 ("The format is case sensitive", closed as duplicate, response: "Given that generators should generate T and Z, not t and z, I'm not particularly inclined to try to fix this."). Confidence: **high**.
- **Lowercase `t`: REJECTED (high-ish, small residual doubt).** Evidence for rejection: issue #20869 explicitly lists case-sensitivity for **both** `t` and `z`; the Go doc says Parse-with-RFC3339 does not accept all RFC-permitted forms; there is no known CL making the `T` separator case-insensitive. Counter-evidence: Go's parser matches *literal* layout chunks with a case-insensitive helper — `// match reports whether s1 and s2 match ignoring case` (`func match(s1, s2 string) bool`, go.dev/src/time/format.go) — which nominally would let a literal `T` match `t`. I could not reach the literal-prefix call site in the fetched slice to settle it. **Confidence: medium.** See "Missing evidence".
- Bonus Go laxness (relevant to akb's strictness goal), direct source comment + issue: single-digit hours are accepted (`0000-01-01T0:00:00Z`), comma fractional separators are accepted, and zone offsets are range-tested with `>` rather than `>=`:
```go
			// The range test use > rather than >=,
			// as some people do write offsets of 24 hours
			// or 60 minutes or 60 seconds.
			if hr > 24 {
				rangeErrString = "time zone offset hour"
			}
```
  Sources: go.dev/src/time/format.go (above); https://github.com/golang/go/issues/54580 (enumerates the three laxnesses and states "This parses all valid RFC 3339 timestamps (except those with leap-seconds)"). Confidence: **high**.

### (c) cel-go `timestamp()`
- Module/version now: **`cel.dev/cel-go` v0.32.x** (v0.32.0 released 2026-08-19; repo moved to `github.com/cel-expr/cel-go` on 2026-06-16). https://pkg.go.dev/cel.dev/cel-go · https://github.com/cel-expr/cel-go/releases/tag/v0.32.0
- **Parsing implementation:** string→timestamp conversions call **`time.Parse(time.RFC3339, s)`**, gated by a strict pre-check `isStrictRFC3339` added by PR #1338, which first shipped in **v0.30.0 (2026-07-26)**. Before v0.30.0 the conversion was raw `time.Parse` (issue #1108: comma fractions, single-digit hours, offsets beyond 23:59 were all silently accepted). PR body: "Cause: the `string`-to-timestamp conversion parses with `time.Parse(time.RFC3339, ...)` ... Fix: gate the conversion on a strict RFC 3339 pattern in the callee before `time.Parse` runs, so the listed forms report the usual conversion error while `time.Parse` keeps doing the calendar validation."
  Sources: https://github.com/cel-expr/cel-go/pull/1338 · https://github.com/cel-expr/cel-go/releases/tag/v0.30.0 · https://github.com/cel-expr/cel-go/commit/41d9149fcbe0c22a91683c10ceed46f8434678ca
- **Exact current source** (`common/types/timestamp.go`, master = cel-expr/cel-go, fetched via https://raw.githubusercontent.com/cel-expr/cel-go/master/common/types/timestamp.go):
```go
// strictRFC3339Pattern gates the strings accepted by the `timestamp()` overload.
// time.Parse accepts inputs that RFC 3339 forbids: a ',' fractional-second
// separator, single-digit time fields, and numeric offsets whose hours exceed
// 23 or minutes exceed 59. ...
var strictRFC3339Pattern = regexp.MustCompile(
	`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])[Tt]([01]\d|2[0-3]):[0-5]\d:([0-5]\d|60)(\.\d+)?([Zz]|[+-]([01]\d|2[0-3]):[0-5]\d)$`)
```
```go
	case string:
		s := strings.TrimSpace(v)
		if s == "" { return time.Time{}, errors.New("invalid RFC 3339 timestamp: ''") }
		if isStrictRFC3339(s) {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid RFC 3339 timestamp %q", s)
			}
			return validateTimestampRange(t.UTC())
		}
```
**Direct answers:**
1. **cel-go `timestamp()` parses with `time.RFC3339`** (Go semantics), after a strict pre-gate. **Confirmed, high confidence.**
2. **cel-go's gate accepts lowercase `t` AND lowercase `z`** (`[Tt]`, `[Zz]`) → cel-go does **not** enforce the uppercase-only variant. (Then `time.Parse` still rejects lowercase `z` per Q6b, so `"…z"` errors anyway; lowercase `t` passes iff Go accepts it.)
3. **Leap seconds:** the pattern allows `:60`, but the subsequent `time.Parse(time.RFC3339, s)` rejects second 60 → `invalid RFC 3339 timestamp`. So cel-go rejects leap seconds, but only via the Go step, not its own gate.
4. Other tightening in v0.31.0: "reject out-of-range hours in timezone offset parsing (#1391)".
Confidence: **high** for 1/2/3/4 as source/quote-level facts; the *runtime* consequence in 3 is a two-step inference from two directly quoted code paths (**medium-high**).

**Implication for RFC-002-A1 (researcher inference, explicitly labelled):** CEL/akb `now` and any CEL-side `timestamp()` string conversion cannot be the place where akb enforces "uppercase T only / uppercase Z only" — cel-go deliberately accepts both cases. Strictness must come from the JSON-Schema format checker akb registers (Q3b), or from akb's own Go code.

---

## Q7. JSON Schema 2020-12 on `format` and the defined formats
Source: https://json-schema.org/draft/2020-12/json-schema-validation

- §7.1 Foreword: "The current URI for this vocabulary, known as the Format-Annotation vocabulary, is: ... Implementing support for this vocabulary is **REQUIRED**." / "The URI for the Format-Assertion vocabulary, is: `https://json-schema.org/draft/2020-12/vocab/format-assertion` ... Implementing support for the Format-Assertion vocabulary is **OPTIONAL**."
- §7.2.1 (default = annotation): "The value of format MUST be collected as an annotation ... **Implementations MAY still treat "format" as an assertion** in addition to an annotation ... The implementation **MUST provide options to enable and disable such evaluation and MUST be disabled by default**."
- §7.2.2 (Format-Assertion vocabulary, if declared `true`): "implementations MUST provide full validation support for all of the formats defined by this specificaion. Implementations that cannot provide full validation support MUST refuse to process the schema."
- §7.2.3: unknown formats must still be collected as annotations; "When the Format-Assertion vocabulary is specified, implementations MUST fail upon encountering unknown formats."
- §7.3.1 Dates/Times/Duration: "Date and time format names are derived from **RFC 3339, section 5.6** ... The duration format is from the ISO 8601 ABNF as given in Appendix A of RFC 3339." then `date-time` ("a valid representation according to the 'date-time' ABNF rule"), `date` (full-date), `time` (full-time), `duration`.
  → **Researcher inference (labelled):** because the spec defers to the RFC 3339 ABNF, and that ABNF's `t`/`z` are case-insensitive per the RFC 3339 NOTE, a *conformant* JSON Schema `date-time` validator is entitled to accept lowercase `t`/`z`. akb's "Go-strict, uppercase-only" rule is therefore a deliberate narrowing (RFC 3339 §5.6 explicitly allows a consuming spec to do this: "MAY further limit the date/time syntax so that the letters 'T' and 'Z' ... must always be upper case").
- Full defined-format list in 2020-12: §7.3.1 date-time, date, time, duration; §7.3.2 email, idn-email; §7.3.3 hostname, idn-hostname; §7.3.4 ipv4, ipv6; §7.3.5 uri, uri-reference, iri, iri-reference, **uuid** ("A string instance is valid against this attribute if it is a valid string representation of a UUID, according to [RFC4122]"); §7.3.6 uri-template; §7.3.7 json-pointer, relative-json-pointer; §7.3.8 regex.
- santhosh v6 built-in format set: json-pointer, relative-json-pointer, uuid, duration, period, ipv4, ipv6, hostname, email, date, time, date-time, uri, iri, uri-reference, iri-reference, uri-template, semver (from `format.go` `var formats = map[string]*Format{...}`). Note: **no `idn-email` / `idn-hostname`** — those names resolve to nil and are silently unchecked.

---

## Contradictions / disputes recorded
1. **Lowercase `t` acceptance in Go's `time.Parse(time.RFC3339, …)`:** issue #20869 says the format is case-sensitive for both `t` and `z`; #54580 says Parse "parses all valid RFC 3339 timestamps (except those with leap-seconds)" (which would include lowercase forms); Go's parser matches literal chunks with the case-insensitive `match` helper, yet the RFC3339 fast path requires `s[10]=='T'`. **Unresolved** (see Missing evidence).
2. **`open_issues_count: 44`** on the GitHub API includes PRs — do not read it as "44 open issues". No bug backlog severity data collected.
3. **CLI version** `cmd/jv/v0.7.0` (README) vs `0.8.0` (newreleases.io). Immaterial to library use.
4. **Unknown-format handling under Format-Assertion:** spec §7.2.3 says implementations MUST fail on unknown formats; santhosh v6's lookup (Q3b) leaves an unknown format as `nil` (silently skipped). Source-derived, not runtime-tested.

## Missing evidence / gaps
- **Not verified by execution:** anything requiring a Go toolchain (this child had no shell/exec tool). Highest-value missing experiments:
  (a) `time.Parse(time.RFC3339, "2023-10-10t09:19:21Z")` and `…"2023-10-10t09:19:21z"` — settles the lowercase-`t` contradiction;
  (b) `c.RegisterFormat(&jsonschema.Format{Name:"date-time", Validate: f})` + `c.AssertFormat()` + `c.Compile(...)` + validate `"2020-01-01t00:00:00Z"` — confirms the override end-to-end;
  (c) concurrent `Validate` on one compiled `*Schema` under `-race`.
- **No bowtie/test-suite compliance evidence found for kaptinlin/jsonschema or google/jsonschema-go**; no verified statement that kaptinlin's `RegisterFormat` can shadow a *built-in* format name.
- **No 2025–2026 comparative benchmark** of 2020-12 Go validators; the shippable numbers (ucarion 2020, vearutop ~2021, santhosh issue #185) are draft-07-era or single-schema.
- **santhosh v6 per-keyword source paths** for `dependentRequired`/`dependentSchemas`/`unevaluatedItems` were not read individually (relied on the 100% bowtie 2020-12 badge).
- **`source_check` unavailable** in this run → no automated claim-vs-source attestation; all critical claims are backed by fetched primary text and exact quotes instead.

## Sources
**Kept (primary / authoritative):**
- santhosh-tekuri/jsonschema — `boon` branch `format.go`, `compiler.go`, `objcompiler.go`, `validator.go`, README — current v6 source for every format/assertion claim (Q2, Q3).
- https://api.github.com/repos/santhosh-tekuri/jsonschema — maintenance/licence/activity metadata (Q1).
- https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6 (+ version index) — tagged releases, Go version, module path (Q1).
- https://bowtie.report/badges/go-jsonschema/compliance/draft2020-12.json — "100% Passing" draft 2020-12 (Q2).
- https://www.rfc-editor.org/rfc/rfc3339.html — ABNF, the `t`/`z` NOTE, leap-second rules (Q6a).
- https://go.dev/src/time/format.go — layout constants, documented Parse caveats, leap-second statement, uppercase-only `Z` branch, `>` range-test comment (Q6b).
- https://github.com/golang/go/issues/50888, /77072, /77163, /20869, /54580 (+ the CL commit message at avyl.ru) — Go behaviour on leap seconds, lowercase, and laxness (Q6b).
- https://raw.githubusercontent.com/cel-expr/cel-go/master/common/types/timestamp.go ; https://github.com/cel-expr/cel-go/pull/1338 ; releases v0.30.0/v0.31.0/v0.32.0 — cel-go timestamp parsing (Q6c).
- https://json-schema.org/draft/2020-12/json-schema-validation — format vocabularies + defined formats (Q7).
- Alternatives: qri README/deps.dev, xeipuuv issue #344, invopop README, google/jsonschema-go pkg docs + Google blog, kaptinlin README (Q4).
- Performance: santhosh issue #185 / discussion #189 / issue #207, vearutop DEV post, ucarion benchmark repo (Q5).

**Rejected / deprioritised:**
- `newreleases.io` CLI-version page — secondary, imprecise (kept only as the CLI discrepancy note).
- `seeadoog/jsonschema` README benchmarks — self-reported, draft-07 era, no methodology.
- `open-circle/schema-benchmarks` — TypeScript/Node oriented, no Go entries relevant to this question.
- `fossies.org` vendored skopeo copy of `v6/format.go` — used only as a byte-identical cross-check of Q3c (not a primary source).
- `github.com/ad3n/jsonschema` pkg.go.dev page — appears to mirror kaptinlin's docs; risk of attribution error, so not used as evidence for kaptinlin.
- DEV.to recap of the Google `jsonschema-go` announcement — secondary; the Google blog post was used instead.

## Next steps (highest value only)
1. Run the three missing Go experiments (a/b/c) in the akb repo — they close the two biggest gaps (lowercase-`t`, override end-to-end) cheaply, and (c) will catch a real concurrency hazard if akb ever validates in parallel.
2. Decide the ADR's normative position explicitly: "akb asserts a *narrowed* RFC 3339 (uppercase `T`/`Z`, numeric offset `±hh:mm`, no leap seconds)" and cite RFC 3339 §5.6's "MAY further limit" sentence plus JSON Schema §7.2.1's "MUST be disabled by default / MAY treat as assertion" as the licence for it.
3. If akb ever considers alternatives, require the candidate to show bowtie/JSON-Schema-Test-Suite 2020-12 results *and* a documented built-in-format override hook before shortlisting kaptinlin/jsonschema.
