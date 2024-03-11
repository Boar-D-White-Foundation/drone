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

func NewBoarDWhiteService(cfg Config) (*boardwhite.Service, func(), error) {
	telegramClient, err := tg.NewClient(cfg.TgKey, cfg.BoarDWhiteChatID)
	if err != nil {
		return nil, nil, fmt.Errorf("new tg client: %w", err)
	}

	db, err := NewBadger(cfg.BadgerPath)
	if err != nil {
		return nil, nil, fmt.Errorf("db open: %w", err)
	}
	closeFn := func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close db", err)
		}
	}

	return boardwhite.NewService(
		cfg.BoarDWhiteLeetCodeThreadID,
		cfg.LCDailyStickerID,
		telegramClient,
		db,
	), closeFn, nil
}

func StartDrone(ctx context.Context, cfg Config) error {
	bw, closeFn, err := NewBoarDWhiteService(cfg)
	if err != nil {
		return err
	}
	defer closeFn()

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return err
	}

	jobs := make([]job, 0)
	jb, err := registerJob(ctx, scheduler, "PublishLCDaily", cfg.LCDailyCron, bw.PublishLCDaily)
	if err != nil {
		return err
	}
	jobs = append(jobs, jb)

	jb, err = registerJob(ctx, scheduler, "PublishNCDaily", cfg.NCDailyCron, bw.PublishNCDaily)
	if err != nil {
		return err
	}
	jobs = append(jobs, jb)

	scheduler.Start()
	slog.Info("started scheduler")
	for _, jb := range jobs {
		t, err := jb.NextRun()
		if err != nil {
			return err
		}
		slog.Info(
			"scheduled job",
			slog.String("name", jb.name),
			slog.String("cron", jb.cron),
			slog.String("next_run", t.String()),
		)
	}

	<-ctx.Done()
	return scheduler.Shutdown()
}

type job struct {
	gocron.Job

	name string
	cron string
}

func registerJob(
	ctx context.Context,
	s gocron.Scheduler,
	name, cron string,
	f func(context.Context) error,
) (job, error) {
	jb, err := s.NewJob(
		gocron.CronJob(cron, false),
		gocron.NewTask(wrapErrors(name, f), ctx),
	)
	if err != nil {
		return job{}, err
	}

	return job{
		Job:  jb,
		name: name,
		cron: cron,
	}, nil
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
