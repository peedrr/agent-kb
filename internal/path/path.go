// Package path resolves and validates knowledge base paths.
package path

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

// KBEnvVar is the environment variable that selects the knowledge base when
// the --kb flag is not given.
const KBEnvVar = "AKB_KB"

// Common error messages
var (
	ErrEmptyPath      = errors.New("path must not be empty")
	ErrAbsolutePath   = errors.New("path must be relative")
	ErrParentDir      = errors.New("path must not contain '..'")
	ErrUseAKBWrite    = errors.New("use `akb write` or `akb read`")
	ErrUseAKBRawWrite = errors.New("use `akb raw write` or `akb raw read`")
	ErrNoKB           = errors.New("no knowledge base selected: pass --kb <path> or set the AKB_KB environment variable")
)

// GuardError reports a rejected invocation: an input path that violates the
// path rules (an absolute path, a '..' escape, or a path that belongs to the
// other command family) or a command that ran without a knowledge base
// selected. The resolvers return it so callers can tell a rejected invocation
// apart from a fault, and the CLI maps it to the usage exit code. It wraps the
// rule's sentinel error, so errors.Is keeps matching the specific rule.
type GuardError struct {
	rule error
}

func (e *GuardError) Error() string { return e.rule.Error() }

func (e *GuardError) Unwrap() error { return e.rule }

// ResolveKBPath resolves a KB-relative path with guard rails.
// It strips redundant "kb/" prefix and validates the path.
func ResolveKBPath(kbRoot, inputPath string) (string, error) {
	// Reject empty path
	if strings.TrimSpace(inputPath) == "" {
		return "", &GuardError{rule: ErrEmptyPath}
	}

	// Reject absolute paths
	if filepath.IsAbs(inputPath) || (len(inputPath) > 1 && inputPath[1] == ':') {
		return "", &GuardError{rule: ErrAbsolutePath}
	}

	// Reject parent directory traversal
	if strings.Contains(inputPath, "..") {
		return "", &GuardError{rule: ErrParentDir}
	}

	// Reject raw/ prefix - user should use akb raw commands
	if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
		return "", &GuardError{rule: ErrUseAKBRawWrite}
	}

	// Strip redundant kb/ prefix
	cleanPath := strings.TrimPrefix(inputPath, "kb/")
	// If input was exactly "kb", TrimPrefix returns "kb", so handle that
	if cleanPath == inputPath && inputPath != "kb" {
		cleanPath = inputPath
	} else if cleanPath == "kb" && inputPath == "kb" {
		cleanPath = "."
	}

	// Normalize the resolved input path so "./" prefixes and redundant
	// separators cannot mask the target from callers.
	cleanPath = filepath.Clean(cleanPath)

	return filepath.Join(kbRoot, cleanPath), nil
}

// ResolveRawPath resolves a raw path with guard rails.
// It strips redundant "raw/" prefix and validates the path.
func ResolveRawPath(kbRoot, inputPath string) (string, error) {
	// Reject empty path
	if strings.TrimSpace(inputPath) == "" {
		return "", &GuardError{rule: ErrEmptyPath}
	}

	// Reject absolute paths
	if filepath.IsAbs(inputPath) || (len(inputPath) > 1 && inputPath[1] == ':') {
		return "", &GuardError{rule: ErrAbsolutePath}
	}

	// Reject parent directory traversal
	if strings.Contains(inputPath, "..") {
		return "", &GuardError{rule: ErrParentDir}
	}

	// Reject kb/ prefix - user should use akb write/read commands
	if strings.HasPrefix(inputPath, "kb/") || inputPath == "kb" {
		return "", &GuardError{rule: ErrUseAKBWrite}
	}

	// Strip redundant raw/ prefix
	cleanPath := strings.TrimPrefix(inputPath, "raw/")
	// If input was exactly "raw", TrimPrefix returns "raw", so handle that
	if cleanPath == inputPath && inputPath != "raw" {
		cleanPath = inputPath
	} else if cleanPath == "raw" && inputPath == "raw" {
		cleanPath = "."
	}

	// Normalize the resolved input path so "./" prefixes and redundant
	// separators cannot mask the target from callers.
	cleanPath = filepath.Clean(cleanPath)

	return filepath.Join(kbRoot, "raw", cleanPath), nil
}

