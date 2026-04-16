package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var akbBinPath, akbTestBinPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "akb-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v", err)
		os.Exit(1)
	}
	akbPath := filepath.Join(tmpDir, "akb")

	buildCmd := exec.Command("go", "build", "-o", akbPath, "./")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build: %s: %v", string(out), err)
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	testVersion := "v0.1.0"
	akbTestPath := filepath.Join(tmpDir, "akb-test")
	buildCmdTest := exec.Command("go", "build", "-ldflags", fmt.Sprintf("-X main.version=%s", testVersion), "-o", akbTestPath, "./")
	if out, err := buildCmdTest.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build test binary: %s: %v", string(out), err)
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	akbBinPath = akbPath
	akbTestBinPath = akbTestPath

	code := m.Run()

	// Cleanup (must be explicit — defer doesn't run with os.Exit)
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

func TestVersionFlag(t *testing.T) {
	if version == "" {
		t.Fatal("version should not be empty")
	}

	expected := "akb " + version
	cmd := exec.Command(akbBinPath, "--version")
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

	cmd := exec.Command(akbTestBinPath, "--version")
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
	cmd := exec.Command(akbBinPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("akb (no args) failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "Usage:") {
		t.Errorf("expected help to contain Usage, got: %s", output)
	}
}
