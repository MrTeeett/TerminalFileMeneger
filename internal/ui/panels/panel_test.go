package panels

import (
    "os"
    "path/filepath"
    "testing"
)

func TestRefreshSortAndHidden(t *testing.T) {
    dir := t.TempDir()
    // create structure
    if err := os.Mkdir(filepath.Join(dir, "Bdir"), 0o755); err != nil { t.Fatal(err) }
    if err := os.Mkdir(filepath.Join(dir, "Adir"), 0o755); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(dir, "z.txt"), []byte("z"), 0o644); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("h"), 0o644); err != nil { t.Fatal(err) }

    p := NewPanel(dir, false)
    if err := p.Refresh(); err != nil { t.Fatalf("Refresh: %v", err) }
    if p.Cwd != dir { t.Fatalf("Cwd = %q; want %q", p.Cwd, dir) }
    // Expect: directories first (sorted), then files; hidden excluded
    if len(p.Entries) != 3 { t.Fatalf("len(entries)=%d; want 3", len(p.Entries)) }
    if !(p.Entries[0].IsDir && p.Entries[0].Name == "Adir") { t.Fatalf("first entry = %#v", p.Entries[0]) }
    if !(p.Entries[1].IsDir && p.Entries[1].Name == "Bdir") { t.Fatalf("second entry = %#v", p.Entries[1]) }
    if p.Entries[2].IsDir || p.Entries[2].Name != "z.txt" { t.Fatalf("third entry = %#v", p.Entries[2]) }
    if p.MaxDirName != 5 { t.Fatalf("MaxDirName=%d; want 5", p.MaxDirName) }
}

func TestChdirFiltersAndMaxDir(t *testing.T) {
    root := t.TempDir()
    sub := filepath.Join(root, "dir")
    if err := os.Mkdir(sub, 0o755); err != nil { t.Fatal(err) }
    if err := os.Mkdir(filepath.Join(sub, "subdirlong"), 0o755); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(sub, ".secret"), []byte("x"), 0o644); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(sub, "a.txt"), []byte("a"), 0o644); err != nil { t.Fatal(err) }

    p := NewPanel(root, false)
    if err := p.Chdir(sub); err != nil { t.Fatalf("Chdir: %v", err) }
    if p.Cwd != sub { t.Fatalf("Cwd after chdir = %q; want %q", p.Cwd, sub) }
    // .secret should be filtered, subdirlong present, a.txt present
    names := []string{p.Entries[0].Name, p.Entries[1].Name}
    if !(names[0] == "subdirlong" && names[1] == "a.txt") {
        t.Fatalf("entries order = %v", names)
    }
    if p.MaxDirName < 11 { // subdirlong + '/'
        t.Fatalf("MaxDirName too small: %d", p.MaxDirName)
    }
}

func TestJoin(t *testing.T) {
    p := NewPanel("/tmp", false)
    if got := p.Join("x"); !filepath.IsAbs(got) || got[len(got)-1] != 'x' {
        t.Fatalf("Join returned %q", got)
    }
}

