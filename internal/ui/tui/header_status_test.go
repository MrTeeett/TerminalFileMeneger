package tui

import (
    "errors"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/charmbracelet/lipgloss"
    "github.com/muesli/termenv"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/clipboard"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/panels"
)

func TestHeaderContent_BasicAndNoTabs(t *testing.T) {
    // No tabs path
    m := &model{}
    if got := headerContent(m); got != "tfm | no tabs" {
        t.Fatalf("header no tabs = %q", got)
    }
    // With one tab
    p := &panels.Panel{Cwd: "/tmp"}
    m.tabs = []tab{{panel: p}}
    got := headerContent(m)
    if !strings.Contains(got, "tfm | tab 1/1 | /tmp") {
        t.Fatalf("header content unexpected: %q", got)
    }
}

func TestStatusContent_BasicFlagsFileAndError(t *testing.T) {
    lipgloss.SetColorProfile(termenv.ANSI)
    dir := t.TempDir()
    file := filepath.Join(dir, "a.txt")
    if err := os.WriteFile(file, []byte("hello"), 0o644); err != nil {
        t.Fatalf("write file: %v", err)
    }
    p := &panels.Panel{Cwd: dir, Entries: []panels.Entry{{Name: "a.txt", IsDir: false}}}
    m := &model{
        tabs:     []tab{{panel: p, selected: 0}},
        active:   0,
        showPrev: true,
        openRight: false,
        focus:    "left",
        rightMode: "panel",
        clip:     clipboard.New(),
        width:    120,
    }
    s := statusContent(m)
    // Should include file name and size in bytes
    if !strings.Contains(s, "a.txt | 5B |") {
        t.Fatalf("status missing filename/size: %q", s)
    }
    if !strings.Contains(s, "F:L") { // focus left
        t.Fatalf("status missing focus flag: %q", s)
    }
    if !strings.Contains(s, "RO:off") || !strings.Contains(s, "PV:on") { // flags from model
        t.Fatalf("status missing right/preview flags: %q", s)
    }
    // With error prefix
    m.err = errors.New("boom")
    s2 := statusContent(m)
    if !strings.HasPrefix(s2, "ERR: boom | ") {
        t.Fatalf("status should prefix error: %q", s2)
    }
}

func TestStatusContent_CmdActiveTrimsToWidth(t *testing.T) {
    // Provide minimal tab/panel so statusContent can compute focused panel
    p := &panels.Panel{Cwd: "/"}
    m := &model{width: 8, cmdActive: true, cmdBuf: []rune("abcdefghij"), tabs: []tab{{panel: p}}}
    got := statusContent(m)
    // ':' + 7 visible chars = 8 total width => ':abcdefg' with ellipsis if narrow,
    // trimToWidth uses ellipsis when w>1 and string longer than w; so for width=8
    // and len(':abcdefghij')>8, it will return first 7 runes + '…'
    if got != ":abcdef…" {
        t.Fatalf("cmdActive trim: %q", got)
    }
}

func TestComputeMaxDirName(t *testing.T) {
    // When there are directories, returns max name width + 1
    entries := []panels.Entry{{Name: "abc", IsDir: true}, {Name: "longer", IsDir: true}, {Name: "file", IsDir: false}}
    got := computeMaxDirName(entries)
    if got != len("longer")+1 {
        t.Fatalf("computeMaxDirName = %d; want %d", got, len("longer")+1)
    }
    // When no dirs, returns minimum 10
    entries = []panels.Entry{{Name: "f1", IsDir: false}}
    if got2 := computeMaxDirName(entries); got2 != 10 {
        t.Fatalf("computeMaxDirName (no dirs) = %d; want 10", got2)
    }
}

func TestStatusContent_PercentAndScroll(t *testing.T) {
    // Build panel with 10 entries, selected first
    entries := make([]panels.Entry, 10)
    for i := 0; i < 10; i++ { entries[i] = panels.Entry{Name: "nope", IsDir: false} }
    p := &panels.Panel{Cwd: "/", Entries: entries}
    m := &model{
        tabs:   []tab{{panel: p, selected: 0}},
        active: 0,
        width:  120,
    }
    // Simulate viewport of height 3 scrolled to offset 2 -> bottom = 5 -> 50%
    m.vp.Height = 3
    m.vp.YOffset = 2
    s := statusContent(m)
    if !strings.Contains(s, "[1/10]") {
        t.Fatalf("status missing sel/total: %q", s)
    }
    if !strings.Contains(s, " 50%  F:L") {
        t.Fatalf("status missing 50%% and focus L: %q", s)
    }
    // Focus right indicator
    m.focus = "right"
    m.rightMode = "panel"
    s2 := statusContent(m)
    if !strings.Contains(s2, "F:R") {
        t.Fatalf("status missing focus R: %q", s2)
    }
}
