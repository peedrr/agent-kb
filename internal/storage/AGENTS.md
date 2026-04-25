# internal/storage/

**Parent:** `./AGENTS.md`

## OVERVIEW

Storage abstraction with two implementations: git-backed (`GitProvider`) and filesystem-only (`FilesystemProvider`). All KB file I/O goes through the `Provider` interface.

## FILES

| File | Purpose |
|------|---------|
| `provider.go` | `Provider` interface: Write/Read/Delete/Exists/List |
| `git.go` | `GitProvider` — auto-commit, merge conflict detection, git config auto-set |
| `filesystem.go` | `FilesystemProvider` — plain file I/O without git operations |
| `storage_test.go` | Unit tests |

## KEY TYPES

| Type | Purpose |
|------|---------|
| `Provider` | Interface abstracting all file operations |
| `GitProvider` | Git-tracked writes with auto-commit and conflict blocking |
| `FilesystemProvider` | Direct filesystem I/O (no git integration) |

## GIT PROVIDER BEHAVIOR

- **Auto-commit**: `akb: write <path>` / `akb: delete <path>` messages
- **Merge conflict blocking**: Checks `git status --porcelain` for `U` (unmerged) before any write/delete
- **Git config auto-set**: Sets `user.name=akb` / `user.email=akb@local` if unset
- **`noCommit` flag**: Global `--no-commit` skips git operations entirely
- **Commit message override**: `WriteWithCommitMsg()` for custom messages (used by `akb index rebuild`)

## TYPICAL USAGE

```go
kbRoot, _ := path.KBRoot()
store := storage.NewGitProvider(kbRoot, noCommit)
store.Write(ctx, path, data)
```

## NOTES

- `GitProvider.Write` delegates to `WriteWithCommitMsg` with auto-generated message
- Path validation happens before storage calls (in `internal/path/`)
- `FilesystemProvider` is used when git is not desired (rare; mostly for testing)
