---
grammar: 1
type: adr
id: ADR-003
title: JSON Schema Validation Uses santhosh v6 and Asserts Every Declared Format
summary: JSON Schema validation delegates to santhosh-tekuri/jsonschema v6; every declared format asserts; akb-registered date-time/date checkers pin the strict RFC 3339 profile cel-go accepts.
tags: [json-schema, validation, format-assertion, rfc-002]
status: proposed
created: "2026-10-06"
updated: "2026-10-06T15:45:00Z"
provenance: agent-drafted
scope: ["internal/**", "cmd/akb/**", "go.mod"]
revisit:
- kaptinlin/jsonschema or another Go validator publishes draft 2020-12 test-suite or bowtie compliance plus a documented built-in-format override hook
- santhosh-tekuri/jsonschema maintenance lapses (no commits for a year) or the module is abandoned
- cel-go changes timestamp() string acceptance again (re-check at every cel-go upgrade)
- a built-in non-temporal format checker false-rejects values a template author reasonably declared
---

# ADR-003: JSON Schema Validation Uses santhosh v6 and Asserts Every Declared Format

> In the context of RFC-002's open validation engine, facing the choice of a Go JSON Schema 2020-12 validator whose format handling must guarantee CEL's parse preconditions across a pinned cel-go upgrade, we decided that akb validates with `santhosh-tekuri/jsonschema/v6`, asserts every format a template declares, and registers its own strict-RFC 3339 `date-time`/`date` checkers, and neglected qri-io, xeipuuv, invopop, google/jsonschema-go, kaptinlin, the library's built-in `date-time` checker, and annotation-default formats, to achieve externally-attested full 2020-12 coverage with a `date-time` acceptance set identical under cel-go v0.28 and v0.32, accepting a single-maintainer dependency and write failures on formats authors may have meant as hints, because a declared constraint that does not enforce is a trap, and the schema layer is the only layer that can assert most formats.

> **For agents:** if your task conflicts with this ADR, stop and name the conflict. Do not silently work around it; propose supersession instead.

## Decision

akb's JSON Schema layer is `github.com/santhosh-tekuri/jsonschema/v6`, compiled for draft 2020-12 with format assertion enabled globally, and akb registers its own `date-time` and `date` format checkers — pinned to the strict RFC 3339 profile of I4 — before any schema is compiled, because the write-time schema pass must guarantee CEL's `timestamp()` preconditions under both the currently pinned cel-go v0.28 and the RFC-002-A3 upgrade target v0.32, and only this library combines verified 2020-12 conformance with an override hook for built-in format checkers.

## Invariants

- **I1**: The system MUST validate the document view against the template's `schema:` block using `github.com/santhosh-tekuri/jsonschema/v6`, compiled for JSON Schema draft 2020-12.
- **I2**: The system MUST enable format assertion on the schema compiler (`Compiler.AssertFormat()`) for every compiled template schema, at write time and at sweep time.
- **I3**: WHEN a template schema declares the `format` keyword on any value, the system MUST assert that format against the value.
- **I4**: The akb-registered `date-time` checker MUST accept exactly the strict profile: four-digit year, two-digit month and day, uppercase `T` separator, two-digit hour, minute, and second with second at most 59, optional dot-separated fractional seconds, and either uppercase `Z` or a numeric offset `±hh:mm` with hour at most 23 and minute at most 59, calendar-validated.
- **I5**: The akb-registered `date` checker MUST accept exactly the RFC 3339 `full-date` production, calendar-validated; WHEN CEL bridges a declared `format: date` value, the system MUST coerce it to midnight UTC in memory only.
- **I6**: WHEN a template declares `format: time`, the system MUST validate the value as a string format only; the system MUST NOT coerce time-only values into CEL timestamps — CEL rules receive the raw string.
- **I7**: WHEN `akb template write` encounters a `format` name outside the known registry, the system MUST warn on stderr that the format will not assert; the template-load path MUST stay lenient so a template authored by a newer akb never bricks an older binary.
- **I8**: The system MUST compile each template schema once and cache the compiled schema keyed by template identity; every `RegisterFormat` call MUST precede compilation.

## Negative Constraints

