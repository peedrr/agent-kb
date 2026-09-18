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
	"time"

	"golang.org/x/sys/unix"
)

// managedCommitPaths are KB files that commands restage alongside the page they
// change: `akb delete` restages the index and the log before committing. They
// join a commit's pathspec when the current operation staged them.
var managedCommitPaths = []string{"kb/index.md", "kb/log.md"}

// gitCommitIdentity is applied to every commit invocation so akb commits carry a
// stable identity without writing user.name/user.email into repository config.
var gitCommitIdentity = []string{"-c", "user.name=akb", "-c", "user.email=akb@local"}

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

		return g.gitCommit(commitMsg, relPath)
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
		return g.gitCommit(commitMsg, relPath)
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

func (g *GitProvider) checkMergeConflicts() error {
	out, err := g.runGit("check git status", "status", "--porcelain")
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
	lockPath, err := g.repoLockPath()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600) //nolint:gosec // fixed lock file name inside the git dir
	if err != nil {
		return fmt.Errorf("open repo lock %s: %w", lockPath, err)
	}
	defer f.Close() //nolint:errcheck // closing the handle also releases the lock

	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil { //nolint:gosec // file descriptors fit in an int
		return fmt.Errorf("lock repo %s: %w", lockPath, err)
	}
	defer unix.Flock(int(f.Fd()), unix.LOCK_UN) //nolint:errcheck,gosec // released with the handle as well; file descriptors fit in an int

	return fn()
}

// repoLockPath returns the lock file shared by all KBs of the git repository the
// KB belongs to. It resolves --git-common-dir rather than --git-dir so linked
// worktrees lock the repository they share with the main checkout.
func (g *GitProvider) repoLockPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir") //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = g.kbRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve git common dir: %s: %w", strings.TrimSpace(string(out)), err)
	}

	commonDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(g.kbRoot, commonDir)
	}
	return filepath.Join(commonDir, "akb.lock"), nil
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
	includeTarget, err := g.isStaged(target)
	if err != nil {
		return nil, err
	}
	if !includeTarget {
		includeTarget, err = g.isInHEAD(target)
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
		staged, err := g.isStaged(managed)
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
func (g *GitProvider) isStaged(relPath string) (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet", "--", relPath) //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = g.kbRoot
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

// isInHEAD reports whether relPath exists in the committed tree.
func (g *GitProvider) isInHEAD(relPath string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--quiet", "--verify", "HEAD:"+relPath) //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = g.kbRoot
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
	_, err := g.runGit("git add "+relPath, "add", relPath)
	return err
}

func (g *GitProvider) gitCommit(msg string, relPath string) error {
	paths, err := g.commitPaths(relPath)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		// The operation recorded nothing git can commit; report that instead of
		// committing changes staged by something else.
		return fmt.Errorf("git commit: nothing to commit for %s", filepath.ToSlash(relPath))
	}

	args := append([]string{}, gitCommitIdentity...)
	args = append(args, "commit", "-m", msg, "--only", "--")
	args = append(args, paths...)

	_, err = g.runGit("git commit", args...)
	return err
}

// runGit runs a git subcommand in the KB root and returns its combined output.
// A command that writes or refreshes the repository index fails while another
// process holds the index lock, so contention is retried with backoff; the
// holder — akb working on another KB, or any other tool — usually releases the
// lock within milliseconds.
func (g *GitProvider) runGit(op string, args ...string) (string, error) {
	for attempt := 0; ; attempt++ {
		cmd := exec.Command("git", args...) //nolint:gosec // launching trusted git binary with controlled args
		cmd.Dir = g.kbRoot
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
