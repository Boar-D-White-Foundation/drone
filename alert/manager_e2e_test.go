//go:build e2e

package alert_test

import (
	"errors"
	"testing"

	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/tg"
	"github.com/stretchr/testify/require"
)

func TestAlert(t *testing.T) {
	cfg, err := config.Load(config.Path())
	require.NoError(t, err)

	adminTGClient, err := tg.NewAdminClientFromConfig(cfg)
	require.NoError(t, err)

	alerts := alert.NewManager(adminTGClient)

	err = errors.New("some err")
	alerts.Errorxf(err, "test err %s", "test arg")
}
