package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/MrTeeett/TerminalFileMeneger/internal/config"
	"github.com/MrTeeett/TerminalFileMeneger/internal/fs/ops"
	"github.com/MrTeeett/TerminalFileMeneger/internal/input/layout"
	"github.com/MrTeeett/TerminalFileMeneger/internal/keymap"
	"github.com/MrTeeett/TerminalFileMeneger/internal/logging"
	"github.com/MrTeeett/TerminalFileMeneger/internal/theme"
	"github.com/MrTeeett/TerminalFileMeneger/internal/ui/cache"
	"github.com/MrTeeett/TerminalFileMeneger/internal/ui/clipboard"
	"github.com/MrTeeett/TerminalFileMeneger/internal/ui/commands"
	"github.com/MrTeeett/TerminalFileMeneger/internal/ui/panels"
	"github.com/MrTeeett/TerminalFileMeneger/internal/ui/preview"
)

// Dependencies wired by app.Run. Keep this small and explicit.
type Dependencies struct {
	Logger   logging.Logger
	Config   *config.Config
	Keymap   keymap.Map
	Theme    theme.Theme
	Registry commands.Registry
	FS       ops.Manager
}

// tab holds state for a single tab/panel.
type tab struct {
	panel    *panels.Panel
	selected int
	scroll   int
}

// model is the Bubble Tea model for the app.
type model struct {
	deps   Dependencies
	tabs   []tab
	active int
	width  int
	height int
	err    error
	vp     viewport.Model
	header int
	status int
	// command-line (Ex) mode
	cmdActive bool
	cmdBuf    []rune
	// simple modal overlay (help/command output)
	modalActive bool
	modalTitle  string
	modalLines  []string
	// key chords
	keySeq []string
	seqGen int
	// focus and right area
	focus     string // "left" or "right" (right focuses the last column)
	rightMode string // "" | "preview" | "panel"
	rightT    tab    // navigable right panel when in panel mode (legacy single)
	rightCols []tab  // dynamic chain of right-side columns
	showPrev  bool   // preview enabled flag
	openRight bool   // open dirs on right flag
	rightPct  int    // width percent (10..80)
	prevProv  preview.Provider
	// profiling
	prof profiler
	// caches to speed up preview and listings
	dirCache  cache.DirCache
	fileCache cache.FileCache
	// async prefetch tracker to avoid duplicate work
	prefetching map[string]struct{}
	// styles
	styHeader   lipgloss.Style
	styStatus   lipgloss.Style
	styNormal   lipgloss.Style
	stySelected lipgloss.Style
	styDir      lipgloss.Style
	// diagnostics
	colorProfile string
	// clipboard
	clip clipboard.Clipboard
}

// profiler is a tiny helper to log step timings when enabled.
type profiler struct {
	enabled bool
	start   time.Time
	last    time.Time
	logger  logging.Logger
}

func (p *profiler) Begin(label string) {
	if !p.enabled {
		return
	}
	p.start = time.Now()
	p.last = p.start
	p.logger.Debugf("prof %s: begin", label)
}
func (p *profiler) Step(label, name string) {
	if !p.enabled {
		return
	}
	now := time.Now()
	p.logger.Debugf("prof %s: %-16s +%s (total %s)", label, name, now.Sub(p.last), now.Sub(p.start))
	p.last = now
}
func (p *profiler) End(label string) {
	if !p.enabled {
		return
	}
	p.logger.Debugf("prof %s: end total=%s", label, time.Since(p.start))
}

func initialModel(deps Dependencies) (model, error) {
	// Ensure a reasonable color profile is selected (config/env)
	usedProfile := configureColorProfile(deps.Config)
	p := panels.NewPanel("", deps.Config.ShowHidden)
	if err := p.Refresh(); err != nil {
		return model{}, err
	}
	m := model{
		deps:   deps,
		tabs:   []tab{{panel: p}},
		active: 0,
		header: 1,
		status: 1,
	}
	m.vp = viewport.Model{}
	// right area defaults
	m.showPrev = deps.Config.ShowPreview
	m.openRight = deps.Config.OpenDirsRight
	m.rightPct = deps.Config.RightPaneWidth
	if m.rightPct < 10 {
		m.rightPct = 10
	}
	if m.rightPct > 80 {
		m.rightPct = 80
	}
	if m.showPrev {
		m.rightMode = "preview"
	}
	m.prevProv = preview.BasicProvider{}
	// profiling controlled by env var TFM_PROFILE (any non-empty value)
	m.prof = profiler{enabled: os.Getenv("TFM_PROFILE") != "", logger: deps.Logger}
	// init caches
	m.dirCache = cache.NewDirCache(64)
	m.fileCache = cache.NewFileCache(64)
	m.prefetching = make(map[string]struct{})
	m.focus = "left"
	m.computeStyles()
	m.colorProfile = usedProfile
	m.refreshContent()
	return m, nil
}

