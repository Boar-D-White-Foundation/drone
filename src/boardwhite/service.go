package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/frosthamster/drone/src/leetcode"
	"github.com/frosthamster/drone/src/tg"
)

type Service struct {
	leetcodeThreadID int
	db               *badger.DB
	telegram         *tg.Client
}

func NewService(leetcodeThreadID int, telegram *tg.Client, db *badger.DB) *Service {
	return &Service{
		leetcodeThreadID: leetcodeThreadID,
		db:               db,
		telegram:         telegram,
	}
}

const (
	defaultDailyHeader = "LeetCode Daily Question"

	keyLeetcodePinnedMessage = "leetcode:pinned_message"
)

func (s *Service) PublishLCDaily(ctx context.Context) error {
	link, err := leetcode.GetDailyLink(ctx)
	if err != nil {
		return fmt.Errorf("get link: %w", err)
	}

	key := []byte(keyLeetcodePinnedMessage)
	err = s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		switch {
		case err == nil:
			err = item.Value(func(val []byte) error {
				pinnedId, _ := strconv.Atoi(string(val))
				err = s.telegram.Unpin(pinnedId)
				if err != nil {
					return fmt.Errorf("unpin %w", err)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("get value %w", err)
			}
		case errors.Is(err, badger.ErrKeyNotFound):
			// do nothing if there is no previous pin
		default:
			return fmt.Errorf("get key %q: %w", key, err)
		}

		messageID, err := s.telegram.SendMessage(s.leetcodeThreadID, defaultDailyHeader, link)
		if err != nil {
			return fmt.Errorf("send daily: %w", err)
		}

		err = s.telegram.Pin(messageID)
		if err != nil {
			return fmt.Errorf("pin: %w", err)
		}

		err = txn.Set(key, []byte(strconv.Itoa(messageID)))
		if err != nil {
			return fmt.Errorf("set key %q: %w", key, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("db update: %w", err)
	}

	return nil
}
