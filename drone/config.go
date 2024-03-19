package main

import (
	"fmt"
	"os"
	"time"

	_ "embed"

	"github.com/boar-d-white-foundation/drone/boardwhite"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

//go:embed default_config.yaml
var defaultConfigBytes []byte

type Config struct {
	BadgerPath string `yaml:"badgerPath"`

	Tg struct {
		Key               string        `envconfig:"BOT_API_KEY" yaml:"-" json:"-"`
		LongPollerTimeout time.Duration `envconfig:"-" yaml:"longPollerTimeout"`
	} `envconfig:"TG" yaml:"tg"`

	Boardwhite struct {
		ChatID           int64 `envconfig:"-" yaml:"chatId"`
		LeetCodeThreadID int   `envconfig:"-" yaml:"leetcodeThreadId"`
	} `envconfig:"-" yaml:"boardwhite"`

	LeetcodeDaily struct {
		Cron string `envconfig:"-" yaml:"cron"`
	} `envconfig:"-" yaml:"leetcodeDaily"`

	NeetcodeDaily struct {
		Cron      string `envconfig:"-" yaml:"cron"`
		StartDate string `envconfig:"-" yaml:"startDate"`
	} `envconfig:"-" yaml:"neetcodeDaily"`

	DailyStickerIDs []string `envconfig:"-" yaml:"dailyStickerIds"`
	DPStickerID     string   `envconfig:"-" yaml:"dpStickerId"`

	Mocks []struct {
		Username  string `envconfig:"-" yaml:"username"`
		Period    string `envconfig:"-" yaml:"period"`
		StickerID string `envconfig:"-" yaml:"stickerId"`
	} `envconfig:"-" yaml:"mocks"`
}

func (cfg Config) ServiceConfig() (boardwhite.ServiceConfig, error) {
	ncDailyStartDate, err := time.Parse("2006-01-02", cfg.NeetcodeDaily.StartDate)
	if err != nil {
		return boardwhite.ServiceConfig{}, fmt.Errorf("parse ncdailystartdate %q: %w", cfg.NeetcodeDaily.StartDate, err)
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
		LeetcodeThreadID: cfg.Boardwhite.LeetCodeThreadID,
		DailyStickersIDs: cfg.DailyStickerIDs,
		DpStickerID:      cfg.DPStickerID,
		DailyNCStartDate: ncDailyStartDate,
		Mocks:            mocks,
	}, nil
}

func DefaultConfig() (cfg Config, err error) {
	err = yaml.Unmarshal(defaultConfigBytes, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal default config: %w", err)
	}

	return cfg, nil
}

func LoadConfig(filename string) (Config, error) {
	cfg, err := DefaultConfig()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", filename, err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	err = envconfig.Process("DRONE", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("read envs: %w", err)
	}

	return cfg, nil
}
