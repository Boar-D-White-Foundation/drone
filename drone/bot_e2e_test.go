//go:build e2e

package main

import (
	"context"
	"testing"

	"github.com/boar-d-white-foundation/drone/config"
	"github.com/stretchr/testify/require"
)

func TestDrone(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.Load(config.Path())
	require.NoError(t, err)

	tgService, err := NewTgServiceFromConfig(cfg)
	require.NoError(t, err)

	database := NewDBFromConfig(cfg)
	err = database.Start(ctx)
	require.NoError(t, err)
	defer database.Stop()

	bwCfg, err := cfg.ServiceConfig()
	require.NoError(t, err)
	bw, err := NewBoarDWhiteServiceFromConfig(tgService, database, bwCfg)
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
