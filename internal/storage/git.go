// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// managedCommitPaths are KB files that commands restage alongside the page they
// change: `akb delete` restages the index and the log before committing. They
// join a commit's pathspec when the current operation staged them.
var managedCommitPaths = []string{"kb/index.md", "kb/log.md"}

const (
	// indexLockRetryAttempts bounds the retries used when another process holds
	// the repository index lock.
	indexLockRetryAttempts = 12
	// indexLockRetryDelay is the first backoff step; every retry doubles it, up
	// to indexLockRetryMaxDelay.
	indexLockRetryDelay    = 10 * time.Millisecond
	indexLockRetryMaxDelay = 250 * time.Millisecond
)

// GitProvider performs file operations with git tracking.
type GitProvider struct {
	kbRoot   string
	noCommit bool
}

// NewGitProvider creates a git-tracking storage provider.
func NewGitProvider(kbRoot string, noCommit bool) *GitProvider {
	return &GitProvider{kbRoot: kbRoot, noCommit: noCommit}
}

// WriteWithCommitMsg writes data and commits with a custom message.
func (g *GitProvider) WriteWithCommitMsg(_ context.Context, path string, data []byte, commitMsg string) error {
	return g.withRepoLock(func() error {
		if err := g.checkMergeConflicts(); err != nil {
			return err
		}

		identity, err := g.commitIdentity()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
			return fmt.Errorf("create parent directories: %w", err)
		}

		if err := os.WriteFile(path, data, 0600); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		relPath, err := filepath.Rel(g.kbRoot, path)
		if err != nil {
			return fmt.Errorf("resolve relative path: %w", err)
		}

		if err := g.gitAdd(relPath); err != nil {
			return err
		}

		if g.noCommit {
			return nil
		}

		return g.gitCommit(commitMsg, relPath, identity)
	})
}

func (g *GitProvider) Write(ctx context.Context, path string, data []byte) error {
	relPath, err := filepath.Rel(g.kbRoot, path)
	if err != nil {
		return fmt.Errorf("resolve relative path: %w", err)
	}
	commitMsg := fmt.Sprintf("akb: write %s", filepath.ToSlash(relPath))
	return g.WriteWithCommitMsg(ctx, path, data, commitMsg)
}

func (g *GitProvider) Read(_ context.Context, path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(g.kbRoot, path)
	}
	data, err := os.ReadFile(path) //nolint:gosec // path validated by provider
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

// Delete removes a file and commits the change.
func (g *GitProvider) Delete(_ context.Context, path string) error {
	return g.withRepoLock(func() error {
		if err := g.checkMergeConflicts(); err != nil {
			return err
		}

		identity, err := g.commitIdentity()
		if err != nil {
			return err
		}

		if err := os.Remove(path); err != nil {
			return fmt.Errorf("delete file: %w", err)
		}

		relPath, err := filepath.Rel(g.kbRoot, path)
		if err != nil {
			return fmt.Errorf("resolve relative path: %w", err)
		}

		if err := g.gitAdd(relPath); err != nil {
			return err
		}

		if g.noCommit {
			return nil
		}

		commitMsg := fmt.Sprintf("akb: delete %s", filepath.ToSlash(relPath))
		return g.gitCommit(commitMsg, relPath, identity)
	})
}

// Exists checks whether a file exists.
func (g *GitProvider) Exists(_ context.Context, path string) (bool, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(g.kbRoot, path)
	}
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("check file existence: %w", err)
}

// List returns files with the given extension under a directory.
func (g *GitProvider) List(_ context.Context, dir string, ext string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ext != "" && !strings.HasSuffix(path, ext) {
			return nil
		}
		relPath, err := filepath.Rel(g.kbRoot, path)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}
		files = append(files, filepath.ToSlash(relPath))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

// commitIdentity resolves the identity of the commit that ends an operation. It
// runs before the operation writes anything, so an identity no source can name
// fails the operation without leaving a half-written update behind. One that
// leaves its commit to the caller (--no-commit) records no identity of its own.
func (g *GitProvider) commitIdentity() (*Identity, error) {
	if g.noCommit {
		return nil, nil
	}
	return ResolveIdentity(g.kbRoot)
}

