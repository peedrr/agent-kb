package path

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
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

func TestResolveKBPathRejectsSymlinkEscapes(t *testing.T) {
	symlinkSupport(t)

	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.md"), []byte("secret\n"), 0600); err != nil {
		t.Fatalf("create outside file: %v", err)
	}

	tests := []struct {
		name  string
		input string
		setup func(t *testing.T, kbRoot string) string
	}{
		{
			name:  "file symlink to an existing file outside",
			input: "evil.md",
			setup: func(t *testing.T, kbRoot string) string {
				link := filepath.Join(kbRoot, "evil.md")
				mustSymlink(t, filepath.Join(outside, "secret.md"), link)
				return link
			},
		},
		{
			name:  "file symlink to a directory outside",
			input: "evil.md",
			setup: func(t *testing.T, kbRoot string) string {
				link := filepath.Join(kbRoot, "evil.md")
				mustSymlink(t, outside, link)
				return link
			},
		},
		{
			name:  "directory symlink used as an intermediate component",
			input: filepath.Join("evil", "secret.md"),
			setup: func(t *testing.T, kbRoot string) string {
				link := filepath.Join(kbRoot, "evil")
				mustSymlink(t, outside, link)
				return link
			},
		},
		{
			name:  "nested intermediate directory symlink",
			input: filepath.Join("docs", "evil", "nested", "secret.md"),
			setup: func(t *testing.T, kbRoot string) string {
				if err := os.MkdirAll(filepath.Join(kbRoot, "docs"), 0750); err != nil {
					t.Fatalf("create docs directory: %v", err)
				}
				link := filepath.Join(kbRoot, "docs", "evil")
				mustSymlink(t, outside, link)
				return link
			},
		},
		{
			name:  "relative symlink pointing outside",
			input: "evil.md",
			setup: func(t *testing.T, kbRoot string) string {
				rel, err := filepath.Rel(kbRoot, filepath.Join(outside, "secret.md"))
				if err != nil {
					t.Fatalf("relative target: %v", err)
				}
				link := filepath.Join(kbRoot, "evil.md")
				mustSymlink(t, rel, link)
				return link
			},
		},
		{
			name:  "dangling symlink whose target lies outside",
			input: "evil.md",
			setup: func(t *testing.T, kbRoot string) string {
				link := filepath.Join(kbRoot, "evil.md")
				mustSymlink(t, filepath.Join(outside, "not-yet.md"), link)
				return link
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kbRoot := t.TempDir()
			link := tt.setup(t, kbRoot)

			got, err := ResolveKBPath(kbRoot, tt.input)
			assertSymlinkEscape(t, err, got, link)
		})
	}
}

func TestResolveRawPathRejectsSymlinkEscapes(t *testing.T) {
	symlinkSupport(t)

	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret\n"), 0600); err != nil {
		t.Fatalf("create outside file: %v", err)
	}

	tests := []struct {
		name  string
		input string
		setup func(t *testing.T, kbRoot string) string
	}{
		{
			name:  "file symlink to an existing file outside",
			input: "evil.txt",
			setup: func(t *testing.T, kbRoot string) string {
				if err := os.MkdirAll(filepath.Join(kbRoot, "raw"), 0750); err != nil {
					t.Fatalf("create raw directory: %v", err)
				}
				link := filepath.Join(kbRoot, "raw", "evil.txt")
				mustSymlink(t, filepath.Join(outside, "secret.txt"), link)
				return link
			},
		},
		{
			name:  "directory symlink used as an intermediate component",
			input: filepath.Join("raw", "evil", "secret.txt"),
			setup: func(t *testing.T, kbRoot string) string {
				if err := os.MkdirAll(filepath.Join(kbRoot, "raw"), 0750); err != nil {
					t.Fatalf("create raw directory: %v", err)
				}
				link := filepath.Join(kbRoot, "raw", "evil")
				mustSymlink(t, outside, link)
				return link
			},
		},
		{
			name:  "the raw directory itself is a symlink outside",
			input: "secret.txt",
			setup: func(t *testing.T, kbRoot string) string {
				link := filepath.Join(kbRoot, "raw")
				mustSymlink(t, outside, link)
				return link
			},
		},
		{
			name:  "dangling symlink whose target lies outside",
			input: "evil.txt",
			setup: func(t *testing.T, kbRoot string) string {
				if err := os.MkdirAll(filepath.Join(kbRoot, "raw"), 0750); err != nil {
					t.Fatalf("create raw directory: %v", err)
				}
				link := filepath.Join(kbRoot, "raw", "evil.txt")
				mustSymlink(t, filepath.Join(outside, "not-yet.txt"), link)
				return link
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kbRoot := t.TempDir()
			link := tt.setup(t, kbRoot)

			got, err := ResolveRawPath(kbRoot, tt.input)
			assertSymlinkEscape(t, err, got, link)
		})
	}
}

