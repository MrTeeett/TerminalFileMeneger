package tui

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/charmbracelet/lipgloss"
    "github.com/muesli/termenv"
    "github.com/MrTeeett/TerminalFileMeneger/internal/config"
    uicache "github.com/MrTeeett/TerminalFileMeneger/internal/ui/cache"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/preview"
)

type stubProv struct{
    calls int
    res   preview.Result
}

func (s *stubProv) Preview(path string, maxBytes int) (preview.Result, error) {
    s.calls++
    _ = path; _ = maxBytes
    return s.res, nil
}

func TestRenderFilePreviewBody_TextFallbackAndCache(t *testing.T) {
    lipgloss.SetColorProfile(termenv.ANSI)
    m := &model{}
    m.styNormal = lipgloss.NewStyle() // no extra SGR
    m.fileCache = uicache.NewFileCache(64)
    sp := &stubProv{res: preview.Result{Kind: "text", Content: "AAA\nBBBBBB", Mime: "text/plain"}}
    m.prevProv = sp
    // Non-image path to force text fallback
    path := filepath.Join(t.TempDir(), "readme.txt")
    _ = os.WriteFile(path, []byte("dummy"), 0o644)
    out := m.renderFilePreviewBody(path, 4, 2)
    if len(out) != 2 { t.Fatalf("len(out)=%d; want 2", len(out)) }
    if out[0] != "AAA " { t.Fatalf("line0 = %q; want 'AAA '", out[0]) }
    if out[1] != "BBBB" { t.Fatalf("line1 = %q; want 'BBBB'", out[1]) }
    if sp.calls != 1 { t.Fatalf("provider calls = %d; want 1", sp.calls) }
    // Second call should hit cache, not increase calls
    _ = m.renderFilePreviewBody(path, 4, 2)
    if sp.calls != 1 { t.Fatalf("provider called again unexpectedly: %d", sp.calls) }
}

func TestRenderFilePreviewBody_InlineImageIterm2(t *testing.T) {
    lipgloss.SetColorProfile(termenv.ANSI)
    // Force SupportsIterm2 via env and enable config.InlineImages
    t.Setenv("TFM_INLINE", "iterm2")
    m := &model{}
    m.deps.Config = config.Default()
    m.deps.Config.InlineImages = true
    // Small fake png file (content does not matter)
    dir := t.TempDir()
    path := filepath.Join(dir, "pic.png")
    if err := os.WriteFile(path, []byte("PNG"), 0o644); err != nil { t.Fatal(err) }
    out := m.renderFilePreviewBody(path, 5, 3)
    if len(out) != 3 { t.Fatalf("len(out)=%d; want 3", len(out)) }
    if !strings.Contains(out[0], "\x1b]1337;") {
        t.Fatalf("missing iterm2 sequence in first line: %q", out[0])
    }
    // All lines should be of requested width
    for i, ln := range out {
        if len(ln) < 5 { // first line includes escape + padding, others are spaces
            t.Fatalf("line %d too short", i)
        }
    }
}

// helper to init an empty FileCache since zero-value lacks maps
// no extra helpers needed; using cache.NewFileCache in tests
