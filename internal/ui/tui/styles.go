package tui

import "github.com/charmbracelet/lipgloss"

// computeStyles initializes styles based on theme and size.
func (m *model) computeStyles() {
	th := m.deps.Config.Theme
	transparent := m.deps.Config != nil && m.deps.Config.BackgroundOpacity < 1.0
	// Header
	m.styHeader = lipgloss.NewStyle()
	if th.Header.Bold {
		m.styHeader = m.styHeader.Bold(true)
	}
	if th.Header.Faint {
		m.styHeader = m.styHeader.Faint(true)
	}
	if th.Header.FG != "" {
		m.styHeader = m.styHeader.Foreground(lipgloss.Color(th.Header.FG))
	} else {
		color := m.deps.Theme.Primary
		if color == "" || color == "default" {
			color = "12"
		}
		m.styHeader = m.styHeader.Foreground(lipgloss.Color(color))
	}
	if th.Header.BG != "" && !transparent {
		m.styHeader = m.styHeader.Background(lipgloss.Color(th.Header.BG))
	}
	// Status
	m.styStatus = lipgloss.NewStyle()
	if th.Status.Bold {
		m.styStatus = m.styStatus.Bold(true)
	}
	if th.Status.Faint {
		m.styStatus = m.styStatus.Faint(true)
	}
	if th.Status.FG != "" {
		m.styStatus = m.styStatus.Foreground(lipgloss.Color(th.Status.FG))
	}
	if th.Status.BG != "" && !transparent {
		m.styStatus = m.styStatus.Background(lipgloss.Color(th.Status.BG))
	}
	// Normal
	m.styNormal = lipgloss.NewStyle()
	if th.Normal.Bold {
		m.styNormal = m.styNormal.Bold(true)
	}
	if th.Normal.Faint {
		m.styNormal = m.styNormal.Faint(true)
	}
	if th.Normal.FG != "" {
		m.styNormal = m.styNormal.Foreground(lipgloss.Color(th.Normal.FG))
	}
	if th.Normal.BG != "" && !transparent {
		m.styNormal = m.styNormal.Background(lipgloss.Color(th.Normal.BG))
	}
	// Selected
	m.stySelected = lipgloss.NewStyle()
	if th.Selected.Bold {
		m.stySelected = m.stySelected.Bold(true)
	}
	if th.Selected.Faint {
		m.stySelected = m.stySelected.Faint(true)
	}
	if th.Selected.Reverse {
		m.stySelected = m.stySelected.Reverse(true)
	}
	if th.Selected.FG != "" {
		m.stySelected = m.stySelected.Foreground(lipgloss.Color(th.Selected.FG))
	}
	if th.Selected.BG != "" && !transparent {
		m.stySelected = m.stySelected.Background(lipgloss.Color(th.Selected.BG))
	}
	// Dir
	m.styDir = lipgloss.NewStyle()
	if th.Dir.Bold {
		m.styDir = m.styDir.Bold(true)
	}
	if th.Dir.Faint {
		m.styDir = m.styDir.Faint(true)
	}
	if th.Dir.FG != "" {
		m.styDir = m.styDir.Foreground(lipgloss.Color(th.Dir.FG))
	} else {
		color := m.deps.Theme.Primary
		if color == "" || color == "default" {
			color = "12"
		}
		m.styDir = m.styDir.Foreground(lipgloss.Color(color))
	}
	if th.Dir.BG != "" && !transparent {
		m.styDir = m.styDir.Background(lipgloss.Color(th.Dir.BG))
	}
}
