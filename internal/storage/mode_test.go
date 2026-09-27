// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/path"
)

// writeKBConfig writes the akb.yaml of the base rooted at kbRoot, so the base
// can be opened as a Store.
func writeKBConfig(t *testing.T, kbRoot, versioning string) {
	t.Helper()

	if err := os.MkdirAll(path.StateDir(kbRoot), 0750); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Name:       "test-kb",
		Created:    "2024-01-15T10:30:00Z",
		Versioning: versioning,
	}
	if err := config.Save(path.ConfigPath(kbRoot), cfg); err != nil {
		t.Fatal(err)
	}
}

// commitKBConfig records the base's config, so the repository starts from a
// clean worktree.
func commitKBConfig(t *testing.T, repo string) {
	t.Helper()

	mustGitIn(t, repo, "add", "--", path.StateDirName)
	mustGitIn(t, repo, "commit", "-m", "add kb config")
}

// lockFree reports whether lockPath is free to lock: it takes a non-blocking
// flock on the file from a second open file description, which conflicts when
// the process already holds the lock through its own handle.
func lockFree(t *testing.T, lockPath string) bool {
	t.Helper()

	f, err := os.OpenFile(lockPath, os.O_RDWR, 0600) //nolint:gosec // test helper locking the file it is given
	if err != nil {
		t.Fatalf("open %s: %v", lockPath, err)
	}
	defer f.Close() //nolint:errcheck // test helper

	err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB) //nolint:gosec // file descriptors fit in an int
	if err == nil {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN) //nolint:errcheck,gosec // test helper releasing its probe lock
		return true
	}
	if errors.Is(err, unix.EWOULDBLOCK) {
		return false
	}
	t.Fatalf("flock %s: %v", lockPath, err)
	return false
}

