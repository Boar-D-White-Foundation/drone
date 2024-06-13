package leetcode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuestions(t *testing.T) {
	t.Parallel()

	questions, err := Questions()
	require.NoError(t, err)
	require.NotEmpty(t, questions)

	ids := make(map[string]struct{})
	for _, q := range questions {
		require.NotEmpty(t, q.ID)
		require.NotEmpty(t, q.Name)
		require.Contains(t, []Difficulty{DifficultyEasy, DifficultyMedium, DifficultyHard}, q.Difficulty)

		_, ok := ids[q.ID]
		require.False(t, ok)
		ids[q.ID] = struct{}{}
	}
}
