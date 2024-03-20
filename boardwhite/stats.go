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

type UserStats struct {
	Messages int64 `json:"messsages"`
}

func (s *Service) UserStatistics(ctx context.Context, c tele.Context) error {

	msg, sender := c.Message(), c.Sender()

	if msg == nil || sender == nil {
		return nil
	}

	username := sender.Username
	key := s.statsKey(username)

	return s.database.Do(ctx, func(tx db.Tx) error {

		newCounter := &UserStats{}
		counter, err := db.GetJson[UserStats](tx, key)
		switch {
		case err == nil:
			newCounter.Messages = counter.Messages + 1

		case errors.Is(err, db.ErrKeyNotFound):
			newCounter.Messages = 1
		default:
			return fmt.Errorf("get %q: %w", key, err)

		}

		err = db.SetJson(tx, key, newCounter)
		if err != nil {
			return fmt.Errorf("set %q: %w", key, err)
		}

		return nil
	})

}
