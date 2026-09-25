# Changelog

Behavior changes and fixes worth acting on. Versions before 0.18.0 predate this file.

## [0.20.0] — 2026-09-25

### Changed

- **BREAKING: the knowledge-base state directory is renamed `.akb/` → `.agent-kb/`,
  and its config file loses the redundant leading dot: `.akb.yaml` → `akb.yaml`.**
  The directory now carries the project's name; the tool-facing files inside it
  keep the tool's. `akb init` creates the new layout, and every command —
  discovery, path resolution, read/write, lint, templates, search — recognizes
  only it. There is no compatibility shim: a base in the old layout is reported
  as not a knowledge base, with the expected marker path named. Migrate an
  existing base by renaming the directory and config file (`mv .akb .agent-kb &&
  mv .agent-kb/.akb.yaml .agent-kb/akb.yaml`) and updating the `.akb/search.db*`
  line in the base's `.gitignore`. Tool-named surfaces are unchanged: the
  `AKB_KB` environment variable, the `*.akb.bak` backup extension, and the
  `page.akb` CEL map key.
- **The on-disk layout now has one definition site.** `internal/path` exports
  `StateDirName`, `ConfigFileName`, and the `StateDir`/`ConfigPath`/
  `TemplatesDir`/`SearchDBPath` helpers, replacing the `".akb"` string literals
  scattered across the command and internal packages.

## [0.19.1] — 2026-09-25

### Fixed

- **`akb init` stages and commits the new base through the index-lock retry.** A concurrent
  process holding the git index lock no longer aborts the initial staging of base files;
  the retry now covers both stages.
- **The compute-budget error is identified by sentinel, not by text.** `internal/cel`
  exports `ErrComputeBudget`, and a rule's evaluation error embeds the key names the rule
  reads — so text matching handed the budget remedy to rules whose field names collide with
  the phrase. Write-path dispatch now matches with `errors.Is`.
- **`akb template get --example` warns about every broken mockup rule, not just the first.**
  Revalidation previously stopped at the first unevaluable rule, hiding later failures and
  dropping the variant's staleness headline. Every failed rule is collected, then named.
- **The template-load failure is the error reported, not a stale page-read error.** When
  `akb index add` falls back to templates and the template set fails to load, the wrapped
  load error is surfaced — it names the cause where the old message named only the page;
  the per-file guard at the append, index-add, lint, and type-directory call sites keeps
  its own error, its classification, and its exit code.

## [0.19.0] — 2026-09-24

### Added

- **`akb lint` gained a tenth check, `required_fields`.** The sweep now reports every on-disk
  page that leaves a schema-required frontmatter field unset, as an error naming the missing
  fields and the template that declares them. Write-time validation only covers pages that
  passed `akb write`, so a page pulled from git, planted by hand, or written before its
  template declared a field stayed invisible until now. A page without frontmatter and a page
  whose `type` has no template declare no schema and stay with `missing_frontmatter` and
  `type_orphan`. The check reads presence only — a field set to an empty value counts as set
  — matching `akb write`, and type and enum constraints stay with the template's CEL rules.

### Changed

- **`akb approve` refuses a draft that leaves a schema-required frontmatter field unset.**
  Approval runs the write path's required-field check before it rewrites the page, so a page
  `akb write` would reject can no longer be published by approving it: the run reports the
  missing fields and the declaring template and exits 1, leaving the page untouched and no
  approval commit behind. `akb approve --all-drafts` refuses the incomplete
  drafts and approves the rest, then fails with the count of the refused drafts, so a refusal
  is never reported as part of a successful batch. A page whose `type` has no template is
  approved as before.

## [0.18.0] — 2026-09-24

### Changed

- **CEL reports wikilink targets in the link graph's spelling.** `page.ast.links[].target`
  for an explicit-destination wikilink (`[[display]](dest)`) is now the normalized
  destination — the spelling the link graph records — instead of goldmark's destination as
  written. `[[Label]](notes/x.md)` reports target `notes/x` in both readers, so a template
  rule comparing `target` with a page path sees one convention. `text` is unchanged: it
  still preserves the raw source token. The `superseded_requires_link` rule in the embedded
  `adr` template now matches a `.md`-bearing body link against a `.md`-less `supersedes`
  value, which it previously could not. Behavior spec: `docs/wikilinks.md`.

- **Explicit-destination link-graph output changed for titled and angle-bracketed
  destinations.** A destination carrying a CommonMark link title
  (`[[a]](notes/x.md "title")`) or angle brackets (`[[a]](<notes/x.md>)`) is now recorded
  under the normalized target (`notes/x`) rather than the bracket label or the raw
  destination. The same pass tightened validity to match goldmark: a malformed explicit
  destination — an unclosed angle-bracket destination (`[[a]](<b)`) or an escaped closing
  paren (`[[a]](\)`) — now falls back to a plain wikilink in both the link graph and CEL
  instead of producing a bogus destination.

- **Pre-existing pages that use `[[a]](dest)` shift their recorded target on the next write
  or index rebuild.** The recorded target becomes the normalized destination (`notes/x`
  instead of the bracket label `a` or the raw `notes/x.md`). Consistent with the new
  convention, but it is a data change: backlinks, orphan status, and broken-link lint for
  such pages reflect the new target after the page is written again or
  `akb index rebuild` runs. This shift was never noted in 0.17.0.
