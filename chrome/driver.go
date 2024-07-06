package chrome

import (
	"fmt"
	"log/slog"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type DriverConfig struct {
	Host string
	Port int
}

type Driver struct {
	browser *rod.Browser
}

func NewDriver(cfg DriverConfig) (*Driver, error) {
	lnch, err := launcher.NewManaged(fmt.Sprintf("ws://%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("create launcher: %w", err)
	}

	lnch = lnch.Headless(false).
		Leakless(true).
		XVFB("--server-num=5", "--server-args=-screen 0 1600x900x16")

	client, err := lnch.Client()
	if err != nil {
		return nil, fmt.Errorf("create cdp client: %w", err)
	}

	browser := rod.New().Client(client)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect browser: %w", err)
	}

	return &Driver{browser: browser}, nil
}

func (d *Driver) Close() {
	if err := d.browser.Close(); err != nil {
		slog.Error("err close chrome browser", slog.Any("err", err))
	}
}

func (d *Driver) Browser() *rod.Browser {
	return d.browser
}
