package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/frosthamster/drone/src/leetcode"
	"github.com/frosthamster/drone/src/tg"
	"github.com/go-co-op/gocron/v2"
	tele "gopkg.in/telebot.v3"
)

type Config struct {
	TgKey                      string
	LCDailyCron                string
	BoarDWhiteChatID           tele.ChatID
	BoarDWhiteLeetCodeThreadID int
}

func StartDrone(ctx context.Context, cfg Config) error {
	bot, err := tele.NewBot(tele.Settings{
		Token: cfg.TgKey,
	})
	if err != nil {
		return err
	}
	dr := drone{
		bot: bot,
		tgManager: tg.Manager{
			BoarDWhiteChatID:           cfg.BoarDWhiteChatID,
			BoarDWhiteLeetCodeThreadID: cfg.BoarDWhiteLeetCodeThreadID,
		},
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
	fmt.Println("Started scheduler, publishLCDaily", cfg.LCDailyCron, "next run:", t.String())

	<-ctx.Done()
	return scheduler.Shutdown()
}

func wrapErrors(name string, f func(context.Context) error) func(context.Context) {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "PANIC:", name, err)
			}
		}()

		err := f(ctx)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "ERROR:", name, err.Error())
		}
	}
}

type drone struct {
	bot       *tele.Bot
	tgManager tg.Manager
}

func (d *drone) publishLCDaily(ctx context.Context) error {
	link, err := leetcode.GetDailyLink(ctx)
	if err != nil {
		return err
	}

	return d.tgManager.SendLCDailyToBoarDWhite(d.bot, tg.DefaultDailyHeader, link)
}
