package storage_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/peedrr/agent-kb/internal/storage"
)

// The concurrency tests below drive provider operations from separate processes:
// only distinct processes can show that akb serializes itself on the repository
// lock and survives the git index lock held by anything else.

const (
	helperModeEnv    = "AKB_TEST_PROVIDER_MODE"
	helperRootEnv    = "AKB_TEST_PROVIDER_ROOT"
	helperPathEnv    = "AKB_TEST_PROVIDER_PATH"
	helperContentEnv = "AKB_TEST_PROVIDER_CONTENT"
	helperMsgEnv     = "AKB_TEST_PROVIDER_MSG"
	helperReadyEnv   = "AKB_TEST_PROVIDER_READY_FILE"
)

// TestGitProviderHelperProcess is the entry point of the child processes the
// concurrency tests start with startHelper. It performs one provider operation
// described by the environment and fails the child process if it does not
// succeed.
func TestGitProviderHelperProcess(t *testing.T) {
	mode := os.Getenv(helperModeEnv)
	if mode == "" {
		t.Skip("helper process for the provider concurrency tests; driven by startHelper")
	}

	root := os.Getenv(helperRootEnv)
	target := filepath.Join(root, os.Getenv(helperPathEnv)) //nolint:gosec // test helper resolving a path inside the temp repository it was given
	if ready := os.Getenv(helperReadyEnv); ready != "" {
		if err := os.WriteFile(ready, []byte("ready\n"), 0600); err != nil { //nolint:gosec // test helper writing its own marker file
			t.Fatalf("mark helper ready: %v", err)
		}
	}

	provider := storage.NewGitProvider(root, false)
	ctx := context.Background()

	var err error
	switch mode {
	case "write":
		err = provider.WriteWithCommitMsg(ctx, target, []byte(os.Getenv(helperContentEnv)), os.Getenv(helperMsgEnv))
	case "delete":
		err = provider.Delete(ctx, target)
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}
	if err != nil {
		t.Fatalf("provider %s %s: %v", mode, target, err)
	}
}

// helperProcess is a provider operation running in its own process.
type helperProcess struct {
	cmd     *exec.Cmd
	output  *strings.Builder
	done    chan struct{}
	waitErr error
}

// startHelper starts one provider operation in a child process of this test
// binary. readyFile, when set, is created by the child right before the
// operation runs.
func startHelper(t *testing.T, mode, root, relPath, content, msg, readyFile string) *helperProcess {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^TestGitProviderHelperProcess$") //nolint:gosec // re-running the test binary with a fixed test name
	cmd.Dir = root
	cmd.Env = helperEnv([]string{
		helperModeEnv + "=" + mode,
		helperRootEnv + "=" + root,
		helperPathEnv + "=" + relPath,
		helperContentEnv + "=" + content,
		helperMsgEnv + "=" + msg,
		helperReadyEnv + "=" + readyFile,
	}...)

	output := &strings.Builder{}
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper process (%s %s): %v", mode, relPath, err)
	}

	helper := &helperProcess{cmd: cmd, output: output, done: make(chan struct{})}
	go func() {
		helper.waitErr = cmd.Wait()
		close(helper.done)
	}()
	return helper
}

// helperEnv returns the environment for a helper process without any helper
// variable inherited from the test run itself.
func helperEnv(extra ...string) []string {
	env := make([]string, 0, len(extra))
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "AKB_TEST_PROVIDER_") {
			continue
		}
		env = append(env, v)
	}
	return append(env, extra...)
}

// running reports whether the helper process is still running.
func (h *helperProcess) running() bool {
	select {
	case <-h.done:
		return false
	default:
		return true
	}
}

// wait blocks until the helper process finished and fails the test unless it
// exited successfully.
func (h *helperProcess) wait(t *testing.T) {
	t.Helper()
	select {
	case <-h.done:
	case <-time.After(30 * time.Second):
		_ = h.cmd.Process.Kill() //nolint:errcheck // test cleanup
		t.Fatalf("helper process did not finish in time\n%s", h.output.String())
	}
	if h.waitErr != nil {
		t.Fatalf("helper process failed: %v\n%s", h.waitErr, h.output.String())
	}
}

// initRepo creates a temporary git repository with one commit and returns it.
func initRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	mustGit(t, repo, "init")
	mustGit(t, repo, "config", "user.name", "test")
	mustGit(t, repo, "config", "user.email", "test@test.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# repo\n"), 0600); err != nil {
		t.Fatal(err)
	}
	mustGit(t, repo, "add", "--", "README.md")
	mustGit(t, repo, "commit", "-m", "initial")
	return repo
}

