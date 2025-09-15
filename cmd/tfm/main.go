package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/MrTeeett/TerminalFileMeneger/internal/app"
)

var version = "0.1.0"

func main() {
	var configPath string
	var logLevel string
	var workingDir string
	var showVersion bool

	flag.StringVar(&configPath, "config", "", "Path to config file (TOML)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level: debug|info|warn|error")
	flag.StringVar(&workingDir, "working-dir", "", "Working directory to start in")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println("tfm", version)
		return
	}

	opts := app.Options{
		ConfigPath: configPath,
		LogLevel:   logLevel,
		WorkingDir: workingDir,
		Version:    version,
	}

	ctx := context.Background()
	if err := app.Run(ctx, opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
