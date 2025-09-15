package tui

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func renderPanelColumn(m *model, t tab, width int, focused bool) []string {
	p := t.panel
	lines := make([]string, 0, len(p.Entries))
	for i, e := range p.Entries {
		name := e.Name
		if e.IsDir {
			name += "/"
		}
		ln := trimToWidth(name, width)
		pad := width - lipgloss.Width(ln)
		if pad > 0 {
			ln += strings.Repeat(" ", pad)
		}
		if focused && i == t.selected {
			ln = m.stySelected.Render(ln)
		} else if e.IsDir {
			ln = m.styDir.Render(ln)
		} else {
			ln = m.styNormal.Render(ln)
		}
		lines = append(lines, ln)
	}
	return lines
}
