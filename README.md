TFM — Terminal File Manager (Go)

English • Русская версия: README.ru.md

TFM is a fast, configurable terminal file manager built on Bubble Tea. It
focuses on convenient navigation, preview, flexible key/theme setup, and easy
extensibility via commands.

- Stack: Bubble Tea, Bubbles, Lip Gloss
- Platforms: Linux (x86_64/arm64)
- Terminals: kitty, wezterm, iTerm2, alacritty, konsole, tmux (see notes)

See also:
- docs/SPEC.en.md — specification and vision
- docs/ARCHITECTURE.en.md — architecture and module layout

## Features
- Panels/Tabs: left panel, right panel, preview, tabs
- Navigation: Vim keys (h/j/k/l, gg/G), arrows, PgUp/PgDn, Ctrl+U/D
- Preview: text/metadata, inline images (iTerm2/WezTerm), ASCII fallback
- Command mode (:): `:help`, `:cd`, `:preview on|off|toggle`, `:theme`
- Copy/Paste: files/dirs and paths (yy/pp, Y/P and corresponding :copy…)
- Themes and colors: configurable styles, color profiles, transparency hints

## Build & Run
Requires Go 1.21+.

```
# Build
make build
# or
go build -o bin/tfm ./cmd/tfm

# Run
./bin/tfm
```

## Configuration
Config file (TOML): `~/.config/tfm/config.toml`.
Example: `configs/config.example.toml`.

Key options (root):
- `show_hidden` — show dotfiles
- `open_dirs_right` — open directories on the right
- `right_pane_width` — width of right panel/preview (percent)
- `color_profile` — `auto|none|ansi|256|truecolor` (recommend `truecolor`)
- `inline_images` — enable image preview (iTerm2/WezTerm)
- `background_opacity` — background transparency (`0..1` or `0..100`%),
  when `< 1` TFM avoids BG fills so terminal transparency shows through
- `blur` — hint flag (actual blur depends on terminal/compositor)

### Key bindings ([keys])
Any action can be remapped:

```
[keys]
"j" = "down"
"k" = "up"
"yy" = "copy"      # copy file/dir
"pp" = "paste"     # paste into current directory
"Y"  = "copy-path" # copy absolute path
"P"  = "paste-path"# jump to copied path
```

See `configs/config.example.toml` for the full action list.

## Command mode (:)
- `:help` — help
- `:cd <path>` — change directory (`~` and relative paths supported)
- `:preview on|off|toggle` — control preview
- `:copy`, `:paste`, `:copy-path`, `:paste-path`
- `:theme` — theme/color diagnostics (profile, TERM/COLORTERM, samples)
- `:opacity <0..1|0..100>` — apply transparency on the fly
- `:blur on|off` — hint toggle (blur is enabled in terminal/compositor)

## Image preview
- Inline images for iTerm2/WezTerm (OSC 1337)
- Kitty/ASCII — automatic fallback (text/ASCII) for now
- Disable via `inline_images = false` or `TFM_NO_INLINE_IMAGES=1`

## Colors & Transparency
- Color profile is auto-detected (or set by `color_profile`)
- For kitty/wezterm TrueColor is recommended
- For transparency set `background_opacity < 1.0` and enable transparency in
  your terminal; blur requires compositor/terminal support

## Debugging (VS Code)
- Use external terminal and legacy adapter for stability:
  - `console: externalTerminal`
  - `debugAdapter: legacy`
- Configure the external terminal via VS Code `terminal.external.*` settings.

## Architecture
- Application: `internal/app`
- Config/Theme/Keymap: `internal/config`, `internal/theme`, `internal/keymap`
- FS operations: `internal/fs/ops` (Copy/Move/Delete, recursive copy)
- UI (TUI): `internal/ui/tui` (model, layout rendering, command mode)
- Preview: `internal/ui/preview` (providers)

More in `docs/ARCHITECTURE.en.md`.

## License
See LICENSE (MIT).
