package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MrTeeett/TerminalFileMeneger/internal/config"
	"github.com/MrTeeett/TerminalFileMeneger/internal/fs/ops"
	"github.com/MrTeeett/TerminalFileMeneger/internal/keymap"
	"github.com/MrTeeett/TerminalFileMeneger/internal/logging"
	"github.com/MrTeeett/TerminalFileMeneger/internal/theme"
	"github.com/MrTeeett/TerminalFileMeneger/internal/ui/commands"
	"github.com/MrTeeett/TerminalFileMeneger/internal/ui/tui"
)

// Options holds startup options from CLI.
type Options struct {
	ConfigPath string
	WorkingDir string
	LogLevel   string
	Version    string
}

// Run wires dependencies and starts the TUI.
func Run(ctx context.Context, opts Options) error {
	// Write logs to file to avoid polluting terminal output before TUI alt-screen.
	var logFile *os.File
	if cacheDir, err := os.UserCacheDir(); err == nil {
		dir := filepath.Join(cacheDir, "tfm", "logs")
		_ = os.MkdirAll(dir, 0o755)
		name := filepath.Join(dir, fmt.Sprintf("tfm-%s.log", time.Now().Format("20060102-150405")))
		if f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
			logFile = f
			defer logFile.Close()
		}
	}
	var logger logging.Logger
	if logFile != nil {
		logger = logging.NewWithWriter(opts.LogLevel, logFile)
	} else {
		logger = logging.New(opts.LogLevel)
	}
	logger.Infof("tfm %s starting", opts.Version)

	wd := opts.WorkingDir
	if wd == "" {
		if cwd, err := os.Getwd(); err == nil {
			wd = cwd
		}
	}
	if wd != "" {
		abs, err := filepath.Abs(wd)
		if err == nil {
			if err := os.Chdir(abs); err != nil {
				logger.Warnf("cannot chdir to %s: %v", abs, err)
			}
		}
	}

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		logger.Warnf("using default config: %v", err)
		cfg = config.Default()
	}

	km := keymap.Default()
	th := theme.Default()
	reg := commands.NewRegistry()
	fsman := ops.NewManager()

	// Apply key overrides from config ([keys] section)
	if cfg != nil && len(cfg.Keys) > 0 {
		if km.Normal == nil {
			km.Normal = make(keymap.Binding)
		}
		for spec, act := range cfg.Keys {
			km.Normal[spec] = keymap.Action(act)
		}
	}

	deps := tui.Dependencies{
		Logger:   logger,
		Config:   cfg,
		Keymap:   km,
		Theme:    th,
		Registry: reg,
		FS:       fsman,
	}

	if err := tui.Start(ctx, deps); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	logger.Infof("tfm exited")
	return nil
}
