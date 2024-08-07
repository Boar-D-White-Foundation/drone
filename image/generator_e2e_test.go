//go:build e2e

package image_test

import (
	"context"
	"os"
	"testing"

	"github.com/boar-d-white-foundation/drone/chrome"
	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/image"
	"github.com/boar-d-white-foundation/drone/leetcode"
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
	browser.ServeMonitor("127.0.0.1:56174")

	imageGenerator := image.NewGeneratorFromCfg(cfg, browser)
	err = imageGenerator.WarmUpCaches(ctx)
	require.NoError(t, err)

	codeBytes, err := os.ReadFile("./image/generator_e2e_test.go")
	require.NoError(t, err)
	code := string(codeBytes)

	buf, err := imageGenerator.GenerateCodeSnippetCarbon(ctx, "1", leetcode.LangGO, code)
	require.NoError(t, err)

	err = os.WriteFile("snippet_carbon.png", buf, 0600)
	require.NoError(t, err)

	buf, err = imageGenerator.GenerateCodeSnippetRayso(ctx, "1", leetcode.LangGO, code)
	require.NoError(t, err)

	err = os.WriteFile("snippet_rayso.png", buf, 0600)
	require.NoError(t, err)

	buf, err = imageGenerator.GenerateCodeSnippetJavaHighlight(ctx, "1", leetcode.LangGO, code)
	require.NoError(t, err)

	err = os.WriteFile("snippet_java_highlight.png", buf, 0600)
	require.NoError(t, err)
}
