package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/boar-d-white-foundation/drone/iterx"
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

	Features struct {
		RodEnabled bool `yaml:"rod_enabled"`
	} `yaml:"features"`

	Tg struct {
		Key               string        `yaml:"api_key" json:"-"` // intentionally hidden from logs
		LongPollerTimeout time.Duration `yaml:"long_poller_timeout"`
		AdminChatID       int64         `yaml:"admin_chat_id"`
	} `yaml:"tg"`

	Rod struct {
		Host            string `yaml:"host"`
		Port            int    `yaml:"port"`
		DownloadsFolder string `yaml:"downloads_folder"`
	} `yaml:"rod"`

	MediaGenerator struct {
		JavaHighlightURL   string `yaml:"javahighlight_url"`
		RodDownloadsFolder string `yaml:"rod_downloads_folder"`
	} `yaml:"media_generator"`

	VC struct {
		Domain string `yaml:"domain"`
		Token  string `yaml:"token" json:"-"` // intentionally hidden from logs
	} `yaml:"vc"`

	Boardwhite struct {
		ChatID                   int64 `yaml:"chat_id"`
		LeetCodeThreadID         int   `yaml:"leetcode_thread_id"`
		LeetcodeChickensThreadID int   `yaml:"leetcode_chickens_thread_id"`
		FloodThreadID            int   `yaml:"flood_thread_id"`
		InterviewsThreadID       int   `yaml:"interviews_thread_id"`
	} `yaml:"boardwhite"`

	Leetcode struct {
		Session string `yaml:"session" json:"-"` // intentionally hidden from logs
		CSRF    string `yaml:"csrf" json:"-"`    // intentionally hidden from logs
	} `yaml:"leetcode"`

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

	GreetingsNewUsersTemplates []string `yaml:"greetings_new_users_templates"`
	GreetingsOldUsersTemplates []string `yaml:"greetings_old_users_templates"`

	Oborona struct {
		Period   time.Duration `yaml:"period"`
		Template string        `yaml:"template"`
		Words    [][]string    `yaml:"words"`
	} `yaml:"oborona"`
}

func (cfg Config) String() string {
	b, _ := json.Marshal(&cfg) //nolint:errchkjson // intentionally omitting the error
	return string(b)
}

func Default() (cfg Config, err error) {
	err = yaml.Unmarshal(defaultConfigBytes, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal default config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
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

	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	if !slices.Equal(cfg.DailyStickerIDs, iterx.Uniq(cfg.DailyStickerIDs)) {
		return errors.New("all daily_sticker_ids must be unique")
	}
	if !slices.Equal(cfg.DailyChickensStickerIDs, iterx.Uniq(cfg.DailyChickensStickerIDs)) {
		return errors.New("all daily_chickens_sticker_ids must be unique")
	}
	for _, mock := range cfg.Mocks {
		if !slices.Equal(mock.StickerIDs, iterx.Uniq(mock.StickerIDs)) {
			return errors.New("all mocks.sticker_ids must be unique")
		}
	}

	if len(cfg.GreetingsNewUsersTemplates) == 0 {
		return errors.New("greetings_new_users_templates must not be empty")
	}
	if len(cfg.GreetingsOldUsersTemplates) == 0 {
		return errors.New("greetings_old_users_templates must not be empty")
	}

	if cfg.Oborona.Template == "" {
		return errors.New("oborona.template must not be empty")
	}
	if len(cfg.Oborona.Words) == 0 {
		return errors.New("oborona.words must not be empty")
	}
	for _, words := range cfg.Oborona.Words {
		if len(words) == 0 {
			return errors.New("oborona.words.[*] must not be empty")
		}
	}
	if strings.Count(cfg.Oborona.Template, "%s") != len(cfg.Oborona.Words) {
		return errors.New("oborona.template must have the same number of %s as there are words in oborona.words")
	}

	return nil
}
