package boardwhite

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/frosthamster/drone/src/tg"
)

type Service struct {
	leetcodeThreadID int
	dailyLCStickerID string
	dailyNCStickerID string
	db               *badger.DB
	telegram         *tg.Client
}

func NewService(
	leetcodeThreadID int,
	dailyLCStickerID string,
	dailyNCStickerID string,
	telegram *tg.Client,
	db *badger.DB,
) *Service {
	return &Service{
		leetcodeThreadID: leetcodeThreadID,
		dailyLCStickerID: dailyLCStickerID,
		dailyNCStickerID: dailyNCStickerID,
		db:               db,
		telegram:         telegram,
	}
}

func (s *Service) publish(header, text, stickerID string, key []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		switch {
		case err == nil:
			err = item.Value(func(val []byte) error {
				pinnedId, _ := strconv.Atoi(string(val))
				err = s.telegram.Unpin(pinnedId)
				if err != nil {
					slog.Error("err unpin", slog.Any("err", err))
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
