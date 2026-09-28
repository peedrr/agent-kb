// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
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
	initAuthorEmail string
	initAuthorName  string
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
	initCmd.Flags().StringVar(&initAuthorName, "author-name", "", "commit identity name the init commit records; never written to akb.yaml (default: AKB_AUTHOR_NAME, then git config)")
	initCmd.Flags().StringVar(&initAuthorEmail, "author-email", "", "commit identity email the init commit records; never written to akb.yaml (default: AKB_AUTHOR_EMAIL, then git config)")
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

	// The versioning mode the invocation chose, and the identity the base
	// records for its commits. A base that is not versioned in git never commits,
	// so it resolves and records no identity.
	versioning := config.VersioningGit
	if initNoGit {
		versioning = config.VersioningNone
	}

	var identity *storage.Identity
	var identityNotice string
	if !initNoGit {
		resolved, source := storage.ResolveInitIdentity(nearestExistingDir(absTarget), initAuthorName, initAuthorEmail)
		identity, identityNotice = initIdentity(resolved, source, !noCommit)
		// The commit path resolves the identity from AKB_AUTHOR_NAME and
		// AKB_AUTHOR_EMAIL before anything the machine configures, so a flag
		// attributes the init commit by naming those variables. Only a value a flag
		// named is exported: an empty variable would shadow the identity the
		// environment or git config names, and an invocation that names no flag
		// leaves both sources untouched.
		if name := strings.TrimSpace(initAuthorName); name != "" {
			if err := os.Setenv(storage.AuthorNameEnv, name); err != nil {
				return fmt.Errorf("export %s: %w", storage.AuthorNameEnv, err)
			}
		}
		if email := strings.TrimSpace(initAuthorEmail); email != "" {
			if err := os.Setenv(storage.AuthorEmailEnv, email); err != nil {
				return fmt.Errorf("export %s: %w", storage.AuthorEmailEnv, err)
			}
		}
	}

	if err := createDirectoryStructure(target); err != nil {
		return err
	}

	if err := writeSeedFiles(target); err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	if err := writeAkbConfig(target, name, now, initDescription, versioning, identity); err != nil {
		return err
	}

	if err := initSearchDB(target); err != nil {
		return err
	}

	// The base's own ignore file is written in every layout. The root .gitignore
	// is skipped for an unversioned base at its host repository's root: nothing
	// versions that base, so the rules would have no consumer, and appending them
	// would mutate a file the host repository tracks, leaving the host's status
	// dirty.
	unversionedAtRepoRoot := initNoGit && targetIsRepoRoot
	if err := writeGitignoreFiles(target, !unversionedAtRepoRoot); err != nil {
		return err
	}

	var modeNotice string
	if initNoGit {
		if hostRoot != "" {
			excluded, count, err := excludeFromHostRepo(absTarget, hostRoot)
			if err != nil {
				return err
			}
			// A base at the repository root is excluded by the directories it
			// owns — three entries — so the notice's undo guidance refers to them
			// in the plural there.
			undo := "remove that line to undo"
			if count > 1 {
				undo = "remove those lines to undo"
			}
			modeNotice = fmt.Sprintf("kb: excluded %s from host git tracking via .git/info/exclude (local to this clone; %s)", excluded, undo)
		}
	} else if err := commitBase(target, name, initEmbed); err != nil {
		return err
	}

	fmt.Printf("Initialized KB %q at %s\n", name, absTarget)
	if initEmbed {
		fmt.Printf("kb: %s is embedded inside repository %s; akb commits only kb/, raw/, .agent-kb/\n", name, hostRoot)
	}
	if modeNotice != "" {
		fmt.Println(modeNotice)
	}
	if identityNotice != "" {
		fmt.Println(identityNotice)
	}
	return nil
}

// initIdentity decides what the new base records for its commits and what the
// invocation reports. An identity the invocation's flags or environment named,
// or the machine's git configuration, belongs to that invocation or machine and
// is never written into a file that travels with the base: the file would
// mis-attribute every commit made from another machine after a clone. Only the
// akb default is recorded, so that a base cloned to a machine without a git
// identity still commits. commits is false for an init that leaves its commit to
// its caller (--no-commit), where no commit of the invocation carries the
// identity.
func initIdentity(identity storage.Identity, source storage.IdentitySource, commits bool) (*storage.Identity, string) {
	switch source {
	case storage.IdentityFromDefault:
		return &identity, fmt.Sprintf("commit identity: %s <%s> (default — no git identity found; recorded in akb.yaml, edit git-author/git-email to change)", identity.Name, identity.Email)
	case storage.IdentityFromGitConfig:
		return nil, fmt.Sprintf("commit identity: from git config (%s <%s>) — not recorded; each machine's git identity applies", identity.Name, identity.Email)
	case storage.IdentityFromFlag:
		if !commits {
			return nil, fmt.Sprintf("commit identity: from --author-name/--author-email (%s <%s>) — not recorded; the commit stays with the caller and its own identity; later commits use each machine's identity", identity.Name, identity.Email)
		}
		return nil, fmt.Sprintf("commit identity: from --author-name/--author-email (%s <%s>) — not recorded; this init commit uses it; later commits use each machine's identity", identity.Name, identity.Email)
	default:
		return nil, fmt.Sprintf("commit identity: from AKB_AUTHOR_NAME/AKB_AUTHOR_EMAIL (%s <%s>) — not recorded; the environment of each invocation applies", identity.Name, identity.Email)
	}
}

