package chrome

import (
	"context"
	"fmt"
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
	code string,
) ([]byte, error) {
	backoff := retry.LinearBackoff{
		Delay:       time.Second,
		MaxAttempts: 5,
	}
	return retry.Do(ctx, "generate code snippet", backoff, func() ([]byte, error) {
		page, err := browser.Page(proto.TargetCreateTarget{
			URL: "https://carbon.now.sh/?t=seti&es=4x&l=auto&code=" + url.QueryEscape(code),
		})
		if err != nil {
			return nil, fmt.Errorf("fecth carbon page: %w", err)
		}

		codeContainer, err := page.Element(".CodeMirror__container")
		if err != nil {
			return nil, fmt.Errorf("get code container: %w", err)
		}
		if err := codeContainer.WaitStable(500 * time.Millisecond); err != nil {
			return nil, fmt.Errorf("wait code container stabilization: %w", err)
		}

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

		img, err := page.Element("body > img")
		if err != nil {
			return nil, fmt.Errorf("get img: %w", err)
		}
		src, err := img.Attribute("src")
		if err != nil || src == nil {
			return nil, fmt.Errorf("get src: %w", err)
		}

		buf, err := page.GetResource(*src)
		if err != nil || len(buf) == 0 {
			return nil, fmt.Errorf("get img data: %w", err)
		}

		return buf, nil
	})
}
