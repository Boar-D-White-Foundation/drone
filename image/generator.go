package image

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/boar-d-white-foundation/drone/retry"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type GeneratorConfig struct {
	CarbonURL              string
	RaysoURL               string
	JavaHighlightURL       string
	UseCarbon              bool
	UseRayso               bool
	UseJavaHighlight       bool
	RodDownloadsSaveFolder string
	RodDownloadsGetFolder  string
}

type Generator struct {
	cfg     GeneratorConfig
	browser *rod.Browser
	client  *http.Client
}

func NewGenerator(
	cfg GeneratorConfig,
	browser *rod.Browser,
) *Generator {
	client := http.Client{Timeout: 5 * time.Second}
	return &Generator{
		cfg:     cfg,
		browser: browser,
		client:  &client,
	}
}

func NewGeneratorFromCfg(
	cfg config.Config,
	browser *rod.Browser,
) *Generator {
	return NewGenerator(
		GeneratorConfig{
			CarbonURL:              cfg.ImageGenerator.CarbonURL,
			RaysoURL:               cfg.ImageGenerator.RaysoURL,
			JavaHighlightURL:       cfg.ImageGenerator.JavaHighlightURL,
			UseCarbon:              cfg.ImageGenerator.UseCarbon,
			UseRayso:               cfg.ImageGenerator.UseRayso,
			UseJavaHighlight:       cfg.ImageGenerator.UseJavaHighlight,
			RodDownloadsSaveFolder: cfg.Rod.DownloadsFolder,
			RodDownloadsGetFolder:  cfg.ImageGenerator.RodDownloadsFolder,
		},
		browser,
	)
}

// WarmUpCaches generates images from all sources to correctly warm up fonts caches
func (g *Generator) WarmUpCaches(ctx context.Context) error {
	if _, err := g.GenerateCodeSnippetCarbon(ctx, "warmup", leetcode.LangUnknown, "warmup"); err != nil {
		return fmt.Errorf("warmup carbon: %w", err)
	}

	if _, err := g.GenerateCodeSnippetRayso(ctx, "warmup", leetcode.LangUnknown, "warmup"); err != nil {
		return fmt.Errorf("warmup rayso: %w", err)
	}

	return nil
}

func normalizeCode(code string) string {
	return strings.Trim(code, "\n")
}

func toJavaHighlightLang(lang leetcode.Lang) string {
	switch lang {
	case leetcode.LangCPP:
		return "cpp"
	case leetcode.LangJava:
		return "java"
	case leetcode.LangPy2:
		return "python"
	case leetcode.LangPy3:
		return "python3"
	case leetcode.LangC:
		return "c"
	case leetcode.LangCSharp:
		return "csharp"
	case leetcode.LangJS:
		return "javascript"
	case leetcode.LangTS:
		return "typescript"
	case leetcode.LangPHP:
		return "php"
	case leetcode.LangSwift:
		return "swift"
	case leetcode.LangKotlin:
		return "kotlin"
	case leetcode.LangGO:
		return "golang"
	case leetcode.LangRuby:
		return "ruby"
	case leetcode.LangScala:
		return "scala"
	case leetcode.LangRust:
		return "rust"
	case leetcode.LangRacket:
		return "racket"
	default:
		return ""
	}
}

func (g *Generator) GenerateCodeSnippetJavaHighlight(
	ctx context.Context,
	submissionID string,
	lang leetcode.Lang,
	code string,
) ([]byte, error) {
	code = normalizeCode(code)
	backoff := retry.LinearBackoff{
		Delay:       time.Second,
		MaxAttempts: 2,
	}
	return retry.Do(ctx, "java highlight snippet "+submissionID, backoff, func() ([]byte, error) {
		slog.Info("start generate code snippet", slog.String("submissionID", submissionID))
		uri := fmt.Sprintf(
			"%s/?l=%s&t=one-dark-vivid&p=30&ligatures=true",
			g.cfg.JavaHighlightURL, toJavaHighlightLang(lang),
		)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, strings.NewReader(code))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		resp, err := g.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch java hightligher image: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				slog.Error("err close resp body", slog.String("submissionID", submissionID), slog.Any("err", err))
			}
		}()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(
				"got non 200 resp for javahighlight generation %d: %s",
				resp.StatusCode, string(respBody),
			)
		}

		return respBody, nil
	})
}

