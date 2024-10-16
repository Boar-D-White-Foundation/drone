package cli

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/tg"
)

func Run(name string, cmd func(context.Context, config.Config, *alert.Manager) error) {
	flag.Parse()

	cfg, err := config.Load(config.Path())
	if err != nil {
		slog.Error("failed to load config", slog.Any("err", err))
		os.Exit(1)
	}
	slog.Info("loaded config", slog.Any("config", cfg))

	adminTGClient, err := tg.NewAdminClientFromConfig(cfg)
	if err != nil {
		slog.Error("failed to create admin tg client", slog.Any("err", err))
		os.Exit(1)
	}

	alerts := alert.NewManager(adminTGClient)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd(ctx, cfg, alerts); err != nil {
		alerts.Errorxf(err, "failed to run cmd: %s", name)
		stop()
		os.Exit(1)
	}
}
