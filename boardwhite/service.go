package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/tg"
)

const (
	keyLastByegorMockTime = "boardwhite:mock:byegor:last_mock"

	keyLCPinnedMessages = "boardwhite:leetcode:pinned_messages"

	keyNCPinnedMessages = "boardwhite:neetcode:pinned_messages"
	keyNCStats          = "boardwhite:neetcode:stats"
)

var (
	lcSubmissionRe = regexp.MustCompile(`^https://leetcode.com/.+/submissions/[^/]+/?$`)
)

type MockEgorConfig struct {
	Enabled   bool
	Period    time.Duration
	StickerID string
}

type ServiceConfig struct {
	LeetcodeThreadID int
	DailyStickersIDs []string
	DpStickerID      string
	DailyNCStartDate time.Time
	MockEgor         MockEgorConfig
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
	ctx context.Context,
	header, text, stickerID string,
	pinnedMsgsKey string,
) error {
	return s.database.Do(ctx, func(tx db.Tx) error {
		pinnedIDs, err := db.GetJsonDefault[[]int](tx, pinnedMsgsKey, nil)
		if err != nil {
			return fmt.Errorf("get key %s: %w", pinnedMsgsKey, err)
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
			return fmt.Errorf("send daily: %w", err)
		}

		_, err = s.telegram.SendSticker(s.cfg.LeetcodeThreadID, stickerID)
		if err != nil {
			return fmt.Errorf("send sticker: %w", err)
		}

		err = s.telegram.Pin(messageID)
		if err != nil {
			return fmt.Errorf("pin: %w", err)
		}

		pinnedIDs = append(pinnedIDs, messageID)
		err = db.SetJson(tx, pinnedMsgsKey, pinnedIDs)
		if err != nil {
			return fmt.Errorf("set key %s: %w", pinnedMsgsKey, err)
		}

		return nil
	})
}
