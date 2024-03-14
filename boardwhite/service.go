package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/tg"
)

type Service struct {
	leetcodeThreadID int
	dailyStickersIDs []string
	dpStickerID      string
	dailyNCStartDate time.Time
	database         db.DB
	telegram         *tg.Client
}

func NewService(
	leetcodeThreadID int,
	dailyStickersIDs []string,
	dpStickerID string,
	dailyNCStartDate time.Time,
	telegram *tg.Client,
	database db.DB,
) (*Service, error) {
	if dailyNCStartDate.After(time.Now()) {
		return nil, errors.New("dailyNCStartDate should be in past")
	}
	return &Service{
		leetcodeThreadID: leetcodeThreadID,
		dailyStickersIDs: dailyStickersIDs,
		dpStickerID:      dpStickerID,
		dailyNCStartDate: dailyNCStartDate,
		database:         database,
		telegram:         telegram,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	s.telegram.RegisterHandler(NeetCodeCounter{db: s.database})
	s.telegram.RegisterHandler(ReactionHandler{})
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

		messageID, err := s.telegram.SendSpoilerLink(s.leetcodeThreadID, header, text)
		if err != nil {
			return fmt.Errorf("send daily: %w", err)
		}

		_, err = s.telegram.SendSticker(s.leetcodeThreadID, stickerID)
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
