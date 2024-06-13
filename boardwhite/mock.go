package boardwhite

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iter"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) mockKey(username string) string {
	return fmt.Sprintf("boardwhite:mock:%s:next", username)
}

func (s *Service) OnMock(ctx context.Context, c tele.Context) error {
	msg, sender := c.Message(), c.Sender()
	if msg == nil || sender == nil {
		return nil
	}

	username := sender.Username
	cfg, ok := s.cfg.Mocks[username]
	if !ok {
		return nil
	}

	key := s.mockKey(username)
	return s.database.Do(ctx, func(tx db.Tx) error {
		mockAt, err := db.GetJson[time.Time](tx, key)
		switch {
		case err == nil:
			if mockAt.After(time.Now()) {
				// not yet
				return nil
			}
		case errors.Is(err, db.ErrKeyNotFound):
			// still do it
		default:
			return fmt.Errorf("get %q: %w", key, err)
		}

		stickerID, err := iter.PickRandom(cfg.StickerIDs)
		if err != nil {
			return fmt.Errorf("pick random sticker: %w", err)
		}

		_, err = s.telegram.ReplyWithSticker(msg.ID, stickerID)
		if err != nil {
			return fmt.Errorf("reply with sticker: %w", err)
		}

		from := cfg.Period.Seconds() * 0.8
		delta := cfg.Period.Seconds() * 0.4
		offset, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
		if err != nil {
			return fmt.Errorf("rand: %w", err)
		}
		next := time.Duration(int64(from)+offset.Int64()) * time.Second
		nextMock := time.Now().Add(next)

		if err := db.SetJson(tx, key, nextMock); err != nil {
			return fmt.Errorf("set %q: %w", key, err)
		}

		return nil
	})
}
