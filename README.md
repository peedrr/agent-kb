# agent-kb (`akb`)

A git-tracked knowledge base for agents: markdown pages with typed templates
and CEL validation, SQLite FTS5 search, a wikilink graph, lint checks, an
append-only log, and raw-source drift tracking.

`akb` never calls an LLM. The agent using it supplies the knowledge; `akb`
reads, writes, searches, links, and lints it.

## Build

```bash
make build     # → bin/akb
make test      # go test ./...
nix develop    # dev shell: Go, gopls, delve, golangci-lint
```

`git` must be on `PATH`: every knowledge base is a git repository, and every
write is committed.

## Quick start

```bash
akb init my-kb                                  # directory, git repo, search index
akb --kb my-kb status

# init seeds no templates: install the note type before the first write
akb template get note --full --examples > my-kb/.akb/templates/note.yaml

akb --kb my-kb write notes/overview.md <<'EOF'
---
type: note
title: Overview
summary: What this knowledge base covers.
tags: [overview]
---

# Overview

Purpose, scope, and open questions.
EOF

akb --kb my-kb search "overview"
akb --kb my-kb lint
```

`akb init` creates the base at `./my-kb` and does not select it. New pages are
drafts; `akb approve` publishes them.

## Selecting a knowledge base

`akb` keeps no stored default: every invocation resolves exactly one base.

1. `--kb <path>` — the flag wins when it is passed.
2. `AKB_KB=<path>` — the environment variable selects when the flag is absent.
3. Neither — the command fails with a usage error (exit 2) that lists the bases
   discovered near the working directory, nearest first.

The resolved path must hold the `.akb/.akb.yaml` config file that marks a
knowledge base — a regular file, not a directory. A path without it is a usage
error (exit 2) whose report names the missing marker and points at
`akb discover` for the bases nearby.

A path may be absolute, relative to the working directory, or start with `~`:

```bash
akb --kb ~/kbs/notes status
AKB_KB=~/kbs/notes akb list
akb --kb . list              # the base you are standing in
```

`akb discover` reports the bases near a directory without selecting one:

```bash
akb discover                 # nearest first
akb discover ~/work          # scan another directory's neighborhood
akb discover --json          # [{"name": ..., "path": ..., "description": ...}] (description optional)
```

There is no registry and no "current KB" to switch: `akb use` and `akb registry`
were removed, and a `~/.config/agent-kb/registry.yaml` left on disk is ignored
(akb prints a one-line note when it exists). Wire a base per invocation
(`--kb`), per shell session (`AKB_KB`), or per directory (the same export from
a direnv `.envrc`, with `direnv allow` once). Mutating commands echo the base
they act on to stderr — `kb: notes (/home/you/kbs/notes)` — before they touch
it.

## Topologies

Where a base lives decides which repository records its history. Every layout is
addressed the same way (`--kb`, `AKB_KB`, or a direnv export); the layouts
differ in where that path points.

### One repository per KB

The default: `akb init` creates a dedicated repository, independent of any
product repository.

```bash
cd ~/kbs && akb init notes

# per shell session
export AKB_KB="$HOME/kbs/notes"

# per directory with direnv — ~/kbs/.envrc:
#   export AKB_KB="$(pwd)/notes"

# per invocation
akb --kb ~/kbs/notes status
```

### Several KBs in one repository

One history to back up and review, several focused bases. `akb init` always
creates a repository, so when a shared repository should own the files, drop
each base's generated `.git` and commit the base directories into it:

```bash
mkdir -p ~/kbs && cd ~/kbs && git init
akb init notes && akb init incidents
rm -rf notes/.git incidents/.git
git add -A -- notes incidents && git commit -m "add knowledge bases"
```

Each base keeps its own `.akb/` (config, templates, search index) and page
tree, so command-level isolation is unchanged — one invocation addresses
exactly one base:

```bash
AKB_KB="$HOME/kbs/notes" akb list
akb --kb "$HOME/kbs/incidents" lint
```

akb commits land in the shared repository, scoped to the files of the
operation.

### A KB inside a product repository

Knowledge distilled next to the code it describes, in the product's history:

```bash
cd ~/src/product
akb init kb
rm -rf kb/.git
git add -A -- kb && git commit -m "add product knowledge base"
```

akb commits now land in the product repository — `akb: write kb/notes/...`
appears in `git log` beside code commits. They are pathspec-scoped
(`git commit --only -- <paths>`), so unrelated work you have staged stays
staged and out of the akb commit, and the base's generated `.gitignore` files
keep its search index out of the product repository. They are authored under
the repository's configured git identity, or — when it has none — a per-commit
`akb <akb@local>` fallback that is never written to repository config. If
the product history must stay free of akb commits, keep the generated
`kb/.git` instead: the base is then its own repository nested in the product
directory, and akb commits stop there.

### An umbrella KB above a projects directory

Cross-project knowledge in one base, outside every product repository:

```
~/work/
├── umbrella-kb/     # base repository
└── projects/
    ├── alpha/       # product repositories
    └── beta/
```

Address it from inside any project:

```bash
cd ~/work/projects/alpha

akb --kb ../../umbrella-kb status          # per invocation
export AKB_KB="$HOME/work/umbrella-kb"     # per session

# per directory — ~/work/projects/alpha/.envrc:
#   export AKB_KB="$HOME/work/umbrella-kb"
```

`akb discover` scans the working directory and its subtree to two levels deep,
and the same for every ancestor up to your home directory, skipping hidden
directories and `.git`, `node_modules`, and `vendor` — so it finds the umbrella
base from below. Per-project bases (for example `alpha/kb`) are reported too,
nearest first:

```bash
cd ~/work/projects/alpha
akb discover
#   kb            .../projects/alpha/kb
#   umbrella-kb   .../work/umbrella-kb
```

## Multi-KB invocation pattern

There is no session state to switch between bases, so a process addresses
several bases by making N invocations, each naming its base. A harness (agent,
CI job, test suite) that owns its environment keeps a variable per base and
passes it with `--kb`; because the flag wins over an inherited `AKB_KB`, a
harness-wide default cannot leak into a call that must act on another base:

```bash
KB_NOTES=/srv/kbs/notes
KB_INCIDENTS=/srv/kbs/incidents

akb --kb "$KB_NOTES" search "deployment checklist"
akb --kb "$KB_INCIDENTS" write incidents/2026-01-14.md < incident.md
akb --kb "$KB_INCIDENTS" approve incidents/2026-01-14.md
```

Mutating commands print the base they act on (`kb: incidents
(/srv/kbs/incidents)`) to stderr, so a run log shows which base every write
touched. When `--kb` is omitted, `AKB_KB` selects the base; the selection is
resolved per invocation and never persisted.

## Where to look next

| Document | Content |
|----------|---------|
| `AGENTS.md` | project structure, conventions, and anti-patterns for contributors |
| `internal/skill/embedded/kb-management/SKILL.md` | agent-facing task router, installed with `akb skill install --location <dir> kb-management` |
| `test/AGENTS.md` | integration-test conventions |
