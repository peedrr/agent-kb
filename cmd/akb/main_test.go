package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var akbBinPath, akbTestBinPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "akb-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v", err)
		os.Exit(1)
	}
	akbPath := filepath.Join(tmpDir, "akb")

	buildCmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"go", "build", "-o", akbPath, "./")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build: %s: %v", string(out), err)
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // cleanup on build failure
		os.Exit(1)
	}

	testVersion := "v0.1.0"
	akbTestPath := filepath.Join(tmpDir, "akb-test")
	buildCmdTest := exec.Command( //nolint:gosec // test helper launching akb binary
		"go", "build", "-ldflags", fmt.Sprintf("-X main.version=%s", testVersion), "-o", akbTestPath, "./")
	if out, err := buildCmdTest.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build test binary: %s: %v", string(out), err)
		_ = os.RemoveAll(tmpDir) //nolint:errcheck // cleanup on build failure
		os.Exit(1)
	}

	akbBinPath = akbPath
	akbTestBinPath = akbTestPath

	code := m.Run()

	// Cleanup (must be explicit — defer doesn't run with os.Exit)
	_ = os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup — failure is non-fatal
	os.Exit(code)
}

func TestVersionFlag(t *testing.T) {
	if version == "" {
		t.Fatal("version should not be empty")
	}

	expected := "akb " + version
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("akb --version failed: %v", err)
	}

	actual := strings.TrimSpace(string(out))
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestVersionFlagWithLdflags(t *testing.T) {
	testVersion := "v0.1.0"

	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbTestBinPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("akb --version failed: %v", err)
	}

	expected := "akb " + testVersion
	actual := strings.TrimSpace(string(out))
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestNoArgsShowsHelp(t *testing.T) {
	cmd := exec.Command( //nolint:gosec // test helper launching akb binary
		akbBinPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("akb (no args) failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "Usage:") {
		t.Errorf("expected help to contain Usage, got: %s", output)
	}
}

// classifiesInvocation runs one CLI invocation in process and returns the exit
// code and stderr report main would produce for it.
func classifiesInvocation(t *testing.T, args ...string) (int, string) {
	t.Helper()

	RootCmd.SetArgs(args)
	return classifyExit(Execute())
}

// assertUsageReport fails the test unless report is a usage report that names
// the command's own mistake exactly once.
func assertUsageReport(t *testing.T, report string, want string) {
	t.Helper()

	if report != "usage: "+want {
		t.Errorf("report = %q, want %q", report, "usage: "+want)
	}
	if strings.Contains(report, "execute command:") {
		t.Errorf("usage report %q carries the execute context", report)
	}
}

func TestExecuteClassifiesUnknownFlagAsUsage(t *testing.T) {
	code, report := classifiesInvocation(t, "--unknown-flag")

	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	assertUsageReport(t, report, "unknown flag: --unknown-flag")
}

func TestExecuteClassifiesUnknownCommandAsUsage(t *testing.T) {
	code, report := classifiesInvocation(t, "frobnicate")

	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	assertUsageReport(t, report, `unknown command "frobnicate" for "akb"`)
}

func TestExecuteClassifiesMissingArgumentAsUsage(t *testing.T) {
	code, report := classifiesInvocation(t, "append")

	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	assertUsageReport(t, report, "accepts 1 arg(s), received 0")
}

func TestExecuteSuggestsKnownSubcommand(t *testing.T) {
	code, report := classifiesInvocation(t, "writ")

	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	if !strings.Contains(report, "Did you mean this?") || !strings.Contains(report, "\twrite") {
		t.Errorf("report = %q, want a suggestion for the write command", report)
	}
}

func TestManagedFileRefusalsClassifyAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	deleteErr := runDeleteCmd(&cobra.Command{}, []string{"kb/index.md"})
	if deleteErr == nil {
		t.Fatal("expected the delete of index.md to be refused")
	}
	var usageErr *usageError
	if !errors.As(deleteErr, &usageErr) {
		t.Errorf("delete refusal %v is not a typed usage error", deleteErr)
	}
	code, report := classifyExit(deleteErr)
	if code != exitFault {
		t.Errorf("delete refusal exit code = %d, want %d", code, exitFault)
	}
	assertUsageReport(t, report, "cannot delete index.md; use 'akb index rebuild' to reset")

	removeErr := runIndexRemove(&cobra.Command{}, []string{"kb/./log.md"})
	if removeErr == nil {
		t.Fatal("expected the index removal of log.md to be refused")
	}
	if !errors.As(removeErr, &usageErr) {
		t.Errorf("index removal refusal %v is not a typed usage error", removeErr)
	}
	code, report = classifyExit(removeErr)
	if code != exitFault {
		t.Errorf("index removal refusal exit code = %d, want %d", code, exitFault)
	}
	assertUsageReport(t, report, "cannot remove log.md from index")
}

func TestPathGuardViolationClassifiesAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	err := runDeleteCmd(&cobra.Command{}, []string{"../escape.md"})
	if err == nil {
		t.Fatal("expected the delete of a '..' path to be refused")
	}

	code, report := classifyExit(commandFailure{err: err})
	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	assertUsageReport(t, report, "resolve path: path must not contain '..'")
}
