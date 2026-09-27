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