// mustGit runs git in dir and returns its trimmed combined output.
func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...) //nolint:gosec // test helper launching the git binary
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %s: %v", strings.Join(args, " "), dir, strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out))
}

// relPathInRepo returns path relative to repo, which contains it.
func relPathInRepo(t *testing.T, repo, path string) string {
	t.Helper()
	rel, err := filepath.Rel(repo, path)
	if err != nil {
		t.Fatalf("relative path of %s in %s: %v", path, repo, err)
	}
	return filepath.ToSlash(rel)
}

// commitKBTree creates the KB-managed files (kb/index.md and kb/log.md) inside
// kbDir and commits them, so provider operations have managed files to update.
func commitKBTree(t *testing.T, repo, kbDir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(kbDir, "kb", "notes"), 0750); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.md", "log.md"} {
		if err := os.WriteFile(filepath.Join(kbDir, "kb", name), []byte("# "+name+"\n"), 0600); err != nil { //nolint:gosec // test helper writing into its temp repository
			t.Fatal(err)
		}
	}
	mustGit(t, repo, "add", "--", relPathInRepo(t, repo, filepath.Join(kbDir, "kb")))
	mustGit(t, repo, "commit", "-m", "add kb tree")
}

// repoLockPath returns the lock file akb uses for the repository at repo. Every
// KB inside the repository — including KBs in linked worktrees — locks this file.
func repoLockPath(repo string) string {
	return filepath.Join(repo, ".git", "akb.lock")
}

// holdRepoLock takes the exclusive repository lock and returns the function that
// releases it.
func holdRepoLock(t *testing.T, repo string) func() {
	t.Helper()

	lock, err := os.OpenFile(repoLockPath(repo), os.O_CREATE|os.O_RDWR, 0600) //nolint:gosec // test helper locking a fixed path in its temp repository
	if err != nil {
		t.Fatalf("open repo lock: %v", err)
	}
	if err := unix.Flock(int(lock.Fd()), unix.LOCK_EX); err != nil { //nolint:gosec // file descriptors fit in an int
		t.Fatalf("flock %s: %v", repoLockPath(repo), err)
	}

	released := false
	t.Cleanup(func() {
		if !released {
			_ = unix.Flock(int(lock.Fd()), unix.LOCK_UN) //nolint:gosec // file descriptors fit in an int
		}
		_ = lock.Close() //nolint:errcheck // test cleanup — closing releases the lock
	})

	return func() {
		if released {
			return
		}
		released = true
		if err := unix.Flock(int(lock.Fd()), unix.LOCK_UN); err != nil { //nolint:gosec // file descriptors fit in an int
			t.Fatalf("unlock %s: %v", repoLockPath(repo), err)
		}
	}
}

// commitFiles returns the sorted paths recorded by HEAD, relative to the repo.
func commitFiles(t *testing.T, repo string) []string {
	t.Helper()
	var files []string
	for _, line := range strings.Split(mustGit(t, repo, "show", "--name-only", "--format=", "HEAD"), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files
}

// commitExists reports whether the log contains a commit with exactly msg.
func commitExists(t *testing.T, repo, msg string) bool {
	t.Helper()
	for _, subject := range strings.Split(mustGit(t, repo, "log", "--format=%s"), "\n") {
		if strings.TrimSpace(subject) == msg {
			return true
		}
	}
	return false
}

// assertCleanWorktree fails the test when the repository has staged or unstaged
// changes left behind.
func assertCleanWorktree(t *testing.T, repo string) {
	t.Helper()
	if status := mustGit(t, repo, "status", "--porcelain"); status != "" {
		t.Errorf("repository is not clean:\n%s", status)
	}
}

// waitForFile waits until path exists.
func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", path)
}

// blockedWriter starts a provider write and asserts that it is waiting for the
// repository lock: the process runs, but neither writes the page nor commits.
func blockedWriter(t *testing.T, root, relPath, content, msg string) *helperProcess {
	t.Helper()

	ready := filepath.Join(t.TempDir(), "ready")
	helper := startHelper(t, "write", root, relPath, content, msg, ready)
	waitForFile(t, ready)
	time.Sleep(200 * time.Millisecond)

	if _, err := os.Stat(filepath.Join(root, relPath)); !os.IsNotExist(err) { //nolint:gosec // test helper checking its temp repository
		t.Errorf("%s exists while the repository lock is held (stat error: %v)", relPath, err)
	}
	if !helper.running() {
		t.Fatalf("helper for %s finished while the repository lock was held\n%s", relPath, helper.output.String())
	}
	if commitExists(t, root, msg) {
		t.Errorf("%s was committed while the repository lock was held", relPath)
	}
	return helper
}

