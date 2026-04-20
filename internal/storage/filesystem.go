package storage

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FilesystemProvider performs file I/O without git operations.
type FilesystemProvider struct {
	kbRoot string
}

func NewFilesystemProvider(kbRoot string) *FilesystemProvider {
	return &FilesystemProvider{kbRoot: kbRoot}
}

func (f *FilesystemProvider) Write(_ context.Context, path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create parent directories: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (f *FilesystemProvider) Read(_ context.Context, path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

func (f *FilesystemProvider) Delete(_ context.Context, path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (f *FilesystemProvider) Exists(_ context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("check file existence: %w", err)
}

func (f *FilesystemProvider) List(_ context.Context, dir string, ext string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ext != "" && !strings.HasSuffix(path, ext) {
			return nil
		}
		relPath, err := filepath.Rel(f.kbRoot, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relPath))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}
	sort.Strings(files)
	return files, nil
}
