package cache

import (
    "reflect"
    "testing"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/panels"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/preview"
)

func TestDirCacheEviction(t *testing.T) {
    c := NewDirCache(2)
    a := []panels.Entry{{Name: "a", IsDir: true}}
    b := []panels.Entry{{Name: "b", IsDir: false}}
    c.Put("a", a)
    c.Put("b", b)
    if !c.Has("a") || !c.Has("b") {
        t.Fatalf("cache should contain a and b")
    }
    c.Put("c", []panels.Entry{{Name: "c"}})
    if c.Has("a") {
        t.Fatalf("oldest entry a should be evicted")
    }
    if !c.Has("b") || !c.Has("c") {
        t.Fatalf("cache should contain b and c")
    }
    got := c.Get("b")
    if !reflect.DeepEqual(got, b) {
        t.Fatalf("Get(b) mismatch: %#v != %#v", got, b)
    }
}

func TestFileCacheEviction(t *testing.T) {
    c := NewFileCache(1)
    c.Put("x", preview.Result{Kind: "text", Content: "hello"})
    if !c.Has("x") {
        t.Fatalf("cache should contain x")
    }
    c.Put("y", preview.Result{Kind: "info", Content: "meta"})
    if c.Has("x") {
        t.Fatalf("x should be evicted")
    }
    if _, ok := c.Get("y"); !ok {
        t.Fatalf("y should exist")
    }
}
