package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigTgKey(t *testing.T) {
	const tgKey = "SECRET KEY"
	t.Setenv("DRONE_TG_BOT_API_KEY", tgKey)

	cfg, err := DefaultConfig()
	require.NoError(t, err)

	assert.Empty(t, cfg.Tg.Key)

	cfg, err = LoadConfig("default_config.yaml")
	require.NoError(t, err)

	assert.Equal(t, tgKey, cfg.Tg.Key)
}

func TestConfigMocks(t *testing.T) {
	cfg, err := DefaultConfig()
	require.NoError(t, err)

	bwCfg, err := cfg.ServiceConfig()
	require.NoError(t, err)

	assert.Len(t, bwCfg.Mocks, 3)
	for username, v := range bwCfg.Mocks {
		assert.NotEmpty(t, username)
		assert.Equal(t, 72*time.Hour, v.Period)
		assert.NotEmpty(t, v.StickerID)
	}
}
