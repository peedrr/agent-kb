# internal/storage/

**Parent:** `./AGENTS.md`

## OVERVIEW

Storage abstraction with two implementations: git-backed (`GitProvider`) and filesystem-only (`FilesystemProvider`). All KB file I/O goes through the `Provider` interface.

## FILES

| File | Purpose |
|------|---------|
| `provider.go` | `Provider` interface: Write/Read/Delete/Exists/List |
| `git.go` | `GitProvider` — repository lock, identity-aware scoped commits, merge conflict detection, index-lock retry |
| `filesystem.go` | `FilesystemProvider` — plain file I/O without git operations |
| `storage_test.go` | Unit tests |

## KEY TYPES

| Type | Purpose |
|------|---------|
| `Provider` | Interface abstracting all file operations |
| `GitProvider` | Git-tracked writes with auto-commit and conflict blocking |
| `RepoLock` | Handle for the exclusive repository lock (`LockRepo()` / `Release()`) |
| `FilesystemProvider` | Direct filesystem I/O (no git integration) |

## GIT PROVIDER BEHAVIOR

- **Repository lock**: `LockRepo()` takes an exclusive `flock` on the repository that hosts the KB, resolved through `git rev-parse --git-common-dir` so linked worktrees share it; a command holds it across its whole mutation, and a nested acquisition of the same repository reuses the held lock
- **Auto-commit**: `akb: write <path>` / `akb: delete <path>` messages, recording only the paths the operation touched
- **Merge conflict blocking**: Checks `git status --porcelain` for `U` (unmerged) before any write/delete
- **Commit identity**: `CommitIdentityArgs()` passes `-c user.name=akb -c user.email=akb@local` only for a repository without `user.name`; a configured identity is used as is. Repository config is never written, except by `akb init` (`ensureGitConfig`)
- **Index-lock retry**: Git commands that refresh the index retry with backoff while another process holds the index lock
- **`noCommit` flag**: Global `--no-commit` skips git operations entirely
- **Commit message override**: `WriteWithCommitMsg()` for custom messages (append, approve, index, log)

## TYPICAL USAGE

```go
kbRoot, _ := path.ResolveKB(kbFlag)   // --kb or AKB_KB selection
store := storage.NewGitProvider(kbRoot, noCommit)
store.Write(ctx, path, data)
```

## NOTES

- `GitProvider.Write` delegates to `WriteWithCommitMsg` with auto-generated message
- Path validation happens before storage calls (in `internal/path/`)
- `FilesystemProvider` is used when git is not desired (rare; mostly for testing)
