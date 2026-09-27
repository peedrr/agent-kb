// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/peedrr/agent-kb/internal/storage"
)

// TestRawSyncNoChangesNeedsNoCommitIdentity pins that a sync that finds nothing
// to reconcile returns before the commit preflight: no commit identity is
// resolvable, yet the no-op succeeds, reports itself, and leaves the manifest
// file unwritten.
func TestRawSyncNoChangesNeedsNoCommitIdentity(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	mustGitInDir(t, kbRoot, "config", "--unset", "user.name")
	mustGitInDir(t, kbRoot, "config", "--unset", "user.email")
	useNoCommitIdentityEnv(t)

	origNoCommit := noCommit
	noCommit = false
	t.Cleanup(func() { noCommit = origNoCommit })

	out, err := captureOutput(func() error { return runRawSync(nil, nil) })
	if err != nil {
		t.Fatalf("raw sync with nothing to sync: %v", err)
	}
	if want := "No changes to sync.\n"; out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
	if _, err := os.Stat(filepath.Join(kbRoot, "raw", "files.log")); !os.IsNotExist(err) {
		t.Errorf("a no-op sync wrote the manifest (stat error: %v)", err)
	}
}

// TestRawSyncRefusesWithoutCommitIdentityWhenThereIsAChange pins that the
// preflight still guards a sync that has something to record: an identity no
// source can name is refused with the sentinel, and the manifest stays
// unwritten.
func TestRawSyncRefusesWithoutCommitIdentityWhenThereIsAChange(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	mustGitInDir(t, kbRoot, "config", "--unset", "user.name")
	mustGitInDir(t, kbRoot, "config", "--unset", "user.email")
	useNoCommitIdentityEnv(t)

	origNoCommit := noCommit
	noCommit = false
	t.Cleanup(func() { noCommit = origNoCommit })

	if err := os.WriteFile(filepath.Join(kbRoot, "raw", "data.csv"), []byte("hello\n"), 0600); err != nil {
		t.Fatal(err)
	}

	err := runRawSync(nil, nil)
	if err == nil {
		t.Fatal("raw sync with a change succeeded without a commit identity")
	}
	if !errors.Is(err, storage.ErrNoCommitIdentity) {
		t.Errorf("refusal = %v, want the commit-identity sentinel", err)
	}
	if _, statErr := os.Stat(filepath.Join(kbRoot, "raw", "files.log")); !os.IsNotExist(statErr) {
		t.Errorf("manifest written after the identity refusal (stat error: %v)", statErr)
	}
}
