// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/peedrr/agent-kb/internal/path"
)

// TestInitDoesNotRegisterOrSetDefault covers `akb init` creating only the
// knowledge base: no registry is written under ~/.config/agent-kb and the
// success output claims no default.
func TestInitDoesNotRegisterOrSetDefault(t *testing.T) {
	tmpDir := t.TempDir()
	home := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(home, 0750); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(akbBinPath, "init", "init-kb") //nolint:gosec // test helper launching akb binary
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("akb init failed: %s: %v", out, err)
	}

	if !strings.Contains(string(out), `Initialized KB "init-kb"`) {
		t.Errorf("output = %q, want the initialization line", out)
	}
	if strings.Contains(string(out), "set as default") {
		t.Errorf("output = %q, want no default claim", out)
	}

	registryDir := filepath.Join(home, ".config", "agent-kb")
	if _, err := os.Stat(registryDir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("registry directory %s exists after init (stat error = %v)", registryDir, err)
	}
}

// TestInitDescription covers the --description flag: when given, akb.yaml
// records it (trimmed) and `akb discover` reports it; when omitted, akb.yaml
// carries no description key at all.
func TestInitDescription(t *testing.T) {
	tmpDir := t.TempDir()

	described := exec.Command(akbBinPath, "init", "described-kb", "--description", "  project notes  ") //nolint:gosec // test helper launching akb binary
	described.Dir = tmpDir
	if out, err := described.CombinedOutput(); err != nil {
		t.Fatalf("akb init --description failed: %s: %v", out, err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "described-kb", ".agent-kb", "akb.yaml")) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "description: project notes") {
		t.Errorf("akb.yaml = %q, want trimmed description", data)
	}

	bare := exec.Command(akbBinPath, "init", "bare-kb") //nolint:gosec // test helper launching akb binary
	bare.Dir = tmpDir
	if out, err := bare.CombinedOutput(); err != nil {
		t.Fatalf("akb init failed: %s: %v", out, err)
	}

	data, err = os.ReadFile(filepath.Join(tmpDir, "bare-kb", ".agent-kb", "akb.yaml")) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "description") {
		t.Errorf("akb.yaml = %q, want no description key", data)
	}
}

// TestInitWaitsOutAnIndexLock pins that `akb init` stages and commits through
// the index-lock retry runner: a bootstrap whose `git add` meets a lock another
// process holds and releases a moment later completes and records its init commit
// instead of failing on the first attempt.
func TestInitWaitsOutAnIndexLock(t *testing.T) {
	tmpDir := t.TempDir()
	name := "locked-kb"

	// A repository skeleton whose index lock is held while init stages the new
	// base; `git init` reinitializes the directory in place.
	if err := os.MkdirAll(filepath.Join(tmpDir, name, ".git"), 0750); err != nil {
		t.Fatal(err)
	}
	indexLock := filepath.Join(tmpDir, name, ".git", "index.lock")
	if err := os.WriteFile(indexLock, nil, 0600); err != nil { //nolint:gosec // test helper creating a fixed file in its temp repository
		t.Fatal(err)
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		_ = os.Remove(indexLock) //nolint:errcheck // test releasing the index lock
	}()

	cmd := exec.Command(akbBinPath, "init", name) //nolint:gosec // test helper launching akb binary
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("akb init failed while the index lock was held: %s: %v", out, err)
	}

	logCmd := exec.Command("git", "log", "-1", "--format=%s") //nolint:gosec // test helper launching trusted git binary
	logCmd.Dir = filepath.Join(tmpDir, name)
	logOut, err := logCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("read the init commit: %s: %v", logOut, err)
	}
	if subject := strings.TrimSpace(string(logOut)); subject != "akb: init "+name {
		t.Errorf("init commit subject = %q, want %q", subject, "akb: init "+name)
	}
}