// configureColorProfile sets lipgloss color profile based on env/TERM, with override via TFM_COLOR.
func configureColorProfile(cfg *config.Config) string {
	// Config override first
	if cfg != nil {
		v := strings.ToLower(strings.TrimSpace(cfg.ColorProfile))
		switch v {
		case "none", "off", "mono", "ascii":
			lipgloss.SetColorProfile(termenv.Ascii)
			return "ascii"
		case "ansi":
			lipgloss.SetColorProfile(termenv.ANSI)
			return "ansi"
		case "256", "ansi256":
			lipgloss.SetColorProfile(termenv.ANSI256)
			return "ansi256"
		case "truecolor", "24bit":
			lipgloss.SetColorProfile(termenv.TrueColor)
			return "truecolor"
		}
	}
	v := strings.ToLower(os.Getenv("TFM_COLOR"))
	switch v {
	case "", "auto":
		// auto-detect
	case "none", "off", "mono", "ascii":
		lipgloss.SetColorProfile(termenv.Ascii)
		return "ascii"
	case "ansi":
		lipgloss.SetColorProfile(termenv.ANSI)
		return "ansi"
	case "256", "ansi256":
		lipgloss.SetColorProfile(termenv.ANSI256)
		return "ansi256"
	case "truecolor", "24bit":
		lipgloss.SetColorProfile(termenv.TrueColor)
		return "truecolor"
	}
	// Auto detection by common envs
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	term := strings.ToLower(os.Getenv("TERM"))
	switch {
	case strings.Contains(term, "kitty"): // kitty supports true color
		lipgloss.SetColorProfile(termenv.TrueColor)
		return "truecolor"
	case strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit"):
		lipgloss.SetColorProfile(termenv.TrueColor)
		return "truecolor"
	case strings.Contains(term, "256color"):
		lipgloss.SetColorProfile(termenv.ANSI256)
		return "ansi256"
	default:
		lipgloss.SetColorProfile(termenv.ANSI)
		return "ansi"
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		// Configure viewport on first size message and subsequent resizes.
		contentH := m.height - m.header - m.status
		if contentH < 1 {
			contentH = 1
		}
		m.vp.Width = m.width
		m.vp.Height = contentH
		m.computeStyles()
		// Keep cursor visible after resize.
		m.ensureVisible()
		m.refreshContent()
		return m, nil
	case tea.KeyMsg:
		// If a modal is active, close it on Esc/Enter/any key (except modifiers)
		if m.modalActive {
			s := msg.String()
			if s != "" {
				m.modalActive = false
				m.refreshContent()
				return m, nil
			}
		}
		// Command-line input mode
		if m.cmdActive {
			if cmd := m.onCmdKey(msg); cmd != nil {
				return m, cmd
			}
			m.refreshContent()
			return m, nil
		}
		if m.prof.enabled {
			m.prof.Begin("update:" + msg.String())
		}
		cmd := m.onKey(msg)
		m.refreshContent()
		if m.prof.enabled {
			m.prof.End("update:" + msg.String())
		}
		return m, cmd
	case chordTimeoutMsg:
		cmd := m.onChordTimeout(msg)
		m.refreshContent()
		return m, cmd
	case dirPrefetchMsg:
		// store prefetched directory listing
		delete(m.prefetching, msg.path)
		if !m.dirCache.Has(msg.path) {
			m.dirCache.Put(msg.path, msg.entries)
		}
		m.refreshContent()
		return m, nil
	case tea.MouseMsg:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	case extRunDoneMsg:
		// After interactive command exits, refresh UI and report error if any.
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
		}
		// Reconfigure color profile and recompute styles (in case capabilities changed)
		m.colorProfile = configureColorProfile(m.deps.Config)
		m.computeStyles()
		// Refresh focused panel contents (FS may have changed)
		if ft := m.focused(); ft != nil && ft.panel != nil {
			_ = ft.panel.Refresh()
		}
		m.refreshContent()
		return m, nil
	}
	return m, nil
}

