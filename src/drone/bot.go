package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/frosthamster/drone/src/boardwhite"
	"github.com/frosthamster/drone/src/tg"
	"github.com/go-co-op/gocron/v2"
)

func StartDrone(ctx context.Context, cfg Config) error {
	telegramClient, err := tg.NewClient(cfg.TgKey, cfg.BoarDWhiteChatID)
	if err != nil {
		return fmt.Errorf("new tg client: %w", err)
	}

	db, err := NewBadger(cfg.BadgerPath)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close db", err)
		}
	}()

	bw := boardwhite.NewService(
		cfg.BoarDWhiteLeetCodeThreadID,
		telegramClient,
		db,
	)

	dr := drone{
		boardwhite: bw,
	}

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return err
	}

	job, err := scheduler.NewJob(
		gocron.CronJob(cfg.LCDailyCron, false),
		gocron.NewTask(wrapErrors("publishLCDaily", dr.publishLCDaily), ctx),
	)
	if err != nil {
		return err
	}

	scheduler.Start()
	t, err := job.NextRun()
	if err != nil {
		return err
	}
	slog.Info(
		"started scheduler",
		slog.String("task", "publishLCDaily"),
		slog.String("cron", cfg.LCDailyCron),
		slog.String("next_run", t.String()),
	)

	<-ctx.Done()
	return scheduler.Shutdown()
}

func wrapErrors(name string, f func(context.Context) error) func(context.Context) {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic in task", slog.String("name", name), slog.Any("err", err))
			}
		}()

		slog.Info("started task run", slog.String("name", name))
		err := f(ctx)
		if err != nil {
			slog.Error("err in task", slog.String("name", name), slog.Any("err", err))
			return
		}
		slog.Info("finished task run", slog.String("name", name))
	}
}

type drone struct {
	boardwhite *boardwhite.Service
}

func (d *drone) publishLCDaily(ctx context.Context) error {
	return d.boardwhite.PublishLCDaily(ctx)
}
