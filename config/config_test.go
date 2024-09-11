package config

import (
	"strings"
	"testing"

	"github.com/boar-d-white-foundation/drone/iterx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigHideSecrets(t *testing.T) {
	t.Parallel()

	cfg, err := Default()
	require.NoError(t, err)

	cfg.Tg.Key = "secret"
	s := cfg.String()
	require.False(t, strings.Contains(s, "apiKey"))
	require.False(t, strings.Contains(s, cfg.Tg.Key))
}

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg, err := Default()
	require.NoError(t, err)

	// all sticker IDs must be unique
	assert.Equal(t, cfg.DailyStickerIDs, iterx.Uniq(cfg.DailyStickerIDs))
	assert.Equal(t, cfg.DailyChickensStickerIDs, iterx.Uniq(cfg.DailyChickensStickerIDs))
	for _, mock := range cfg.Mocks {
		assert.Equal(t, mock.StickerIDs, iterx.Uniq(mock.StickerIDs))
	}
}

func TestConfigOverride(t *testing.T) {
	t.Parallel()

	cfg, err := Default()
	require.NoError(t, err)

	tgKey := cfg.Tg.Key
	assert.NotEmpty(t, cfg.Tg.LongPollerTimeout)
	lcSession := cfg.Leetcode.Session
	lcCSRF := cfg.Leetcode.CSRF

	cfg, err = Load("testdata/override.yaml")
	require.NoError(t, err)

	assert.NotEqual(t, tgKey, cfg.Tg.Key)
	assert.NotEmpty(t, cfg.Tg.LongPollerTimeout)
	assert.NotEqual(t, lcSession, cfg.Leetcode.Session)
	assert.NotEqual(t, lcCSRF, cfg.Leetcode.CSRF)
}
