# Wikilinks — Behavior Spec

**Pinned at v0.18.0.** This file records what a wikilink does in akb: the forms, the
destination rules, the exclusion set, the degenerate-token outcomes, and the one target
spelling both readers share. A behavior change that does not land here is a deviation;
deliberate divergences from CommonMark and goldmark are listed in
[`KNOWN-LIMITATIONS.md`](../KNOWN-LIMITATIONS.md).

## Two readers, one spelling

A wikilink is read twice:

| Reader | Entry point | Output |
|---|---|---|
| Link graph | `markdown.ParseWikilinks` (`internal/markdown/wikilink.go`), consumed by `internal/linkgraph` | one row per token: `RawTarget`, `Display`, resolved path |
| CEL | `flattenLinks` (`internal/cel/pagebuilder.go`) | `page.ast.links[]`: `target`, `text`, `is_wikilink`, `line` |

Both readers must name the same target for every form. `page.ast.links[].target` carries the
link graph's spelling by construction: the parser's normalized destination when the token has
one, and otherwise the bracket target. A plain CommonMark `[text](dest)` link is not a
wikilink (`is_wikilink: false`), keeps goldmark's destination as written, and is ignored by
the link graph — there is no second spelling to reconcile.

## The four forms

| Form | `target` (both readers) | Link-graph `Display` | CEL `text` |
|---|---|---|---|
| `[[target]]` | `target` | `target` | `target` |
| `[[target\|display]]` | `target` | `display` | `display` |
| `[[target#heading]]` | `target` | `heading` | `heading` |
| `[[display]](dest)` | the normalized `dest` | the bracket label, or the pipe display when one is written | goldmark's rendered label — bracketed: `[display]`, `[a\|b]` |

The pipe and heading forms take the target from the bracket part. The explicit-destination
form takes the target from the destination, which beats the bracket label
(`[[a|b]](dest)` → target `dest`), and a `#heading` in the bracket part is discarded rather
than merged into the destination. `text` preserves the raw source token — for the
explicit-destination form that is goldmark's bracketed label, so only `target` changes
spelling there.

Every wikilink yields exactly one entry per token, so `.all()` and `.exists()` over
`page.ast.links` see each token once. A wikilink nested inside a markdown link label is the
exception, and it is two tokens, not one: `[text [[f]]](dest)` yields the outer goldmark
entry (target `dest`, `is_wikilink: false`) plus the wikilink's own entry (target `f`). A
bracket run inside an explicit destination is destination text rather than a token: nested
in the quoted title or in a path destination it carries no entry of its own in either
reader — see Degenerate tokens.

## Adjacency rule

An explicit destination requires the opening paren to be **adjacent** to the closing
brackets, with no whitespace between them. `[[Paris]] (the city)` is a plain wikilink whose
target is `Paris` and whose span ends at `]]`.

## Destination validity

The destination is scanned with the CommonMark inline-link tail:

- an angle-bracketed destination ends at its closing `>`, and may not contain a newline;
- a bare destination ends at the first unescaped whitespace or unbalanced `)`, with balanced
  nested parentheses accepted;
- an optional trailing link title (double-quoted, single-quoted, or parenthesized) follows
  the destination and may contain a `)`; a quoted title may span a newline, so **a token's
  span can cross lines**.

A destination is accepted only when it is a complete CommonMark link destination with an
optional trailing title. Otherwise the parens are ignored and the token is a **plain
wikilink**: one entry whose target is the bracket target, and whose span ends at the closing
brackets. Plain-wikilink outcomes:

| Source | Outcome |
|---|---|
| `[[a]]()` | plain wikilink, target `a` |
| `[[a]](   )` | plain wikilink, target `a` |
| `[[a]](<>)` | plain wikilink, target `a` |
| `[[a]](unclosed` | plain wikilink, target `a` |
| `[[a]](<b)` — unclosed angle-bracket destination | plain wikilink, target `a` (goldmark agrees: no link) |
| `[[a]](\)` — escaped closing paren | plain wikilink, target `a` (goldmark agrees: no link) |
| `[[a]](notes/x.md "title)` — unclosed title | plain wikilink, target `a` |
| `[[a]](notes/x.md "title" extra)` — text after the title | plain wikilink, target `a` |
| `[[a]](.md)`, `[[a]](./)` — destination normalizes away | plain wikilink, target `a` |

