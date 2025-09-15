package config

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse("")
	if err != nil {
		t.Fatalf("Parse empty: %v", err)
	}
	want := Default()
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("defaults mismatch:\n got: %#v\nwant: %#v", cfg, want)
	}
}

func TestParseRootKeysAndClamps(t *testing.T) {
	src := `
show_hidden = true
theme = "dracula"
keymap = "emacs"
show_preview = false
open_on_right = true
right_pane_width = 100
inline_images = false
color_profile = "256"
background_opacity = 150
blur = true
`
	cfg, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !cfg.ShowHidden {
		t.Errorf("ShowHidden = %v; want true", cfg.ShowHidden)
	}
	if cfg.ThemeName != "dracula" {
		t.Errorf("ThemeName = %q; want dracula", cfg.ThemeName)
	}
	if cfg.Keymap != "emacs" {
		t.Errorf("Keymap = %q; want emacs", cfg.Keymap)
	}
	if cfg.ShowPreview {
		t.Errorf("ShowPreview = %v; want false", cfg.ShowPreview)
	}
	if !cfg.OpenDirsRight {
		t.Errorf("OpenDirsRight = %v; want true", cfg.OpenDirsRight)
	}
	if cfg.RightPaneWidth != 80 {
		t.Errorf("RightPaneWidth = %d; want 80 (clamped)", cfg.RightPaneWidth)
	}
	if cfg.InlineImages {
		t.Errorf("InlineImages = %v; want false", cfg.InlineImages)
	}
	if cfg.ColorProfile != "256" {
		t.Errorf("ColorProfile = %q; want 256", cfg.ColorProfile)
	}
	if cfg.BackgroundOpacity != 1.0 {
		t.Errorf("BackgroundOpacity = %v; want 1.0 (clamped)", cfg.BackgroundOpacity)
	}
	if !cfg.Blur {
		t.Errorf("Blur = %v; want true", cfg.Blur)
	}
}

func TestSectionsCommandsKeysPreviewViewTheme(t *testing.T) {
	src := `
[commands]
open = "xdg-open {path}"
[keys]
gg = "top"
ctrl+x = "close-right"
[preview]
enabled = true
width = 5
inline = true
color_profile = "ansi"
[view]
open_on_right = false
right_pane_width = 5
[theme.header]
fg = "#11223344"
bg = "red"
bold = true
`
	cfg, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.CustomCommands["open"] == "" {
		t.Fatalf("commands not parsed")
	}
	if got := cfg.Keys["gg"]; got != "top" {
		t.Fatalf("keys[gg] = %q; want top", got)
	}
	if got := cfg.Keys["ctrl+x"]; got != "close-right" {
		t.Fatalf("keys[ctrl+x] = %q; want close-right", got)
	}
	if !cfg.ShowPreview {
		t.Errorf("ShowPreview (preview.enabled) = false; want true")
	}
	if cfg.RightPaneWidth != 10 {
		t.Errorf("RightPaneWidth (preview.width clamp) = %d; want 10", cfg.RightPaneWidth)
	}
	if !cfg.InlineImages {
		t.Errorf("InlineImages (preview.inline) = false; want true")
	}
	if cfg.ColorProfile != "ansi" {
		t.Errorf("ColorProfile (preview.color_profile) = %q; want ansi", cfg.ColorProfile)
	}
	if cfg.OpenDirsRight { // set false in [view]
		t.Errorf("OpenDirsRight (view.open_on_right) = true; want false")
	}
	if cfg.RightPaneWidth != 10 { // 5 clamped to 10
		t.Errorf("RightPaneWidth (view.right_pane_width clamp) = %d; want 10", cfg.RightPaneWidth)
	}
	if got := cfg.Theme.Header.FG; got != "#112233" {
		t.Errorf("theme.header.fg = %q; want #112233 (alpha stripped)", got)
	}
	if got := cfg.Theme.Header.BG; got != "red" {
		t.Errorf("theme.header.bg = %q; want red", got)
	}
	if !cfg.Theme.Header.Bold {
		t.Errorf("theme.header.bold = false; want true")
	}
}

func TestSplitKVAndTrimQuotesAndParseBool(t *testing.T) {
	if k, v, ok := splitKV("a = b"); !ok || k != "a" || v != "b" {
		t.Fatalf("splitKV failed: %v %q %q", ok, k, v)
	}
	if _, _, ok := splitKV("noequal"); ok {
		t.Fatalf("splitKV accepted invalid line")
	}
	if got := trimQuotes("'abc'"); got != "abc" {
		t.Errorf("trimQuotes single: %q", got)
	}
	if got := trimQuotes("\"abc\""); got != "abc" {
		t.Errorf("trimQuotes double: %q", got)
	}
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"true", true}, {"1", true}, {"yes", true}, {"on", true},
		{"false", false}, {"0", false}, {"no", false}, {"off", false},
	} {
		got, err := parseBool(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("parseBool(%q) = %v, %v; want %v, nil", tc.in, got, err, tc.want)
		}
	}
}

func TestNormalizeColorString(t *testing.T) {
	if got := normalizeColorString("#AABBCCDD"); got != "#AABBCC" {
		t.Errorf("normalize #RRGGBBAA -> %q; want #AABBCC", got)
	}
	if got := normalizeColorString("#ABC"); got != "#ABC" {
		t.Errorf("normalize #ABC -> %q; want #ABC", got)
	}
	if got := normalizeColorString("red"); got != "red" {
		t.Errorf("normalize name -> %q; want red", got)
	}
}

func TestLoadMissingReturnsDefault(t *testing.T) {
	// Use a path that does not exist
	dir := t.TempDir()
	cfg, err := Load(filepath.Join(dir, "nope.toml"))
	if err == nil {
		t.Fatalf("Load expected error for missing file")
	}
	if cfg == nil || cfg.ThemeName == "" {
		t.Fatalf("Load should return defaults on error")
	}
}

// Ensure DefaultPath does not panic and returns a plausible path.
func TestDefaultPath(t *testing.T) {
	// Override env var to a temp dir for determinism
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	got := DefaultPath()
	if !strings.HasPrefix(got, dir) {
		t.Fatalf("DefaultPath should use XDG_CONFIG_HOME, got %q", got)
	}
}
