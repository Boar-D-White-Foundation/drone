package boardwhite

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/tg"
)

const (
	keyLCPinnedMessages = "boardwhite:leetcode:pinned_messages"

	keyNCPinnedMessages = "boardwhite:neetcode:pinned_messages"
	keyNCPinnedToDayIdx = "boardwhite:neetcode:pinned_to_day_idx"
	keyNCStats          = "boardwhite:neetcode:stats"
)

var (
	lcSubmissionRe = regexp.MustCompile(`https://leetcode.com/.+/submissions/[^/]+/?`)
)

type MockConfig struct {
	Period     time.Duration
	StickerIDs []string
}

type ServiceConfig struct {
	LeetcodeThreadID int
	DailyStickersIDs []string
	DpStickerID      string
	DailyNCStartDate time.Time
	Mocks            map[string]MockConfig
}

func (cfg ServiceConfig) Validate() error {
	if cfg.DailyNCStartDate.After(time.Now()) {
		return errors.New("dailyNCStartDate should be in past")
	}

	return nil
}

type Service struct {
	cfg      ServiceConfig
	database db.DB
	telegram tg.Client
}

func NewService(
	cfg ServiceConfig,
	telegram tg.Client,
	database db.DB,
) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Service{
		cfg:      cfg,
		database: database,
		telegram: telegram,
	}, nil
}

func (s *Service) publish(
	tx db.Tx,
	header, text, stickerID string,
	pinnedMsgsKey string,
) (int, error) {
	pinnedIDs, err := db.GetJsonDefault[[]int](tx, pinnedMsgsKey, nil)
	if err != nil {
		return 0, fmt.Errorf("get key %s: %w", pinnedMsgsKey, err)
	}
	if len(pinnedIDs) > 0 {
		// last is considered active
		err = s.telegram.Unpin(pinnedIDs[len(pinnedIDs)-1])
		if err != nil {
			slog.Error("err unpin", slog.Any("err", err))
		}
	}

	messageID, err := s.telegram.SendSpoilerLink(s.cfg.LeetcodeThreadID, header, text)
	if err != nil {
		return 0, fmt.Errorf("send daily: %w", err)
	}

	_, err = s.telegram.SendSticker(s.cfg.LeetcodeThreadID, stickerID)
	if err != nil {
		return 0, fmt.Errorf("send sticker: %w", err)
	}

	err = s.telegram.Pin(messageID)
	if err != nil {
		return 0, fmt.Errorf("pin: %w", err)
	}

	pinnedIDs = append(pinnedIDs, messageID)
	err = db.SetJson(tx, pinnedMsgsKey, pinnedIDs)
	if err != nil {
		return 0, fmt.Errorf("set key %s: %w", pinnedMsgsKey, err)
	}

	return messageID, nil
}
