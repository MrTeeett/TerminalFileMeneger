package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// message sent when an interactive external command finishes
type extRunDoneMsg struct{ err error }

// onCmdKey handles key input for command-line (Ex) mode.
func (m *model) onCmdKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.cmdActive = false
		return nil
	case tea.KeyEnter:
		cmd := strings.TrimSpace(string(m.cmdBuf))
		m.cmdActive = false
		m.cmdBuf = nil
		return m.execCommand(cmd)
	case tea.KeyBackspace, tea.KeyCtrlH:
		if len(m.cmdBuf) > 0 {
			m.cmdBuf = m.cmdBuf[:len(m.cmdBuf)-1]
		}
		return nil
	default:
		if len(msg.Runes) > 0 {
			m.cmdBuf = append(m.cmdBuf, msg.Runes...)
		} else {
			// handle space
			if msg.String() == "space" {
				m.cmdBuf = append(m.cmdBuf, ' ')
			}
		}
		return nil
	}
}

// execCommand parses and executes a command-line string.
func (m *model) execCommand(s string) tea.Cmd {
	if s == "" {
		return nil
	}
	if s[0] == ':' {
		s = s[1:]
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil
	}
	name := strings.ToLower(fields[0])
	args := fields[1:]
	switch name {
	case "help", "h", "?":
		m.showHelp()
		return nil
	case "theme":
		// Show current theme + color profile diagnostics and samples
		th := m.deps.Config.Theme
		lines := []string{
			"Color profile (used): " + m.colorProfile,
			"TERM=" + os.Getenv("TERM") + " COLORTERM=" + os.Getenv("COLORTERM"),
			"",
			"[theme.header] fg=" + th.Header.FG + " bg=" + th.Header.BG + fmt.Sprintf(" bold=%v faint=%v reverse=%v", th.Header.Bold, th.Header.Faint, th.Header.Reverse),
			"[theme.status] fg=" + th.Status.FG + " bg=" + th.Status.BG + fmt.Sprintf(" bold=%v faint=%v reverse=%v", th.Status.Bold, th.Status.Faint, th.Status.Reverse),
			"[theme.dir]    fg=" + th.Dir.FG + " bg=" + th.Dir.BG + fmt.Sprintf(" bold=%v faint=%v reverse=%v", th.Dir.Bold, th.Dir.Faint, th.Dir.Reverse),
			"[theme.normal] fg=" + th.Normal.FG + " bg=" + th.Normal.BG + fmt.Sprintf(" bold=%v faint=%v reverse=%v", th.Normal.Bold, th.Normal.Faint, th.Normal.Reverse),
			"[theme.selected] fg=" + th.Selected.FG + " bg=" + th.Selected.BG + fmt.Sprintf(" bold=%v faint=%v reverse=%v", th.Selected.Bold, th.Selected.Faint, th.Selected.Reverse),
			"",
			"Samples:",
		}
		// Add rendered samples to verify the applied styles
		lines = append(lines,
			m.styHeader.Render("Header sample"),
			m.styStatus.Render("Status sample"),
			m.styDir.Render("Directory/"),
			m.styNormal.Render("file.txt"),
			m.stySelected.Render("<selected>"),
		)
		m.modalTitle = "Theme"
		m.modalLines = lines
		m.modalActive = true
		return nil
	case "q", "quit", "exit":
		return tea.Quit
	case "cd":
		if len(args) == 0 {
			m.setError(fmt.Errorf("usage: :cd <path>"))
		}
		return nil
	case "copy":
		m.copySelectedFile()
		return nil
	case "paste":
		return m.pasteFiles()
	case "copy-path":
		m.copySelectedPath()
		return nil
	case "paste-path":
		return m.pastePath()
		path := args[0]
		if strings.HasPrefix(path, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				path = filepath.Join(home, strings.TrimPrefix(path, "~"))
			}
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(m.focused().panel.Cwd, path)
		}
		if fi, err := os.Stat(path); err == nil && fi.IsDir() {
			// change directory in focused panel
			t := m.focused()
			if err := t.panel.Chdir(path); err != nil {
				m.setError(err)
			} else {
				m.err = nil
				t.selected = 0
				t.scroll = 0
			}
		} else if err != nil {
			m.setError(err)
		} else {
			m.setError(fmt.Errorf("not a directory: %s", path))
		}
		return nil
	case "preview":
		if len(args) == 0 || args[0] == "toggle" {
			m.togglePreview()
		} else if args[0] == "on" {
			if !m.showPrev {
				m.togglePreview()
			}
		} else if args[0] == "off" {
			if m.showPrev {
				m.togglePreview()
			}
		}
		return nil
	case "opacity":
		if len(args) == 0 {
			m.modalTitle = "Opacity"
			m.modalLines = []string{fmt.Sprintf("background_opacity = %.0f%%", m.deps.Config.BackgroundOpacity*100)}
			m.modalActive = true
			return nil
		}
		val := args[0]
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			if f > 1 && f <= 100 {
				f = f / 100
			}
			if f < 0 {
				f = 0
			}
			if f > 1 {
				f = 1
			}
			if m.deps.Config != nil {
				m.deps.Config.BackgroundOpacity = f
			}
			m.computeStyles()
			m.refreshContent()
			return nil
		}
		m.setError(fmt.Errorf("usage: :opacity <0..1 | 0..100>"))
		return nil
	case "blur":
		if len(args) == 0 {
			m.modalTitle = "Blur"
			m.modalLines = []string{fmt.Sprintf("blur = %v (requires terminal/compositor support)", m.deps.Config.Blur)}
			m.modalActive = true
			return nil
		}
		on := strings.ToLower(args[0])
		switch on {
		case "on", "1", "true", "yes":
			if m.deps.Config != nil {
				m.deps.Config.Blur = true
			}
		case "off", "0", "false", "no":
			if m.deps.Config != nil {
				m.deps.Config.Blur = false
			}
		default:
			m.setError(fmt.Errorf("usage: :blur on|off"))
			return nil
		}
		// We cannot programmatically enable blur; hint only.
		m.refreshContent()
		return nil
	default:
		// Try custom command from config
		if m.deps.Config != nil {
			tpl, ok := m.deps.Config.CustomCommands[name]
			if ok {
				rawTpl := tpl
				cmdStr := strings.TrimSpace(tpl)
				// placeholders
				ft := m.focused()
				p := ft.panel
				var selName string
				var selPath string
				if len(p.Entries) > 0 && ft.selected >= 0 && ft.selected < len(p.Entries) {
					e := p.Entries[ft.selected]
					selName = e.Name
					selPath = filepath.Join(p.Cwd, e.Name)
				} else {
					selName = ""
					selPath = p.Cwd
				}
				rep := func(s string) string {
					s = strings.ReplaceAll(s, "{cwd}", shellQuote(p.Cwd))
					s = strings.ReplaceAll(s, "{file}", shellQuote(selName))
					s = strings.ReplaceAll(s, "{path}", shellQuote(selPath))
					return s
				}
				// Remember if template explicitly uses placeholders
				hasPh := strings.Contains(cmdStr, "{cwd}") || strings.Contains(cmdStr, "{file}") || strings.Contains(cmdStr, "{path}")
				cmdStr = rep(cmdStr)
				// If no placeholders were specified, append selected path (file or dir) by default.
				if !hasPh {
					// If there is a selected entry, use its path; otherwise use current directory.
					defArg := selPath
					if defArg == "" {
						defArg = p.Cwd
					}
					if defArg != "" {
						cmdStr = cmdStr + " " + shellQuote(defArg)
					}
				}
				// Determine if this should be run interactively (attach TTY)
				interactive := false
				if strings.HasPrefix(strings.TrimSpace(rawTpl), "!") || strings.HasPrefix(strings.TrimSpace(cmdStr), "!") {
					interactive = true
					cmdStr = strings.TrimSpace(strings.TrimPrefix(cmdStr, "!"))
				}
				// Common TUI/TTY apps we auto-run interactively
				switch strings.ToLower(name) {
				case "nano", "vim", "nvim", "vi", "less", "more", "micro":
					interactive = true
				}
				if interactive {
					// Suspend Bubble Tea, attach TTY and run interactively
					c := exec.Command("sh", "-c", cmdStr)
					return tea.ExecProcess(c, func(err error) tea.Msg { return extRunDoneMsg{err: err} })
				}
				// Non-interactive: capture output and show in modal
				return func() tea.Msg {
					out, err := runShell(cmdStr)
					m.modalLines = strings.Split(strings.TrimRight(out, "\n"), "\n")
					if err != nil {
						m.modalTitle = "Command error"
					} else {
						m.modalTitle = "Command output"
					}
					m.modalActive = true
					return nil
				}
			}
		}
		// Unknown command
		m.modalTitle = "Unknown command"
		m.modalLines = []string{"Unknown: " + name}
		m.modalActive = true
		return nil
	}
}

