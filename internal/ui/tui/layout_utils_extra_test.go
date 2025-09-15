package tui

import (
    "testing"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/panels"
)

func TestJoinHelper(t *testing.T) {
    if got := join(nil, ","); got != "" { t.Fatalf("join nil = %q", got) }
    if got := join([]string{"x"}, ","); got != "x" { t.Fatalf("join single = %q", got) }
    if got := join([]string{"a","b","c"}, "-"); got != "a-b-c" { t.Fatalf("join many = %q", got) }
}

func TestAutoColumnWidthsEdgeSmallTotal(t *testing.T) {
    // totalW is too small to fit min 10 each; function still enforces per-col minimums
    c0 := tab{panel: &panels.Panel{MaxDirName: 20}}
    c1 := tab{panel: &panels.Panel{MaxDirName: 20}}
    c2 := tab{panel: &panels.Panel{MaxDirName: 20}}
    widths := autoColumnWidths([]tab{c0,c1,c2}, 25, 1)
    if len(widths) != 3 { t.Fatalf("len=%d", len(widths)) }
    if widths[0] < 10 || widths[1] < 10 || widths[2] < 10 {
        t.Fatalf("each col should be >=10, got %v", widths)
    }
}
