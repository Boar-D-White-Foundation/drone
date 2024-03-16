package boardwhite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/tg"
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
func (r ReactionHandler) Handle(client *tg.Client, c tele.Context) error {
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
	ctx      context.Context
	db       db.DB
	telegram *tg.Client
}

func newMockEgorHandler(cfg MockEgorConfig, ctx context.Context, db db.DB, telegram *tg.Client) *mockEgorHandler {
	return &mockEgorHandler{
		MockEgorConfig: cfg,
		ctx:            ctx,
		db:             db,
		telegram:       telegram,
	}
}

func (h *mockEgorHandler) Match(c tele.Context) bool {
	return c.Sender().Username == byegor
}

func (h *mockEgorHandler) Handle(client *tg.Client, c tele.Context) error {
	message := c.Message()

	key := "mock:byegor"

	return h.db.Do(h.ctx, func(tx db.Tx) error {
		shouldMock := true
		t, err := db.GetJson[time.Time](tx, key)
		switch {
		case err == nil:
			if time.Since(t) < h.Period {
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

		_, err = h.telegram.ReplyWithSticker(h.StickerID, message)
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
