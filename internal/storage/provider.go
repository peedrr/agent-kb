// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package storage

import "context"

// Provider abstracts file operations for the knowledge base.
type Provider interface {
	Write(ctx context.Context, path string, data []byte) error
	Read(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	List(ctx context.Context, dir string, ext string) ([]string, error)
}
