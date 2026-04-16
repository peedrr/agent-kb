package path

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Common error messages
var (
	ErrEmptyPath      = errors.New("path must not be empty")
	ErrAbsolutePath   = errors.New("path must be relative")
	ErrParentDir      = errors.New("path must not contain '..'")
	ErrUseAKBWrite    = errors.New("use `akb write` or `akb read`")
	ErrUseAKBRawWrite = errors.New("use `akb raw write` or `akb raw read`")
)

// ResolveKBPath resolves a KB-relative path with guard rails.
// It strips redundant "kb/" prefix and validates the path.
func ResolveKBPath(kbRoot, inputPath string) (string, error) {
	// Reject empty path
	if strings.TrimSpace(inputPath) == "" {
		return "", ErrEmptyPath
	}

	// Reject absolute paths
	if filepath.IsAbs(inputPath) || (len(inputPath) > 1 && inputPath[1] == ':') {
		return "", ErrAbsolutePath
	}

	// Reject parent directory traversal
	if strings.Contains(inputPath, "..") {
		return "", ErrParentDir
	}

	// Reject raw/ prefix - user should use akb raw commands
	if strings.HasPrefix(inputPath, "raw/") || inputPath == "raw" {
		return "", ErrUseAKBRawWrite
	}

	// Strip redundant kb/ prefix
	cleanPath := strings.TrimPrefix(inputPath, "kb/")
	// If input was exactly "kb", TrimPrefix returns "kb", so handle that
	if cleanPath == inputPath && inputPath != "kb" {
		cleanPath = inputPath
	} else if cleanPath == "kb" && inputPath == "kb" {
		cleanPath = "."
	}

	return filepath.Join(kbRoot, cleanPath), nil
}

// ResolveRawPath resolves a raw path with guard rails.
// It strips redundant "raw/" prefix and validates the path.
func ResolveRawPath(kbRoot, inputPath string) (string, error) {
	// Reject empty path
	if strings.TrimSpace(inputPath) == "" {
		return "", ErrEmptyPath
	}

	// Reject absolute paths
	if filepath.IsAbs(inputPath) || (len(inputPath) > 1 && inputPath[1] == ':') {
		return "", ErrAbsolutePath
	}

	// Reject parent directory traversal
	if strings.Contains(inputPath, "..") {
		return "", ErrParentDir
	}

	// Reject kb/ prefix - user should use akb write/read commands
	if strings.HasPrefix(inputPath, "kb/") || inputPath == "kb" {
		return "", ErrUseAKBWrite
	}

	// Strip redundant raw/ prefix
	cleanPath := strings.TrimPrefix(inputPath, "raw/")
	// If input was exactly "raw", TrimPrefix returns "raw", so handle that
	if cleanPath == inputPath && inputPath != "raw" {
		cleanPath = inputPath
	} else if cleanPath == "raw" && inputPath == "raw" {
		cleanPath = "."
	}

	return filepath.Join(kbRoot, "raw", cleanPath), nil
}

// KBRoot finds the KB root by walking up from cwd looking for .akb directory.
func KBRoot() (string, error) {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
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
