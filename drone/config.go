package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/boardwhite"
	"github.com/kelseyhightower/envconfig"
)

func defaultDailyStickersIDs() []string {
	return []string{
		"CAACAgIAAxkBAAELpiFl7G4Rn8WQBK3AaDiAMn6ixTUR7gACzzkAAr-TAAFK91qMnVpp9TQ0BA",
		"CAACAgIAAxkBAAELqk5l72yJkbx4_vskG3n6zWoWaAnA3QACazYAArJX2UsY5inoNwaFoTQE",
		"CAACAgIAAxkBAAELsFZl8hFWX-2VvyppubPYPnLFrVeLPgACZjEAAmSSoUrAsMs8kC1omDQE",
		"CAACAgIAAxkBAAELsFRl8hE8RstoYI8E3RfKk2N1LRX-qQAC4TUAAhJs2ElFZyKM69wJfDQE",
		"CAACAgIAAxkBAAELsFJl8hEuJlU_1RK4nbpq3_iuZYsAATAAAi01AALtZihJuuf3iNlNcbU0BA",
		"CAACAgIAAxkBAAELsFBl8hEnQ9Ie_ANBN_TfzbDjvMck2wACVToAAhqdoEiRsQbJLo-jijQE",
		"CAACAgIAAxkBAAELsE5l8hEfhr-0Dn7BfsI9sDWA3I0t0wACSzUAAtS3aUj_OJb99K6DkjQE",
		"CAACAgIAAxkBAAELsExl8hEZTP0DFsf6hUDiZnx-i9I25wACpTEAAlp3YUhJNRORPRx_5DQE",
		"CAACAgIAAxkBAAELsEpl8hERg5mpamVfTo_SCgZRatbi6wACCzcAAqVsKEjLAvd1EuqdLzQE",
		"CAACAgIAAxkBAAELsEhl8hENAzygO8iFauBYZD0XPYqD3gACkTQAArnoCUjvaNgUd-BoHDQE",
		"CAACAgIAAxkBAAELsEZl8hD-i6wYaeLUMtP8MWhZwZoy3gACAjMAAs0gAUjsT0apWa4cRTQE",
		"CAACAgIAAxkBAAELsERl8hD1xsuLUYRD9F4a8ekVAgg8VAACgDgAAsuH-UvwActGs5DfMzQE",
		"CAACAgIAAxkBAAELsEJl8hDUc-b0jyDfeH6Ct2McMp4mlAACOzcAAquEmUsMP7ObsCcumTQE",
	}
}

type Config struct {
	BadgerPath string `envconfig:"BADGER_PATH" default:"data/badger"`

	TgKey               string        `envconfig:"TG_BOT_API_KEY"`
	TgLongPollerTimeout time.Duration `envconfig:"TG_LONG_POLLER_TIMEOUT" default:"10s"`

	BoarDWhiteChatID           int64 `envconfig:"BOAR_D_WHITE_CHAT_ID" default:"-1001640461540"`
	BoarDWhiteLeetCodeThreadID int   `envconfig:"BOAR_D_WHITE_LEET_CODE_THREAD_ID" default:"10095"`

	// every day at 01:00 UTC
	LCDailyCron string `envconfig:"LC_DAILY_CRON" default:"0 1 * * *"`

	// every day at 12:00 UTC
	NCDailyCron      string `envconfig:"NC_DAILY_CRON" default:"0 12 * * *"`
	NCDailyStartDate string `envconfig:"NC_DAILY_START_DATE" default:"2024-03-14"`
	DailyStickerIDs  []string
	DPStickerID      string `envconfig:"DP_STICKER_ID" default:"CAACAgIAAxkBAAELsFhl8hGciGxpkKi4-jhou97SOqwkvwACpT0AAkV6sEpqc1XNPnvOIDQE"`

	Mocks MockConfigs `envconfig:"MOCKS"`
}

type MockConfig struct {
	Username  string
	Period    string
	StickerID string
}

type MockConfigs []MockConfig

func (cfg *MockConfigs) Decode(s string) error {
	if s == "" {
		return nil
	}
	s = strings.TrimLeft(s, "[")
	s = strings.TrimRight(s, "]")
	ss := strings.Split(s, ",")

	cfgs := make([]MockConfig, 0, len(ss))
	for _, v := range ss {
		vv := strings.Split(v, ";")
		if len(vv) != 3 {
			return fmt.Errorf("incorrect mock config value %q", v)
		}
		cfgs = append(cfgs, MockConfig{
			Username:  vv[0],
			Period:    vv[1],
			StickerID: vv[2],
		})
	}

	*cfg = cfgs
	return nil
}

func (cfg Config) ServiceConfig() (boardwhite.ServiceConfig, error) {
	ncDailyStartDate, err := time.Parse("2006-01-02", cfg.NCDailyStartDate)
	if err != nil {
		return boardwhite.ServiceConfig{}, fmt.Errorf("parse ncdailystartdate %q: %w", cfg.NCDailyStartDate, err)
	}

	mocks := make(map[string]boardwhite.MockConfig)
	for _, v := range cfg.Mocks {
		period, err := time.ParseDuration(v.Period)
		if err != nil {
			return boardwhite.ServiceConfig{}, fmt.Errorf("parse duration %q: %w", v.Period, err)
		}
		mocks[v.Username] = boardwhite.MockConfig{
			Period:    period,
			StickerID: v.StickerID,
		}
	}

	return boardwhite.ServiceConfig{
		LeetcodeThreadID: cfg.BoarDWhiteLeetCodeThreadID,
		DailyStickersIDs: cfg.DailyStickerIDs,
		DpStickerID:      cfg.DPStickerID,
		DailyNCStartDate: ncDailyStartDate,
		Mocks:            mocks,
	}, nil
}

func LoadConfig() (Config, error) {
	var cfg Config

	err := envconfig.Process("DRONE", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("process: %w", err)
	}

	if len(cfg.DailyStickerIDs) == 0 {
		cfg.DailyStickerIDs = defaultDailyStickersIDs()
	}

	return cfg, nil
}