// writerIndex returns the writer number named by a helper commit message of the
// form "akb: write <path> (writer N)".
func writerIndex(t *testing.T, msg string) int {
	t.Helper()
	open := strings.LastIndex(msg, "(writer ")
	if open < 0 || !strings.HasSuffix(msg, ")") {
		t.Fatalf("commit message %q does not name a writer", msg)
	}
	n, err := strconv.Atoi(msg[open+len("(writer ") : len(msg)-1])
	if err != nil {
		t.Fatalf("commit message %q: %v", msg, err)
	}
	return n
}

// TestGitProviderWriteWaitsForRepoLock asserts the repository lock itself: while
// a process holds it, another process running a provider write neither writes
// the page nor commits, and releasing the lock lets it finish.
func TestGitProviderWriteWaitsForRepoLock(t *testing.T) {
	repo := initRepo(t)
	commitKBTree(t, repo, repo)

	release := holdRepoLock(t, repo)

	const rel = "kb/notes/blocked.md"
	helper := blockedWriter(t, repo, rel, "blocked content\n", "akb: write "+rel)

	release()
	helper.wait(t)

	data, err := os.ReadFile(filepath.Join(repo, rel)) //nolint:gosec // test reading a path inside its temp repository
	if err != nil {
		t.Fatalf("read page: %v", err)
	}
	if string(data) != "blocked content\n" {
		t.Errorf("page content = %q, want %q", string(data), "blocked content\n")
	}
	if head := mustGit(t, repo, "log", "-1", "--format=%s"); head != "akb: write "+rel {
		t.Errorf("HEAD subject = %q, want %q", head, "akb: write "+rel)
	}
}

// TestGitProviderConcurrentWritersInRepoRootKB covers a KB whose root is the git
// repository root: concurrent writer processes commit every page and leave the
// worktree clean.
func TestGitProviderConcurrentWritersInRepoRootKB(t *testing.T) {
	repo := initRepo(t)
	commitKBTree(t, repo, repo)

	const writers = 4
	helpers := make([]*helperProcess, 0, writers)
	for i := range writers {
		rel := fmt.Sprintf("kb/notes/writer-%d.md", i)
		helpers = append(helpers, startHelper(t, "write", repo, rel, fmt.Sprintf("content from writer %d\n", i), "akb: write "+rel, ""))
	}
	for _, helper := range helpers {
		helper.wait(t)
	}

	for i := range writers {
		rel := fmt.Sprintf("kb/notes/writer-%d.md", i)
		data, err := os.ReadFile(filepath.Join(repo, rel)) //nolint:gosec // test reading a path inside its temp repository
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if want := fmt.Sprintf("content from writer %d\n", i); string(data) != want {
			t.Errorf("content of %s = %q, want %q", rel, string(data), want)
		}
		if !commitExists(t, repo, "akb: write "+rel) {
			t.Errorf("no commit for %s", rel)
		}
	}

	assertCleanWorktree(t, repo)
	if _, err := os.Stat(repoLockPath(repo)); err != nil {
		t.Errorf("repository lock file missing after concurrent writes: %v", err)
	}
}

