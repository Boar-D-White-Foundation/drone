//go:build e2e

package chrome_test

import (
	"context"
	"os"
	"testing"

	"github.com/boar-d-white-foundation/drone/chrome"
	"github.com/stretchr/testify/require"
)

func TestCarbon(t *testing.T) {
	ctx := context.Background()
	browser, cleanup, err := chrome.NewRemote("127.0.0.1", 7317)
	require.NoError(t, err)
	defer cleanup()

	code, err := os.ReadFile("./carbon_e2e_test.go")
	require.NoError(t, err)

	buf, err := chrome.GenerateCodeSnippet(ctx, browser, string(code))
	require.NoError(t, err)

	err = os.WriteFile("code_snippet.png", buf, 0644)
	require.NoError(t, err)
}
