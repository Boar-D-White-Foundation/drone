package boardwhite

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/boar-d-white-foundation/drone/tg"
	"github.com/dgraph-io/badger/v4"
	tele "gopkg.in/telebot.v3"
)

const (
	byegor = "byegor"
)

type ReactionHandler struct{}

func (r ReactionHandler) Match(c tele.Context) bool {
	return true
}

// TODO: rewrite this to do something useful
func (r ReactionHandler) Handle(c tele.Context) error {
	// user := c.Sender()
	// message := c.Message()

	// reaction := []*tg.ReactionEmoji{}
	// switch user.Username {
	// case "quazeeee":
	// 	reaction = append(reaction, tg.ReactionClown)
	// case "ollkostin":
	// 	reaction = append(reaction, tg.ReactionHotDog)
	// }

	// if len(reaction) > 0 {
	// 	reactionOptions := &tg.ReactionOptions{MessageID: message.ID, ChatID: message.Chat.ID, Reaction: reaction, IsBig: true}
	// 	_, err := c.Bot().Raw("setMessageReaction", reactionOptions)
	// 	if err != nil {
	// 		slog.Error("err react", slog.Any("err", err))
	// 	}

	// }
	return nil
}

type MockEgorConfig struct {
	Enabled   bool
	Period    time.Duration
	StickerID string
}

type mockEgorHandler struct {
	MockEgorConfig
	db       *badger.DB
	telegram *tg.Client
}

func newMockEgorHandler(cfg MockEgorConfig, db *badger.DB, telegram *tg.Client) *mockEgorHandler {
	return &mockEgorHandler{
		MockEgorConfig: cfg,
		db:             db,
		telegram:       telegram,
	}
}

func (h *mockEgorHandler) Match(c tele.Context) bool {
	return c.Sender().Username == byegor
}

func (h *mockEgorHandler) Handle(c tele.Context) error {
	message := c.Message()

	key := []byte("mock:byegor")

	err := h.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		shouldMock := true
		switch {
		case err == nil:
			err := item.Value(func(val []byte) error {
				t, err := time.Parse(time.RFC3339, string(val))
				if err != nil {
					slog.Warn("invalid time value", slog.Any("value", t), slog.Any("err", err))
					return nil
				}

				if time.Since(t) < h.Period {
					shouldMock = false
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("value: %w", err)
			}
		case errors.Is(err, badger.ErrKeyNotFound):
			// still do it
		default:
			return fmt.Errorf("get key %q: %w", key, err)
		}

		if !shouldMock {
			return nil
		}

		_, err = h.telegram.ReplyWithSticker(h.StickerID, message)
		if err != nil {
			return fmt.Errorf("reply with sticker: %w", err)
		}

		t := time.Now().Format(time.RFC3339)
		err = txn.Set(key, []byte(t))
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