// chordTimeoutMsg is emitted when we should resolve a pending key chord.
type chordTimeoutMsg struct{ gen int }

// dirPrefetchMsg delivers async directory listing results to the model.
type dirPrefetchMsg struct {
	path    string
	entries []panels.Entry
}

const chordTimeout = 350 * time.Millisecond

func (m *model) onChordTimeout(msg chordTimeoutMsg) tea.Cmd {
	if msg.gen != m.seqGen {
		// Stale timer
		return nil
	}
	// Decide based on current sequence
	if act, exact, _ := m.matchSequence(m.keySeq); exact {
		if cmd := m.doAction(act); cmd != nil {
			m.keySeq = nil
			return cmd
		}
	}
	// Clear incomplete/unknown sequence
	m.keySeq = nil
	return nil
}

func (m *model) onKey(msg tea.KeyMsg) tea.Cmd {
	key := normalizeKey(msg.String())
	// Append to current sequence and try resolve.
	m.keySeq = append(m.keySeq, key)

	act, exact, hasPrefix := m.matchSequence(m.keySeq)
	switch {
	case exact && !hasPrefix:
		// Unambiguous exact match: execute immediately.
		m.keySeq = nil
		m.seqGen++ // invalidate any pending timers
		return m.doAction(act)
	case hasPrefix:
		// Wait for more input (or timeout). If also exact, timeout will trigger it.
		m.seqGen++
		gen := m.seqGen
		return tea.Tick(chordTimeout, func(time.Time) tea.Msg { return chordTimeoutMsg{gen: gen} })
	case !exact && !hasPrefix:
		// No match — reset buffer and retry with single key.
		m.keySeq = []string{key}
		if act2, exact2, hasPrefix2 := m.matchSequence(m.keySeq); exact2 {
			m.keySeq = nil
			m.seqGen++ // invalidate any pending timers
			return m.doAction(act2)
		} else if hasPrefix2 {
			m.seqGen++
			gen := m.seqGen
			return tea.Tick(chordTimeout, func(time.Time) tea.Msg { return chordTimeoutMsg{gen: gen} })
		}
		// Still nothing: drop.
		m.keySeq = nil
		return nil
	}
	return nil
}

// normalizeKey maps keys from non-English layouts (currently Cyrillic RU) to
// their physical US QWERTY equivalents so bindings like "h/j/k/l" work
// regardless of active layout. Non-letter keys and chord tokens remain as-is.
func normalizeKey(s string) string { return layout.NormalizeToken(s) }

// Small LRU helpers for caches
const cacheMaxDirs = 64
const cacheMaxFiles = 64

// legacy cache helpers were removed; using cache.DirCache/FileCache instead

// matchSequence checks current binding set for an exact match and/or prefix.
// Returns the action for exact match (if any), whether exact match exists and whether any binding has the given sequence as prefix.
func (m *model) matchSequence(seq []string) (keymap.Action, bool, bool) {
	var (
		exactAct keymap.Action
		hasExact bool
		hasPref  bool
	)
	bind := m.deps.Keymap.Normal
	for spec, act := range bind {
		toks := keymap.TokensOf(spec)
		// Check prefix
		if len(toks) >= len(seq) {
			ok := true
			for i := range seq {
				if toks[i] != seq[i] {
					ok = false
					break
				}
			}
			if ok {
				if len(toks) == len(seq) {
					hasExact = true
					exactAct = act
				} else {
					hasPref = true
				}
			}
		}
	}
	return exactAct, hasExact, hasPref
}

