package boardwhite

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/boar-d-white-foundation/drone/tg"
	"github.com/dgraph-io/badger/v4"
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
	db       *badger.DB
	telegram *tg.Client
}

func NewService(
	cfg ServiceConfig,
	telegram *tg.Client,
	db *badger.DB,
) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Service{
		ServiceConfig: cfg,
		db:            db,
		telegram:      telegram,
	}, nil
}

func (s *Service) Start() {
	s.telegram.RegisterHandler(NeetCodeCounter{db: s.db})
	s.telegram.RegisterHandler(ReactionHandler{})
	if s.MockEgor.Enabled {
		s.telegram.RegisterHandler(newMockEgorHandler(s.MockEgor, s.db, s.telegram))
	}
	s.telegram.Start()
}

func (s *Service) Stop() {
	s.telegram.Stop()
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