// checkMergeConflicts rejects a mutation while the repository that hosts the KB
// carries a merge that is not finished. It is the provider's side of the
// package-level merge preflight, which every commit path runs.
func (g *GitProvider) checkMergeConflicts() error {
	return checkMergeState(g.kbRoot)
}

// checkMergeState rejects a mutation while the repository of kbRoot carries a
// merge that is not finished: a conflicted merge, or a clean one that is still
// in progress. Git refuses a partial commit — the only kind akb makes — until
// the merge is committed or aborted, so a mutation must not start.
func checkMergeState(kbRoot string) error {
	out, err := RunGit(kbRoot, "check git status", "status", "--porcelain")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(out, "\n") {
		if len(line) >= 3 {
			x := line[0]
			y := line[1]
			if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
				return fmt.Errorf("merge conflict detected: resolve conflicts before proceeding (file: %s)", strings.TrimSpace(line[2:]))
			}
		}
	}
	return nil
}

// withRepoLock runs fn while holding an exclusive lock on the git repository
// that hosts the KB. The lock is keyed by the repository, so every KB inside it
// — including KBs in linked worktrees, which share a common dir — serializes its
// file write, staging and commit steps against other processes running akb on
// the same repository.
func (g *GitProvider) withRepoLock(fn func() error) error {
	lock, err := LockRepo(g.kbRoot)
	if err != nil {
		return err
	}
	defer lock.Release()

	return fn()
}

// RepoLock is a handle on the exclusive lock the process holds on one lock
// file: the repository lock of a git-versioned base, or the lock file of a base
// that is not versioned in git. Releasing the handle drops the one acquisition
// it represents and is idempotent, so a release run twice on the same handle
// cannot drop an acquisition another holder of the process holds. The process
// keeps the flock until the outermost acquisition is released.
type RepoLock struct {
	path     string
	released bool
}

// lockEntry is the lock the process holds on one lock file: the open file
// carrying the flock and the number of acquisitions sharing it.
type lockEntry struct {
	file *os.File
	refs int
}

// heldLocks tracks the lock files the process holds. A flock is keyed by the
// open file description, so a second flock on a fresh handle of the same file
// would block against the process's own lock; nested acquisitions reuse the
// held lock instead.
var (
	heldLocksMu sync.Mutex
	heldLocks   = make(map[string]*lockEntry)
)

// LockRepo acquires the exclusive lock on the git repository that hosts the KB
// and returns its handle. A command holds that lock across its whole mutation
// sequence — the page or raw file write, the index and log updates, the commit,
// and the search and link-graph updates that follow them — so concurrent akb
// processes serialize instead of losing each other's updates. An acquisition
// nested inside an already held lock of the same repository shares it, which
// lets a command hold the lock while the provider commits its write.
func LockRepo(kbRoot string) (*RepoLock, error) {
	lockPath, err := repoLockPath(kbRoot)
	if err != nil {
		return nil, err
	}
	return lockFile(lockPath)
}

// lockFile acquires the exclusive lock carried by lockPath and returns its
// handle. Every lock of the process goes through it, so acquisitions of the same
// lock file nested inside each other share the lock instead of blocking on the
// process's own flock.
func lockFile(lockPath string) (*RepoLock, error) {
	heldLocksMu.Lock()
	defer heldLocksMu.Unlock()

	if entry, held := heldLocks[lockPath]; held {
		entry.refs++
		return &RepoLock{path: lockPath}, nil
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600) //nolint:gosec // fixed lock file name inside the git dir
	if err != nil {
		return nil, fmt.Errorf("open repo lock %s: %w", lockPath, err)
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil { //nolint:gosec // file descriptors fit in an int
		_ = f.Close() //nolint:errcheck // a failed lock leaves nothing to release
		return nil, fmt.Errorf("lock repo %s: %w", lockPath, err)
	}

	heldLocks[lockPath] = &lockEntry{file: f, refs: 1}
	return &RepoLock{path: lockPath}, nil
}

// Release drops one acquisition of the repository lock. A repeat release of the
// same handle, like a release of a nil handle, is a no-op: the process holds the
// flock until every acquisition that was actually taken is released. The flock
// is released with the outermost acquisition; flock also releases it when the
// process exits.
func (l *RepoLock) Release() {
	if l == nil {
		return
	}

	heldLocksMu.Lock()
	defer heldLocksMu.Unlock()

	if l.released {
		return
	}
	l.released = true

	entry, held := heldLocks[l.path]
	if !held {
		return
	}
	entry.refs--
	if entry.refs > 0 {
		return
	}

	delete(heldLocks, l.path)
	_ = unix.Flock(int(entry.file.Fd()), unix.LOCK_UN) //nolint:errcheck,gosec // released with the handle as well; file descriptors fit in an int
	_ = entry.file.Close()                             //nolint:errcheck // releasing the handle drops the lock too
}

// repoLockPath returns the lock file shared by all KBs of the git repository the
// KB belongs to. It resolves --git-common-dir rather than --git-dir so linked
// worktrees lock the repository they share with the main checkout.
func repoLockPath(kbRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir") //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = kbRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve git common dir: %s: %w", strings.TrimSpace(string(out)), err)
	}

	commonDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(kbRoot, commonDir)
	}
	return filepath.Join(commonDir, "akb.lock"), nil
}