// doAction performs a semantic action according to the keymap.
func (m *model) doAction(act keymap.Action) tea.Cmd {
	switch string(act) {
	case "copy":
		m.copySelectedFile()
	case "paste":
		return m.pasteFiles()
	case "copy-path":
		m.copySelectedPath()
	case "paste-path":
		return m.pastePath()
	case "command":
		m.cmdActive = true
		m.cmdBuf = nil
		return nil
	case "quit":
		return tea.Quit
	case "down":
		m.move(1)
		return m.maybePrefetchSelected()
	case "up":
		m.move(-1)
		return m.maybePrefetchSelected()
	case "left":
		// Go to parent directory
		m.up()
		return m.maybePrefetchSelected()
	case "right":
		// Enter/open
		m.enter()
		return m.maybePrefetchSelected()
	case "top":
		m.setSelected(0)
		return m.maybePrefetchSelected()
	case "bottom":
		if n := m.entriesCount(); n > 0 {
			m.setSelected(n - 1)
		}
		return m.maybePrefetchSelected()
	case "toggle-hidden":
		m.toggleHidden()
		return m.maybePrefetchSelected()
	case "toggle-preview":
		m.togglePreview()
	case "toggle-right-open-mode":
		m.toggleOpenRightMode()
	case "close-right":
		m.closeRight()
	case "toggle-focus":
		if m.rightMode == "panel" && m.rightT.panel != nil {
			if m.focus == "left" {
				m.focus = "right"
			} else {
				m.focus = "left"
			}
		} else {
			m.focus = "left"
		}
	case "focus-left":
		m.focus = "left"
	case "focus-right":
		if m.rightMode == "panel" && m.rightT.panel != nil {
			m.focus = "right"
		}
	case "new-tab":
		m.newTab()
	case "next-tab":
		m.nextTab()
	case "prev-tab":
		m.prevTab()
	case "close-tab":
		m.closeTab()
	case "page-down":
		m.page(1)
		return m.maybePrefetchSelected()
	case "page-up":
		m.page(-1)
		return m.maybePrefetchSelected()
	case "half-page-down":
		m.halfPage(1)
		return m.maybePrefetchSelected()
	case "half-page-up":
		m.halfPage(-1)
		return m.maybePrefetchSelected()
	default:
		// Unimplemented actions (rename, delete, copy, paste, filter, fuzzy) are ignored for now.
	}
	return nil
}

func (m model) View() string {
	if len(m.tabs) == 0 {
		return "no tabs"
	}
	// Header
	header := headerContent(&m)
	// Layout sizes
	if m.width <= 0 {
		m.width = 80
	}
	if m.height <= 0 {
		m.height = 24
	}
	status := statusContent(&m)

	// Compose full view: header, viewport view, status
	lines := []string{
		m.styHeader.Render(trimToWidth(header, m.width)),
		m.vp.View(),
		m.styStatus.Render(trimToWidth(status, m.width)),
	}
	return join(lines, "\n")
}

func (m *model) current() *tab { return &m.tabs[m.active] }

func (m *model) entriesCount() int { return len(m.focused().panel.Entries) }

func (m *model) focused() *tab {
	if m.focus == "right" && m.rightMode == "panel" {
		if n := len(m.rightCols); n > 0 {
			return &m.rightCols[n-1]
		}
		if m.rightT.panel != nil {
			return &m.rightT
		}
	}
	return m.current()
}

func (m *model) setSelected(idx int) {
	t := m.focused()
	if idx < 0 {
		idx = 0
	}
	n := len(t.panel.Entries)
	if n == 0 {
		t.selected = 0
		m.ensureVisible()
		return
	}
	if idx >= n {
		idx = n - 1
	}
	t.selected = idx
	m.ensureVisible()
}

func (m *model) move(delta int) {
	m.setSelected(m.focused().selected + delta)
}

func (m *model) page(delta int) {
	step := m.viewportHeight()
	if step < 1 {
		step = 1
	}
	m.setSelected(m.focused().selected + delta*step)
}

func (m *model) halfPage(delta int) {
	step := m.viewportHeight() / 2
	if step < 1 {
		step = 1
	}
	m.setSelected(m.focused().selected + delta*step)
}

func (m *model) toggleHidden() {
	p := m.focused().panel
	p.ShowHidden = !p.ShowHidden
	_ = p.Refresh()
	m.setSelected(0)
}

func (m *model) enter() {
	m.prof.Begin("enter")
	t := m.focused()
	p := t.panel
	if len(p.Entries) == 0 {
		m.prof.End("enter")
		return
	}
	e := p.Entries[t.selected]
	if e.IsDir {
		newPath := filepath.Join(p.Cwd, e.Name)
		if m.openRight {
			// open as a new column and focus it
			m.prof.Step("enter", "open-right")
			m.openRightPanel(newPath)
			m.focus = "right"
		} else {
			m.prof.Step("enter", "chdir")
			if entries := m.dirCache.Get(newPath); entries != nil {
				// Use cached listing to avoid extra I/O
				p.Cwd = newPath
				p.Entries = entries
				p.MaxDirName = computeMaxDirName(entries)
				m.err = nil
				t.selected = 0
				t.scroll = 0
			} else if err := p.Chdir(newPath); err != nil {
				m.err = err
			} else {
				m.err = nil
				t.selected = 0
				t.scroll = 0
			}
		}
	}
	m.prof.End("enter")
}