- **N1** (MUST NOT · scope: `internal/**`, `cmd/akb/**`): The system MUST NOT hand-roll JSON Schema keyword semantics; the system MUST delegate draft 2020-12 keyword evaluation to the library and limit akb-authored validation code to the compiler seam and the two temporal format checkers.
- **N2** (MUST NOT · scope: `internal/**`): The system MUST NOT use the library's built-in `date-time` checker — it accepts lowercase `t`, lowercase `z`, and the leap-second form `23:59:60` (source- and runtime-verified 2026-10-06); the system MUST register the akb strict-profile checker of I4 before compilation.
- **N3** (MUST NOT · scope: `internal/**`): The system MUST NOT define the `date-time` acceptance set as whatever Go's `time.Parse(time.RFC3339, …)` accepts — raw Go parse admits comma fractional seconds, single-digit hours, and out-of-range offsets such as `+24:00` under the currently pinned cel-go v0.28; the system MUST implement the strict pattern of I4, which is a subset of both cel-go v0.28 and v0.32 `timestamp()` acceptance.
- **N4** (MUST NOT · scope: `internal/**`, `cmd/akb/**`): The system MUST NOT offer per-field or per-template opt-outs from format assertion in the V3 template format; a template author who wants a hint-only field MUST drop the `format` keyword and express the expectation in prose or a CEL `lint_rule`.

## Exceptions

No exceptions are permitted. The library's one un-overridable format (`regex`) is a non-issue — akb does not override it; RE2 via Go's stdlib `regexp` is the desired semantics. Any case that appears to need a waiver is a proposal to supersede this record.

## Verification

- **I4, N2, N3**: unit test matrix over the acceptance set — accept `2023-10-10T09:19:21Z`, `…T09:19:21.123+05:30`; reject lowercase `t`, lowercase `z`, `1990-12-31T23:59:60Z`, comma fraction `…T09:19:21,5Z`, single-digit hour `…T9:19:21Z`, offsets `+24:00` and `+03:60`, and non-dates · gate: `go test ./internal/...` · mode: **block** once implemented · remediation: make the checker match I4 exactly; never widen the test to match a looser checker.
- **I2, I3**: testscript case: a template declaring `format: email` rejects an invalid email at `akb write` with exit 1, and the sweep-time schema checker flags the same page · gate: `go test ./test/` · mode: **block** once implemented.
- **I5**: unit tests for `full-date` acceptance and rejection (`2023-02-30` MUST fail) · gate: `go test ./internal/...` · mode: **block** once implemented.
- **I6**: unit test: a `format: time` value arrives in the CEL `page` map as the raw string · mode: **block** once implemented.
- **I7**: testscript: `akb template write` with an unknown `format` name succeeds with a stderr warning naming the format · gate: `go test ./test/` · mode: **block** once implemented.
- **I8**: assertion in template-loading tests that schemas compile once per template · mode: **advisory**.
- **I1 conformance**: bowtie draft 2020-12 badge for `go-jsonschema` reports 100% passing (`https://bowtie.report/badges/go-jsonschema/compliance/draft2020-12.json`) · mode: **info** — re-fetch at every dependency upgrade.
- **Human-only residue**: whether a built-in non-temporal checker (e.g. `email`) matches a template author's expectation is judgment; the `revisit` tripwire on false rejections is the watched signal, not a mechanical check.

## Context

