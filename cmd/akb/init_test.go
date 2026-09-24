package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestInitDoesNotRegisterOrSetDefault covers `akb init` creating only the
// knowledge base: no registry is written under ~/.config/agent-kb and the
// success output claims no default.
func TestInitDoesNotRegisterOrSetDefault(t *testing.T) {
	tmpDir := t.TempDir()
	home := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(home, 0750); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(akbBinPath, "init", "init-kb") //nolint:gosec // test helper launching akb binary
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "HOME="+home)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("akb init failed: %s: %v", out, err)
	}

	if !strings.Contains(string(out), `Initialized KB "init-kb"`) {
		t.Errorf("output = %q, want the initialization line", out)
	}
	if strings.Contains(string(out), "set as default") {
		t.Errorf("output = %q, want no default claim", out)
	}

	registryDir := filepath.Join(home, ".config", "agent-kb")
	if _, err := os.Stat(registryDir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("registry directory %s exists after init (stat error = %v)", registryDir, err)
	}
}

// TestInitWaitsOutAnIndexLock pins that `akb init` stages and commits through
// the index-lock retry runner: a bootstrap whose `git add` meets a lock another
// process holds and releases a moment later completes and records its init commit
// instead of failing on the first attempt.
func TestInitWaitsOutAnIndexLock(t *testing.T) {
	tmpDir := t.TempDir()
	name := "locked-kb"

	// A repository skeleton whose index lock is held while init stages the new
	// base; `git init` reinitializes the directory in place.
	if err := os.MkdirAll(filepath.Join(tmpDir, name, ".git"), 0750); err != nil {
		t.Fatal(err)
	}
	indexLock := filepath.Join(tmpDir, name, ".git", "index.lock")
	if err := os.WriteFile(indexLock, nil, 0600); err != nil { //nolint:gosec // test helper creating a fixed file in its temp repository
		t.Fatal(err)
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		_ = os.Remove(indexLock) //nolint:errcheck // test releasing the index lock
	}()

	cmd := exec.Command(akbBinPath, "init", name) //nolint:gosec // test helper launching akb binary
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("akb init failed while the index lock was held: %s: %v", out, err)
	}

	logCmd := exec.Command("git", "log", "-1", "--format=%s") //nolint:gosec // test helper launching trusted git binary
	logCmd.Dir = filepath.Join(tmpDir, name)
	logOut, err := logCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("read the init commit: %s: %v", logOut, err)
	}
	if subject := strings.TrimSpace(string(logOut)); subject != "akb: init "+name {
		t.Errorf("init commit subject = %q, want %q", subject, "akb: init "+name)
	}
}
