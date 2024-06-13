package boardwhite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLCChickenQuestions(t *testing.T) {
	t.Parallel()

	questions, err := newLCChickenQuestions()
	require.NoError(t, err)
	require.NotEmpty(t, questions.questions)
	require.NotEmpty(t, questions.shuffledPosition)
	require.Equal(t, len(questions.questions), len(questions.shuffledPosition))

	counts := make([]int, len(questions.questions))
	for _, i := range questions.shuffledPosition {
		counts[i]++
		_ = questions.questions[i]
	}
	for _, cnt := range counts {
		require.Equal(t, 1, cnt)
	}

	for i := 0; i < 10; i++ {
		repeatedQuestions, err := newLCChickenQuestions()
		require.NoError(t, err)
		require.Equal(t, questions.shuffledPosition, repeatedQuestions.shuffledPosition)
	}
}
