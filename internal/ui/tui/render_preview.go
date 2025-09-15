package tui

import (
    "github.com/charmbracelet/lipgloss"
    "path/filepath"
    "strings"
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/preview"
)

// isImagePath returns true for common image file extensions.
func isImagePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".tiff", ".tif":
		return true
	default:
		return false
	}
}

// renderFilePreviewBody builds body lines for the right preview for a file path.
// It tries inline image rendering if supported, otherwise falls back to text preview.
func (m *model) renderFilePreviewBody(path string, width int, maxBodyLines int) []string {
	if width < 1 || maxBodyLines < 1 {
		return nil
	}
	// Inline image path: only if enabled in config and terminal supports it.
	if m.deps.Config != nil && m.deps.Config.InlineImages && isImagePath(path) && preview.SupportsIterm2() {
		if lines, ok := preview.BuildIterm2Inline(path, width, maxBodyLines); ok {
			return lines
		}
	}
	// Text fallback using provider (with cache)
	res, ok := m.fileCache.Get(path)
	if !ok {
		res, _ = m.prevProv.Preview(path, 8192)
		m.fileCache.Put(path, res)
	}
	out := []string{}
	for _, l := range strings.Split(res.Content, "\n") {
		ln := trimToWidth(l, width)
		pad := width - lipgloss.Width(ln)
		if pad > 0 {
			ln += strings.Repeat(" ", pad)
		}
		out = append(out, m.styNormal.Render(ln))
		if len(out) >= maxBodyLines {
			break
		}
	}
	return out
}
