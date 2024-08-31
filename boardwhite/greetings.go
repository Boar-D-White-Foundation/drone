package boardwhite

import (
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/iterx"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) OnGreetJoinedUser(ctx context.Context, c tele.Context) error {
	msg := c.Message()
	if msg == nil || msg.UserJoined == nil {
		return nil
	}

	template, err := iterx.PickRandom(s.cfg.GreetingsTemplates)
	if err != nil {
		return fmt.Errorf("pick template :%w", err)
	}

	username := tg.BuildMentionMarkdownV2(msg.UserJoined)
	greetMessage := fmt.Sprintf(template, username)
	_, err = s.telegram.SendMarkdownV2(s.cfg.FloodThreadID, greetMessage)
	if err != nil {
		return fmt.Errorf("greet user: %w", err)
	}

	return nil
}
