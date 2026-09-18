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

// assertUsageFailure fails the test unless err is a typed usage error that the
// CLI exits on with the fault exit code and a report carrying want verbatim.
func assertUsageFailure(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected the invocation to fail with %q, got nil error", want)
	}
	var usageErr *usageError
	if !errors.As(err, &usageErr) {
		t.Errorf("failure %v is not a typed usage error", err)
	}
	code, report := classifyExit(err)
	if code != exitFault {
		t.Errorf("exit code = %d, want %d", code, exitFault)
	}
	assertUsageReport(t, report, want)
}

// charDeviceStdin points os.Stdin at the null device for the duration of the
// test, so a command that guards against terminal input sees one.
func charDeviceStdin(t *testing.T) {
	t.Helper()

	devNull, err := os.OpenFile(os.DevNull, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}

	old := os.Stdin
	os.Stdin = devNull
	t.Cleanup(func() {
		os.Stdin = old
		_ = devNull.Close() //nolint:errcheck // test cleanup — failure is non-fatal
	})
}

// withWriteFlags sets the write command's flag state for one invocation and
// restores it afterwards.
func withWriteFlags(t *testing.T, frontmatter []string, appendMode bool) {
	t.Helper()

	origFrontmatter, origAppend := writeFrontmatter, writeAppend
	writeFrontmatter, writeAppend = frontmatter, appendMode
	t.Cleanup(func() { writeFrontmatter, writeAppend = origFrontmatter, origAppend })
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

// TestWriteInvocationMistakesClassifyAsUsage drives each input branch of
// `akb write` into its invocation mistakes: every one of them reports the same
// message as before and now exits with the usage fault code.
func TestWriteInvocationMistakesClassifyAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	const noteContent = "---\ntype: note\ntitle: A note\n---\nBody."

	cases := []struct {
		name        string
		args        []string
		frontmatter []string
		appendMode  bool
		stdin       string
		want        string
	}{
		{
			name:        "frontmatter with append",
			args:        []string{"notes/any.md"},
			frontmatter: []string{"title=Other"},
			appendMode:  true,
			want:        "--frontmatter and --append cannot be used together",
		},
		{
			name:        "frontmatter without .md extension",
			args:        []string{"notes/any"},
			frontmatter: []string{"title=Other"},
			want:        "page filename must end with .md",
		},
		{
			name:        "frontmatter on a raw path",
			args:        []string{"raw/any.md"},
			frontmatter: []string{"title=Other"},
			want:        "use `akb raw write`",
		},
		{
			name:        "frontmatter on index.md",
			args:        []string{"kb/index.md"},
			frontmatter: []string{"title=Other"},
			want:        "cannot write index.md; use 'akb index add' to update",
		},
		{
			name:        "frontmatter on log.md",
			args:        []string{"kb/log.md"},
			frontmatter: []string{"title=Other"},
			want:        "cannot write log.md; it is a managed file",
		},
		{
			name:       "append without .md extension",
			args:       []string{"notes/any"},
			appendMode: true,
			stdin:      "Appended.",
			want:       "page filename must end with .md",
		},
		{
			name:       "append on a raw path",
			args:       []string{"raw/any.md"},
			appendMode: true,
			stdin:      "Appended.",
			want:       "use `akb raw write`",
		},
		{
			name:       "append on index.md",
			args:       []string{"kb/index.md"},
			appendMode: true,
			stdin:      "Appended.",
			want:       "cannot write index.md; use 'akb index add' to update",
		},
		{
			name:       "append on log.md",
			args:       []string{"kb/log.md"},
			appendMode: true,
			stdin:      "Appended.",
			want:       "cannot write log.md; it is a managed file",
		},
		{
			name:  "stdin write without .md extension",
			args:  []string{"notes/any"},
			stdin: noteContent,
			want:  "page filename must end with .md",
		},
		{
			name:  "stdin write on a raw path",
			args:  []string{"raw/any.md"},
			stdin: noteContent,
			want:  "use `akb raw write`",
		},
		{
			name:  "stdin write on index.md",
			args:  []string{"kb/index.md"},
			stdin: noteContent,
			want:  "cannot write index.md; use 'akb index add' to update",
		},
		{
			name:  "stdin write on log.md",
			args:  []string{"kb/log.md"},
			stdin: noteContent,
			want:  "cannot write log.md; it is a managed file",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withWriteFlags(t, tc.frontmatter, tc.appendMode)
			pointStdinAtTempFile(t, tc.stdin)

			assertUsageFailure(t, runWrite(nil, tc.args), tc.want)
		})
	}

	t.Run("write without stdin", func(t *testing.T) {
		withWriteFlags(t, nil, false)
		charDeviceStdin(t)

		assertUsageFailure(t, runWrite(nil, []string{"notes/any.md"}), "input required: pipe content to stdin")
	})
}

