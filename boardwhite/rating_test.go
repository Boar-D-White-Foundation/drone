package boardwhite

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/nc_stats.json
var rawNCStats []byte

func TestBuildRating(t *testing.T) {
	t.Parallel()

	var stats stats
	err := json.Unmarshal(rawNCStats, &stats)
	require.NoError(t, err)

	rating := buildRating(stats)
	require.Len(t, rating, 12)
	require.Equal(t, "faucct", rating[0].Username)
	require.Equal(t, 5, rating[0].Solved)
	for _, row := range rating {
		require.LessOrEqual(t, row.MaxStreak, row.Solved)
		require.LessOrEqual(t, row.CurrentStreak, row.Solved)
		require.LessOrEqual(t, row.CurrentStreak, row.MaxStreak)
	}
}
