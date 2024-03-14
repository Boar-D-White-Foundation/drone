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
	bw, closeFn, err := NewBoarDWhiteService(cfg)
	require.NoError(t, err)
	defer closeFn()

	err = bw.PublishLCDaily(ctx)
	require.NoError(t, err)

	err = bw.PublishNCDaily(ctx)
	require.NoError(t, err)
}
