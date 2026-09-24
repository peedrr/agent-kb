# Changelog

Behavior changes and fixes worth acting on. Versions before 0.18.0 predate this file.

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
