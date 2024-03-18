package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Setenv(
		"DRONE_MOCKS",
		`byegor;72h;CAACAgIAAxkBAAELtUJl9FKjhIGnyaUwO_IXh_SepPBgSAACzzwAAiTYUEn0kbWw7nXa1zQE,lk4d4;72h;CAACAgQAAxkBAAELu8Rl90uOqEMPwdCvcFIm8nBMpVNyoAACBwIAAnBt9gd0v3XadwsPfzQE,ollkostin;72h;CAACAgQAAxkBAAELvRpl-EweMqCeuggLoAo3ysvFmONONgACvRAAAqbxcR5BuzVAQyP23DQE'`,
	)

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Len(t, cfg.Mocks, 3)
	for _, v := range cfg.Mocks {
		assert.NotEmpty(t, v.Username)
		assert.NotEmpty(t, v.Period)
		assert.NotEmpty(t, v.StickerID)
	}

	bwCfg, err := cfg.ServiceConfig()
	require.NoError(t, err)

	assert.Len(t, bwCfg.Mocks, 3)
	for username, v := range bwCfg.Mocks {
		assert.NotEmpty(t, username)
		assert.Equal(t, 72*time.Hour, v.Period)
		assert.NotEmpty(t, v.StickerID)
	}
}
