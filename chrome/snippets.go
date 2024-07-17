package chrome

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
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
	UseCarbon              bool
	UseRayso               bool
	RodDownloadsSaveFolder string
	RodDownloadsGetFolder  string
}

type ImageGenerator struct {
	cfg     GeneratorConfig
	browser *rod.Browser
}

func NewImageGenerator(
	cfg GeneratorConfig,
	browser *rod.Browser,
) *ImageGenerator {
	return &ImageGenerator{
		cfg:     cfg,
		browser: browser,
	}
}

func NewImageGeneratorFromCfg(
	cfg config.Config,
	browser *rod.Browser,
) *ImageGenerator {
	return NewImageGenerator(
		GeneratorConfig{
			CarbonURL:              cfg.ImageGenerator.CarbonURL,
			RaysoURL:               cfg.ImageGenerator.RaysoURL,
			UseCarbon:              cfg.ImageGenerator.UseCarbon,
			UseRayso:               cfg.ImageGenerator.UseRayso,
			RodDownloadsSaveFolder: cfg.Rod.DownloadsFolder,
			RodDownloadsGetFolder:  cfg.ImageGenerator.RodDownloadsFolder,
		},
		browser,
	)
}

// WarmUpCaches generates images from all sources to correctly warm up fonts caches
func (g *ImageGenerator) WarmUpCaches(ctx context.Context) error {
	if _, err := g.GenerateCodeSnippetCarbon(ctx, "warmup", leetcode.LangUnknown, "warmup"); err != nil {
		return fmt.Errorf("warmup carbon: %w", err)
	}

	if _, err := g.GenerateCodeSnippetRayso(ctx, "warmup", leetcode.LangUnknown, "warmup"); err != nil {
		return fmt.Errorf("warmup rayso: %w", err)
	}

	return nil
}

func (g *ImageGenerator) GenerateCodeSnippetCarbon(
	ctx context.Context,
	submissionID string,
	lang leetcode.Lang,
	code string,
) ([]byte, error) {
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

func (g *ImageGenerator) GenerateCodeSnippetRayso(
	ctx context.Context,
	submissionID string,
	lang leetcode.Lang,
	code string,
) ([]byte, error) {
	backoff := retry.LinearBackoff{
		Delay:       time.Second,
		MaxAttempts: 8,
	}
	return retry.Do(ctx, "rayso snippet "+submissionID, backoff, func() ([]byte, error) {
		slog.Info("start generate rayso code snippet", slog.String("submissionID", submissionID))
		uri := fmt.Sprintf(
			"%s/#theme=candy&background=true&darkMode=true&padding=16&language=%s&code=%s",
			g.cfg.RaysoURL, toRaysoLang(lang), base64.URLEncoding.EncodeToString([]byte(code)),
		)
		page, err := g.browser.Timeout(30 * time.Second).Page(proto.TargetCreateTarget{URL: uri})
		if err != nil {
			return nil, fmt.Errorf("fetch rayso page: %w", err)
		}
		if err := page.WaitStable(300 * time.Millisecond); err != nil {
			return nil, fmt.Errorf("wait page stabilization: %w", err)
		}
		slog.Info("fetched page", slog.String("submissionID", submissionID))

		// TODO change size to 6x
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

func (g *ImageGenerator) GenerateCodeSnippet(
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
	}

	return nil, errors.New("no preferred image generator enabled")
}