func (g *Generator) GenerateCodeSnippetCarbon(
	ctx context.Context,
	submissionID string,
	lang leetcode.Lang,
	code string,
) ([]byte, error) {
	code = normalizeCode(code)
	backoff := retry.LinearBackoff{
		Delay:       time.Second,
		MaxAttempts: 8,
	}
	return retry.Do(ctx, "carbon snippet "+submissionID, backoff, func() ([]byte, error) {
		slog.Info("start generate code snippet", slog.String("submissionID", submissionID))
		uri := fmt.Sprintf(
			"%s/?t=vscode&es=4x&l=auto&ln=false&fm=Hack&code=%s",
			g.cfg.CarbonURL, url.QueryEscape(code),
		)
		page, err := g.browser.Timeout(30 * time.Second).Page(proto.TargetCreateTarget{URL: uri})
		if err != nil {
			return nil, fmt.Errorf("fetch carbon page: %w", err)
		}
		defer func() {
			if err := page.Close(); err != nil {
				slog.Error("err closing page", slog.String("submissionID", submissionID), slog.Any("err", err))
			}
		}()
		if err := page.WaitStable(200 * time.Millisecond); err != nil {
			return nil, fmt.Errorf("wait page stabilization: %w", err)
		}
		slog.Info("fetched page", slog.String("submissionID", submissionID))

		codeContainer, err := page.Element(".CodeMirror__container")
		if err != nil {
			return nil, fmt.Errorf("get code container: %w", err)
		}
		if err := codeContainer.WaitStable(500 * time.Millisecond); err != nil {
			return nil, fmt.Errorf("wait code container stabilization: %w", err)
		}
		// need to click to correctly apply code highlighting
		if err := codeContainer.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return nil, fmt.Errorf("click code container: %w", err)
		}
		time.Sleep(time.Second)
		slog.Info("selected code container", slog.String("submissionID", submissionID))

		exportMenu, err := page.Element(`#export-menu`)
		if err != nil {
			return nil, fmt.Errorf("get export menu: %w", err)
		}
		if err := exportMenu.WaitStable(500 * time.Millisecond); err != nil {
			return nil, fmt.Errorf("wait export menu stabilization: %w", err)
		}
		if err := exportMenu.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return nil, fmt.Errorf("click export menu: %w", err)
		}
		slog.Info("opened export menu", slog.String("submissionID", submissionID))

		exportBtns, err := page.Elements(".export-menu-container button")
		if err != nil {
			return nil, fmt.Errorf("get export buttons: %w", err)
		}
		exportBtns = iter.FilterMut(exportBtns, func(e *rod.Element) bool {
			return e.MustText() == "Open"
		})
		if len(exportBtns) != 1 {
			return nil, errors.New("open button not found or found multiple")
		}
		if err := exportBtns[0].Click(proto.InputMouseButtonLeft, 1); err != nil {
			return nil, fmt.Errorf("click export button: %w", err)
		}
		time.Sleep(time.Second)
		slog.Info("started exporting", slog.String("submissionID", submissionID))

		img, err := page.Element("body > img")
		if err != nil {
			return nil, fmt.Errorf("get img: %w", err)
		}
		src, err := img.Attribute("src")
		if err != nil {
			return nil, fmt.Errorf("get src: %w", err)
		}
		if src == nil {
			return nil, errors.New("src is nil")
		}
		slog.Info("located image", slog.String("submissionID", submissionID))

		buf, err := page.GetResource(*src)
		if err != nil {
			return nil, fmt.Errorf("get img data: %w", err)
		}
		if len(buf) == 0 {
			return nil, errors.New("img data is empty")
		}
		slog.Info("got image data", slog.String("submissionID", submissionID))

		return buf, nil
	})
}