// CommitFiles stages and commits exactly the given KB-relative paths with
// commitMsg while holding the repository lock. The pathspec keeps the commit
// scoped to the operation's own files, so changes another tool staged stay
// staged. The commit carries the identity akb supplies — AKB_AUTHOR_NAME and
// AKB_AUTHOR_EMAIL, or git-author and git-email in the base's akb.yaml — and
// git's own identity when akb supplies none. Paths git cannot record — absent
// from both the worktree and the committed tree — are skipped.
func CommitFiles(kbRoot, commitMsg string, paths ...string) error {
	lock, err := LockRepo(kbRoot)
	if err != nil {
		return err
	}
	defer lock.Release()

	if err := checkMergeState(kbRoot); err != nil {
		return err
	}

	identity, err := ResolveIdentity(kbRoot)
	if err != nil {
		return err
	}

	recorded, err := recordablePaths(kbRoot, paths)
	if err != nil {
		return err
	}
	if len(recorded) == 0 {
		return fmt.Errorf("git commit: nothing to commit")
	}

	if err := StageFiles(kbRoot, recorded...); err != nil {
		return err
	}

	args, env := commitArgs(identity, commitMsg, recorded)
	_, err = runGitEnv(kbRoot, "git commit", env, args...)
	return err
}

// NothingToCommit reports whether committing the given KB-relative paths would
// record nothing: every path matches the committed tree, with no staged change,
// no worktree change and no untracked file. It lets a command that would write
// exactly what is already committed treat the operation as a no-op instead of
// reaching git with an empty change set.
func NothingToCommit(kbRoot string, paths ...string) (bool, error) {
	args := append([]string{"status", "--porcelain", "--"}, paths...)
	out, err := RunGit(kbRoot, "git status", args...)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// recordablePaths keeps the paths git can record: the ones present in the
// worktree, plus the ones recorded in the committed tree, which an operation
// may have deleted.
func recordablePaths(kbRoot string, paths []string) ([]string, error) {
	recorded := make([]string, 0, len(paths))
	for _, path := range paths {
		relPath := filepath.ToSlash(path)
		if _, err := os.Stat(filepath.Join(kbRoot, filepath.FromSlash(relPath))); err == nil {
			recorded = append(recorded, relPath)
			continue
		}

		inHEAD, err := isInHEAD(kbRoot, relPath)
		if err != nil {
			return nil, err
		}
		if inHEAD {
			recorded = append(recorded, relPath)
		}
	}
	return recorded, nil
}

// StageFiles stages the given KB-relative paths, including the deletions of
// paths that are gone from the worktree. Like the other git writes, it retries
// while another process holds the repository index lock.
func StageFiles(kbRoot string, paths ...string) error {
	args := append([]string{"add", "-A", "--"}, paths...)
	_, err := RunGit(kbRoot, "git add", args...)
	return err
}

// commitPaths returns the pathspec of the commit for the operation that wrote or
// deleted relPath: the file itself when git records it, plus the KB-managed files
// the operation staged, such as the index and the log. Committing with an
// explicit pathspec keeps akb from sweeping unrelated staged changes of the
// enclosing repository into its commit.
func (g *GitProvider) commitPaths(relPath string) ([]string, error) {
	target := filepath.ToSlash(relPath)
	var paths []string

	// Include the target when the operation staged it, or when it exists in the
	// committed tree. A page git has never recorded — staged, then deleted
	// before its first commit — has nothing to contribute and is rejected by git.
	includeTarget, err := isStaged(g.kbRoot, target)
	if err != nil {
		return nil, err
	}
	if !includeTarget {
		includeTarget, err = isInHEAD(g.kbRoot, target)
		if err != nil {
			return nil, err
		}
	}
	if includeTarget {
		paths = append(paths, target)
	}

	for _, managed := range managedCommitPaths {
		if managed == target {
			continue
		}
		staged, err := isStaged(g.kbRoot, managed)
		if err != nil {
			return nil, err
		}
		if staged {
			paths = append(paths, managed)
		}
	}

	return paths, nil
}

// isStaged reports whether relPath has staged changes in the git index.
func isStaged(kbRoot, relPath string) (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet", "--", relPath) //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = kbRoot
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("inspect staged state of %s: %w", relPath, err)
}

// isInHEAD reports whether relPath exists in the committed tree. The path is
// resolved relative to the KB root, which may be a subdirectory of the
// repository.
func isInHEAD(kbRoot, relPath string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--quiet", "--verify", "HEAD:./"+relPath) //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = kbRoot
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, fmt.Errorf("inspect committed tree for %s: %w", relPath, err)
}

func (g *GitProvider) gitAdd(relPath string) error {
	_, err := RunGit(g.kbRoot, "git add "+relPath, "add", relPath)
	return err
}

func (g *GitProvider) gitCommit(msg string, relPath string, identity *Identity) error {
	paths, err := g.commitPaths(relPath)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		// The operation recorded nothing git can commit; report that instead of
		// committing changes staged by something else.
		return fmt.Errorf("git commit: nothing to commit for %s", filepath.ToSlash(relPath))
	}

	args, env := commitArgs(identity, msg, paths)
	_, err = runGitEnv(g.kbRoot, "git commit", env, args...)
	return err
}

