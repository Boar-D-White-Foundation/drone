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
	leetcodeThreadID int
	dailyStickersIDs []string
	dpStickerID      string
	dailyNCStartDate time.Time
	mocks            map[string]MockConfig
}

func NewServiceConfig(
	leetcodeThreadID int,
	dailyStickersIDs []string,
	dpStickerID string,
	dailyNCStartDate time.Time,
	mocks map[string]MockConfig,
) (ServiceConfig, error) {
	if dailyNCStartDate.After(time.Now()) {
		return ServiceConfig{}, errors.New("dailyNCStartDate should be in past")
	}

	return ServiceConfig{
		leetcodeThreadID: leetcodeThreadID,
		dailyStickersIDs: dailyStickersIDs,
		dpStickerID:      dpStickerID,
		dailyNCStartDate: dailyNCStartDate,
		mocks:            mocks,
	}, nil
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
) *Service {
	return &Service{
		cfg:      cfg,
		database: database,
		telegram: telegram,
	}
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

	messageID, err := s.telegram.SendSpoilerLink(s.cfg.leetcodeThreadID, header, text)
	if err != nil {
		return 0, fmt.Errorf("send daily: %w", err)
	}

	_, err = s.telegram.SendSticker(s.cfg.leetcodeThreadID, stickerID)
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
