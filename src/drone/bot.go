package main

import (
	"context"
	"fmt"
	"time"

	"github.com/frosthamster/drone/src/leetcode"
	"github.com/frosthamster/drone/src/tg"
	"github.com/go-co-op/gocron/v2"
	tele "gopkg.in/telebot.v3"
)

type Config struct {
	TgKey       string
	LCDailyCron string
}

func StartDrone(ctx context.Context, cfg Config) error {
	bot, err := tele.NewBot(tele.Settings{
		Token: cfg.TgKey,
	})
	if err != nil {
		return err
	}
	dr := drone{bot: bot}

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return err
	}

	if len(cfg.LCDailyCron) == 0 {
		cfg.LCDailyCron = "0 1 * * *" // every day at 01:00 UTC
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
	fmt.Println("Started scheduler, publishLCDaily", cfg.LCDailyCron, "next run:", t.String())

	<-ctx.Done()
	return scheduler.Shutdown()
}

func wrapErrors(name string, f func(context.Context) error) func(context.Context) {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("PANIC:", name, err)
			}
		}()

		err := f(ctx)
		if err != nil {
			fmt.Println("ERROR:", name, err.Error())
		}
	}
}

type drone struct {
	bot *tele.Bot
}

func (d *drone) publishLCDaily(ctx context.Context) error {
	link, err := leetcode.GetDailyLink(ctx)
	if err != nil {
		return err
	}

	return tg.SendLCDailyToBoarDWhite(d.bot, tg.DefaultDailyHeader, link)
}
