// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/peedrr/agent-kb/internal/manifest"
	"github.com/peedrr/agent-kb/internal/storage"
)

// TestRawDeleteRefusesBeforeRemovingWithoutCommitIdentity pins that the raw
// delete path runs the commit-identity preflight before its first mutation: an
// identity no source can name is refused with the sentinel, and the raw file
// and its manifest entry stay as they were.
func TestRawDeleteRefusesBeforeRemovingWithoutCommitIdentity(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	mustGitInDir(t, kbRoot, "config", "--unset", "user.name")
	mustGitInDir(t, kbRoot, "config", "--unset", "user.email")
	useNoCommitIdentityEnv(t)

	origNoCommit := noCommit
	noCommit = false
	t.Cleanup(func() { noCommit = origNoCommit })

	const relPath = "data/config.json"
	fullPath := filepath.Join(kbRoot, "raw", filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte("raw content"), 0600); err != nil {
		t.Fatal(err)
	}
	sha256hash, err := manifest.ComputeSHA256(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.NewManager(kbRoot).AddEntry(relPath, sha256hash); err != nil {
		t.Fatal(err)
	}

	err = runRawDelete(nil, []string{relPath})
	if err == nil {
		t.Fatal("expected the raw delete to be refused without a commit identity")
	}
	if !errors.Is(err, storage.ErrNoCommitIdentity) {
		t.Errorf("refusal = %v, want the commit-identity sentinel", err)
	}

	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("raw file gone after the identity refusal: %v", err)
	}
	_, found, err := manifest.NewManager(kbRoot).FindByFilename(relPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !found {
		t.Error("manifest entry gone after the identity refusal")
	}
}
