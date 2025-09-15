package ops

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(b)
}

func TestCopyFile(t *testing.T) {
	m := NewManager()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	writeFile(t, src, "hello", 0o644)
	if err := m.Copy(context.Background(), src, dst); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if got := readFile(t, dst); got != "hello" {
		t.Fatalf("dst content = %q; want hello", got)
	}
	if fi, err := os.Stat(dst); err != nil || fi.Mode().Perm() != 0o644 {
		t.Fatalf("dst mode = %v, err=%v; want 0644", fi.Mode(), err)
	}
}

func TestCopyDirRecursive(t *testing.T) {
	m := NewManager()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFile(t, filepath.Join(src, "a", "b.txt"), strings.Repeat("x", 10), 0o600)
	if err := m.Copy(context.Background(), src, dst); err != nil {
		t.Fatalf("Copy dir: %v", err)
	}
	got := readFile(t, filepath.Join(dst, "a", "b.txt"))
	if len(got) != 10 {
		t.Fatalf("copied content len = %d; want 10", len(got))
	}
}

func TestCopySymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require privileges on Windows")
	}
	m := NewManager()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	writeFile(t, target, "T", 0o644)
	link := filepath.Join(dir, "link")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}
	dst := filepath.Join(dir, "dstlink")
	if err := m.Copy(context.Background(), link, dst); err != nil {
		t.Fatalf("Copy symlink: %v", err)
	}
	if got, err := os.Readlink(dst); err != nil || got != "target.txt" {
		t.Fatalf("dst symlink -> %q, err=%v; want target.txt", got, err)
	}
}

func TestMoveAndDelete(t *testing.T) {
	m := NewManager()
	dir := t.TempDir()
	src := filepath.Join(dir, "m.txt")
	dst := filepath.Join(dir, "moved.txt")
	writeFile(t, src, "m", 0o644)
	if err := m.Move(context.Background(), src, dst); err != nil {
		t.Fatalf("Move: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("src should not exist after move")
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("dst should exist after move: %v", err)
	}
	// Delete
	if err := m.Delete(context.Background(), filepath.Dir(dst)); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(dst)); !os.IsNotExist(err) {
		t.Fatalf("directory should be removed")
	}
}

func TestCopyContextCanceled(t *testing.T) {
	m := NewManager()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	// Non-empty to hit ctx check in loop
	writeFile(t, filepath.Join(src, "x.txt"), "x", 0o644)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before
	if err := m.Copy(ctx, src, dst); err == nil {
		t.Fatalf("expected cancellation error")
	}
}
