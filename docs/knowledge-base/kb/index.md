# Index

## Adrs

- [KB Diagnosis and Repair Live in a Dedicated akb Doctor Command](kb/decisions/ADR-001-akb-doctor.md) — akb doctor owns diagnosis and repair of nonconformant KBs; remediation names akb commands, never raw git
- [Init Author Flags Record a Durable Commit Identity in akb.yaml](kb/decisions/ADR-002-init-author-flags.md) — a complete --author-name/--author-email pair at init is recorded in akb.yaml as git-author/git-email; ambient identities never are

## Specs

- [Init Selects the KB Versioning Mode and Resolves Commit Identity](kb/specs/SPEC-001-init-versioning.md) — akb init chooses the KB versioning mode (embed/no-git/standalone) and commit identity resolves env -> akb.yaml -> git; migrated from raw/spec/init-versioning-spec.md

## Rfcs

- [Give akb a Declared Actor Identity So Rules Can Assert Who Acted](kb/rfcs/RFC-003-declared-actor-identity.md) — Expose a resolved actor identity to akb's rule engine so templates can assert who drafted, stewards, and ratifies — mechanism in akb, roster policy in the KB.

