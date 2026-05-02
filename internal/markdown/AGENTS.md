# internal/markdown/

**Parent:** `./AGENTS.md`

## OVERVIEW

Markdown parsing utilities: wikilinks, quality annotations, provenance markers, AST flattening. All parsers exclude fenced code blocks, inline code, and HTML comments.

## FILES

| File | Purpose |
|------|---------|
| `wikilink.go` | Parse `[[target]]`, `[[target|display]]`, `[[target#heading]]` |
| `annotation.go` | Parse `<!-- olw-auto: key=val -->` HTML comments |
| `provenance.go` | Parse `^[inferred]`, `^[ambiguous]`, `^[extracted]` markers |
| `ast.go` | Goldmark AST flattener: FlattenHeadings, FlattenLinks, FlattenCodeBlocks |
| `offset.go` | `offsetToLine(source, offset)` helper |

## KEY TYPES

| Type | Fields / Purpose |
|------|------------------|
| `Wikilink` | `Target`, `Display`, `HasHeading`, `Heading` |
| `Annotation` | `Type` (always "olw-auto"), `Fields` map, `Position` |
| `ProvenanceMarker` | `Type` (inferred/ambiguous/extracted), `Position` |

## EXCLUSION LOGIC

All three parsers skip matches inside:
- Fenced code blocks (```...```)
- Inline code (backtick-delimited)
- HTML comments (<!-- ... -->)

`computeExclusions()` builds ranges; `exclusionSet.isExcluded()` checks before accepting matches.

## KEY FUNCTIONS

| Function | Purpose |
|----------|---------|
| `ParseWikilinks(content)` | `[]Wikilink` with heading anchor support |
| `ParseAnnotations(content)` | `[]Annotation` from olw-auto HTML comments |
| `ParseProvenanceMarkers(content)` | `[]ProvenanceMarker` excluding protected regions |
| `StripProvenanceMarkers(content)` | Removes `^[inferred/ambiguous/extracted]` markers |
| `CountMarkersByType(markers)` | Returns `map[type]count` |
| `FlattenHeadings(doc, source)` | `[]cel.Heading` from Goldmark AST |
| `FlattenLinks(doc, source)` | `[]cel.Link` from Goldmark AST |
| `FlattenCodeBlocks(doc, source)` | `[]cel.CodeBlock` from Goldmark AST |
| `offsetToLine(source, offset)` | Returns 1-based line number from byte offset |

## NOTES

- Wikilink heading anchors (`#heading`) are stripped before link resolution
- `StripProvenanceMarkers` preserves markers inside code blocks/comments
- Used by `cmd/akb/approve.go` to clean annotations before setting `is_draft: false`
- `internal/cel/pagebuilder.go` inlines equivalent AST flattening to avoid import cycle
