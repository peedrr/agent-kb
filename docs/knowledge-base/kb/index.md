# Index

## Adrs

- [KB Diagnosis and Repair Live in a Dedicated akb Doctor Command](kb/decisions/ADR-001-akb-doctor.md) — akb doctor owns diagnosis and repair of nonconformant KBs; remediation names akb commands, never raw git
- [Init Author Flags Record a Durable Commit Identity in akb.yaml](kb/decisions/ADR-002-init-author-flags.md) — a complete --author-name/--author-email pair at init is recorded in akb.yaml as git-author/git-email; ambient identities never are
- [JSON Schema Validation Uses santhosh v6 and Asserts Every Declared Format](kb/decisions/ADR-003-santhosh-v6-format-assertion.md) — JSON Schema validation delegates to santhosh-tekuri/jsonschema v6; every declared format asserts; akb-registered date-time/date checkers pin the strict RFC 3339 profile cel-go accepts.
- [CEL Runs on cel.dev/cel-go v0.32 with a Pinned Extension Set](kb/decisions/ADR-004-cel-go-upgrade-pinned-extensions.md) — cel-go upgrades to cel.dev/cel-go v0.32.0 in one require+import rewrite; five ext libraries enabled in NewEnv pinned at their highest v0.32.0 versions; regex plan-size knob stays unbounded.
- [The Date Bridge Coerces Only Schema-Declared Fields and Is Statically Checked](kb/decisions/ADR-005-date-bridge-determinism.md) — The date bridge coerces only schema-declared temporal fields, identically for page/old_page; template write warns on unbacked temporal calls; the duration seam is documented, bridging deferred to A5.
- [TemplateV3 Retires schema.frontmatter — Migration Is a Recipe, Not a Command](kb/decisions/ADR-006-schema-frontmatter-retirement.md) — Migration Is a Recipe, Not a Command](kb/decisions/ADR-006-schema-frontmatter-retirement.md) — TemplateV3 hard-rejects schema.frontmatter; migration is a CHANGELOG recipe driven by ADR-005 warnings — no command, no alias

## Specs

- [Init Selects the KB Versioning Mode and Resolves Commit Identity](kb/specs/SPEC-001-init-versioning.md) — akb init chooses the KB versioning mode (embed/no-git/standalone) and commit identity resolves env -> akb.yaml -> git; migrated from raw/spec/init-versioning-spec.md
- [Write-Time Validation Failures Merge Into One Structured Report](kb/specs/SPEC-002-structured-validation-report.md) — merged write-time validation report (schema keyword + JSON Pointer, CEL rule ID) on stdout under write --json; widened runTemplateValidations; RFC-002 candidate RFC-002-S1

## Rfcs

- [Give akb a Declared Actor Identity So Rules Can Assert Who Acted](kb/rfcs/RFC-003-declared-actor-identity.md) — Expose a resolved actor identity to akb's rule engine so templates can assert who drafted, stewards, and ratifies — mechanism in akb, roster policy in the KB.
- [JSON Schema / CEL Coexistence and JSON I/O in akb](kb/rfcs/RFC-001-json-schema-cel-coexistence.md) — JSON Schema 2020-12 alongside CEL plus JSON I/O for agent pipelines; superseded by RFC-002.
- [The Open Validation Engine — JSON Schema + CEL as akb's only validators](kb/rfcs/RFC-002-open-validation-engine.md) — JSON Schema + CEL as akb's only validators](kb/rfcs/RFC-002-open-validation-engine.md) — JSON Schema + CEL as akb's only validators](kb/rfcs/RFC-002-open-validation-engine.md) — JSON Schema + CEL as akb's only validators](kb/rfcs/RFC-002-open-validation-engine.md) — JSON Schema + CEL as akb's only validators](kb/rfcs/RFC-002-open-validation-engine.md) — JSON Schema + CEL as akb's only validators](kb/rfcs/RFC-002-open-validation-engine.md) — The open validation engine: JSON Schema + CEL as akb's only validators; byte-faithful writes; prescriptive core removed or made configurable.

