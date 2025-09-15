package tui

import (
	"fmt"
	"os"
	"path/filepath"
)

func statusContent(m *model) string {
	ft := m.focused()
	p := ft.panel
	n := len(p.Entries)
	sel := 0
	if n > 0 {
		sel = ft.selected + 1
	}
	bottom := m.vp.YOffset + m.vp.Height
	if bottom > n {
		bottom = n
	}
	pct := 0
	if n > 0 {
		pct = (bottom * 100) / n
	}
	ro := "off"
	if m.openRight {
		ro = "on"
	}
	pv := "off"
	if m.showPrev {
		pv = "on"
	}
	fc := "L"
	if m.rightMode == "panel" && m.focus == "right" {
		fc = "R"
	}
	status := fmt.Sprintf("[%d/%d] %3d%%  F:%s RO:%s PV:%s CB:%s  [j/k] move  [h] up  [l/Enter] open  [.] hidden  [yy] copy  [pp] paste  [Y] copy-path  [P] paste-path  [:] cmd  [q] quit", sel, n, pct, fc, ro, pv, m.clip.Kind())
	if n > 0 {
		e := p.Entries[ft.selected]
		fpath := filepath.Join(p.Cwd, e.Name)
		if fi, err := os.Stat(fpath); err == nil {
			if !fi.IsDir() {
				status = fmt.Sprintf("%s | %dB | %s", e.Name, fi.Size(), status)
			} else {
				status = fmt.Sprintf("%s | %s", e.Name, status)
			}
		} else {
			status = fmt.Sprintf("%s | %s", e.Name, status)
		}
	}
	if m.err != nil {
		status = fmt.Sprintf("ERR: %s | %s", m.err.Error(), status)
	}
	if m.cmdActive {
		prompt := ":" + string(m.cmdBuf)
		return trimToWidth(prompt, m.width)
	}
	return status
}
