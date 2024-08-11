package main

import (
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/cli"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/db"
)

var dumpPath = flag.String("p", "db_dump.json.gz", "path to the dump file")

func restoreDB(ctx context.Context, cfg config.Config, alerts *alert.Manager) error {
	database := db.NewBadgerDBFromConfig(cfg)
	if err := database.Start(ctx); err != nil {
		return fmt.Errorf("failed to start database: %w", err)
	}
	defer database.Stop()

	fd, err := os.Open(*dumpPath)
	if err != nil {
		return fmt.Errorf("failed to open dump file: %w", err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			slog.Error("failed to close dump file", slog.Any("err", err))
		}
	}()

	reader, err := gzip.NewReader(fd)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			slog.Error("failed to close gzip reader", slog.Any("err", err))
		}
	}()

	if err := db.RestoreJson(ctx, database, reader); err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	slog.Info("restore OK", slog.String("path", *dumpPath))
	return nil
}

func main() {
	cli.Run("db-restore", restoreDB)
}