## Destination normalization

`normalizeDest` cleans an accepted destination before it becomes the target:

1. surrounding whitespace is trimmed;
2. an optional trailing link title is dropped;
3. surrounding angle brackets are stripped, then whitespace is trimmed again;
4. a leading `./` and a trailing `.md` are collapsed — repeated affixes collapse too, so
   `././a.md.md` → `a` and `notes/x.md.md` → `notes/x`.

Every spelling of the same `.md`-bearing destination therefore normalizes to one target:
`[[Paris]](concepts/paris.md)`, `[[Paris]](<concepts/paris.md>)`,
`[[Paris]](./concepts/paris.md)`, and `[[Paris]](concepts/paris.md "Paris")` all have target
`concepts/paris`, and the graph resolves that spelling to the same page.

**`://` scheme guard.** A scheme-qualified destination is left as written on the
goldmark-reported path only (`normalizeWikilinkTarget`, used when the parser has no
destination for a wikilink-form token). On the parser path the destination is normalized
like any other, so `[[a]](https://x/y.md)` records `https://x/y` in both readers — see
[`KNOWN-LIMITATIONS.md`](../KNOWN-LIMITATIONS.md).

## Exclusion set

A wikilink-shaped token is skipped — recorded by neither reader — when its bracket part lies
inside:

- a **fenced code block** (``` … ```);
- **inline code** (backtick-delimited spans);
- an **HTML comment** (`<!-- … -->`);
- a **backslash-escaped bracket**: an opening `[` or a closing `]` preceded by an odd number
  of backslashes. `\[[alpha]](notes/x.md)` and `[[alpha\]](notes/x.md)` both render as
  literal text, so neither is a link; `\\[[alpha]]` is not escaped and still links;
- an **indented code block**: a line indented four or more columns past the content column
  of the innermost open list item (four columns at the top level). A tab advances to the
  next multiple of four, both in a line's leading indent and in the whitespace after a list
  marker — `-\titem` puts the item's content at column four. An indented line that follows a
  paragraph with no intervening blank line is a lazy continuation — prose, not code — and it
  closes no list item, whatever its indent. A line that opens a block which interrupts a
  paragraph — a fence delimiter, a thematic break, an ATX heading, a block quote, or an HTML
  block of CommonMark type 1-6 — instead runs the list bookkeeping and closes the items it
  falls outside of; a setext heading underline never interrupts. A line immediately after
  such an interrupting block, with no blank line between them, is still judged a lazy
  continuation — its wikilink is recorded where goldmark renders the line as an indented
  code block — a divergence ledgered in [`KNOWN-LIMITATIONS.md`](../KNOWN-LIMITATIONS.md).

The exclusion is line-structured while a token's span can cross lines: a token whose span
merely crosses an indented line (a quoted title spanning a newline) is still recorded; only
a token whose bracket part lies inside the block is excluded.

## Degenerate tokens

Bracket runs of three or more are one balanced token: the token opens with a run of two or
more `[` and closes at the first following `]` run at least as long as the opening run,
consuming exactly as many `]` as the opening run has `[`. An opening run that cannot close
that way is retried one `[` further right, so the inner text is never a bracket fragment
such as `[a`. Pinned outcomes:

