package chrome

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/retry"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pkg/errors"
)

func GenerateCodeSnippet(
	ctx context.Context,
	browser *rod.Browser,
	submissionID string,
	code string,
) ([]byte, error) {
	backoff := retry.LinearBackoff{
		Delay:       time.Second,
		MaxAttempts: 8,
	}
	return retry.Do(ctx, "generate code snippet "+submissionID, backoff, func() ([]byte, error) {
		slog.Info("start generate code snippet", slog.String("submissionID", submissionID))
		page, err := browser.Page(proto.TargetCreateTarget{
			URL: "http://carbon:3000/?t=vscode&es=4x&l=auto&ln=false&fm=Hack&code=" + url.QueryEscape(code),
		})
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
