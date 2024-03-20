package neetcode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroups(t *testing.T) {
	t.Parallel()

	groups, err := Groups()
	require.NoError(t, err)

	totalCount := 0
	for _, g := range groups {
		totalCount += len(g.Questions)
		require.NotEmpty(t, g.Name)
		for _, q := range g.Questions {
			require.NotEmpty(t, q.Name)
			require.NotEmpty(t, q.Difficulty)
			require.NotEmpty(t, q.LCLink)
			require.NotEqual(t, -1, DifficultyToSortOrder(q.Difficulty))
		}
	}
	require.Equal(t, QuestionsTotalCount, totalCount)
}
