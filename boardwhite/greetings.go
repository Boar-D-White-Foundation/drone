package boardwhite

import (
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/db"
	"github.com/boar-d-white-foundation/drone/iterx"
	"github.com/boar-d-white-foundation/drone/tg"
	tele "gopkg.in/telebot.v3"
)

func (s *Service) OnGreetJoinedUser(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || msg.UserJoined == nil || msg.UserJoined.IsBot || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}

	return s.database.Do(ctx, func(tx db.Tx) error {
		greetedUsers, err := db.GetJsonDefault(tx, keyOnJoinGreetedUsers, make(map[int64]struct{}))
		if err != nil {
			return fmt.Errorf("get greetedUsers: %w", err)
		}

		var templates []string
		if _, ok := greetedUsers[msg.UserJoined.ID]; ok {
			templates = s.cfg.GreetingsOldUsersTemplates
		} else {
			templates = s.cfg.GreetingsNewUsersTemplates
		}

		template, err := iterx.PickRandom(templates)
		if err != nil {
			return fmt.Errorf("pick template :%w", err)
		}

		username := tg.BuildMentionMarkdownV2(*msg.UserJoined)
		greetMessage := fmt.Sprintf(template, username)
		_, err = s.telegram.SendMarkdownV2(s.cfg.FloodThreadID, greetMessage)
		if err != nil {
			return fmt.Errorf("greet user: %w", err)
		}

		greetedUsers[msg.UserJoined.ID] = struct{}{}
		if err := db.SetJson(tx, keyOnJoinGreetedUsers, greetedUsers); err != nil {
			return fmt.Errorf("set keyOnJoinGreetedUsers: %w", err)
		}

		return nil
	})
}
