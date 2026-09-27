// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/path"
)

// writeBaseIdentity records the commit identity of the base rooted at kbRoot in
// its akb.yaml.
func writeBaseIdentity(t *testing.T, kbRoot, author, email string) {
	t.Helper()

	writeKBConfig(t, kbRoot, config.VersioningGit)
	configPath := path.ConfigPath(kbRoot)
	data, err := os.ReadFile(configPath) //nolint:gosec // test reading its own base config
	if err != nil {
		t.Fatal(err)
	}
	if author != "" {
		data = append(data, []byte("git-author: "+author+"\n")...)
	}
	if email != "" {
		data = append(data, []byte("git-email: "+email+"\n")...)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil { //nolint:gosec // test writing its own base config
		t.Fatal(err)
	}
}

func TestResolveIdentity(t *testing.T) {
	isolateGitConfig(t)

	t.Run("supplies the identity of the base", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")
		writeBaseIdentity(t, repo, "reviewer", "reviewer@example.com")

		identity, err := ResolveIdentity(repo)
		if err != nil {
			t.Fatalf("ResolveIdentity: %v", err)
		}
		if identity == nil || *identity != (Identity{Name: "reviewer", Email: "reviewer@example.com"}) {
			t.Errorf("identity = %v, want the identity of the base", identity)
		}
	})

	t.Run("prefers the environment over the base", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")
		writeBaseIdentity(t, repo, "reviewer", "reviewer@example.com")
		t.Setenv(AuthorNameEnv, "agent-7")
		t.Setenv(AuthorEmailEnv, "agent-7@example.com")

		identity, err := ResolveIdentity(repo)
		if err != nil {
			t.Fatalf("ResolveIdentity: %v", err)
		}
		if identity == nil || *identity != (Identity{Name: "agent-7", Email: "agent-7@example.com"}) {
			t.Errorf("identity = %v, want the environment identity", identity)
		}
	})

	t.Run("ignores an empty environment value", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		writeBaseIdentity(t, repo, "reviewer", "reviewer@example.com")
		t.Setenv(AuthorNameEnv, "")

		identity, err := ResolveIdentity(repo)
		if err != nil {
			t.Fatalf("ResolveIdentity: %v", err)
		}
		if identity == nil || *identity != (Identity{Name: "reviewer", Email: "reviewer@example.com"}) {
			t.Errorf("identity = %v, want the identity of the base", identity)
		}
	})

	t.Run("fills a field the base leaves out from git", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")
		writeBaseIdentity(t, repo, "reviewer", "")

		identity, err := ResolveIdentity(repo)
		if err != nil {
			t.Fatalf("ResolveIdentity: %v", err)
		}
		if identity == nil || *identity != (Identity{Name: "reviewer", Email: "ada@example.com"}) {
			t.Errorf("identity = %v, want the base name with git's email", identity)
		}
	})

	t.Run("lets git resolve the identity when akb supplies none", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")

		identity, err := ResolveIdentity(repo)
		if err != nil {
			t.Fatalf("ResolveIdentity: %v", err)
		}
		if identity != nil {
			t.Errorf("identity = %v, want nil so that git resolves it", identity)
		}
	})

	t.Run("reports an identity no source names", func(t *testing.T) {
		forceUnsetUser(t)
		repo := initRepoWithoutIdentity(t)

		identity, err := ResolveIdentity(repo)
		if !errors.Is(err, ErrNoCommitIdentity) {
			t.Fatalf("ResolveIdentity = %v, %v, want %v", identity, err, ErrNoCommitIdentity)
		}
		for _, remedy := range []string{AuthorNameEnv, AuthorEmailEnv, "git-author", "git-email"} {
			if !strings.Contains(err.Error(), remedy) {
				t.Errorf("error %q does not name the remedy %q", err, remedy)
			}
		}
	})

	t.Run("refuses an identity git cannot attribute a committer for", func(t *testing.T) {
		forceUnsetUser(t)
		repo := initRepoWithoutIdentity(t)
		t.Setenv("GIT_AUTHOR_NAME", "escape")
		t.Setenv("GIT_AUTHOR_EMAIL", "escape@example.com")

		if _, err := ResolveIdentity(repo); !errors.Is(err, ErrNoCommitIdentity) {
			t.Errorf("ResolveIdentity = %v, want %v for an identity only the author carries", err, ErrNoCommitIdentity)
		}
	})
}

