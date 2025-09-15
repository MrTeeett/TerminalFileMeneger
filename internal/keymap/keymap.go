package keymap

// Action describes a semantic action in the app.
type Action string

// Binding maps keys or chords (e.g., "gg", "Ctrl+C") to actions.
type Binding map[string]Action

// Map groups bindings by mode (normal/insert/command). For MVP — only normal.
type Map struct {
	Normal Binding
}

// TokensOf splits a binding specification into a sequence of key tokens.
// Supported forms:
//   - Space-separated sequences: "g g", "ctrl+x ctrl+e"
//   - Single special key with modifiers: "ctrl+c", "enter", "pgdown"
//   - Repeated runes without spaces: "gg", "yy"
//
// Returns a slice of tokens that should be matched one-by-one against incoming keys
// as produced by Bubble Tea's tea.KeyMsg.String().
func TokensOf(spec string) []string {
	s := spec
	// Normalize internal spaces
	// If the spec contains spaces, treat as explicit sequence
	hasSpace := false
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			hasSpace = true
			break
		}
	}
	if hasSpace {
		out := make([]string, 0, 4)
		w := ""
		for _, r := range s {
			if r == ' ' || r == '\t' || r == '\n' {
				if w != "" {
					out = append(out, w)
					w = ""
				}
				continue
			}
			w += string(r)
		}
		if w != "" {
			out = append(out, w)
		}
		return out
	}
	// No spaces: handle named special keys as a single token
	switch s {
	case "left", "right", "up", "down", "enter", "backspace", "pgdown", "pgup", "home", "end", "tab":
		return []string{s}
	}
	// If it's a plain alpha string like "gg" split to runes
	plainAlpha := true
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') {
			plainAlpha = false
			break
		}
	}
	if plainAlpha && len(s) > 1 {
		out := make([]string, 0, len([]rune(s)))
		for _, r := range s {
			out = append(out, string(r))
		}
		return out
	}
	// Otherwise, treat entire spec as a single token
	return []string{s}
}

// Default returns a Vim-like keymap.
func Default() Map {
	return Map{
		Normal: Binding{
			// Navigation (Vim)
			"h":  "left",
			"j":  "down",
			"k":  "up",
			"l":  "right",
			"gg": "top",
			"G":  "bottom",
			".":  "toggle-hidden",
			// Navigation (arrows, paging)
			"left":      "left",
			"right":     "right",
			"up":        "up",
			"down":      "down",
			"enter":     "right",
			"backspace": "left",
			"pgdown":    "page-down",
			"pgup":      "page-up",
			"ctrl+d":    "half-page-down",
			"ctrl+u":    "half-page-up",
			// Tabs
			"t": "new-tab",
			"]": "next-tab",
			"[": "prev-tab",
			"w": "close-tab",
			// View toggles
			"ctrl+p": "toggle-preview",
			"ctrl+o": "toggle-right-open-mode",
			"ctrl+x": "close-right",
			// Command-line (Ex) mode
			":": "command",
			// Focus
			"tab": "toggle-focus",
			// Quit
			"q":      "quit",
			"ctrl+c": "quit",
			// File ops (not yet implemented in TUI; placeholders)
			"r":  "rename",
			"d":  "delete",
			"yy": "copy",
			"pp": "paste",
			"Y":  "copy-path",
			"P":  "paste-path",
			"/":  "filter",
			"f":  "fuzzy",
		},
	}
}
