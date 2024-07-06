package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/boar-d-white-foundation/drone/boardwhite"
	"gopkg.in/yaml.v3"
)

func Path() string {
	path, ok := os.LookupEnv("DRONE_CONFIG_FILENAME")
	if !ok {
		return "config.yaml"
	}

	return path
}

//go:embed default_config.yaml
var defaultConfigBytes []byte

type Config struct {
	BadgerPath string `yaml:"badger_path"`

	Tg struct {
		Key               string        `yaml:"api_key" json:"-"` // intentionally hidden from logs
		LongPollerTimeout time.Duration `yaml:"long_poller_timeout"`
	} `yaml:"tg"`

	Boardwhite struct {
		ChatID                   int64 `yaml:"chat_id"`
		LeetCodeThreadID         int   `yaml:"leetcode_thread_id"`
		LeetcodeChickensThreadID int   `yaml:"leetcode_chickens_thread_id"`
	} `yaml:"boardwhite"`

	LeetcodeDaily struct {
		Cron       string `yaml:"cron"`
		RatingCron string `yaml:"rating_cron"`
	} `yaml:"leetcode_daily"`

	NeetcodeDaily struct {
		Cron       string `yaml:"cron"`
		RatingCron string `yaml:"rating_cron"`
	} `yaml:"neetcode_daily"`

	DailyStickerIDs         []string `yaml:"daily_sticker_ids"`
	DailyChickensStickerIDs []string `yaml:"daily_chickens_sticker_ids"`
	DPStickerID             string   `yaml:"dp_sticker_id"`

	Mocks []struct {
		Username   string   `yaml:"username"`
		Period     string   `yaml:"period"`
		StickerIDs []string `yaml:"sticker_ids"`
	} `yaml:"mocks"`
}

func (cfg Config) String() string {
	b, _ := json.Marshal(&cfg) //nolint:errchkjson // intentionally omitting the error
	return string(b)
}

func (cfg Config) ServiceConfig() (boardwhite.ServiceConfig, error) {
	mocks := make(map[string]boardwhite.MockConfig)
	for _, v := range cfg.Mocks {
		period, err := time.ParseDuration(v.Period)
		if err != nil {
			return boardwhite.ServiceConfig{}, fmt.Errorf("parse duration %q: %w", v.Period, err)
		}
		mocks[v.Username] = boardwhite.MockConfig{
			Period:     period,
			StickerIDs: v.StickerIDs,
		}
	}

	return boardwhite.ServiceConfig{
		LeetcodeThreadID:         cfg.Boardwhite.LeetCodeThreadID,
		LeetcodeChickensThreadID: cfg.Boardwhite.LeetcodeChickensThreadID,
		DailyStickersIDs:         cfg.DailyStickerIDs,
		DailyChickensStickerIDs:  cfg.DailyChickensStickerIDs,
		DpStickerID:              cfg.DPStickerID,
		Mocks:                    mocks,
	}, nil
}

func Default() (cfg Config, err error) {
	err = yaml.Unmarshal(defaultConfigBytes, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal default config: %w", err)
	}

	return cfg, nil
}

func Load(filename string) (Config, error) {
	cfg, err := Default()
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

	return cfg, nil
}
