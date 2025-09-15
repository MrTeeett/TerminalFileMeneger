package tui

import (
    "strings"
    "testing"

    "github.com/charmbracelet/lipgloss"
    "github.com/muesli/termenv"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/panels"
)

func TestRenderPanelColumn_StylesAndWidth(t *testing.T) {
    // Ensure lipgloss emits SGR sequences for styles
    lipgloss.SetColorProfile(termenv.ANSI)

    m := &model{}
    // Make directory entries bold, normal entries unstyled
    m.styDir = lipgloss.NewStyle().Bold(true)
    // selected style not used in this test
    m.stySelected = lipgloss.NewStyle()
    m.styNormal = lipgloss.NewStyle()

    p := &panels.Panel{Entries: []panels.Entry{{Name: "dir", IsDir: true}, {Name: "file.txt", IsDir: false}}}
    ttab := tab{panel: p, selected: 0}
    lines := renderPanelColumn(m, ttab, 6, false)
    if len(lines) != 2 {
        t.Fatalf("lines = %d; want 2", len(lines))
    }
    // Width should be exactly 6 printable cells for each line
    for i, ln := range lines {
        if lipgloss.Width(ln) != 6 {
            t.Fatalf("line %d width=%d; want 6 (ln=%q)", i, lipgloss.Width(ln), ln)
        }
    }
    // First line is directory with trailing '/', should be bold (SGR 1)
    if !strings.Contains(lines[0], "\x1b[1m") {
        t.Fatalf("directory line not bold: %q", lines[0])
    }
    // Second line is normal file, should not contain SGR if unstyled
    if strings.Contains(lines[1], "\x1b[") {
        t.Fatalf("normal line unexpectedly styled: %q", lines[1])
    }
}

func TestRenderPanelColumn_SelectedGetsSelectedStyle(t *testing.T) {
    lipgloss.SetColorProfile(termenv.ANSI)

    m := &model{}
    // Use distinct styles to detect selection vs dir
    m.styDir = lipgloss.NewStyle().Faint(true)      // SGR 2
    m.stySelected = lipgloss.NewStyle().Bold(true)  // SGR 1
    m.styNormal = lipgloss.NewStyle()

    p := &panels.Panel{Entries: []panels.Entry{{Name: "dir", IsDir: true}, {Name: "file", IsDir: false}}}
    ttab := tab{panel: p, selected: 1}
    lines := renderPanelColumn(m, ttab, 5, true) // focused: selected style should apply
    if len(lines) != 2 {
        t.Fatalf("lines = %d; want 2", len(lines))
    }
    // Directory (line 0) should be faint (SGR 2), file selected (line 1) should be bold (SGR 1)
    if !strings.Contains(lines[0], "\x1b[2m") {
        t.Fatalf("dir line not faint-styled: %q", lines[0])
    }
    if !strings.Contains(lines[1], "\x1b[1m") {
        t.Fatalf("selected line not bold-styled: %q", lines[1])
    }
    // Printable widths preserved
    for i, ln := range lines {
        if lipgloss.Width(ln) != 5 {
            t.Fatalf("line %d width=%d; want 5", i, lipgloss.Width(ln))
        }
    }
}
