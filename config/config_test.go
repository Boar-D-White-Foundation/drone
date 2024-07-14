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
	tgSession := cfg.Tg.Session
	tgCSRF := cfg.Tg.CSRF
	assert.NotEmpty(t, cfg.Tg.LongPollerTimeout)

	cfg, err = Load("testdata/override.yaml")
	require.NoError(t, err)

	assert.NotEqual(t, tgKey, cfg.Tg.Key)
	assert.NotEqual(t, tgSession, cfg.Tg.Session)
	assert.NotEqual(t, tgCSRF, cfg.Tg.CSRF)
	assert.NotEmpty(t, cfg.Tg.LongPollerTimeout)
}
