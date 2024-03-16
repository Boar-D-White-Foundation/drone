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

type MockEgorConfig struct {
	Enabled   bool
	Period    time.Duration
	StickerID string
}

func (s *Service) newMockEgorHandler(ctx context.Context, cfg MockEgorConfig) tele.HandlerFunc {

	return func(c tele.Context) error {

		if c.Sender().Username != byegor {
			return nil
		}

		message := c.Message()

		key := "mock:byegor"

		return s.database.Do(ctx, func(tx db.Tx) error {
			shouldMock := true
			t, err := db.GetJson[time.Time](tx, key)
			switch {
			case err == nil:
				if time.Since(t) < cfg.Period {
					shouldMock = false
				}
			case errors.Is(err, db.ErrKeyNotFound):
				// still do it
			default:
				return fmt.Errorf("get key %q: %w", key, err)
			}

			if !shouldMock {
				return nil
			}

			_, err = s.telegram.ReplyWithSticker(cfg.StickerID, message)
			if err != nil {
				return fmt.Errorf("reply with sticker: %w", err)
			}

			err = db.SetJson[time.Time](tx, key, time.Now())

			if err != nil {
				return fmt.Errorf("set key %q: %w", key, err)
			}

			return nil
		})
	}
}
