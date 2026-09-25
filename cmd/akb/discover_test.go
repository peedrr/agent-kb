// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// discoverTestRoot isolates a discovery test in one temporary tree: home is
// the boundary the scan stops at and work is the directory it runs from, so
// the scan cannot reach beyond the fixtures. It returns work.
func discoverTestRoot(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	work := filepath.Join(home, "work")
	if err := os.MkdirAll(work, 0750); err != nil {
		t.Fatalf("create work directory: %v", err)
	}
	t.Setenv("HOME", home)
	t.Chdir(work)
	return work
}

// makeDiscoverKB creates a knowledge base in dir, with a config when one is
// given.
func makeDiscoverKB(t *testing.T, dir, config string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(dir, ".agent-kb"), 0750); err != nil {
		t.Fatalf("create .agent-kb directory: %v", err)
	}
	if config == "" {
		return
	}
	if err := os.WriteFile(filepath.Join(dir, ".agent-kb", "akb.yaml"), []byte(config), 0600); err != nil {
		t.Fatalf("write akb.yaml: %v", err)
	}
}

// resetDiscoverJSON clears the --json flag state for one test and restores it
// afterwards, so a test that sets the flag cannot leak into the next one.
func resetDiscoverJSON(t *testing.T) {
	t.Helper()

	original := discoverJSON
	discoverJSON = false
	t.Cleanup(func() { discoverJSON = original })
}

func TestDiscoverCommandListsNearbyBases(t *testing.T) {
	work := discoverTestRoot(t)
	resetDiscoverJSON(t)

	makeDiscoverKB(t, filepath.Join(work, "project"), "name: project\n")
	makeDiscoverKB(t, filepath.Join(work, "project-kb"), "name: project-kb\ndescription: project notes\n")

	out, err := captureOutput(func() error { return runDiscover(nil, nil) })
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}

	want := "discovered knowledge bases (nearest first):\n" +
		"  project  " + filepath.Join(work, "project") + "\n" +
		"  project-kb  " + filepath.Join(work, "project-kb") + "  project notes\n" +
		"\n" +
		"select one with --kb <path> or AKB_KB=<path>\n"
	if out != want {
		t.Errorf("discover output = %q, want %q", out, want)
	}
}

func TestDiscoverCommandJSONShape(t *testing.T) {
	work := discoverTestRoot(t)
	resetDiscoverJSON(t)

	makeDiscoverKB(t, filepath.Join(work, "project"), "name: project\n")
	makeDiscoverKB(t, filepath.Join(work, "described"), "name: described\ndescription: has a description\n")

	discoverJSON = true
	out, err := captureOutput(func() error { return runDiscover(nil, nil) })
	if err != nil {
		t.Fatalf("discover --json failed: %v", err)
	}

	want := `[{"name":"described","path":"` + filepath.Join(work, "described") + `","description":"has a description"},` +
		`{"name":"project","path":"` + filepath.Join(work, "project") + `"}]` + "\n"
	if out != want {
		t.Errorf("discover --json output = %q, want %q", out, want)
	}
}

func TestDiscoverCommandScansGivenDirectory(t *testing.T) {
	work := discoverTestRoot(t)
	resetDiscoverJSON(t)

	makeDiscoverKB(t, filepath.Join(work, "docs", "kb"), "name: nested\n")

	out, err := captureOutput(func() error {
		return runDiscover(nil, []string{filepath.Join(work, "docs")})
	})
	if err != nil {
		t.Fatalf("discover <dir> failed: %v", err)
	}

	want := "discovered knowledge bases (nearest first):\n" +
		"  nested  " + filepath.Join(work, "docs", "kb") + "\n" +
		"\n" +
		"select one with --kb <path> or AKB_KB=<path>\n"
	if out != want {
		t.Errorf("discover output = %q, want %q", out, want)
	}
}

func TestDiscoverCommandEmptyNeighborhood(t *testing.T) {
	home := t.TempDir()
	work := filepath.Join(home, "work", "empty")
	if err := os.MkdirAll(work, 0750); err != nil {
		t.Fatalf("create work directory: %v", err)
	}
	t.Setenv("HOME", home)
	t.Chdir(work)
	resetDiscoverJSON(t)

	out, err := captureOutput(func() error { return runDiscover(nil, nil) })
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}
	if want := "no knowledge bases found near " + work + "\n"; out != want {
		t.Errorf("discover output = %q, want %q", out, want)
	}

	discoverJSON = true
	out, err = captureOutput(func() error { return runDiscover(nil, nil) })
	if err != nil {
		t.Fatalf("discover --json failed: %v", err)
	}
	if out != "[]\n" {
		t.Errorf("discover --json output = %q, want %q", out, "[]\n")
	}
}

func TestDiscoverCommandRejectsMissingDirectory(t *testing.T) {
	discoverTestRoot(t)
	resetDiscoverJSON(t)

	err := runDiscover(nil, []string{filepath.Join("does", "not", "exist")})

	var usageErr *usageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("expected a usage error, got %T: %v", err, err)
	}
	if !strings.Contains(usageErr.Error(), "does/not/exist") {
		t.Errorf("usage error %q does not name the missing directory", usageErr.Error())
	}
}

func TestDiscoverCommandIsRegisteredReadOnly(t *testing.T) {
	cmd := resolveCommandPath(t, "akb discover")
	if cmd.RunE == nil {
		t.Fatal("discover has no RunE")
	}
	if mutatingCommands[cmd.CommandPath()] {
		t.Error("discover is listed as a mutating command, so it would report a knowledge base it never resolves")
	}
}
