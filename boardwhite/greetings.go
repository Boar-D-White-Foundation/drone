package boardwhite

import (
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/iter"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) makeGreetJoinedUserHandler() func(context.Context, tele.Context) error {
	return func(ctx context.Context, c tele.Context) error {
		msg := c.Message()

		if msg == nil || msg.UserJoined == nil {
			return nil
		}

		template, err := iter.PickRandom(s.cfg.Greetings)

		if err != nil {
			return fmt.Errorf("greet user :%w", err)
		}

		username := BuildUsername(msg.UserJoined)

		greetMessage := fmt.Sprintf(template, username)

		_, err = s.telegram.SendMarkdownV2(s.cfg.GreetingsThreadID, greetMessage)

		if err != nil {
			return fmt.Errorf("greet user: %w", err)
		}

		return nil
	}
}

func BuildUsername(user *tele.User) (name string) {
	if len(user.Username) > 0 {
		name = "@" + tg.EscapeMD(user.Username)
	} else {
		name = fmt.Sprintf("[%s](tg://user?id=%d)", tg.EscapeMD(user.FirstName), user.ID)
	}
	return name
}
