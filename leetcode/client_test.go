package leetcode_test

import (
	"encoding/json"
	"testing"

	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/stretchr/testify/require"
)

func TestMarshalling(t *testing.T) {
	submission := leetcode.Submission{
		ID:                "123",
		Runtime:           1,
		RuntimePercentile: 2.5,
		Memory:            2,
		MemoryPercentile:  3.5,
		Code:              "code",
		Lang:              leetcode.LangGO,
		TotalCorrect:      6,
		TotalTestcases:    10,
	}

	bytes, err := json.Marshal(submission)
	require.NoError(t, err)

	var unmarshalled leetcode.Submission
	err = json.Unmarshal(bytes, &unmarshalled)
	require.NoError(t, err)

	require.Equal(t, submission, unmarshalled)
}
