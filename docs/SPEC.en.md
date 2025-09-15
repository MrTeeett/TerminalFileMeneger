Terminal File Manager (Go) — Specification (draft)

Document version: 0.1 (draft). Intended to align vision, priorities,
and architecture before and during implementation.

## Purpose
A fast, convenient, and extensible terminal file manager (TUI) for Linux,
usable locally and over SSH. Focus on deep customization via config, key
bindings, themes, macros, and external commands/plugins.

## Defaults
- TUI stack: Bubble Tea + Bubbles + Lip Gloss
- Config format: TOML (`~/.config/tfm/config.toml`)
- Binding style: Vim-like (fully remappable)
- Binary/project name: `tfm` (Terminal File Manager)

## Goals
- Performance: responsive UI on large directories; background operations;
  minimal input latency.
- Usability: predictable actions; command palette; clear status; one-hand nav.
- Customization: themes, bindings, macros, external commands; hot reload.
- Reliability: robust to FS errors; correct behavior with permissions,
  symlinks, and special files.

## Non-functional Requirements
- Platforms: Linux x86_64/arm64; popular terminals (xterm, kitty, wezterm, alacritty)
- Distributed as a single statically-linked binary (no CGO by default)
- Clear error messages; no panics in normal operation
- Logs: `~/.cache/tfm/logs/` with daily rotation and level from CLI flag

## MVP Features
- Panels and Tabs
  - Single and two-panel layout (left/right)
  - At least 2 tabs with independent working directories
- Navigation
  - Arrows and Vim keys (h/j/k/l), Home/End, PageUp/PageDown
  - History nav (back/forward), quick `cd` to parent
  - Toggle show/hide dotfiles
  - Bookmarks and quick jumps
- File Ops
  - Copy, move, delete (with confirm/"trash"), rename
  - Create: `mkdir`, `touch`
  - Bulk operations via selection
- Task Queue
  - Background operations with progress, cancelation, and parallelism limits
- Search & Filter
  - Incremental filter in current listing
  - Fuzzy search by names (time/limit constrained)
- Preview
  - Text/hex/metadata; for binaries — short info
  - Image preview where supported (e.g. iTerm2/WezTerm), ASCII fallback

## Roadmap (initial)
1) Initialize `go mod` and base layout (`cmd/tfm`, `internal/...`).
2) Integrate Bubble Tea/Bubbles/Lip Gloss; render “Hello TUI”.
3) Draft config and keymap, load from TOML.
4) Panels skeleton and minimal navigation.

This document is a living draft. Please propose adjustments: terminology,
priorities, UX, stack choices.

