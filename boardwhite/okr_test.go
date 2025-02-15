package boardwhite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildOkrMsg(t *testing.T) {
	t.Parallel()

	counts := make(map[okrTag]int)
	i := 0
	for tag := range okrGoals {
		counts[tag] = i
		i++
	}
	msg, err := buildOkrProgressMsg(counts)
	require.NoError(t, err)
	require.NotEmpty(t, msg)
}

func TestOkrInit(t *testing.T) {
	t.Parallel()

	okrs := okrs{Updates: []okrUpdate{{}}}
	okrs.init()
	require.NotNil(t, okrs.TotalCount)
	require.NotNil(t, okrs.Updates[0].Counts)
}
