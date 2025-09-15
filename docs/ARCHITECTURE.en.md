Project Architecture (draft)

The project is organized as a Go module with the `tfm` executable and
internal packages that encapsulate implementation details.

## Current Directory Layout

- `cmd/tfm/` — CLI entry point
- `internal/app/` — dependency wiring and startup
- `internal/logging/` — lightweight logger with levels
- `internal/config/` — configuration (TOML), loading/defaults
- `internal/keymap/` — key maps, default Vim-like scheme
- `internal/theme/` — theme model (colors/styles)
- `internal/ui/commands/` — command registry and execution
- `internal/fs/ops/` — file operations (queue/workers in future)
- `internal/ui/panels/` — panels and directory listings
- `internal/ui/preview/` — preview providers
- `internal/ui/tui/` — TUI shell (Bubble Tea)
- `docs/` — specification and architecture docs

## Next Steps
1. Integrate Bubble Tea/Bubbles/Lip Gloss and implement the base TUI model.
2. Implement TOML config loading and hot reload.
3. Add a task queue for filesystem operations with progress.
4. Expand the command palette and key bindings.

