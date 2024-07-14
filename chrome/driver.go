package chrome

import (
	"fmt"
	"log/slog"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func NewRemote(host string, port int) (*rod.Browser, func(), error) {
	// docker run --rm -p 7317:7317 ghcr.io/go-rod/rod:v0.116.1
	lnch, err := launcher.NewManaged(fmt.Sprintf("ws://%s:%d", host, port))
	if err != nil {
		return nil, nil, fmt.Errorf("create launcher: %w", err)
	}

	lnch = lnch.Headless(false).
		Leakless(true).
		XVFB("--server-num=5", "--server-args=-screen 0 1600x900x16")

	client, err := lnch.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("create cdp client: %w", err)
	}

	browser := rod.New().Client(client)
	if err := browser.Connect(); err != nil {
		return nil, nil, fmt.Errorf("connect browser: %w", err)
	}

	cleanup := func() {
		if err := browser.Close(); err != nil {
			slog.Error("err close browser", slog.Any("err", err))
		}
	}
	return browser, cleanup, nil
}

func NewLocal() (*rod.Browser, func(), error) {
	lnch := launcher.New().Headless(false).Leakless(true)
	url, err := lnch.Launch()
	if err != nil {
		return nil, nil, fmt.Errorf("create launcher: %w", err)
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		lnch.Kill()
		lnch.Cleanup()
		return nil, nil, fmt.Errorf("connect browser: %w", err)
	}

	cleanup := func() {
		if err := browser.Close(); err != nil {
			slog.Error("err close browser", slog.Any("err", err))
		}
		lnch.Cleanup()
	}
	return browser, cleanup, nil
}
