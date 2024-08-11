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
	lcSession := cfg.Leetcode.Session
	lcCSRF := cfg.Leetcode.CSRF

	cfg, err = Load("testdata/override.yaml")
	require.NoError(t, err)

	assert.NotEqual(t, tgKey, cfg.Tg.Key)
	assert.NotEmpty(t, cfg.Tg.LongPollerTimeout)
	assert.NotEqual(t, lcSession, cfg.Leetcode.Session)
	assert.NotEqual(t, lcCSRF, cfg.Leetcode.CSRF)
}
