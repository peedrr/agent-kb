package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	if version == "" {
		t.Fatal("version should not be empty")
	}

	dir, _ := os.Getwd()
	akbPath := filepath.Join(dir, "akb")

	expected := "akb " + version
	cmd := exec.Command(akbPath, "--version")
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
	dir, _ := os.Getwd()
	akbPath := filepath.Join(dir, "..", "..", "akb-test")

	if _, err := os.Stat(akbPath); os.IsNotExist(err) {
		t.Skip("akb-test binary not built - run: CGO_ENABLED=0 go build -ldflags \"-X main.version=v0.1.0\" -o akb-test ./cmd/akb/")
	}

	cmd := exec.Command(akbPath, "--version")
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
	dir, _ := os.Getwd()
	akbPath := filepath.Join(dir, "akb")

	cmd := exec.Command(akbPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("akb (no args) failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "Usage:") {
		t.Errorf("expected help to contain Usage, got: %s", output)
	}
}
