// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/path"
)

// resetKBFlag clears the --kb flag state for one test and restores it
// afterwards, so a test that drives the flag cannot leak into the next one.
func resetKBFlag(t *testing.T) {
	t.Helper()

	orig := kbFlag
	kbFlag = ""
	t.Cleanup(func() { kbFlag = orig })
}

// captureStderr runs fn with os.Stderr redirected to a pipe and returns the
// text fn wrote.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stderr
	os.Stderr = w
	fn()
	os.Stderr = old

	if err := w.Close(); err != nil {
		t.Fatalf("close stderr pipe: %v", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stderr pipe: %v", err)
	}
	return string(data)
}

// resolveCommandPath walks the registered command tree from RootCmd down a
// space-separated command path and returns the command it names, failing the
// test when a segment matches no registered command.
func resolveCommandPath(t *testing.T, commandPath string) *cobra.Command {
	t.Helper()

	segments := strings.Fields(commandPath)
	if len(segments) == 0 || segments[0] != RootCmd.Name() {
		t.Fatalf("command path %q does not start at the root command %q", commandPath, RootCmd.Name())
	}

	cmd := RootCmd
	for _, segment := range segments[1:] {
		var next *cobra.Command
		for _, sub := range cmd.Commands() {
			if sub.Name() == segment {
				next = sub
				break
			}
		}
		if next == nil {
			t.Fatalf("command path %q: no registered command named %q under %q", commandPath, segment, cmd.CommandPath())
		}
		cmd = next
	}
	return cmd
}

func TestKBFlagSelectsKnowledgeBase(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)
	resetKBFlag(t)

	// The environment points elsewhere, so only the flag can make status
	// succeed.
	t.Setenv(path.KBEnvVar, filepath.Join(t.TempDir(), "elsewhere"))

	RootCmd.SetArgs([]string{"status", "--kb", kbRoot})
	out, err := captureOutput(Execute)
	if err != nil {
		t.Fatalf("status --kb failed: %v", err)
	}
	if !strings.Contains(out, "Name: write-test") {
		t.Errorf("status output = %q, want the name of the flag's knowledge base", out)
	}
}

func TestStatusWithoutKBSelectionClassifiesAsUsage(t *testing.T) {
	home := t.TempDir()
	work := filepath.Join(home, "work")
	if err := os.MkdirAll(work, 0750); err != nil {
		t.Fatalf("create work directory: %v", err)
	}
	t.Setenv("HOME", home)
	t.Chdir(work)
	t.Setenv(path.KBEnvVar, "")
	resetKBFlag(t)

	code, report := classifiesInvocation(t, "status")
	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	assertUsageReport(t, report, "resolve knowledge base: no knowledge base selected: pass --kb <path> or set the AKB_KB environment variable")
}

func TestRemovedCommandsClassifyAsUsage(t *testing.T) {
	for _, name := range []string{"use", "registry"} {
		t.Run(name, func(t *testing.T) {
			code, report := classifiesInvocation(t, name)
			if code != exitFault {
				t.Errorf("exit code = %d, want %d", code, exitFault)
			}
			if !strings.Contains(report, `unknown command "`+name+`"`) {
				t.Errorf("report = %q, want an unknown-command report for %q", report, name)
			}
		})
	}
}

func TestKBIdentityLineNamesTheSelectedBase(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	line, err := kbIdentityLine(kbRoot)
	if err != nil {
		t.Fatalf("kbIdentityLine: %v", err)
	}
	if want := "kb: write-test (" + kbRoot + ")"; line != want {
		t.Errorf("kbIdentityLine = %q, want %q", line, want)
	}
}

func TestMutatingCommandsReportIdentity(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)
	resetKBFlag(t)

	stderr := captureStderr(t, func() {
		if err := reportKBIdentity(writeCmd, nil); err != nil {
			t.Errorf("reportKBIdentity: %v", err)
		}
	})
	if want := "kb: write-test (" + kbRoot + ")\n"; stderr != want {
		t.Errorf("stderr = %q, want %q", stderr, want)
	}
}

func TestCommandsOutsideAKnowledgeBaseReportNothing(t *testing.T) {
	// skill install works without a knowledge base: it must neither resolve
	// one nor report one.
	t.Setenv("HOME", t.TempDir())
	t.Setenv(path.KBEnvVar, "")
	resetKBFlag(t)

	stderr := captureStderr(t, func() {
		if err := reportKBIdentity(skillInstallCmd, nil); err != nil {
			t.Errorf("reportKBIdentity: %v", err)
		}
	})
	if stderr != "" {
		t.Errorf("stderr = %q, want no identity line", stderr)
	}
}

// TestMutatingCommandsMatchRegisteredTree pins the hand-maintained
// mutatingCommands map against the registered command tree: the key set must
// equal the expected list and every key must resolve to a registered command
// with a RunE. Renaming a command's Use string, restructuring the tree, or
// adding a mutating command without updating the map fails here instead of
// silently dropping the stderr identity line.
func TestMutatingCommandsMatchRegisteredTree(t *testing.T) {
	expected := []string{
		"akb write",
		"akb append",
		"akb delete",
		"akb approve",
		"akb index add",
		"akb index remove",
		"akb index rebuild",
		"akb log append",
		"akb raw write",
		"akb raw delete",
		"akb raw sync",
		"akb template write",
		"akb template delete",
	}

	got := make([]string, 0, len(mutatingCommands))
	for commandPath := range mutatingCommands {
		got = append(got, commandPath)
	}
	sort.Strings(got)

	want := append([]string(nil), expected...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mutatingCommands keys = %v, want %v", got, want)
	}

	for _, commandPath := range got {
		if cmd := resolveCommandPath(t, commandPath); cmd.RunE == nil {
			t.Errorf("mutating command %q has no RunE, so the identity hook cannot run", commandPath)
		}
	}
}

// TestMutatingCommandReportsIdentityOnStderr drives the identity line through
// the real binary, covering the root hook wiring and the environment selection.
func TestMutatingCommandReportsIdentityOnStderr(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	content := "---\ntype: note\ntitle: Identity Note\nsummary: An identity test\ntags: test\n---\nBody."
	out, err := writeRun(kbRoot, "notes/identity.md", content)
	if err != nil {
		t.Fatalf("akb write failed: %s: %v", out, err)
	}
	if want := "kb: write-test (" + kbRoot + ")"; !strings.Contains(out, want) {
		t.Errorf("output = %q, want it to contain %q", out, want)
	}
}
