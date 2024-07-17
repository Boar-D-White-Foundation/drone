package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/boardwhite"
	"github.com/boar-d-white-foundation/drone/chrome"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/dbq"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/boar-d-white-foundation/drone/tg"
	"github.com/go-co-op/gocron/v2"
)

func StartDrone(ctx context.Context, cfg config.Config) error {
	adminTGClient, err := tg.NewAdminClientFromConfig(cfg)
	if err != nil {
		return err
	}

	alerts := alert.NewManager(adminTGClient)

	var imageGenerator *chrome.ImageGenerator
	if cfg.Features.SnippetsGenerationEnabled {
		browser, cleanup, err := chrome.NewRemote(cfg.Rod.Host, cfg.Rod.Port)
		if err != nil {
			return err
		}
		defer cleanup()

		imageGenerator = chrome.NewImageGeneratorFromCfg(cfg, browser)
		if err := imageGenerator.WarmUpCaches(ctx); err != nil {
			return fmt.Errorf("fonts cache loading: %w", err)
		}
	}

	lcClient := leetcode.NewClientFromConfig(cfg)

	tgService, err := tg.NewBoardwhiteServiceFromConfig(cfg)
	if err != nil {
		return err
	}

	database := db.NewBadgerDBFromConfig(cfg)
	if err := database.Start(ctx); err != nil {
		return err
	}
	defer database.Stop()

	bw, err := boardwhite.NewServiceFromConfig(cfg, tgService, database, alerts, imageGenerator, lcClient)
	if err != nil {
		return err
	}

	bw.RegisterHandlers(ctx, tgService)
	tgService.Start()
	defer tgService.Stop()
	slog.Info("started tg handlers")

	dbqRegistry := dbq.NewRegistry()
	if err := bw.RegisterTasks(dbqRegistry); err != nil {
		return err
	}

	queue, err := dbq.NewQueue(dbqRegistry, database)
	if err != nil {
		return err
	}

	dbqDone := make(chan struct{})
	go func() {
		queue.StartHandlers(ctx, time.Second)
		dbqDone <- struct{}{}
	}()
	slog.Info("started dbq")

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
	<-dbqDone
	return scheduler.Shutdown()
}

type job struct {
	gocron.Job

	name string
	cron string
}

func registerCronJobs(
	ctx context.Context,
	cfg config.Config,
	scheduler gocron.Scheduler,
	bw *boardwhite.Service,
) ([]job, error) {
	jobs := make([]job, 0)
	jb, err := registerJob(ctx, scheduler, "PublishLCDaily", cfg.LeetcodeDaily.Cron, bw.PublishLCDaily)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, jb)

	jb, err = registerJob(ctx, scheduler, "PublishLCChickensDaily", cfg.LeetcodeDaily.Cron, bw.PublishLCChickensDaily)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, jb)

	jb, err = registerJob(ctx, scheduler, "PublishLCRating", cfg.LeetcodeDaily.RatingCron, bw.PublishLCRating)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, jb)

	jb, err = registerJob(
		ctx,
		scheduler,
		"PublishLCChickensRating",
		cfg.LeetcodeDaily.RatingCron,
		bw.PublishLCChickensRating,
	)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, jb)

	jb, err = registerJob(ctx, scheduler, "PublishNCDaily", cfg.NeetcodeDaily.Cron, bw.PublishNCDaily)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, jb)

	jb, err = registerJob(ctx, scheduler, "PublishNCRating", cfg.NeetcodeDaily.RatingCron, bw.PublishNCRating)
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
