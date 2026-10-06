# Index

## Adrs

- [KB Diagnosis and Repair Live in a Dedicated akb Doctor Command](kb/decisions/ADR-001-akb-doctor.md) — akb doctor owns diagnosis and repair of nonconformant KBs; remediation names akb commands, never raw git
- [Init Author Flags Record a Durable Commit Identity in akb.yaml](kb/decisions/ADR-002-init-author-flags.md) — a complete --author-name/--author-email pair at init is recorded in akb.yaml as git-author/git-email; ambient identities never are
- [JSON Schema Validation Uses santhosh v6 and Asserts Every Declared Format](kb/decisions/ADR-003-santhosh-v6-format-assertion.md) — JSON Schema validation delegates to santhosh-tekuri/jsonschema v6; every declared format asserts; akb-registered date-time/date checkers pin the strict RFC 3339 profile cel-go accepts.

## Specs

- [Init Selects the KB Versioning Mode and Resolves Commit Identity](kb/specs/SPEC-001-init-versioning.md) — akb init chooses the KB versioning mode (embed/no-git/standalone) and commit identity resolves env -> akb.yaml -> git; migrated from raw/spec/init-versioning-spec.md

## Rfcs

- [Give akb a Declared Actor Identity So Rules Can Assert Who Acted](kb/rfcs/RFC-003-declared-actor-identity.md) — Expose a resolved actor identity to akb's rule engine so templates can assert who drafted, stewards, and ratifies — mechanism in akb, roster policy in the KB.
- [JSON Schema / CEL Coexistence and JSON I/O in akb](kb/rfcs/RFC-001-json-schema-cel-coexistence.md) — JSON Schema 2020-12 alongside CEL plus JSON I/O for agent pipelines; superseded by RFC-002.
- [The Open Validation Engine — JSON Schema + CEL as akb's only validators](kb/rfcs/RFC-002-open-validation-engine.md) — JSON Schema + CEL as akb's only validators](kb/rfcs/RFC-002-open-validation-engine.md) — The open validation engine: JSON Schema + CEL as akb's only validators; byte-faithful writes; prescriptive core removed or made configurable.

