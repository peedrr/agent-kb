// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/peedrr/agent-kb/internal/config"
	"github.com/peedrr/agent-kb/internal/db"
	"github.com/peedrr/agent-kb/internal/path"
	"github.com/peedrr/agent-kb/internal/storage"
)

var (
	initDescription string
	initEmbed       bool
	initNoGit       bool
	initForce       bool
)

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize a new Agent Knowledge Base",
	Long:  `Create a new KB directory with the standard structure, its versioning mode, and configuration.`,
	Example: `  # Initialize a standalone KB: a git repository of its own
  akb init my-kb

  # Initialize a KB inside the current repository, recorded in that repository's history
  akb init docs --embed

  # Initialize a KB that is not versioned at all
  akb init scratch --no-git

  # Initialize a KB with a description (shown by akb discover)
  akb init my-kb --description "Project notes and decisions"`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initDescription, "description", "", "short description of the knowledge base's contents (optional; shown by akb discover)")
	initCmd.Flags().BoolVar(&initEmbed, "embed", false, "version the KB in the enclosing repository's history; akb commits only the KB's own paths")
	initCmd.Flags().BoolVar(&initNoGit, "no-git", false, "do not version the KB; exclude it from the enclosing repository's git status")
	initCmd.Flags().BoolVar(&initForce, "force", false, "write .agent-kb/, kb/ and raw/ directly to a repository root")
}

// The flag explanations of the versioning refusal: what each mode does with the
// repository the KB either joins or stays out of.
const (
	embedFlagHelp = "  --embed    the KB joins this repository's history; akb commits only its own paths (kb/, raw/, .agent-kb/), never other worktree changes"
	noGitFlagHelp = "  --no-git   the KB is not versioned; akb excludes it from the host's git status via .git/info/exclude (local to this clone)"
)

// repoRootLayoutRefusal reports the layout that writes the base's directories
// into the root of a repository. --force is the only way past it, and it chooses
// no versioning mode: the modes stay orthogonal to the layout.
const repoRootLayoutRefusal = "This will write .agent-kb/, kb/ and raw/ directly to the repo root. --force if desired, otherwise prefer a subdirectory: akb init <name> --embed."

func runInit(_ *cobra.Command, args []string) error {
	if err := validateName(args[0]); err != nil {
		return err
	}
	if initEmbed && initNoGit {
		return &usageError{msg: "--embed and --no-git cannot be used together"}
	}

	// `akb init .` initializes the directory it runs in and adopts that
	// directory's name: the name reaches akb.yaml, the init commit, and the
	// success line. Every other argument is the name and the directory at once.
	target := filepath.Clean(args[0])
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve absolute path: %w", err)
	}
	name := args[0]
	if target == "." {
		name = filepath.Base(absTarget)
		if err := validateName(name); err != nil {
			return err
		}
	}

	if _, err := os.Stat(path.StateDir(target)); err == nil {
		return fmt.Errorf("kb already initialized at %s. Use a different name or remove existing kb", name)
	}

	if err := checkGitAvailable(); err != nil {
		return err
	}

	// The repository that would host the new base, and whether the target
	// directory is that repository's root: together they decide which refusal an
	// invocation without a versioning mode gets, and whether the repo-root
	// layout warning applies. A target that is not there yet is created inside
	// the repository of its parent, so detection starts from the nearest
	// directory that exists.
	hostRoot, err := detectHostRepo(nearestExistingDir(absTarget))
	if err != nil {
		return err
	}
	targetIsRepoRoot := sameDirectory(hostRoot, absTarget)

	if hostRoot != "" && !initEmbed && !initNoGit {
		return &usageError{msg: versioningRefusal(name, hostRoot, targetIsRepoRoot)}
	}
	if targetIsRepoRoot && !initForce {
		return &usageError{msg: repoRootLayoutRefusal}
	}
	if initEmbed && hostRoot == "" {
		return &usageError{msg: fmt.Sprintf("--embed: %s is not inside a git repository — omit --embed to create a standalone repository", name)}
	}

	if err := createDirectoryStructure(target); err != nil {
		return err
	}

	if err := writeSeedFiles(target); err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	if err := writeAkbConfig(target, name, now, initDescription); err != nil {
		return err
	}

	if err := initSearchDB(target); err != nil {
		return err
	}

	if err := writeGitignoreFiles(target); err != nil {
		return err
	}

	if err := gitInitAndCommit(target, name); err != nil {
		return err
	}

	fmt.Printf("Initialized KB %q at %s\n", name, absTarget)
	return nil
}

