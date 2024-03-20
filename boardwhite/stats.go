package boardwhite

import (
	"context"
	"errors"
	"fmt"

	"github.com/boar-d-white-foundation/drone/db"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) statsKey(username string) string {
	return "boardwhite:stats:" + username
}

type userStats struct {
	MessagesCount int64 `json:"messsagesCount"`
}

func (s *Service) UserStatistics(ctx context.Context, c tele.Context) error {

	msg, sender := c.Message(), c.Sender()

	if msg == nil || sender == nil {
		return nil
	}

	username := sender.Username
	key := s.statsKey(username)

	return s.database.Do(ctx, func(tx db.Tx) error {

		counter, err := db.GetJsonDefault[userStats](tx, key, userStats{})
		switch {
		case err == nil:
			counter.MessagesCount++

		case errors.Is(err, db.ErrKeyNotFound):
			counter.MessagesCount = 1
		default:
			return fmt.Errorf("get %q: %w", key, err)

		}

		err = db.SetJson(tx, key, counter)
		if err != nil {
			return fmt.Errorf("set %q: %w", key, err)
		}

		return nil
	})

}