// repoWithInProgressMerge returns a repository with a merge in progress and the
// path the merge changed, relative to the repository root. The merge conflicts
// when conflicting is set — both branches change the same file — and is clean
// otherwise, with the branches changing different files.
func repoWithInProgressMerge(t *testing.T, conflicting bool) (string, string) {
	t.Helper()

	repo := initGitRepo(t)
	branch := mustGitIn(t, repo, "rev-parse", "--abbrev-ref", "HEAD")

	writeAndCommit := func(file, content string) {
		t.Helper()
		fullPath := filepath.Join(repo, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil { //nolint:gosec // test helper writing into its temp repository
			t.Fatal(err)
		}
		mustGitIn(t, repo, "add", "--", file)
		mustGitIn(t, repo, "commit", "-m", content)
	}

	mustGitIn(t, repo, "checkout", "-b", "feature")
	featureFile := "src/other.go"
	if conflicting {
		featureFile = "src/app.go"
	}
	writeAndCommit(featureFile, "feature change\n")

	mustGitIn(t, repo, "checkout", branch)
	if conflicting {
		writeAndCommit(featureFile, "base change\n")
	} else {
		writeAndCommit("src/app.go", "base change\n")
	}

	if conflicting {
		// The merge conflicts: git leaves the conflict in the index and
		// MERGE_HEAD behind, which blocks every commit until it is resolved.
		if _, err := gitIn(t, repo, "merge", "feature"); err == nil {
			t.Fatal("precondition: the merge did not conflict")
		}
		return repo, featureFile
	}

	mustGitIn(t, repo, "merge", "--no-commit", "feature")
	return repo, featureFile
}

func TestParseMode(t *testing.T) {
	t.Run("treats a missing mode as git", func(t *testing.T) {
		mode, err := ParseMode("")
		if err != nil {
			t.Fatalf("ParseMode: %v", err)
		}
		if mode != ModeGit {
			t.Errorf("mode = %q, want %q", mode, ModeGit)
		}
	})

	t.Run("reads both configured modes", func(t *testing.T) {
		for value, want := range map[string]Mode{
			config.VersioningGit:  ModeGit,
			config.VersioningNone: ModeNone,
		} {
			mode, err := ParseMode(value)
			if err != nil {
				t.Fatalf("ParseMode(%q): %v", value, err)
			}
			if mode != want {
				t.Errorf("ParseMode(%q) = %q, want %q", value, mode, want)
			}
		}
	})

	t.Run("rejects an unknown mode", func(t *testing.T) {
		mode, err := ParseMode("svn")
		if err == nil {
			t.Fatalf("ParseMode(\"svn\") = %q, want an error", mode)
		}
		if !strings.Contains(err.Error(), config.VersioningNone) {
			t.Errorf("error %q does not name the accepted modes", err)
		}
	})
}

func TestNewProviderSelectsByMode(t *testing.T) {
	kbRoot := t.TempDir()

	t.Run("selects the git provider for a git-versioned base", func(t *testing.T) {
		provider := newProvider(kbRoot, ModeGit, false)
		git, ok := provider.(*GitProvider)
		if !ok {
			t.Fatalf("provider for %q is %T, want *GitProvider", ModeGit, provider)
		}
		if git.noCommit {
			t.Error("the git provider did not carry the no-commit flag")
		}
		if noCommit := newProvider(kbRoot, ModeGit, true).(*GitProvider).noCommit; !noCommit { //nolint:forcetypeassert // the mode selects the provider under test
			t.Error("the git provider dropped the no-commit flag")
		}
	})

	t.Run("selects the filesystem provider for an unversioned base", func(t *testing.T) {
		provider := newProvider(kbRoot, ModeNone, false)
		if _, ok := provider.(*FilesystemProvider); !ok {
			t.Fatalf("provider for %q is %T, want *FilesystemProvider", ModeNone, provider)
		}
	})
}

func TestOpenStoreReadsTheVersioningMode(t *testing.T) {
	t.Run("opens a git-versioned base", func(t *testing.T) {
		kbRoot := t.TempDir()
		writeKBConfig(t, kbRoot, config.VersioningGit)

		store, err := OpenStore(kbRoot, false)
		if err != nil {
			t.Fatalf("OpenStore: %v", err)
		}
		if store.Mode() != ModeGit {
			t.Errorf("Mode = %q, want %q", store.Mode(), ModeGit)
		}
		if _, ok := store.provider.(*GitProvider); !ok {
			t.Errorf("provider of a git-versioned base is %T, want *GitProvider", store.provider)
		}
	})

	t.Run("opens an unversioned base", func(t *testing.T) {
		kbRoot := t.TempDir()
		writeKBConfig(t, kbRoot, config.VersioningNone)

		store, err := OpenStore(kbRoot, false)
		if err != nil {
			t.Fatalf("OpenStore: %v", err)
		}
		if store.Mode() != ModeNone {
			t.Errorf("Mode = %q, want %q", store.Mode(), ModeNone)
		}
		if _, ok := store.provider.(*FilesystemProvider); !ok {
			t.Errorf("provider of an unversioned base is %T, want *FilesystemProvider", store.provider)
		}
	})

	t.Run("rejects an unknown versioning mode", func(t *testing.T) {
		kbRoot := t.TempDir()
		writeKBConfig(t, kbRoot, "svn")

		if _, err := OpenStore(kbRoot, false); err == nil {
			t.Fatal("OpenStore accepted an unknown versioning mode")
		}
	})
}

// TestStoreUnversionedBaseNeverInvokesGit asserts a base versioned in nothing
// performs its file operations without running git: the only git binary on PATH
// fails and records that it was called.
func TestStoreUnversionedBaseNeverInvokesGit(t *testing.T) {
	kbRoot := t.TempDir()
	writeKBConfig(t, kbRoot, config.VersioningNone)

	binDir := t.TempDir()
	shimRan := filepath.Join(t.TempDir(), "git-ran")
	shim := "#!/bin/sh\necho \"$@\" >> " + shimRan + "\nexit 1\n"
	if err := os.WriteFile(filepath.Join(binDir, "git"), []byte(shim), 0750); err != nil { //nolint:gosec // test helper writing an executable shim
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	store, err := OpenStore(kbRoot, false)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	ctx := context.Background()
	page := filepath.Join(kbRoot, "kb", "notes", "page.md")
	if err := store.Write(ctx, page, []byte("page body\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, err := os.ReadFile(page) //nolint:gosec // test reading a path inside its temp base
	if err != nil {
		t.Fatalf("read page: %v", err)
	}
	if string(data) != "page body\n" {
		t.Errorf("page content = %q, want %q", string(data), "page body\n")
	}

	if err := store.WriteWithCommitMsg(ctx, page, []byte("updated body\n"), "akb: write kb/notes/page.md"); err != nil {
		t.Fatalf("WriteWithCommitMsg: %v", err)
	}
	if err := store.StageFiles("kb/index.md", "kb/log.md"); err != nil {
		t.Errorf("StageFiles in an unversioned base = %v, want no-op", err)
	}
	if err := store.Commit("akb: write kb/notes/page.md", "kb/notes/page.md"); err != nil {
		t.Errorf("Commit in an unversioned base = %v, want no-op", err)
	}
	nothing, err := store.NothingToCommit("kb/notes/page.md")
	if err != nil || !nothing {
		t.Errorf("NothingToCommit in an unversioned base = %v, %v, want true, nil", nothing, err)
	}

	if err := store.Delete(ctx, page); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(page); !os.IsNotExist(err) {
		t.Errorf("page still exists after Delete (stat error: %v)", err)
	}

	if recorded, err := os.ReadFile(shimRan); err == nil { //nolint:gosec // test reading its own marker file
		t.Errorf("git was invoked in a base versioned in nothing: %s", recorded)
	}
}

// TestStoreUnversionedBaseLockLocksItsOwnLockFile asserts the lock of a base
// versioned in nothing: it is the flock on .agent-kb/akb.lock, it stays held
// while a nested acquisition is released, and releasing it frees the lock.
func TestStoreUnversionedBaseLockLocksItsOwnLockFile(t *testing.T) {
	kbRoot := t.TempDir()
	writeKBConfig(t, kbRoot, config.VersioningNone)

	store, err := OpenStore(kbRoot, false)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	lockPath := path.LockPath(kbRoot)
	lock, err := store.Lock()
	if err != nil {
		t.Fatalf("Lock: %v", err)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("base lock file %s: %v", lockPath, err)
	}
	if lockFree(t, lockPath) {
		t.Fatal("the base lock is not held after Lock")
	}

	nested, err := store.Lock()
	if err != nil {
		t.Fatalf("nested Lock: %v", err)
	}
	nested.Release()
	if lockFree(t, lockPath) {
		t.Fatal("releasing the nested acquisition dropped the held lock")
	}

	lock.Release()
	if !lockFree(t, lockPath) {
		t.Fatal("the base lock is still held after releasing the outer acquisition")
	}
}

// TestStoreGitModeLockLocksTheRepository asserts a git-versioned base locks the
// repository that hosts it, not a base-local lock file.
func TestStoreGitModeLockLocksTheRepository(t *testing.T) {
	repo := initGitRepo(t)
	writeKBConfig(t, repo, config.VersioningGit)

	store, err := OpenStore(repo, false)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	lock, err := store.Lock()
	if err != nil {
		t.Fatalf("Lock: %v", err)
	}
	repoLock := filepath.Join(repo, ".git", "akb.lock")
	if lockFree(t, repoLock) {
		t.Fatalf("the repository lock %s is not held after Lock", repoLock)
	}
	baseLock := path.LockPath(repo)
	if _, err := os.Stat(baseLock); !os.IsNotExist(err) {
		t.Errorf("a git-versioned base wrote a base lock file (stat error: %v)", err)
	}

	lock.Release()
	if !lockFree(t, repoLock) {
		t.Fatal("the repository lock is still held after releasing it")
	}
}

// TestStoreGitModeWriteCommits asserts the Store's write path commits the page
// of a git-versioned base.
func TestStoreGitModeWriteCommits(t *testing.T) {
	repo := initGitRepo(t)
	writeKBConfig(t, repo, config.VersioningGit)
	commitKBConfig(t, repo)

	store, err := OpenStore(repo, false)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	const rel = "kb/notes/page.md"
	if err := store.Write(context.Background(), filepath.Join(repo, rel), []byte("page body\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if subject := mustGitIn(t, repo, "log", "-1", "--format=%s"); subject != "akb: write "+rel {
		t.Errorf("HEAD subject = %q, want %q", subject, "akb: write "+rel)
	}
	if status := mustGitIn(t, repo, "status", "--porcelain"); status != "" {
		t.Errorf("repository is not clean:\n%s", status)
	}
}

// TestStoreCommitRecordsTheGivenPaths asserts the Store's commit funnel records
// exactly the paths it is handed, which is how the raw and template commands
// commit files they wrote themselves.
func TestStoreCommitRecordsTheGivenPaths(t *testing.T) {
	repo := initGitRepo(t)
	writeKBConfig(t, repo, config.VersioningGit)
	commitKBConfig(t, repo)

	store, err := OpenStore(repo, false)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	content := filepath.Join(repo, "raw", "data.csv")
	if err := os.MkdirAll(filepath.Dir(content), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(content, []byte("a,b\n"), 0600); err != nil { //nolint:gosec // test writing into its temp repository
		t.Fatal(err)
	}
	if err := store.Commit("akb: raw write data.csv", "raw/data.csv"); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	recorded := mustGitIn(t, repo, "show", "--name-only", "--format=", "HEAD")
	if strings.TrimSpace(recorded) != "raw/data.csv" {
		t.Errorf("commit recorded %q, want %q", recorded, "raw/data.csv")
	}
	if status := mustGitIn(t, repo, "status", "--porcelain"); status != "" {
		t.Errorf("repository is not clean:\n%s", status)
	}
}

// TestStoreCommitRejectsConflictedMerge asserts the commit funnel refuses to
// commit while the repository carries a conflicted merge.
func TestStoreCommitRejectsConflictedMerge(t *testing.T) {
	repo, _ := repoWithInProgressMerge(t, true)
	writeKBConfig(t, repo, config.VersioningGit)

	store, err := OpenStore(repo, false)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	err = store.Commit("akb: write kb/notes/page.md", "kb/notes/page.md")
	if err == nil {
		t.Fatal("Commit succeeded while the repository carried a merge conflict")
	}
	if !strings.Contains(err.Error(), "merge conflict") {
		t.Errorf("error = %q, want a merge conflict report", err)
	}
}

// mergeInProgressReport returns the report of a merge that is in progress in
// repo and changed file, as the merge preflight words it.
func mergeInProgressReport(t *testing.T, repo, file string) string {
	t.Helper()

	return fmt.Sprintf("merge in progress in repository %s (outside the knowledge base, in: %s): "+
		"akb commits are blocked until it is completed or aborted. This merge belongs to the host project "+
		"— do not resolve it from the KB; retry later or surface to the user.",
		mustGitIn(t, repo, "rev-parse", "--show-toplevel"), file)
}

// TestGitProviderWriteRejectsCleanInProgressMerge asserts the merge preflight of
// the provider refuses a clean merge that is still in progress: git would refuse
// the partial commit that follows, so the report of the merge replaces the raw
// git failure.
func TestGitProviderWriteRejectsCleanInProgressMerge(t *testing.T) {
	repo, changed := repoWithInProgressMerge(t, false)

	provider := NewGitProvider(repo, false)
	page := filepath.Join(repo, "kb", "notes", "page.md")
	err := provider.Write(context.Background(), page, []byte("page body\n"))
	if err == nil {
		t.Fatal("Write succeeded while the repository carried a merge in progress")
	}
	if want := mergeInProgressReport(t, repo, changed); err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
	if !errors.Is(err, ErrMergeInProgress) {
		t.Errorf("error %v does not wrap ErrMergeInProgress", err)
	}
	if _, err := os.Stat(page); !os.IsNotExist(err) {
		t.Errorf("the page was written while the merge report was due (stat error: %v)", err)
	}
}

// TestStoreCommitRejectsCleanInProgressMerge asserts the commit funnel refuses a
// clean merge in progress and names the merge's path, so the agent sees the host
// project's merge instead of a git failure it cannot place.
func TestStoreCommitRejectsCleanInProgressMerge(t *testing.T) {
	repo, changed := repoWithInProgressMerge(t, false)
	writeKBConfig(t, repo, config.VersioningGit)

	store, err := OpenStore(repo, false)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	err = store.Commit("akb: write kb/notes/page.md", "kb/notes/page.md")
	if err == nil {
		t.Fatal("Commit succeeded while the repository carried a merge in progress")
	}
	if want := mergeInProgressReport(t, repo, changed); err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}
