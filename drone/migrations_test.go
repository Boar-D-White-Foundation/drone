package main

import (
	"context"
	"testing"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	bdb := db.NewBadgerDB(":memory:")
	err := bdb.Start(ctx)
	require.NoError(t, err)
	defer bdb.Stop()

	err = migrate(ctx, bdb)
	require.NoError(t, err)
}
