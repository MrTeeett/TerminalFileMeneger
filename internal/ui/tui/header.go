package tui

import "fmt"

func headerContent(m *model) string {
	if len(m.tabs) == 0 {
		return "tfm | no tabs"
	}
	ft := m.focused()
	cwd := ft.panel.Cwd
	return fmt.Sprintf("tfm | tab %d/%d | %s", m.active+1, len(m.tabs), cwd)
}
