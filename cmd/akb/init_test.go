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

// setRepoIdentity gives the repository the identity its commits are made under,
// so the scenario does not depend on the machine's git configuration.
func setRepoIdentity(t *testing.T, repo, name, email string) {
	t.Helper()

	gitRun(t, repo, "config", "user.name", name)
	gitRun(t, repo, "config", "user.email", email)
}

// gitRun runs a git command in dir and fails the test when it does not succeed.
func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // test helper launching trusted git binary
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %s: %v", args, out, err)
	}
}

// gitOutput returns the trimmed output of a git command run in dir.
func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // test helper launching trusted git binary
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %s: %v", args, out, err)
	}
	return strings.TrimSpace(string(out))
}

// initRun runs `akb init` in dir and returns its combined output with the exit
// code it ended on.
func initRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()

	return initRunEnv(t, dir, nil, args...)
}

// initRunEnv runs `akb init` with an explicit environment: a scenario that
// controls what git resolves as the commit identity sets one.
func initRunEnv(t *testing.T, dir string, env []string, args ...string) (string, int) {
	t.Helper()

	cmd := exec.Command(akbBinPath, append([]string{"init"}, args...)...) //nolint:gosec // test helper launching akb binary
	cmd.Dir = dir
	cmd.Env = env
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

// TestInitEmbedCommitsOnlyTheBasesOwnPaths covers embedded mode: the base is
// committed into the host repository, the commit records exactly the base's own
// paths, and changes another tool staged in the host worktree stay staged and
// uncommitted.
func TestInitEmbedCommitsOnlyTheBasesOwnPaths(t *testing.T) {
	repo := initTestRepo(t)
	setRepoIdentity(t, repo, "ada", "ada@example.com")

	hostFile := filepath.Join(repo, "src", "app.go")
	if err := os.MkdirAll(filepath.Dir(hostFile), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hostFile, []byte("package main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", "src/app.go")
	gitRun(t, repo, "commit", "-m", "host work")

	stagedFile := filepath.Join(repo, "src", "staged.go")
	if err := os.WriteFile(stagedFile, []byte("package main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", "src/staged.go")

	out, code := initRun(t, repo, "docs", "--embed")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}
	for _, want := range []string{
		fmt.Sprintf("Initialized KB %q at %s", "docs", filepath.Join(repo, "docs")),
		"kb: docs is embedded inside repository " + repo + "; akb commits only kb/, raw/, .agent-kb/",
		"commit identity: from git config (ada <ada@example.com>) — not recorded; each machine's git identity applies",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want it to contain %q", out, want)
		}
	}

	data, err := os.ReadFile(path.ConfigPath(filepath.Join(repo, "docs")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "versioning: git") {
		t.Errorf("akb.yaml = %q, want the mode recorded explicitly", data)
	}
	if strings.Contains(string(data), "git-author") {
		t.Errorf("akb.yaml = %q, want no identity the machine owns", data)
	}

	// The init commit records exactly the base's own paths.
	wantPaths := []string{
		"docs/.agent-kb/.gitignore",
		"docs/.agent-kb/akb.yaml",
		"docs/.gitignore",
		"docs/kb/index.md",
		"docs/kb/log.md",
		"docs/raw/files.log",
	}
	if recorded := gitOutput(t, repo, "show", "--name-only", "--format=", "HEAD"); recorded != strings.Join(wantPaths, "\n") {
		t.Errorf("init commit records %q, want %q", recorded, strings.Join(wantPaths, "\n"))
	}

	// The unrelated staged change is still staged and uncommitted.
	if staged := gitOutput(t, repo, "diff", "--cached", "--name-only"); staged != "src/staged.go" {
		t.Errorf("staged files = %q, want the host's own change untouched", staged)
	}
}

// TestInitEmbedOutsideARepositoryRefuses covers `--embed` with no repository to
// join: there is no history for the base to become part of.
func TestInitEmbedOutsideARepositoryRefuses(t *testing.T) {
	dir := t.TempDir()

	out, code := initRun(t, dir, "kb", "--embed")
	if code != exitFault {
		t.Errorf("exit code = %d, want %d (output: %s)", code, exitFault, out)
	}
	if want := "usage: --embed: kb is not inside a git repository — omit --embed to create a standalone repository"; !strings.Contains(out, want) {
		t.Errorf("output = %q, want it to contain %q", out, want)
	}
	if _, err := os.Stat(path.StateDir(filepath.Join(dir, "kb"))); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the refused invocation created a base (stat error: %v)", err)
	}
}

// TestInitNoGitExcludesTheBaseFromTheHost covers non-git mode inside a
// repository: the base is unversioned, it is appended to the host's local
// exclude file, and the host's git status stays clear of it.
func TestInitNoGitExcludesTheBaseFromTheHost(t *testing.T) {
	repo := initTestRepo(t)

	out, code := initRun(t, repo, "docs", "--no-git")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}
	for _, want := range []string{
		fmt.Sprintf("Initialized KB %q at %s", "docs", filepath.Join(repo, "docs")),
		"kb: excluded docs/ from host git tracking via .git/info/exclude (local to this clone; remove that line to undo)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want it to contain %q", out, want)
		}
	}

	data, err := os.ReadFile(path.ConfigPath(filepath.Join(repo, "docs")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "versioning: none") {
		t.Errorf("akb.yaml = %q, want the mode recorded explicitly", data)
	}

	exclude, err := os.ReadFile(filepath.Join(repo, ".git", "info", "exclude")) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(exclude), "docs/") {
		t.Errorf("the host exclude file = %q, want the base's directory", exclude)
	}

	if status := gitOutput(t, repo, "status", "--porcelain"); strings.Contains(status, "docs") {
		t.Errorf("host status = %q, want the excluded base out of it", status)
	}

	if _, err := os.Stat(filepath.Join(repo, "docs", ".git")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the unversioned base is a repository of its own (stat error: %v)", err)
	}
}

