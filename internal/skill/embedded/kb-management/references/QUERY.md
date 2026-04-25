# QUERY — Answer Questions from the KB

**When:** User asks "what do we know about X?", "compare A and B", or any domain question.

## Core Rule

Never answer from general knowledge. The KB is the source of truth. Search and read the wiki first.

## Procedure

1. Search for relevant pages:
   ```bash
   akb search "<query>"
   ```

2. Read the top results:
   ```bash
   akb read <path>
   ```

3. Follow links to gather context:
   ```bash
   akb links <path>
   akb read <linked-path>
   ```

4. Synthesize the answer. Cite pages using `[[path]]` notation.

5. Log the query:
   ```bash
   akb log append query "<user's question>"
   ```

## Filtered Search

Narrow results with dimensional filters:

```bash
# Search by tag
akb search "database" --tag go

# Search by page type
akb search "architecture" --type adr

# Search by creation date
akb search "API" --after 2025-01-01

# Combined filters
akb search "deployment" --tag kubernetes --type note --after 2025-01-01
```

Filters work with `--json` for structured output.

Note: Search without filters continues to work as before, returning all matching results.

## Rules

- Search first, read selectively. `akb search` ranks by relevance.
- Use `akb backlinks <path>` to find pages that reference a topic.
- Use `akb read` for full content; `akb links` for relationships.
- Do NOT write to `kb/log.md` directly — use `akb log append`.
