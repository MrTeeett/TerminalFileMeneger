package ops

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Manager manages file operations.
type Manager struct{}

func NewManager() Manager { return Manager{} }

// Copy copies a file or directory recursively from src to dst.
func (m Manager) Copy(ctx context.Context, src, dst string) error {
	fi, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		// For simplicity: copy symlink as symlink
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		_ = os.RemoveAll(dst)
		return os.Symlink(target, dst)
	}
	if fi.IsDir() {
		// Create destination dir
		if err := os.MkdirAll(dst, fi.Mode().Perm()); err != nil {
			return err
		}
		// Walk contents
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			s := filepath.Join(src, e.Name())
			d := filepath.Join(dst, e.Name())
			if err := m.Copy(ctx, s, d); err != nil {
				return err
			}
		}
		return nil
	}
	// Regular file copy
	return copyFile(src, dst, fi)
}

func copyFile(src, dst string, fi fs.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fi.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func (m Manager) Move(ctx context.Context, src, dst string) error {
	_ = ctx
	return os.Rename(src, dst)
}

func (m Manager) Delete(ctx context.Context, path string) error {
	_ = ctx
	return os.RemoveAll(path)
}