// TestGitProviderCommitLeavesUnrelatedStagedChanges covers a KB living in a
// subdirectory of a larger repository: akb commits only the paths of its own
// operation and leaves the staged work of the surrounding project alone.
func TestGitProviderCommitLeavesUnrelatedStagedChanges(t *testing.T) {
	repo := initRepo(t)
	kbRoot := filepath.Join(repo, "product")
	commitKBTree(t, repo, kbRoot)

	unrelated := filepath.Join(repo, "src", "app.go")
	if err := os.MkdirAll(filepath.Dir(unrelated), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unrelated, []byte("package main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	mustGit(t, repo, "add", "--", "src/app.go")

	if staged := mustGit(t, repo, "diff", "--cached", "--name-only"); staged != "src/app.go" {
		t.Fatalf("precondition: staged changes = %q, want src/app.go", staged)
	}

	t.Run("write commits only the page", func(t *testing.T) {
		const rel = "kb/notes/page.md"
		helper := startHelper(t, "write", kbRoot, rel, "page body\n", "akb: write "+rel, "")
		helper.wait(t)

		if files := commitFiles(t, repo); !reflect.DeepEqual(files, []string{"product/kb/notes/page.md"}) {
			t.Errorf("commit recorded %v, want only product/kb/notes/page.md", files)
		}
		if status := mustGit(t, repo, "status", "--porcelain"); status != "A  src/app.go" {
			t.Errorf("unrelated staged changes = %q, want %q", status, "A  src/app.go")
		}
	})

	t.Run("delete commits the page with the restaged index and log", func(t *testing.T) {
		const rel = "kb/notes/doomed.md"
		page := filepath.Join(kbRoot, rel)
		if err := os.MkdirAll(filepath.Dir(page), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(page, []byte("doomed body\n"), 0600); err != nil { //nolint:gosec // test helper writing into its temp repository
			t.Fatal(err)
		}
		mustGit(t, kbRoot, "add", "--", rel)
		// A pathspec keeps the surrounding project's staged work out of this
		// setup commit.
		mustGit(t, kbRoot, "commit", "-m", "add doomed page", "--only", "--", rel)

		// `akb delete` updates and restages the managed files before the page
		// deletion is committed.
		for _, name := range []string{"index.md", "log.md"} {
			path := filepath.Join(kbRoot, "kb", name)
			file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600) //nolint:gosec // test helper appending to a file in its temp repository
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.WriteString("updated by delete\n"); err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
		}
		mustGit(t, kbRoot, "add", "--", "kb/index.md", "kb/log.md")
		staged := mustGit(t, repo, "diff", "--cached", "--name-only")
		if want := "product/kb/index.md\nproduct/kb/log.md\nsrc/app.go"; staged != want {
			t.Fatalf("precondition: staged files = %q, want %q", staged, want)
		}

		helper := startHelper(t, "delete", kbRoot, rel, "", "akb: delete "+rel, "")
		helper.wait(t)

		want := []string{"product/kb/index.md", "product/kb/log.md", "product/kb/notes/doomed.md"}
		if files := commitFiles(t, repo); !reflect.DeepEqual(files, want) {
			t.Errorf("commit recorded %v, want %v", files, want)
		}
		if status := mustGit(t, repo, "status", "--porcelain"); status != "A  src/app.go" {
			t.Errorf("unrelated staged changes = %q, want %q", status, "A  src/app.go")
		}
	})
}

// TestGitProviderTwoKBsInOneRepoConcurrentWriters covers two KBs sharing one git
// repository: both KBs wait on the same lock and every writer's page survives.
func TestGitProviderTwoKBsInOneRepoConcurrentWriters(t *testing.T) {
	repo := initRepo(t)
	kbNames := []string{"kbA", "kbB"}
	for _, name := range kbNames {
		commitKBTree(t, repo, filepath.Join(repo, name))
	}

	// Precondition: the lock of the shared repository gates writers of both KBs.
	release := holdRepoLock(t, repo)
	blocked := make([]*helperProcess, 0, len(kbNames))
	for i, name := range kbNames {
		rel := fmt.Sprintf("kb/notes/blocked-%d.md", i)
		blocked = append(blocked, blockedWriter(t, filepath.Join(repo, name), rel, "blocked\n", "akb: write "+name+"/"+rel))
	}
	release()
	for _, helper := range blocked {
		helper.wait(t)
	}

	const writersPerKB = 3
	helpers := make([]*helperProcess, 0, len(kbNames)*writersPerKB)
	for _, name := range kbNames {
		for i := range writersPerKB {
			rel := fmt.Sprintf("kb/notes/%s-writer-%d.md", name, i)
			content := fmt.Sprintf("content from %s writer %d\n", name, i)
			helpers = append(helpers, startHelper(t, "write", filepath.Join(repo, name), rel, content, "akb: write "+name+"/"+rel, ""))
		}
	}
	for _, helper := range helpers {
		helper.wait(t)
	}

	for _, name := range kbNames {
		for i := range writersPerKB {
			rel := fmt.Sprintf("kb/notes/%s-writer-%d.md", name, i)
			data, err := os.ReadFile(filepath.Join(repo, name, rel)) //nolint:gosec // test reading a path inside its temp repository
			if err != nil {
				t.Fatalf("read %s: %v", rel, err)
			}
			if want := fmt.Sprintf("content from %s writer %d\n", name, i); string(data) != want {
				t.Errorf("content of %s/%s = %q, want %q", name, rel, string(data), want)
			}
			if !commitExists(t, repo, "akb: write "+name+"/"+rel) {
				t.Errorf("no commit for %s/%s", name, rel)
			}
		}
	}

	assertCleanWorktree(t, repo)
}

// TestGitProviderConcurrentWritesToSamePage covers concurrent writers of the same
// page: their writes serialize, so each commit holds exactly the content that
// writer produced and no update is lost.
func TestGitProviderConcurrentWritesToSamePage(t *testing.T) {
	repo := initRepo(t)
	commitKBTree(t, repo, repo)

	const rel = "kb/notes/shared.md"
	const writers = 3
	helpers := make([]*helperProcess, 0, writers)
	for i := range writers {
		content := fmt.Sprintf("content from writer %d\n", i)
		msg := fmt.Sprintf("akb: write %s (writer %d)", rel, i)
		helpers = append(helpers, startHelper(t, "write", repo, rel, content, msg, ""))
	}
	for _, helper := range helpers {
		helper.wait(t)
	}

	commits := strings.Split(mustGit(t, repo, "rev-list", "HEAD", "--", rel), "\n")
	if len(commits) != writers {
		t.Fatalf("commits touching %s = %d, want %d", rel, len(commits), writers)
	}

	seen := make(map[int]bool, writers)
	for _, sha := range commits {
		msg := mustGit(t, repo, "log", "-1", "--format=%s", sha)
		index := writerIndex(t, msg)
		content := mustGit(t, repo, "show", sha+":"+rel)
		if want := fmt.Sprintf("content from writer %d", index); content != want {
			t.Errorf("commit %s (%s) recorded %q, want %q — concurrent writes interleaved", sha[:7], msg, content, want)
		}
		seen[index] = true
	}
	if len(seen) != writers {
		t.Errorf("commits named %d of %d writers: %v", len(seen), writers, seen)
	}

	worktree, err := os.ReadFile(filepath.Join(repo, rel)) //nolint:gosec // test reading a path inside its temp repository
	if err != nil {
		t.Fatalf("read page: %v", err)
	}
	if head := mustGit(t, repo, "show", "HEAD:"+rel); strings.TrimSpace(string(worktree)) != head {
		t.Errorf("worktree content = %q, want the content of HEAD (%q)", string(worktree), head)
	}
	assertCleanWorktree(t, repo)
}

// TestGitProviderWorktreeWriteWaitsForRepoLock asserts that a KB in a linked
// worktree locks the repository it shares with the main checkout rather than its
// own git dir.
func TestGitProviderWorktreeWriteWaitsForRepoLock(t *testing.T) {
	repo := initRepo(t)
	worktree := t.TempDir()
	mustGit(t, repo, "worktree", "add", "--detach", worktree)

	kbRoot := filepath.Join(worktree, "kb-worktree")
	commitKBTree(t, worktree, kbRoot)

	release := holdRepoLock(t, repo)

	const rel = "kb/notes/blocked-worktree.md"
	helper := blockedWriter(t, kbRoot, rel, "worktree content\n", "akb: write "+rel)
	release()
	helper.wait(t)

	if head := mustGit(t, kbRoot, "log", "-1", "--format=%s"); head != "akb: write "+rel {
		t.Errorf("HEAD subject = %q, want %q", head, "akb: write "+rel)
	}
}

// TestGitProviderRetriesWhileIndexLockHeld asserts that a write survives the git
// index lock held by another tool: it waits for the lock instead of failing.
func TestGitProviderRetriesWhileIndexLockHeld(t *testing.T) {
	repo := initRepo(t)
	commitKBTree(t, repo, repo)

	indexLock := filepath.Join(repo, ".git", "index.lock")
	if err := os.WriteFile(indexLock, nil, 0600); err != nil { //nolint:gosec // test helper creating a fixed file in its temp repository
		t.Fatal(err)
	}
	const holdFor = 300 * time.Millisecond
	released := make(chan struct{})
	go func() {
		defer close(released)
		time.Sleep(holdFor)
		_ = os.Remove(indexLock) //nolint:errcheck // test releasing the index lock
	}()

	provider := storage.NewGitProvider(repo, false)
	started := time.Now()
	err := provider.Write(context.Background(), filepath.Join(repo, "kb", "notes", "retried.md"), []byte("retried content\n"))
	if err != nil {
		t.Fatalf("write failed while the index lock was held: %v", err)
	}
	elapsed := time.Since(started)
	<-released

	if elapsed < holdFor {
		t.Errorf("write finished after %v, before the index lock was released after %v", elapsed, holdFor)
	}
	if !commitExists(t, repo, "akb: write kb/notes/retried.md") {
		t.Errorf("write was not committed")
	}
	assertCleanWorktree(t, repo)
}
