//go:build e2e

package alert_test

import (
	"errors"
	"testing"

	"github.com/boar-d-white-foundation/drone/alert"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/stretchr/testify/require"
)

func TestAlert(t *testing.T) {
	cfg, err := config.Load(config.Path())
	require.NoError(t, err)

	alertManager, err := alert.NewManagerFromConfig(cfg)
	require.NoError(t, err)

	err = errors.New("some err")
	alertManager.Errorxf(err, "test err %s", "test arg")
}
