package boardwhite

import (
	// "log/slog"

	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
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