// TestAppendInvocationMistakesClassifyAsUsage drives the invocation mistakes of
// `akb append`.
func TestAppendInvocationMistakesClassifyAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	cases := []struct {
		name string
		path string
		want string
	}{
		{"raw path", "raw/any.md", "use `akb raw write`"},
		{"index.md", "kb/index.md", "cannot append to index.md; use 'akb index add' to update"},
		{"log.md", "kb/log.md", "cannot append to log.md; it is a managed file"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pointStdinAtTempFile(t, "Appended.")

			assertUsageFailure(t, runAppend(nil, []string{tc.path}), tc.want)
		})
	}

	t.Run("append without stdin", func(t *testing.T) {
		charDeviceStdin(t)

		assertUsageFailure(t, runAppend(nil, []string{"notes/any.md"}), "input required: pipe content to stdin")
	})
}

// TestDeleteRawPrefixClassifiesAsUsage accesses a raw file through the page
// command, which akb answers with the raw command to use instead.
func TestDeleteRawPrefixClassifiesAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	assertUsageFailure(t, runDeleteCmd(nil, []string{"raw/any.md"}), "use `akb raw delete`")
}

// TestRawWriteInvocationMistakesClassifyAsUsage drives the invocation mistakes
// of `akb raw write`.
func TestRawWriteInvocationMistakesClassifyAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	t.Run("files.log", func(t *testing.T) {
		pointStdinAtTempFile(t, "content")

		assertUsageFailure(t, runRawWrite(nil, []string{"files.log"}), "cannot write files.log directly; it is a managed file")
	})

	t.Run("raw write without stdin", func(t *testing.T) {
		charDeviceStdin(t)

		assertUsageFailure(t, runRawWrite(nil, []string{"data.txt"}), "input required: pipe content to stdin")
	})
}

// TestReadFamilyAdviceClassifiesAsUsage covers the commands of each family
// answering a path of the other one: the advice is unchanged and now reports as
// a usage mistake.
func TestReadFamilyAdviceClassifiesAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	t.Run("read on a raw path", func(t *testing.T) {
		assertUsageFailure(t, runRead(nil, []string{"raw/data.csv"}), "use `akb raw read`")
	})

	t.Run("raw read on a kb path", func(t *testing.T) {
		assertUsageFailure(t, runRawRead(nil, []string{"kb/notes/any.md"}), "use `akb read`")
	})
}

// TestIndexAddManagedFileRefusalClassifiesAsUsage covers `akb index add`
// refusing to index a managed file.
func TestIndexAddManagedFileRefusalClassifiesAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	assertUsageFailure(t, runIndexAdd(nil, []string{"kb/index.md", "Bad entry"}), "cannot add index.md to index")
	assertUsageFailure(t, runIndexAdd(nil, []string{"kb/log.md", "Bad entry"}), "cannot add log.md to index")
}

// TestTemplateDeleteForceRefusalClassifiesAsUsage covers `akb template delete`
// refusing to delete without --force.
func TestTemplateDeleteForceRefusalClassifiesAsUsage(t *testing.T) {
	kbRoot := writeSetupTestKB(t)
	defer writeCleanup(kbRoot)

	origForce := tdForce
	tdForce = false
	t.Cleanup(func() { tdForce = origForce })

	_, err := captureOutput(func() error { return runTemplateDelete(nil, []string{"note"}) })
	assertUsageFailure(t, err, "deletion refused without --force")
}
