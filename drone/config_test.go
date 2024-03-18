package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Setenv("DRONE_MOCKS", `byegor;72h;CAACAgIAAxkBAAELtUJl9FKjhIGnyaUwO_IXh_SepPBgSAACzzwAAiTYUEn0kbWw7nXa1zQE,lk4d4;72h;CAACAgQAAxkBAAELu8Rl90uOqEMPwdCvcFIm8nBMpVNyoAACBwIAAnBt9gd0v3XadwsPfzQE`)

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 2, len(cfg.Mocks))
	for _, v := range cfg.Mocks {
		assert.NotEmpty(t, v.Username)
		assert.NotEmpty(t, v.Period)
		assert.NotEmpty(t, v.StickerID)
	}
	fmt.Println(cfg.Mocks)
}
