//go:build e2e

package main

import (
	"context"
	"log/slog"
	"testing"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/boardwhite"
	"github.com/boar-d-white-foundation/drone/chrome"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/boar-d-white-foundation/drone/tg"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

func TestDrone(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.Load(config.Path())
	require.NoError(t, err)

	alerts, err := alert.NewManagerFromConfig(cfg)
	require.NoError(t, err)

	browser, cleanup, err := chrome.NewRemote(cfg.Rod.Host, cfg.Rod.Port)
	require.NoError(t, err)
	defer cleanup()

	lcClient := leetcode.NewClientFromConfig(cfg)

	tgService, err := tg.NewBoardwhiteServiceFromConfig(cfg)
	require.NoError(t, err)

	database := db.NewBadgerDBFromConfig(cfg)
	err = database.Start(ctx)
	require.NoError(t, err)
	defer database.Stop()

	bw, err := boardwhite.NewServiceFromConfig(cfg, tgService, database, alerts, browser, lcClient)
	require.NoError(t, err)

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
	tgService.Start()
	defer tgService.Stop()
	<-ctx.Done()
}