func toRaysoLang(lang leetcode.Lang) string {
	switch lang {
	case leetcode.LangCPP:
		return "cpp"
	case leetcode.LangJava:
		return "java"
	case leetcode.LangPy2:
		return "python"
	case leetcode.LangPy3:
		return "python"
	case leetcode.LangC:
		return "cpp"
	case leetcode.LangCSharp:
		return "csharp"
	case leetcode.LangJS:
		return "javascript"
	case leetcode.LangTS:
		return "typescript"
	case leetcode.LangPHP:
		return "php"
	case leetcode.LangSwift:
		return "swift"
	case leetcode.LangKotlin:
		return "kotlin"
	case leetcode.LangGO:
		return "go"
	case leetcode.LangRuby:
		return "ruby"
	case leetcode.LangScala:
		return "scala"
	case leetcode.LangRust:
		return "rust"
	case leetcode.LangRacket:
		return "lisp"
	default:
		return ""
	}
}

func (g *Generator) GenerateCodeSnippetRayso(
	ctx context.Context,
	submissionID string,
	lang leetcode.Lang,
	code string,
) ([]byte, error) {
	code = normalizeCode(code)
	backoff := retry.LinearBackoff{
		Delay:       time.Second,
		MaxAttempts: 8,
	}
	return retry.Do(ctx, "rayso snippet "+submissionID, backoff, func() ([]byte, error) {
		slog.Info("start generate rayso code snippet", slog.String("submissionID", submissionID))
		uri := fmt.Sprintf(
			"%s/#theme=vercel&background=true&darkMode=true&padding=16&language=%s&code=%s",
			g.cfg.RaysoURL, toRaysoLang(lang), base64.URLEncoding.EncodeToString([]byte(code)),
		)
		page, err := g.browser.Timeout(30 * time.Second).Page(proto.TargetCreateTarget{URL: uri})
		if err != nil {
			return nil, fmt.Errorf("fetch rayso page: %w", err)
		}
		defer func() {
			if err := page.Close(); err != nil {
				slog.Error("err closing page", slog.String("submissionID", submissionID), slog.Any("err", err))
			}
		}()
		if err := page.WaitStable(300 * time.Millisecond); err != nil {
			return nil, fmt.Errorf("wait page stabilization: %w", err)
		}
		slog.Info("fetched page", slog.String("submissionID", submissionID))

		// TODO change size to 6x
		// padding via query just doesn't work for vercel theme
		paddingBt, err := page.Element(`div[dir="ltr"] > button[aria-label="16"]`)
		if err != nil {
			return nil, fmt.Errorf("get padding button: %w", err)
		}
		if err := paddingBt.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return nil, fmt.Errorf("click padding button: %w", err)
		}
		slog.Info("set padding", slog.String("submissionID", submissionID))

		wait := g.browser.WaitDownload(g.cfg.RodDownloadsSaveFolder)
		exportBtn, err := page.Element(`button[aria-label="Export as PNG"]`)
		if err != nil {
			return nil, fmt.Errorf("get export button: %w", err)
		}
		if err := exportBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return nil, fmt.Errorf("click export button: %w", err)
		}
		slog.Info("started exporting", slog.String("submissionID", submissionID))

		path := filepath.Join(g.cfg.RodDownloadsGetFolder, wait().GUID)
		defer func() {
			if err := os.Remove(path); err != nil {
				slog.Error("err removing tmp snippet", slog.String("path", path))
			}
		}()
		buf, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read snippet: %w", err)
		}
		if len(buf) == 0 {
			return nil, errors.New("snippet data is empty")
		}
		slog.Info("got snippet data", slog.String("submissionID", submissionID))

		return buf, nil
	})
}

func (g *Generator) GenerateCodeSnippet(
	ctx context.Context,
	submissionID string,
	lang leetcode.Lang,
	code string,
) ([]byte, error) {
	switch {
	case g.cfg.UseCarbon:
		return g.GenerateCodeSnippetCarbon(ctx, submissionID, lang, code)
	case g.cfg.UseRayso:
		return g.GenerateCodeSnippetRayso(ctx, submissionID, lang, code)
	case g.cfg.UseJavaHighlight:
		return g.GenerateCodeSnippetJavaHighlight(ctx, submissionID, lang, code)
	}

	return nil, errors.New("no preferred image generator enabled")
}
