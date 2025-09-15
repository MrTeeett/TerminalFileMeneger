package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ColorStyle describes basic style properties for a text region.
type ColorStyle struct {
	FG      string
	BG      string
	Bold    bool
	Faint   bool
	Reverse bool
}

// ThemeConfig groups styles for key UI areas.
type ThemeConfig struct {
	Header   ColorStyle
	Status   ColorStyle
	Dir      ColorStyle
	Selected ColorStyle
	Normal   ColorStyle
}

// Config holds user preferences loaded from TOML-like config.
type Config struct {
	ShowHidden bool
	ThemeName  string
	Keymap     string
	// Keys holds user keybindings (section [keys] or [keys.normal])
	// mapping key sequence specs (e.g., "j", "g g", "ctrl+d") to actions.
	Keys map[string]string
	// UI options
	ShowPreview       bool    // show preview pane on the right
	OpenDirsRight     bool    // when entering a dir, open it in right pane instead of replacing left
	RightPaneWidth    int     // right pane width percent (10..80)
	InlineImages      bool    // enable inline image previews (iTerm2/WezTerm/Kitty etc.)
	ColorProfile      string  // color profile: auto|none|ansi|256|truecolor
	BackgroundOpacity float64 // 0..1 hint: if <1, avoid BG fills to let terminal transparency show
	Blur              bool    // hint flag (actual blur depends on terminal/compositor)
	// CustomCommands maps command names to shell snippets.
	// Example:
	//   [commands]
	//   open = "xdg-open {path}"
	// Placeholders: {path}, {cwd}, {file}
	CustomCommands map[string]string
	Theme          ThemeConfig
}

func Default() *Config {
	return &Config{
		ShowHidden:        false,
		ThemeName:         "default",
		Keymap:            "vim",
		Keys:              map[string]string{},
		ShowPreview:       true,
		OpenDirsRight:     false,
		RightPaneWidth:    40,
		InlineImages:      true,
		ColorProfile:      "auto",
		BackgroundOpacity: 1.0,
		Blur:              false,
		CustomCommands:    map[string]string{},
		Theme: ThemeConfig{
			Header:   ColorStyle{Bold: true},
			Status:   ColorStyle{Faint: true},
			Dir:      ColorStyle{FG: "12"}, // ANSI blue
			Selected: ColorStyle{Bold: true, Reverse: true},
			Normal:   ColorStyle{},
		},
	}
}

// Load attempts to load config from path. Falls back to defaults on error.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Default(), errors.New("config not found, using defaults")
	}
	cfg, err := Parse(string(b))
	if err != nil {
		return Default(), fmt.Errorf("config parse error: %w", err)
	}
	return cfg, nil
}

// DefaultPath returns XDG-compliant default config path.
func DefaultPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tfm", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tfm", "config.toml")
}