// TestHostExcludeEntriesResolvesASymlinkedTarget covers an init invocation whose
// target is reached through a symlink: git reports the resolved repository root,
// so the exclude entry has to name the resolved directory — the lexical
// '../<link>/docs/' pattern would match nothing — while a target at the
// repository root still yields the directories the base owns.
func TestHostExcludeEntriesResolvesASymlinkedTarget(t *testing.T) {
	repo := initTestRepo(t)
	hostRoot, err := detectHostRepo(repo)
	if err != nil {
		t.Fatalf("detectHostRepo(%s): %v", repo, err)
	}

	if err := os.MkdirAll(filepath.Join(repo, "docs"), 0750); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatal(err)
	}
	lexicalTarget := filepath.Join(link, "docs")

	entries, err := hostExcludeEntries(lexicalTarget, hostRoot)
	if err != nil {
		t.Fatalf("hostExcludeEntries(%s): %v", lexicalTarget, err)
	}
	if got := strings.Join(entries, ", "); got != "docs/" {
		t.Errorf("hostExcludeEntries(%s) = %q, want %q", lexicalTarget, got, "docs/")
	}

	excluded, count, err := excludeFromHostRepo(lexicalTarget, hostRoot)
	if err != nil {
		t.Fatalf("excludeFromHostRepo(%s): %v", lexicalTarget, err)
	}
	if excluded != "docs/" {
		t.Errorf("excludeFromHostRepo(%s) = %q, want %q", lexicalTarget, excluded, "docs/")
	}
	if count != 1 {
		t.Errorf("excludeFromHostRepo(%s) appended %d entries, want 1", lexicalTarget, count)
	}

	excludeData, err := os.ReadFile(filepath.Join(repo, ".git", "info", "exclude")) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(excludeData), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line != "docs/" {
			t.Errorf("host exclude file carries pattern %q, want only the resolved base directory", line)
		}
	}

	// A target at the repository root keeps yielding the directories the base
	// owns instead of the root itself, each anchored to the root so the pattern
	// matches those directories there alone.
	entries, err = hostExcludeEntries(repo, hostRoot)
	if err != nil {
		t.Fatalf("hostExcludeEntries(%s): %v", repo, err)
	}
	if got, want := strings.Join(entries, ", "), "/.agent-kb/, /kb/, /raw/"; got != want {
		t.Errorf("hostExcludeEntries at the repository root = %q, want %q", got, want)
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry, "/") {
			t.Errorf("hostExcludeEntries at the repository root yielded %q, want a pattern anchored to the root", entry)
		}
	}
}

