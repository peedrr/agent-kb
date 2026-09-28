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
- **Commit identity**: `ResolveIdentity()` supplies the identity of a commit in precedence order — `AKB_AUTHOR_NAME`/`AKB_AUTHOR_EMAIL`, then `git-author`/`git-email` in the base's `akb.yaml`, then git's own resolution (`GIT_AUTHOR_*` in the environment, then the merged repository and global configuration). An identity akb supplies is exported to the commit as `GIT_AUTHOR_*` and `GIT_COMMITTER_*`; when it supplies none, git resolves the commit identity itself. When no source names an identity, resolution reports `ErrNoCommitIdentity`, which the CLI maps to the usage exit code
- **Index-lock retry**: Git commands that refresh the index retry with backoff while another process holds the index lock
- **`noCommit` flag**: Global `--no-commit` skips git operations entirely
- **Commit message override**: `WriteWithCommitMsg()` for custom messages (append, approve, index, log)

## TYPICAL USAGE

```go
kbRoot, _ := path.ResolveKB(kbFlag)                 // --kb or AKB_KB selection
store, err := storage.OpenStore(kbRoot, noCommit)   // the versioning mode selects the provider
if err != nil {
    return err
}
store.Write(ctx, filepath.Join(kbRoot, "kb", "notes", "page.md"), data)   // in git mode commits `akb: write <path>`
store.Delete(ctx, filepath.Join(kbRoot, "kb", "notes", "page.md"))        // in git mode commits `akb: delete <path>`
store.Commit("akb: index add kb/notes/page.md", "kb/index.md")            // stages and commits base-relative paths as one unit
store.NothingToCommit("kb/index.md")                                      // would a commit of these paths record anything?
```

## NOTES

- `GitProvider.Write` delegates to `WriteWithCommitMsg` with auto-generated message
- Path validation happens before storage calls (in `internal/path/`)
- `FilesystemProvider` is used when git is not desired (rare; mostly for testing)
