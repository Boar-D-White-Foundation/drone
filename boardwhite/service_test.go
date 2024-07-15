package boardwhite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLCSubmissionRE(t *testing.T) {
	type testCase struct {
		link       string
		expectedID string
	}
	testCases := []testCase{
		{
			link:       "https://leetcode.com/problems/create-binary-tree-from-descriptions/submissions/1321938777/",
			expectedID: "1321938777",
		},
		{
			link:       "https://leetcode.com/submissions/detail/1322062899/",
			expectedID: "1322062899",
		},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			matches := lcSubmissionRe.FindStringSubmatch(tc.link)
			require.Len(t, matches, 2)
			require.Equal(t, tc.expectedID, matches[1])
		})
	}
}