// TestInitNoGitOutsideARepository covers non-git mode with no host repository:
// the base is unversioned and nothing is created around it.
func TestInitNoGitOutsideARepository(t *testing.T) {
	dir := t.TempDir()

	out, code := initRun(t, dir, "scratch", "--no-git")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}

	data, err := os.ReadFile(path.ConfigPath(filepath.Join(dir, "scratch")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "versioning: none") {
		t.Errorf("akb.yaml = %q, want the mode recorded explicitly", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "scratch", ".git")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the unversioned base is a repository of its own (stat error: %v)", err)
	}
	if strings.Contains(out, "info/exclude") {
		t.Errorf("output = %q, want no exclusion without a host repository", out)
	}
}

// TestInitForceBypassesTheLayoutOnly covers --force: it gets past the repo-root
// layout warning and chooses no versioning mode of its own.
func TestInitForceBypassesTheLayoutOnly(t *testing.T) {
	repo := initTestRepo(t)
	setRepoIdentity(t, repo, "ada", "ada@example.com")

	out, code := initRun(t, repo, ".", "--force")
	if code != exitFault {
		t.Errorf("exit code with --force alone = %d, want %d (output: %s)", code, exitFault, out)
	}
	if !strings.Contains(out, "choose how the KB is versioned") {
		t.Errorf("output = %q, want the versioning refusal: --force implies no mode", out)
	}

	out, code = initRun(t, repo, ".", "--embed", "--force")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}
	if want := fmt.Sprintf("Initialized KB %q at %s", filepath.Base(repo), repo); !strings.Contains(out, want) {
		t.Errorf("output = %q, want it to contain %q", out, want)
	}
	if _, err := os.Stat(path.ConfigPath(repo)); err != nil {
		t.Errorf("the base at the repository root: %v", err)
	}
}

// TestInitNoGitAtRepositoryRootKeepsTheHostGitignore covers the unversioned base
// at its host repository's root: the base is never versioned, so the root
// .gitignore has no consumer, and init leaves the host's committed file
// byte-identical and the host's status clean while still writing the base's own
// .agent-kb/.gitignore and the exclude entries that keep the base invisible,
// whose notice tells the caller to remove all three of them.
func TestInitNoGitAtRepositoryRootKeepsTheHostGitignore(t *testing.T) {
	repo := initTestRepo(t)
	setRepoIdentity(t, repo, "ada", "ada@example.com")

	hostGitignore := filepath.Join(repo, ".gitignore")
	hostContent := "host-rules\n"
	if err := os.WriteFile(hostGitignore, []byte(hostContent), 0600); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", ".gitignore")
	gitRun(t, repo, "commit", "-m", "host ignore rules")

	out, code := initRun(t, repo, ".", "--no-git", "--force")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}

	// The root layout appends three entries, so the notice's undo guidance has
	// to refer to them in the plural, and each entry is anchored to the
	// repository root.
	if want := "kb: excluded /.agent-kb/, /kb/, /raw/ from host git tracking via .git/info/exclude (local to this clone; remove those lines to undo)"; !strings.Contains(out, want) {
		t.Errorf("output = %q, want it to contain %q", out, want)
	}

	data, err := os.ReadFile(hostGitignore) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != hostContent {
		t.Errorf(".gitignore = %q, want the host's committed content %q untouched", data, hostContent)
	}
	if status := gitOutput(t, repo, "status", "--porcelain"); status != "" {
		t.Errorf("host status = %q, want a clean status", status)
	}

	baseGitignore, err := os.ReadFile(filepath.Join(path.StateDir(repo), ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if string(baseGitignore) != "search.db*\n" {
		t.Errorf(".agent-kb/.gitignore = %q, want the base's own rules", baseGitignore)
	}

	exclude, err := os.ReadFile(filepath.Join(repo, ".git", "info", "exclude")) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	anchored := map[string]bool{"/.agent-kb/": true, "/kb/": true, "/raw/": true}
	patterns := 0
	for _, line := range strings.Split(string(exclude), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !anchored[line] {
			t.Errorf("host exclude file carries pattern %q, want only the anchored directories of the base", line)
		}
		patterns++
	}
	if patterns != len(anchored) {
		t.Errorf("host exclude file carries %d patterns, want the base's three anchored directories", patterns)
	}
}

// TestInitEmbedAtRepositoryRootMergesTheHostGitignore covers the layout the
// unversioned root case sits beside: an embedded base at the repository root is
// versioned, so the host's tracked .gitignore still gains the base's rules.
func TestInitEmbedAtRepositoryRootMergesTheHostGitignore(t *testing.T) {
	repo := initTestRepo(t)
	setRepoIdentity(t, repo, "ada", "ada@example.com")

	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("host-rules\n"), 0600); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", ".gitignore")
	gitRun(t, repo, "commit", "-m", "host ignore rules")

	out, code := initRun(t, repo, ".", "--embed", "--force")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}

	data, err := os.ReadFile(filepath.Join(repo, ".gitignore")) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	if want := "host-rules\n*.akb.bak\n.agent-kb/search.db*\n"; string(data) != want {
		t.Errorf(".gitignore = %q, want %q", data, want)
	}
}