// ResolveKB returns the absolute path of the knowledge base selected for an
// invocation: the --kb flag when given, otherwise the AKB_KB environment
// variable. A leading ~ is expanded to the home directory and a relative path
// is resolved against the working directory. With neither set, the command has
// no base to act on and ResolveKB reports a usage error, noting the removed
// registry file when it is still on disk.
func ResolveKB(flagKB string) (string, error) {
	selected := strings.TrimSpace(flagKB)
	if selected == "" {
		selected = strings.TrimSpace(os.Getenv(KBEnvVar))
	}
	if selected == "" {
		if note := DeprecatedRegistryNote(); note != "" {
			fmt.Fprintln(os.Stderr, note)
		}
		return "", &GuardError{rule: noKBError()}
	}

	expanded, err := expandHome(selected)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("resolve knowledge base path %q: %w", selected, err)
	}
	return absPath, nil
}

// expandHome expands a leading ~ in p to the user's home directory.
func expandHome(p string) (string, error) {
	if p != "~" && !strings.HasPrefix(p, "~"+string(filepath.Separator)) {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand ~: %w", err)
	}
	if p == "~" {
		return home, nil
	}
	return filepath.Join(home, p[2:]), nil
}

// DeprecatedRegistryNote returns the one-line notice printed when a command
// ran without a knowledge base selected while the removed registry file is
// still on disk, and when `akb discover` runs. The file is never read for
// resolution; its presence only triggers the notice. It returns "" when there
// is nothing to note.
func DeprecatedRegistryNote() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	registryPath := filepath.Join(home, ".config", "agent-kb", "registry.yaml")
	if _, err := os.Stat(registryPath); err != nil {
		return ""
	}
	return fmt.Sprintf("note: %s is no longer used; select a knowledge base with --kb or the %s environment variable", registryPath, KBEnvVar)
}

// KBRoot reports the root of the knowledge base that dir belongs to: dir
// itself when it holds a .akb directory, otherwise the nearest ancestor that
// does. It reports an error when no directory from dir up to the filesystem
// root holds one. A directory is a knowledge base root exactly when KBRoot
// reports the directory itself, which is how a discovery scan detects the
// .akb/ marker.
func KBRoot(dir string) (string, error) {
	dir = filepath.Clean(dir)

	// Walk up the directory tree
	for {
		if isKBRoot(dir) {
			return dir, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding KB
			break
		}
		dir = parent
	}

	return "", errors.New("not in a knowledge base directory (no .akb/ found)")
}

// isKBRoot reports whether dir itself holds the .akb directory that marks a
// knowledge base root.
func isKBRoot(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".akb"))
	return err == nil && info.IsDir()
}

// Discovery bounds: the scan checks every directory it visits for the .akb/
// marker, descends to discoverDepth levels below each visited directory, and
// stops after discoverBudget filesystem entries, so a deep or crowded tree
// cannot make an invocation crawl.
const (
	discoverDepth  = 2
	discoverBudget = 1000
)

// discoverSkipDirs names the directories the scan never descends into or
// records: repository and dependency internals rather than knowledge bases
// users address. Hidden directories are skipped for the same reason, next to
// these names.
var discoverSkipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
}

// DiscoveredKB reports one knowledge base found by a discovery scan.
type DiscoveredKB struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// Discover reports the knowledge bases in the neighborhood of root: root
// itself, the directories below it to two levels, and the same for every
// ancestor up to the home directory — or, when root does not live below it, up
// to the filesystem root. Hidden directories and the .git, node_modules, and
// vendor directories are skipped, and the walk stops after a fixed entry
// budget.
//
// Entries come nearest-first, by their distance in path components from root
// and then by path, so the base a caller stands in is reported before the ones
// further out. Discovery is read-only and never selects a base: callers
// address what it finds with --kb or AKB_KB.
func Discover(root string) []DiscoveredKB {
	// Distances and the home boundary compare between absolute paths, so a
	// relative root is resolved against the working directory first.
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	} else {
		root = filepath.Clean(root)
	}

	candidates := map[string]int{}
	remaining := discoverBudget
	for _, anchor := range discoverAnchors(root) {
		walkDiscoveryAnchor(anchor, root, candidates, &remaining)
		if remaining <= 0 {
			break
		}
	}

	type ranked struct {
		kb       DiscoveredKB
		distance int
	}
	found := make([]ranked, 0, len(candidates))
	for dir, distance := range candidates {
		// A directory is a knowledge base root when it is its own KBRoot: a
		// directory that merely lives inside a base is not reported.
		kbRoot, err := KBRoot(dir)
		if err != nil || kbRoot != dir {
			continue
		}

		name, description := readKBMetadata(dir)
		found = append(found, ranked{
			kb:       DiscoveredKB{Name: name, Path: dir, Description: description},
			distance: distance,
		})
	}

	sort.Slice(found, func(i, j int) bool {
		if found[i].distance != found[j].distance {
			return found[i].distance < found[j].distance
		}
		return found[i].kb.Path < found[j].kb.Path
	})

	discovered := make([]DiscoveredKB, 0, len(found))
	for _, entry := range found {
		discovered = append(discovered, entry.kb)
	}
	return discovered
}

