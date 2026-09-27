// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package storage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/path"
)

// kbLockFileName names the lock file of a base that is not versioned in git. It
// sits in the base's state directory, next to the config it belongs to.
const kbLockFileName = "akb.lock"

// Mode is the versioning mode of a knowledge base: the value of its akb.yaml
// `versioning` key.
type Mode string

const (
	// ModeGit versions the base in git: every mutation is committed.
	ModeGit Mode = config.VersioningGit
	// ModeNone versions nothing: mutations write their files and stop.
	ModeNone Mode = config.VersioningNone
)

// ParseMode maps a `versioning` config value to a Mode. An empty value — a
// config that predates the key — means git, the mode those bases were created
// in. Any other value is rejected instead of read as git, so a base whose mode
// is misspelled never commits under the wrong assumption.
func ParseMode(value string) (Mode, error) {
	switch value {
	case "", config.VersioningGit:
		return ModeGit, nil
	case config.VersioningNone:
		return ModeNone, nil
	default:
		return "", fmt.Errorf("unknown versioning mode %q: use %q or %q", value, config.VersioningGit, config.VersioningNone)
	}
}

// newProvider returns the provider the base in mode reads and writes through: a
// git-tracking provider for a base versioned in git, a plain filesystem
// provider for one that is not. Selection happens here, at construction time;
// the Provider interface carries no versioning semantics.
func newProvider(kbRoot string, mode Mode, noCommit bool) Provider {
	if mode == ModeNone {
		return NewFilesystemProvider(kbRoot)
	}
	return NewGitProvider(kbRoot, noCommit)
}

// committing is the provider capability the Store needs besides Provider: a
// provider that records what it writes, under a caller-supplied message.
type committing interface {
	WriteWithCommitMsg(ctx context.Context, path string, data []byte, commitMsg string) error
}

// Store is the storage of one knowledge base: the provider its versioning mode
// selects, plus the lock, merge preflight and commit steps its commands go
// through. Commands work through a Store rather than a provider directly, so a
// base that is not versioned in git runs the same command code as a
// git-versioned one and simply writes its files without committing them.
type Store struct {
	kbRoot   string
	mode     Mode
	noCommit bool
	provider Provider
}

// OpenStore returns the storage of the base rooted at kbRoot, with the
// versioning mode its akb.yaml configures: a base versioned in git writes and
// commits through git, a base versioned in nothing writes files only, with no
// git invocation at all. noCommit is the base-wide --no-commit flag; it has no
// effect on a base that is not versioned in git, which never commits.
func OpenStore(kbRoot string, noCommit bool) (*Store, error) {
	cfg, err := config.Load(path.ConfigPath(kbRoot))
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	mode, err := ParseMode(cfg.Versioning)
	if err != nil {
		return nil, err
	}
	return &Store{
		kbRoot:   kbRoot,
		mode:     mode,
		noCommit: noCommit,
		provider: newProvider(kbRoot, mode, noCommit),
	}, nil
}

// Mode returns the versioning mode of the base.
func (s *Store) Mode() Mode {
	return s.mode
}

// Preflight checks what a mutation of the base needs before anything is
// written: a repository without a merge in progress, which would refuse the
// commit the mutation ends in. A base that is not versioned in git has nothing
// to check and no commit to protect. Commands call it before their first
// mutation; the commit steps themselves check again under the lock.
func (s *Store) Preflight() error {
	if s.mode == ModeNone {
		return nil
	}
	return checkMergeState(s.kbRoot)
}

// Lock acquires the exclusive lock of the base's mutation sequence and returns
// the handle that releases it. A base versioned in git locks the repository that
// hosts it — every base inside that repository serializes against the others —
// while a base versioned in nothing locks its own .agent-kb/akb.lock, which
// resolves no git state. An acquisition nested in a lock the process already
// holds shares it, so a command may lock around its whole mutation and let the
// steps inside it lock again.
func (s *Store) Lock() (*RepoLock, error) {
	if s.mode == ModeNone {
		return lockFile(filepath.Join(path.StateDir(s.kbRoot), kbLockFileName))
	}
	return LockRepo(s.kbRoot)
}

// Write writes data to path and, in a base versioned in git, commits it as
// `akb: write <path>`.
func (s *Store) Write(ctx context.Context, path string, data []byte) error {
	if err := s.Preflight(); err != nil {
		return err
	}
	//nolint:wrapcheck // the provider's error already names the file and the step that failed
	return s.provider.Write(ctx, path, data)
}

// WriteWithCommitMsg writes data to path and, in a base versioned in git,
// commits it with commitMsg.
func (s *Store) WriteWithCommitMsg(ctx context.Context, path string, data []byte, commitMsg string) error {
	if err := s.Preflight(); err != nil {
		return err
	}
	if committer, ok := s.provider.(committing); ok {
		//nolint:wrapcheck // the provider's error already names the file and the step that failed
		return committer.WriteWithCommitMsg(ctx, path, data, commitMsg)
	}
	//nolint:wrapcheck // the provider's error already names the file and the step that failed
	return s.provider.Write(ctx, path, data)
}

// Read returns the contents of path.
func (s *Store) Read(ctx context.Context, path string) ([]byte, error) {
	//nolint:wrapcheck // the provider's error already names the file and the step that failed
	return s.provider.Read(ctx, path)
}

// Delete removes path and, in a base versioned in git, commits the removal as
// `akb: delete <path>`.
func (s *Store) Delete(ctx context.Context, path string) error {
	if err := s.Preflight(); err != nil {
		return err
	}
	//nolint:wrapcheck // the provider's error already names the file and the step that failed
	return s.provider.Delete(ctx, path)
}

// Exists reports whether path exists.
func (s *Store) Exists(ctx context.Context, path string) (bool, error) {
	//nolint:wrapcheck // the provider's error already names the file and the step that failed
	return s.provider.Exists(ctx, path)
}

// List returns the files with the given extension under dir.
func (s *Store) List(ctx context.Context, dir string, ext string) ([]string, error) {
	//nolint:wrapcheck // the provider's error already names the file and the step that failed
	return s.provider.List(ctx, dir, ext)
}

// StageFiles stages the given base-relative paths for the commit that follows
// them. It is the staging step of the commit funnel, so it takes the same
// preflight: a base that is not versioned in git stages nothing.
func (s *Store) StageFiles(paths ...string) error {
	if s.mode == ModeNone {
		return nil
	}
	if err := s.Preflight(); err != nil {
		return err
	}
	return StageFiles(s.kbRoot, paths...)
}

// Commit stages and commits exactly the given base-relative paths with commitMsg.
// A base that is not versioned in git keeps no commit history, so the call
// records nothing there — which also makes --no-commit a no-op in that mode.
func (s *Store) Commit(commitMsg string, paths ...string) error {
	if s.mode == ModeNone || s.noCommit {
		return nil
	}
	if err := s.Preflight(); err != nil {
		return err
	}
	return CommitFiles(s.kbRoot, commitMsg, paths...)
}

// NothingToCommit reports whether committing the given base-relative paths
// would record nothing. A base that is not versioned in git has recorded
// nothing and will record nothing, so the answer is always yes.
func (s *Store) NothingToCommit(paths ...string) (bool, error) {
	if s.mode == ModeNone {
		return true, nil
	}
	return NothingToCommit(s.kbRoot, paths...)
}