| Source | Outcome |
|---|---|
| `[[[a]]]` | one link: target `a`, display `a`, span `[0,7)` |
| `[[[a]]](notes/x.md)` | one link: target `notes/x` in **both** the link graph and `page.ast.links` (never the bracket fragment `[a`), display `a`, span covers the whole token |
| `[[[[a]]]](x.md)` | one link: target `x`, display `a`, span covers the whole token |
| `[[[[a]]` | one link: target `a`, display `a`, span `[2,7)` — the four-bracket run never closes, so the retry pairs the innermost `[[a]]` inside it |
| `[[[a]]](x.md "t")` | one link: target `x`, display `a`, span covers the whole token — the destination's trailing link title is stripped by normalization |
| `[[a]]([[b]])` | `page.ast.links` emits one entry, target `b`, text `b` — the outer token and its goldmark link are dropped. The link graph still records the bracket-shaped `[[b]]`; ledgered in [`KNOWN-LIMITATIONS.md`](../KNOWN-LIMITATIONS.md) |
| `[[a]](<[[b]]>)` | `page.ast.links` emits one entry, target `b`, text `b` — destination normalization strips the angle brackets, after which the outer token and its goldmark link are dropped. The link graph still records the bracket-shaped `[[b]]`; ledgered in [`KNOWN-LIMITATIONS.md`](../KNOWN-LIMITATIONS.md) |
| `[[a]](notes/[[weird]].md)` | one entry, the outer link: target `notes/[[weird]]`, text `[a]` — a bracket run inside the destination path is path text, so the parser records no token for it and both readers name only the outer link |
| `[[a]](x.md "see [[b]]")` | one entry, the outer link: target `x`, text `[a]` — a bracket run inside the quoted title is title text, so the parser records no token for it and both readers name only the outer link |
| `[[#h]](dest)` | target `dest`, empty display (the bracket part holds only a heading, and no pipe display was written); CEL `text` is goldmark's label `[#h]` |
| `[[a#h|]](dest)` | target `dest`, display `a` — an empty pipe display does not count, so the heading-derived display falls back to the bracket label; CEL `text` is goldmark's label `[a#h|]` |
| `[[the page]](.md)` | plain wikilink: target `the page`, display `the page`, CEL `text` `the page` — the parser display, **not** the bracketed label, the same treatment `[[a]](<>)` gets |
| `[[a [b](c)]]` | one entry: the inner goldmark link (target `c`, text `b`, `is_wikilink: false`) — the closing run must be at least as long as the opening run, so a token never spans the inner `]` of a nested markdown link |

## Where it is pinned

- Parser: `TestParseWikilinks`, `TestParseWikilinksExplicitDestination`,
  `TestParseWikilinksExplicitDestinationDisplayFallbacks`,
  `TestParseWikilinksDegenerateBracketRuns`,
  `TestParseWikilinksDegenerateBracketRunsAtOffsetZero`, `TestParseWikilinksEscapedBrackets`,
  `TestParseWikilinksIndentedCodeBlocks`, `TestParseWikilinksIndentedCodeBlocksInLists`,
  `TestParseWikilinksIndentedCodeBlockSpanBoundary` (`internal/markdown/wikilink_test.go`).
- CEL: `TestFlattenLinksWikilinkForms`, `TestFlattenLinksMatchesLinkGraphTargetSpelling`,
  `TestFlattenLinksDegenerateBracketRunNamesTheDestination`,
  `TestFlattenLinksWikilinkInsideDestination`, `TestFlattenLinksWikilinkInsidePathDestination`,
  `TestFlattenLinksExplicitDestinationSharesGoldmarkLinkStart`, `TestNormalizeWikilinkTarget`
  (`internal/cel/pagebuilder_test.go`).
- End to end: `test/testdata/wikilink_forms.txt` writes each form and asserts the target the
  link graph records equals the target CEL matched; `test/testdata/superseded_adr.txt`
  exercises a `.md`-bearing destination against a `.md`-less `supersedes` value.

## Parser spec-frozen

Parser spec-frozen as of v0.19.0 — divergences from goldmark on degenerate input are documented in this file and KNOWN-LIMITATIONS.md, not patched.

One documented divergence: a lowercase HTML declaration (`<!doctype html>`, `<!a>`) or a lowercase `<![cdata[` is treated as a paragraph interrupt here, where goldmark requires an uppercase letter after `<!` or the exact `<![CDATA[` spelling and reads the line as a lazy paragraph continuation.