// TestInitRecordsTheDefaultIdentity covers the identity an environment without
// any git identity leaves behind: the base records the akb default in akb.yaml,
// and the init commit is attributed to it.
func TestInitRecordsTheDefaultIdentity(t *testing.T) {
	dir := t.TempDir()

	out, code := initRunEnv(t, dir, envWithoutIdentity(), "defaulted-kb")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}
	if want := "commit identity: agent-kb <agent@agent-kb> (default — no git identity found; recorded in akb.yaml, edit git-author/git-email to change)"; !strings.Contains(out, want) {
		t.Errorf("output = %q, want it to contain %q", out, want)
	}

	data, err := os.ReadFile(path.ConfigPath(filepath.Join(dir, "defaulted-kb")))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"git-author: agent-kb", "git-email: agent@agent-kb"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("akb.yaml = %q, want it to contain %q", data, want)
		}
	}

	if author := gitOutput(t, filepath.Join(dir, "defaulted-kb"), "log", "-1", "--format=%an <%ae>"); author != "agent-kb <agent@agent-kb>" {
		t.Errorf("init commit author = %q, want the recorded default", author)
	}
}

// TestInitRegistersAuthorFlags covers the two identity flags `akb init` carries:
// both are registered with an empty default, so an init that names neither
// resolves its identity from the environment or git config as before.
func TestInitRegistersAuthorFlags(t *testing.T) {
	for _, name := range []string{"author-name", "author-email"} {
		flag := initCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("initCmd registers no --%s flag", name)
		}
		if flag.DefValue != "" {
			t.Errorf("--%s default = %q, want the empty default an unnamed identity has", name, flag.DefValue)
		}
	}
}

// TestInitAuthorFlagsOutrankEnvironmentAndGitConfig covers an init whose
// identity comes from a complete flag pair: the invoker declared the identity
// for the base itself, so the report names the flags and the base records the
// pair in akb.yaml, and the init commit is attributed to it — whether git
// config, the environment, or no other source names one.
func TestInitAuthorFlagsOutrankEnvironmentAndGitConfig(t *testing.T) {
	cases := map[string][]string{
		"flags outrank git config": envWithGitIdentity("machine-user", "machine@example.com"),
		"flags outrank the environment": append(envWithGitIdentity("machine-user", "machine@example.com"),
			"AKB_AUTHOR_NAME=env-user", "AKB_AUTHOR_EMAIL=env@example.com"),
		"flags resolve on their own": envWithoutIdentity(),
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()

			out, code := initRunEnv(t, dir, env, "flagged-kb", "--author-name", "Grace Hopper", "--author-email", "grace@example.com")
			if code != exitSuccess {
				t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
			}
			assertIdentityReport(t, out, "commit identity: from --author-name/--author-email (Grace Hopper <grace@example.com>) — recorded in akb.yaml, edit git-author/git-email to change")

			kbDir := filepath.Join(dir, "flagged-kb")
			assertRecordedIdentity(t, kbDir, "Grace Hopper", "grace@example.com")
			if author := gitOutput(t, kbDir, "log", "-1", "--format=%an <%ae>"); author != "Grace Hopper <grace@example.com>" {
				t.Errorf("init commit author = %q, want the flag identity", author)
			}
		})
	}
}

