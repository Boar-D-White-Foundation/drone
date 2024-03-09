package leetcode_test

import (
	"context"
	"testing"

	"github.com/frosthamster/drone/src/leetcode"
	"github.com/stretchr/testify/require"
)

func TestGetDailyLink(t *testing.T) {
	ctx := context.Background()
	link, err := leetcode.GetDailyLink(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, link)
}
