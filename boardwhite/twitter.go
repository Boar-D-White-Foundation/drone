package boardwhite

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	tele "gopkg.in/telebot.v3"
)

var twitterLinkRe = regexp.MustCompile(`(https?://)?(www\.)?\b(x\.com|twitter\.com)/[-a-zA-Z0-9@:%_+~#?&/=]+`)

func (s *Service) OnPostTwitterEmbed(ctx context.Context, c tele.Context) error {
	msg, chat := c.Message(), c.Chat()
	if msg == nil || chat == nil || chat.ID != s.cfg.ChatID {
		return nil
	}

	matchedTwitterLinks := twitterLinkRe.FindStringSubmatch(msg.Text)
	if len(matchedTwitterLinks) == 0 {
		return nil
	}

	firstTwitterLink := matchedTwitterLinks[0]
	embedTwitterLink := strings.Replace(firstTwitterLink, "x.com/", "i.fixupx.com/", 1)
	embedTwitterLink = strings.Replace(embedTwitterLink, "twitter.com/", "i.fixupx.com/", 1)

	_, err := s.telegram.ReplyWithText(msg.ID, embedTwitterLink)
	if err != nil {
		return fmt.Errorf("reply with twitter embed: %w", err)
	}

	return nil
}
