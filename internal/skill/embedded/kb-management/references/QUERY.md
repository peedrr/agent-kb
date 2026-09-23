# QUERY — Answer Questions from the KB

**When:** User asks "what do we know about X?", "compare A and B", or any domain question.

**Selection:** every command here picks its KB per invocation with `--kb <path>` or the `AKB_KB` environment variable; there is no stored default (`akb discover` lists nearby bases).

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

```bash
akb search "database" --tag go                       # by tag
akb search "architecture" --type adr                 # by page type
akb search "API" --after 2025-01-01                  # by creation date
akb search "deployment" --tag k8s --type note        # combined
```

All filters work with `--json` for structured output.

## Rules

- Search first, read selectively. `akb search` ranks by relevance.
- Use `akb backlinks <path>` to find pages that reference a topic.
- Use `akb read` for full content; `akb links` for relationships.
- **DO NOT edit the log directly** — use `akb log append`.