// The refusal text of an invocation inside a repository: the explanation of both
// versioning modes, and the warning of a target that is the repository root.
const (
	wantEmbedFlagHelp = "  --embed    the KB joins this repository's history; akb commits only its own paths (kb/, raw/, .agent-kb/), never other worktree changes"
	wantNoGitFlagHelp = "  --no-git   the KB is not versioned; akb excludes it from the host's git status via .git/info/exclude (local to this clone)"
	wantLayoutRefusal = "This will write .agent-kb/, kb/ and raw/ directly to the repo root. --force if desired, otherwise prefer a subdirectory: akb init <name> --embed."
)

// TestVersioningRefusalPinsMembershipAndModes pins the refusal text: the
// membership wording distinguishes a subdirectory of a repository from a target
// that is the repository root, both modes are explained, and only the root
// carries the layout warning.
func TestVersioningRefusalPinsMembershipAndModes(t *testing.T) {
	inside := versioningRefusal("kb", "/repo", false)
	wantInside := "kb is inside git repository /repo — choose how the KB is versioned:\n\n" +
		wantEmbedFlagHelp + "\n" + wantNoGitFlagHelp
	if inside != wantInside {
		t.Errorf("versioningRefusal of a subdirectory = %q, want %q", inside, wantInside)
	}

	root := versioningRefusal("repo", "/repo", true)
	wantRoot := "repo is a git repository — choose how the KB is versioned:\n\n" +
		wantEmbedFlagHelp + "\n" + wantNoGitFlagHelp + "\n\n" + wantLayoutRefusal
	if root != wantRoot {
		t.Errorf("versioningRefusal of a repository root = %q, want %q", root, wantRoot)
	}
}

// TestDetectHostRepoReportsTheEnclosingRepository covers host detection: a
// subdirectory resolves the top level of the repository it is in, and a
// directory outside every repository resolves none.
func TestDetectHostRepoReportsTheEnclosingRepository(t *testing.T) {
	repo := initTestRepo(t)
	nested := filepath.Join(repo, "docs", "nested")
	if err := os.MkdirAll(nested, 0750); err != nil {
		t.Fatal(err)
	}

	root, err := detectHostRepo(nested)
	if err != nil {
		t.Fatalf("detectHostRepo(%s): %v", nested, err)
	}
	if !sameDirectory(root, repo) {
		t.Errorf("detectHostRepo(%s) = %q, want %q", nested, root, repo)
	}

	outside := t.TempDir()
	root, err = detectHostRepo(outside)
	if err != nil {
		t.Fatalf("detectHostRepo(%s): %v", outside, err)
	}
	if root != "" {
		t.Errorf("detectHostRepo(%s) = %q, want no repository", outside, root)
	}
}

// TestNearestExistingDirFallsBackToTheParent covers the directory detection
// starts from: a target that is not there yet is judged by the nearest ancestor
// that is a directory.
func TestNearestExistingDirFallsBackToTheParent(t *testing.T) {
	parent := t.TempDir()
	missing := filepath.Join(parent, "not-there", "kb")

	if got := nearestExistingDir(missing); !sameDirectory(got, parent) {
		t.Errorf("nearestExistingDir(%s) = %q, want %q", missing, got, parent)
	}
	if got := nearestExistingDir(parent); !sameDirectory(got, parent) {
		t.Errorf("nearestExistingDir(%s) = %q, want %q", parent, got, parent)
	}
}

// TestInitInsideRepositoryRefusesWithoutAMode covers the inside-repo row of the
// matrix: the invocation exits 2 with the refusal that names the repository and
// both versioning modes, and creates nothing.
func TestInitInsideRepositoryRefusesWithoutAMode(t *testing.T) {
	repo := initTestRepo(t)
	root, err := detectHostRepo(repo)
	if err != nil {
		t.Fatalf("detectHostRepo(%s): %v", repo, err)
	}

	out, code := initRun(t, repo, "kb")
	if code != exitFault {
		t.Errorf("exit code = %d, want %d (output: %s)", code, exitFault, out)
	}
	for _, want := range []string{
		"usage: kb is inside git repository " + root + " — choose how the KB is versioned:",
		wantEmbedFlagHelp,
		wantNoGitFlagHelp,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want it to contain %q", out, want)
		}
	}

	if _, err := os.Stat(filepath.Join(repo, "kb")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the refused invocation created %s (stat error: %v)", filepath.Join(repo, "kb"), err)
	}
}

