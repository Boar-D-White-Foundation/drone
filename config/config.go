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
	BadgerPath string `yaml:"badgerPath"`

	Tg struct {
		Key               string        `yaml:"apiKey" json:"-"` // intentionally hidden from logs
		LongPollerTimeout time.Duration `yaml:"longPollerTimeout"`
	} `yaml:"tg"`

	Boardwhite struct {
		ChatID                   int64 `yaml:"chatId"`
		LeetCodeThreadID         int   `yaml:"leetcodeThreadId"`
		LeetcodeChickensThreadID int   `yaml:"leetcodeChickensThreadId"`
	} `yaml:"boardwhite"`

	LeetcodeDaily struct {
		Cron       string `yaml:"cron"`
		RatingCron string `yaml:"ratingCron"`
	} `yaml:"leetcodeDaily"`

	NeetcodeDaily struct {
		Cron       string `yaml:"cron"`
		RatingCron string `yaml:"ratingCron"`
	} `yaml:"neetcodeDaily"`

	DailyStickerIDs         []string `yaml:"dailyStickerIds"`
	DailyChickensStickerIDs []string `yaml:"dailyChickensStickerIds"`
	DPStickerID             string   `yaml:"dpStickerId"`

	Mocks []struct {
		Username   string   `yaml:"username"`
		Period     string   `yaml:"period"`
		StickerIDs []string `yaml:"stickerIds"`
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
