package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/db"
)

var dumpPath = flag.String("p", "db_dump.json", "path to the dump file")

func main() {
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.Load(config.Path())
	if err != nil {
		slog.Error("failed load config", slog.Any("err", err))
		os.Exit(1)
	}

	database := db.NewBadgerDBFromConfig(cfg)
	if err := database.Start(ctx); err != nil {
		slog.Error("failed to start database", slog.Any("err", err))
		os.Exit(1)
	}
	defer database.Stop()

	fd, err := os.OpenFile(*dumpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		slog.Error("failed to open dump file", slog.Any("err", err))
		os.Exit(1)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			slog.Error("failed to close dump file", slog.Any("err", err))
		}
	}()

	backup := db.JsonBackup{DB: database}
	if err := backup.Dump(ctx, fd); err != nil {
		slog.Error("failed to dump database", slog.Any("err", err))
		os.Exit(1)
	}

	slog.Info("backup OK", slog.String("path", *dumpPath))
}
