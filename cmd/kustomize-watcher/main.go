package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/polygone-app/kustomize-watcher/internal/applier"
	"github.com/polygone-app/kustomize-watcher/internal/config"
	"github.com/polygone-app/kustomize-watcher/internal/watcher"
)

var (
	cfgPath  string
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "kustomize-watcher",
	Short: "Watch kustomize directories and apply them to the local cluster",
	RunE:  run,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "path to config file (env: KUSTOMIZE_WATCHER_CONFIG)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level: debug/info/warn/error (env: KUSTOMIZE_WATCHER_LOG_LEVEL)")
}

func run(cmd *cobra.Command, _ []string) error {
	if cfgPath == "" {
		cfgPath = os.Getenv("KUSTOMIZE_WATCHER_CONFIG")
	}
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	// Bootstrap logger for config-load errors before we know the real level.
	bootstrap := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cfg, err := config.Load(cfgPath)
	if err != nil {
		bootstrap.Error("load config", "path", cfgPath, "err", err)
		return err
	}

	// Resolve effective log level: flag > env > config > default.
	effective := cfg.LogLevel
	if ev := os.Getenv("KUSTOMIZE_WATCHER_LOG_LEVEL"); ev != "" {
		effective = ev
	}
	if cmd.Flags().Changed("log-level") {
		effective = logLevel
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LevelFromString(effective),
	}))
	slog.SetDefault(logger)

	a, err := applier.New(logger)
	if err != nil {
		return fmt.Errorf("applier: %w", err)
	}

	w, err := watcher.New(cfg.Glob, a, logger)
	if err != nil {
		return fmt.Errorf("watcher: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return w.Run(ctx)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
