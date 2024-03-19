package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	_ "go.uber.org/automaxprocs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := LoadConfig(os.Getenv("CONFIG_FILENAME"))
	if err != nil {
		slog.Error("failed load config", slog.Any("err", err))
		os.Exit(1)
	}
	slog.Info("loaded config", slog.Any("config", cfg))

	if err = StartDrone(ctx, cfg); err != nil {
		slog.Error("failed to start drone", slog.Any("err", err))
		os.Exit(1)
	}
}
