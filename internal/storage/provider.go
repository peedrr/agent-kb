package storage

import "context"

// StorageProvider abstracts file operations for the knowledge base.
type StorageProvider interface {
	Write(ctx context.Context, path string, data []byte) error
	Read(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	List(ctx context.Context, dir string, ext string) ([]string, error)
}