Origin: RFC-002 candidate RFC-002-A1 (RFC-002 §18 lifecycle — on ratification the RFC's A1 row moves to ratified as ADR-003 and body references are rewritten). RFC-002 §9 committed to library-first full 2020-12 coverage and left the library choice and the `format` assertion default to this record; §5.2 pinned a "Go-strict RFC3339" `date-time` profile to guarantee that any schema-accepted value is parseable by CEL `timestamp()`. Two findings from the research run (`.pi/research/rfc-002-a1-validator/`, 2026-10-06) shaped the final form. First, "Go-strict" is version-dependent: akb pins cel-go v0.28.0 (`go.mod:12`), whose `timestamp()` is raw `time.Parse(time.RFC3339)` — it rejects lowercase `t`/`z` and leap seconds but accepts comma fractions, single-digit hours, and `+24:00` offsets; cel-go v0.30 added a `strictRFC3339Pattern` gate (PR #1338) that rejects those too. I4's strict pattern is the intersection both versions accept, so the guarantee survives the A3 upgrade. Second, the library's built-in `date-time` checker accepts lowercase `t`/`z` and leap seconds (`format.go` source, confirmed by an in-repo runtime probe), so overriding it is necessary, not cosmetic — and `Compiler.RegisterFormat` verifiably shadows built-ins. The narrowing itself is licensed by RFC 3339 §5.6 ("a consuming spec MAY further limit the date/time syntax so that the letters 'T' and 'Z' must always be upper case") and by JSON Schema 2020-12 §7.2.1 (assertion is an implementation option; akb opts in globally). The owner ratified the assert-everything-declared policy in the research session: template writers own their mistakes.

## Decision Drivers

- The schema pass must guarantee CEL's `timestamp()` preconditions independent of the pinned cel-go version (v0.28 today, v0.32 per RFC-002-A3).
- A declared-but-unenforced constraint is the D9 species of trap; P7 refuses at the door you configure.
- CEL has no email/uri/uuid/hostname validators; the schema layer is the only layer that can assert those formats.
- The library must expose a hook that overrides a built-in format checker — load-bearing for N2, verified in source and at runtime.
- Draft 2020-12 coverage must carry external conformance evidence (bowtie / JSON-Schema-Test-Suite), not self-report.
- Both validators then share Go's RE2 pattern semantics, keeping "akb is RE2 everywhere" true.
- Single-maintainer dependency risk is acceptable only behind a thin seam that keeps swap cost to one package.

## Alternatives Considered

- **qri-io/jsonschema** — rejected: drafts 7 and 2019-09 only; last release 2021-03, no activity since. Do not re-propose unless it ships verified 2020-12 coverage and resumes maintenance.
- **xeipuuv/gojsonschema** — rejected: unmaintained since 2020, draft 7 or earlier. Do not re-propose unless maintenance resumes with 2020-12 support.
- **invopop/jsonschema** — rejected: generates schemas from Go types; it does not validate. Do not re-propose unless it gains a validator.
- **google/jsonschema-go** — rejected: ignores `format` entirely — no assertion, not even annotations (published package docs). Do not re-propose unless it gains an overridable format-assertion hook.
- **kaptinlin/jsonschema** — not selected: claims 2020-12 and offers `RegisterFormat`, but no bowtie or JSON-Schema-Test-Suite compliance evidence and no documentation that a built-in format name can be shadowed (verified 2026-10). Do not re-propose unless it publishes 2020-12 compliance results and documents built-in-name override.
- **The library's built-in `date-time` checker** — rejected: accepts lowercase `t`/`z` and `23:59:60` leap seconds. Do not re-propose unless the library narrows its built-in to the I4 profile.
- **Raw Go `time.RFC3339` acceptance as the profile** — rejected: version-dependent (lax under cel-go v0.28) and therefore breaks the §5.2 guarantee across the A3 upgrade. Do not re-propose unless akb pins cel-go at v0.30 or newer permanently and adopts cel-go's gate as the profile definition.
- **Temporal-only format assertion (the 2020-12 §7.2.1 default)** — rejected: a declared non-temporal format would silently not enforce, and no CEL backstop exists for email/uri/uuid. Do not re-propose unless assert-all produces false rejections in practice — the remediation then is a deliberate per-template opt-out, not a return to annotation-default.
- **Per-template assertion opt-in list** — rejected for the V3 template format: new declaration machinery ahead of demonstrated need. Do not re-propose unless a concrete KB needs mixed assertion posture between templates.

## Consequences

- Good, because full draft 2020-12 coverage is delegated to an externally-attested implementation (bowtie 100% passing), and the format-override hook the date bridge depends on is verified in source and at runtime.
- Good, because the I4 acceptance set is a subset of both cel-go v0.28 and v0.32 `timestamp()` acceptance, so the RFC-002 §5.2 precondition guarantee holds today and after the A3 upgrade without re-validation of existing pages.
- Good, because every declared `format` enforces something — no silent annotation-only traps — and both validators share RE2 pattern semantics.
- Bad, because the library is single-maintainer with a low bus factor; mitigated by confining it behind one compiler-construction seam and by the `revisit` trigger watching kaptinlin and maintenance lapse.
- Bad, because assert-all fails writes on formats an author meant as documentation (a loose `format: email`); mitigated by the I7 unknown-format warning and by doctrine that teaches `format` equals assertion.
- Bad, because `Compiler.Compile` is documented as not thread-safe and concurrent `Validate` on one compiled schema is undocumented; akb compiles once per template at load and validates serially in the CLI today — safe, but a constraint to re-check before any concurrent validation is introduced.
- Neutral, because RFC-002 §5.2's "Go-strict RFC3339" wording is refined by I4 (strict pattern rather than raw Go acceptance); the RFC is amended per its §18 lifecycle when this record is ratified.

## References

- Research run record: `.pi/research/rfc-002-a1-validator/` — `LOG.md` (decision log), `research-web.md` (sourced ecosystem findings), `recon-code.md` (integration points), `extract-prior-research.md` (prior research base).
- santhosh-tekuri/jsonschema `format.go`, `compiler.go`, `objcompiler.go` (branch `boon`) — built-in `date-time` laxity and the `RegisterFormat` shadowing order, quoted in `research-web.md` Q3.
- Runtime probes 2026-10-06 (coordinator-run, results inlined in `LOG.md` findings): Go `time.RFC3339` acceptance matrix; `RegisterFormat` override end-to-end.
- RFC 3339 §5.6 (ABNF, the `t`/`z` NOTE, leap seconds); Go `src/time/format.go`; cel-go PR #1338 and `common/types/timestamp.go`; JSON Schema 2020-12 validation spec §7.2.
- RFC-002 §5.2 (date bridge), §9 (coverage + format policy), §18 (candidate lifecycle).
