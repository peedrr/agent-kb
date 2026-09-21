package path

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		{
			name:        "filename with spaces",
			inputPath:   "notes/my note.md",
			expected:    filepath.Join(kbRoot, "notes/my note.md"),
			expectError: false,
		},
		{
			name:        "unicode filename",
			inputPath:   "notes/日本語.md",
			expected:    filepath.Join(kbRoot, "notes/日本語.md"),
			expectError: false,
		},
		{
			name:        "emoji filename",
			inputPath:   "notes/🚀-rocket.md",
			expected:    filepath.Join(kbRoot, "notes/🚀-rocket.md"),
			expectError: false,
		},
		{
			name:        "very long filename",
			inputPath:   "notes/" + strings.Repeat("a", 255) + ".md",
			expected:    filepath.Join(kbRoot, "notes/"+strings.Repeat("a", 255)+".md"),
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
				if tt.errType != nil && !errors.Is(err, tt.errType) {
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
		{
			name:        "filename with spaces",
			inputPath:   "data/my file.txt",
			expected:    filepath.Join(kbRoot, "raw", "data/my file.txt"),
			expectError: false,
		},
		{
			name:        "unicode filename",
			inputPath:   "data/日本語.txt",
			expected:    filepath.Join(kbRoot, "raw", "data/日本語.txt"),
			expectError: false,
		},
		{
			name:        "emoji filename",
			inputPath:   "data/🚀-rocket.txt",
			expected:    filepath.Join(kbRoot, "raw", "data/🚀-rocket.txt"),
			expectError: false,
		},
		{
			name:        "very long filename",
			inputPath:   "data/" + strings.Repeat("a", 255) + ".txt",
			expected:    filepath.Join(kbRoot, "raw", "data/"+strings.Repeat("a", 255)+".txt"),
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
				if tt.errType != nil && !errors.Is(err, tt.errType) {
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
	defer os.RemoveAll(tmpDir) //nolint:errcheck,gosec // test cleanup — failure is non-fatal

	// Test case: KB root found with .akb directory
	t.Run("finds .akb directory", func(t *testing.T) {
		workDir, err := os.MkdirTemp(tmpDir, "work")
		if err != nil {
			t.Fatalf("failed to create work dir: %v", err)
		}

		akbDir := filepath.Join(workDir, ".akb")
		if err := os.MkdirAll(akbDir, 0750); err != nil {
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
		if !strings.Contains(err.Error(), "not in a knowledge base") {
			t.Errorf("error should contain 'not in a knowledge base', got: %v", err)
		}
	})

	// Test case: finds root in parent directory
	t.Run("finds root in parent directory", func(t *testing.T) {
		parentDir, err := os.MkdirTemp(tmpDir, "parent")
		if err != nil {
			t.Fatalf("failed to create parent dir: %v", err)
		}

		subDir := filepath.Join(parentDir, "sub")
		if err := os.MkdirAll(subDir, 0750); err != nil {
			t.Fatalf("failed to create sub dir: %v", err)
		}

		akbDir := filepath.Join(parentDir, ".akb")
		if err := os.MkdirAll(akbDir, 0750); err != nil {
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

func TestResolveKB(t *testing.T) {
	t.Run("flag wins over environment", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv(KBEnvVar, filepath.Join(home, "from-env"))
		flagPath := filepath.Join(home, "from-flag")

		got, err := ResolveKB(flagPath)
		if err != nil {
			t.Fatalf("ResolveKB failed: %v", err)
		}
		if got != flagPath {
			t.Fatalf("expected %q, got %q", flagPath, got)
		}
	})

	t.Run("environment used when flag is empty", func(t *testing.T) {
		home := t.TempDir()
		envPath := filepath.Join(home, "from-env")
		t.Setenv(KBEnvVar, envPath)

		got, err := ResolveKB("")
		if err != nil {
			t.Fatalf("ResolveKB failed: %v", err)
		}
		if got != envPath {
			t.Fatalf("expected %q, got %q", envPath, got)
		}
	})

	t.Run("empty environment counts as unset", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv(KBEnvVar, "")

		_, err := ResolveKB("")
		if !errors.Is(err, ErrNoKB) {
			t.Fatalf("expected ErrNoKB, got %v", err)
		}
	})

	t.Run("relative path resolves against the working directory", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("get working directory: %v", err)
		}

		got, err := ResolveKB("kb")
		if err != nil {
			t.Fatalf("ResolveKB failed: %v", err)
		}
		if want := filepath.Join(cwd, "kb"); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("tilde expands to the home directory", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		got, err := ResolveKB("~/kb")
		if err != nil {
			t.Fatalf("ResolveKB failed: %v", err)
		}
		if want := filepath.Join(home, "kb"); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}

		got, err = ResolveKB("~")
		if err != nil {
			t.Fatalf("ResolveKB failed: %v", err)
		}
		if got != home {
			t.Fatalf("expected %q, got %q", home, got)
		}
	})

	t.Run("absolute path is normalized", func(t *testing.T) {
		home := t.TempDir()

		got, err := ResolveKB(filepath.Join(home, "kb", "..", "other"))
		if err != nil {
			t.Fatalf("ResolveKB failed: %v", err)
		}
		if want := filepath.Join(home, "other"); got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("missing selection reports the usage rule", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv(KBEnvVar, "")

		_, err := ResolveKB("")
		if err == nil {
			t.Fatal("expected error when no knowledge base is selected")
		}
		var guardErr *GuardError
		if !errors.As(err, &guardErr) {
			t.Fatalf("expected a GuardError, got %T", err)
		}
		if !errors.Is(err, ErrNoKB) {
			t.Fatalf("expected ErrNoKB, got %v", err)
		}
	})

	t.Run("removed registry file notes deprecation on the usage-error path", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv(KBEnvVar, "")

		registryPath := filepath.Join(home, ".config", "agent-kb", "registry.yaml")
		if err := os.MkdirAll(filepath.Dir(registryPath), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(registryPath, []byte("default: elsewhere\n"), 0600); err != nil {
			t.Fatal(err)
		}

		stderr := captureStderr(t, func() {
			_, err := ResolveKB("")
			if !errors.Is(err, ErrNoKB) {
				t.Errorf("expected ErrNoKB, got %v", err)
			}
		})
		if !strings.Contains(stderr, registryPath) {
			t.Errorf("note %q does not name %q", stderr, registryPath)
		}
		if !strings.Contains(stderr, "no longer used") {
			t.Errorf("note %q does not say the registry is unused", stderr)
		}
	})

	t.Run("no note when the registry file is absent", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv(KBEnvVar, "")

		stderr := captureStderr(t, func() {
			_, _ = ResolveKB("") //nolint:errcheck // only the stderr note is under test
		})
		if stderr != "" {
			t.Errorf("expected no note, got %q", stderr)
		}
	})
}

// captureStderr runs fn with os.Stderr redirected to a pipe and returns the
// text fn wrote.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stderr
	os.Stderr = w
	fn()
	os.Stderr = old

	if err := w.Close(); err != nil {
		t.Fatalf("close stderr pipe: %v", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stderr pipe: %v", err)
	}
	return string(data)
}