func TestResolveInitIdentity(t *testing.T) {
	isolateGitConfig(t)

	t.Run("prefers the flags", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")
		t.Setenv(AuthorNameEnv, "agent-7")
		t.Setenv(AuthorEmailEnv, "agent-7@example.com")

		identity, source := ResolveInitIdentity(repo, "reviewer", "reviewer@example.com")
		if want := (Identity{Name: "reviewer", Email: "reviewer@example.com"}); identity != want {
			t.Errorf("identity = %v, want %v", identity, want)
		}
		if source != IdentityFromFlag {
			t.Errorf("source = %d, want %d", source, IdentityFromFlag)
		}
	})

	t.Run("prefers the environment over git config", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")
		t.Setenv(AuthorNameEnv, "agent-7")
		t.Setenv(AuthorEmailEnv, "agent-7@example.com")

		identity, source := ResolveInitIdentity(repo, "", "")
		if want := (Identity{Name: "agent-7", Email: "agent-7@example.com"}); identity != want {
			t.Errorf("identity = %v, want %v", identity, want)
		}
		if source != IdentityFromEnv {
			t.Errorf("source = %d, want %d", source, IdentityFromEnv)
		}
	})

	t.Run("reads the identity git configures", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")

		identity, source := ResolveInitIdentity(repo, "", "")
		if want := (Identity{Name: "ada", Email: "ada@example.com"}); identity != want {
			t.Errorf("identity = %v, want %v", identity, want)
		}
		if source != IdentityFromGitConfig {
			t.Errorf("source = %d, want %d", source, IdentityFromGitConfig)
		}
	})

	t.Run("falls back to the default", func(t *testing.T) {
		forceUnsetUser(t)
		repo := initRepoWithoutIdentity(t)

		identity, source := ResolveInitIdentity(repo, "", "")
		if identity != DefaultIdentity {
			t.Errorf("identity = %v, want %v", identity, DefaultIdentity)
		}
		if source != IdentityFromDefault {
			t.Errorf("source = %d, want %d", source, IdentityFromDefault)
		}
	})

	t.Run("treats a partial configuration as naming no identity", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")
		forceUnsetUser(t)

		identity, source := ResolveInitIdentity(repo, "", "")
		if identity != DefaultIdentity {
			t.Errorf("identity = %v, want %v", identity, DefaultIdentity)
		}
		if source != IdentityFromDefault {
			t.Errorf("source = %d, want %d", source, IdentityFromDefault)
		}
	})

	t.Run("fills a field from the next source down", func(t *testing.T) {
		repo := initRepoWithoutIdentity(t)
		mustGitIn(t, repo, "config", "user.name", "ada")
		mustGitIn(t, repo, "config", "user.email", "ada@example.com")

		identity, source := ResolveInitIdentity(repo, "reviewer", "")
		if want := (Identity{Name: "reviewer", Email: "ada@example.com"}); identity != want {
			t.Errorf("identity = %v, want %v", identity, want)
		}
		if source != IdentityFromFlag {
			t.Errorf("source = %d, want %d", source, IdentityFromFlag)
		}
	})
}

func TestGitProviderCommitWithEnvironmentIdentity(t *testing.T) {
	isolateGitConfig(t)
	forceUnsetUser(t)
	repo := initRepoWithoutIdentity(t)
	t.Setenv(AuthorNameEnv, "agent-7")
	t.Setenv(AuthorEmailEnv, "agent-7@example.com")

	provider := NewGitProvider(repo, false)
	path := filepath.Join(repo, "kb", "env-identity.md")
	if err := provider.Write(context.Background(), path, []byte("env identity test")); err != nil {
		t.Fatalf("write with an environment identity: %v", err)
	}

	if author := commitAuthor(t, repo); author != "agent-7 <agent-7@example.com>" {
		t.Errorf("author = %q, want %q", author, "agent-7 <agent-7@example.com>")
	}
	if committer := mustGitIn(t, repo, "log", "-1", "--format=%cn <%ce>"); committer != "agent-7 <agent-7@example.com>" {
		t.Errorf("committer = %q, want %q", committer, "agent-7 <agent-7@example.com>")
	}
	assertRepoIdentityUnset(t, repo)
}