// commitBase records the new base in the versioning mode it was created in: a
// standalone base gets a repository of its own, while an embedded one is
// committed into the repository that hosts it. The commit records exactly the
// base's own paths, so changes another tool staged in the host worktree stay
// staged and uncommitted. An invocation that leaves the commit to its caller
// stages those paths instead.
func commitBase(target, name string, embed bool) error {
	if !embed {
		// `git init` runs directly: it neither creates nor contends on the
		// repository index lock, which is the lock the staging and commit steps
		// behind it wait out.
		gitInit := exec.Command("git", "init", target) //nolint:gosec // launching trusted git binary with controlled args
		if out, err := gitInit.CombinedOutput(); err != nil {
			return fmt.Errorf("git init: %s: %w", strings.TrimSpace(string(out)), err)
		}
	}

	store, err := storage.OpenStore(target, noCommit)
	if err != nil {
		return fmt.Errorf("open base storage: %w", err)
	}

	basePaths := []string{"kb", "raw", path.StateDirName, ".gitignore"}
	if noCommit {
		if err := store.StageFiles(basePaths...); err != nil {
			return fmt.Errorf("stage the base's files: %w", err)
		}
		return nil
	}
	if err := store.Commit(fmt.Sprintf("akb: init %s", name), basePaths...); err != nil {
		return fmt.Errorf("commit the base: %w", err)
	}
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

func writeAkbConfig(target, name, created, description, versioning string, identity *storage.Identity) error {
	cfg := &config.Config{
		Name:        name,
		Created:     created,
		Description: strings.TrimSpace(description),
		Versioning:  versioning,
	}
	// Only an identity the base itself owns is recorded: one from the machine or
	// the invocation stays out of a file that travels with the base.
	if identity != nil {
		cfg.GitAuthor = identity.Name
		cfg.GitEmail = identity.Email
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

// writeGitignoreFiles writes the ignore rules init owns. Both files merge: a
// target that already carries a .gitignore keeps every line of it and gains the
// missing ones, so writing them twice changes nothing. rootGitignore is false
// for an unversioned base at its host repository's root, where the root file
// belongs to the host repository and carries no rules for the base.
func writeGitignoreFiles(target string, rootGitignore bool) error {
	if err := appendMissingLines(filepath.Join(path.StateDir(target), ".gitignore"), "search.db*"); err != nil {
		return fmt.Errorf("write .agent-kb/.gitignore: %w", err)
	}
	if !rootGitignore {
		return nil
	}

	if err := appendMissingLines(filepath.Join(target, ".gitignore"), "*.akb.bak", ".agent-kb/search.db*"); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
}

// appendMissingLines appends the lines target does not carry yet, creating the
// file when it is not there. Every line that is already in the file stays as it
// is: an ignore file that predates the base keeps its own rules.
func appendMissingLines(target string, lines ...string) error {
	existing, err := os.ReadFile(target) //nolint:gosec // init writes fixed files under the base it creates
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", target, err)
	}

	present := make(map[string]bool)
	for _, line := range strings.Split(string(existing), "\n") {
		present[strings.TrimSpace(line)] = true
	}

	var missing []string
	for _, line := range lines {
		if !present[strings.TrimSpace(line)] {
			missing = append(missing, line)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(target), err)
	}

	// A file that does not end in a newline gets one before the appended lines,
	// so its last line does not run into the first appended one.
	content := strings.Join(missing, "\n") + "\n"
	if len(existing) > 0 && !bytes.HasSuffix(existing, []byte("\n")) {
		content = "\n" + content
	}

	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) //nolint:gosec // init writes fixed files under the base it creates
	if err != nil {
		return fmt.Errorf("open %s: %w", target, err)
	}
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close() //nolint:errcheck // the write error is the one to report
		return fmt.Errorf("append to %s: %w", target, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", target, err)
	}
	return nil
}
