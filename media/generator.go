package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/boar-d-white-foundation/drone/config"
	"github.com/boar-d-white-foundation/drone/leetcode"
	"github.com/boar-d-white-foundation/drone/retry"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type GeneratorConfig struct {
	JavaHighlightURL       string
	RodDownloadsSaveFolder string
	RodDownloadsGetFolder  string
	VCDomain               string
	VCToken                string
}

type Generator struct {
	mu      sync.Mutex
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
			JavaHighlightURL:       cfg.MediaGenerator.JavaHighlightURL,
			RodDownloadsSaveFolder: cfg.Rod.DownloadsFolder,
			RodDownloadsGetFolder:  cfg.MediaGenerator.RodDownloadsFolder,
			VCDomain:               cfg.VC.Domain,
			VCToken:                cfg.VC.Token,
		},
		browser,
	)
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
	case leetcode.LangDart:
		return "dart"
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
	case leetcode.LangElixir:
		return "elixir"
	default:
		return ""
	}
}

func (g *Generator) GenerateCodeSnippet(
	ctx context.Context,
	id leetcode.SubmissionID,
	lang leetcode.Lang,
	code string,
) ([]byte, error) {
	code = normalizeCode(code)
	backoff := retry.LinearBackoff{
		Delay:       time.Second,
		MaxAttempts: 2,
	}
	return retry.Do(ctx, "java highlight snippet "+id.String(), backoff, func() ([]byte, error) {
		slog.Info("start generate code snippet", slog.String("submissionID", id.String()))
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
				slog.Error(
					"err close resp body",
					slog.String("submissionID", id.String()), slog.Any("err", err),
				)
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

func (g *Generator) GenerateVCPagePdf(
	ctx context.Context,
	link string,
) ([]byte, error) {
	backoff := retry.LinearBackoff{
		Delay:       time.Second,
		MaxAttempts: 8,
	}
	return retry.Do(ctx, "vc page pdf "+link, backoff, func() ([]byte, error) {
		slog.Info("start generate vc page pdf", slog.String("link", link))
		err := g.browser.SetCookies([]*proto.NetworkCookieParam{{
			Name:   "token",
			Value:  g.cfg.VCToken,
			Domain: g.cfg.VCDomain,
		}})
		if err != nil {
			return nil, fmt.Errorf("set auth cookies: %w", err)
		}
		time.Sleep(time.Second)

		page, err := g.browser.Timeout(30 * time.Second).Page(proto.TargetCreateTarget{URL: link})
		if err != nil {
			return nil, fmt.Errorf("fetch vc page: %w", err)
		}
		defer func() {
			if err := page.Close(); err != nil {
				slog.Error("err closing page", slog.String("link", link), slog.Any("err", err))
			}
		}()
		if err := page.WaitStable(300 * time.Millisecond); err != nil {
			return nil, fmt.Errorf("wait page stabilization: %w", err)
		}
		slog.Info("fetched page", slog.String("link", link))

		menuRight, err := page.Element(`.menu-right`)
		if err != nil {
			return nil, fmt.Errorf("get menu right: %w", err)
		}
		if err := menuRight.Remove(); err != nil {
			return nil, fmt.Errorf("remove menu right: %w", err)
		}
		slog.Info("removed menu right", slog.String("link", link))

		commentsForm, err := page.Element(`#post-comments-form`)
		if err != nil {
			return nil, fmt.Errorf("get comments form: %w", err)
		}
		if err := commentsForm.Remove(); err != nil {
			return nil, fmt.Errorf("remove comments form: %w", err)
		}
		slog.Info("removed comments form", slog.String("link", link))

		slog.Info("start pdf generation", slog.String("link", link))
		reader, err := page.PDF(&proto.PagePrintToPDF{})
		if err != nil {
			return nil, fmt.Errorf("get pdf: %w", err)
		}
		slog.Info("generate pdf ok", slog.String("link", link))

		return io.ReadAll(reader)
	})
}

func (g *Generator) downloadFile(
	ctx context.Context,
	initiateDownload func(context.Context) error,
) ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	wait := g.browser.WaitDownload(g.cfg.RodDownloadsSaveFolder)
	if err := initiateDownload(ctx); err != nil {
		return nil, fmt.Errorf("initiate download: %w", err)
	}

	path := filepath.Join(g.cfg.RodDownloadsGetFolder, wait().GUID)
	defer func() {
		if err := os.Remove(path); err != nil {
			slog.Error("err removing tmp file", slog.String("path", path))
		}
	}()
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read downloaded file: %w", err)
	}
	if len(buf) == 0 {
		return nil, errors.New("downloaded file data is empty")
	}
	slog.Info("got downloaded file data")

	return buf, nil
}
