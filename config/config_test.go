package config

import (
	"strings"
	"testing"

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

func TestConfigOverride(t *testing.T) {
	t.Parallel()

	cfg, err := Default()
	require.NoError(t, err)

	tgKey := cfg.Tg.Key
	assert.NotEmpty(t, cfg.Tg.LongPollerTimeout)

	cfg, err = Load("testdata/override.yaml")
	require.NoError(t, err)

	assert.NotEqual(t, tgKey, cfg.Tg.Key)
	assert.NotEmpty(t, cfg.Tg.LongPollerTimeout)
}

func TestConfigMocks(t *testing.T) {
	t.Parallel()

	cfg, err := Default()
	require.NoError(t, err)

	bwCfg, err := cfg.ServiceConfig()
	require.NoError(t, err)

	assert.NotEmpty(t, bwCfg.Mocks)
	for username, v := range bwCfg.Mocks {
		assert.NotEmpty(t, username)
		assert.NotEmpty(t, v.Period)
		assert.NotEmpty(t, v.StickerID)
	}
}