func (m *model) up() {
	// If exploring in right columns, closing current column returns to previous
	if m.focus == "right" && m.rightMode == "panel" {
		if n := len(m.rightCols); n > 0 {
			// pop last column
			m.rightCols = m.rightCols[:n-1]
			if len(m.rightCols) == 0 {
				// No more columns — switch focus back to left and restore preview if enabled
				m.focus = "left"
				if m.showPrev {
					m.rightMode = "preview"
				} else {
					m.rightMode = ""
				}
			}
			return
		}
	}
	// Otherwise go to parent in the focused panel
	t := m.focused()
	p := t.panel
	parent := filepath.Dir(p.Cwd)
	if parent == p.Cwd {
		return
	}
	if err := p.Chdir(parent); err != nil {
		m.err = err
	} else {
		m.err = nil
		t.selected = 0
		t.scroll = 0
	}
}

func (m *model) newTab() {
	// Clone current CWD and ShowHidden
	curr := m.current().panel
	p := panels.NewPanel(curr.Cwd, curr.ShowHidden)
	_ = p.Refresh()
	m.tabs = append(m.tabs, tab{panel: p})
	m.active = len(m.tabs) - 1
	m.setSelected(0)
}

func (m *model) closeTab() {
	if len(m.tabs) <= 1 {
		return
	}
	idx := m.active
	m.tabs = append(m.tabs[:idx], m.tabs[idx+1:]...)
	if m.active >= len(m.tabs) {
		m.active = len(m.tabs) - 1
	}
}

func (m *model) nextTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.active = (m.active + 1) % len(m.tabs)
	m.ensureVisible()
	m.refreshContent()
}

func (m *model) openRightPanel(path string) {
	p := panels.NewPanel(path, m.deps.Config.ShowHidden)
	if entries := m.dirCache.Get(path); entries != nil {
		p.Cwd = path
		p.Entries = entries
		p.MaxDirName = computeMaxDirName(entries)
	} else {
		_ = p.Refresh()
	}
	m.rightCols = append(m.rightCols, tab{panel: p})
	m.rightMode = "panel"
}

func (m *model) closeRight() {
	// Close all right columns and revert focus
	m.rightCols = nil
	m.rightT = tab{}
	if m.focus == "right" {
		m.focus = "left"
	}
	if m.showPrev {
		m.rightMode = "preview"
	} else {
		m.rightMode = ""
	}
}

func (m *model) togglePreview() {
	m.showPrev = !m.showPrev
	if m.showPrev {
		m.rightMode = "preview"
		// No interactive right column in preview-only mode
		m.focus = "left"
	} else if m.rightMode == "preview" {
		m.rightMode = ""
		m.focus = "left"
	}
}

func (m *model) toggleOpenRightMode() { m.openRight = !m.openRight }

func (m *model) prevTab() {
	if len(m.tabs) == 0 {
		return
	}
	m.active = (m.active - 1 + len(m.tabs)) % len(m.tabs)
	m.ensureVisible()
	m.refreshContent()
}

func join(ss []string, sep string) string {
	switch len(ss) {
	case 0:
		return ""
	case 1:
		return ss[0]
	}
	// Use a builder to avoid O(n^2) concatenations
	var b strings.Builder
	// Rough capacity hint: sum of lengths + separators
	total := 0
	for _, s := range ss {
		total += len(s)
	}
	total += (len(ss) - 1) * len(sep)
	b.Grow(total)
	b.WriteString(ss[0])
	for i := 1; i < len(ss); i++ {
		b.WriteString(sep)
		b.WriteString(ss[i])
	}
	return b.String()
}

func (m *model) viewportHeight() int {
	// Reserve 1 line for header and 1 for status.
	h := m.height - m.header - m.status
	if h < 1 {
		h = 1
	}
	return h
}

