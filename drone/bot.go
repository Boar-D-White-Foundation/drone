package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/boar-d-white-foundation/drone/boardwhite"
	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/tg"
	"github.com/go-co-op/gocron/v2"
)

func NewTgServiceFromConfig(cfg Config) (*tg.Service, error) {
	tgService, err := tg.NewService(cfg.Tg.Key, cfg.Boardwhite.ChatID, cfg.Tg.LongPollerTimeout)
	if err != nil {
		return nil, fmt.Errorf("new tg client: %w", err)
	}

	return tgService, nil
}

func NewDBFromConfig(cfg Config) db.DB {
	return db.NewBadgerBD(cfg.BadgerPath)
}

func NewBoarDWhiteServiceFromConfig(
	telegram tg.Client,
	database db.DB,
	cfg boardwhite.ServiceConfig,
) (*boardwhite.Service, error) {
	return boardwhite.NewService(
		cfg,
		telegram,
		database,
	)
}

func StartDrone(ctx context.Context, cfg Config) error {
	tgService, err := NewTgServiceFromConfig(cfg)
	if err != nil {
		return err
	}

	database := NewDBFromConfig(cfg)
	if err := database.Start(ctx); err != nil {
		return err
	}
	defer database.Stop()

	bwCfg, err := cfg.ServiceConfig()
	if err != nil {
		return fmt.Errorf("service config: %w", err)
	}
	bw, err := NewBoarDWhiteServiceFromConfig(tgService, database, bwCfg)
	if err != nil {
		return err
	}

	bw.RegisterHandlers(ctx, tgService)
	tgService.Start()
	defer tgService.Stop()
	slog.Info("started tg handlers")

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return err
	}

	jobs, err := registerCronJobs(ctx, cfg, scheduler, bw)
	if err != nil {
		return err
	}

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

func registerCronJobs(
	ctx context.Context,
	cfg Config,
	scheduler gocron.Scheduler,
	bw *boardwhite.Service,
) ([]job, error) {
	jobs := make([]job, 0)
	jb, err := registerJob(ctx, scheduler, "PublishLCDaily", cfg.LeetcodeDaily.Cron, bw.PublishLCDaily)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, jb)

	jb, err = registerJob(ctx, scheduler, "PublishNCDaily", cfg.NeetcodeDaily.Cron, bw.PublishNCDaily)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, jb)
	return jobs, nil
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
				slog.Error("panic in cron task", slog.String("name", name), slog.Any("err", err))
			}
		}()

		slog.Info("started cron task run", slog.String("name", name))
		err := f(ctx)
		if err != nil {
			slog.Error("err in cron task", slog.String("name", name), slog.Any("err", err))
			return
		}
		slog.Info("finished cron task run", slog.String("name", name))
	}
}
