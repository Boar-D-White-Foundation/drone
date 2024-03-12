package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSkipPublishLCDaily(t *testing.T) {
	ctx := context.Background()
	cfg, err := LoadConfig()
	require.NoError(t, err)
	bw, closeFn, err := NewBoarDWhiteService(cfg)
	require.NoError(t, err)
	defer closeFn()
	err = bw.PublishLCDaily(ctx)
	require.NoError(t, err)
}
