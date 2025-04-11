//go:build e2e

package media_test

import (
	"context"
	"os"
	"testing"

	"github.com/boar-d-white-foundation/drone/chrome"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/boar-d-white-foundation/drone/media"
	"github.com/stretchr/testify/require"
)

func TestSnippetsGeneration(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.Load(config.Path())
	require.NoError(t, err)

	browser, cleanup, err := chrome.NewRemote("127.0.0.1", 7317)
	//browser, cleanup, err := chrome.NewLocal()
	require.NoError(t, err)
	defer cleanup()
	//browser.ServeMonitor("127.0.0.1:56174")

	mediaGenerator := media.NewGeneratorFromCfg(cfg, browser)

	codeBytes, err := os.ReadFile("./media/testdata/main.rs")
	require.NoError(t, err)
	code := string(codeBytes)

	buf, err := mediaGenerator.GenerateCodeSnippet(ctx, "1", leetcode.LangGO, code)
	require.NoError(t, err)

	err = os.WriteFile("snippet_java_highlight.png", buf, 0600)
	require.NoError(t, err)
}
