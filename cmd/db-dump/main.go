package main

import (
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/cli"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/db"
)

var dumpPath = flag.String("p", "db_dump.json.gz", "path to the dump file")

func dumpDB(ctx context.Context, cfg config.Config, alerts *alert.Manager) error {
	database := db.NewBadgerDBFromConfig(cfg)
	if err := database.Start(ctx); err != nil {
		return fmt.Errorf("failed to start database: %w", err)
	}
	defer database.Stop()

	fd, err := os.OpenFile(*dumpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open dump file: %w", err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			slog.Error("failed to close dump file", slog.Any("err", err))
		}
	}()

	writer, err := gzip.NewWriterLevel(fd, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	writer.Name = *dumpPath
	writer.ModTime = time.Now()
	defer func() {
		if err := writer.Close(); err != nil {
			slog.Error("failed to close gzip writer", slog.Any("err", err))
		}
	}()

	if err := db.DumpJson(ctx, database, writer); err != nil {
		return fmt.Errorf("failed to dump database: %w", err)
	}

	slog.Info("backup OK", slog.String("path", *dumpPath))
	return nil
}

func main() {
	cli.Run("db-dump", dumpDB)
}
