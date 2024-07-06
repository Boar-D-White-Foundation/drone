//go:build e2e

package chrome_test

import (
	"context"
	"os"
	"testing"

	"github.com/boar-d-white-foundation/drone/chrome"
	"github.com/stretchr/testify/require"
)

func TestT(t *testing.T) {
	// docker run --rm -p 7317:7317 ghcr.io/go-rod/rod:v0.116.1
	ctx := context.Background()
	d, err := chrome.NewDriver(chrome.DriverConfig{
		Host: "127.0.0.1",
		Port: 7317,
	})
	require.NoError(t, err)
	defer d.Close()

	buf, err := chrome.GenerateCodeSnippet(ctx, d.Browser(), "# some code\ndef f():\n    print('hello world')")
	require.NoError(t, err)

	err = os.WriteFile("code_snippet.png", buf, 0644)
	require.NoError(t, err)
}
