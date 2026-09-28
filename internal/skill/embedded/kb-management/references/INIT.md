# INIT — Bootstrap a New KB

**When:** No KB exists; user wants a new knowledge base.

## Decide how the KB is versioned first

`akb init` writes the same KB structure in three versioning modes, and the mode
is a decision, not a default: inside a git repository an invocation that names
none is refused (exit 2) with both flags spelled out. Choose before running the
command.

| Mode | Command | What it means |
|------|---------|---------------|
| Standalone | `akb init <kb-name>` | The KB is a git repository of its own; `akb` commits the pages into it. |
| Embedded | `akb init <kb-name> --embed` | The KB joins the history of the repository it lands in; `akb` commits only `kb/`, `raw/` and `.agent-kb/`, never other worktree changes. |
| Unversioned | `akb init <kb-name> --no-git` | The KB is not versioned at all; `akb` writes its files and stops. |

- Choose **embedded** when the KB documents the project around it and should be
  reviewed, branched and reverted with that project's code. The KB lives in a
  subdirectory of that project.
- Choose **standalone** when the KB is not the project's — a personal reference,
  a KB that spans several projects, or one that keeps its own history.
- Choose **`--no-git`** when the content must stay out of version control
  (private or machine-local notes, generated material). The KB is excluded from
  the enclosing repository's git status through `.git/info/exclude`, a file
  local to this clone: the host does not track an ignore rule for the KB, and a
  clone of the host does not carry the exclusion.
- `--embed` needs a repository: outside one it is refused. A KB that is not
  versioned anywhere is the `--no-git` case.
- `--force` is not a versioning mode. It bypasses the warning for a KB laid out
  directly in a repository root (`akb init . --embed --force`), and it never
  implies a mode: a repository-root invocation still needs `--embed` or
  `--no-git`.
- The mode is recorded as the `versioning` key (`git` or `none`) in
  `.agent-kb/akb.yaml`.

## Procedure

1. Initialize the KB:
   ```bash
   akb init <kb-name>                # standalone: a repository of its own
   akb init <kb-name> --embed        # inside a repository: versioned with it
   akb init <kb-name> --no-git       # not versioned
   ```
   Replace `<kb-name>` with a short identifier (no slashes, no `..`). The KB
   lands at `<working directory>/<kb-name>`. `akb init` does not select it.
   `--description "what this KB holds"` is optional; `akb discover` displays it.

2. Verify (every command below addresses the new KB with `--kb`):
   ```bash
   akb --kb <kb-name> status
   akb --kb <kb-name> list
   ```

3. (Optional) Create an overview page — copy the `note` showcase first, because
   a fresh KB has no templates and refuses writes of unknown types:
   ```bash
   akb template get note --full --examples > <kb-name>/.agent-kb/templates/note.yaml
   akb --kb <kb-name> write overview.md <<'EOF'
   ---
   title: Overview
   type: note
   summary: Purpose and scope of this knowledge base.
   tags: [overview]
   ---
   
   # Overview
   
   <purpose statement>
   
   ## Open Questions
   EOF
   ```

4. Update the index and log:
   ```bash
   akb --kb <kb-name> index add "overview.md" "KB purpose and open questions"
   akb --kb <kb-name> log append ingest "Initialized KB <kb-name>"
   ```

## What the invocation reports

Every successful init prints where the KB landed:

```
Initialized KB "docs" at /abs/path/to/docs
```

### Embedded

```bash
cd <host-project>
akb init docs --embed
```

The KB lands in `<host-project>/docs`, and the invocation adds:

```
kb: docs is embedded inside repository /abs/path/to/host-project; akb commits only kb/, raw/, .agent-kb/
```

That line names the repository the KB now shares. The init commit records the
KB's own paths only, so changes another author staged in the host worktree stay
staged and uncommitted.

### Commit identity

Every init that versions the KB in git — standalone as well as embedded —
also reports the commit identity it resolved, in one of three forms:

```
commit identity: from git config (Ada Lovelace <ada@example.com>) — not recorded; each machine's git identity applies
commit identity: from AKB_AUTHOR_NAME/AKB_AUTHOR_EMAIL (Ada Lovelace <ada@example.com>) — not recorded; the environment of each invocation applies
commit identity: agent-kb <agent@agent-kb> (default — no git identity found; recorded in akb.yaml, edit git-author/git-email to change)
```

