package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	tele "gopkg.in/telebot.v3"
)

const (
	byegor = "byegor"
)

func (s *Service) OnMockEgor(ctx context.Context, c tele.Context) error {
	msg, sender := c.Message(), c.Sender()
	if msg == nil || sender == nil {
		return nil
	}
	if sender.Username != byegor {
		return nil
	}

	key := keyLastByegorMockTime
	return s.database.Do(ctx, func(tx db.Tx) error {
		lastMock, err := db.GetJson[time.Time](tx, key)
		switch {
		case err == nil:
			if time.Since(lastMock) < s.cfg.MockEgor.Period {
				return nil
			}
		case errors.Is(err, db.ErrKeyNotFound):
			// still do it
		default:
			return fmt.Errorf("get lastMock: %w", err)
		}

		_, err = s.telegram.ReplyWithSticker(msg.ID, s.cfg.MockEgor.StickerID)
		if err != nil {
			return fmt.Errorf("reply with sticker: %w", err)
		}

		err = db.SetJson[time.Time](tx, key, time.Now())
		if err != nil {
			return fmt.Errorf("set lastMock: %w", err)
		}

		return nil
	})
}
