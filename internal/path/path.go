// Package path resolves and validates knowledge base paths.
package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		if note := deprecatedRegistryNote(); note != "" {
			fmt.Fprintln(os.Stderr, note)
		}
		return "", &GuardError{rule: ErrNoKB}
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

// deprecatedRegistryNote returns the one-line notice printed when a command ran
// without a knowledge base selected while the removed registry file is still
// on disk. The file is never read for resolution; its presence only triggers
// the notice. It returns "" when there is nothing to note.
func deprecatedRegistryNote() string {
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

// KBRoot finds the KB root by walking up from cwd looking for .akb directory.
func KBRoot() (string, error) {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// Walk up the directory tree
	for {
		// Check for .akb directory
		akbDir := filepath.Join(dir, ".akb")
		if info, err := os.Stat(akbDir); err == nil && info.IsDir() {
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