An identity from the environment or from git config belongs to that invocation
or to that machine, so it is not written into `.agent-kb/akb.yaml`. Only the akb
default is recorded there, so a KB cloned to a machine without a git identity
still commits. A `--no-git` init commits nothing and reports no identity.

### Unversioned

```bash
akb init scratch --no-git
```

Inside a repository the invocation reports what it excluded from the host's git
status:

```
kb: excluded scratch/ from host git tracking via .git/info/exclude (local to this clone; remove that line to undo)
```

A KB initialized at the repository root is excluded by the three directories it
owns, and the undo guidance refers to them in the plural:

```
kb: excluded .agent-kb/, kb/, raw/ from host git tracking via .git/info/exclude (local to this clone; remove those lines to undo)
```

## Refusals

Each refusal exits 2 and prints one `usage:` message. Every message is
self-sufficient: it states what was found and what each choice does, so the
decision can be made from the message alone.

**No versioning mode, inside a repository.** `--embed` and `--no-git` are the
choices; neither is the default:

```
usage: docs is inside git repository /abs/path/to/host-project — choose how the KB is versioned:

  --embed    the KB joins this repository's history; akb commits only its own paths (kb/, raw/, .agent-kb/), never other worktree changes
  --no-git   the KB is not versioned; akb excludes it from the host's git status via .git/info/exclude (local to this clone)
```

A target that *is* the repository root is named as a repository of its own, and
the layout warning is appended below the same two flag lines:

```
usage: host-project is a git repository — choose how the KB is versioned:

  --embed    the KB joins this repository's history; akb commits only its own paths (kb/, raw/, .agent-kb/), never other worktree changes
  --no-git   the KB is not versioned; akb excludes it from the host's git status via .git/info/exclude (local to this clone)

This will write .agent-kb/, kb/ and raw/ directly to the repo root. --force if desired, otherwise prefer a subdirectory: akb init <name> --embed.
```

**A KB laid out directly in a repository root.**

```
usage: This will write .agent-kb/, kb/ and raw/ directly to the repo root. --force if desired, otherwise prefer a subdirectory: akb init <name> --embed.
```

`--force` is the right judgment only when the KB genuinely owns the repository —
the KB's `kb/`, `raw/` and `.agent-kb/` sit beside the project's own files. In
every other case the subdirectory of the message is the better structure.

**`--embed` outside any repository.**

```
usage: --embed: docs is not inside a git repository — omit --embed to create a standalone repository
```

Here `--embed` has nothing to join: either run the init inside the repository
the KB belongs to, or drop `--embed` for a standalone KB. `--no-git` applies
when the KB should not be versioned at all.

**`--embed` together with `--no-git`.**

```
usage: --embed and --no-git cannot be used together
```

The two name opposite outcomes for the same repository, so no reconciliation is
possible: pick the one the user asked for and run it alone.

## A host merge in progress blocks the KB

Committing commands preflight the repository that hosts the KB before they write
anything, and refuse while that repository carries a merge in progress — even a
merge whose conflicts are already resolved:

```
usage: write page: merge in progress in repository /abs/path/to/host-project (outside the knowledge base, in: hostnotes.txt): akb commits are blocked until it is completed or aborted. This merge belongs to the host project — do not resolve it from the KB; retry later or surface to the user.
```

The merge is the host project's work: **do not resolve it from the KB.** Report
the blockage, retry after the merge is completed or aborted, or surface it to
the user — never commit, stash or abort the host's merge to unblock a KB write.
The context before `merge in progress` names the step that refused (`write
page:`, `commit preflight:`); the refusal itself is the same however it is
reached, and the changed path it names belongs to the host's merge.

## Notes

- `akb init` sets up everything but the page types — structure, an **empty** template directory, search index, and version control. No template is seeded, so every typed page write is refused with `unknown type` until a template exists for that type. Copy a showcase (`akb template get note --full --examples > .agent-kb/templates/note.yaml`) or author one; see `TEMPLATE.md`.
- Run `akb init` from the directory where you want the KB to live.
- After `akb init` (which does not select a base), address the KB on every
  invocation with `--kb <path>` or `AKB_KB=<path>`; a relative path resolves
  against the working directory, so `--kb <kb-name>` works from the directory
  you initialized in. See `SKILL.md` for the full selection rules.
- An embedded KB and a standalone one are addressed the same way: the layout
  changes what is committed, never how the base is selected.
