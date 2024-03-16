package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

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
	ServiceConfig
	database db.DB
	telegram *tg.Client
}

func NewService(
	cfg ServiceConfig,
	telegram *tg.Client,
	database db.DB,
) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Service{
		ServiceConfig: cfg,
		database:      database,
		telegram:      telegram,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.telegram.RegisterHandler(newNeetCodeCounter(ctx, s.database), tele.OnText)
	s.telegram.RegisterHandler(ReactionHandler{})
	if s.MockEgor.Enabled {
		s.telegram.RegisterHandler(newMockEgorHandler(s.MockEgor, ctx, s.database, s.telegram))
	}
	s.telegram.Start()
	return s.database.Start(ctx)
}

func (s *Service) Stop() {
	s.telegram.Stop()
	s.database.Stop()
}

func (s *Service) publish(ctx context.Context, header, text, stickerID string, pinnedMsgIDKey string) error {
	err := s.database.Do(ctx, func(tx db.Tx) error {
		pinnedId, err := db.GetJsonDefault[int](tx, pinnedMsgIDKey, 0)
		if err != nil {
			return fmt.Errorf("get key %q: %w", pinnedMsgIDKey, err)
		}
		if pinnedId != 0 {
			err = s.telegram.Unpin(pinnedId)
			if err != nil {
				slog.Error("err unpin", slog.Any("err", err))
			}
		}

		messageID, err := s.telegram.SendSpoilerLink(s.LeetcodeThreadID, header, text)
		if err != nil {
			return fmt.Errorf("send daily: %w", err)
		}

		_, err = s.telegram.SendSticker(s.LeetcodeThreadID, stickerID)
		if err != nil {
			return fmt.Errorf("send sticker: %w", err)
		}

		err = s.telegram.Pin(messageID)
		if err != nil {
			return fmt.Errorf("pin: %w", err)
		}

		err = db.SetJson(tx, pinnedMsgIDKey, messageID)
		if err != nil {
			return fmt.Errorf("set key %s: %w", pinnedMsgIDKey, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("db update: %w", err)
	}

	return nil
}