// TestInitPartialAuthorFlagIsNotRecorded covers a partial flag pair: the
// field the flag named attributes the init commit and the other comes from the
// next source down the chain, the machine's git identity — but the pair is
// ambient for the field the flag left out, so the base records nothing. Only a
// complete pair is an identity declared for the base itself.
func TestInitPartialAuthorFlagIsNotRecorded(t *testing.T) {
	cases := map[string]struct {
		args  []string
		given string
		want  string
	}{
		"name only":  {[]string{"--author-name", "Grace Hopper"}, "--author-name", "Grace Hopper <machine@example.com>"},
		"email only": {[]string{"--author-email", "grace@example.com"}, "--author-email", "machine-user <grace@example.com>"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()

			args := append([]string{"named-kb"}, tc.args...)
			out, code := initRunEnv(t, dir, envWithGitIdentity("machine-user", "machine@example.com"), args...)
			if code != exitSuccess {
				t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
			}
			assertIdentityReport(t, out, "commit identity: from "+tc.given+" alone ("+tc.want+") — not recorded; only a complete --author-name/--author-email pair is recorded in akb.yaml; later commits use each machine's identity")

			kbDir := filepath.Join(dir, "named-kb")
			assertNoRecordedIdentity(t, kbDir)
			if author := gitOutput(t, kbDir, "log", "-1", "--format=%an <%ae>"); author != tc.want {
				t.Errorf("init commit author = %q, want the flag field filled from git config, %q", author, tc.want)
			}
		})
	}
}

// TestInitAuthorFlagsWithNoCommitLeaveTheCommitToTheCaller covers a complete
// flag pair of an init that records no commit: the scaffold succeeds without an
// exit-2 identity refusal, nothing is committed, and the declared identity is
// still recorded in akb.yaml — it is the base's own, and the later commits akb
// makes are attributed to it — while the report leaves the init commit itself
// to the caller.
func TestInitAuthorFlagsWithNoCommitLeaveTheCommitToTheCaller(t *testing.T) {
	dir := t.TempDir()
	kbDir := filepath.Join(dir, "staged-kb")

	out, code := initRunEnv(t, dir, envWithoutIdentity(), "staged-kb", "--no-commit", "--author-name", "Grace Hopper", "--author-email", "grace@example.com")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}
	assertIdentityReport(t, out, "commit identity: from --author-name/--author-email (Grace Hopper <grace@example.com>) — recorded in akb.yaml, edit git-author/git-email to change; the commit itself stays with the caller")

	logCmd := exec.Command("git", "log", "-1", "--format=%s") //nolint:gosec // test helper launching trusted git binary
	logCmd.Dir = kbDir
	if err := logCmd.Run(); err == nil {
		t.Error("--no-commit recorded a commit")
	}

	assertRecordedIdentity(t, kbDir, "Grace Hopper", "grace@example.com")
}

// assertIdentityReport fails the test unless out carries the identity report the
// scenario resolved.
func assertIdentityReport(t *testing.T, out, want string) {
	t.Helper()

	if !strings.Contains(out, want) {
		t.Errorf("output = %q, want it to contain %q", out, want)
	}
}