func (m *model) ensureVisible() {
	t := m.current()
	if m.focus == "right" && m.rightMode == "panel" {
		if n := len(m.rightCols); n > 0 {
			t = &m.rightCols[n-1]
		} else if m.rightT.panel != nil {
			t = &m.rightT
		}
	}
	vh := m.viewportHeight()
	// Sync viewport height too.
	if m.vp.Height != vh && vh > 0 {
		m.vp.Height = vh
	}
	// Adjust viewport offset to keep selection visible.
	if t.selected < m.vp.YOffset {
		m.vp.YOffset = t.selected
	} else if t.selected >= m.vp.YOffset+vh {
		m.vp.YOffset = t.selected - vh + 1
	}
	if m.vp.YOffset < 0 {
		m.vp.YOffset = 0
	}
}

func (m *model) copySelectedFile() {
	t := m.focused()
	if t == nil || t.panel == nil || len(t.panel.Entries) == 0 {
		return
	}
	if t.selected < 0 || t.selected >= len(t.panel.Entries) {
		return
	}
	e := t.panel.Entries[t.selected]
	src := filepath.Join(t.panel.Cwd, e.Name)
	m.clip.SetFiles([]string{src})
}

func (m *model) copySelectedPath() {
	t := m.focused()
	if t == nil || t.panel == nil {
		return
	}
	var p string
	if len(t.panel.Entries) > 0 && t.selected >= 0 && t.selected < len(t.panel.Entries) {
		e := t.panel.Entries[t.selected]
		p = filepath.Join(t.panel.Cwd, e.Name)
	} else {
		p = t.panel.Cwd
	}
	m.clip.SetPath(p)
}

func (m *model) pasteFiles() tea.Cmd {
	if m.clip.Kind() != "file" || len(m.clip.Items()) == 0 {
		return nil
	}
	destDir := m.focused().panel.Cwd
	items := m.clip.Items()
	fs := m.deps.FS
	return func() tea.Msg {
		ctx := context.Background()
		for _, src := range items {
			base := filepath.Base(src)
			dst := uniqueDestPath(destDir, base)
			if err := fs.Copy(ctx, src, dst); err != nil {
				return extRunDoneMsg{err: err}
			}
		}
		return extRunDoneMsg{err: nil}
	}
}

func uniqueDestPath(dir, name string) string {
	dst := filepath.Join(dir, name)
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return dst
	}
	ext := ""
	base := name
	if i := strings.LastIndex(name, "."); i > 0 {
		base = name[:i]
		ext = name[i:]
	}
	for i := 1; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s copy %d%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	// fallback
	return filepath.Join(dir, fmt.Sprintf("%s copy%s", base, ext))
}

func (m *model) pastePath() tea.Cmd {
	if m.clip.Kind() != "path" || len(m.clip.Items()) == 0 {
		return nil
	}
	p := m.clip.Items()[0]
	return func() tea.Msg {
		fi, err := os.Stat(p)
		if err != nil {
			return extRunDoneMsg{err: err}
		}
		if fi.IsDir() {
			// navigate to directory
			t := m.focused()
			if err := t.panel.Chdir(p); err != nil {
				return extRunDoneMsg{err: err}
			}
			return extRunDoneMsg{err: nil}
		}
		// navigate to parent and select file
		dir := filepath.Dir(p)
		base := filepath.Base(p)
		t := m.focused()
		if err := t.panel.Chdir(dir); err != nil {
			return extRunDoneMsg{err: err}
		}
		for i, e := range t.panel.Entries {
			if e.Name == base {
				m.setSelected(i)
				break
			}
		}
		return extRunDoneMsg{err: nil}
	}
}