func TestGitProviderCommitWithBaseIdentity(t *testing.T) {
	isolateGitConfig(t)
	forceUnsetUser(t)
	repo := initRepoWithoutIdentity(t)
	writeBaseIdentity(t, repo, "reviewer", "reviewer@example.com")

	provider := NewGitProvider(repo, false)
	path := filepath.Join(repo, "kb", "base-identity.md")
	if err := provider.Write(context.Background(), path, []byte("base identity test")); err != nil {
		t.Fatalf("write with the identity of the base, which git does not configure: %v", err)
	}

	if author := commitAuthor(t, repo); author != "reviewer <reviewer@example.com>" {
		t.Errorf("author = %q, want %q", author, "reviewer <reviewer@example.com>")
	}
	if committer := mustGitIn(t, repo, "log", "-1", "--format=%cn <%ce>"); committer != "reviewer <reviewer@example.com>" {
		t.Errorf("committer = %q, want %q", committer, "reviewer <reviewer@example.com>")
	}
	assertRepoIdentityUnset(t, repo)
}

// TestGitProviderKeepsTheExportedIdentity asserts git's own identity stays in
// charge when akb supplies none: an identity the caller exports through
// GIT_AUTHOR_* and GIT_COMMITTER_* is the one the commit records.
func TestGitProviderKeepsTheExportedIdentity(t *testing.T) {
	isolateGitConfig(t)
	forceUnsetUser(t)
	repo := initRepoWithoutIdentity(t)
	t.Setenv("GIT_AUTHOR_NAME", "escape")
	t.Setenv("GIT_AUTHOR_EMAIL", "escape@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "escape")
	t.Setenv("GIT_COMMITTER_EMAIL", "escape@example.com")

	provider := NewGitProvider(repo, false)
	path := filepath.Join(repo, "kb", "exported-identity.md")
	if err := provider.Write(context.Background(), path, []byte("exported identity test")); err != nil {
		t.Fatalf("write with an exported identity: %v", err)
	}

	if author := commitAuthor(t, repo); author != "escape <escape@example.com>" {
		t.Errorf("author = %q, want %q", author, "escape <escape@example.com>")
	}
}

// TestGitProviderWithoutIdentityLeavesNothingBehind asserts the identity
// preflight runs before the file is written: a base whose identity cannot
// resolve keeps its worktree untouched.
func TestGitProviderWithoutIdentityLeavesNothingBehind(t *testing.T) {
	isolateGitConfig(t)
	forceUnsetUser(t)
	repo := initRepoWithoutIdentity(t)

	provider := NewGitProvider(repo, false)
	path := filepath.Join(repo, "kb", "unattributable.md")
	if err := provider.Write(context.Background(), path, []byte("unattributable")); !errors.Is(err, ErrNoCommitIdentity) {
		t.Fatalf("write without a resolvable identity = %v, want %v", err, ErrNoCommitIdentity)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("the page was written before the identity could be resolved (stat error: %v)", err)
	}
	if status := mustGitIn(t, repo, "status", "--porcelain"); status != "" {
		t.Errorf("the refused write left changes behind:\n%s", status)
	}
	assertRepoIdentityUnset(t, repo)
}

// TestGitProviderNoCommitNeedsNoIdentity asserts an invocation that leaves the
// commit to its caller writes its file even where no commit identity resolves.
func TestGitProviderNoCommitNeedsNoIdentity(t *testing.T) {
	isolateGitConfig(t)
	forceUnsetUser(t)
	repo := initRepoWithoutIdentity(t)

	provider := NewGitProvider(repo, true)
	path := filepath.Join(repo, "kb", "no-commit.md")
	if err := provider.Write(context.Background(), path, []byte("no commit test")); err != nil {
		t.Fatalf("write without a commit: %v", err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // test reading a file in its temp repository
	if err != nil {
		t.Fatalf("read page: %v", err)
	}
	if string(data) != "no commit test" {
		t.Errorf("page content = %q, want %q", string(data), "no commit test")
	}
}

func TestStorePreflightReportsTheIdentity(t *testing.T) {
	isolateGitConfig(t)
	forceUnsetUser(t)
	repo := initRepoWithoutIdentity(t)
	writeKBConfig(t, repo, config.VersioningGit)

	store, err := OpenStore(repo, false)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	if err := store.Preflight(); !errors.Is(err, ErrNoCommitIdentity) {
		t.Fatalf("Preflight = %v, want %v", err, ErrNoCommitIdentity)
	}
	if err := store.Write(context.Background(), filepath.Join(repo, "kb", "page.md"), []byte("body\n")); !errors.Is(err, ErrNoCommitIdentity) {
		t.Fatalf("Write = %v, want %v", err, ErrNoCommitIdentity)
	}
	if _, err := os.Stat(filepath.Join(repo, "kb", "page.md")); !os.IsNotExist(err) {
		t.Errorf("the page was written before the identity could be resolved (stat error: %v)", err)
	}
}
