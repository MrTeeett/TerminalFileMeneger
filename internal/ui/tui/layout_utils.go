package tui

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// autoColumnWidths computes column widths based on the longest directory name in each column.
func autoColumnWidths(cols []tab, totalW int, sep int) []int {
	n := len(cols)
	widths := make([]int, n)
	desired := make([]int, n)
	for i, c := range cols {
		maxLen := 10
		if c.panel != nil && c.panel.MaxDirName > 0 {
			maxLen = c.panel.MaxDirName
		} else if c.panel != nil {
			for _, e := range c.panel.Entries {
				if e.IsDir {
					l := lipgloss.Width(e.Name) + 1
					if l > maxLen {
						maxLen = l
					}
				}
			}
		}
		desired[i] = maxLen + 2
		if desired[i] < 10 {
			desired[i] = 10
		}
	}
	remaining := totalW - (n-1)*sep
	for i := 0; i < n; i++ {
		minForRest := (n - i - 1) * 10
		w := desired[i]
		if w > remaining-minForRest {
			w = remaining - minForRest
		}
		if w < 10 {
			w = 10
		}
		widths[i] = w
		remaining -= w
	}
	if remaining > 0 {
		widths[n-1] += remaining
	}
	return widths
}

func mergeColumns(cols [][]string, widths []int, sep string) string {
	maxLines := 0
	for _, c := range cols {
		if len(c) > maxLines {
			maxLines = len(c)
		}
	}
	out := make([]string, 0, maxLines)
	for i := 0; i < maxLines; i++ {
		row := ""
		for j, c := range cols {
			if j > 0 {
				row += sep
			}
			w := widths[j]
			cell := ""
			if i < len(c) {
				cell = c[i]
			} else {
				cell = strings.Repeat(" ", w)
			}
			row += cell
		}
		out = append(out, row)
	}
	return join(out, "\n")
}
