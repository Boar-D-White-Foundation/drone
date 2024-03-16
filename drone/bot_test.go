package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSkipDrone(t *testing.T) {
	ctx := context.Background()
	cfg, err := LoadConfig()
	require.NoError(t, err)
	bw, err := NewBoarDWhiteService(cfg)
	require.NoError(t, err)

	err = bw.Start(ctx)
	require.NoError(t, err)
	defer bw.Stop()

	err = bw.PublishLCDaily(ctx)
	require.NoError(t, err)

	err = bw.PublishNCDaily(ctx)
	require.NoError(t, err)
}
