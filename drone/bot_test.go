package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSkipE2EDrone(t *testing.T) {
	ctx := context.Background()
	cfg, err := LoadConfig()
	require.NoError(t, err)

	tgService, err := NewTgServiceFromConfig(cfg)
	require.NoError(t, err)

	database := NewDBFromConfig(cfg)
	err = database.Start(ctx)
	require.NoError(t, err)
	defer database.Stop()

	bw, err := NewBoarDWhiteServiceFromConfig(tgService, database, cfg.ServiceConfig())
	require.NoError(t, err)

	err = bw.PublishLCDaily(ctx)
	require.NoError(t, err)

	err = bw.PublishNCDaily(ctx)
	require.NoError(t, err)

	bw.RegisterHandlers(ctx, tgService)
	tgService.Start()
	defer tgService.Stop()
	<-ctx.Done()
}
