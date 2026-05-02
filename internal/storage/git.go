package storage

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
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

// WriteWithCommitMsg writes data to a file and commits with a custom message.
// WriteWithCommitMsg writes data and commits with a custom message.
func (g *GitProvider) WriteWithCommitMsg(_ context.Context, path string, data []byte, commitMsg string) error {
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

	if err := g.ensureGitConfig(); err != nil {
		return err
	}
	return g.gitCommit(commitMsg)
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

	if err := g.ensureGitConfig(); err != nil {
		return err
	}

	commitMsg := fmt.Sprintf("akb: delete %s", filepath.ToSlash(relPath))
	return g.gitCommit(commitMsg)
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
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.kbRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("check git status: %s: %w", strings.TrimSpace(string(out)), err)
	}
	for _, line := range strings.Split(string(out), "\n") {
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

func (g *GitProvider) ensureGitConfig() error {
	gitConfig := func(args ...string) error {
		cmd := exec.Command("git", args...) //nolint:gosec // launching trusted git binary with controlled args
		cmd.Dir = g.kbRoot
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("get git config: %w", err)
		}
		return nil
	}

	if err := gitConfig("config", "--local", "user.name"); err != nil {
		if err := gitConfig("config", "user.name", "akb"); err != nil {
			return fmt.Errorf("set git user.name: %w", err)
		}
	}

	if err := gitConfig("config", "--local", "user.email"); err != nil {
		if err := gitConfig("config", "user.email", "akb@local"); err != nil {
			return fmt.Errorf("set git user.email: %w", err)
		}
	}

	return nil
}

func (g *GitProvider) gitAdd(relPath string) error {
	cmd := exec.Command("git", "add", relPath) //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = g.kbRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add %s: %s: %w", relPath, strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (g *GitProvider) gitCommit(msg string) error {
	cmd := exec.Command("git", "commit", "-m", msg) //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = g.kbRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