// TestInitAtRepositoryRootRefusesTheLayout covers the repo-root row of the
// matrix: without a mode the refusal explains the modes and the layout warning,
// with a mode it is the layout warning alone, and neither creates anything.
func TestInitAtRepositoryRootRefusesTheLayout(t *testing.T) {
	repo := initTestRepo(t)
	name := filepath.Base(repo)

	out, code := initRun(t, repo, ".")
	if code != exitFault {
		t.Errorf("exit code = %d, want %d (output: %s)", code, exitFault, out)
	}
	for _, want := range []string{
		"usage: " + name + " is a git repository — choose how the KB is versioned:",
		wantEmbedFlagHelp,
		wantNoGitFlagHelp,
		wantLayoutRefusal,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want it to contain %q", out, want)
		}
	}

	// A chosen mode does not get past the layout: --force is the only way.
	out, code = initRun(t, repo, ".", "--embed")
	if code != exitFault {
		t.Errorf("exit code with --embed = %d, want %d (output: %s)", code, exitFault, out)
	}
	if !strings.Contains(out, wantLayoutRefusal) {
		t.Errorf("output = %q, want the layout refusal %q", out, wantLayoutRefusal)
	}
	if strings.Contains(out, "choose how the KB is versioned") {
		t.Errorf("output = %q, want no mode refusal once the mode was chosen", out)
	}

	if _, err := os.Stat(path.StateDir(repo)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the refused invocation created %s (stat error: %v)", path.StateDir(repo), err)
	}
}

// TestInitDotAdoptsTheDirectoryName covers `akb init .` outside a repository:
// the base takes the name of the directory it initializes — akb.yaml, the init
// commit, and the success line all name it — and becomes a repository of its
// own at that directory.
func TestInitDotAdoptsTheDirectoryName(t *testing.T) {
	kbDir := filepath.Join(t.TempDir(), "adopted-kb")
	if err := os.MkdirAll(kbDir, 0750); err != nil {
		t.Fatal(err)
	}

	out, code := initRun(t, kbDir, ".")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}
	absDir, err := filepath.Abs(kbDir)
	if err != nil {
		t.Fatal(err)
	}
	if want := fmt.Sprintf("Initialized KB %q at %s", "adopted-kb", absDir); !strings.Contains(out, want) {
		t.Errorf("output = %q, want it to contain %q", out, want)
	}

	data, err := os.ReadFile(path.ConfigPath(kbDir))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "name: adopted-kb") {
		t.Errorf("akb.yaml = %q, want the adopted name", data)
	}

	logCmd := exec.Command("git", "log", "-1", "--format=%s") //nolint:gosec // test helper launching trusted git binary
	logCmd.Dir = kbDir
	logOut, err := logCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("read the init commit: %s: %v", logOut, err)
	}
	if subject := strings.TrimSpace(string(logOut)); subject != "akb: init adopted-kb" {
		t.Errorf("init commit subject = %q, want %q", subject, "akb: init adopted-kb")
	}
}

// initTestRepo creates an empty git repository: the enclosing repository of a
// scenario that initializes a KB inside one.
func initTestRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	cmd := exec.Command("git", "init", "-q", repo) //nolint:gosec // test helper launching trusted git binary
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	return repo
}

// initRun runs `akb init` in dir and returns its combined output with the exit
// code it ended on.
func initRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()

	cmd := exec.Command(akbBinPath, append([]string{"init"}, args...)...) //nolint:gosec // test helper launching akb binary
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), exitSuccess
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("akb init %v: %v: %s", args, err, out)
	}
	return string(out), exitErr.ExitCode()
}
