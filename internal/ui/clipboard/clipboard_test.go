package clipboard

import "testing"

func TestClipboard(t *testing.T) {
    c := New()
    if c.Kind() != "" || len(c.Items()) != 0 {
        t.Fatalf("new clipboard not empty")
    }
    c.SetFiles([]string{"a", "b"})
    if c.Kind() != "file" { t.Fatalf("kind file expected") }
    if got := c.Items(); len(got) != 2 || got[0] != "a" || got[1] != "b" {
        t.Fatalf("items mismatch: %v", got)
    }
    c.SetPath("/tmp/x")
    if c.Kind() != "path" { t.Fatalf("kind path expected") }
    if got := c.Items(); len(got) != 1 || got[0] != "/tmp/x" { t.Fatalf("items %v", got) }
    c.Clear()
    if c.Kind() != "" || len(c.Items()) != 0 { t.Fatalf("clipboard not cleared") }
}