func TestResolveKBPathKeepsSymlinksInsideTheBase(t *testing.T) {
	symlinkSupport(t)

	kbRoot := t.TempDir()
	docs := filepath.Join(kbRoot, "docs")
	if err := os.MkdirAll(docs, 0750); err != nil {
		t.Fatalf("create docs directory: %v", err)
	}
	target := filepath.Join(docs, "target.md")
	if err := os.WriteFile(target, []byte("body\n"), 0600); err != nil {
		t.Fatalf("create target page: %v", err)
	}

	mustSymlink(t, target, filepath.Join(kbRoot, "absolute-link.md"))
	mustSymlink(t, filepath.Join("docs", "target.md"), filepath.Join(kbRoot, "relative-link.md"))
	mustSymlink(t, docs, filepath.Join(kbRoot, "alias"))
	mustSymlink(t, filepath.Join(docs, "not-yet.md"), filepath.Join(kbRoot, "dangling-link.md"))

	tests := []struct {
		name  string
		input string
	}{
		{name: "absolute file symlink", input: "absolute-link.md"},
		{name: "relative file symlink", input: "relative-link.md"},
		{name: "directory symlink", input: filepath.Join("alias", "target.md")},
		{name: "dangling symlink inside the base", input: "dangling-link.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveKBPath(kbRoot, tt.input)
			if err != nil {
				t.Fatalf("ResolveKBPath(%q) rejected a symlink inside the base: %v", tt.input, err)
			}
			if want := filepath.Join(kbRoot, tt.input); got != want {
				t.Fatalf("ResolveKBPath(%q) = %q, want %q", tt.input, got, want)
			}
		})
	}
}

