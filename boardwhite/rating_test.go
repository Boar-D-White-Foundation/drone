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

	rating := buildRating(stats, 6, 8, ratingOpts{})
	require.Len(t, rating.rows, 10)
	require.Equal(t, "@cauchy2384", rating.rows[0].Mention)
	require.Equal(t, 3, rating.rows[0].Solved)
	for _, row := range rating.rows {
		require.LessOrEqual(t, row.MaxStreak, row.Solved)
		require.LessOrEqual(t, row.CurrentStreak, row.Solved)
		require.LessOrEqual(t, row.CurrentStreak, row.MaxStreak)
	}

	require.NotEmpty(t, rating.toMarkdownV2("header"))
}
