package tui

import (
    "testing"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/panels"
)

func TestShellQuote(t *testing.T) {
    if got := shellQuote(""); got != "''" {
        t.Fatalf("shellQuote empty = %q; want ''", got)
    }
    in := "ab'cd"
    got := shellQuote(in)
    // raw string to avoid escapes; expected: 'ab'\''cd'
    want := `'ab'\''cd'`
    if got != want {
        t.Fatalf("shellQuote = %q; want %q", got, want)
    }
}

func TestTrimToWidth(t *testing.T) {
    if got := trimToWidth("hello", 10); got != "hello" {
        t.Fatalf("trimToWidth no trim = %q", got)
    }
    if got := trimToWidth("hello", 1); got != "h" {
        t.Fatalf("trimToWidth w=1 = %q; want 'h'", got)
    }
    if got := trimToWidth("hello", 4); got != "hel…" {
        t.Fatalf("trimToWidth w=4 = %q; want 'hel…'", got)
    }
}

func TestMergeColumns(t *testing.T) {
    cols := [][]string{{"a", "b"}, {"11"}}
    widths := []int{1, 2}
    sep := " "
    got := mergeColumns(cols, widths, sep)
    // Column 2 second line is sep + padded cell (width=2) => 3 spaces after 'b'
    want := "a 11\nb   "
    if got != want {
        t.Fatalf("mergeColumns = %q; want %q", got, want)
    }
}

func TestAutoColumnWidths(t *testing.T) {
    // Two columns, with MaxDirName hints
    c0 := tab{panel: &panels.Panel{MaxDirName: 12}}
    c1 := tab{panel: &panels.Panel{MaxDirName: 5}}
    w := autoColumnWidths([]tab{c0, c1}, 40, 1)
    if len(w) != 2 {
        t.Fatalf("widths len = %d; want 2", len(w))
    }
    // desired: c0=14, c1=10(min); remaining distribution -> [14, 25]
    if w[0] != 14 || w[1] != 25 {
        t.Fatalf("widths = %v; want [14 25]", w)
    }
}
