package neetcode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuestions(t *testing.T) {
	qs, err := Questions()
	require.NoError(t, err)

	for _, q := range qs {
		require.NotEmpty(t, q.LeetcodeLink())
		require.NotEmpty(t, q.LeetcodeCaLink())
	}
}
