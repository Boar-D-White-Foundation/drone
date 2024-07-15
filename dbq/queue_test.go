package dbq_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/dbq"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	database := db.NewBadgerDB(":memory:")
	err := database.Start(ctx)
	require.NoError(t, err)
	defer database.Stop()

	registry := dbq.NewRegistry()
	result := make(chan int)
	shouldRetry := true
	task, err := dbq.RegisterHandler(registry, "task", func(ctx context.Context, tx db.Tx, i int) error {
		if shouldRetry {
			shouldRetry = false
			return errors.New("retry")
		}

		result <- i
		return nil
	})
	require.NoError(t, err)

	queue, err := dbq.NewQueue(registry, database)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		queue.StartHandlers(ctx, time.Second)
		done <- struct{}{}
	}()

	err = task.Schedule(ctx, 1, 42)
	require.NoError(t, err)

	require.Equal(t, 42, <-result)
	cancel()
	<-done
}