// commitArgs builds the arguments and the environment of a commit that records
// exactly paths with msg. The identity akb supplies reaches git as
// GIT_AUTHOR_* and GIT_COMMITTER_* overrides; when akb supplies none, git
// resolves the identity from its own configuration.
func commitArgs(identity *Identity, msg string, paths []string) (args []string, env []string) {
	args = append(args, "commit", "-m", msg, "--only", "--")
	return append(args, paths...), commitEnv(identity)
}

// RunGit runs a git subcommand in kbRoot and returns its combined output. A
// command that writes or refreshes the repository index fails while another
// process holds the index lock, so contention is retried with backoff; the
// holder — akb working on another KB, or any other tool — usually releases the
// lock within milliseconds. It is the runner the provider uses, and the one a
// command outside this package uses for the git steps that package does not
// wrap (the bootstrap staging and commit of `akb init`).
func RunGit(kbRoot, op string, args ...string) (string, error) {
	return runGitEnv(kbRoot, op, nil, args...)
}

// runGitEnv is RunGit with the environment of the git process spelled out: nil
// inherits akb's environment, which is what a commit that akb supplies no
// identity for needs — git resolves the identity from it.
func runGitEnv(kbRoot, op string, env []string, args ...string) (string, error) {
	for attempt := 0; ; attempt++ {
		cmd := exec.Command("git", args...) //nolint:gosec // launching trusted git binary with controlled args
		cmd.Dir = kbRoot
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err == nil {
			return string(out), nil
		}
		if attempt >= indexLockRetryAttempts-1 || !isIndexLockContention(out) {
			return "", fmt.Errorf("%s: %s: %w", op, strings.TrimSpace(string(out)), err)
		}
		time.Sleep(indexLockBackoff(attempt))
	}
}

// isIndexLockContention reports whether git refused to run because another
// process holds the repository index lock.
func isIndexLockContention(out []byte) bool {
	return bytes.Contains(out, []byte("index.lock"))
}

func indexLockBackoff(attempt int) time.Duration {
	delay := indexLockRetryDelay << attempt
	if delay <= 0 || delay > indexLockRetryMaxDelay {
		return indexLockRetryMaxDelay
	}
	return delay
}
