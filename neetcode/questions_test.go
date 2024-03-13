package neetcode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroups(t *testing.T) {
	groups, err := Groups()
	require.NoError(t, err)

	for _, g := range groups {
		require.NotEmpty(t, g.Name)
		for _, q := range g.Questions {
			require.NotEmpty(t, q.Name)
			require.NotEmpty(t, q.Difficulty)
			require.NotEmpty(t, q.LCLink)
			require.NotEqual(t, -1, DifficultyToSortOrder(q.Difficulty))
		}
	}
}