func TestResolveRawPathKeepsSymlinksInsideTheBase(t *testing.T) {
	symlinkSupport(t)

	kbRoot := t.TempDir()
	raw := filepath.Join(kbRoot, "raw")
	if err := os.MkdirAll(raw, 0750); err != nil {
		t.Fatalf("create raw directory: %v", err)
	}
	target := filepath.Join(raw, "target.txt")
	if err := os.WriteFile(target, []byte("body\n"), 0600); err != nil {
		t.Fatalf("create target file: %v", err)
	}

	mustSymlink(t, target, filepath.Join(raw, "link.txt"))
	mustSymlink(t, raw, filepath.Join(raw, "alias"))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "absolute file symlink", input: "link.txt", want: filepath.Join(raw, "link.txt")},
		{name: "directory symlink", input: filepath.Join("raw", "alias", "target.txt"), want: filepath.Join(raw, "alias", "target.txt")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveRawPath(kbRoot, tt.input)
			if err != nil {
				t.Fatalf("ResolveRawPath(%q) rejected a symlink inside the base: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ResolveRawPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveKBPathRejectsEscapesDeeperUnderAContainedSymlink(t *testing.T) {
	symlinkSupport(t)

	kbRoot := t.TempDir()
	docs := filepath.Join(kbRoot, "docs")
	if err := os.MkdirAll(docs, 0750); err != nil {
		t.Fatalf("create docs directory: %v", err)
	}
	outside := t.TempDir()

	mustSymlink(t, docs, filepath.Join(kbRoot, "alias"))
	mustSymlink(t, outside, filepath.Join(docs, "evil"))

	// The link below the contained alias is named by its resolved path: the
	// inspection continues from where the alias landed.
	realRoot, err := filepath.EvalSymlinks(kbRoot)
	if err != nil {
		t.Fatalf("resolve base root: %v", err)
	}

	got, err := ResolveKBPath(kbRoot, filepath.Join("alias", "evil", "secret.md"))
	assertSymlinkEscape(t, err, got, filepath.Join(realRoot, "docs", "evil"))
}

func TestResolveKBPathRejectsSiblingNameSharingTheRootPrefix(t *testing.T) {
	symlinkSupport(t)

	parent := t.TempDir()
	kbRoot := filepath.Join(parent, "kb")
	sibling := filepath.Join(parent, "kb2")
	for _, dir := range []string{kbRoot, sibling} {
		if err := os.MkdirAll(dir, 0750); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(sibling, "secret.md"), []byte("secret\n"), 0600); err != nil {
		t.Fatalf("create sibling file: %v", err)
	}

	link := filepath.Join(kbRoot, "evil.md")
	mustSymlink(t, filepath.Join(sibling, "secret.md"), link)

	got, err := ResolveKBPath(kbRoot, "evil.md")
	assertSymlinkEscape(t, err, got, link)
}

func TestResolversAllowNonexistentWriteTargets(t *testing.T) {
	kbRoot := t.TempDir()
	for _, dir := range []string{"notes", filepath.Join("raw", "logs")} {
		if err := os.MkdirAll(filepath.Join(kbRoot, dir), 0750); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}

	tests := []struct {
		name    string
		resolve func(string, string) (string, error)
		input   string
		want    string
	}{
		{
			name:    "new page in an existing directory",
			resolve: ResolveKBPath,
			input:   "notes/new-page.md",
			want:    filepath.Join(kbRoot, "notes", "new-page.md"),
		},
		{
			name:    "new nested tree",
			resolve: ResolveKBPath,
			input:   "brand/new/tree/page.md",
			want:    filepath.Join(kbRoot, "brand", "new", "tree", "page.md"),
		},
		{
			name:    "new raw file",
			resolve: ResolveRawPath,
			input:   "logs/app.log",
			want:    filepath.Join(kbRoot, "raw", "logs", "app.log"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.resolve(kbRoot, tt.input)
			if err != nil {
				t.Fatalf("resolver rejected the write target %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("resolver returned %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveKBPathAcceptsBaseReachedThroughASymlink(t *testing.T) {
	symlinkSupport(t)

	realRoot := filepath.Join(t.TempDir(), "kb")
	if err := os.MkdirAll(filepath.Join(realRoot, "docs"), 0750); err != nil {
		t.Fatalf("create docs directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realRoot, "docs", "readme.md"), []byte("body\n"), 0600); err != nil {
		t.Fatalf("create readme: %v", err)
	}

	linkParent := filepath.Join(t.TempDir(), "parent")
	mustSymlink(t, filepath.Dir(realRoot), linkParent)
	linkedRoot := filepath.Join(t.TempDir(), "kblink")
	mustSymlink(t, realRoot, linkedRoot)

	tests := []struct {
		name   string
		kbRoot string
	}{
		{name: "symlinked parent", kbRoot: filepath.Join(linkParent, filepath.Base(realRoot))},
		{name: "symlinked base root", kbRoot: linkedRoot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveKBPath(tt.kbRoot, filepath.Join("docs", "readme.md"))
			if err != nil {
				t.Fatalf("ResolveKBPath rejected %q reached through a symlink: %v", tt.kbRoot, err)
			}
			if want := filepath.Join(tt.kbRoot, "docs", "readme.md"); got != want {
				t.Fatalf("ResolveKBPath = %q, want %q", got, want)
			}
		})
	}
}

// symlinkSupport skips a test on a platform where symlinks cannot be created,
// which is the one capability the containment tests need.
func symlinkSupport(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	if err := os.Symlink(dir, filepath.Join(dir, "probe")); err != nil {
		t.Skipf("symlinks are not supported here: %v", err)
	}
}

// mustSymlink creates a symlink or fails the test.
func mustSymlink(t *testing.T, target, link string) {
	t.Helper()

	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink %s -> %s: %v", link, target, err)
	}
}

// assertSymlinkEscape pins the rejection of a path that leaves the base
// through the symlink at link: a GuardError matching ErrSymlinkEscape whose
// message names that link.
func assertSymlinkEscape(t *testing.T, err error, got, link string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected the symlink at %s to be rejected, got %q", link, got)
	}
	var guardErr *GuardError
	if !errors.As(err, &guardErr) {
		t.Fatalf("expected a GuardError, got %T: %v", err, err)
	}
	if !errors.Is(err, ErrSymlinkEscape) {
		t.Fatalf("expected ErrSymlinkEscape, got %v", err)
	}
	if !strings.Contains(err.Error(), link) {
		t.Errorf("error %q does not name the offending symlink %q", err.Error(), link)
	}
}

func TestKBRoot(t *testing.T) {
	t.Run("reports a directory that holds .akb", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".akb"), 0750); err != nil {
			t.Fatalf("failed to create .akb dir: %v", err)
		}

		got, err := KBRoot(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != root {
			t.Fatalf("expected %q but got %q", root, got)
		}
	})

	t.Run("reports the nearest ancestor that holds .akb", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".akb"), 0750); err != nil {
			t.Fatalf("failed to create .akb dir: %v", err)
		}
		subDir := filepath.Join(root, "docs", "notes")
		if err := os.MkdirAll(subDir, 0750); err != nil {
			t.Fatalf("failed to create sub dir: %v", err)
		}

		got, err := KBRoot(subDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != root {
			t.Fatalf("expected %q but got %q", root, got)
		}
	})

	t.Run("not found returns error", func(t *testing.T) {
		subDir := filepath.Join(t.TempDir(), "plain")
		if err := os.MkdirAll(subDir, 0750); err != nil {
			t.Fatalf("failed to create sub dir: %v", err)
		}

		_, err := KBRoot(subDir)
		if err == nil {
			t.Fatalf("expected error but got nil")
		}
		if !strings.Contains(err.Error(), "not in a knowledge base") {
			t.Errorf("error should contain 'not in a knowledge base', got: %v", err)
		}
	})
}

func TestResolveKB(t *testing.T) {
	t.Run("flag wins over environment", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv(KBEnvVar, filepath.Join(home, "from-env"))
		flagPath := filepath.Join(home, "from-flag")
		makeKBFixture(t, flagPath, "name: flag\n")

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
		makeKBFixture(t, envPath, "name: env\n")
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
		makeKBFixture(t, filepath.Join(dir, "kb"), "name: kb\n")

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
		makeKBFixture(t, filepath.Join(home, "kb"), "name: kb\n")
		makeKBFixture(t, home, "name: home\n")

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
		makeKBFixture(t, filepath.Join(home, "other"), "name: other\n")

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

	t.Run("nonexistent directory is not a knowledge base", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing")

		_, err := ResolveKB(missing)
		assertNotAKB(t, err, missing)
	})

	t.Run("plain directory is not a knowledge base", func(t *testing.T) {
		dir := t.TempDir()

		_, err := ResolveKB(dir)
		assertNotAKB(t, err, dir)
	})

	t.Run("directory with .akb but no config is not a knowledge base", func(t *testing.T) {
		dir := t.TempDir()
		makeKBFixture(t, dir, "")

		_, err := ResolveKB(dir)
		assertNotAKB(t, err, dir)
	})

	t.Run("config that is not a regular file is not a knowledge base", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".akb", ".akb.yaml"), 0750); err != nil {
			t.Fatalf("create config directory: %v", err)
		}

		_, err := ResolveKB(dir)
		assertNotAKB(t, err, dir)
	})

	t.Run("initialized knowledge base resolves", func(t *testing.T) {
		dir := t.TempDir()
		makeKBFixture(t, dir, "name: fixture\n")

		got, err := ResolveKB(dir)
		if err != nil {
			t.Fatalf("ResolveKB failed: %v", err)
		}
		if got != dir {
			t.Fatalf("expected %q, got %q", dir, got)
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

// assertNotAKB pins the usage error for a selected path that is not a
// knowledge base: a GuardError matching ErrNotAKB that names the path, the
// missing .akb/.akb.yaml marker, and the `akb discover` pointer.
func assertNotAKB(t *testing.T, err error, path string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected a not-a-knowledge-base error for %q", path)
	}
	var guardErr *GuardError
	if !errors.As(err, &guardErr) {
		t.Fatalf("expected a GuardError, got %T", err)
	}
	if !errors.Is(err, ErrNotAKB) {
		t.Fatalf("expected ErrNotAKB, got %v", err)
	}
	for _, want := range []string{path, filepath.Join(".akb", ".akb.yaml"), "akb discover"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}
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

// scanEnvironment isolates a discovery scan in a temporary tree: home is the
// boundary the scan stops at and work is the directory it runs from, so the
// tree below home is the whole scan region.
func scanEnvironment(t *testing.T) (home, work string) {
	t.Helper()

	home = t.TempDir()
	work = filepath.Join(home, "work")
	if err := os.MkdirAll(work, 0750); err != nil {
		t.Fatalf("create work directory: %v", err)
	}
	t.Setenv("HOME", home)
	t.Chdir(work)
	return home, work
}

// makeKBFixture creates a minimal knowledge base in dir: the .akb directory
// that marks a root, plus a config when one is given.
func makeKBFixture(t *testing.T, dir, config string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(dir, ".akb"), 0750); err != nil {
		t.Fatalf("create .akb directory: %v", err)
	}
	if config == "" {
		return
	}
	if err := os.WriteFile(filepath.Join(dir, ".akb", ".akb.yaml"), []byte(config), 0600); err != nil {
		t.Fatalf("write .akb.yaml: %v", err)
	}
}

// treeSnapshot records every path below root with its size and modification
// time, so a test can pin that a read-only scan changed nothing.
func treeSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()

	snapshot := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		snapshot[path] = fmt.Sprintf("%d %s", info.Size(), info.ModTime())
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return snapshot
}

func TestDiscoverReportsNearestFirst(t *testing.T) {
	_, work := scanEnvironment(t)

	// A base nested two levels below the scan root, one one level below, and
	// one beside the parent of the scan root.
	makeKBFixture(t, filepath.Join(work, "docs", "kb"), "name: nested\ndescription: nested base\n")
	makeKBFixture(t, filepath.Join(work, "project"), "name: project\n")
	makeKBFixture(t, filepath.Join(filepath.Dir(work), "project-kb"), "name: project-kb\n")

	got := Discover(work)
	want := []DiscoveredKB{
		{Name: "project", Path: filepath.Join(work, "project")},
		{Name: "project-kb", Path: filepath.Join(filepath.Dir(work), "project-kb")},
		{Name: "nested", Path: filepath.Join(work, "docs", "kb"), Description: "nested base"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Discover = %#v, want %#v", got, want)
	}
}

func TestDiscoverResolvesRelativeRoots(t *testing.T) {
	_, work := scanEnvironment(t)

	makeKBFixture(t, filepath.Join(work, "kb"), "name: kb\n")

	got := Discover(".")
	want := []DiscoveredKB{{Name: "kb", Path: filepath.Join(work, "kb")}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Discover = %#v, want %#v", got, want)
	}
}

func TestDiscoverSkipsInternalAndHiddenDirs(t *testing.T) {
	_, work := scanEnvironment(t)

	makeKBFixture(t, filepath.Join(work, "visible"), "name: visible\n")
	for _, name := range []string{".git", "node_modules", "vendor", ".hidden"} {
		makeKBFixture(t, filepath.Join(work, name, "hidden-kb"), "name: hidden-kb\n")
	}

	got := Discover(work)
	want := []DiscoveredKB{{Name: "visible", Path: filepath.Join(work, "visible")}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Discover = %#v, want %#v", got, want)
	}
}

func TestDiscoverSkipsBasesWithoutConfig(t *testing.T) {
	_, work := scanEnvironment(t)

	makeKBFixture(t, filepath.Join(work, "initialized"), "name: initialized\n")
	makeKBFixture(t, filepath.Join(work, "torn"), "")

	got := Discover(work)
	want := []DiscoveredKB{{Name: "initialized", Path: filepath.Join(work, "initialized")}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Discover = %#v, want %#v", got, want)
	}
}

func TestDiscoverRespectsEntryBudget(t *testing.T) {
	_, work := scanEnvironment(t)

	makeKBFixture(t, filepath.Join(work, "0-first"), "name: first\n")

	crowded := filepath.Join(work, "a-crowded")
	if err := os.MkdirAll(crowded, 0750); err != nil {
		t.Fatalf("create crowded directory: %v", err)
	}
	for i := 0; i <= discoverBudget; i++ {
		name := filepath.Join(crowded, fmt.Sprintf("entry-%04d", i))
		if err := os.WriteFile(name, nil, 0600); err != nil {
			t.Fatalf("create budget entry: %v", err)
		}
	}

	makeKBFixture(t, filepath.Join(work, "z-last"), "name: last\n")

	got := Discover(work)
	foundFirst := false
	for _, kb := range got {
		if kb.Path == filepath.Join(work, "z-last") {
			t.Errorf("Discover reported %q past its %d-entry budget", kb.Path, discoverBudget)
		}
		if kb.Path == filepath.Join(work, "0-first") {
			foundFirst = true
		}
	}
	if !foundFirst {
		t.Errorf("Discover = %#v, want the base visited before the crowded directory", got)
	}
}

func TestDiscoverReadsNameAndDescription(t *testing.T) {
	_, work := scanEnvironment(t)

	makeKBFixture(t, filepath.Join(work, "named"), "name: named-kb\ndescription: a named base\n")
	makeKBFixture(t, filepath.Join(work, "no-description"), "name: bare\n")
	makeKBFixture(t, filepath.Join(work, "no-config"), "other: value\n")

	got := Discover(work)
	want := []DiscoveredKB{
		{Name: "named-kb", Path: filepath.Join(work, "named"), Description: "a named base"},
		{Name: "no-config", Path: filepath.Join(work, "no-config")},
		{Name: "bare", Path: filepath.Join(work, "no-description")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Discover = %#v, want %#v", got, want)
	}
}

func TestDiscoverLeavesTheTreeUnchanged(t *testing.T) {
	_, work := scanEnvironment(t)

	makeKBFixture(t, filepath.Join(work, "kb"), "name: kb\n")
	if err := os.WriteFile(filepath.Join(work, "kb", "note.md"), []byte("body\n"), 0600); err != nil {
		t.Fatalf("create page: %v", err)
	}

	before := treeSnapshot(t, work)
	Discover(work)
	after := treeSnapshot(t, work)

	if !reflect.DeepEqual(before, after) {
		t.Errorf("Discover changed the tree:\nbefore: %v\nafter:  %v", before, after)
	}
}

func TestResolveKBNoSelectionListsNearbyBases(t *testing.T) {
	_, work := scanEnvironment(t)
	t.Setenv(KBEnvVar, "")

	makeKBFixture(t, filepath.Join(work, "nearby"), "name: nearby\ndescription: nearby base\n")

	_, err := ResolveKB("")
	if !errors.Is(err, ErrNoKB) {
		t.Fatalf("expected ErrNoKB, got %v", err)
	}

	message := err.Error()
	for _, want := range []string{
		"nearby",
		filepath.Join(work, "nearby"),
		"nearby base",
		"--kb <path>",
		"AKB_KB=<path>",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("error %q does not contain %q", message, want)
		}
	}
}

func TestResolveKBNoSelectionWithoutNearbyBasesKeepsTheRuleText(t *testing.T) {
	scanEnvironment(t)
	t.Setenv(KBEnvVar, "")

	_, err := ResolveKB("")
	if err == nil {
		t.Fatal("expected ErrNoKB")
	}
	if err.Error() != ErrNoKB.Error() {
		t.Errorf("error = %q, want the bare rule text %q", err.Error(), ErrNoKB.Error())
	}
}