// discoverAnchors lists the directories whose neighborhoods a scan covers:
// root and every ancestor up to the home directory, or up to the filesystem
// root when the home directory does not contain root.
func discoverAnchors(root string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	anchors := make([]string, 0, 8)
	for dir := root; ; {
		anchors = append(anchors, dir)

		if dir == home {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return anchors
}

// walkDiscoveryAnchor visits the directories in anchor's neighborhood: anchor
// itself and its descendants to discoverDepth levels. Every visited directory
// becomes a candidate at its distance from root, skipped names are neither
// recorded nor descended into, and the shared budget stops the walk once the
// scan has seen enough entries.
func walkDiscoveryAnchor(anchor, root string, candidates map[string]int, remaining *int) {
	// WalkDir reports per-entry problems to the callback and returns only the
	// callback's own result, which is always nil here.
	_ = filepath.WalkDir(anchor, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			// An unreadable branch is skipped, not reported: the scan informs,
			// so it must not fail an invocation.
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if *remaining <= 0 {
			return fs.SkipAll
		}
		*remaining--

		if !entry.IsDir() {
			return nil
		}
		if path != anchor && isDiscoverSkip(entry.Name()) {
			return fs.SkipDir
		}

		distance := pathDistance(root, path)
		if current, ok := candidates[path]; !ok || distance < current {
			candidates[path] = distance
		}

		if pathDistance(anchor, path) >= discoverDepth {
			return fs.SkipDir
		}
		return nil
	})
}

// isDiscoverSkip reports whether a discovery scan avoids the directory with
// this name: repository and dependency internals, plus hidden directories.
func isDiscoverSkip(name string) bool {
	return discoverSkipDirs[name] || strings.HasPrefix(name, ".")
}

// pathDistance counts the path components between two directories: 0 when they
// name the same directory, 1 for a directory and its direct child, and so on.
// Paths the platform cannot relate sort last.
func pathDistance(from, to string) int {
	rel, err := filepath.Rel(from, to)
	if err != nil {
		return math.MaxInt
	}
	if rel == "." {
		return 0
	}
	return len(strings.Split(rel, string(filepath.Separator)))
}

// kbMetadata is the part of a knowledge base's .akb.yaml that discovery
// reports.
type kbMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// readKBMetadata reads the name and optional description a knowledge base
// reports about itself. A base whose config is missing, unreadable, or without
// a name is still reported, under its directory name: discovery informs, so a
// torn config must not hide the base.
func readKBMetadata(root string) (name, description string) {
	data, err := os.ReadFile(filepath.Join(root, ".akb", ".akb.yaml")) //nolint:gosec // KB root supplied by the scan
	if err == nil {
		var metadata kbMetadata
		if err := yaml.Unmarshal(data, &metadata); err == nil {
			name = strings.TrimSpace(metadata.Name)
			description = strings.TrimSpace(metadata.Description)
		}
	}
	if name == "" {
		name = filepath.Base(root)
	}
	return name, description
}

// FormatDiscovered renders discovered knowledge bases, nearest first, as the
// listing a human reads. It is the single spelling of that listing, shared by
// the no-selection usage error and `akb discover`.
func FormatDiscovered(discovered []DiscoveredKB) string {
	var b strings.Builder
	b.WriteString("discovered knowledge bases (nearest first):")
	for _, kb := range discovered {
		b.WriteString("\n  ")
		b.WriteString(kb.Name)
		b.WriteString("  ")
		b.WriteString(kb.Path)
		if kb.Description != "" {
			b.WriteString("  ")
			b.WriteString(kb.Description)
		}
	}
	b.WriteString("\n\nselect one with --kb <path> or AKB_KB=<path>")
	return b.String()
}

// noKBError reports the missing knowledge base selection, listing the bases
// the neighborhood scan found so the next invocation can select one of them.
// The error still wraps ErrNoKB.
func noKBError() error {
	cwd, err := os.Getwd()
	if err != nil {
		return ErrNoKB
	}

	discovered := Discover(cwd)
	if len(discovered) == 0 {
		return ErrNoKB
	}
	return fmt.Errorf("%w\n\n%s", ErrNoKB, FormatDiscovered(discovered))
}
