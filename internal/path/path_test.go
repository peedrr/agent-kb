package path

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveKBPath(t *testing.T) {
	kbRoot := "/tmp/kb"

	tests := []struct {
		name        string
		inputPath   string
		expected    string
		expectError bool
		errType     error
	}{
		// Basic resolution
		{
			name:        "basic path resolution",
			inputPath:   "docs/readme.md",
			expected:    filepath.Join(kbRoot, "docs/readme.md"),
			expectError: false,
		},

		// Empty path - MUST reject
		{
			name:        "empty path",
			inputPath:   "",
			expectError: true,
			errType:     ErrEmptyPath,
		},
		{
			name:        "whitespace only path",
			inputPath:   "   ",
			expectError: true,
			errType:     ErrEmptyPath,
		},

		// Absolute paths - MUST reject
		{
			name:        "absolute path",
			inputPath:   "/etc/passwd",
			expectError: true,
			errType:     ErrAbsolutePath,
		},
		{
			name:        "absolute path with drive",
			inputPath:   "C:\\Windows",
			expectError: true,
			errType:     ErrAbsolutePath,
		},

		// Parent directory traversal - MUST reject
		{
			name:        "parent directory with ..",
			inputPath:   "../secret.txt",
			expectError: true,
			errType:     ErrParentDir,
		},
		{
			name:        "parent in middle of path",
			inputPath:   "docs/../secret.txt",
			expectError: true,
			errType:     ErrParentDir,
		},
		{
			name:        "double parent",
			inputPath:   "a/b/../../secret.txt",
			expectError: true,
			errType:     ErrParentDir,
		},

		// raw/ prefix - MUST reject with special error
		{
			name:        "raw prefix",
			inputPath:   "raw/file.txt",
			expectError: true,
			errType:     ErrUseAKBRawWrite,
		},
		{
			name:        "raw prefix only",
			inputPath:   "raw",
			expectError: true,
			errType:     ErrUseAKBRawWrite,
		},
		{
			name:        "raw prefix with path",
			inputPath:   "raw/secrets/api.key",
			expectError: true,
			errType:     ErrUseAKBRawWrite,
		},

		// kb/ prefix - SHOULD strip
		{
			name:        "kb prefix stripped",
			inputPath:   "kb/docs/readme.md",
			expected:    filepath.Join(kbRoot, "docs/readme.md"),
			expectError: false,
		},
		{
			name:        "kb prefix only",
			inputPath:   "kb",
			expected:    filepath.Join(kbRoot, "."),
			expectError: false,
		},

		// Normal paths
		{
			name:        "simple filename",
			inputPath:   "notes.txt",
			expected:    filepath.Join(kbRoot, "notes.txt"),
			expectError: false,
		},
		{
			name:        "nested path",
			inputPath:   "a/b/c/d.txt",
			expected:    filepath.Join(kbRoot, "a/b/c/d.txt"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveKBPath(kbRoot, tt.inputPath)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Fatalf("expected error type %v but got %v", tt.errType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestResolveRawPath(t *testing.T) {
	kbRoot := "/tmp/kb"

	tests := []struct {
		name        string
		inputPath   string
		expected    string
		expectError bool
		errType     error
	}{
		// Basic resolution
		{
			name:        "basic path resolution",
			inputPath:   "logs/app.log",
			expected:    filepath.Join(kbRoot, "raw", "logs/app.log"),
			expectError: false,
		},

		// Empty path - MUST reject
		{
			name:        "empty path",
			inputPath:   "",
			expectError: true,
			errType:     ErrEmptyPath,
		},
		{
			name:        "whitespace only path",
			inputPath:   "   ",
			expectError: true,
			errType:     ErrEmptyPath,
		},

		// Absolute paths - MUST reject
		{
			name:        "absolute path",
			inputPath:   "/etc/passwd",
			expectError: true,
			errType:     ErrAbsolutePath,
		},

		// Parent directory traversal - MUST reject
		{
			name:        "parent directory with ..",
			inputPath:   "../secret.txt",
			expectError: true,
			errType:     ErrParentDir,
		},
		{
			name:        "parent in middle",
			inputPath:   "logs/../secret.txt",
			expectError: true,
			errType:     ErrParentDir,
		},

		// kb/ prefix - MUST reject with special error
		{
			name:        "kb prefix",
			inputPath:   "kb/file.txt",
			expectError: true,
			errType:     ErrUseAKBWrite,
		},
		{
			name:        "kb prefix only",
			inputPath:   "kb",
			expectError: true,
			errType:     ErrUseAKBWrite,
		},

		// raw/ prefix - SHOULD strip
		{
			name:        "raw prefix stripped",
			inputPath:   "raw/logs/app.log",
			expected:    filepath.Join(kbRoot, "raw", "logs/app.log"),
			expectError: false,
		},
		{
			name:        "raw prefix only",
			inputPath:   "raw",
			expected:    filepath.Join(kbRoot, "raw", "."),
			expectError: false,
		},

		// Normal paths
		{
			name:        "simple filename",
			inputPath:   "config.json",
			expected:    filepath.Join(kbRoot, "raw", "config.json"),
			expectError: false,
		},
		{
			name:        "nested path",
			inputPath:   "a/b/c.txt",
			expected:    filepath.Join(kbRoot, "raw", "a/b/c.txt"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveRawPath(kbRoot, tt.inputPath)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Fatalf("expected error type %v but got %v", tt.errType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestKBRoot(t *testing.T) {
	// Create temp directory structure for testing
	tmpDir, err := os.MkdirTemp("", "kb-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case: KB root found with .akb directory
	t.Run("finds .akb directory", func(t *testing.T) {
		workDir, err := os.MkdirTemp(tmpDir, "work")
		if err != nil {
			t.Fatalf("failed to create work dir: %v", err)
		}

		akbDir := filepath.Join(workDir, ".akb")
		if err := os.MkdirAll(akbDir, 0755); err != nil {
			t.Fatalf("failed to create .akb dir: %v", err)
		}

		if err := os.Chdir(workDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		root, err := KBRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != workDir {
			t.Fatalf("expected %q but got %q", workDir, root)
		}
	})

	// Test case: KB root found with .akb.yaml file
	t.Run("finds .akb.yaml file", func(t *testing.T) {
		workDir, err := os.MkdirTemp(tmpDir, "work2")
		if err != nil {
			t.Fatalf("failed to create work dir: %v", err)
		}

		akbFile := filepath.Join(workDir, ".akb.yaml")
		if err := os.WriteFile(akbFile, []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create .akb.yaml: %v", err)
		}

		if err := os.Chdir(workDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		root, err := KBRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != workDir {
			t.Fatalf("expected %q but got %q", workDir, root)
		}
	})

	// Test case: KB root NOT found
	t.Run("not found returns error", func(t *testing.T) {
		workDir, err := os.MkdirTemp(tmpDir, "work3")
		if err != nil {
			t.Fatalf("failed to create work dir: %v", err)
		}

		if err := os.Chdir(workDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		_, err = KBRoot()
		if err == nil {
			t.Fatalf("expected error but got nil")
		}
	})

	// Test case: finds root in parent directory
	t.Run("finds root in parent directory", func(t *testing.T) {
		parentDir, err := os.MkdirTemp(tmpDir, "parent")
		if err != nil {
			t.Fatalf("failed to create parent dir: %v", err)
		}

		subDir := filepath.Join(parentDir, "sub")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("failed to create sub dir: %v", err)
		}

		akbDir := filepath.Join(parentDir, ".akb")
		if err := os.MkdirAll(akbDir, 0755); err != nil {
			t.Fatalf("failed to create .akb dir: %v", err)
		}

		if err := os.Chdir(subDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		root, err := KBRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != parentDir {
			t.Fatalf("expected %q but got %q", parentDir, root)
		}
	})
}