// Parse reads a very small subset of TOML sufficient for our theme config.
// Supported constructs:
//   - Comments starting with '#'
//   - Sections: [theme], [theme.header], [theme.status], [theme.dir], [theme.selected], [theme.normal]
//   - Keys with values: key = "value" | true | false
//   - Root keys: show_hidden, theme_name, keymap
func Parse(s string) (*Config, error) {
	cfg := Default()
	sec := ""
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sec = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		k, v, ok := splitKV(line)
		if !ok {
			continue
		}
		k = strings.ToLower(k)
		switch sec {
		case "": // root
			switch k {
			case "show_hidden":
				if b, err := parseBool(v); err == nil {
					cfg.ShowHidden = b
				}
			case "theme_name", "theme":
				cfg.ThemeName = trimQuotes(v)
			case "keymap":
				cfg.Keymap = trimQuotes(v)
			case "show_preview":
				if b, err := parseBool(v); err == nil {
					cfg.ShowPreview = b
				}
			case "open_dirs_right", "open_on_right":
				if b, err := parseBool(v); err == nil {
					cfg.OpenDirsRight = b
				}
			case "right_pane_width", "right_width":
				if n, err := strconv.Atoi(trimQuotes(v)); err == nil {
					if n < 10 {
						n = 10
					} else if n > 80 {
						n = 80
					}
					cfg.RightPaneWidth = n
				}
			case "inline_images", "inline":
				if b, err := parseBool(v); err == nil {
					cfg.InlineImages = b
				}
			case "color_profile", "colors", "color":
				cfg.ColorProfile = strings.ToLower(trimQuotes(v))
			case "background_opacity", "opacity":
				if f, err := strconv.ParseFloat(trimQuotes(v), 64); err == nil {
					if f > 1.0 && f <= 100.0 {
						f = f / 100.0
					}
					if f < 0 {
						f = 0
					}
					if f > 1 {
						f = 1
					}
					cfg.BackgroundOpacity = f
				}
			case "blur":
				if b, err := parseBool(v); err == nil {
					cfg.Blur = b
				}
			}
		case "commands", "cmd", "ex":
			if cfg.CustomCommands == nil {
				cfg.CustomCommands = make(map[string]string)
			}
			cfg.CustomCommands[strings.TrimSpace(k)] = trimQuotes(v)
		case "keys", "keymap.keys", "keys.normal":
			if cfg.Keys == nil {
				cfg.Keys = make(map[string]string)
			}
			cfg.Keys[strings.TrimSpace(k)] = trimQuotes(v)
		case "preview":
			switch k {
			case "enabled":
				if b, err := parseBool(v); err == nil {
					cfg.ShowPreview = b
				}
			case "width", "right_pane_width":
				if n, err := strconv.Atoi(trimQuotes(v)); err == nil {
					if n < 10 {
						n = 10
					} else if n > 80 {
						n = 80
					}
					cfg.RightPaneWidth = n
				}
			case "inline_images", "inline":
				if b, err := parseBool(v); err == nil {
					cfg.InlineImages = b
				}
			case "color_profile", "colors", "color":
				cfg.ColorProfile = strings.ToLower(trimQuotes(v))
			case "background_opacity", "opacity":
				if f, err := strconv.ParseFloat(trimQuotes(v), 64); err == nil {
					if f > 1.0 && f <= 100.0 {
						f = f / 100.0
					}
					if f < 0 {
						f = 0
					}
					if f > 1 {
						f = 1
					}
					cfg.BackgroundOpacity = f
				}
			case "blur":
				if b, err := parseBool(v); err == nil {
					cfg.Blur = b
				}
			}
		case "view":
			switch k {
			case "open_dirs_right", "open_on_right":
				if b, err := parseBool(v); err == nil {
					cfg.OpenDirsRight = b
				}
			case "right_pane_width":
				if n, err := strconv.Atoi(trimQuotes(v)); err == nil {
					if n < 10 {
						n = 10
					} else if n > 80 {
						n = 80
					}
					cfg.RightPaneWidth = n
				}
			case "background_opacity", "opacity":
				if f, err := strconv.ParseFloat(trimQuotes(v), 64); err == nil {
					if f > 1.0 && f <= 100.0 {
						f = f / 100.0
					}
					if f < 0 {
						f = 0
					}
					if f > 1 {
						f = 1
					}
					cfg.BackgroundOpacity = f
				}
			case "blur":
				if b, err := parseBool(v); err == nil {
					cfg.Blur = b
				}
			}
		case "theme", "theme.header", "theme.status", "theme.dir", "theme.selected", "theme.normal":
			applyStyleKey(&cfg.Theme, sec, k, v)
		default:
			// ignore unknown sections
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func splitKV(line string) (key, val string, ok bool) {
	i := strings.Index(line, "=")
	if i < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:i])
	val = strings.TrimSpace(line[i+1:])
	return key, val, true
}

func trimQuotes(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}

func parseBool(v string) (bool, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		// try to parse numeric
		n, err := strconv.Atoi(v)
		if err == nil {
			return n != 0, nil
		}
	}
	return false, fmt.Errorf("invalid bool: %q", v)
}

func applyStyleKey(th *ThemeConfig, section, key, val string) {
	var target *ColorStyle
	switch section {
	case "theme", "theme.normal":
		target = &th.Normal
	case "theme.header":
		target = &th.Header
	case "theme.status":
		target = &th.Status
	case "theme.dir":
		target = &th.Dir
	case "theme.selected":
		target = &th.Selected
	}
	if target == nil {
		return
	}
	switch key {
	case "fg", "foreground":
		target.FG = normalizeColorString(trimQuotes(val))
	case "bg", "background":
		target.BG = normalizeColorString(trimQuotes(val))
	case "bold":
		if b, err := parseBool(val); err == nil {
			target.Bold = b
		}
	case "faint":
		if b, err := parseBool(val); err == nil {
			target.Faint = b
		}
	case "reverse":
		if b, err := parseBool(val); err == nil {
			target.Reverse = b
		}
	}
}

// normalizeColorString accepts common color formats and strips unsupported alpha.
// Supported examples: "12" (ANSI), "red" (name), "#RGB", "#RRGGBB", "#RRGGBBAA" (alpha ignored).
func normalizeColorString(s string) string {
	if s == "" {
		return s
	}
	// Only handle hex-like strings
	if len(s) >= 2 && s[0] == '#' {
		// #RRGGBBAA -> #RRGGBB (drop alpha, terminals don't support it)
		if len(s) == 9 { // '#' + 8 hex digits
			return s[:7]
		}
		// accept #RGB/#RRGGBB as-is
	}
	return s
}
