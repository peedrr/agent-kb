// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/peedrr/agent-kb/internal/path"
)

// detectHostRepo returns the top level of the git repository that contains dir,
// or "" when dir is not inside one. dir has to exist: git resolves the
// repository of the directory it runs in, and the directory an init invocation
// creates belongs to the repository of its parent.
func detectHostRepo(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel") //nolint:gosec // launching trusted git binary with controlled args
	out, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(out, []byte("not a git repository")) {
			return "", nil
		}
		return "", fmt.Errorf("resolve repository of %s: %s: %w", dir, strings.TrimSpace(string(out)), err)
	}

	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("resolve repository of %s: git reported no top level", dir)
	}
	return root, nil
}

// nearestExistingDir returns dir when a directory occupies it, and the nearest
// ancestor that is a directory otherwise. Git cannot enter a directory that is
// not there yet or that a file occupies, so detection starts from the closest
// directory that exists.
func nearestExistingDir(dir string) string {
	for {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

// sameDirectory reports whether both paths name the same existing directory.
// The repository root git reports is resolved through symlinks, so the paths are
// compared by the directory they point at rather than by their spelling.
func sameDirectory(a, b string) bool {
	infoA, errA := os.Stat(a)
	infoB, errB := os.Stat(b)
	return errA == nil && errB == nil && infoA.IsDir() && infoB.IsDir() && os.SameFile(infoA, infoB)
}

// excludeFromHostRepo appends the new base to the exclude file of the host
// repository — the file git reads for this clone alone — and returns the
// patterns it appended, in the form the invocation reports them, along with how
// many there are, so the invocation's undo guidance can refer to them in the
// singular or the plural. It is the single write an init invocation makes to a
// repository the base does not own: the base stays out of the host's git status
// without the host tracking an ignore rule for it.
func excludeFromHostRepo(target, hostRoot string) (string, int, error) {
	entries, err := hostExcludeEntries(target, hostRoot)
	if err != nil {
		return "", 0, err
	}

	excludePath, err := hostExcludePath(target)
	if err != nil {
		return "", 0, err
	}
	if err := appendMissingLines(excludePath, entries...); err != nil {
		return "", 0, fmt.Errorf("append to %s: %w", excludePath, err)
	}

	return strings.Join(entries, ", "), len(entries), nil
}

// hostExcludeEntries returns the patterns that keep the new base out of the host
// repository's status. The target is resolved through symlinks first, because
// the repository root git reports is resolved too: a target reached through a
// symlink would otherwise render an entry relative to the repository and
// outside it at once, which git reads as a pattern that matches nothing. A base
// in a subdirectory is excluded as that directory. A base at the repository
// root is excluded by the directories it owns, because excluding the directory
// the host's own files live in would hide them too.
func hostExcludeEntries(target, hostRoot string) ([]string, error) {
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", target, err)
	}
	rel, err := filepath.Rel(hostRoot, resolved)
	if err != nil {
		return nil, fmt.Errorf("locate %s in %s: %w", target, hostRoot, err)
	}
	if rel == "." {
		return []string{path.StateDirName + "/", "kb/", "raw/"}, nil
	}
	return []string{filepath.ToSlash(rel) + "/"}, nil
}

// hostExcludePath returns the exclude file of the repository dir belongs to.
// `git rev-parse --git-path` resolves it against the directory git runs in and
// names the shared file of a linked worktree, which is the file git reads.
func hostExcludePath(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-path", "info/exclude") //nolint:gosec // launching trusted git binary with controlled args
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve exclude file of %s: %s: %w", dir, strings.TrimSpace(string(out)), err)
	}

	excludePath := strings.TrimSpace(string(out))
	if excludePath == "" {
		return "", fmt.Errorf("resolve exclude file of %s: git reported none", dir)
	}
	if !filepath.IsAbs(excludePath) {
		excludePath = filepath.Join(dir, excludePath)
	}
	return excludePath, nil
}