func (m *model) showHelp() {
	lines := []string{
		"Ex commands:",
		":help                 — показать эту справку",
		":q | :quit | :exit    — выйти",
		":cd <path>            — перейти в каталог",
		":preview on|off|toggle — управлять панелью предпросмотра",
		":copy                 — скопировать выделенный файл/папку (в буфер TFM)",
		":paste                — вставить в текущий каталог",
		":copy-path            — скопировать полный путь выделенного (в буфер TFM)",
		":paste-path           — перейти по скопанному пути (cd/выделить файл)",
		"",
		"Кастомные команды [commands] в config.toml:",
		"  name = \"shell snippet\"",
		"Подстановки: {cwd} {file} {path}",
		"Пример: open = \"xdg-open {path}\"",
		"",
		"Нажмите любую клавишу для закрытия",
	}
	m.modalTitle = "tfm help"
	m.modalLines = lines
	m.modalActive = true
}

// shellQuote returns a single-quoted string safe for sh -c.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	// Replace ' with '\'' sequence
	s = strings.ReplaceAll(s, "'", "'\\''")
	return "'" + s + "'"
}

func runShell(cmd string) (string, error) {
	c := exec.Command("sh", "-c", cmd)
	out, err := c.CombinedOutput()
	return string(out), err
}

// runInteractive executes a command attached to the current TTY (stdin/stdout/stderr).
// Suitable for editors/pagers like nano, vim, less.
// runInteractive was replaced by tea.ExecProcess; kept here previously, now removed.

func (m *model) setError(err error) {
	m.err = err
}
