//go:build e2e

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/boardwhite"
	"github.com/boar-d-white-foundation/drone/chrome"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/dbq"
	"github.com/boar-d-white-foundation/drone/image"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/boar-d-white-foundation/drone/tg"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

func TestDrone(t *testing.T) {
	// arrange
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.Load(config.Path())
	require.NoError(t, err)

	adminTGClient, err := tg.NewAdminClientFromConfig(cfg)
	require.NoError(t, err)

	alerts := alert.NewManager(adminTGClient)

	browser, cleanup, err := chrome.NewLocal()
	require.NoError(t, err)
	defer cleanup()

	imageGenerator := image.NewGeneratorFromCfg(cfg, browser)
	err = imageGenerator.WarmUpCaches(ctx)
	require.NoError(t, err)

	lcClient := leetcode.NewClientFromConfig(cfg)

	tgService, err := tg.NewBoardwhiteServiceFromConfig(cfg, alerts)
	require.NoError(t, err)

	database := db.NewBadgerDBFromConfig(cfg)
	err = database.Start(ctx)
	require.NoError(t, err)
	defer database.Stop()

	dbqRegistry := dbq.NewRegistry()

	bw, err := boardwhite.NewServiceFromConfig(cfg, tgService, database, alerts, imageGenerator, lcClient)
	require.NoError(t, err)

	// act
	err = bw.PublishLCDaily(ctx)
	require.NoError(t, err)

	err = bw.PublishLCChickensDaily(ctx)
	require.NoError(t, err)

	err = bw.PublishNCDaily(ctx)
	require.NoError(t, err)

	err = bw.PublishLCRating(ctx)
	require.NoError(t, err)

	err = bw.PublishLCChickensRating(ctx)
	require.NoError(t, err)

	err = bw.PublishNCRating(ctx)
	require.NoError(t, err)

	bw.RegisterHandlers(ctx, tgService)
	tgService.RegisterHandler(tele.OnText, "OnDummy", func(c tele.Context) error {
		slog.Info("got update", slog.Any("ctx", c))
		return nil
	})

	err = bw.RegisterTasks(dbqRegistry)
	require.NoError(t, err)

	// assert
	queue, err := dbq.NewQueue(dbqRegistry, database)
	require.NoError(t, err)

	dbqDone := make(chan struct{})
	go func() {
		queue.StartHandlers(ctx, 30*time.Second)
		dbqDone <- struct{}{}
	}()

	tgService.Start()
	defer tgService.Stop()

	<-ctx.Done()
	<-dbqDone
}