// versioningRefusal reports that the target directory is inside a git repository
// and no versioning mode was chosen, and explains what each mode does. The
// wording names the membership of the target: a subdirectory is inside the
// repository, while a target that is the repository root is a repository of its
// own. A target at the repository root carries the layout warning as well.
func versioningRefusal(name, hostRoot string, targetIsRepoRoot bool) string {
	membership := fmt.Sprintf("%s is inside git repository %s", name, hostRoot)
	if targetIsRepoRoot {
		membership = fmt.Sprintf("%s is a git repository", name)
	}

	refusal := fmt.Sprintf("%s — choose how the KB is versioned:\n\n%s\n%s", membership, embedFlagHelp, noGitFlagHelp)
	if targetIsRepoRoot {
		refusal += "\n\n" + repoRootLayoutRefusal
	}
	return refusal
}

func validateName(name string) error {
	if name == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("name must not be empty")
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("name must not contain '/'")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("name must not contain '..'")
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("name must not start with '-'")
	}
	return nil
}

func checkGitAvailable() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is required but not found. Please install git")
	}
	return nil
}

// createDirectoryStructure creates the templates directory empty: page types
// are authored per KB, and `akb template get --examples` is the copy source
// for a KB that has none of its own.
func createDirectoryStructure(target string) error {
	dirs := []string{
		filepath.Join(target, "kb"),
		filepath.Join(target, "raw"),
		path.TemplatesDir(target),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	return nil
}

func writeSeedFiles(target string) error {
	seeds := map[string]string{
		filepath.Join(target, "kb", "index.md"):   "# Index\n\n",
		filepath.Join(target, "kb", "log.md"):     "# Log\n\n",
		filepath.Join(target, "raw", "files.log"): "# Auto-generated by agent-kb. Do not edit manually.\n# filename | sha256 | last_updated\n",
	}
	for path, content := range seeds {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

func writeAkbConfig(target, name, created, description string) error {
	cfg := &config.Config{
		Name:        name,
		Created:     created,
		Description: strings.TrimSpace(description),
	}
	if err := config.Save(path.ConfigPath(target), cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	return nil
}

func initSearchDB(target string) error {
	dbPath := path.SearchDBPath(target)
	conn, err := db.InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	defer conn.Close() //nolint:errcheck // DB close error non-critical on command exit

	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("enable WAL mode: %w", err)
	}
	conn.SetMaxOpenConns(1)

	if err := db.CreateSchema(conn); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return nil
}

func writeGitignoreFiles(target string) error {
	akbGitignore := filepath.Join(path.StateDir(target), ".gitignore")
	if err := os.WriteFile(akbGitignore, []byte("search.db*\n"), 0600); err != nil {
		return fmt.Errorf("write .agent-kb/.gitignore: %w", err)
	}

	rootGitignore := filepath.Join(target, ".gitignore")
	if err := os.WriteFile(rootGitignore, []byte("*.akb.bak\n.agent-kb/search.db*\n"), 0600); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
}

func gitInitAndCommit(target, name string) error {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	gitInit := exec.Command("git", "init", target) //nolint:gosec // launching trusted git binary with controlled args
	if out, err := gitInit.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %s: %w", strings.TrimSpace(string(out)), err)
	}

	if err := ensureGitConfig(absTarget); err != nil {
		return err
	}

	// The staging and commit steps touch the repository index, so they run
	// through the retry runner every other staging and commit step uses:
	// another process holding the index lock is waited out instead of failing
	// the fresh base. The `git init` above and the `git config` invocations in
	// ensureGitConfig run directly via exec.Command on purpose: neither creates
	// nor contends on the repository index lock.
	if _, err := storage.RunGit(absTarget, "git add", "add", "-A"); err != nil {
		//nolint:wrapcheck // RunGit's error already names the git step and its output
		return err
	}

	commitMsg := fmt.Sprintf("akb: init %s", name)
	if _, err := storage.RunGit(absTarget, "git commit", "commit", "-m", commitMsg); err != nil {
		//nolint:wrapcheck // RunGit's error already names the git step and its output
		return err
	}

	return nil
}

// ensureGitConfig gives a repository without a git identity the akb identity to
// commit under. Only `akb init` runs it: every other command commits under the
// repository's configured identity or the per-invocation akb fallback, and none
// of them write repository config.
func ensureGitConfig(repoPath string) error {
	gitConfig := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...) //nolint:gosec // launching trusted git binary with controlled args
		cmd.Dir = repoPath
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}

	if _, err := gitConfig("config", "user.name"); err != nil {
		if _, err := gitConfig("config", "user.name", "akb"); err != nil {
			return fmt.Errorf("set git user.name: %w", err)
		}
	}

	if _, err := gitConfig("config", "user.email"); err != nil {
		if _, err := gitConfig("config", "user.email", "akb@local"); err != nil {
			return fmt.Errorf("set git user.email: %w", err)
		}
	}

	return nil
}
