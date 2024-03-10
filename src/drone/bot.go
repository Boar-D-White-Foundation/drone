package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/frosthamster/drone/src/leetcode"
	"github.com/frosthamster/drone/src/tg"
	"github.com/go-co-op/gocron/v2"
	tele "gopkg.in/telebot.v3"
)

func StartDrone(ctx context.Context, cfg Config) error {
	bot, err := tele.NewBot(tele.Settings{
		Token: cfg.TgKey,
	})
	if err != nil {
		return err
	}

	db, err := badger.Open(badger.DefaultOptions(cfg.BadgerPath))
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}

	dr := drone{
		bot: bot,
		tgManager: tg.Manager{
			BoarDWhiteChatID:           cfg.BoarDWhiteChatID,
			BoarDWhiteLeetCodeThreadID: cfg.BoarDWhiteLeetCodeThreadID,
			LCDailyStickerID:           cfg.LCDailyStickerID,
		},
		db: db,
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
	bot       *tele.Bot
	db        *badger.DB
	tgManager tg.Manager
}

func (d *drone) publishLCDaily(ctx context.Context) error {
	link, err := leetcode.GetDailyLink(ctx)
	if err != nil {
		return fmt.Errorf("get link: %w", err)
	}

	key := []byte("last_link")
	err = d.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		switch {
		case err == nil:
			// ok
		case errors.Is(err, badger.ErrKeyNotFound):
			// ok
		default:
			return fmt.Errorf("get key %q: %w", key, err)
		}

		if item.String() == link {
			return nil
		}

		err = d.tgManager.SendLCDailyToBoarDWhite(d.bot, tg.DefaultDailyHeader, link)
		if err != nil {
			return fmt.Errorf("send daily: %w", err)
		}

		err = txn.Set(key, []byte(link))
		if err != nil {
			return fmt.Errorf("set key %q: %w", key, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("db update: %w", err)
	}

	return nil
}
