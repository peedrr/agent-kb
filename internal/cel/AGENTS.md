# internal/cel/

**Parent:** `./AGENTS.md`

## OVERVIEW

CEL expression evaluation engine for template-based page validation and linting. Bridges frontmatter, markdown AST, and template rules.

## FILES

| File | Purpose |
|------|---------|
| `engine.go` | CEL environment builder, rule compiler with caching, evaluator with panic recovery |
| `types.go` | AST structs (Heading, Link, CodeBlock) with CEL tags for native type registration |
| `errors.go` | ValidationError struct with RuleID, Message, Line, Severity |
| `pagebuilder.go` | BuildPage/BuildOldPage — assembles `page`/`old_page` maps from frontmatter + AST |

## KEY TYPES

| Type | Purpose |
|------|---------|
| `ValidationError` | Structured error: RuleID/Message/Line/Severity |
| `Heading` | `level`, `text`, `line` — for CEL `page.ast.headings` |
| `Link` | `target`, `text`, `is_wikilink`, `line` — for CEL `page.ast.links`; `is_wikilink` is true for every wikilink form, and `target` carries the link graph's spelling |
| `CodeBlock` | `language`, `line` — for CEL `page.ast.code_blocks` |

## KEY FUNCTIONS

| Function | Purpose |
|----------|---------|
| `NewEnv()` | Creates CEL env with `page`, `old_page`, `now` variables |
| `CompileRule(env, expr)` | Parses/compiles CEL expression; caches result in `sync.Map` |
| `Evaluate(ctx, prg, vars)` | Runs compiled program; recovers from panics; converts cost-limit-exceeded to "exceeded compute budget" error |
| `BuildPage(relPath, fm, body, astDoc, source)` | Assembles `page` map: `file`, `frontmatter`, `content`, `ast`, `akb` |
| `BuildOldPage(relPath, store)` | Reads on-disk file, returns `page` map or `nil` for new files |

## CEL VARIABLES

| Variable | Type | Description |
|----------|------|-------------|
| `page` | `map[string]any` | Post-modification page state |
| `old_page` | `map[string]any` | Pre-modification page state (nil for new files) |
| `now` | `timestamp` | Current time (injected during lint sweeps) |

## PAGE MAP STRUCTURE

```
page.file.path      — relative path
page.file.name      — basename
page.file.dir       — directory
page.frontmatter.type
page.frontmatter.title
page.frontmatter.<field>
page.content.raw
page.content.word_count
page.content.char_count
page.ast.headings   — []{level, text, line}
page.ast.links      — []{target, text, is_wikilink, line}; is_wikilink true for all four wikilink forms, target in the link graph's spelling
page.ast.code_blocks — []{language, line}
page.akb.provenance_markers
page.akb.annotations
```

## NOTES

- Programs cached in `sync.Map` keyed by expression string
- Cost budget: `MaxCostLimit` (100000), applied when `CompileRule` compiles the program
- Panic recovery catches `interpreter.EvalCancelledError` (cost exceeded) and unknown panics
- `old_page` is nil for new files; `has(old_page)` returns `false` in CEL
- Any frontmatter string that parses as RFC3339 or date-only `2006-01-02` is converted to `time.Time`
- `pagebuilder.go` walks the Goldmark AST itself to build `page.ast`
- `page.ast.links` covers all four wikilink forms — `[[target]]`, `[[target|display]]`, `[[target#heading]]`, and `[[display]](dest)`. `flattenLinks` merges `markdown.ParseWikilinks` into the Goldmark `ast.Link` walk, so `is_wikilink` is true for every wikilink and false for a plain markdown link.
- The pipe and heading forms take `target` from the bracket target and `text` from the display: `[[target|display]]` → text `display`, `[[target#heading]]` → text `heading`.
- The explicit-destination form takes `target` from the destination, which beats the bracket label (`[[a|b]](dest)` → target `dest`, display `b`). `target` carries the one spelling the link graph records, so `[[display]](notes/page.md)` → target `notes/page` and the graph resolves that spelling to the page. The spelling is normalized by dropping surrounding whitespace and angle brackets, a leading `./`, a trailing `.md`, and an optional trailing CommonMark link title; where only Goldmark reports a destination, it is normalized the same way except that a scheme-qualified destination (`://`) is left as written. `text` still preserves the raw source token: for this form it is Goldmark's rendered link label — bracketed, `[display]` — not the bare display text, so only `target` changes spelling.
- Empty-destination hybrids — `[[g]]()`, `[[g]](   )`, `[[g]](<>)`, an unclosed `(`, and a destination that normalizes away (`[[g]](.md)`, `[[g]](./)`) — are plain wikilinks: one entry whose `target` is the bracket target, the parens ignored, which is the target the link graph records for them.
- Dedupe is start-equality on the token offset, which keeps nesting deterministic: `[text [[f]]](dest)` yields the outer Goldmark entry plus the wikilink's own entry, while `[[a [b](c)]]` yields only the inner Goldmark entry (target `c`, text `b`, `is_wikilink: false`) and no wikilink entry, because the wikilink token regex cannot span the inner `]` of `[b]`.