func trimToWidth(s string, w int) string {
	if w <= 0 {
		return s
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	if w <= 1 {
		return string([]rune(s)[:w])
	}
	// rune-safe truncation with ellipsis
	r := []rune(s)
	if len(r) <= w-1 {
		return s
	}
	return string(r[:w-1]) + "…"
}

// preview helpers moved to render_preview.go and preview/iterm2.go

// computeMaxDirName returns the maximum width for directory names (with trailing slash)
// used for column width hints.
func computeMaxDirName(entries []panels.Entry) int {
	maxLen := 0
	for _, e := range entries {
		if e.IsDir {
			l := lipgloss.Width(e.Name) + 1
			if l > maxLen {
				maxLen = l
			}
		}
	}
	if maxLen == 0 {
		maxLen = 10
	}
	return maxLen
}

// maybePrefetchSelected schedules an async prefetch of the currently selected directory
// (in the focused panel/column). It returns a tea.Cmd to run the prefetch, or nil.
func (m *model) maybePrefetchSelected() tea.Cmd {
	t := m.focused()
	if t == nil || t.panel == nil {
		return nil
	}
	p := t.panel
	if len(p.Entries) == 0 {
		return nil
	}
	if t.selected < 0 || t.selected >= len(p.Entries) {
		return nil
	}
	e := p.Entries[t.selected]
	if !e.IsDir {
		return nil
	}
	path := filepath.Join(p.Cwd, e.Name)
	if m.dirCache.Has(path) {
		return nil
	}
	if _, inflight := m.prefetching[path]; inflight {
		return nil
	}
	// Mark as in-flight to avoid duplicates
	m.prefetching[path] = struct{}{}
	showHidden := p.ShowHidden
	return func() tea.Msg {
		dp := panels.NewPanel(path, showHidden)
		_ = dp.Refresh()
		return dirPrefetchMsg{path: path, entries: dp.Entries}
	}
}

// tryRefreshMulti renders multi-column layout when right columns or preview are active.
// Returns true if it handled rendering and set the viewport content.
func (m *model) tryRefreshMulti() bool {
	m.prof.Begin("refresh-multi")
	if len(m.tabs) == 0 {
		m.vp.SetContent("")
		m.prof.End("refresh-multi")
		return true
	}
	leftTab := m.tabs[m.active]
	totalW := m.vp.Width
	if totalW <= 0 {
		totalW = m.width
	}
	if totalW <= 0 {
		totalW = 80
	}

	// Collect columns (panel mode)
	cols := []tab{leftTab}
	if m.rightMode == "panel" {
		cols = append(cols, m.rightCols...)
		if len(m.rightCols) == 0 && m.rightT.panel != nil {
			cols = append(cols, m.rightT)
		}
	}

	if len(cols) == 1 {
		if !m.showPrev {
			// Single left column only
			leftW := totalW
			if leftW < 10 {
				leftW = 10
			}
			leftLines := renderPanelColumn(m, cols[0], leftW, true)
			m.vp.SetContent(join(leftLines, "\n"))
			m.prof.End("refresh-multi")
			return true
		}
		// Render left | preview
		rightW := (totalW * m.rightPct) / 100
		if rightW < 10 {
			rightW = 10
		}
		if rightW > totalW-10 {
			rightW = totalW - 10
		}
		leftW := totalW - rightW - 1
		if leftW < 10 {
			leftW = 10
		}
		// In preview-only mode there is no interactive right column,
		// so the left column should always be considered focused for styling.
		leftLines := renderPanelColumn(m, cols[0], leftW, true)
		var rightLines []string
		p := cols[0].panel
		if len(p.Entries) > 0 {
			e := p.Entries[cols[0].selected]
			path := filepath.Join(p.Cwd, e.Name)
			header := trimToWidth(path, rightW)
			rightLines = append(rightLines, m.styStatus.Render(header))
			if e.IsDir {
				// Use cached directory entries for preview
				entries := m.dirCache.Get(path)
				if entries == nil {
					dp := panels.NewPanel(path, m.deps.Config.ShowHidden)
					_ = dp.Refresh()
					entries = dp.Entries
					m.dirCache.Put(path, entries)
				}
				for _, de := range entries {
					name := de.Name
					if de.IsDir {
						name += "/"
					}
					ln := trimToWidth(name, rightW)
					pad := rightW - lipgloss.Width(ln)
					if pad > 0 {
						ln += strings.Repeat(" ", pad)
					}
					if de.IsDir {
						rightLines = append(rightLines, m.styDir.Render(ln))
					} else {
						rightLines = append(rightLines, m.styNormal.Render(ln))
					}
				}
			} else {
				// File preview: try inline image if supported; else text fallback
				body := m.renderFilePreviewBody(path, rightW, m.viewportHeight()-1)
				rightLines = append(rightLines, body...)
			}
		}
		m.vp.SetContent(mergeColumns([][]string{leftLines, rightLines}, []int{leftW, rightW}, " "))
		m.prof.End("refresh-multi")
		return true
	}

	// Multi-column panel mode
	if m.showPrev {
		// Allocate rightmost preview column width and distribute the rest among panel columns
		prevW := (totalW * m.rightPct) / 100
		if prevW < 10 {
			prevW = 10
		}
		if prevW > totalW-10 {
			prevW = totalW - 10
		}
		avail := totalW - prevW - 1
		if avail < 10 {
			avail = 10
		}
		widths := autoColumnWidths(cols, avail, 1)
		linesPerCol := make([][]string, 0, len(cols))
		for i, c := range cols {
			focusCol := (i == 0 && m.focus == "left") || (i == len(cols)-1 && m.focus == "right")
			linesPerCol = append(linesPerCol, renderPanelColumn(m, c, widths[i], focusCol))
		}
		// Build preview lines for the currently focused selection
		var prevLines []string
		ft := m.focused()
		if ft != nil && ft.panel != nil {
			p := ft.panel
			if len(p.Entries) > 0 && ft.selected >= 0 && ft.selected < len(p.Entries) {
				e := p.Entries[ft.selected]
				path := filepath.Join(p.Cwd, e.Name)
				header := trimToWidth(path, prevW)
				prevLines = append(prevLines, m.styStatus.Render(header))
				if e.IsDir {
					entries := m.dirCache.Get(path)
					if entries == nil {
						dp := panels.NewPanel(path, p.ShowHidden)
						_ = dp.Refresh()
						entries = dp.Entries
						m.dirCache.Put(path, entries)
					}
					for _, de := range entries {
						name := de.Name
						if de.IsDir {
							name += "/"
						}
						ln := trimToWidth(name, prevW)
						pad := prevW - lipgloss.Width(ln)
						if pad > 0 {
							ln += strings.Repeat(" ", pad)
						}
						if de.IsDir {
							prevLines = append(prevLines, m.styDir.Render(ln))
						} else {
							prevLines = append(prevLines, m.styNormal.Render(ln))
						}
					}
				} else {
					// File preview: try inline image if supported; else text fallback
					body := m.renderFilePreviewBody(path, prevW, m.viewportHeight()-1)
					prevLines = append(prevLines, body...)
				}
			}
		}
		m.vp.SetContent(mergeColumns(append(linesPerCol, prevLines), append(widths, prevW), " "))
		m.prof.End("refresh-multi")
		return true
	}
	widths := autoColumnWidths(cols, totalW, 1)
	linesPerCol := make([][]string, 0, len(cols))
	for i, c := range cols {
		focusCol := (i == 0 && m.focus == "left") || (i == len(cols)-1 && m.focus == "right")
		linesPerCol = append(linesPerCol, renderPanelColumn(m, c, widths[i], focusCol))
	}
	m.vp.SetContent(mergeColumns(linesPerCol, widths, " "))
	m.prof.End("refresh-multi")
	return true
}

// autoColumnWidths computes column widths based on the longest directory name in each column.
// moved to layout_utils.go and render_columns.go

// Start runs the TUI event loop using Bubble Tea.
func Start(ctx context.Context, deps Dependencies) error {
	deps.Logger.Infof("starting TUI (Bubble Tea)")
	m, err := initialModel(deps)
	if err != nil {
		return err
	}
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)
	_, err = p.Run()
	return err
}

// refreshContent rebuilds the viewport content from the current panel state.
func (m *model) refreshContent() {
	m.prof.Begin("refresh")
	// Modal overlay content (help/command output)
	if m.modalActive {
		totalW := m.vp.Width
		if totalW <= 0 {
			totalW = m.width
		}
		if totalW <= 0 {
			totalW = 80
		}
		lines := make([]string, 0, len(m.modalLines)+1)
		title := m.modalTitle
		if title != "" {
			lines = append(lines, m.styStatus.Render(trimToWidth(title, totalW)))
		}
		for _, l := range m.modalLines {
			ln := trimToWidth(l, totalW)
			pad := totalW - lipgloss.Width(ln)
			if pad > 0 {
				ln += strings.Repeat(" ", pad)
			}
			lines = append(lines, m.styNormal.Render(ln))
		}
		if len(lines) == 0 {
			lines = []string{""}
		}
		m.vp.SetContent(join(lines, "\n"))
		m.prof.End("refresh")
		return
	}
	// Single path: unified renderer handles all layouts
	_ = m.tryRefreshMulti()
	m.prof.End("refresh")
}

// computeStyles moved to styles.go