// assertNoRecordedIdentity fails the test unless the base's akb.yaml carries no
// commit identity: an ambient identity — the environment's, git config's, or a
// partial flag pair's — is never written into a file that travels with the
// base.
func assertNoRecordedIdentity(t *testing.T, kbDir string) {
	t.Helper()

	data, err := os.ReadFile(path.ConfigPath(kbDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"git-author", "git-email"} {
		if strings.Contains(string(data), key) {
			t.Errorf("akb.yaml = %q, want no %s key", data, key)
		}
	}
}

// assertRecordedIdentity fails the test unless the base's akb.yaml records the
// given commit identity: a complete flag pair the invoker declared for the
// base, or the akb default.
func assertRecordedIdentity(t *testing.T, kbDir, name, email string) {
	t.Helper()

	data, err := os.ReadFile(path.ConfigPath(kbDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"git-author: " + name, "git-email: " + email} {
		if !strings.Contains(string(data), want) {
			t.Errorf("akb.yaml = %q, want it to contain %q", data, want)
		}
	}
}

// TestInitNoCommitLeavesTheCommitToTheCaller covers `akb init --no-commit`: the
// base is scaffolded and staged, and its commit is the caller's to make.
func TestInitNoCommitLeavesTheCommitToTheCaller(t *testing.T) {
	dir := t.TempDir()
	kbDir := filepath.Join(dir, "staged-kb")

	out, code := initRun(t, dir, "staged-kb", "--no-commit")
	if code != exitSuccess {
		t.Fatalf("exit code = %d, want %d (output: %s)", code, exitSuccess, out)
	}

	logCmd := exec.Command("git", "log", "-1", "--format=%s") //nolint:gosec // test helper launching trusted git binary
	logCmd.Dir = kbDir
	if err := logCmd.Run(); err == nil {
		t.Error("--no-commit recorded a commit")
	}

	status := gitOutput(t, kbDir, "status", "--porcelain")
	for _, want := range []string{"kb/index.md", "raw/files.log", ".agent-kb/akb.yaml"} {
		if !strings.Contains(status, want) {
			t.Errorf("status = %q, want %q staged for the caller", status, want)
		}
	}
}

// TestAppendMissingLinesKeepsExistingLines covers the ignore-file merge: the
// lines that are already in the file survive, the missing ones are appended, a
// file without a final newline is not run into, and writing the same lines again
// changes nothing.
func TestAppendMissingLinesKeepsExistingLines(t *testing.T) {
	target := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(target, []byte("unrelated.txt"), 0600); err != nil {
		t.Fatal(err)
	}

	want := "unrelated.txt\n*.akb.bak\n.agent-kb/search.db*\n"
	if err := appendMissingLines(target, "*.akb.bak", ".agent-kb/search.db*"); err != nil {
		t.Fatalf("appendMissingLines: %v", err)
	}
	data, err := os.ReadFile(target) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Errorf("merged .gitignore = %q, want %q", data, want)
	}

	if err := appendMissingLines(target, "*.akb.bak", ".agent-kb/search.db*"); err != nil {
		t.Fatalf("appendMissingLines: %v", err)
	}
	data, err = os.ReadFile(target) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Errorf("re-merged .gitignore = %q, want it unchanged at %q", data, want)
	}

	missing := filepath.Join(t.TempDir(), "new", ".gitignore")
	if err := appendMissingLines(missing, "search.db*"); err != nil {
		t.Fatalf("appendMissingLines: %v", err)
	}
	data, err = os.ReadFile(missing) //nolint:gosec // test temp file
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "search.db*\n" {
		t.Errorf("created .gitignore = %q, want the missing lines", data)
	}
}

// envWithGitIdentity returns the process environment with the machine's git
// identity replaced by name and email: the akb and git identity variables are
// removed and git's merged configuration names exactly the given identity, so a
// scenario resolving one does not depend on the host's configuration.
func envWithGitIdentity(name, email string) []string {
	env := make([]string, 0, len(os.Environ())+4)
	for _, variable := range os.Environ() {
		key, _, _ := strings.Cut(variable, "=")
		if strings.HasPrefix(key, "GIT_") || strings.HasPrefix(key, "AKB_AUTHOR_") {
			continue
		}
		env = append(env, variable)
	}
	return append(env,
		"GIT_CONFIG_COUNT=2",
		"GIT_CONFIG_KEY_0=user.name",
		"GIT_CONFIG_VALUE_0="+name,
		"GIT_CONFIG_KEY_1=user.email",
		"GIT_CONFIG_VALUE_1="+email,
	)
}

// envWithoutIdentity returns the process environment with every source of a
// git identity removed: the akb and git identity variables, and the machine's
// configuration, which the empty user.name and user.email values override. Git
// reads an empty value as no identity at all.
func envWithoutIdentity() []string {
	return envWithGitIdentity("", "")
}
