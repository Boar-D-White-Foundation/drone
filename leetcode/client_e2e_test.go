//go:build e2e

package leetcode_test

import (
	"context"
	"testing"

	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/stretchr/testify/require"
)

func TestGetSubmission(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.Load(config.Path())
	require.NoError(t, err)

	client := leetcode.NewClientFromConfig(cfg)
	submission, err := client.GetSubmission(ctx, "1312071772")
	require.NoError(t, err)
	require.NotEmpty(t, submission.Code)
}
